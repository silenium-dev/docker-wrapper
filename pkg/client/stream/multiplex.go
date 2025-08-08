package stream

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strings"

	"go.uber.org/zap"
)

type StreamType int

const (
	StreamTypeStdin       StreamType = 0
	StreamTypeStdout      StreamType = 1
	StreamTypeStderr      StreamType = 2
	StreamTypeSystemError StreamType = 3
)

func (s StreamType) Name() string {
	switch s {
	case StreamTypeStdin:
		return "stdin"
	case StreamTypeStdout:
		return "stdout"
	case StreamTypeStderr:
		return "stderr"
	case StreamTypeSystemError:
		return "system_error"
	default:
		return "unknown"
	}
}

func (s StreamType) IsValid() bool {
	return s.Name() != "unknown"
}

type MultiplexedStream struct {
	io.Closer
	reader io.Reader
	writer io.Writer

	logger *zap.SugaredLogger

	stdOut      chan []byte
	stdErr      chan []byte
	systemError chan []byte
	done        chan struct{}
}

func NewMultiplexedStream(
	ctx context.Context,
	reader io.Reader,
	closer io.Closer,
	writer io.Writer,
	multiplex bool,
	logger *zap.SugaredLogger,
) *MultiplexedStream {
	m := &MultiplexedStream{
		Closer:      closer,
		reader:      reader,
		writer:      writer,
		stdOut:      make(chan []byte, 100),
		stdErr:      make(chan []byte, 100),
		systemError: make(chan []byte, 100),
		done:        make(chan struct{}),
		logger:      logger,
	}

	if multiplex {
		go m.handleMultiplexOutput(ctx)
	} else {
		go m.handleSimplexOutput(ctx)
	}

	return m
}

func (m *MultiplexedStream) Stdout() <-chan []byte {
	return m.stdOut
}

func (m *MultiplexedStream) Stderr() <-chan []byte {
	return m.stdErr
}

func (m *MultiplexedStream) SystemError() <-chan []byte {
	return m.systemError
}

func (m *MultiplexedStream) Done() <-chan struct{} {
	return m.done
}

func (m *MultiplexedStream) Write(data []byte) (int, error) {
	if m.writer == nil {
		return 0, errors.New("writer is nil")
	}
	return m.writer.Write(data)
}

func (m *MultiplexedStream) handleSimplexOutput(ctx context.Context) {
	defer close(m.done)
	defer close(m.stdOut)
	defer close(m.stdErr)
	defer close(m.systemError)
	go func() {
		<-ctx.Done()
		_ = m.Close()
	}()

	scanner := bufio.NewScanner(m.reader)
	for ctx.Err() == nil && scanner.Scan() {
		line := scanner.Bytes()
		m.stdOut <- line
	}
	err := scanner.Err()
	if err != nil {
		m.logger.Errorf("failed to read simplex output: %v", err)
	}
}

func (m *MultiplexedStream) handleMultiplexOutput(ctx context.Context) {
	defer close(m.done)
	defer close(m.stdOut)
	defer close(m.stdErr)
	defer close(m.systemError)
	go func() {
		<-ctx.Done()
		_ = m.Close()
	}()

	header := make([]byte, 8)
	for ctx.Err() == nil {
		_, err := io.ReadFull(m.reader, header)
		if isEOF(err) {
			m.logger.Debugf("EOF")
			break
		} else if err != nil {
			m.logger.Errorf("failed to read multiplexed output header: %v", err)
			return
		}
		streamType := StreamType(header[0])
		if !streamType.IsValid() {
			m.logger.Errorf("invalid stream type: %d", streamType)
			return
		}
		size := uint32(0)
		for i := 0; i < 4; i++ {
			size <<= 8
			size |= uint32(header[i+4])
		}

		data := make([]byte, size)
		_, err = io.ReadFull(m.reader, data)
		if isEOF(err) {
			m.logger.Debugf("EOF")
			break
		} else if err != nil {
			m.logger.Errorf("failed to read multiplexed output data: %v", err)
			return
		}

		switch streamType {
		case StreamTypeStdout:
			m.stdOut <- data
		case StreamTypeStderr:
			m.stdErr <- data
		case StreamTypeSystemError:
			m.systemError <- data
		default:
			m.logger.Warnf("unexpected stream type: %s", streamType.Name())
		}
	}
	m.logger.Debugf("multiplexed output handling completed, closing channels")
}

func isEOF(err error) bool {
	return err != nil && (errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		strings.Contains(err.Error(), "http: read on closed response body") ||
		strings.Contains(err.Error(), "file has already been closed") ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded))
}

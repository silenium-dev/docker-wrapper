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

type Type int

const (
	TypeStdin       Type = 0
	TypeStdout      Type = 1
	TypeStderr      Type = 2
	TypeSystemError Type = 3
)

func (s Type) Name() string {
	switch s {
	case TypeStdin:
		return "stdin"
	case TypeStdout:
		return "stdout"
	case TypeStderr:
		return "stderr"
	case TypeSystemError:
		return "system_error"
	default:
		return "unknown"
	}
}

func (s Type) IsValid() bool {
	return s.Name() != "unknown"
}

type Message struct {
	StreamType Type
	Content    []byte
}

type MultiplexedStream struct {
	closer io.Closer
	reader io.Reader
	writer io.Writer

	logger *zap.SugaredLogger

	messageChan chan Message
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
		closer:      closer,
		reader:      reader,
		writer:      writer,
		messageChan: make(chan Message, 100),
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

func (m *MultiplexedStream) Messages() chan Message {
	return m.messageChan
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

func (m *MultiplexedStream) Close() error {
	if m.closer == nil {
		return nil
	}
	err := m.closer.Close()
	if err != nil {
		m.logger.Errorf("failed to close multiplexed stream: %v", err)
		return err
	}
	m.logger.Debugf("multiplexed stream closed")
	return nil
}

func (m *MultiplexedStream) handleSimplexOutput(ctx context.Context) {
	defer close(m.done)
	defer close(m.messageChan)
	go func() {
		<-ctx.Done()
		_ = m.Close()
	}()

	scanner := bufio.NewScanner(m.reader)
	for ctx.Err() == nil && scanner.Scan() {
		line := scanner.Bytes()
		m.messageChan <- Message{
			Content:    line,
			StreamType: TypeStdout,
		}
	}
	err := scanner.Err()
	if err != nil {
		m.logger.Errorf("failed to read simplex output: %v", err)
	}
}

func (m *MultiplexedStream) handleMultiplexOutput(ctx context.Context) {
	defer close(m.done)
	defer close(m.messageChan)
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
		streamType := Type(header[0])
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

		m.messageChan <- Message{
			StreamType: streamType,
			Content:    data,
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

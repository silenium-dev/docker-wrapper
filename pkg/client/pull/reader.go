package pull

import (
	"bufio"
	"context"
	"docker-wrapper/pkg/client/pull/base"
	"docker-wrapper/pkg/client/pull/events"
	"encoding/json"
	"io"
	"log"
)

func ParseStream(ctx context.Context, reader io.ReadCloser) chan events.PullEvent {
	result := make(chan events.PullEvent)
	go parseEvents(ctx, reader, result)
	return result
}

func parseEvents(ctx context.Context, reader io.ReadCloser, ch chan events.PullEvent) {
	defer close(ch)
	defer func() { _ = reader.Close() }()

	scan := bufio.NewScanner(reader)
	for scan.Scan() {
		var raw base.PullProgressEvent
		err := json.Unmarshal(scan.Bytes(), &raw)
		if err != nil {
			panic(err)
		}
		event, err := events.Parse(raw)
		if err != nil {
			panic(err)
		}
		select {
		case ch <- event:
		case <-ctx.Done():
			return
		}
	}
	if err := scan.Err(); err != nil {
		log.Printf("error reading pull stream: %v", err)
	}
}

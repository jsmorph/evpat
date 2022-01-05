package sse

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jsmorph/evpat/bus"
	"github.com/jsmorph/evpat/pat"
)

type SSE struct {
	SessionLimit int
	Bus          *bus.Bus
}

func NewSSE(bus *bus.Bus) *SSE {
	s := &SSE{
		SessionLimit: 10000,
		Bus:          bus,
	}
	if bus != nil {
		bus.History = s.History
	}
	return s
}

func (s *SSE) History(ctx context.Context, q *bus.Query) (chan []bus.Msg, error) {
	// ToDo: Query some data store to get old messages

	return nil, nil
}

func (s *SSE) Ingest(ctx context.Context, msgs []bus.Msg) error {
	// ToDo: Add to history data store

	select {
	case <-ctx.Done():
		return bus.Canceled
	case s.Bus.Incoming <- msgs:
	}
	return nil
}

func (s *SSE) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	log.Printf("SSE.Handle")

	// ToDo: Get query parameters (if any)

	// Register bus.Consumer and defer deregistration.

	c := &bus.Consumer{
		Outgoing: make(chan []bus.Msg),
	}

	select {
	case <-ctx.Done():
		return bus.Canceled
	case s.Bus.AddConsumer <- c:
		log.Printf("SSE.Handle added consumer")
		defer func() {
			select {
			case <-ctx.Done():
				return
			case s.Bus.RemConsumer <- c:
			}
		}()
	}

	count := 0

LOOP:
	for {
		select {
		case <-ctx.Done():
			return bus.Canceled
		case msgs := <-c.Outgoing:
			log.Printf("SSE.Handle got %d msgs", len(msgs))
			for _, msg := range msgs {
				js := pat.JSON(msg)
				js = strings.TrimSpace(js)
				e := fmt.Sprintf("event: %s\nid: %s\ndata: %s\n\n", msg.Type, msg.Id, js)
				if _, err := w.Write([]byte(e)); err != nil {
					log.Printf("INFO Handler Write error %s", err)
					break LOOP
				}
				count++

				if s.SessionLimit <= count {
					break LOOP
				}
			}
		}
	}

	log.Printf("SSE.Handle terminating")

	return nil
}

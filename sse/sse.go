package sse

import (
	"context"
	"net/http"

	"github.com/jsmorph/evpat/bus"
)

type SSE struct {
	Bus *bus.Bus
}

func NewSSE(bus *bus.Bus) *SSE {
	s := &SSE{
		Bus: bus,
	}
	if bus != nil {
		bus.History = s.History
	}
	return s
}

func (s *SSE) History(ctx context.Context, q *bus.Query) (chan []bus.Msg, error) {
	return nil, nil
}

func (s *SSE) Ingest(ctx context.Context, msgs []bus.Msg) error {
	// Add to history
	// Send to Bus
	return nil
}

func (s *SSE) Handle(w http.ResponseWriter, r http.Request) error {
	// Get last id (if any)
	// Get pattern (if any)
	// Create bus.Consumer
	// Register bus.Consumer and defer deregistration
	// Consume our Consumer channel
	return nil
}

func (s *SSE) forward(ctx context.Context, msg bus.Msg, w http.ResponseWriter) error {
	return nil
}

package sse

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jsmorph/evpat/bus"
)

func TestSSE(t *testing.T) {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		b           = bus.NewBus(10, time.Second, time.Second)
		s           = NewSSE(b)
	)
	defer cancel()
	s.SessionLimit = 3

	go b.Run(ctx)

	input := func() {
		msg := bus.Msg{
			Type: "desire",
			Id:   time.Now().Format(time.RFC3339Nano),
			Payload: map[string]interface{}{
				"want": "tacos",
			},
		}

		select {
		case <-ctx.Done():
			t.Fatal("canceled")
		case b.Incoming <- []bus.Msg{msg}:
			log.Printf("sending msgs")
		}
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		s.Handle(ctx, w, r)
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()

	go func() {
		res, err := http.Get(ts.URL)
		if err != nil {
			log.Fatal(err)
		}
		bs, err := io.ReadAll(res.Body)
		res.Body.Close()
		fmt.Printf("%s\n", bs)
	}()

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		input()
		time.Sleep(time.Second)
	}
}

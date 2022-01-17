package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

	return s
}

func punt(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}

func (s *SSE) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	log.Printf("SSE.Handle")

	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		if err != nil {
			punt(w, http.StatusBadRequest, "bad query string %s: %s\n", r.URL.RawQuery, err)
		}
	}

	p := q.Get("limit")
	limit := bus.DefaultQuery.Limit
	if p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			punt(w, http.StatusBadRequest, "bad limit %s: %s\n", p, err)
		}
		limit = n
	}

	p = q.Get("replay")
	log.Printf("debug replay %s", p)
	replay := bus.DefaultQuery.Replay
	switch strings.ToLower(p) {
	case "true":
		replay = true
	case "false":
		replay = false
	case "":
	default:
		punt(w, http.StatusBadRequest, "bad replay %s\n", p)
		return nil
	}

	js, err := ioutil.ReadAll(r.Body)
	if err != nil {
		punt(w, http.StatusBadRequest, "failed to read filter: %s\n", err)
		return nil
	}
	var filter pat.Constraint = pat.Pass
	if 0 < len(js) {
		var x interface{}
		if err = json.Unmarshal(js, &x); err != nil {
			punt(w, http.StatusBadRequest, "bad filter %s: (%s)\n", js, err)
			return nil
		}
		p, err := pat.DefaultCfg.ParsePattern(x)
		if err != nil {
			punt(w, http.StatusBadRequest, "bad filter %s: (%s)\n", js, err)
			return nil
		}
		filter = p
	}

	// Register bus.Consumer and defer deregistration.

	c := &bus.Consumer{
		Outgoing: make(chan []bus.Msg),
		Query: &bus.Query{
			Replay: replay,
			Filter: filter,
			Limit:  limit,
		},
	}

	log.Printf("SSE.Handler adding consumer")
	select {
	case <-ctx.Done():
		return bus.Canceled
	case s.Bus.AddConsumer <- c:
		defer func() {
			log.Printf("SSE.Handler removing consumer")
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
				var e string

				if msg.Type != "" {
					e = fmt.Sprintf("event: %s\n", msg.Type)
				}

				if msg.Id != "" {
					e += fmt.Sprintf("id: %s\n", msg.Id)
				}

				js := pat.JSON(msg)
				js = strings.TrimSpace(js)
				e += fmt.Sprintf("data: %s\n\n", js)

				if _, err := w.Write([]byte(e)); err != nil {
					log.Printf("INFO Handler Write error %s", err)
					break LOOP
				}

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
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

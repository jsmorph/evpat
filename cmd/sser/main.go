package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jsmorph/evpat/bus"
	"github.com/jsmorph/evpat/sse"

	"github.com/go-redis/redis/v8"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		topics       = flag.String("topics", "test", "comma-separated Redis PUBSUB keys")
		httpPort     = flag.String("listen", ":8000", "HTTP service port")
		redisPort    = flag.String("redis", "localhost:6379", "Redis host:port")
		sessionLimit = flag.Int("session-limit", 1000, "Max events per session")
		maxReplay    = flag.Int("max-replay", 100, "max messages to replay for a client")

		ctx, cancel = context.WithCancel(context.Background())
		b           = bus.NewBus(10, time.Second, time.Second) // ToDo
		db          = bus.NewRing(100)                         // ToDo
		s           = sse.NewSSE(b)
	)
	defer cancel()

	flag.Parse()

	s.SessionLimit = *sessionLimit
	b.DB = db
	b.MaxReplay = *maxReplay

	ropts := &redis.Options{
		Addr: *redisPort,
	}

	go b.Run(ctx)

	for _, topic := range strings.Split(*topics, ",") {
		go func(topic string) {
			var (
				r   = redis.NewClient(ropts)
				sub = r.Subscribe(ctx, topic)
			)

			for {
				m, err := sub.ReceiveMessage(ctx)
				if err != nil {
					log.Fatal(err)
				}

				var x interface{}
				if err := json.Unmarshal([]byte(m.Payload), &x); err != nil {
					x = map[string]interface{}{
						"go": fmt.Sprintf("%#v", m.Payload),
					}
				}

				msg := bus.Msg{
					Type:    topic, // Eh
					Id:      time.Now().Format(time.RFC3339Nano),
					Payload: x,
				}
				select {
				case <-ctx.Done():
					log.Fatal("canceled")
				case b.Incoming <- []bus.Msg{msg}:
				}
			}
		}(strings.TrimSpace(topic))
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		s.Handle(ctx, w, r)
	}

	return http.ListenAndServe(*httpPort, http.HandlerFunc(h))
}

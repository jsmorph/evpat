package bus

import (
	"context"
	"fmt"
	"time"

	"github.com/jsmorph/evpat/pat"
)

type Msg struct {
	Type    string      `json:"type,omitempty"`
	Payload interface{} `json:"payload"`
	Id      string      `json:"id,omitempty"`
}

type Consumer struct {
	Query    *Query
	Outgoing chan []Msg
}

type Query struct {
	Replay   bool
	From, To string
	Limit    int
	Filter   pat.Constraint
}

type Bus struct {
	History func(ctx context.Context, q *Query) (chan []Msg, error)

	Incoming    chan []Msg
	AddConsumer chan *Consumer
	RemConsumer chan *Consumer

	consumerTimeout time.Duration
	workersTimeout  time.Duration
	ws              *Workers
}

func NewBus(workers int, workersTimeout, consumerTimeout time.Duration) *Bus {
	return &Bus{
		Incoming:        make(chan []Msg),
		AddConsumer:     make(chan *Consumer),
		RemConsumer:     make(chan *Consumer),
		workersTimeout:  workersTimeout,
		consumerTimeout: consumerTimeout,
		ws:              NewWorkers(workers),
	}
}

func (b *Bus) work(ctx context.Context, f func(context.Context) error) error {
	wsctx, _ := context.WithTimeout(ctx, b.workersTimeout)
	i, err := b.ws.Get(wsctx)
	if err != nil {
		return err
	}

	go func() {
		defer b.ws.Return(ctx, i)
		f(ctx)
	}()

	return nil
}

var (
	Canceled = fmt.Errorf("canceled")
	Timeout  = fmt.Errorf("timeout")
)

func (b *Bus) Run(ctx context.Context) error {

	clients := make(map[*Consumer]bool)

	for {
		select {
		case <-ctx.Done():
			return Canceled
		case msgs := <-b.Incoming:
			f := func(ctx context.Context) error {
				for c := range clients {
					b.forward(ctx, c, msgs)
				}
				return nil
			}
			b.work(ctx, f)
		case c := <-b.AddConsumer:
			clients[c] = true
			f := func(ctx context.Context) error {
				return b.replay(ctx, c)
			}
			b.work(ctx, f)
		case c := <-b.RemConsumer:
			delete(clients, c)
		}
	}
	return nil
}

func (b *Bus) forward(ctx context.Context, c *Consumer, msgs []Msg) error {

	var filtered []Msg
	if c.Query == nil || c.Query.Filter == nil {
		filtered = msgs
	} else {
		filtered = make([]Msg, 0, len(msgs))
		for _, msg := range msgs {
			if ok, _ := c.Query.Filter.Matches(msg.Payload); ok {
				filtered = append(filtered, msg)
			}
		}
	}

	select {
	case <-ctx.Done():
		return Canceled
	case <-time.NewTimer(b.consumerTimeout).C:
		return Timeout
	case c.Outgoing <- filtered:
		return nil
	}
}

func (b *Bus) replay(ctx context.Context, c *Consumer) error {
	if c.Query == nil || !c.Query.Replay {
		return nil
	}
	in, err := b.History(ctx, c.Query)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return Canceled
		case msgs := <-in:
			if err := b.forward(ctx, c, msgs); err != nil {
				return err
			}
		}
	}
}

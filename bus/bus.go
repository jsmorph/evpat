package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
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
	Replay bool
	Limit  int
	Filter pat.Constraint

	// From, To string
}

var DefaultQuery = &Query{
	Replay: true,
	Limit:  10,
	Filter: pat.Pass,
}

type Cfg struct {
	// NumWorkers is the size of the pool of workers that handle
	// connections.
	//
	// The default is based on the number of CPU cores.
	NumWorkers int

	// MaxReplay is the maximum number of messages to replay.
	//
	// With many systems, this maximum will frequently bit hit.
	MaxReplay int

	// ConsumerTimeout is the length of time to wait on a consumer
	// to accept a message.
	ConsumerTimeout time.Duration

	// WorkersTimeout is the length of time to wait for a worker
	// to handle a connection.
	WorkersTimeout time.Duration
}

var DefaultCfg = &Cfg{
	NumWorkers:      runtime.NumCPU() * 50, // ?
	MaxReplay:       100,
	ConsumerTimeout: 20 * time.Second,
	WorkersTimeout:  10 * time.Second,
}

type Bus struct {
	*Cfg

	DB DB

	Incoming    chan []Msg
	AddConsumer chan *Consumer
	RemConsumer chan *Consumer

	ws *WorkersPool
}

func (cfg *Cfg) New() *Bus {
	return &Bus{
		Cfg:         cfg,
		Incoming:    make(chan []Msg),
		AddConsumer: make(chan *Consumer),
		RemConsumer: make(chan *Consumer),
		ws:          NewWorkersPool(cfg.NumWorkers),
	}
}

func NewBus() *Bus {
	return DefaultCfg.New()
}

func (b *Bus) work(ctx context.Context, f func(context.Context) error) error {
	wsctx, _ := context.WithTimeout(ctx, b.WorkersTimeout)
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
			if b.DB != nil {
				if err := b.DB.Write(ctx, msgs); err != nil {
					return err
				}
			}
			f := func(ctx context.Context) error {
				for c := range clients {
					b.forward(ctx, c, msgs)
				}
				return nil
			}
			b.work(ctx, f)
		case c := <-b.AddConsumer:
			if 0 < b.MaxReplay && b.MaxReplay < c.Query.Limit {
				c.Query.Limit = b.MaxReplay
			}
			clients[c] = true
			f := func(ctx context.Context) error {
				return b.Replay(ctx, c)
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
	if c.Query == nil {
		c.Query = DefaultQuery
	}
	filtered = make([]Msg, 0, len(msgs))
	for _, msg := range msgs {
		x := Canonicalize(msg)
		if ok, _ := c.Query.Filter.Matches(x); ok {
			filtered = append(filtered, msg)
		}
	}

	select {
	case <-ctx.Done():
		return Canceled
	case <-time.NewTimer(b.ConsumerTimeout).C:
		return Timeout
	case c.Outgoing <- filtered:
		return nil
	}
}

func Canonicalize(x interface{}) interface{} {
	js, err := json.Marshal(&x)
	if err != nil {
		return x
	}
	var y interface{}
	if err = json.Unmarshal(js, &y); err != nil {
		return x
	}
	return y
}

func (b *Bus) Replay(ctx context.Context, c *Consumer) error {
	log.Printf("Bus.Replay %#v (DB: %v)", c.Query, b.DB != nil)

	if c.Query == nil || !c.Query.Replay || b.DB == nil {
		return nil
	}
	in, err := b.DB.Read(ctx, c.Query)
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

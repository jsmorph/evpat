package bus

import (
	"context"
	"log"
	"sync"
)

type Ring struct {
	size, at int
	buf      []*Msg

	sync.RWMutex
}

func NewRing(size int) *Ring {
	return &Ring{
		size: size,
		buf:  make([]*Msg, size),
	}
}

func (r *Ring) Open(context.Context) error {
	return nil
}

func (r *Ring) Close(context.Context) error {
	return nil
}

func (r *Ring) add(msg *Msg) {
	r.buf[r.at] = msg
	r.at++
	if r.size <= r.at {
		r.at = 0
	}
}

func (r *Ring) Write(ctx context.Context, msgs []Msg) error {
	r.Lock()
	for _, msg := range msgs {
		r.add(&msg)
	}
	r.Unlock()
	return nil
}

func (r *Ring) ReplayForward(n int) []*Msg {

	acc := make([]*Msg, 0, n)

	r.RLock()

	at := r.at
	for i := 0; i < r.size; i++ {
		msg := r.buf[at]
		at++
		if r.size <= at {
			at = 0
		}
		if msg == nil {
			continue
		}
		acc = append(acc, msg)
		if n <= len(acc) {
			break
		}
	}

	r.RUnlock()

	return acc
}

func (r *Ring) ReplayRecent(n int) []*Msg {
	if n <= 0 {
		n = r.size
	}
	if n > r.size {
		n = r.size
	}
	acc := make([]*Msg, 0, n)

	r.RLock()

	at := r.at - n
	if at < 0 {
		at = r.size + at
	}
	for i := 0; i < r.size; i++ {
		msg := r.buf[at]
		at++
		if r.size <= at {
			at = 0
		}
		if msg == nil {
			continue
		}
		acc = append(acc, msg)
		if n <= len(acc) {
			break
		}
	}

	r.RUnlock()

	return acc
}

func (r *Ring) Read(ctx context.Context, q *Query) (chan []Msg, error) {
	if q == nil {
		q = DefaultQuery
	}

	var (
		c    = make(chan []Msg, q.Limit)
		msgs = r.ReplayRecent(0)
		n    = 0
	)

LOOP:
	for _, msg := range msgs {
		if pass, err := q.Filter.Matches(msg.Payload); err != nil {
			log.Printf("Ring.Read debug error %s", err)
			return nil, err
		} else if pass {
			select {
			case c <- []Msg{*msg}:
				n++
			default:
				break LOOP
			}
		}
	}

	return c, nil
}

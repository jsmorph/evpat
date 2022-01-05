package bus

import "context"

type Workers struct {
	ws chan int
}

func NewWorkers(n int) *Workers {
	ws := make(chan int, n)
	for i := 0; i < n; i++ {
		ws <- i
	}
	return &Workers{
		ws: ws,
	}
}

func (w *Workers) Get(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, Canceled
	case i := <-w.ws:
		return i, nil
	}
}

func (w *Workers) Return(ctx context.Context, i int) error {
	select {
	case <-ctx.Done():
		return Canceled
	case w.ws <- i:
		return nil
	}
}

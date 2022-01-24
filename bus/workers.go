package bus

import "context"

type WorkersPool struct {
	ws chan int
}

func NewWorkersPool(n int) *WorkersPool {
	ws := make(chan int, n)
	for i := 0; i < n; i++ {
		ws <- i
	}
	return &WorkersPool{
		ws: ws,
	}
}

func (w *WorkersPool) Get(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, Canceled
	case i := <-w.ws:
		return i, nil
	}
}

func (w *WorkersPool) Return(ctx context.Context, i int) error {
	select {
	case <-ctx.Done():
		return Canceled
	case w.ws <- i:
		return nil
	}
}

package bus

import "context"

type DB interface {
	Open(context.Context) error
	Close(context.Context) error
	Write(context.Context, []Msg) error
	Read(context.Context, *Query) (chan []Msg, error)
}

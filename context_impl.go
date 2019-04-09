package context_impl

import (
	"errors"
	"time"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

type emptyCtx int

func (emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}
func (emptyCtx) Done() <-chan struct{}             { return nil }
func (emptyCtx) Err() error                        { return nil }
func (emptyCtx) Value(key interface{}) interface{} { return nil }

func Background() Context {
	return new(emptyCtx)
}

func TODO() Context {
	return new(emptyCtx)
}

type CancelFunc func()

type cancelCtx struct {
	parent Context
	done   chan struct{}
	err    error
}

func (ctx *cancelCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.parent.Deadline()
}
func (ctx *cancelCtx) Done() <-chan struct{}             { return ctx.done }
func (ctx *cancelCtx) Err() error                        { return ctx.err }
func (ctx *cancelCtx) Value(key interface{}) interface{} { return ctx.parent.Value(key) }

var Canceled = errors.New("context canceled")

func WithCancel(parent Context) (Context, CancelFunc) {
	ctx := cancelCtx{
		parent: parent,
		done:   make(chan struct{}),
	}
	cancel := func() {
		ctx.err = Canceled
		close(ctx.done)
	}
	return &ctx, cancel
}
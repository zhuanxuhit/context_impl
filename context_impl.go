package context_impl

import (
	"errors"
	"sync"
	"time"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

type emptyCtx int

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func (emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}
func (emptyCtx) Done() <-chan struct{}             { return nil }
func (emptyCtx) Err() error                        { return nil }
func (emptyCtx) Value(key interface{}) interface{} { return nil }

func Background() Context {
	return background
}

func TODO() Context {
	return todo
}

type CancelFunc func()

type cancelCtx struct {
	parent Context
	done   chan struct{}
	err    error
	mu     sync.Mutex
}

func (ctx *cancelCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.parent.Deadline()
}
func (ctx *cancelCtx) Done() <-chan struct{} { return ctx.done }
func (ctx *cancelCtx) Err() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.err
}
func (ctx *cancelCtx) Value(key interface{}) interface{} { return ctx.parent.Value(key) }

var Canceled = errors.New("context canceled")

func WithCancel(parent Context) (Context, CancelFunc) {
	ctx := cancelCtx{
		parent: parent,
		done:   make(chan struct{}),
	}
	cancel := func() {
		ctx.mu.Lock()
		ctx.err = Canceled
		ctx.mu.Unlock()
		close(ctx.done)
	}
	return &ctx, cancel
}

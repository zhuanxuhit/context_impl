package context_impl

import (
	"errors"
	"reflect"
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
	Context
	done chan struct{}
	err  error
	mu   sync.Mutex
}

//func (ctx *cancelCtx) Deadline() (deadline time.Time, ok bool) {
//	return ctx.parent.Deadline()
//}
func (ctx *cancelCtx) Done() <-chan struct{} { return ctx.done }
func (ctx *cancelCtx) Err() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.err
}

//func (ctx *cancelCtx) Value(key interface{}) interface{} { return ctx.parent.Value(key) }

var Canceled = errors.New("context canceled")

func WithCancel(parent Context) (Context, CancelFunc) {
	ctx := cancelCtx{
		Context: parent,
		done:    make(chan struct{}),
	}
	cancel := func() {
		ctx.cancel(Canceled)
	}
	go func() {
		select {
		case <-parent.Done():
			ctx.cancel(parent.Err())
		case <-ctx.done:
		}
	}()
	return &ctx, cancel
}

func (ctx *cancelCtx) cancel(err error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.err != nil {
		return
	}
	ctx.err = err

	close(ctx.done)
}

//var DeadlineExceeded = errors.New("deadline exceeded")
var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string   { return "context deadline exceeded" }
func (deadlineExceededError) Timeout() bool   { return true }
func (deadlineExceededError) Temporary() bool { return true }


type deadlineCtx struct {
	*cancelCtx
	deadline time.Time
}

func (ctx *deadlineCtx) Deadline() (deadline time.Time, ok bool) {
	return ctx.deadline, true
}

func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc) {
	cctx, cancel := WithCancel(parent)

	ctx := deadlineCtx{
		cancelCtx: cctx.(*cancelCtx),
		deadline:  deadline,
	}

	time.AfterFunc(time.Until(deadline), func() {
		ctx.cancel(DeadlineExceeded)
	})

	return &ctx, cancel
}

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

type valueCtx struct {
	Context
	key, value interface{}
}

func (ctx *valueCtx) Value(key interface{}) interface{} {
	if key == ctx.key {
		return ctx.value
	}
	return ctx.Context.Value(key)
}

func WithValue(parent Context, key, val interface{}) Context {
	if key == nil {
		panic("key is nil")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not Comparable")
	}

	ctx := valueCtx{
		Context: parent,
		key:     key,
		value:   val,
	}
	return &ctx
}

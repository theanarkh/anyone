package anyone

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Worker[T any] func() (T, error)

func Run[T any](ctx context.Context, workers []Worker[T]) (T, error) {
	var total = int32(len(workers))
	var count atomic.Int32
	var mutex sync.RWMutex
	var done bool
	var result T
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	for _, worker := range workers {
		go func() {
			var res T
			var e error
			defer func() {
				// make panic as an error
				if p := recover(); p != nil {
					e = fmt.Errorf("worker panic: %v", p)
				}
				count.Add(1)
				mutex.RLock()
				if done {
					mutex.RUnlock()
					return
				}
				mutex.RUnlock()
				mutex.Lock()
				defer mutex.Unlock()
				if done {
					return
				}
				if e == nil || count.Load() == total {
					done = true
					result = res
					err = e
					wg.Done()
				}
			}()
			res, e = worker()
		}()
	}
	wg.Wait()
	return result, err
}

func WithTimeout[T any](ctx context.Context, worker Worker[T], timeout time.Duration) func() (T, error) {
	return func() (result T, err error) {
		return Timeout(ctx, worker, timeout)
	}
}

func Timeout[T any](ctx context.Context, worker Worker[T], timeout time.Duration) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	type Result struct {
		Result T
		Error  error
	}
	dc := make(chan *Result)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				select {
				case dc <- &Result{
					Error: fmt.Errorf("worker panic: %v", p),
				}:
				default:
				}
			}
			close(dc)
		}()
		res, e := worker()
		select {
		case dc <- &Result{
			Result: res,
			Error:  e,
		}:
		default:
		}
	}()
	select {
	case <-ctx.Done():
		var res T
		return res, ctx.Err()
	case res := <-dc:
		return res.Result, res.Error
	}
}

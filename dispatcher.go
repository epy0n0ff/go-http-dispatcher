package dispatcher

import (
	"context"
	"net/http"
)

type Request *http.Request
type Response *responseAndError

type Dispatcher struct {
	job        chan Request
	reqChan    chan chan Request
	resChan    chan Response
	ResultFunc func(Response)
}

func NewDispatcher(workerSize int) *Dispatcher {
	job := make(chan Request, workerSize)
	reqPool := make(chan chan Request, workerSize)
	resPool := make(chan Response, workerSize)
	return &Dispatcher{job, reqPool, resPool, nil}
}

func (d *Dispatcher) Run(ctx context.Context) {
	for i := 0; i < cap(d.reqChan); i++ {
		w := NewDefaultWorker(d.reqChan, d.resChan)
		w.Start(ctx)
	}

	go d.dispatch(ctx)
	go d.fetchResponse(ctx)
}

func (d *Dispatcher) Add(req Request) {
	d.job <- req
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	for {
		select {
		case req := <-d.reqChan:
			go func(req chan Request) {
				r := <-d.job
				req <- r
			}(req)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) fetchResponse(ctx context.Context) {
	go func() {
		for {
			select {
			case res := <-d.resChan:
				if d.ResultFunc != nil {
					d.ResultFunc(res)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

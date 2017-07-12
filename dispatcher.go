package dispatcher

import (
	"context"
	"net/http"
)

type Request *http.Request
type Response *responseAndError

type Dispatcher struct {
	ctx        context.Context
	job        chan Request
	reqChan    chan chan Request
	resChan    chan Response
	ResultFunc func(Response)
}

func NewDispatcher(ctx context.Context, workerSize int) *Dispatcher {
	job := make(chan Request)
	reqPool := make(chan chan Request, workerSize)
	resPool := make(chan Response, workerSize)
	return &Dispatcher{ctx, job, reqPool, resPool, nil}
}

func (d *Dispatcher) Run() {
	for i := 0; i < cap(d.reqChan); i++ {
		w := NewDefaultWorker(d.ctx, d.reqChan, d.resChan)
		w.Start()
	}

	go d.dispatch()
	go d.fetchResponse()
}

func (d *Dispatcher) Add(req Request) {
	d.job <- req
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case req := <-d.reqChan:
			go func(req chan Request) {
				r := <-d.job
				req <- r
			}(req)
		case <- d.ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) fetchResponse() {
	go func() {
		for {
			select {
			case res := <-d.resChan:
				if d.ResultFunc != nil {
					d.ResultFunc(res)
				}
			case <- d.ctx.Done():
				return
			}
		}
	}()
}

package dispatcher

import (
	"context"
	"net/http"
)

type Request *http.Request
type Response *responseAndError

type Dispatcher struct {
	reqJob     chan Request
	reqPool    chan chan Request
	resPool    chan Response
	ResultFunc func(Response)
}

func NewDispatcher(workerSize int) *Dispatcher {
	reqJob := make(chan Request, workerSize)
	reqPool := make(chan chan Request, workerSize)
	resPool := make(chan Response, workerSize)
	return &Dispatcher{reqJob, reqPool, resPool, nil}
}

func (d *Dispatcher) Run(ctx context.Context) {
	for i := 0; i < cap(d.reqPool); i++ {
		w := NewDefaultWorker(d.reqPool, d.resPool)
		w.Start(ctx)
	}

	go d.dispatch(ctx)
	go d.fetchResponse(ctx)
}

func (d *Dispatcher) RunWithHttpClient(ctx context.Context, client *http.Client) {
	for i := 0; i < cap(d.reqPool); i++ {
		w := NewWorkerWithHttpClient(d.reqPool, d.resPool, client)
		w.Start(ctx)
	}

	go d.dispatch(ctx)
	go d.fetchResponse(ctx)
}

func (d *Dispatcher) Add(req Request) {
	d.reqJob <- req
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	for {
		select {
		case req := <-d.reqPool:
			go func(req chan Request) {
				r := <-d.reqJob
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
			case res := <-d.resPool:
				if d.ResultFunc != nil {
					d.ResultFunc(res)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

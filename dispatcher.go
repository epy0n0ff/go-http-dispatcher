package dispatcher

import (
	"context"
	"net/http"
)

type Request *http.Request
type Response *responseAndError

type Dispatcher struct {
	reqJob  chan Request
	reqPool chan chan Request
	resPool chan chan Response
}

func NewDispatcher(workerSize int) *Dispatcher {
	reqJob := make(chan Request, workerSize)
	reqPool := make(chan chan Request, workerSize)
	resPool := make(chan chan Response, workerSize)
	return &Dispatcher{reqJob, reqPool, resPool}
}

func (d *Dispatcher) Run(ctx context.Context) chan chan Response {
	for i := 0; i < cap(d.reqPool); i++ {
		w := NewDefaultWorker(d.reqPool, d.resPool)
		w.Start(ctx)
	}

	go d.dispatch(ctx)

	return d.resPool
}

func (d *Dispatcher) RunWithHttpClient(ctx context.Context, client *http.Client) chan chan Response {
	for i := 0; i < cap(d.reqPool); i++ {
		w := NewWorkerWithHttpClient(d.reqPool, d.resPool, client)
		w.Start(ctx)
	}

	go d.dispatch(ctx)

	return d.resPool
}

func (d *Dispatcher) Add(req Request) {
	d.reqJob <- req
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	for {
		select {
		case req := <-d.reqPool:
			req <- <-d.reqJob
		case <-ctx.Done():
			return
		default:
		}
	}
}

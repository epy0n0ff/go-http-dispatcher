package dispatcher

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

var copyBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return b
	},
}

type Worker struct {
	reqPool chan chan Request
	reqJob  chan Request
	resPool chan chan Response
	resJob  chan Response
	*http.Client
}

type responseAndError struct {
	Resp *http.Response
	Err  error
}

// NewDefaultWorker returns worker pointer having the default http client
func NewDefaultWorker(reqPool chan chan Request, resPool chan chan Response) *Worker {
	return &Worker{
		reqPool,
		make(chan Request, 1),
		resPool,
		make(chan Response, 1),
		http.DefaultClient,
	}
}

// NewDefaultWorker returns worker pointer having the custom http client
func NewWorkerWithHttpClient(reqPool chan chan Request, resPool chan chan Response, client *http.Client) *Worker {
	return &Worker{
		reqPool,
		make(chan Request, 1),
		resPool,
		make(chan Response, 1),
		client,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.reqPool <- w.reqJob

	go func() {
		for {
			select {
			case req := <-w.reqJob:
				resp, err := w.Do(req)
				w.resJob <- &responseAndError{w.copyResponse(resp), err}
				if err == nil {
					resp.Body.Close()
				}
				w.resPool <- w.resJob
				w.reqPool <- w.reqJob
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (w *Worker) copyReader(dst io.Writer, src io.Reader) (n int64, err error) {
	bufp := copyBufPool.Get().([]byte)
	defer copyBufPool.Put(bufp)
	// context.CanceledのハンドリングするならcopyBufferを実装してそこでやる
	return io.CopyBuffer(dst, src, bufp)
}

func (w *Worker) copyResponse(resp *http.Response) *http.Response {
	if resp == nil {
		return nil
	}

	cpy := new(bytes.Buffer)
	w.copyReader(cpy, resp.Body)

	body := struct {
		io.Writer
		io.ReadCloser
	}{
		cpy,
		ioutil.NopCloser(cpy),
	}

	return &http.Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Header:           resp.Header,
		Body:             body,
		ContentLength:    resp.ContentLength,
		TransferEncoding: resp.TransferEncoding,
		Close:            resp.Close,
		Uncompressed:     resp.Uncompressed,
		Trailer:          resp.Trailer,
		Request:          resp.Request,
		TLS:              resp.TLS,
	}
}

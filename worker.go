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
	ctx     context.Context
	reqChan chan chan Request
	request chan Request
	resChan chan Response
	*http.Client
}

type responseAndError struct {
	Resp *http.Response
	Err error
}

// NewDefaultWorker returns worker pointer having the default http client
func NewDefaultWorker(ctx context.Context, reqChan chan chan Request, resChan chan Response) *Worker {
	return &Worker{
		ctx,
		reqChan,
		make(chan Request, 1),
		resChan,
		http.DefaultClient,
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.reqChan <- w.request

			select {
			case req := <-w.request:
				resp, err := w.Do(req)
				w.resChan <- &responseAndError{w.copyResponse(resp), err}
				resp.Body.Close()
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

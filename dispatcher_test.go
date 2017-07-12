package dispatcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var echo string = "hello"
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	wg := sync.WaitGroup{}
	f := func(resp Response) {
		go func(resp Response) {
			t.Logf("%v", resp.Err)
			dump, _ := httputil.DumpResponse(resp.Resp, true)
			t.Logf("%s", string(dump))
			wg.Done()
		}(resp)
	}

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.Run(ctx)

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", ts.URL, nil)
		d.Add(req)
	}

	wg.Wait()
}

// WIP
func TestRunWithCancel(t *testing.T) {
	var echo string = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Second)
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	f := func(resp Response) {
		go func(resp Response) {
			t.Logf("%v", resp.Err)
			dump, _ := httputil.DumpResponse(resp.Resp, true)
			t.Logf("%s", string(dump))
		}(resp)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.Run(ctx)

	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", ts.URL, nil)
		req = req.WithContext(ctx)
		d.Add(req)
	}
	defer cancel()
	<-ctx.Done()
}

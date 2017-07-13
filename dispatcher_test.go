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
	d := NewDispatcher(context.Background(), 5)
	d.ResultFunc = f
	d.Run()

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", ts.URL, nil)
		d.Add(req)
	}

	wg.Wait()
}

func TestRunWithCancel(t *testing.T) {
	var echo string = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
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
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(5 * time.Second))
	defer cancel()

	d := NewDispatcher(ctx, 5)
	d.ResultFunc = f
	d.Run()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", ts.URL, nil)
		req = req.WithContext(ctx)
		d.Add(req)
	}

	wg.Wait()
}
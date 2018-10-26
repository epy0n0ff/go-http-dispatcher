package dispatcher

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	var echo = "hello"
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
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

	for i := 0; i < 100; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", ts.URL, nil)
		d.Add(req)
	}

	wg.Wait()
}

func TestRunSingleWorker(t *testing.T) {
	var echo = "hello"
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
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

	d := NewDispatcher(1)
	d.ResultFunc = f
	d.Run(ctx)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%d", ts.URL, i), nil)
		d.Add(req)
	}

	wg.Wait()
}

func TestRunWithCancel(t *testing.T) {
	var echo = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	f := func(resp Response) {
		go func(resp Response) {
			if resp.Err != nil {
				t.Logf("err:%v", resp.Err)
			} else {
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
			}
		}(resp)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1*time.Second))

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.Run(ctx)

	go func() {
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", ts.URL, nil)
			req = req.WithContext(ctx)
			d.Add(req)
		}

	}()

	defer cancel()
	<-ctx.Done()
}

func TestRunWithDeadline(t *testing.T) {
	var echo = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	f := func(resp Response) {
		go func(resp Response) {
			if resp.Err != nil {
				t.Logf("err:%v", resp.Err)
			} else {
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
			}
		}(resp)
	}

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.Run(ctx)

	go func() {
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", ts.URL, nil)
			req = req.WithContext(ctx)
			d.Add(req)
		}

	}()
	defer cancel()
	<-ctx.Done()
}

func TestRunWithHttpClient(t *testing.T) {
	var echo = "hello"
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
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

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.RunWithHttpClient(ctx, &http.Client{Transport: tp})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", ts.URL, nil)
		d.Add(req)
	}

	wg.Wait()
}

func TestRunWithCancelAndHttpClient(t *testing.T) {
	var echo = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	f := func(resp Response) {
		go func(resp Response) {
			if resp.Err != nil {
				t.Logf("err:%v", resp.Err)
			} else {
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
			}
		}(resp)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1*time.Second))

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.RunWithHttpClient(ctx, &http.Client{Transport: tp})

	go func() {
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", ts.URL, nil)
			req = req.WithContext(ctx)
			d.Add(req)
		}

	}()

	defer cancel()
	<-ctx.Done()
}

func TestRunWithDeadlineAndHttpClient(t *testing.T) {
	var echo = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	f := func(resp Response) {
		go func(resp Response) {
			if resp.Err != nil {
				t.Logf("err:%v", resp.Err)
			} else {
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
			}
		}(resp)
	}

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	d.ResultFunc = f
	d.RunWithHttpClient(ctx, &http.Client{Transport: tp})
	go func() {
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", ts.URL, nil)
			req = req.WithContext(ctx)
			d.Add(req)
		}

	}()
	defer cancel()
	<-ctx.Done()
}

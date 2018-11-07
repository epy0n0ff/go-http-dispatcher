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

	d := NewDispatcher(5)
	resPool := d.Run(ctx)
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				t.Logf("%v", resp.Err)
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
				wg.Done()
			default:
			}
		}
	}()

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
	d := NewDispatcher(1)
	resPool := d.Run(ctx)
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				t.Logf("%v", resp.Err)
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
				wg.Done()
			default:
			}
		}
	}()

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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1*time.Second))

	d := NewDispatcher(5)
	resPool := d.Run(ctx)
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				if resp.Err != nil {
					t.Logf("err:%v", resp.Err)
				} else {
					dump, _ := httputil.DumpResponse(resp.Resp, true)
					t.Logf("%s", string(dump))
				}
			default:
			}
		}
	}()

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

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))

	d := NewDispatcher(5)
	resPool := d.Run(ctx)
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				if resp.Err != nil {
					t.Logf("err:%v", resp.Err)
				} else {
					dump, _ := httputil.DumpResponse(resp.Resp, true)
					t.Logf("%s", string(dump))
				}
			default:
			}
		}
	}()

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

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	resPool := d.RunWithHttpClient(ctx, &http.Client{Transport: tp})
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				t.Logf("%v", resp.Err)
				dump, _ := httputil.DumpResponse(resp.Resp, true)
				t.Logf("%s", string(dump))
				wg.Done()
			default:
			}
		}
	}()

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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(1*time.Second))

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	resPool := d.RunWithHttpClient(ctx, &http.Client{Transport: tp})
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res
				if resp.Err != nil {
					t.Logf("err:%v", resp.Err)
				} else {
					dump, _ := httputil.DumpResponse(resp.Resp, true)
					t.Logf("%s", string(dump))
				}
			default:
			}
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

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Second))

	tp := http.DefaultTransport.(*http.Transport)
	tp.MaxIdleConns = 0
	tp.MaxIdleConnsPerHost = 100

	d := NewDispatcher(5)
	resPool := d.RunWithHttpClient(ctx, &http.Client{Transport: tp})
	go func() {
		for {
			select {
			case res := <-resPool:
				resp := <-res

				if resp == nil {
					t.Log("resp is nil")
					continue
				}

				if resp.Err != nil {
					t.Logf("err:%v", resp.Err)
				} else {
					dump, _ := httputil.DumpResponse(resp.Resp, true)
					t.Logf("%s", string(dump))
				}
			default:
			}
		}
	}()

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

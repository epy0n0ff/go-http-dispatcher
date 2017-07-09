package dispatcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
	"testing"
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
			t.Logf("%v", resp.err)
			dump, _ := httputil.DumpResponse(resp.res, true)
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

package dispatcher

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopyResponse(t *testing.T) {
	var echo string = "hello"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(echo))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w := NewDefaultWorker(make(chan chan Request), make(chan Response))
	cpy := w.copyResponse(res)
	if cpy == nil {
		t.Fatalf("unexpected error : %v", err)
	}
	raw, err := ioutil.ReadAll(cpy.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if echo != string(raw) {
		t.Fatalf("unexpected error: invalid response body")
	}
}

func BenchmarkCopyReader(b *testing.B) {
	w := NewDefaultWorker(make(chan chan Request), make(chan Response))
	b.ResetTimer()
	r := bytes.NewBufferString("hogehoge")
	for i := 0; i < b.N; i++ {
		w.copyReader(ioutil.Discard, r)
	}
}

func BenchmarkCopyResponse(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	w := NewDefaultWorker(make(chan chan Request), make(chan Response))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.copyResponse(res)
	}
}

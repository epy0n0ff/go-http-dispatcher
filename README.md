# go-http-dispatcher
[![CircleCI](https://circleci.com/gh/epy0n0ff/go-http-dispatcher/tree/master.svg?style=svg)](https://circleci.com/gh/epy0n0ff/go-http-dispatcher/tree/master) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/epy0n0ff/go-http-dispatcher/master/LICENSE)

golang http dispatcher

## Install

```bash
$ go get github.com/epy0n0ff/go-http-dispatcher
```

## Usage

```go
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	f := func(resp dispatcher.Response) {
		go func(resp dispatcher.Response) {
			lock.Lock()
			t.Logf("%v", resp.Err)
			dump, _ := httputil.DumpResponse(resp.Resp, true)
			t.Logf("%s", string(dump))
			lock.Unlock()

			wg.Done()
		}(resp)
	}
	// create five worker threads
	d := dispatcher.NewDispatcher(context.Background(), 5)
	d.ResultFunc = f
	d.Run()

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		req, _ := http.NewRequest("GET", "http://xxxx", nil)
		// enqueue http.Request to workers
		d.Add(req)
	}
	wg.Wait()
```

## License
[Apache 2.0](https://raw.githubusercontent.com/epy0n0ff/go-http-dispatcher/master/LICENSE)

package webgetter

import (
	"context"

	"time"

	"bytes"
	"io/ioutil"

	"github.com/dave/scrapy/scraper/getter"
)

type Getter struct {
	Results map[string]Dummy
}

type Dummy struct {
	Body    string
	Code    int
	Latency time.Duration
	Err     error
}

func (h *Getter) Get(ctx context.Context, url string) chan getter.Result {
	out := make(chan getter.Result)
	go func() {
		defer close(out)

		result, ok := h.Results[url]

		// Not found in mock results: return 404
		if !ok {
			out <- getter.Result{
				Code: 404,
				Body: ioutil.NopCloser(bytes.NewBufferString("404 error body")),
			}
			return
		}

		// Wait for latency but respect cancellation
		select {
		case <-time.After(result.Latency):
			// great!
		case <-ctx.Done():
			out <- getter.Result{Err: ctx.Err()}
			return
		}

		code := 200
		if result.Code > 0 {
			code = result.Code
		}

		// Return the mock result
		out <- getter.Result{
			Code: code,
			Body: ioutil.NopCloser(bytes.NewBufferString(result.Body)),
			Err:  result.Err,
		}
	}()
	return out

}

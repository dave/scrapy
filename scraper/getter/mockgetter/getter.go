// Package mockgetter defines a getter.Interface that returns mock results for use in tests
package mockgetter

import (
	"bytes"
	"context"
	"io/ioutil"
	"time"

	"github.com/dave/scrapy/scraper/getter"
)

// Getter is a getter.Interface that returns results mock results for use in tests
type Getter struct {
	Results map[string]Dummy // The results to return: url -> result
}

// Dummy contains information about the result
type Dummy struct {
	Body    string        // Contents of the body as a string
	Code    int           // Response code
	Latency time.Duration // Time to wait before returning
	Err     error         // Error to return
}

// Get returns a channel. Later it sends the response, and closes the channel.
func (h *Getter) Get(ctx context.Context, url string) chan getter.Result {
	out := make(chan getter.Result)
	go func() {
		// Make sure we close the channel.
		defer close(out)

		// Look up the result in the mock results collection by url.
		result, ok := h.Results[url]

		// If we don't have a result for this URL, return a 404 error.
		if !ok {
			out <- getter.Result{
				Code: 404,
				Body: ioutil.NopCloser(bytes.NewBufferString("404 not found")),
			}
			return
		}

		if result.Latency > 0 {
			// Wait for latency but respect cancellation
			select {
			case <-time.After(result.Latency):
				// great!
			case <-ctx.Done():
				out <- getter.Result{Err: ctx.Err()}
				return
			}
		}

		// Return an error if required
		if result.Err != nil {
			out <- getter.Result{
				Err: result.Err,
			}
			return
		}

		// If code isn't specified, default to 200
		code := 200
		if result.Code > 0 {
			code = result.Code
		}

		// Return the mock result
		out <- getter.Result{
			Code: code,
			Body: ioutil.NopCloser(bytes.NewBufferString(result.Body)),
			Html: true,
		}
	}()
	return out

}

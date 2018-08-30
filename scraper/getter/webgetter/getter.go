// Package webgetter defines a getter.Interface that gets real results by HTTP
package webgetter

import (
	"context"
	"net/http"

	"strings"

	"github.com/dave/scrapy/scraper/getter"
)

// Getter is a getter.Interface that returns results real results by HTTP
type Getter struct {
	client http.Client // the http client to use
}

// Get returns a channel. Later it sends the response, and closes the channel.
func (h *Getter) Get(ctx context.Context, url string) chan getter.Result {
	out := make(chan getter.Result)
	go func() {
		// Make sure we close the channel
		defer close(out)

		// Create a standard GET request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			out <- getter.Result{Err: err}
			return
		}

		// Add the context to the request to ensure we respect cancellation
		req = req.WithContext(ctx)

		// Start the request processing
		response, err := h.client.Do(req)

		select {
		case <-ctx.Done():
			// Was the context cancelled? If so, return the context error.
			// TODO: Is this needed? If the context is cancelled I would think Do will return the context error?
			out <- getter.Result{Err: ctx.Err()}
			return
		default:
			if err != nil {
				out <- getter.Result{Err: err}
				return
			}
			// Send the result on the channel - remember the caller of Get is responsible for closing Body.
			out <- getter.Result{Code: response.StatusCode, Body: response.Body, HTML: strings.Contains(response.Header.Get("content-type"), "text/html")}
			return
		}
	}()
	return out
}

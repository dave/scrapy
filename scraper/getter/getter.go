// Package getter defines an interface that is used to request results by URL
package getter

import (
	"context"
	"io"
)

// Interface is used to request results by URL
type Interface interface {
	Get(ctx context.Context, url string) chan Result // Get returns a channel. Later it sends the response, and closes the channel.
}

// Result is the result of a Get
type Result struct {
	Code int           // The http status code
	Body io.ReadCloser // The body - remember the caller of Get is responsible for closing this.
	HTML bool          // Did the content-type header indicates HTML?
	Err  error         // Any error (all other fields will be zero if Err != nil)
}

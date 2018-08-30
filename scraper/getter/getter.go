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

// Interface is used to request results by URL
type Result struct {
	Code int           // The http status code
	Body io.ReadCloser // The body
	Html bool          // The content-type header indicates HTML
	Err  error         // Any error (all other fields will be zero if Err != nil)
}

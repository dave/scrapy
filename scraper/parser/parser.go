// Package parser defines an interface used to parse HTML and extract links
package parser

import (
	"context"
	"io"
)

// Interface parses HTML and returns the urls from anchor href attributes
type Interface interface {
	// Parse parses the document and returns the urls and parse errors
	Parse(ctx context.Context, url string, body io.Reader) (urls []string, errs []error)
}

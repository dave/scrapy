// Package parser defines an interface used to parse HTML and extract links
package parser

import (
	"io"
)

type Interface interface {
	Parse(url string, body io.Reader) (urls []string, errs []error)
}

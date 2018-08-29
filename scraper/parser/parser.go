package parser

import (
	"io"
)

type Interface interface {
	Parse(url string, body io.Reader) (urls []string, errs []error)
}

package parser

import (
	"io"
)

type Interface interface {
	Parse(reader io.ReadCloser) (urls []string, errs []error)
}

package parser

import (
	"io"
)

type Interface interface {
	Parse(reader io.Reader) (urls []string, errs []error)
}

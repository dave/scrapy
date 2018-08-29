package getter

import (
	"context"
	"io"
)

type Interface interface {
	Get(ctx context.Context, url string) chan Result
}

type Result struct {
	Code int
	Body io.ReadCloser
	Err  error
	Mime string
}

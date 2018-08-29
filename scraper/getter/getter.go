package getter

import (
	"context"
	"io"
)

type Interface interface {
	GetPage(ctx context.Context, url string) chan Result
}

type Result struct {
	Code int
	Body io.ReadCloser
	Err  error
}

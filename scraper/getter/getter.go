package getter

import "context"

type Interface interface {
	GetPage(ctx context.Context, url string) chan Result
}

type Result struct {
	Code int
	Body []byte
	Err  error
}

package webgetter

import (
	"context"
	"net/http"

	"github.com/dave/scrapy/scraper/getter"
)

type Getter struct {
	client http.Client
}

func (h *Getter) Get(ctx context.Context, url string) chan getter.Result {
	out := make(chan getter.Result)
	go func() {
		defer close(out)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			out <- getter.Result{Err: err}
			return
		}
		response, err := h.client.Do(req.WithContext(ctx))
		select {
		case <-ctx.Done():
			out <- getter.Result{Err: ctx.Err()}
			return
		default:
			if err != nil {
				out <- getter.Result{Err: err}
				return
			}
			out <- getter.Result{Code: response.StatusCode, Body: response.Body}
			return
		}
	}()
	return out
}

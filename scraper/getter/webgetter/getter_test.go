package webgetter

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dave/scrapy/scraper/getter"
)

func TestGetter(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		code        int
		err         string
		overrideURL string
		returnAfter int
		cancelAfter int
	}{
		{
			name: "simple",
			body: "a\n",
			code: 200,
		},
		{
			name:        "broken url",
			overrideURL: "a",
			err:         "unsupported protocol scheme",
		},
		{
			name: "code 404",
			body: "a\n",
			code: 404,
		},
		{
			name:        "timeout",
			returnAfter: 100,
			cancelAfter: 5,
			err:         "context deadline exceeded",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := server(test.returnAfter, test.code, test.body)
			defer ts.Close()

			g := &Getter{}

			ctx := context.Background()
			if test.cancelAfter > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(test.cancelAfter)*time.Millisecond)
				defer cancel()
			}

			url := ts.URL
			if test.overrideURL != "" {
				url = test.overrideURL
			}

			c := g.Get(ctx, url)

			var r getter.Result
			select {
			case r = <-c:
				// great.
			case <-time.After(time.Millisecond * 200):
				// fatal time out after 200ms
				t.Fatal("test took too long")
			}

			var body string
			if r.Body != nil {
				b, err := ioutil.ReadAll(r.Body)
				r.Body.Close()
				if err != nil {
					t.Errorf("error reading body: %v", err)
				}
				body = string(b)
			}

			if body != test.body {
				t.Errorf("expected body %q, got %q", test.body, body)
			}

			if r.Code != test.code {
				t.Errorf("expected code %d, got %d", test.code, r.Code)
			}

			if test.err == "" {
				if r.Err != nil {
					t.Errorf("expected success, got error: %v", r.Err)
				}
			} else {
				if r.Err == nil {
					t.Errorf("expected error %s, got nil", test.err)
				} else if !strings.Contains(r.Err.Error(), test.err) {
					t.Errorf("expected error to contain %s, got %v", test.err, r.Err)
				}
			}
		})
	}
}

func server(returnAfter, code int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if returnAfter > 0 {
			<-time.After(time.Duration(returnAfter) * time.Millisecond)
		}
		if code != 0 && code != 200 {
			w.WriteHeader(code)
		}
		fmt.Fprint(w, body)
	}))
}

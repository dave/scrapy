package scraper

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dave/scrapy/scraper/getter"
	"github.com/dave/scrapy/scraper/logger"
	"github.com/dave/scrapy/scraper/parser"
	"github.com/dave/scrapy/scraper/queuer"
)

type State struct {
	Timeout time.Duration
	Getter  getter.Interface
	Parser  parser.Interface
	Queuer  queuer.Interface
	Logger  logger.Interface
}

func (s *State) Start(ctx context.Context, url string) {
	s.Logger.Init()
	s.Queuer.Start(func(url string) {

		ctx, cancel := context.WithTimeout(ctx, s.Timeout)
		defer cancel()

		// Log that the url has started processing
		s.Logger.Start(url)

		start := time.Now()

		// Start the getter
		c := s.Getter.Get(ctx, url)

		// Wait for the getter to start streaming the contents, but respect context cancellation
		var body io.ReadCloser
		var code int
		var err error
		var mime string
		select {
		case <-ctx.Done():
			s.Logger.Error(url, ctx.Err())
			return
		case r := <-c:
			err = r.Err
			body = r.Body
			code = r.Code
			mime = r.Mime
		}

		// Log error
		if err != nil {
			s.Logger.Error(url, err)
			return
		}

		if !strings.Contains(mime, "text/html") {
			s.Logger.Error(url, fmt.Errorf("unsupported mime type: %s", mime))
			return
		}

		// Close body if it is non nil
		if body != nil {
			defer body.Close()
		}

		// Don't continue if the code is not 200
		if code != 200 {
			s.Logger.Finish(url, code, time.Now().Sub(start), 0, 0)
			return
		}

		// Parse the body
		urls, errs := s.Parser.Parse(url, body)

		// Perhaps the parser ended early because of cancellation? If so, log the error.
		select {
		case <-ctx.Done():
			s.Logger.Error(url, ctx.Err())
			return
		default:
			// great!
		}

		// Log the finish event
		s.Logger.Finish(url, code, time.Now().Sub(start), len(urls), len(errs))

		// Queue all the resulting urls
		for _, u := range urls {
			if added, err := s.Queuer.Push(u); err != nil {
				s.Logger.Full(u)
			} else if added {
				s.Logger.Queue(u)
			}
		}
	})

	if added, err := s.Queuer.Push(url); err != nil {
		s.Logger.Full(url)
	} else if added {
		s.Logger.Queue(url)
	}

	s.Queuer.Wait()
	s.Logger.Exit()
}

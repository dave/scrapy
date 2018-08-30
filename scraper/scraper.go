package scraper

import (
	"context"
	"errors"
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
		s.Logger.Starting(url)

		start := time.Now()

		// Start the getter
		c := s.Getter.Get(ctx, url)

		// Wait for the getter to start streaming the contents, but respect context cancellation
		var r getter.Result
		select {
		case <-ctx.Done():
			s.Logger.Error(url, ctx.Err())
			return
		case r = <-c:
			// great!
		}

		// Log error
		if r.Err != nil {
			s.Logger.Error(url, r.Err)
			return
		}

		// Close body if it is non nil
		if r.Body != nil {
			defer r.Body.Close()
		}

		// Don't continue if the code is not 200
		if r.Code != 200 {
			s.Logger.Finished(url, r.Code, time.Now().Sub(start), 0, 0)
			return
		}

		if !r.Html {
			s.Logger.Error(url, errors.New("contents were not HTML"))
			return
		}

		// Parse the body
		urls, errs := s.Parser.Parse(url, r.Body)

		// Perhaps the parser ended early because of cancellation? If so, log the error.
		select {
		case <-ctx.Done():
			s.Logger.Error(url, ctx.Err())
			return
		default:
			// great!
		}

		// Log the finish event
		s.Logger.Finished(url, r.Code, time.Now().Sub(start), len(urls), len(errs))

		// Queue all the resulting urls
		for _, u := range urls {
			if err := s.Queuer.Push(u); err != nil {
				s.Logger.Error(u, err)
				continue
			}
			// Log if the push succeeded
			s.Logger.Queued(u)
		}
	})

	if err := s.Queuer.Push(url); err != nil {
		panic("error in initial queue push")
	}
	s.Logger.Queued(url)

	s.Queuer.Wait()
	s.Logger.Exit()
}

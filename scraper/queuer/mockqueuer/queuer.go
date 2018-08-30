package mockqueuer

import (
	"context"
	"sync"

	"github.com/dave/scrapy/scraper/queuer"
)

// Mock queuer executes the action as soon as the url is pushed
// Deprecated: using the concurrent queuer with Workers: 1 gives reproducible results and correct log order

type Queuer struct {
	action func(url string)
	urls   map[string]bool
	ctx    context.Context
	once   sync.Once
}

func (q *Queuer) Start(action func(url string)) {
	q.once.Do(func() {
		q.urls = map[string]bool{}
	})
	q.action = action
}

func (q *Queuer) Push(url string) error {
	if q.urls == nil {
		panic("Start must be called before Push")
	}
	if q.urls[url] {
		return queuer.DuplicateError
	}
	q.urls[url] = true
	q.action(url)
	return nil
}

func (*Queuer) Wait() {
	// no-op
}

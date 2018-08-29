package mockqueuer

import "context"

// Mock queuer executes the action as soon as the url is pushed

type Queuer struct {
	action func(ctx context.Context, url string)
	urls   map[string]bool
	ctx    context.Context
}

func (q *Queuer) Push(url string) {
	if q.urls == nil {
		q.urls = map[string]bool{}
	}
	if q.urls[url] {
		return
	}
	q.urls[url] = true
	q.action(q.ctx, url)
}

func (q *Queuer) Action(action func(ctx context.Context, url string)) {
	q.action = action
}

func (q *Queuer) Start(ctx context.Context) {
	q.ctx = ctx
}

func (*Queuer) Wait() {}

package concurrentqueuer

import (
	"sync"

	"github.com/dave/scrapy/scraper/queuer"
)

type Queuer struct {
	Length  int // Max queue length
	Workers int // Number of concurrent workers

	seen        sync.Map       // Tracks the urls that have been pushed in the past
	queue       chan string    // The queue
	wg          sync.WaitGroup // Waitgroup tracking progress
	initialised bool           // Ensures initialisation orderss
}

func (q *Queuer) Start(action func(url string)) {

	if q.initialised {
		panic("Start should only be called once")
	}

	q.initialised = true
	q.queue = make(chan string, q.Length)

	for i := 0; i < q.Workers; i++ {
		go func() {
			for u := range q.queue {
				action(u)
				q.wg.Done()
			}
		}()
	}
}

func (q *Queuer) Push(url string) error {

	if !q.initialised {
		panic("Start must be called before Push")
	}

	if _, loaded := q.seen.LoadOrStore(url, true); loaded {
		return queuer.DuplicateError
	}

	select {
	case q.queue <- url:
		// Url was added to the queue
		q.wg.Add(1)
		return nil
	default:
		// queue was full - don't want to wait here...
		return queuer.FullError
	}

}

func (q *Queuer) Wait() {
	q.wg.Wait()
}

package concurrentqueuer

import (
	"sync"

	"github.com/dave/scrapy/scraper/queuer"
)

type Queuer struct {
	Length  int // Max queue length
	Workers int // Number of concurrent workers

	urls  map[string]bool // Urls that have been processed in the past
	m     sync.RWMutex    // To protect the urls map
	once  sync.Once       // For initialisation
	queue chan string     // The queue
	wg    *sync.WaitGroup
}

func (q *Queuer) Start(action func(url string)) {
	q.once.Do(func() {
		q.urls = map[string]bool{}
		q.queue = make(chan string, q.Length)
		q.wg = &sync.WaitGroup{}
	})
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

	if q.urls == nil {
		panic("Start must be called before Push")
	}

	if !q.needsProcessing(url) {
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

func (q *Queuer) needsProcessing(url string) bool {

	// First do a RLock
	q.m.RLock()
	if q.urls[url] {
		q.m.RUnlock()
		return false
	}
	q.m.RUnlock()

	// If url wasn't found, do a Lock
	q.m.Lock()
	defer q.m.Unlock()

	// We need to check again after the Lock
	if q.urls[url] {
		return false
	}

	// Finally set the url
	q.urls[url] = true
	return true
}

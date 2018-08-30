package concurrentqueuer

import (
	"sync"

	"github.com/dave/scrapy/scraper/queuer"
)

type Queuer struct {
	Length  int // Max queue length
	Workers int // Number of concurrent workers

	seen                  sync.Map       // Tracks the items that have been pushed in the past
	queue                 chan string    // The queue of items waiting to process
	queueWait, workerWait sync.WaitGroup // Waitgroup tracking queue and workers
	initialised           bool           // Ensures initialisation order
}

func (q *Queuer) Start(action func(string)) {

	if q.initialised {
		panic("Start should only be called once")
	}

	q.initialised = true
	q.queue = make(chan string, q.Length)

	for i := 0; i < q.Workers; i++ {
		go func() {

			// Use a waitgroup to ensure we don't exit before the workers have finished exiting.
			q.workerWait.Add(1)
			defer q.workerWait.Done()

			// Read from the queue channel and perform the action on each item
			for u := range q.queue {
				action(u)
				q.queueWait.Done()
			}
		}()
	}
}

func (q *Queuer) Push(payload string) error {

	if !q.initialised {
		panic("Start must be called before Push")
	}

	if _, loaded := q.seen.LoadOrStore(payload, true); loaded {
		return queuer.DuplicateError
	}

	select {
	case q.queue <- payload:
		// Url was added to the queue
		q.queueWait.Add(1)
		return nil
	default:
		// queue was full - don't want to wait here...
		return queuer.FullError
	}

}

func (q *Queuer) Wait() {
	q.queueWait.Wait()  // wait for the queue to finish
	close(q.queue)      // close the queue channel so workers will start to exit
	q.workerWait.Wait() // wait for all workers to finish exiting
}

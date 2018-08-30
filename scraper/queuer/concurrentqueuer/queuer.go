// Package concurrentqueuer defines a queuer.Interface that runs several workers concurrently on a queue
package concurrentqueuer

import (
	"sync"

	"github.com/dave/scrapy/scraper/queuer"
)

// Queuer is a queuer.Interface that runs several workers concurrently on a queue.
type Queuer struct {
	Length                int            // Max queue length
	Workers               int            // Number of concurrent workers
	seen                  sync.Map       // Tracks the items that have been pushed in the past
	queue                 chan string    // The queue of items waiting to process
	queueWait, workerWait sync.WaitGroup // Waitgroup tracking queue and workers
	once                  sync.Once      // For initialisation
}

// Start starts processing the queue.
func (q *Queuer) Start(action func(string)) {

	q.ensureInitialised()

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

// Push attempts to add an item to the queue. On failure, returns queuer.DuplicateError or queuer.FullError.
func (q *Queuer) Push(payload string) error {

	q.ensureInitialised()

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

// Wait waits for all items to be processed before returning.
func (q *Queuer) Wait() {
	q.queueWait.Wait()  // wait for the queue to finish
	close(q.queue)      // close the queue channel so workers will start to exit
	q.workerWait.Wait() // wait for all workers to finish exiting
}

// initialises the queue
func (q *Queuer) ensureInitialised() {
	q.once.Do(func() {
		q.queue = make(chan string, q.Length)
	})
}

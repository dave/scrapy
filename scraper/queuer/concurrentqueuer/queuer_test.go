package concurrentqueuer

import (
	"testing"
	"time"

	"github.com/dave/scrapy/scraper/queuer"
)

// TestQueuerConcurrent tests that a queuer with several workers starts work on the correct number of concurrent actions
func TestQueuer_concurrent(t *testing.T) {
	q := &Queuer{Length: 10, Workers: 2}

	aSignal := make(chan struct{})
	aStarted := make(chan struct{})
	bSignal := make(chan struct{})
	bStarted := make(chan struct{})
	cSignal := make(chan struct{})
	cStarted := make(chan struct{})

	q.Start(func(s string) {
		switch s {
		case "a":
			close(aStarted)
			<-aSignal
		case "b":
			close(bStarted)
			<-bSignal
		case "c":
			close(cStarted)
			<-cSignal
		}
	})

	// Push two items onto the queue
	if err := q.Push("a"); err != nil {
		t.Errorf("a should succeed, this failed with %v", err)
	}

	if err := q.Push("b"); err != nil {
		t.Errorf("b should succeed, this failed with %v", err)
	}

	// Wait for the queue to start processing both actions
	if timeout(aStarted) {
		t.Errorf("timed out waiting for a to start processing")
	}
	if timeout(bStarted) {
		t.Errorf("timed out waiting for b to start processing")
	}

	// A third action will be queued, but should not start processing until one of the previous actions finishes.
	if err := q.Push("c"); err != nil {
		t.Errorf("c should succeed, this failed with %v", err)
	}
	if !timeout(cStarted) {
		t.Errorf("c should not start processing, but it did")
	}

	// Finish a and b so can start
	close(aSignal)

	// Wait for the queue to start processing c
	if timeout(cStarted) {
		t.Errorf("timed out waiting for c to start processing")
	}

	// Tidy up
	close(bSignal)
	close(cSignal)

	q.Wait()

}

// TestQueuer tests that the queueing system works correctly and FullError, DuplicateErrors are returned correctly
func TestQueuer_queue(t *testing.T) {
	q := &Queuer{Length: 1, Workers: 1}

	aSignal := make(chan struct{})
	aStarted := make(chan struct{})
	bSignal := make(chan struct{})
	bStarted := make(chan struct{})

	q.Start(func(s string) {
		switch s {
		case "a":
			close(aStarted)
			<-aSignal
		case "b":
			close(bStarted)
			<-bSignal
		}
	})

	if err := q.Push("a"); err != nil {
		t.Errorf("a should succeed, this failed with %v", err)
	}

	// Wait for the queue to start processing a, but it won't finish until we close aSignal
	if timeout(aStarted) {
		t.Errorf("timed out waiting for a to start processing")
	}

	// We push another item. The queue is max length 1, so this will fill the queue
	if err := q.Push("b"); err != nil {
		t.Errorf("b should succeed, this failed with %v", err)
	}

	// Pushing another item should fail with a full queue
	if err := q.Push("c"); err != queuer.FullError {
		t.Errorf("c should fail with FullError, this failed with %v", err)
	}

	// Finally we let a finish by closing aSignal
	close(aSignal)

	// b should start processing
	if timeout(bStarted) {
		t.Errorf("timed out waiting for b to start processing")
	}

	// Now if we re-push any of the three items we should get DuplicateError
	if err := q.Push("a"); err != queuer.DuplicateError {
		t.Errorf("a should now fail with DuplicateError, this failed with %v", err)
	}
	if err := q.Push("b"); err != queuer.DuplicateError {
		t.Errorf("b should now fail with DuplicateError, this failed with %v", err)
	}
	if err := q.Push("c"); err != queuer.DuplicateError {
		t.Errorf("c should now fail with DuplicateError, this failed with %v", err)
	}

	// Tidy up
	close(bSignal)

	q.Wait()

}

func timeout(c chan struct{}) bool {
	select {
	case <-c:
		return false
	case <-time.After(time.Millisecond * 200):
		// TODO: Arbitrary timeouts in tests are brittle - is there another way?
		return true
	}
}

// Package queuer defines an interface used to queue and execute an action on items
package queuer

import "errors"

// Interface is used to queue and execute an action on items
type Interface interface {
	Start(action func(string)) // Start starts processing the queue.
	Push(item string) error    // Push attempts to add an item to the queue. On failure, returns ErrDuplicate or ErrFull.
	Wait()                     // Wait waits for all items to be processed before returning.
}

// ErrDuplicate is returned by Push when the URL has been pushed before
var ErrDuplicate = errors.New("duplicate url")

// ErrFull is returned by Push when the queue is full
var ErrFull = errors.New("queue full")

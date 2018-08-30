// Package queuer defines an interface used to queue and execute an action on items
package queuer

import "errors"

// Interface is used to queue and execute an action on items
type Interface interface {
	Start(action func(string)) // Start starts processing the queue.
	Push(item string) error    // Push attempts to add an item to the queue. On failure, returns DuplicateError or FullError.
	Wait()                     // Wait waits for all items to be processed before returning.
}

// DuplicateError is returned by Push when the URL has been pushed before
var DuplicateError = errors.New("duplicate url")

// FullError is returned by Push when the queue is full
var FullError = errors.New("queue full")

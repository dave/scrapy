package queuer

import "errors"

type Interface interface {
	Start(func(string)) // Starts processing the queue.
	Push(string) error  // Attempt to add a url to the queue. On failure, returns DuplicateError or FullError.
	Wait()              // Waits for all items to be processed before returning.
}

var DuplicateError = errors.New("duplicate url")
var FullError = errors.New("queue full")

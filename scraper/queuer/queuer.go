package queuer

type Interface interface {
	Start(func(string))        // Starts processing the queue.
	Push(string) (bool, error) // Add a url to the queue. Returns false if the url was a duplicate.
	Wait()                     // Waits for all items to be processed before returning.
}

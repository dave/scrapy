package queuer

type Interface interface {
	Start(func(string))        // Starts processing the queue.
	Push(string) (bool, error) // Attempt to add a url to the queue. Returns true if the url was added (false if duplicate).
	Wait()                     // Waits for all items to be processed before returning.
}

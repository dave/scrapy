package queuer

import "context"

type Interface interface {
	Push(url string)                              // Add a url to the queue - push maintains a list of the previously added urls and does automatic de-duplication
	Action(func(ctx context.Context, url string)) // Sets the action to perform for each url
	Start(ctx context.Context)                    // Starts the queue processing and returns immediately
	Wait()                                        // Waits for all items to be processed before returning
}

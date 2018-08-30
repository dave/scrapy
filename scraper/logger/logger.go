// Package logger defines an interface that is used to log events and metrics during execution
package logger

import "time"

// Interface is used to log events and metrics during execution
type Interface interface {
	Init()               // Initialise the logger
	Queued(url string)   // Queued is called each time a url is successfully queued
	Starting(url string) // Starting is called each time a url starts processing
	Finished(url string, code int, latency time.Duration,
		urls, errors int) // Finished is called each time a URL successfully finishes processing (even for non-200 results)
	Error(url string, err error) // Error is called on every error
	Exit()                       // Exit is called when the queue has finished and the logger should finalise
}

// Package logger defines an interface that is used to log events and metrics during execution
package logger

import "time"

type Interface interface {
	Init()
	Queued(url string)
	Starting(url string)
	Finished(url string, code int, latency time.Duration, urls, errors int)
	Error(url string, err error)
	Exit()
}

package logger

import "time"

type Interface interface {
	Init()
	Queue(url string)
	Start(url string)
	Full(url string)
	Error(url string, err error)
	Finish(url string, code int, latency time.Duration, urls, errors int)
	Exit()
}

// Package mocklogger defines a logger.Interface that stores a string representation of each logged event for testing
package mocklogger

import (
	"fmt"
	"sync"
	"time"
)

// Logger is a logger.Interface that stores a string representation of each logged event for testing
type Logger struct {
	Log []string
	m   sync.Mutex
}

// Init initialises the logger
func (l *Logger) Init() {}

// Queued is called each time a url is successfully queued
func (l *Logger) Queued(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("queue %s", url))
}

// Starting is called each time a url starts processing
func (l *Logger) Starting(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("start %s", url))
}

// Finished is called each time a URL successfully finishes processing (even for non-200 results)
func (l *Logger) Finished(url string, code int, latency time.Duration, urls, errors int) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("finish %s: %d, %d, %d", url, code, urls, errors))
}

// Error is called on every error
func (l *Logger) Error(url string, err error) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("error %s: %v", url, err))
}

// Exit is called when the queue has finished and the logger should finalise
func (l *Logger) Exit() {}

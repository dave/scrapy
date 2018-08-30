package mocklogger

import (
	"fmt"
	"sync"
	"time"
)

type Logger struct {
	Log []string
	m   sync.Mutex
}

func (l *Logger) Init() {}

func (l *Logger) Queued(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("queue %s", url))
}

func (l *Logger) Starting(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("start %s", url))
}

func (l *Logger) Finished(url string, code int, latency time.Duration, urls, errors int) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("finish %s: %d, %d, %d", url, code, urls, errors))
}

func (l *Logger) Error(url string, err error) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("error %s: %v", url, err))
}

func (l *Logger) Exit() {}

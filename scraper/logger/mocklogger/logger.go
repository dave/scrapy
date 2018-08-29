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

func (l *Logger) Start(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("start %s", url))
}

func (l *Logger) Error(url string, err error) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("error %s: %v", url, err))
}

func (l *Logger) Finish(url string, code int, latency time.Duration, errors int) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Log = append(l.Log, fmt.Sprintf("finish %s: %d, %d", url, code, errors))
}

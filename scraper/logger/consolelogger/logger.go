package consolelogger

import (
	"fmt"
	"io"
	"time"
)

type Logger struct {
	Writer io.Writer
}

func (l *Logger) Start(url string) {
	fmt.Fprintf(l.Writer, "starting %s\n", url)
}

func (l *Logger) Error(url string, err error) {
	fmt.Fprintf(l.Writer, "error %s: %v\n", url, err)
}

func (l *Logger) Finish(url string, code int, latency time.Duration, urls, errors int) {
	fmt.Fprintf(l.Writer, "finished %s: code %d in %v with %d url(s) and %d error(s)\n", url, code, latency, urls, errors)
}

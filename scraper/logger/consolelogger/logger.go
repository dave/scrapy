package consolelogger

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"strings"

	"github.com/buger/goterm"
)

type Logger struct {
	Urls                           []string
	urlM                           sync.Mutex
	err                            error
	errM                           sync.Mutex
	queued, started, errs, success uint64
	ticker                         *time.Ticker
	ticking                        bool
}

func (l *Logger) log() {

	goterm.Clear()
	goterm.MoveCursor(1, 1)

	queued := atomic.LoadUint64(&l.queued)
	started := atomic.LoadUint64(&l.started)
	errs := atomic.LoadUint64(&l.errs)
	success := atomic.LoadUint64(&l.success)

	totals := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprintf(totals, "Queued\t%d\n", queued-started)
	fmt.Fprintf(totals, "Started\t%d\n", started)
	fmt.Fprintf(totals, "Errors\t%d\n", errs)
	fmt.Fprintf(totals, "Success\t%d\n", success)
	//fmt.Fprintf(totals, "NumGoroutine\t%d\n", runtime.NumGoroutine())
	if lastErr := l.getErr(); lastErr != "" {
		fmt.Fprintf(totals, "Last error\t%s\n", lastErr)
	}
	goterm.Println(totals)
	goterm.Flush()

}

func (l *Logger) Init() {
	l.ticker = time.NewTicker(200 * time.Millisecond)
	l.ticking = true
	go func() {
		for range l.ticker.C {
			if l.ticking {
				l.log()
			}
		}
	}()
}

func (l *Logger) getErr() string {
	l.errM.Lock()
	defer l.errM.Unlock()
	if l.err == nil {
		return ""
	}
	return strings.TrimSpace(l.err.Error())
}

func (l *Logger) addUrl(url string) {
	l.urlM.Lock()
	defer l.urlM.Unlock()
	l.Urls = append(l.Urls, url)
}

func (l *Logger) Queue(url string) {
	atomic.AddUint64(&l.queued, 1)
}

func (l *Logger) Start(url string) {
	atomic.AddUint64(&l.started, 1)
}

func (l *Logger) Error(url string, err error) {
	atomic.AddUint64(&l.errs, 1)
	if !strings.Contains(err.Error(), "context canceled") {
		l.errM.Lock()
		l.err = err
		l.errM.Unlock()
	}
}

func (l *Logger) Finish(url string, code int, latency time.Duration, urls, errors int) {
	if code != 200 {
		atomic.AddUint64(&l.errs, 1)
		return
	}
	atomic.AddUint64(&l.success, 1)
	l.addUrl(url)
}

func (l *Logger) Exit() {
	l.ticking = false
	l.ticker.Stop()
	sort.Strings(l.Urls)
	goterm.Clear()
	goterm.MoveCursor(1, 1)
	goterm.Print("\r")
	for _, u := range l.Urls {
		goterm.Println(u)
	}
	goterm.Flush()
}

package consolelogger

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/dave/ghistogram"
)

type Logger struct {
	Urls                                 []string
	urlM                                 sync.Mutex
	err                                  error
	errM                                 sync.Mutex
	last                                 string
	lastM                                sync.Mutex
	queued, started, errs, success, full uint64
	ticker                               *time.Ticker
	ticking                              bool
	hist                                 *ghistogram.Histogram
}

func (l *Logger) log() {

	queued := atomic.LoadUint64(&l.queued)
	started := atomic.LoadUint64(&l.started)
	errs := atomic.LoadUint64(&l.errs)
	success := atomic.LoadUint64(&l.success)
	full := atomic.LoadUint64(&l.full)
	lastErr := l.getErr()
	lastUrl := l.getLast()

	fmt.Print("\033[H\033[2J")
	fmt.Println("Summary")
	fmt.Println("-------")
	totals := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(totals, "Queued\t%d\n", queued-started)
	fmt.Fprintf(totals, "In progress\t%d\n", started-success-errs)
	fmt.Fprintf(totals, "Success\t%d\n", success)
	fmt.Fprintf(totals, "Errors\t%d\n", errs)
	if lastUrl != "" {
		fmt.Fprintf(totals, "Last url\t%s\n", lastUrl)
	}
	if full > 0 {
		fmt.Fprintf(totals, "Queue was full\t%d\n", full)
	}
	if lastErr != "" {
		fmt.Fprintf(totals, "Last error\t%s\n", lastErr)
	}
	totals.Flush()

	if false {
		fmt.Println("")
		fmt.Println("Mem stats")
		fmt.Println("---------")
		stats := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		bToMb := func(b uint64) uint64 {
			return b / 1024 / 1024
		}
		fmt.Fprintf(stats, "Alloc\t%v MiB\n", bToMb(m.Alloc))
		fmt.Fprintf(stats, "TotalAlloc\t%v MiB\n", bToMb(m.TotalAlloc))
		fmt.Fprintf(stats, "Sys\t%v MiB\n", bToMb(m.Sys))
		fmt.Fprintf(stats, "Goroutines\t%d\n", runtime.NumGoroutine())
		stats.Flush()
	}

	fmt.Println("")
	fmt.Println("Latency")
	fmt.Println("-------")
	g := l.hist.EmitGraph(nil, nil)
	fmt.Println(g.String())

}

func (l *Logger) Init() {
	l.ticker = time.NewTicker(200 * time.Millisecond)
	l.ticking = true
	l.hist = ghistogram.NewHistogram(21, 100, 0.0)
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

func (l *Logger) getLast() string {
	l.lastM.Lock()
	defer l.lastM.Unlock()
	return l.last
}

func (l *Logger) addUrl(url string) {
	l.urlM.Lock()
	defer l.urlM.Unlock()
	l.Urls = append(l.Urls, url)
}

func (l *Logger) Full(url string) {
	atomic.AddUint64(&l.full, 1)
}

func (l *Logger) Queue(url string) {
	atomic.AddUint64(&l.queued, 1)
}

func (l *Logger) Start(url string) {
	atomic.AddUint64(&l.started, 1)
	l.lastM.Lock()
	l.last = url
	l.lastM.Unlock()
}

func (l *Logger) Error(url string, err error) {
	atomic.AddUint64(&l.errs, 1)
	l.errM.Lock()
	l.err = err
	l.errM.Unlock()
}

func (l *Logger) Finish(url string, code int, latency time.Duration, urls, errors int) {
	l.hist.Add(uint64(latency/time.Millisecond), 1)
	if code != 200 {
		l.errM.Lock()
		l.err = fmt.Errorf("response code %d: %s", code, url)
		l.errM.Unlock()
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

	fmt.Println("URLs")
	fmt.Println("----")
	for _, u := range l.Urls {
		fmt.Println(u)
	}
}

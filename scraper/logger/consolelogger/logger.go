package consolelogger

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/dave/ghistogram"
	"github.com/dave/scrapy/scraper/queuer"
)

type Logger struct {
	successfulUrls                       []string              // all successful urls (will be sorted and listed at exit)
	lastUrlStarted                       string                // last url that started processing
	lastErr                              error                 // last error received
	queued, started, errs, success, full uint64                // counters for various stats
	ticker                               *time.Ticker          // ticker ticks every 200ms to display stats
	ticking                              bool                  // used to ensure stats don't display after ticker is stopped
	hist                                 *ghistogram.Histogram // displays a histogram of latencies
	m                                    sync.Mutex            // If ultimate performance was a concern, we could have a mutex per variable but this will simplify
}

func (l *Logger) log() {

	stats := l.loadDisplayStats()

	fmt.Print(ClearScreen)
	fmt.Println("Summary")
	fmt.Println("-------")

	w := tabwriter.NewWriter(os.Stdout, 4, 3, 3, ' ', 0)
	fmt.Fprintf(w, "Queued\t%d\n", stats.inQueue)
	fmt.Fprintf(w, "In progress\t%d\t%s\n", stats.inProgress, l.getLastUrlStarted())
	fmt.Fprintf(w, "Success\t%d\n", stats.success)
	fmt.Fprintf(w, "Errors\t%d\t%s\n", stats.allErrors, l.getLastErr())
	w.Flush()

	// printMemStats()

	fmt.Println("")
	fmt.Println("Latency")
	fmt.Println("-------")
	fmt.Println(l.hist.EmitGraph(nil, nil).String())

}

func (l *Logger) Init() {
	l.ticker = time.NewTicker(200 * time.Millisecond)
	l.ticking = true
	l.hist = ghistogram.NewHistogram(21, 100, 0.0)
	go func() {
		for range l.ticker.C {
			if l.isTicking() {
				l.log()
			}
		}
	}()
}

func (l *Logger) Queued(url string) {
	atomic.AddUint64(&l.queued, 1)
}

func (l *Logger) Starting(url string) {
	atomic.AddUint64(&l.started, 1)
	l.setLastUrlStarted(url)
}

func (l *Logger) Error(url string, err error) {

	// TODO: add latency here and log to the histogram

	switch err {
	case queuer.DuplicateError:
		// ignore duplicate errors
	case queuer.FullError:
		atomic.AddUint64(&l.full, 1)
		l.setLastErr(err)
	default:
		atomic.AddUint64(&l.errs, 1)
		l.setLastErr(err)
	}
}

func (l *Logger) Finished(url string, code int, latency time.Duration, urls, errors int) {

	// Log the latency for all finished requests for the histogram
	l.hist.Add(uint64(latency/time.Millisecond), 1)

	// If the code isn't 200, log as an error
	if code != 200 {
		atomic.AddUint64(&l.errs, 1)
		l.setLastErr(fmt.Errorf("response code %d: %s", code, url))
		return
	}

	atomic.AddUint64(&l.success, 1)
	l.addUrlSuccess(url)
}

func (l *Logger) Exit() {

	l.stopTicker()

	// Sort the strings
	sort.Strings(l.successfulUrls)

	fmt.Println("URLs")
	fmt.Println("----")
	for _, u := range l.successfulUrls {
		fmt.Println(u)
	}
}

func (l *Logger) isTicking() bool {
	l.m.Lock()
	defer l.m.Unlock()
	return l.ticking
}

func (l *Logger) stopTicker() {
	l.ticker.Stop()
	l.m.Lock()
	defer l.m.Unlock()
	l.ticking = false
}

func (l *Logger) getLastErr() string {
	l.m.Lock()
	defer l.m.Unlock()
	if l.lastErr == nil {
		return ""
	}
	return strings.TrimSpace(l.lastErr.Error())
}

func (l *Logger) setLastErr(err error) {
	l.m.Lock()
	defer l.m.Unlock()
	l.lastErr = err
}

func (l *Logger) getLastUrlStarted() string {
	l.m.Lock()
	defer l.m.Unlock()
	return l.lastUrlStarted
}

func (l *Logger) setLastUrlStarted(u string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.lastUrlStarted = u
}

func (l *Logger) addUrlSuccess(url string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.successfulUrls = append(l.successfulUrls, url)
}

type displayStats struct {
	inQueue, inProgress, success, allErrors uint64
}

func (l *Logger) loadDisplayStats() displayStats {
	var (
		queued  = atomic.LoadUint64(&l.queued)
		started = atomic.LoadUint64(&l.started)
		errs    = atomic.LoadUint64(&l.errs)
		success = atomic.LoadUint64(&l.success)
		full    = atomic.LoadUint64(&l.full)
	)
	return displayStats{
		inQueue:    queued - started,
		inProgress: started - success - errs,
		success:    success,
		allErrors:  errs + full,
	}
}

const ClearScreen = "\033[H\033[2J"

/*
func printMemStats() {
	fmt.Println("")
	fmt.Println("Mem stats")
	fmt.Println("---------")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}
	w := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	fmt.Fprintf(w, "Alloc\t%v MiB\n", bToMb(m.Alloc))
	fmt.Fprintf(w, "TotalAlloc\t%v MiB\n", bToMb(m.TotalAlloc))
	fmt.Fprintf(w, "Sys\t%v MiB\n", bToMb(m.Sys))
	fmt.Fprintf(w, "Goroutines\t%d\n", runtime.NumGoroutine())
	w.Flush()
}
*/

// Package consolelogger defines a logger.Interface that emits logs to a writer (usually the console)
package consolelogger

import (
	"fmt"
	"io"
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

// Logger is a logger.Interface that emits logs to a writer (usually the console)
type Logger struct {
	Writer                               io.Writer             // where to print the logs
	successfulUrls                       []string              // all successful urls (will be sorted and listed at exit)
	lastURLStarted                       string                // last url that started processing
	lastErr                              error                 // last error received
	queued, started, errs, success, full uint64                // counters for various stats
	ticker                               *time.Ticker          // ticker ticks every 200ms to display stats
	exiting                              bool                  // used to ensure stats don't display after ticker is stopped
	hist                                 *ghistogram.Histogram // displays a histogram of latencies
	m                                    sync.Mutex            // If ultimate performance was a concern, we could have a mutex per variable but this will simplify
}

// printSummary prints a summary of the logs to the writer
func (l *Logger) printSummary() {

	stats := l.loadDisplayStats()

	fmt.Fprint(l.Writer, ClearScreen)
	fmt.Fprintln(l.Writer, "Summary")
	fmt.Fprintln(l.Writer, "-------")

	w := tabwriter.NewWriter(l.Writer, 4, 3, 3, ' ', 0)
	fmt.Fprintf(w, "Queued\t%d\n", stats.inQueue)
	fmt.Fprintf(w, "In progress\t%d\t%s\n", stats.inProgress, l.getLastURLStarted())
	fmt.Fprintf(w, "Success\t%d\n", stats.success)
	fmt.Fprintf(w, "Errors\t%d\t%s\n", stats.allErrors, l.getLastErr())
	w.Flush()

	// l.printMemStats()

	fmt.Fprintln(l.Writer, "")
	fmt.Fprintln(l.Writer, "Latency")
	fmt.Fprintln(l.Writer, "-------")
	fmt.Fprintln(l.Writer, l.hist.EmitGraph(nil, nil).String())

}

// Init initialises the logger and starts the summary ticker
func (l *Logger) Init() {

	// Default to stdout if no Writer is specified
	if l.Writer == nil {
		l.Writer = os.Stdout
	}

	// Initialise the histogram
	l.hist = ghistogram.NewHistogram(21, 100, 0.0)

	// Start the ticker
	l.ticker = time.NewTicker(200 * time.Millisecond)

	// Print a summary on every tick
	go func() {
		for range l.ticker.C {
			if l.isExiting() {
				return
			}
			l.printSummary()
		}
	}()
}

// Queued is called each time a url is successfully queued
func (l *Logger) Queued(url string) {
	atomic.AddUint64(&l.queued, 1)
}

// Starting is called each time a url starts processing
func (l *Logger) Starting(url string) {
	atomic.AddUint64(&l.started, 1)
	l.setLastURLStarted(url)
}

// Finished is called each time a URL successfully finishes processing (even for non-200 results)
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
	l.addURLSuccess(url)
}

// Error is called on every error
func (l *Logger) Error(url string, err error) {

	// TODO: add latency here and log to the histogram

	switch err {
	case queuer.ErrDuplicate:
		// ignore duplicate errors
	case queuer.ErrFull:
		atomic.AddUint64(&l.full, 1)
		l.setLastErr(err)
	default:
		atomic.AddUint64(&l.errs, 1)
		l.setLastErr(err)
	}
}

// Exit stops the status ticker, and prints a sorted list of the successful urls
func (l *Logger) Exit() {

	l.stopTicker()

	// Sort the strings
	sort.Strings(l.successfulUrls)

	fmt.Fprintln(l.Writer, "URLs")
	fmt.Fprintln(l.Writer, "----")
	for _, u := range l.successfulUrls {
		fmt.Fprintln(l.Writer, u)
	}
}

func (l *Logger) isExiting() bool {
	l.m.Lock()
	defer l.m.Unlock()
	return l.exiting
}

func (l *Logger) stopTicker() {
	l.ticker.Stop()
	l.m.Lock()
	defer l.m.Unlock()
	l.exiting = true
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

func (l *Logger) getLastURLStarted() string {
	l.m.Lock()
	defer l.m.Unlock()
	return l.lastURLStarted
}

func (l *Logger) setLastURLStarted(u string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.lastURLStarted = u
}

func (l *Logger) addURLSuccess(url string) {
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

// ClearScreen is the control code to clear the screen
const ClearScreen = "\033[H\033[2J"

/*
func (l *Logger) printMemStats() {
	fmt.Fprintln(l.Writer, "")
	fmt.Fprintln(l.Writer, "Mem stats")
	fmt.Fprintln(l.Writer, "---------")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}
	w := tabwriter.NewWriter(l.Writer, 2, 2, 2, ' ', 0)
	fmt.Fprintf(w, "Alloc\t%v MiB\n", bToMb(m.Alloc))
	fmt.Fprintf(w, "TotalAlloc\t%v MiB\n", bToMb(m.TotalAlloc))
	fmt.Fprintf(w, "Sys\t%v MiB\n", bToMb(m.Sys))
	fmt.Fprintf(w, "Goroutines\t%d\n", runtime.NumGoroutine())
	w.Flush()
}
*/

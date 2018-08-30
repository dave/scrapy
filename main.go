// Package main is a simple command line interface for the scraper library.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dave/scrapy/scraper"
	"github.com/dave/scrapy/scraper/getter/webgetter"
	"github.com/dave/scrapy/scraper/logger/consolelogger"
	"github.com/dave/scrapy/scraper/parser/htmlparser"
	"github.com/dave/scrapy/scraper/queuer/concurrentqueuer"
)

func main() {

	config := struct {
		url             string
		length, workers int
		timeout         int
	}{}

	flag.StringVar(&config.url, "url", "https://monzo.com", "The start page")
	flag.IntVar(&config.length, "length", 1000, "Length of the queue")
	flag.IntVar(&config.workers, "workers", 5, "Number of concurrent workers")
	flag.IntVar(&config.timeout, "timeout", 10000, "Request timeout in ms")
	flag.Parse()

	// If there is an anonymous command line argument, use it as the url
	if arg := flag.Arg(0); arg != "" {
		config.url = arg
	}

	// Make sure we can parse the URL
	base, err := url.Parse(config.url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create a context that will be cancelled on Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Set up graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		// Wait for shutdown signal
		<-stop

		// clear the "^C" emitted to the console
		// TODO: Is this cross-platform?
		fmt.Print("\r")

		// Call the context cancellation function
		cancel()
	}()

	// Create a scraper
	s := &scraper.State{
		Timeout: time.Duration(config.timeout) * time.Millisecond,
		Getter:  &webgetter.Getter{},
		Parser: &htmlparser.Parser{
			Include: func(u *url.URL) bool {
				// Only accept the url if the host matches the host of the base page - e.g. some domain.
				return u != nil && u.Host == base.Host
			},
		},
		Queuer: &concurrentqueuer.Queuer{Length: config.length, Workers: config.workers},
		Logger: &consolelogger.Logger{},
	}

	// Start the scraper
	s.Start(ctx, base.String())
}

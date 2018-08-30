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

	// The initial url is the first command line argument - default to https://monzo.com if not found.
	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		arg = "https://monzo.com"
	}

	// Make sure we can parse the URL
	base, err := url.Parse(arg)
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
		Timeout: time.Second * 10,
		Getter:  &webgetter.Getter{},
		Parser: &htmlparser.Parser{
			Include: func(u *url.URL) bool {
				// Only accept the url if the host matches the host of the base page - e.g. some domain.
				return u != nil && u.Host == base.Host
			},
		},
		Queuer: &concurrentqueuer.Queuer{Length: 1000, Workers: 5},
		Logger: &consolelogger.Logger{},
	}

	// Start the scraper
	s.Start(ctx, base.String())
}

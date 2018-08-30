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

	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		arg = "https://monzo.com"
	}

	base, err := url.Parse(arg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Set up graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		// Wait for shutdown signal
		<-stop

		fmt.Print("\r") // clear the "^C" emitted to the console TODO: Is this cross-platform?

		// Call the context cancellation function
		cancel()
	}()

	s := &scraper.State{
		Timeout: time.Second * 10,
		Getter:  &webgetter.Getter{},
		Parser: &htmlparser.Parser{
			Include: func(u *url.URL) bool { return u != nil && u.Host == base.Host },
		},
		Queuer: &concurrentqueuer.Queuer{Length: 1000, Workers: 5},
		Logger: &consolelogger.Logger{},
	}

	s.Start(ctx, base.String())
}

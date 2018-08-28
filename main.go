package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dave/scrapy/scraper"
)

func main() {
	flag.Parse()
	s := scraper.New()
	if err := s.Start(flag.Arg(0)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

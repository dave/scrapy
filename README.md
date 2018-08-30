[![Build Status](https://travis-ci.org/dave/scrapy.svg?branch=master)](https://travis-ci.org/dave/scrapy) 
[![Go Report Card](https://goreportcard.com/badge/github.com/dave/scrapy)](https://goreportcard.com/report/github.com/dave/scrapy) 
[![codecov](https://codecov.io/gh/dave/scrapy/branch/master/graph/badge.svg)](https://codecov.io/gh/dave/scrapy)

# A simple web scraper

### Install

```
go get -u github.com/dave/scrapy
```

### Usage

```
scrapy [url]
```

The `scrapy` command will get get the page at `url`, parse it for links and get all pages that are 
on the same domain.

Some stats will be outputted during the processing, and a list of URLs will be printed when it's 
finished. You can end the job early with Ctrl+C.

### Flags

Several command line flags are available:

```
  -length int
    	Length of the queue (default 1000)
  -timeout int
    	Request timeout in ms (default 10000)
  -url string
    	The start page (default "https://monzo.com")
  -workers int
    	Number of concurrent workers (default 5)
```

### Library

This scraper can also be used as a library. See the [scraper](https://godoc.org/github.com/dave/scrapy/scraper) package.

### Notes

See [here](https://github.com/dave/scrapy/blob/master/NOTES.md) for design notes and brainstorming.

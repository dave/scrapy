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

### Example output

```
Summary
-------
Queued        46
In progress   5   https://monzo.com/blog/2018/08/30/manage-your-bills
Success       22
Errors        0   

Latency
-------
   0 - 100  ***
 100 - 200 
 200 - 300 
 300 - 400  **************************
 400 - 500  ******************************
 500 - 600  ***************
 600 - 700  ***
 700 - 800  ***
 800 - 900 
 900 - 1000
1000 - 1100
1100 - 1200
1200 - 1300
1300 - 1400
1400 - 1500
1500 - 1600
1600 - 1700
1700 - 1800
1800 - 1900
1900 - 2000
2000+ 

URLs
----
https://monzo.com
https://monzo.com/-play-store-redirect
https://monzo.com/about
https://monzo.com/blog
https://monzo.com/blog/2018/07/02/publishing-our-2018-annual-report
https://monzo.com/blog/2018/07/10/making-quarterly-goals-public
https://monzo.com/blog/2018/07/25/monzo-reliability-report
https://monzo.com/blog/how-money-works
https://monzo.com/blog/latest

...
```
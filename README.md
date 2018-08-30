[![Build Status](https://travis-ci.org/dave/scrapy.svg?branch=master)](https://travis-ci.org/dave/scrapy) 
[![Go Report Card](https://goreportcard.com/badge/github.com/dave/scrapy)](https://goreportcard.com/report/github.com/dave/scrapy) 
[![codecov](https://codecov.io/gh/dave/scrapy/branch/master/graph/badge.svg)](https://codecov.io/gh/dave/scrapy)

# Web scraper brainstorming

### Features
* Keep it simple. Don't go crazy. 
* Would love to do a PID controller to optimize concurrency but that's not what they want.

### Design process
* Split out independent parts
* Can each part be represented by a nice interface?
* Create mocks / real services
* Test independently
* Glue them together
* Test together
* How to test concurrent code - always hard.

### Optimize
* Don't go too crazy on this, but let's do a couple of optimizations and some benchmarking.

### Offline testing
* Make a mode that records the entire site structure, along with server latencies, then a testing mode that replays this. 
* This can be used for more real-world benchmarks

### Independent systems
* Get a page -> Raw page: `Getter`
* Raw page -> Parse the page -> Page stats: `Parser`
* Page stats -> Queue new actions, log stats
* Take from the queue and start getting -> Start getting page

### Considerations
* Make it usable as a library, and usable in a server system - e.g. respect context.Context and cancellation.
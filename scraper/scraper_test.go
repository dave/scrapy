package scraper

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/dave/scrapy/scraper/getter/mockgetter"
	"github.com/dave/scrapy/scraper/logger/mocklogger"
	"github.com/dave/scrapy/scraper/parser/mockparser"
	"github.com/dave/scrapy/scraper/queuer/concurrentqueuer"
)

func TestScraper(t *testing.T) {
	tests := []struct {
		name            string
		length, workers int
		timeout         time.Duration
		start           string
		get             map[string]mockgetter.Dummy
		parse           map[string]mockparser.Dummy
		expected        []string
		cancel          bool
	}{
		{
			name: "simple",
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body"},
			},
			parse:    map[string]mockparser.Dummy{},
			expected: []string{"queue a", "start a", "finish a: 200, 0, 0"},
		},
		{
			name: "simple parsed",
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body"},
			},
			parse: map[string]mockparser.Dummy{
				"a_body": {Urls: []string{"b"}},
			},
			expected: []string{"queue a", "start a", "finish a: 200, 1, 0", "queue b", "start b", "finish b: 404, 0, 0"},
		},
		{
			name:    "queue full",
			length:  2,
			workers: 1,
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body"},
			},
			parse: map[string]mockparser.Dummy{
				"a_body": {Urls: []string{"b", "c", "d"}},
			},
			expected: []string{"queue a", "start a", "finish a: 200, 3, 0", "queue b", "queue c", "error d: queue full", "start b", "finish b: 404, 0, 0", "start c", "finish c: 404, 0, 0"},
		},
		{
			name:    "duplicate",
			length:  2,
			workers: 1,
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body"},
			},
			parse: map[string]mockparser.Dummy{
				"a_body": {Urls: []string{"b", "b"}},
			},
			expected: []string{"queue a", "start a", "finish a: 200, 2, 0", "queue b", "error b: duplicate url", "start b", "finish b: 404, 0, 0"},
		},
		{
			name: "complex",
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body"},
				"b": {Body: "b_body"},
				"c": {Body: "c_body"},
				"d": {Body: "d_body"},
			},
			parse: map[string]mockparser.Dummy{
				"a_body": {Urls: []string{"b", "c"}},
				"c_body": {Urls: []string{"d", "e"}},
			},
			expected: []string{"queue a", "start a", "finish a: 200, 2, 0", "queue b", "queue c", "start b", "finish b: 200, 0, 0", "start c", "finish c: 200, 2, 0", "queue d", "queue e", "start d", "finish d: 200, 0, 0", "start e", "finish e: 404, 0, 0"},
		},
		{
			name:    "timeout",
			timeout: time.Millisecond * 10,
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body", Latency: time.Second},
			},
			parse:    map[string]mockparser.Dummy{},
			expected: []string{"queue a", "start a", "error a: context deadline exceeded"},
		},
		{
			name:   "cancel",
			cancel: true,
			get: map[string]mockgetter.Dummy{
				"a": {Body: "a_body", Latency: time.Second},
			},
			parse:    map[string]mockparser.Dummy{},
			expected: []string{"queue a", "start a", "error a: context canceled"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			log := &mocklogger.Logger{}

			timeout := time.Second
			if test.timeout > 0 {
				timeout = test.timeout
			}

			workers := 1
			if test.workers > 0 {
				workers = test.workers
			}

			length := 10
			if test.length > 0 {
				length = test.length
			}

			start := "a"
			if test.start != "" {
				start = test.start
			}

			ctx := context.Background()
			if test.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			state := &State{
				Timeout: timeout,
				Getter:  &mockgetter.Getter{Results: test.get},
				Parser:  &mockparser.Parser{Results: test.parse},
				Queuer:  &concurrentqueuer.Queuer{Length: length, Workers: workers},
				Logger:  log,
			}

			state.Start(ctx, start)

			if !reflect.DeepEqual(log.Log, test.expected) {
				t.Errorf("unexpected log contents - found %#v", log.Log)
			}
		})
	}
}

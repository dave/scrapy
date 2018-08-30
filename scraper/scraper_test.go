package scraper

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/dave/scrapy/scraper/getter/mockgetter"
	"github.com/dave/scrapy/scraper/logger/mocklogger"
	"github.com/dave/scrapy/scraper/parser/mockparser"
	"github.com/dave/scrapy/scraper/queuer/concurrentqueuer"
)

var tests = map[string]testSpec{
	"simple": {
		get: map[string]mockgetter.Dummy{
			"a": {Body: "a_body"},
		},
		parse:    map[string]mockparser.Dummy{},
		expected: []string{"queue a", "start a", "finish a: 200, 0, 0"},
	},
	"simple parsed": {
		get: map[string]mockgetter.Dummy{
			"a": {Body: "a_body"},
		},
		parse: map[string]mockparser.Dummy{
			"a_body": {Urls: []string{"b"}},
		},
		expected: []string{"queue a", "start a", "finish a: 200, 1, 0", "queue b", "start b", "finish b: 404, 0, 0"},
	},
	"queue full": {
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
	"duplicate": {
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
	"complex": {
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
	"timeout": {
		timeout: time.Millisecond * 10,
		get: map[string]mockgetter.Dummy{
			"a": {Body: "a_body", Latency: time.Second},
		},
		parse:    map[string]mockparser.Dummy{},
		expected: []string{"queue a", "start a", "error a: context deadline exceeded"},
	},
	"cancel": {
		cancel: true,
		get: map[string]mockgetter.Dummy{
			"a": {Body: "a_body", Latency: time.Second},
		},
		parse:    map[string]mockparser.Dummy{},
		expected: []string{"queue a", "start a", "error a: context canceled"},
	},
}

type testSpec struct {
	single, skip    bool
	length, workers int
	timeout         time.Duration
	start           string
	get             map[string]mockgetter.Dummy
	parse           map[string]mockparser.Dummy
	expected        []string
	cancel          bool
}

func (s testSpec) run(name string, t *testing.T) {
	t.Helper()

	log := &mocklogger.Logger{}

	timeout := time.Second
	if s.timeout > 0 {
		timeout = s.timeout
	}

	workers := 1
	if s.workers > 0 {
		workers = s.workers
	}

	length := 10
	if s.length > 0 {
		length = s.length
	}

	start := "a"
	if s.start != "" {
		start = s.start
	}

	ctx := context.Background()
	if s.cancel {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
	}

	state := &State{
		Timeout: timeout,
		Getter:  &mockgetter.Getter{Results: s.get},
		Parser:  &mockparser.Parser{Results: s.parse},
		Queuer:  &concurrentqueuer.Queuer{Length: length, Workers: workers},
		Logger:  log,
	}

	state.Start(ctx, start)

	if !reflect.DeepEqual(log.Log, s.expected) {
		t.Errorf("unexpected log contents - found %#v", log.Log)
	}
}

func TestScraper(t *testing.T) {
	var single bool
	for name, test := range tests {
		if test.single {
			if single {
				panic("two tests marked as single")
			}
			single = true
			tests = map[string]testSpec{name: test}
		}
	}

	// order tests by name to ensure consistent execution order
	type named struct {
		testSpec
		name string
	}
	var ordered []named
	for name, spec := range tests {
		ordered = append(ordered, named{spec, name})
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].name < ordered[j].name })

	// run tests, skipping marked tests
	var skipped bool
	for _, spec := range ordered {
		if spec.skip {
			skipped = true
			continue
		}
		spec.run(spec.name, t)
	}

	// fail in single mode or if any tests are skipped
	if single {
		t.Fatal("test passed, but failed because single mode is set")
	}
	if skipped {
		t.Fatal("tests passed, but skipped some")
	}
}

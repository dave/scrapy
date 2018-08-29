package scraper

import (
	"testing"
	"time"

	"context"

	"reflect"

	"github.com/dave/scrapy/scraper/getter/mockgetter"
	"github.com/dave/scrapy/scraper/logger/mocklogger"
	"github.com/dave/scrapy/scraper/parser/mockparser"
	"github.com/dave/scrapy/scraper/queuer/concurrentqueuer"
)

func TestScraper(t *testing.T) {
	log := &mocklogger.Logger{}
	s := &State{
		Timeout: time.Second,
		Getter: &mockgetter.Getter{
			Results: map[string]mockgetter.Dummy{
				"a": {
					Body:    "a_body",
					Latency: time.Millisecond,
				},
				"b": {
					Body:    "b_body",
					Latency: time.Millisecond,
				},
				"c": {
					Body:    "c_body",
					Latency: time.Millisecond,
				},
				"d": {
					Body:    "d_body",
					Latency: time.Millisecond,
				},
			},
		},
		Parser: &mockparser.Parser{
			Results: map[string]mockparser.Dummy{
				"a_body": {
					Urls: []string{"b", "c"},
				},
				"c_body": {
					Urls: []string{"d", "e"},
				},
			},
		},
		Queuer: &concurrentqueuer.Queuer{Length: 10, Workers: 1},
		Logger: log,
	}

	s.Start(context.Background(), "a")

	expected := []string{"queue a", "start a", "finish a: 200, 2, 0", "queue b", "queue c", "start b", "finish b: 200, 0, 0", "start c", "finish c: 200, 2, 0", "queue d", "queue e", "start d", "finish d: 200, 0, 0", "start e", "finish e: 404, 0, 0"}

	if !reflect.DeepEqual(log.Log, expected) {
		t.Errorf("unexpected log contents - found %#v", log.Log)
	}
}

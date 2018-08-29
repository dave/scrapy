package mockparser

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

// TODO: Quick test for mock parser - perhaps improve if have time.

func TestParser(t *testing.T) {
	p := &Parser{
		Results: map[string]Dummy{
			"a": {
				Urls: []string{"b"},
			},
		},
	}
	urls, errs := p.Parse("", ioutil.NopCloser(bytes.NewBufferString("a")))
	expected := []string{"b"}
	if !reflect.DeepEqual(urls, expected) {
		t.Errorf("expected urls: %#v, found %#v", expected, urls)
	}
	if len(errs) > 0 {
		t.Errorf("expected nil errs, found %#v", errs)
	}
}

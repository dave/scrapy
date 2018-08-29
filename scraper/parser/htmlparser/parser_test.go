package htmlparser

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

var tests = map[string]testSpec{
	"simple": {
		body: `<a href="a"></a>`,
		urls: []string{"a"},
	},
	"url error": {
		body: `<a href=":"></a>`,
		errs: []string{"parse :: missing protocol scheme"},
	},
	"complex html": {
		body: `<body><p><a href="a"></a></p><table><td><a href="b"></a></td></table><!--<a href="c"></a>--></body>`,
		urls: []string{"a", "b"},
	},
	"html and errors": {
		body: `<body><a href=":"></a><p><a href="a"></a></p><div><a href="b"></a><a href="1:2"></a></div></body>`,
		urls: []string{"a", "b"},
		errs: []string{"parse :: missing protocol scheme", "parse 1:2: first path segment in URL cannot contain colon"},
	},
	"include function": {
		body: `<a href="http://a.com/a"></a><a href="http://b.com/b"></a>`,
		inc:  func(url *url.URL) bool { return url != nil && url.Host == "b.com" },
		urls: []string{"http://b.com/b"},
	},
}

type testSpec struct {
	single, skip bool
	body         string
	urls         []string
	errs         []string
	inc          func(url *url.URL) bool
}

func (s testSpec) run(name string, t *testing.T) {
	t.Helper()

	p := &Parser{}

	if s.inc != nil {
		p.Include = s.inc
	}

	body := ioutil.NopCloser(bytes.NewBufferString(s.body))

	urls, errs := p.Parse(body)

	if !reflect.DeepEqual(urls, s.urls) {
		t.Errorf("%s: unexpected urls - got: %#v, expected: %#v", name, urls, s.urls)
	}
	var errorStrings []string
	for _, e := range errs {
		errorStrings = append(errorStrings, e.Error())
	}
	if !reflect.DeepEqual(errorStrings, s.errs) {
		t.Errorf("%s: unexpected errors - got: %#v, expected: %#v", name, errorStrings, s.errs)
	}
}

func TestParser(t *testing.T) {
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

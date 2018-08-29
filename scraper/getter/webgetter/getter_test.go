package webgetter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"io/ioutil"

	"github.com/dave/scrapy/scraper/getter"
)

var tests = map[string]testSpec{
	"simple": {
		body: "a\n",
		code: 200,
	},
	"broken url": {
		overrideUrl: "a",
		err:         "unsupported protocol scheme",
	},
	"404": {
		body: "a\n",
		code: 404,
	},
	"timeout": {
		returnAfter: 100,
		cancelAfter: 5,
		err:         "context deadline exceeded",
	},
}

type testSpec struct {
	single, skip bool
	body         string
	code         int
	err          string
	overrideUrl  string
	returnAfter  int
	cancelAfter  int
}

func (s testSpec) run(name string, t *testing.T) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.returnAfter > 0 {
			<-time.After(time.Duration(s.returnAfter) * time.Millisecond)
		}
		if s.code != 0 && s.code != 200 {
			w.WriteHeader(s.code)
		}
		fmt.Fprint(w, s.body)
	}))
	defer ts.Close()

	g := &Getter{}

	ctx := context.Background()
	if s.cancelAfter > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.cancelAfter)*time.Millisecond)
		defer cancel()
	}

	url := ts.URL
	if s.overrideUrl != "" {
		url = s.overrideUrl
	}

	c := g.Get(ctx, url)

	var r getter.Result
	select {
	case r = <-c:
		// great.
	case <-time.After(time.Millisecond * 200):
		// fatal time out after 200ms
		t.Fatalf("%s: test took too long", name)
	}

	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("%s: error reading body: %v", name, err)
	}

	if string(b) != s.body {
		t.Errorf("%s: expected body %q, got %q", name, s.body, r.Body)
	}

	if r.Code != s.code {
		t.Errorf("%s: expected code %d, got %d", name, s.code, r.Code)
	}

	if s.err == "" {
		if r.Err != nil {
			t.Errorf("%s: expected success, got error: %v", name, r.Err)
		}
	} else {
		if r.Err == nil {
			t.Errorf("%s: expected error %s, got nil", name, s.err)
		} else if !strings.Contains(r.Err.Error(), s.err) {
			t.Errorf("%s: expected error to contain %s, got %v", name, s.err, r.Err)
		}
	}
}

func TestGetter(t *testing.T) {
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

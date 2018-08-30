package htmlparser

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestNormalise(t *testing.T) {
	tests := []struct {
		name, url, page, expected, err string
	}{
		{
			name:     "simple",
			url:      "https://a",
			expected: "https://a",
		},
		{
			name:     "throw away tel",
			url:      "tel:0",
			expected: "",
		},
		{
			name:     "throw away mailto",
			url:      "mailto:a@a",
			expected: "",
		},
		{
			name:     "throw away javascript",
			url:      "javascript:a",
			expected: "",
		},
		{
			name:     "throw away binary",
			url:      "https://a/b.pdf",
			expected: "",
		},
		{
			name: "parse error",
			url:  ":",
			err:  "missing protocol scheme",
		},
		{
			name:     "copy scheme and host from page if missing",
			url:      "/d",
			page:     "https://a/b/c",
			expected: "https://a/d",
		},
		{
			name:     "relative path",
			url:      "./../../e",
			page:     "https://a/b/c/d",
			expected: "https://a/b/e",
		},
		{
			name:     "remove trailing slash",
			url:      "https://a/b/",
			expected: "https://a/b",
		},
		{
			name:     "prevent jumping from https to http",
			url:      "http://a/b",
			page:     "https://a",
			expected: "https://a/b",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			page, err := url.Parse(test.page)
			if err != nil {
				t.Fatal("parsing page url failed")
			}
			out, err := normalise(test.url, page)
			if test.err == "" && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if test.err != "" && !strings.Contains(err.Error(), test.err) {
				t.Errorf("expected error %s but got %v", test.err, err)
			}
			if test.expected == "" && out != nil {
				t.Errorf("expected nil url, but got %s", out.String())
			}
			if test.expected != "" && (out == nil || out.String() != test.expected) {
				t.Errorf("expected %s, but got %v", test.expected, out)
			}
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		name string
		body string
		urls []string
		errs []string
		inc  func(url *url.URL) bool
	}{
		{
			name: "simple",
			body: `<a href="a"></a>`,
			urls: []string{"a"},
		},
		{
			name: "url error",
			body: `<a href=":"></a>`,
			errs: []string{"parse :: missing protocol scheme"},
		},
		{
			name: "complex html",
			body: `<body><p><a href="a"></a></p><table><td><a href="b"></a></td></table><!--<a href="c"></a>--></body>`,
			urls: []string{"a", "b"},
		},
		{
			name: "html and errors",
			body: `<body><a href=":"></a><p><a href="a"></a></p><div><a href="b"></a><a href="1:2"></a></div></body>`,
			urls: []string{"a", "b"},
			errs: []string{"parse :: missing protocol scheme", "parse 1:2: first path segment in URL cannot contain colon"},
		},
		{
			name: "include function",
			body: `<a href="http://a.com/a"></a><a href="http://b.com/b"></a>`,
			inc:  func(url *url.URL) bool { return url != nil && url.Host == "b.com" },
			urls: []string{"http://b.com/b"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &Parser{}

			if test.inc != nil {
				p.Include = test.inc
			}

			body := ioutil.NopCloser(bytes.NewBufferString(test.body))

			urls, errs := p.Parse(context.Background(), "", body)

			if !reflect.DeepEqual(urls, test.urls) {
				t.Errorf("unexpected urls - got: %#v, expected: %#v", urls, test.urls)
			}
			var errorStrings []string
			for _, e := range errs {
				errorStrings = append(errorStrings, e.Error())
			}
			if !reflect.DeepEqual(errorStrings, test.errs) {
				t.Errorf("unexpected errors - got: %#v, expected: %#v", errorStrings, test.errs)
			}
		})
	}
}

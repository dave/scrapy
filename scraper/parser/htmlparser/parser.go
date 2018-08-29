package htmlparser

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

type Parser struct {
	Include func(*url.URL) bool
}

func (p *Parser) Parse(body io.ReadCloser) (urls []string, errs []error) {
	defer body.Close()
	t := html.NewTokenizer(body)
	for {
		typ := t.Next()
		switch typ {
		case html.ErrorToken:
			if t.Err() == io.EOF {
				// end of document
				return
			}
			// parser error - log
			errs = append(errs, t.Err())
		case html.StartTagToken:
			tok := t.Token()
			if tok.Data != "a" {
				continue
			}
			for _, att := range tok.Attr {
				if att.Key != "href" {
					continue
				}
				u, err := url.Parse(att.Val)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if p.Include != nil && !p.Include(u) {
					continue
				}
				urls = append(urls, att.Val)
			}
		}
	}
}

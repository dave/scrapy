package htmlparser

import (
	"io"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type Parser struct {
	Include func(*url.URL) bool
}

func (p *Parser) Parse(urlPage string, body io.Reader) (urls []string, errs []error) {
	t := html.NewTokenizer(body)

	page, err := url.Parse(urlPage)
	if err != nil {
		return nil, []error{err}
	}

	for {
		typ := t.Next()
		switch typ {
		case html.ErrorToken:

			// End of document
			if t.Err() == io.EOF {
				return
			}

			// Log a parser error
			errs = append(errs, t.Err())
			continue

		case html.StartTagToken:
			tok := t.Token()

			// Look for a tags. TODO: Look for more tags?
			if tok.Data != "a" {
				continue
			}

			for _, att := range tok.Attr {

				// Look for href attributes. TODO: Look for more attributes?
				if att.Key != "href" {
					continue
				}

				// Let's throw away a common error
				if strings.HasPrefix(att.Val, "tel:") {
					break
				}

				u, err := url.Parse(att.Val)
				if err != nil {
					errs = append(errs, err)
					break
				}

				// Add scheme and host from page for relative URLs
				if u.Scheme == "" {
					u.Scheme = page.Scheme
				}
				if u.Host == "" {
					u.Host = page.Host
				}
				// Handle relative paths - e.g. "../foo/bar.html"
				if !path.IsAbs(u.Path) {
					u.Path = path.Join(page.Path, u.Path)
				}

				// Clear path fragment to reduce duplication
				u.Fragment = ""

				// Run the include function if it exists and skip this url if needed
				if p.Include != nil && !p.Include(u) {
					break
				}

				// Stop scanning attributes once we've found a url
				urls = append(urls, u.String())
				break
			}
		}
	}
}

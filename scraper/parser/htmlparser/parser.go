// Package htmlparser defines a parser.Interface that parses HTML and returns the urls from anchor href attributes
package htmlparser

import (
	"context"
	"io"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Parser is a parser.Interface that parses HTML and returns the urls from anchor href attributes
type Parser struct {
	Include func(*url.URL) bool
}

// Parse parses the document and returns the urls and parse errors
func (p *Parser) Parse(ctx context.Context, urlPage string, body io.Reader) (urls []string, errs []error) {
	t := html.NewTokenizer(body)

	page, err := url.Parse(urlPage)
	if err != nil {
		return nil, []error{err}
	}

	for {
		select {
		case <-ctx.Done():
			return nil, []error{ctx.Err()}
		default:
			// great!
		}
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

			// Look for a tags.
			// TODO: Look for more tags?
			if tok.Data != "a" {
				continue
			}

			for _, att := range tok.Attr {

				// Look for href attributes.
				// TODO: Look for more attributes?
				if att.Key != "href" {
					continue
				}

				u, err := normalise(att.Val, page)
				if err != nil {
					errs = append(errs, err)
					break
				}
				if u == nil {
					break
				}

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

// normalise performs modifications on a url in order to reduce duplicates and errors
func normalise(raw string, page *url.URL) (*url.URL, error) {

	// Let's throw away a common errors
	if strings.HasPrefix(raw, "tel:") || strings.HasPrefix(raw, "mailto:") {
		return nil, nil
	}

	// URLs we don't want to get
	if strings.HasSuffix(raw, ".zip") || strings.HasSuffix(raw, ".pdf") || strings.HasSuffix(raw, ".png") || strings.HasSuffix(raw, ".jpg") {
		return nil, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
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

	// Normalize URLs to remove trailing slashes
	if strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path[:len(u.Path)-1]
	}

	// Prevent https page from linking to non-https version?
	if page.Scheme == "https" && u.Scheme == "http" {
		u.Scheme = "https"
	}

	// Clear path fragment to reduce duplication
	u.Fragment = ""

	return u, nil
}

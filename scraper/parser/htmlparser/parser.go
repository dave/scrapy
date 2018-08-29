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

				// Let's throw away a common errors
				if strings.HasPrefix(att.Val, "tel:") || strings.HasPrefix(att.Val, "mailto:") {
					break
				}

				// URLs we don't want to get
				if strings.HasSuffix(att.Val, ".zip") || strings.HasSuffix(att.Val, ".pdf") || strings.HasSuffix(att.Val, ".png") || strings.HasSuffix(att.Val, ".jpg") {
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

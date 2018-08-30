// Package mockparser defines a parser.Interface that returns dummy urls for a given input, and is used in tests
package mockparser

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
)

// Parser is a parser.Interface that returns dummy urls for a given input, and is used in tests
type Parser struct {
	Results map[string]Dummy // The results to return: body -> result
}

// Dummy responses
type Dummy struct {
	Urls []string // List of urls
	Errs []string // List of parse errors as strings
}

// Parse returns the dummy data if Results contains a matching record.
func (p *Parser) Parse(ctx context.Context, url string, body io.Reader) (urls []string, errs []error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, []error{err}
	}
	result, ok := p.Results[string(b)]
	if !ok {
		return nil, nil
	}
	for _, e := range result.Errs {
		errs = append(errs, errors.New(e))
	}
	return result.Urls, errs
}

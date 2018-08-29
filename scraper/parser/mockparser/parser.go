package mockparser

import (
	"errors"
	"io"
	"io/ioutil"
)

type Parser struct {
	Results map[string]Dummy
}

type Dummy struct {
	Urls []string
	Errs []string
}

func (p *Parser) Parse(body io.Reader) (urls []string, errs []error) {
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

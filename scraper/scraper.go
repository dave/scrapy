package scraper

import (
	"fmt"
)

type State struct {
}

func New() *State {
	return &State{}
}

func (s *State) Start(url string) error {
	fmt.Println(url)
	return nil
}

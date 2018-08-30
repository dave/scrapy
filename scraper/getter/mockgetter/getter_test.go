package mockgetter

import (
	"context"
	"io/ioutil"
	"testing"
	"time"
)

// A quick test for mock getter
func TestGetter(t *testing.T) {

	// TODO: perhaps improve if have time.

	g := &Getter{
		Results: map[string]Dummy{
			"a": {
				Body:    "b",
				Latency: time.Millisecond,
			},
		},
	}
	r := <-g.Get(context.Background(), "a")
	b, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "b" {
		t.Errorf("expected %s, found %s", "b", string(b))
	}
}

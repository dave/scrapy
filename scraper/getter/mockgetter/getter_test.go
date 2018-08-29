package mockgetter

import (
	"context"
	"io/ioutil"
	"testing"
	"time"
)

// TODO: Quick test for mock getter - perhaps improve if have time.

func TestGetter(t *testing.T) {
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

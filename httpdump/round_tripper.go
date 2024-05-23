package httpdump

import (
	"net/http"
	"testing"

	"github.com/alextanhongpin/testdump/httpdump/internal"
)

type RoundTripper struct {
	t    *testing.T
	opts []Option
}

func RoundTrip(t *testing.T, opts ...Option) *RoundTripper {
	return &RoundTripper{
		t:    t,
		opts: opts,
	}
}

func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Copy the body.
	rc, err := internal.CloneRequest(r)
	if err != nil {
		return nil, err
	}

	w, err := http.DefaultTransport.RoundTrip(r)
	New(rt.opts...).Dump(rt.t, w, rc)

	return w, err
}

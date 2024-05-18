package http

import (
	"net/http"
	"testing"

	"github.com/alextanhongpin/dump/http/internal"
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

	h := Handler(rt.t, nil, rt.opts...)
	if err := h.dump(w, rc); err != nil {
		return nil, err
	}

	return w, err
}

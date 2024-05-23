package httpdump

import (
	"net/http"
	"testing"

	"github.com/alextanhongpin/testdump/httpdump/internal"
)

// RoundTripper is a struct that holds a testing object and a slice of options.
type RoundTripper struct {
	t    *testing.T
	opts []Option
}

// RoundTrip is a function that takes a testing object and a variadic list of options.
// It returns a new RoundTripper instance with these values.
func RoundTrip(t *testing.T, opts ...Option) *RoundTripper {
	return &RoundTripper{
		t:    t,
		opts: opts,
	}
}

// RoundTrip is a method on the RoundTripper struct that takes an HTTP request.
// It clones the request, sends the request to the default transport, dumps the response, and then returns the response and any error.
func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Copy the body.
	rc, err := internal.CloneRequest(r)
	if err != nil {
		return nil, err
	}

	// Send the request to the default transport.
	w, err := http.DefaultTransport.RoundTrip(r)

	// Dump the response.
	New(rt.opts...).Dump(rt.t, w, rc)

	return w, err
}

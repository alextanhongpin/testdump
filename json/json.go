package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/alextanhongpin/dump/json/internal"
	"github.com/google/go-cmp/cmp"
)

type Middleware func(b []byte) ([]byte, error)

type Option struct {
	CompareOptions []cmp.Option
	Middlewares    []Middleware
}
type Dump struct {
	t   *testing.T
	opt *Option
}

func NewDump(t *testing.T, opt *Option) *Dump {
	return &Dump{
		t:   t,
		opt: opt,
	}
}

func (d *Dump) Dump(t *testing.T, v any, opt *Option) error {
	t.Helper()

	received, err := json.Marshal(v)
	if err != nil {
		return err
	}

	for _, m := range append(d.opt.Middlewares, opt.Middlewares...) {
		received, err = m(received)
		if err != nil {
			return err
		}
	}

	// TODO: infer filename.
	file := fmt.Sprintf("testdata/%s.json", t.Name())
	overwrite := false
	written, err := internal.WriteFile(file, received, overwrite)
	if err != nil {
		return err
	}

	if written {
		return nil
	}

	snapshot, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if !bytes.Equal(snapshot, received) {
		return errors.New("mismatch")
	}

	return nil
}

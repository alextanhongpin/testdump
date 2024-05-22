package jsondump

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/alextanhongpin/dump/jsondump/internal"
	"github.com/alextanhongpin/dump/pkg/diff"
)

// New creates a new dumper with the given options.
// The dumper can be used to dump values to a file.
func New(opts ...Option) interface {
	Dump(t *testing.T, v any, opts ...Option)
} {
	return &dumper{opts: opts}
}

func Dump(t *testing.T, v any, opts ...Option) {
	d := &dumper{}
	d.Dump(t, v, opts...)
}

type dumper struct {
	opts []Option
}

func (d *dumper) Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	if err := dump(t, v, append(d.opts, opts...)...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, v any, opts ...Option) error {
	// Extract from struct tags.
	opt := newOption(v, opts...)

	receivedBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	for _, p := range opt.Processors {
		receivedBytes, err = p(receivedBytes)
		if err != nil {
			return err
		}
	}

	file := filepath.Join(
		"testdata",
		t.Name(),
		fmt.Sprintf("%s.json", internal.Or(opt.File, internal.TypeName(v))),
	)

	overwrite, _ := strconv.ParseBool(os.Getenv(opt.Env))
	written, err := internal.WriteFile(file, receivedBytes, overwrite)
	if err != nil {
		return err
	}

	if written {
		return nil
	}

	snapshotBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Since google's cmp does not have an option to ignore paths, we just mask
	// the values before comparing.
	// The masked values will not be written to the file.
	for _, p := range opt.IgnorePathsProcessors {
		snapshotBytes, err = p(snapshotBytes)
		if err != nil {
			return err
		}

		receivedBytes, err = p(receivedBytes)
		if err != nil {
			return err
		}
	}

	// Convert back to map[string]any for nicer diff.
	var snapshot, received any
	if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
		return err
	}
	if err := json.Unmarshal(receivedBytes, &received); err != nil {
		return err
	}

	return diff.ANSI(snapshot, received, opt.CmpOpts...)
}

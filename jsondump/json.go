package jsondump

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
)

var d *Dumper

func init() {
	d = new(Dumper)
}

// New creates a new Dumper with the given options.
// The Dumper can be used to dump values to a file.
func New(opts ...Option) *Dumper {
	return &Dumper{opts: opts}
}

func Dump(t *testing.T, v any, opts ...Option) {
	d.Dump(t, v, opts...)
}

type Dumper struct {
	opts []Option
}

func (d *Dumper) Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := dump(t, v, opts...); err != nil {
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

	for _, transform := range opt.transformers {
		receivedBytes, err = transform(receivedBytes)
		if err != nil {
			return err
		}
	}

	file := filepath.Join(
		"testdata",
		t.Name(),
		fmt.Sprintf("%s.json", internal.Or(opt.file, internal.TypeName(v))),
	)

	overwrite, _ := strconv.ParseBool(os.Getenv(opt.env))
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
	for _, transform := range opt.ignorePathsTransformers {
		snapshotBytes, err = transform(snapshotBytes)
		if err != nil {
			return err
		}

		receivedBytes, err = transform(receivedBytes)
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

	comparer := diff.Text
	if opt.colors {
		comparer = diff.ANSI
	}

	return comparer(snapshot, received, opt.cmpOpts...)
}

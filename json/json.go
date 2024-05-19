package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/dump/json/internal"
	"github.com/alextanhongpin/dump/pkg/diff"
)

func Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	if err := dump(t, v, opts...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, v any, opts ...Option) error {
	// Extract from struct tags.
	ignorePaths := IgnorePathsFromStructTag(v)
	maskPaths := MaskPathsFromStructTag(v)
	opts = append(opts, MaskPaths(maskPaths...))
	opts = append(opts, IgnorePaths(ignorePaths...))

	opt := newOption(opts...)

	receivedBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	for _, v := range opt.Validators {
		if err := v(receivedBytes); err != nil {
			return err
		}
	}

	for _, p := range opt.Processors {
		receivedBytes, err = p(receivedBytes)
		if err != nil {
			return err
		}
	}

	file := filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.json", internal.Or(opt.Name, internal.TypeName(v))))
	// TODO: change into a flag or env vars.
	overwrite := false
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
	for _, p := range opt.IgnorePathsProcessor {
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

package jsondump

import (
	"os"
	"strconv"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

// Define a constant for ignored values
const (
	ignoreValue = "[IGNORED]"
	env         = "TESTDUMP"
)

// Define a function type Option that takes a pointer to an options struct
type Option func(o *options)

// Define the options struct with various fields
type options struct {
	cmpOpts      []cmp.Option
	colors       bool
	env          string // The environment variable name to overwrite the snapsnot.
	file         string // A custom file name.
	ignorePaths  []string
	rawOutput    bool
	transformers []func([]byte) ([]byte, error)
}

func newOptions() *options {
	return &options{
		colors:    true,
		env:       env,
		rawOutput: true,
	}
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

func (o *options) overwrite() bool {
	t, _ := strconv.ParseBool(os.Getenv(o.env))
	return t
}

func (o *options) encoder() *encoder {
	return &encoder{
		marshalFns: o.transformers,
	}
}

func (o *options) comparer() *comparer {
	return &comparer{
		colors:      o.colors,
		ignorePaths: o.ignorePaths,
		opts:        o.cmpOpts,
	}
}

// File is an Option that sets the file name
func File(name string) Option {
	return func(o *options) {
		o.file = name
	}
}

// Env is an Option that sets the environment variable name
func Env(name string) Option {
	return func(o *options) {
		o.env = name
	}
}

// Colors is an Option that sets the colors flag
func Colors(colors bool) Option {
	return func(o *options) {
		o.colors = colors
	}
}

// Transformers is an Option that adds a transformer function
func Transformers(p ...func([]byte) ([]byte, error)) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, p...)
	}
}

// CmpOpts is an Option that adds comparison options
func CmpOpts(opts ...cmp.Option) Option {
	return func(o *options) {
		o.cmpOpts = append(o.cmpOpts, opts...)
	}
}

// IgnoreFields is an Option that ignores certain fields
func IgnoreFields(fields ...string) Option {
	return func(o *options) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths is an Option that ignores certain paths
func IgnorePaths(paths ...string) Option {
	return func(o *options) {
		o.ignorePaths = paths
	}
}

// MaskFields is an Option that masks certain fields
func MaskFields(mask string, fields []string) Option {
	return Transformers(internal.MaskFields(mask, fields))
}

// MaskPaths is an Option that masks certain paths
func MaskPaths(mask string, paths []string) Option {
	return Transformers(internal.MaskPaths(mask, paths))
}

// Define a struct for a Masker
type Masker struct {
	mask string
}

// NewMask is a constructor for the Masker struct
func NewMask(mask string) *Masker {
	return &Masker{mask: mask}
}

// MaskFields is a method on Masker that masks certain fields
func (m *Masker) MaskFields(fields ...string) Option {
	return MaskFields(m.mask, fields)
}

// MaskPaths is a method on Masker that masks certain paths
func (m *Masker) MaskPaths(paths ...string) Option {
	return MaskPaths(m.mask, paths)
}

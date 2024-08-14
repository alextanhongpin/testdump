package grpcdump

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/alextanhongpin/testdump/grpcdump/internal"
)

const env = "TESTDUMP"

// Option is a type that defines a function that modifies an options object.
// The function takes a pointer to an options object and does not return any value.
type Option func(o *options)

type options struct {
	cmpOpt       CompareOption
	colors       bool
	env          string
	transformers []func(*GRPC) error
}

func newOptions() *options {
	return &options{
		colors: true,
		env:    env,
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
		opt:    o.cmpOpt,
		colors: o.colors,
	}
}

// IgnoreMetadata is a function that returns an Option.
// This Option, when applied, configures the options object to ignore certain metadata keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreMetadata(keys ...string) Option {
	return func(o *options) {
		o.cmpOpt.Metadata = append(o.cmpOpt.Metadata, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreTrailer is a function that returns an Option.
// This Option, when applied, configures the options object to ignore certain trailer keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreTrailer(keys ...string) Option {
	return func(o *options) {
		o.cmpOpt.Trailer = append(o.cmpOpt.Trailer, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreHeader is a function that returns an Option.
// This Option, when applied, configures the options object to ignore certain header keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreHeader(keys ...string) Option {
	return func(o *options) {
		// Append the keys to be ignored to the Header field of the cmpOpt object.
		o.cmpOpt.Header = append(o.cmpOpt.Header, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreMessageFields is a function that returns an Option.
// This Option, when applied, configures the options object to ignore certain message fields.
// The fields to ignore are provided as arguments to the function.
func IgnoreMessageFields(keys ...string) Option {
	return func(o *options) {
		// Append the fields to be ignored to the Message field of the cmpOpt object.
		o.cmpOpt.Message = append(o.cmpOpt.Message, internal.IgnoreMapEntries(keys...))
	}
}

// MaskMetadata is a function that returns an Option.
// This Option, when applied, configures the options object to mask certain metadata keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskMetadata(mask string, keys []string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			md, err := maskMetadata(g.Metadata, mask, keys)
			if err != nil {
				return err
			}
			g.Metadata = md
			return nil
		})
	}
}

// MaskTrailer is a function that returns an Option.
// This Option, when applied, configures the options object to mask certain trailer keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskTrailer(mask string, keys []string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			md, err := maskMetadata(g.Trailer, mask, keys)
			if err != nil {
				return err
			}
			g.Trailer = md
			return nil
		})
	}
}

// MaskHeader is a function that returns an Option.
// This Option, when applied, configures the options object to mask certain header keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskHeader(mask string, keys []string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			// Apply the mask to the header metadata.
			md, err := maskMetadata(g.Header, mask, keys)
			if err != nil {
				return err // Return the error if the mask cannot be applied.
			}
			g.Header = md // Update the header metadata with the masked values.
			return nil
		})
	}
}

// MaskMessageFields is a function that returns an Option.
// This Option, when applied, configures the options object to mask certain message fields.
// The mask and the fields to mask are provided as arguments to the function.
func MaskMessageFields(mask string, fields []string) Option {
	masker := internal.MaskFields(mask, fields) // Create a masker function.
	return func(o *options) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			// Apply the mask to each message in the gRPC object.
			for i, msg := range g.Messages {
				b, err := json.Marshal(msg.Message) // Convert the message to JSON.
				if err != nil {
					return err // Return the error if the message cannot be converted to JSON.
				}

				b, err = masker(b) // Apply the mask to the JSON message.
				if err != nil {
					return err // Return the error if the mask cannot be applied.
				}
				var a any
				if err := json.Unmarshal(b, &a); err != nil { // Convert the masked JSON message back to a message object.
					return err // Return the error if the masked JSON message cannot be converted back to a message object.
				}
				g.Messages[i].Message = a // Update the message with the masked values.
			}

			return nil
		})
	}
}

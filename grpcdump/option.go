package grpcdump

import (
	"encoding/json"

	"github.com/alextanhongpin/testdump/grpcdump/internal"
)

// Option is a type that defines a function that modifies an option object.
// The function takes a pointer to an option object and does not return any value.
type Option func(o *option)

type option struct {
	transformers []Transformer
	cmpOpt       CompareOption
	env          string
	colors       bool
}

func newOption(opts ...Option) *option {
	o := new(option)
	o.colors = true

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// IgnoreMetadata is a function that returns an Option.
// This Option, when applied, configures the option object to ignore certain metadata keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreMetadata(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Metadata = append(o.cmpOpt.Metadata, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreTrailer is a function that returns an Option.
// This Option, when applied, configures the option object to ignore certain trailer keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreTrailer(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Trailer = append(o.cmpOpt.Trailer, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreHeader is a function that returns an Option.
// This Option, when applied, configures the option object to ignore certain header keys.
// The keys to ignore are provided as arguments to the function.
func IgnoreHeader(keys ...string) Option {
	return func(o *option) {
		// Append the keys to be ignored to the Header field of the cmpOpt object.
		o.cmpOpt.Header = append(o.cmpOpt.Header, internal.IgnoreMapEntries(keys...))
	}
}

// IgnoreMessageFields is a function that returns an Option.
// This Option, when applied, configures the option object to ignore certain message fields.
// The fields to ignore are provided as arguments to the function.
func IgnoreMessageFields(keys ...string) Option {
	return func(o *option) {
		// Append the fields to be ignored to the Message field of the cmpOpt object.
		o.cmpOpt.Message = append(o.cmpOpt.Message, internal.IgnoreMapEntries(keys...))
	}
}

// MaskMetadata is a function that returns an Option.
// This Option, when applied, configures the option object to mask certain metadata keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskMetadata(mask string, keys []string) Option {
	return func(o *option) {
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
// This Option, when applied, configures the option object to mask certain trailer keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskTrailer(mask string, keys []string) Option {
	return func(o *option) {
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
// This Option, when applied, configures the option object to mask certain header keys.
// The mask and the keys to mask are provided as arguments to the function.
func MaskHeader(mask string, keys []string) Option {
	return func(o *option) {
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
// This Option, when applied, configures the option object to mask certain message fields.
// The mask and the fields to mask are provided as arguments to the function.
func MaskMessageFields(mask string, fields []string) Option {
	masker := internal.MaskFields(mask, fields) // Create a masker function.
	return func(o *option) {
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

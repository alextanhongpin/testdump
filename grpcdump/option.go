package grpcdump

import (
	"encoding/json"

	"github.com/alextanhongpin/testdump/grpcdump/internal"
)

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

func IgnoreMetadata(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Metadata = append(o.cmpOpt.Metadata, internal.IgnoreMapEntries(keys...))
	}
}

func IgnoreTrailer(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Trailer = append(o.cmpOpt.Trailer, internal.IgnoreMapEntries(keys...))
	}
}

func IgnoreHeader(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Header = append(o.cmpOpt.Header, internal.IgnoreMapEntries(keys...))
	}
}

func IgnoreMessageFields(keys ...string) Option {
	return func(o *option) {
		o.cmpOpt.Message = append(o.cmpOpt.Message, internal.IgnoreMapEntries(keys...))
	}
}

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

func MaskHeader(mask string, keys []string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			md, err := maskMetadata(g.Header, mask, keys)
			if err != nil {
				return err
			}
			g.Header = md
			return nil
		})
	}
}

func MaskMessageFields(mask string, fields []string) Option {
	masker := internal.MaskFields(mask, fields)
	return func(o *option) {
		o.transformers = append(o.transformers, func(g *GRPC) error {
			for i, msg := range g.Messages {
				b, err := json.Marshal(msg.Message)
				if err != nil {
					return err
				}

				b, err = masker(b)
				if err != nil {
					return err
				}
				var a any
				if err := json.Unmarshal(b, &a); err != nil {
					return err
				}
				g.Messages[i].Message = a
			}

			return nil
		})
	}
}

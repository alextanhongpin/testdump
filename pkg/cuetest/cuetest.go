package cuetest

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type Validator struct {
	Schemas     []string
	SchemaPaths []string
	Options     []cue.Option
}

func (o *Validator) Validate(data []byte) error {
	bothEmpty := len(o.Schemas)+len(o.SchemaPaths) == 0
	if bothEmpty {
		// Nothing to validate.
		return nil
	}

	if len(o.Options) == 0 {
		// Concrete allows us to check for required fields.
		o.Options = append(o.Options, cue.Concrete(true))
	}

	ctx := cuecontext.New()

	var schema cue.Value

	// Load string schemas.
	for _, s := range o.Schemas {
		schema = schema.Unify(ctx.CompileString(s))
	}

	// Load path schemas.
	if len(o.SchemaPaths) > 0 {
		bins := load.Instances(o.SchemaPaths, nil)
		values, err := ctx.BuildInstances(bins)
		if err != nil {
			return err
		}

		for _, v := range values {
			schema = schema.Unify(v)
		}
	}

	value := ctx.CompileString(string(data))
	return schema.Unify(value).Validate(o.Options...)
}

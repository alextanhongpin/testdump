package textdump

type Option func(o *option)

type option struct {
	transformers []Transformer
	env          string
	colors       bool
	file         string
}

type Transformer func(b []byte) ([]byte, error)

func Transformers(t ...Transformer) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, t...)
	}
}

func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

func Env(env string) Option {
	return func(o *option) {
		o.env = env
	}
}

func File(file string) Option {
	return func(o *option) {
		o.file = file
	}
}

func newOption(opts ...Option) *option {
	opt := new(option)
	opt.colors = true

	for _, o := range opts {
		o(opt)
	}

	return opt
}

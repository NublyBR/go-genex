package genex

type Option func(Generator) Generator

func OptionRNG(rng RNG) Option {
	return func(g Generator) Generator {
		switch cast := g.(type) {
		case *Charset:
			cast.rng = rng
		case *Choice:
			cast.rng = rng
		case *Repeat:
			cast.rng = rng
		case *Numeric:
			cast.rng = rng
		}

		return g
	}
}

func optionApply(opts ...Option) func(Generator) Generator {
	return func(g Generator) Generator {
		for _, o := range opts {
			g = o(g)
		}
		return g
	}
}

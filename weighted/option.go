package weighted

type drawOptions struct {
	roller Roller
}
type Option func(*drawOptions)

func WithRand(r Roller) Option {
	return func(o *drawOptions) {
		o.roller = r
	}
}

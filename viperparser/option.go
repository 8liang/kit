package viperparser

type Option func(p *Parser)

func WithUrl(url string) Option {
	return func(p *Parser) {
		p.url = url
	}
}

func WithEnvPrefix(envPrefix string) Option {
	return func(p *Parser) {
		p.envPrefix = envPrefix
	}
}

func WithEnvFile(envFile string) Option {
	return func(p *Parser) {
		p.envFile = envFile
	}
}

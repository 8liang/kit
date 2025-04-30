package protobuf

import "path"

type Config struct {
	includePaths []string
	getOutPath   GetOutPath
}
type Option func(c *Config)

type GetOutPath func(dir string) string

func WithGetOutPath(getOutPath GetOutPath) Option {
	return func(c *Config) {
		c.getOutPath = getOutPath
	}
}

func WithIncludePaths(includePaths ...string) Option {
	return func(c *Config) {
		c.includePaths = includePaths
	}
}

func DefaultGetOutPath(dir string) string {
	return path.Dir(dir)
}

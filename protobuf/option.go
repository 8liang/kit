package protobuf

import "path"

type Config struct {
	includePaths []string
	goPath       string
	getOutPath   GetOutPath
	grain        bool
	debug        bool
}
type Option func(c *Config)

type GetOutPath func(filePath string) string

func WithGetOutPath(getOutPath GetOutPath) Option {
	return func(c *Config) {
		c.getOutPath = getOutPath
	}
}

func WithIncludePaths(includePaths ...string) Option {
	return func(c *Config) {
		c.includePaths = append(c.includePaths, includePaths...)
	}
}

func WithDebug() Option {
	return func(c *Config) {
		c.debug = true
	}
}

func WithGrain() Option {
	return func(c *Config) {
		c.grain = true
	}
}

func DefaultGetOutPath(filePath string) string {
	return path.Dir(filePath)
}

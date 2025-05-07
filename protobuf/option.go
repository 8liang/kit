package protobuf

import (
	"os"
	"path"
)

type Config struct {
	includePaths []string
	goPath       string
	getOutPath   GetOutPath
	grain        bool
	injectTag    bool
	debug        bool
}

func newDefaultConfig() *Config {
	cfg := &Config{
		getOutPath: DefaultGetOutPath,
		goPath:     os.Getenv("GOPATH"),
	}
	cfg.includePaths = append(cfg.includePaths, ".", path.Join(cfg.goPath, "src"))
	return cfg
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

func WithInjectTag() Option {
	return func(c *Config) {
		c.injectTag = true
	}
}

func DefaultGetOutPath(filePath string) string {
	return path.Dir(filePath)
}

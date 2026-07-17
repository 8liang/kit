package protobuf

import (
	"os"
	"path"
)

// Plugin describes a protoc plugin to invoke.
// Plugin 描述一个需要调用的 protoc 插件。
type Plugin struct {
	Name    string // plugin name, e.g. "go-grain", "es" — protoc-gen-{Name}
	OutPath string // output directory for --{Name}_out
	Module  string // Go module for --{Name}_opt=module={Module}
}

type Config struct {
	includePaths []string
	goPath       string
	getOutPath   GetOutPath
	plugins      []Plugin
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

// WithPlugin registers a protoc plugin.
// WithPlugin 注册一个 protoc 插件。
// name: plugin name, e.g. "go-grain" → protoc-gen-go-grain, --go-grain_out, --go-grain_opt=module=...
func WithPlugin(name, outPath, module string) Option {
	return func(c *Config) {
		c.plugins = append(c.plugins, Plugin{Name: name, OutPath: outPath, Module: module})
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

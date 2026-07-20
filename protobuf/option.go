package protobuf

import (
	"os"
	"path"
)

// Plugin describes a protoc plugin to invoke.
// Plugin 描述一个需要调用的 protoc 插件。
type Plugin struct {
	Name       string   // plugin name, e.g. "go-grain", "es" - protoc-gen-{Name}
	OutPath    string   // output directory for --{Name}_out
	Module     string   // Go module for --{Name}_opt=module={Module}
	Binary     string   // optional full path to the plugin binary; 空则按 LookPath / GOPATH/bin 解析
	ExtraOpts  []string // extra --{Name}_opt= options, e.g. "target=ts"
}

// PluginFilter decides whether a plugin should run for a given proto file.
// PluginFilter 决定某个插件是否对指定 proto 文件生效，返回 false 则跳过该文件。
type PluginFilter func(protoFile string, plugin Plugin) bool

type Config struct {
	includePaths []string
	goPath       string
	goModule     string
	getOutPath   GetOutPath
	plugins      []Plugin
	pluginFilter PluginFilter
	injectTag    bool
	nonRecursive bool
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

// WithGoModule sets the Go module prefix used for --go_opt=module=.
// WithGoModule 设置 Go module 前缀，用于 --go_opt=module=，使输出按 go_package 相对 module 分目录。
func WithGoModule(module string) Option {
	return func(c *Config) {
		c.goModule = module
	}
}

// WithPlugin registers a protoc plugin.
// WithPlugin 注册一个 protoc 插件。
// name: plugin name, e.g. "go-grain" -> protoc-gen-go-grain, --go-grain_out, --go-grain_opt=module=...
func WithPlugin(name, outPath, module string) Option {
	return WithPluginConfig(Plugin{Name: name, OutPath: outPath, Module: module})
}

// WithPluginBinary registers a protoc plugin with an explicit binary path.
// WithPluginBinary 注册一个带显式二进制路径的 protoc 插件，适用于不在 GOPATH/bin 的插件（如 npm 安装的 protoc-gen-es）。
func WithPluginBinary(name, binary, outPath, module string) Option {
	return WithPluginConfig(Plugin{Name: name, OutPath: outPath, Module: module, Binary: binary})
}

// WithPluginConfig registers a fully-specified protoc plugin (binary, module, extra opts).
// WithPluginConfig 注册一个完整配置的 protoc 插件（含 binary、module、额外 opt）。
func WithPluginConfig(p Plugin) Option {
	return func(c *Config) {
		c.plugins = append(c.plugins, p)
	}
}

// WithPluginFilter attaches a filter that selectively enables plugins per proto file.
// WithPluginFilter 附加过滤器，按 proto 文件选择性启用插件。
func WithPluginFilter(filter PluginFilter) Option {
	return func(c *Config) {
		c.pluginFilter = filter
	}
}

// WithNonRecursive makes findProtoFiles only scan the top-level directory.
// WithNonRecursive 使 findProtoFiles 只扫描顶层目录，不递归子目录。
func WithNonRecursive() Option {
	return func(c *Config) {
		c.nonRecursive = true
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

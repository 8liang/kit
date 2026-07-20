package protobuf

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/spf13/afero"
)

// KnownPlugins lists the protoc plugins that proto setup checks for.
// KnownPlugins 列出 proto setup 会检查的 protoc 插件。
var KnownPlugins = []string{"go", "go-grain", "es"}

// SetupConfig holds all configuration for proto setup.
// SetupConfig 保存 proto setup 的全部配置。
type SetupConfig struct {
	install       bool
	cacheDir      string
	protocVersion string
	ignoreImports map[string]bool
	nonRecursive  bool
	verbose       bool
	fs            afero.Fs
	httpClient    *http.Client
}

// newSetupConfig returns a SetupConfig with sensible defaults.
// newSetupConfig 返回带合理默认值的 SetupConfig。
func newSetupConfig() *SetupConfig {
	return &SetupConfig{
		protocVersion: "latest",
		ignoreImports: make(map[string]bool),
		fs:            afero.NewOsFs(),
		httpClient:    http.DefaultClient,
	}
}

// SetupOption is a functional option for Setup.
// SetupOption 是 Setup 的函数式选项。
type SetupOption func(*SetupConfig)

// SetupWithInstall enables automatic installation of missing dependencies.
// SetupWithInstall 启用自动安装缺失依赖。
func SetupWithInstall() SetupOption {
	return func(c *SetupConfig) { c.install = true }
}

// SetupWithCacheDir sets the cache root directory (default: <protoDir>/.proto-cache).
// SetupWithCacheDir 设置缓存根目录。
func SetupWithCacheDir(dir string) SetupOption {
	return func(c *SetupConfig) { c.cacheDir = dir }
}

// SetupWithProtocVersion sets the target protoc version (default: "latest").
// SetupWithProtocVersion 设置目标 protoc 版本。
func SetupWithProtocVersion(version string) SetupOption {
	return func(c *SetupConfig) { c.protocVersion = version }
}

// SetupWithIgnoreImport skips the given import paths during dependency resolution.
// SetupWithIgnoreImport 跳过指定 import 路径的依赖解析。
func SetupWithIgnoreImport(paths ...string) SetupOption {
	return func(c *SetupConfig) {
		for _, p := range paths {
			c.ignoreImports[p] = true
		}
	}
}

// SetupWithNonRecursive makes Setup only scan the top-level directory.
// SetupWithNonRecursive 使 Setup 仅扫描顶层目录。
func SetupWithNonRecursive() SetupOption {
	return func(c *SetupConfig) { c.nonRecursive = true }
}

// SetupWithVerbose enables verbose logging.
// SetupWithVerbose 启用详细日志。
func SetupWithVerbose() SetupOption {
	return func(c *SetupConfig) { c.verbose = true }
}

// --- test-only options, 仅测试使用 ---

func setupWithFs(fs afero.Fs) SetupOption {
	return func(c *SetupConfig) { c.fs = fs }
}

func setupWithHTTPClient(client *http.Client) SetupOption {
	return func(c *SetupConfig) { c.httpClient = client }
}

// --- report types, 报告类型 ---

// ToolStatus describes the installation status of a tool (e.g. protoc).
// ToolStatus 描述工具的安装状态。
type ToolStatus struct {
	Found      bool   // 是否已安装
	Path       string // 二进制路径
	Version    string // 版本号
	InstallCmd string // 安装命令建议
}

// PluginStatus describes the installation status of a protoc plugin.
// PluginStatus 描述 protoc 插件的安装状态。
type PluginStatus struct {
	Name       string // 插件名（如 "go", "es"）
	Found      bool   // 是否已安装
	Path       string // 二进制路径
	InstallCmd string // 安装命令建议
}

// ProtoDepStatus describes the status of a single proto import dependency.
// ProtoDepStatus 描述单个 proto import 依赖的状态。
type ProtoDepStatus struct {
	ImportPath string // import 路径，如 "google/protobuf/any.proto"
	Cached     bool   // 是否已在缓存中
	URL        string // 下载 URL（Cached=false 时有效）
	Source     string // 来源: "wkt", "github", ""（空表示无法解析）
	Error      string // 下载错误信息（成功时为空）
}

// SetupReport aggregates all setup check/install results.
// SetupReport 汇总全部 setup 检查/安装结果。
type SetupReport struct {
	Protoc       ToolStatus
	Plugins      []PluginStatus
	ProtoDeps    []ProtoDepStatus
	MissingCount int // 缺失依赖总数
}

// cachePath returns the path within the cache directory.
// cachePath 返回缓存目录内的子路径。
func cachePath(cacheDir, sub string) string {
	return filepath.Join(cacheDir, sub)
}

// Setup detects and optionally installs proto dependencies for the given directory.
// Setup 检测并可选择安装指定目录的 proto 依赖。
func Setup(protoDir string, opts ...SetupOption) (*SetupReport, error) {
	cfg := newSetupConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.cacheDir == "" {
		cfg.cacheDir = filepath.Join(protoDir, ".proto-cache")
	}

	report := &SetupReport{}

	// 1. 扫描 proto 文件
	var fileMap map[string][]string
	var err error
	if cfg.nonRecursive {
		fileMap, err = scanProtoDirNonRecursive(cfg.fs, protoDir)
	} else {
		fileMap, err = scanProtoDir(cfg.fs, protoDir)
	}
	if err != nil {
		return nil, fmt.Errorf("scan proto dir: %w", err)
	}
	imports := deduplicateImports(fileMap)

	// 2. 解析 URL
	resolver := newImportResolver(cfg.protocVersion)
	type depTask struct {
		importPath string
		url        string
		source     string
	}
	var tasks []depTask
	for _, imp := range imports {
		if cfg.ignoreImports[imp] {
			if cfg.verbose {
				fmt.Printf("  跳过: %s (--ignore-import)\n", imp)
			}
			continue
		}
		url, source, err := resolver.resolve(imp)
		if err != nil && cfg.verbose {
			fmt.Printf("  解析失败: %s: %v\n", imp, err)
		}
		tasks = append(tasks, depTask{importPath: imp, url: url, source: source})
	}

	// 3. 下载 proto 依赖
	downloader := newProtoDownloader(cfg.cacheDir, cfg.httpClient, cfg.verbose)
	if cfg.install {
		for _, task := range tasks {
			if task.url == "" {
				report.ProtoDeps = append(report.ProtoDeps, ProtoDepStatus{
					ImportPath: task.importPath,
					Source:     task.source,
				})
				report.MissingCount++
				continue
			}
			if downloader.isCached(task.importPath) {
				report.ProtoDeps = append(report.ProtoDeps, ProtoDepStatus{
					ImportPath: task.importPath,
					Cached:     true,
					Source:     task.source,
				})
				continue
			}
			err := downloader.download(task.importPath, task.url)
			status := ProtoDepStatus{
				ImportPath: task.importPath,
				URL:        task.url,
				Source:     task.source,
				Cached:     err == nil,
			}
			if err != nil {
				status.Error = err.Error()
				report.MissingCount++
			}
			report.ProtoDeps = append(report.ProtoDeps, status)
		}
	} else {
		for _, task := range tasks {
			dep := ProtoDepStatus{
				ImportPath: task.importPath,
				URL:        task.url,
				Source:     task.source,
				Cached:     downloader.isCached(task.importPath),
			}
			if !dep.Cached {
				report.MissingCount++
			}
			report.ProtoDeps = append(report.ProtoDeps, dep)
		}
	}

	// 4. 安装 protoc
	if cfg.install {
		installer := newProtocInstaller(cfg.cacheDir, cfg.protocVersion, cfg.httpClient, cfg.verbose)
		binPath, err := installer.install()
		if err != nil {
			report.Protoc = ToolStatus{InstallCmd: fmt.Sprintf("download from github: %v", err)}
			report.MissingCount++
		} else {
			report.Protoc = ToolStatus{Found: true, Path: binPath, Version: cfg.protocVersion}
		}
	} else {
		report.Protoc = checkProtoc()
		if !report.Protoc.Found {
			report.MissingCount++
		}
	}

	// 5. 检查插件
	for _, name := range KnownPlugins {
		ps := checkPlugin(name)
		if !ps.Found {
			report.MissingCount++
		}
		report.Plugins = append(report.Plugins, ps)
	}

	return report, nil
}

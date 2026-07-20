package protobuf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FormatReport formats a SetupReport as a human-readable string.
// FormatReport 将 SetupReport 格式化为可读字符串。
func FormatReport(r *SetupReport) string {
	var b strings.Builder

	b.WriteString("=== 工具链 ===\n")
	if r.Protoc.Found {
		b.WriteString(fmt.Sprintf("  ✓ protoc %-20s (PATH: %s)\n", r.Protoc.Version, r.Protoc.Path))
	} else {
		b.WriteString("  ✗ protoc\n     → 安装: game-dev-cli proto setup --install\n")
	}

	b.WriteString("=== 插件 ===\n")
	for _, p := range r.Plugins {
		name := "protoc-gen-" + p.Name
		if p.Found {
			b.WriteString(fmt.Sprintf("  ✓ %-25s (PATH: %s)\n", name, p.Path))
		} else {
			b.WriteString(fmt.Sprintf("  ✗ %-25s → 安装: %s\n", name, p.InstallCmd))
		}
	}

	b.WriteString("=== proto 依赖 ===\n")
	for _, d := range r.ProtoDeps {
		if d.Cached {
			b.WriteString(fmt.Sprintf("  ✓ %s\n", d.ImportPath))
		} else if d.Source == "" {
			b.WriteString(fmt.Sprintf("  ? %s (无法自动解析，请用 --include 手动添加)\n", d.ImportPath))
		} else {
			b.WriteString(fmt.Sprintf("  ✗ %s\n     → %s\n", d.ImportPath, d.URL))
		}
	}

	if r.MissingCount == 0 {
		b.WriteString("\n所有依赖就绪 ✓\n")
	} else {
		b.WriteString(fmt.Sprintf("\n%d 个依赖缺失，运行 --install 安装。\n", r.MissingCount))
	}

	return b.String()
}

// checkProtoc detects whether protoc is installed and returns its status.
// checkProtoc 检测 protoc 是否安装并返回状态。
func checkProtoc() ToolStatus {
	status := ToolStatus{}
	path, err := exec.LookPath("protoc")
	if err != nil {
		status.InstallCmd = "game-dev-cli proto setup --install"
		return status
	}
	status.Found = true
	status.Path = path

	out, err := exec.Command("protoc", "--version").CombinedOutput()
	if err == nil {
		version := strings.TrimSpace(string(out))
		version = strings.TrimPrefix(version, "libprotoc ")
		status.Version = version
	}
	return status
}

// checkPlugin detects whether a protoc plugin (protoc-gen-<name>) is installed.
// checkPlugin 检测 protoc-gen-<name> 插件是否安装。
func checkPlugin(name string) PluginStatus {
	status := PluginStatus{Name: name}
	binName := "protoc-gen-" + name
	path, err := exec.LookPath(binName)
	if err == nil {
		status.Found = true
		status.Path = path
		return status
	}
	if gopath := filepath.Join(goPath(), "bin", binName); fileExists(gopath) {
		status.Found = true
		status.Path = gopath
		return status
	}
	status.InstallCmd = installCmdFor(name)
	return status
}

// installCmdFor returns the suggested install command for a plugin.
// installCmdFor 返回插件的建议安装命令。
func installCmdFor(name string) string {
	switch name {
	case "es":
		return "npm install -g @bufbuild/protoc-gen-es"
	default:
		return "go install google.golang.org/protobuf/cmd/protoc-gen-" + name + "@latest"
	}
}

// goPath returns GOPATH or the default $HOME/go.
func goPath() string {
	cmd := exec.Command("go", "env", "GOPATH")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// fileExists checks if a file exists on the real filesystem.
// fileExists 检查真实文件系统中文件是否存在。
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

package protobuf

import (
	"bufio"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

type Summary struct {
	GoPackage string
	OutPath   string
	ProtoFile string
	Args      []string
}

func analyze(af afero.Fs, protoFile string, cfg *Config) (s *Summary, err error) {
	s = &Summary{
		ProtoFile: protoFile,
	}
	if err = s.parsePackage(af, protoFile); err != nil {
		return
	}
	err = s.prepareArgs(cfg)
	return
}

func (s *Summary) parsePackage(af afero.Fs, protoFile string) (err error) {
	var file afero.File
	if file, err = af.Open(protoFile); err != nil {
		return
	}
	defer file.Close()
	re := regexp.MustCompile(`^option\s+go_package\s*=\s*"(.+)";$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			s.GoPackage = matches[1]
			return
		}
	}
	return
}

func (s *Summary) prepareArgs(cfg *Config) (err error) {
	for _, clause := range cfg.includePaths {
		s.Args = append(s.Args, "-I", clause)
	}
	// Go module for --go_opt=module=：优先用 cfg.goModule，否则回退完整 go_package（保持旧行为）。
	// Go module for --go_opt=module=: prefer cfg.goModule, fall back to full go_package (legacy behavior).
	goModule := cfg.goModule
	if goModule == "" {
		goModule = s.GoPackage
	}
	// 基础输出目录由 getOutPath 决定，protoc 会按 go_package 去掉 module 前缀后落到子目录。
	// Base output dir from getOutPath; protoc strips the module prefix from go_package to compute the subdir.
	baseOut := cfg.getOutPath(s.ProtoFile)
	s.Args = append(s.Args,
		"--go_out="+baseOut,
		"--go_opt=module="+goModule,
		"--proto_path="+path.Dir(s.ProtoFile),
	)
	// 记录最终落盘目录，供 inject-tag 使用。
	// Record the final on-disk dir for inject-tag.
	s.OutPath = finalGoOutPath(baseOut, s.GoPackage, goModule)
	for _, p := range cfg.plugins {
		if cfg.pluginFilter != nil && !cfg.pluginFilter(s.ProtoFile, p) {
			continue
		}
		s.Args = append(s.Args, "--plugin="+resolvePluginBinary(cfg, p))
		s.Args = append(s.Args, "--"+p.Name+"_out="+p.OutPath)
		// 仅当插件指定 module 时才传 --<name>_opt=module=（protoc-gen-es 等不认 module opt）。
		// Only emit --<name>_opt=module= when the plugin specifies a module (protoc-gen-es rejects it).
		if p.Module != "" {
			s.Args = append(s.Args, "--"+p.Name+"_opt=module="+p.Module)
		}
		for _, o := range p.ExtraOpts {
			s.Args = append(s.Args, "--"+p.Name+"_opt="+o)
		}
	}
	return
}

// finalGoOutPath computes the directory where protoc-gen-go actually writes files,
// i.e. baseOut joined with the go_package suffix after stripping the module prefix.
// finalGoOutPath 计算 protoc-gen-go 实际落盘目录：baseOut 拼上 go_package 去掉 module 前缀后的相对路径。
func finalGoOutPath(baseOut, goPackage, goModule string) string {
	rel := strings.TrimPrefix(goPackage, goModule)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || rel == goPackage {
		// go_package 不以 module 开头时，protoc 用完整 go_package 作为子目录。
		// When go_package does not start with module, protoc uses the full go_package as the subdir.
		return path.Join(baseOut, path.Clean(goPackage))
	}
	return path.Join(baseOut, rel)
}

// resolvePluginBinary resolves the plugin binary path:
// explicit Binary -> exec.LookPath("protoc-gen-<name>") -> GOPATH/bin/protoc-gen-<name>.
// resolvePluginBinary 解析插件二进制路径：显式 Binary -> PATH 查找 -> GOPATH/bin。
func resolvePluginBinary(cfg *Config, p Plugin) string {
	if p.Binary != "" {
		return p.Binary
	}
	name := "protoc-gen-" + p.Name
	if found, err := exec.LookPath(name); err == nil {
		return found
	}
	return path.Join(cfg.goPath, "bin", name)
}

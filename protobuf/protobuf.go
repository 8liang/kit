package protobuf

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// CheckProtoc verifies that protoc is installed and accessible.
// CheckProtoc 检查 protoc 是否已安装。
func CheckProtoc() error {
	if _, err := exec.LookPath("protoc"); err != nil {
		return fmt.Errorf("protoc not found in PATH: %w\ninstall: https://github.com/protocolbuffers/protobuf/releases", err)
	}
	return nil
}

func Compile(protoDir string, opts ...Option) (err error) {
	var cmds []*exec.Cmd
	if cmds, err = GenerateCommands(afero.NewOsFs(), protoDir, opts...); err != nil {
		return
	}
	for _, cmd := range cmds {
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s\n%s", err, string(out))
		}
	}
	return
}

func GenerateCommands(af afero.Fs, protoDir string, opts ...Option) (cmds []*exec.Cmd, err error) {
	cfg := newDefaultConfig()
	// 自动探测 .proto-cache/ 并追加 -I 和 protoc 路径
	// 从 protoDir 向上遍历祖先目录查找 .proto-cache/，支持分目录 proto gen 调用。
	// Auto-detect .proto-cache/ by walking up from protoDir, so subdir proto gen
	// calls still find the shared cache.
	cacheDir := findProtoCache(af, protoDir)
	if cacheDir != "" {
		if importsDir := filepath.Join(cacheDir, "imports"); dirExists(af, importsDir) {
			cfg.includePaths = append(cfg.includePaths, importsDir)
		}
		if binPath := filepath.Join(cacheDir, "bin", "protoc"); fileExistsFs(af, binPath) {
			cfg.protocPath = binPath
		}
	}
	for _, opt := range opts {
		opt(cfg)
	}
	var files []string
	if files, err = findProtoFiles(af, protoDir, cfg); err != nil {
		return
	}
	// 构造带 GOPATH/bin 的 PATH，让 protoc 能找到 go install 安装的插件（如 protoc-gen-go）
	// Build PATH with GOPATH/bin prepended so protoc finds go-installed plugins.
	gobin := filepath.Join(goPathOrDefault(cfg.goPath), "bin")
	pathEnv := "PATH=" + gobin + ":" + os.Getenv("PATH")
	env := replaceEnv(os.Environ(), pathEnv)
	outPath := map[string]struct{}{}
	for _, file := range files {
		var s *Summary
		if s, err = analyze(af, file, cfg); err != nil {
			return
		}
		protocBin := "protoc"
		if cfg.protocPath != "" {
			protocBin = cfg.protocPath
		}
		cmd := exec.Command(protocBin, append(s.Args, s.ProtoFile)...)
		cmd.Env = env
		if cfg.debug {
			fmt.Println(cmd)
		}
		cmds = append(cmds, cmd)
		outPath[s.OutPath] = struct{}{}
	}
	if cfg.injectTag {
		injectTagBin := filepath.Join(gobin, "protoc-go-inject-tag")
		for out := range outPath {
			cmd := exec.Command(injectTagBin, "-input="+out+"/*.pb.go")
			cmd.Env = env
			if cfg.debug {
				fmt.Println(cmd)
			}
			cmds = append(cmds, cmd)
		}
	}
	return
}

// replaceEnv returns a copy of env with the key=value entry (matched by prefix) replaced or appended.
// replaceEnv 返回 env 的副本，替换（匹配 prefix）或追加 key=value。
func replaceEnv(env []string, entry string) []string {
	out := make([]string, 0, len(env))
	key := strings.SplitN(entry, "=", 2)[0] + "="
	replaced := false
	for _, e := range env {
		if strings.HasPrefix(e, key) {
			out = append(out, entry)
			replaced = true
		} else {
			out = append(out, e)
		}
	}
	if !replaced {
		out = append(out, entry)
	}
	return out
}

func dirExists(fs afero.Fs, path string) bool {
	info, err := fs.Stat(path)
	return err == nil && info.IsDir()
}

func fileExistsFs(fs afero.Fs, path string) bool {
	info, err := fs.Stat(path)
	return err == nil && !info.IsDir()
}

// findProtoCache walks up from dir to find the nearest .proto-cache directory.
// findProtoCache 从 dir 向上遍历祖先目录查找 .proto-cache。
func findProtoCache(fs afero.Fs, dir string) string {
	for {
		cacheDir := filepath.Join(dir, ".proto-cache")
		if dirExists(fs, cacheDir) {
			return cacheDir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func findProtoFiles(af afero.Fs, protoDir string, cfg *Config) (protoFiles []string, err error) {
	var of afero.File
	if of, err = af.Open(protoDir); err != nil {
		return
	}
	files, err := of.Readdir(-1)
	if err != nil {
		return
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		filePath := path.Join(protoDir, file.Name())
		if file.IsDir() {
			if cfg.nonRecursive {
				continue
			}
			var _files []string
			if _files, err = findProtoFiles(af, filePath, cfg); err != nil {
				return nil, err
			}
			protoFiles = append(protoFiles, _files...)
			continue
		}
		if strings.HasSuffix(file.Name(), ".proto") {
			protoFiles = append(protoFiles, path.Join(protoDir, file.Name()))
		}
	}
	return
}

// goPathOrDefault returns GOPATH or its default ($HOME/go).
// goPathOrDefault 返回 GOPATH 或默认值 ($HOME/go)。
func goPathOrDefault(s string) string {
	if s != "" {
		return s
	}
	return filepath.Join(os.Getenv("HOME"), "go")
}

package protobuf

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/afero"
)

func Export(protoDir string, opts ...Option) (err error) {
	var cmds []*exec.Cmd
	var result []byte
	if cmds, err = GenerateCommands(afero.NewOsFs(), protoDir, opts...); err != nil {
		return
	}
	for _, cmd := range cmds {
		if result, err = cmd.Output(); err != nil {
			return
		}
		fmt.Println(string(result))
	}
	return
}

func NewDefaultConfig() *Config {
	cfg := &Config{
		getOutPath: DefaultGetOutPath,
		goPath:     os.Getenv("GOPATH"),
	}
	cfg.includePaths = append(cfg.includePaths, path.Join(cfg.goPath, "src"))
	return cfg
}

func GenerateCommands(af afero.Fs, protoDir string, opts ...Option) (cmds []*exec.Cmd, err error) {
	cfg := NewDefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	files, err := findProtoFiles(af, protoDir)
	if err != nil {
		return
	}
	for _, file := range files {
		var s *Summary
		if s, err = analyze(af, file, cfg); err != nil {
			return
		}
		cmds = append(cmds, exec.Command("protoc", append(s.Args, s.ProtoFile)...))
	}
	return
}


func findProtoFiles(af afero.Fs, protoDir string) (protoFiles []string, err error) {
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
			var _files []string
			if _files, err = findProtoFiles(af, filePath); err != nil {
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

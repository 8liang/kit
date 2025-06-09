package protobuf

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/afero"
)

func Compile(protoDir string, opts ...Option) (err error) {
	var cmds []*exec.Cmd
	if cmds, err = GenerateCommands(afero.NewOsFs(), protoDir, opts...); err != nil {
		return
	}
	for _, cmd := range cmds {
		if _, err = cmd.Output(); err != nil {
			var _err *exec.ExitError
			if errors.As(err, &_err) {
				return fmt.Errorf("%s,%s", err, string(_err.Stderr))
			}
			return
		}
	}
	return
}

func GenerateCommands(af afero.Fs, protoDir string, opts ...Option) (cmds []*exec.Cmd, err error) {
	cfg := newDefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	var files []string
	if files, err = findProtoFiles(af, protoDir); err != nil {
		return
	}
	outPath := map[string]struct{}{}
	for _, file := range files {
		var s *Summary
		if s, err = analyze(af, file, cfg); err != nil {
			return
		}
		cmd := exec.Command("protoc", append(s.Args, s.ProtoFile)...)
		if cfg.debug {
			fmt.Println(cmd)
		}
		cmds = append(cmds, cmd)
		outPath[s.OutPath] = struct{}{}

	}
	if cfg.injectTag {
		for out := range outPath {
			cmd := exec.Command("protoc-go-inject-tag", "-input="+out+"/*.pb.go")
			if cfg.debug {
				fmt.Println(cmd)
			}
			cmds = append(cmds, cmd)
		}
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

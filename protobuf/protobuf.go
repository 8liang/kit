package protobuf

import (
	"bufio"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

func Export(protoDir string) (err error) {
	var appFs = afero.NewOsFs()
	err = process(appFs, protoDir)
	if err != nil {
		return
	}

	return
}

func NewDefaultConfig() *Config {
	cfg := &Config{
		getOutPath: DefaultGetOutPath,
	}
	cfg.includePaths = append(cfg.includePaths, path.Join(os.Getenv("GOPATH"), "src"))
	return cfg
}

func process(af afero.Fs, protoDir string, opts ...Option) (err error) {
	cfg := NewDefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	files, err := findProtoFiles(af, protoDir)
	if err != nil {
		return
	}
	for _, file := range files {
		var cmd *exec.Cmd
		if cmd, err = protoc(af, file, cfg); err != nil {
			return
		}
		cmd.String()
	}
	return
}

func protoc(af afero.Fs, protoFile string, cfg *Config) (cmd *exec.Cmd, err error) {
	var summary *Summary
	if summary, err = analyze(af, protoFile); err != nil {
		return
	}
	args := []string{}
	for _, clause := range cfg.includePaths {
		args = append(args, "-I", clause)
	}
	outPath := cfg.getOutPath(protoFile)
	args = append(args, "--go_out="+outPath,
		"--go_opt=module="+summary.GoPackage,
		"--proto_path="+outPath,
	)
	return exec.Command("protoc", append(args, protoFile)...), nil
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

func analyze(af afero.Fs, protoFile string) (summary *Summary, err error) {
	var file afero.File
	if file, err = af.Open(protoFile); err != nil {
		return
	}
	defer file.Close()
	summary = &Summary{}
	re := regexp.MustCompile(`^option\s+go_package\s*=\s*"(.+)";$`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			summary.GoPackage = matches[1]
			return
		}
	}
	return
}

type Summary struct {
	GoPackage string
}

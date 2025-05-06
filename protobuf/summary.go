package protobuf

import (
	"bufio"
	"path"
	"regexp"

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
	s.OutPath = cfg.getOutPath(s.ProtoFile)
	s.Args = append(s.Args, "--go_out="+s.OutPath,
		"--go_opt=module="+s.GoPackage,
		"--proto_path="+s.OutPath,
		"--plugin="+path.Join(cfg.goPath, "bin", "protoc-gen-go-grain"),
		"--go-grain_out="+s.OutPath,
		"--go-grain_opt=module="+s.GoPackage,
	)
	return
}

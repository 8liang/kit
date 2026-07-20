package protobuf

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// importRe matches `import "path/to/file.proto";` lines.
// importRe 匹配 import 声明行。
var importRe = regexp.MustCompile(`^\s*import\s+"([^"]+)";`)

// parseImportLines extracts all import paths from a single proto file.
// parseImportLines 从单个 proto 文件提取全部 import 路径。
func parseImportLines(fs afero.Fs, filePath string) ([]string, error) {
	f, err := fs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var imports []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := importRe.FindStringSubmatch(line); len(matches) == 2 {
			imports = append(imports, matches[1])
		}
	}
	return imports, scanner.Err()
}

// scanProtoDir recursively finds .proto files under dir and extracts their imports.
// scanProtoDir 递归查找 dir 下的 .proto 文件并提取其 import 声明。
func scanProtoDir(fs afero.Fs, dir string) (map[string][]string, error) {
	return scanDir(fs, dir, false)
}

// scanProtoDirNonRecursive only scans the top-level directory.
// scanProtoDirNonRecursive 仅扫描顶层目录。
func scanProtoDirNonRecursive(fs afero.Fs, dir string) (map[string][]string, error) {
	return scanDir(fs, dir, true)
}

func scanDir(fs afero.Fs, dir string, nonRecursive bool) (map[string][]string, error) {
	result := make(map[string][]string)
	f, err := fs.Open(dir)
	if err != nil {
		return nil, err
	}
	entries, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		fullPath := dir + "/" + e.Name()
		fullPath = strings.ReplaceAll(fullPath, "//", "/")
		if e.IsDir() {
			if nonRecursive {
				continue
			}
			sub, err := scanDir(fs, fullPath, false)
			if err != nil {
				return nil, err
			}
			for k, v := range sub {
				result[k] = v
			}
			continue
		}
		if !strings.HasSuffix(e.Name(), ".proto") {
			continue
		}
		imports, err := parseImportLines(fs, fullPath)
		if err != nil {
			return nil, err
		}
		result[fullPath] = imports
	}
	return result, nil
}

// deduplicateImports returns a sorted list of unique import paths across all files.
// deduplicateImports 返回跨所有文件的去重 import 路径列表。
func deduplicateImports(fileMap map[string][]string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, imports := range fileMap {
		for _, imp := range imports {
			if !seen[imp] {
				seen[imp] = true
				result = append(result, imp)
			}
		}
	}
	return result
}

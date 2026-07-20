package protobuf

import (
	"fmt"
	"strings"
)

// wktFiles is the complete list of Google Well-Known Type proto files.
// wktFiles 是 Google Well-Known Type 的完整 proto 文件列表。
var wktFiles = map[string]bool{
	"google/protobuf/any.proto":             true,
	"google/protobuf/api.proto":             true,
	"google/protobuf/descriptor.proto":      true,
	"google/protobuf/duration.proto":        true,
	"google/protobuf/empty.proto":           true,
	"google/protobuf/field_mask.proto":      true,
	"google/protobuf/source_context.proto":  true,
	"google/protobuf/struct.proto":          true,
	"google/protobuf/timestamp.proto":       true,
	"google/protobuf/type.proto":            true,
	"google/protobuf/wrappers.proto":        true,
}

// isWKT reports whether importPath is a Google Well-Known Type.
// isWKT 判断 importPath 是否为 Google Well-Known Type。
func isWKT(importPath string) bool {
	return wktFiles[importPath]
}

// wktURLTemplate is the base URL for downloading WKT proto files from GitHub.
// wktURLTemplate 是下载 WKT proto 文件的 GitHub raw URL 模板。
const wktURLTemplate = "https://raw.githubusercontent.com/protocolbuffers/protobuf/%s/src/%s"

// importResolver maps proto import paths to download URLs.
// importResolver 将 proto import 路径映射为下载 URL。
type importResolver struct {
	protocVersion string            // protoc 版本，用于 WKT URL
	branchCache   map[string]string // owner/repo -> default branch
}

// newImportResolver creates a new importResolver.
// newImportResolver 创建新的 importResolver。
func newImportResolver(protocVersion string) *importResolver {
	return &importResolver{
		protocVersion: protocVersion,
		branchCache:   make(map[string]string),
	}
}

// resolve maps an import path to a download URL.
// Resolve 将 import 路径映射为下载 URL。
// Returns url="" and source="" when the import cannot be resolved (non-GitHub, non-WKT).
func (r *importResolver) resolve(importPath string) (url string, source string, err error) {
	if isWKT(importPath) {
		version := r.protocVersion
		if version == "" || version == "latest" {
			version = "main" // ponytail: "latest" 应通过 GitHub API 查最新 release tag，此处简化
		}
		url = fmt.Sprintf(wktURLTemplate, version, importPath)
		return url, "wkt", nil
	}
	if strings.HasPrefix(importPath, "github.com/") {
		return r.resolveGitHub(importPath)
	}
	return "", "", nil
}

// resolveGitHub resolves github.com/<owner>/<repo>/<path> style imports.
// resolveGitHub 解析 github.com/<owner>/<repo>/<path> 格式的 import。
func (r *importResolver) resolveGitHub(importPath string) (url string, source string, err error) {
	parts := strings.SplitN(importPath, "/", 4)
	if len(parts) < 4 {
		return "", "", fmt.Errorf("invalid github import path: %s", importPath)
	}
	owner := parts[1]
	repo := parts[2]
	rest := parts[3]

	key := owner + "/" + repo
	branch, ok := r.branchCache[key]
	if !ok {
		branch = "main"
	}
	url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, rest)
	return url, "github", nil
}

// setBranch caches the default branch for a GitHub repo.
// setBranch 缓存 GitHub 仓库的默认分支。
func (r *importResolver) setBranch(ownerRepo, branch string) {
	r.branchCache[ownerRepo] = branch
}

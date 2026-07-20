package protobuf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatReport(t *testing.T) {
	r := &SetupReport{
		Protoc: ToolStatus{
			Found:   true,
			Path:    "/usr/local/bin/protoc",
			Version: "27.3",
		},
		Plugins: []PluginStatus{
			{Name: "go", Found: true, Path: "/go/bin/protoc-gen-go"},
			{Name: "es", Found: false, InstallCmd: "npm install -g @bufbuild/protoc-gen-es"},
		},
		ProtoDeps: []ProtoDepStatus{
			{ImportPath: "google/protobuf/any.proto", Cached: true, Source: "wkt"},
			{ImportPath: "google/protobuf/empty.proto", Cached: false, URL: "https://raw.githubusercontent.com/...", Source: "wkt"},
			{ImportPath: "unknown/registry/file.proto", Cached: false, Source: ""},
		},
		MissingCount: 2,
	}

	out := FormatReport(r)

	assert.Contains(t, out, "=== 工具链 ===")
	assert.Contains(t, out, "protoc 27.3")
	assert.Contains(t, out, "/usr/local/bin/protoc")

	assert.Contains(t, out, "=== 插件 ===")
	assert.Contains(t, out, "protoc-gen-go")
	assert.Contains(t, out, "npm install -g")

	assert.Contains(t, out, "=== proto 依赖 ===")
	assert.Contains(t, out, "google/protobuf/any.proto")
	assert.Contains(t, out, "2 个依赖缺失")
}

func TestCheckPlugin(t *testing.T) {
	status := checkPlugin("nonexistent-xyz-12345")
	assert.False(t, status.Found)
	assert.Equal(t, "nonexistent-xyz-12345", status.Name)
	assert.Contains(t, status.InstallCmd, "go install")
}

func TestFormatReport_AllOK(t *testing.T) {
	r := &SetupReport{
		Protoc:  ToolStatus{Found: true},
		Plugins: []PluginStatus{{Name: "go", Found: true}},
		ProtoDeps: []ProtoDepStatus{
			{ImportPath: "google/protobuf/any.proto", Cached: true},
		},
	}
	out := FormatReport(r)
	assert.Contains(t, out, "所有依赖就绪")
	assert.False(t, strings.Contains(out, "缺失"))
}

package protobuf

import (
	"path"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// newTestFs 构造一个含顶层与子目录 proto 的内存文件系统，用于分目录 / 非递归测试。
// newTestFs builds an in-memory fs with top-level and nested protos for split-dir / non-recursive tests.
func newTestFs(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	cases := map[string]string{
		"/proto/messages.proto": `syntax = "proto3";
package messages;
option go_package = "canlonggame.com/pub-master/pkg/messages";
message Role { string id = 1; }`,
		"/proto/svc/agent.svc.proto": `syntax = "proto3";
package service;
option go_package = "canlonggame.com/pub-master/internal/game/agent/agentsvc";
service Agent { rpc Ping(messages.Base) returns (messages.Base) {} }`,
	}
	for p, c := range cases {
		assert.NoError(t, afero.WriteFile(fs, p, []byte(c), 0644))
	}
	return fs
}

func TestPrepareArgs_GoModule_SplitsByPackage(t *testing.T) {
	fs := newTestFs(t)
	cfg := newDefaultConfig()
	cfg.goModule = "canlonggame.com/pub-master"
	cfg.getOutPath = func(_ string) string { return "/go-project" }

	// 顶层 messages：应落到 pkg/messages 子目录。
	// top-level messages: should land under pkg/messages.
	s, err := analyze(fs, "/proto/messages.proto", cfg)
	assert.NoError(t, err)
	assert.Equal(t, "/go-project/pkg/messages", s.OutPath)
	assert.Contains(t, s.Args, "--go_opt=module=canlonggame.com/pub-master")
	assert.Contains(t, s.Args, "--go_out=/go-project")

	// 子目录 svc：应落到 internal/game/agent/agentsvc 子目录。
	// nested svc: should land under internal/game/agent/agentsvc.
	s2, err := analyze(fs, "/proto/svc/agent.svc.proto", cfg)
	assert.NoError(t, err)
	assert.Equal(t, "/go-project/internal/game/agent/agentsvc", s2.OutPath)
}

func TestPrepareArgs_DefaultModule_Legacy(t *testing.T) {
	// 不设 goModule 时，--go_opt=module= 回退为完整 go_package（旧行为）。
	// Without goModule, --go_opt=module= falls back to the full go_package (legacy).
	fs := newTestFs(t)
	cfg := newDefaultConfig()
	cfg.getOutPath = func(_ string) string { return "/out" }
	s, err := analyze(fs, "/proto/messages.proto", cfg)
	assert.NoError(t, err)
	assert.Contains(t, s.Args, "--go_opt=module=canlonggame.com/pub-master/pkg/messages")
}

func TestPrepareArgs_PluginFilter(t *testing.T) {
	fs := newTestFs(t)
	cfg := newDefaultConfig()
	cfg.goModule = "canlonggame.com/pub-master"
	cfg.getOutPath = func(_ string) string { return "/go-project" }
	cfg.plugins = []Plugin{{Name: "go-grain", OutPath: "/go-project", Module: "canlonggame.com/pub-master"}}
	// 只对 svc 目录的 proto 启用 go-grain。
	// Enable go-grain only for protos under the svc directory.
	cfg.pluginFilter = func(protoFile string, _ Plugin) bool {
		return strings.Contains(protoFile, "/svc/")
	}

	// 顶层 messages：go-grain 被过滤掉。
	// top-level messages: go-grain filtered out.
	s, err := analyze(fs, "/proto/messages.proto", cfg)
	assert.NoError(t, err)
	assert.NotContains(t, s.Args, "--go-grain_out=")

	// svc：go-grain 生效。
	// svc: go-grain enabled.
	s2, err := analyze(fs, "/proto/svc/agent.svc.proto", cfg)
	assert.NoError(t, err)
	assert.Contains(t, s2.Args, "--go-grain_out=/go-project")
}

func TestPrepareArgs_PluginExtraOpts(t *testing.T) {
	fs := newTestFs(t)
	cfg := newDefaultConfig()
	cfg.goModule = "canlonggame.com/pub-master"
	cfg.getOutPath = func(_ string) string { return "/go-project" }
	cfg.plugins = []Plugin{{
		Name:      "es",
		OutPath:   "/ts-project",
		ExtraOpts: []string{"target=ts"},
	}}
	s, err := analyze(fs, "/proto/messages.proto", cfg)
	assert.NoError(t, err)
	// es 不设 module，不应出现 --es_opt=module=；但 target=ts 应出现。
	assert.NotContains(t, s.Args, "--es_opt=module=")
	assert.Contains(t, s.Args, "--es_opt=target=ts")
	assert.Contains(t, s.Args, "--es_out=/ts-project")
}

func TestResolvePluginBinary(t *testing.T) {
	cfg := newDefaultConfig()
	// 显式 Binary 优先。
	// Explicit Binary wins.
	assert.Equal(t, "/npm/bin/protoc-gen-es",
		resolvePluginBinary(cfg, Plugin{Name: "es", Binary: "/npm/bin/protoc-gen-es"}))
	// Binary 空 + LookPath 失败时回退到 GOPATH/bin。
	// Empty Binary + LookPath miss falls back to GOPATH/bin.
	got := resolvePluginBinary(cfg, Plugin{Name: "nonexistent-xyz"})
	assert.True(t, strings.HasSuffix(got, path.Join("bin", "protoc-gen-nonexistent-xyz")),
		"expected GOPATH/bin fallback, got %s", got)
}

func TestFindProtoFiles_NonRecursive(t *testing.T) {
	fs := newTestFs(t)
	cfg := newDefaultConfig()
	cfg.nonRecursive = true
	// 非递归只收集顶层 messages.proto，不进 svc/。
	// Non-recursive collects only top-level messages.proto, skipping svc/.
	files, err := findProtoFiles(fs, "/proto", cfg)
	assert.NoError(t, err)
	assert.Equal(t, []string{"/proto/messages.proto"}, files)

	// 递归（默认）则收集全部。
	// Recursive (default) collects all.
	cfg.nonRecursive = false
	all, err := findProtoFiles(fs, "/proto", cfg)
	assert.NoError(t, err)
	assert.Len(t, all, 2)
}

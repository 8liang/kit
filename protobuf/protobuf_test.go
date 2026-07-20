package protobuf

import (
	"fmt"
	"os"
	"testing"

	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var tesFs afero.Fs
var testCases = map[string]string{
	"/kit/protobuf/web/websvc/web.svc.proto": `syntax = "proto3";
package message;
option go_package = "github.com/8liang/kit/protobuf/web/websvc";
import "google/protobuf/empty.proto";

service WebSvc {
    rpc Start(google.protobuf.Empty)returns(google.protobuf.Empty){}
}`,
	"/kit/protobuf/role.proto": `syntax = "proto3";
package message;
option go_package = "github.com/8liang/kit/protobuf/role";
message Role {
    string id = 1;
    string name = 2;
    string description = 3;
}`,
}

func TestProcess(t *testing.T) {
	c, err := GenerateCommands(tesFs, "/")
	assert.NoError(t, err)
	fmt.Println(c)
}

func TestFindProtoFiles(t *testing.T) {
	files, err := findProtoFiles(tesFs, "/kit", newDefaultConfig())
	assert.NoError(t, err)
	assert.Equal(t, len(testCases), len(files))
	for _, file := range files {
		assert.Contains(t, testCases, file)
	}
}

func TestMain(m *testing.M) {
	tesFs = afero.NewMemMapFs()
	for path, content := range testCases {
		afero.WriteFile(tesFs, path, []byte(content), 0644)
	}
	os.Exit(m.Run())
}

func TestAnalyze(t *testing.T) {
	summary, err := analyze(tesFs, "/kit/protobuf/web/websvc/web.svc.proto", newDefaultConfig())
	assert.NoError(t, err)
	assert.Equal(t, "github.com/8liang/kit/protobuf/web/websvc", summary.GoPackage)
	spew.Dump(summary.Args)
}

func TestGenerateCommands_WithProtoCache(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/proto/.proto-cache/imports/google/protobuf", 0755)
	afero.WriteFile(fs, "/proto/.proto-cache/imports/google/protobuf/any.proto",
		[]byte("syntax = \"proto3\";"), 0644)

	afero.WriteFile(fs, "/proto/test.proto", []byte(`
syntax = "proto3";
option go_package = "github.com/test/pkg";
import "google/protobuf/any.proto";
`), 0644)

	cmds, err := GenerateCommands(fs, "/proto")
	assert.NoError(t, err)
	assert.NotEmpty(t, cmds)

	args := strings.Join(cmds[0].Args, " ")
	assert.Contains(t, args, "-I")
	assert.Contains(t, args, ".proto-cache/imports")
}

func TestFindProtoCache(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/project/proto/.proto-cache/imports", 0755)

	assert.Equal(t, "/project/proto/.proto-cache", findProtoCache(fs, "/project/proto"))
	assert.Equal(t, "/project/proto/.proto-cache", findProtoCache(fs, "/project/proto/sub"))
	assert.Equal(t, "/project/proto/.proto-cache", findProtoCache(fs, "/project/proto/sub/deep"))
	assert.Equal(t, "", findProtoCache(fs, "/project"))
	assert.Equal(t, "", findProtoCache(fs, "/"))
}

package protobuf

import (
	"fmt"
	"os"
	"testing"

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
	files, err := findProtoFiles(tesFs, "/kit")
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

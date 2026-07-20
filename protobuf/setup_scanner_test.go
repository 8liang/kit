package protobuf

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestParseImportLines(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/a.proto", []byte(`
syntax = "proto3";
import "google/protobuf/any.proto";
import "other/pkg.proto";
import "google/protobuf/empty.proto";
// import "commented/out.proto";
`), 0644)

	imports, err := parseImportLines(fs, "/proto/a.proto")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"google/protobuf/any.proto",
		"other/pkg.proto",
		"google/protobuf/empty.proto",
	}, imports)
}

func TestParseImportLines_NoImports(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/empty.proto", []byte(`syntax = "proto3";
package foo;
`), 0644)

	imports, err := parseImportLines(fs, "/proto/empty.proto")
	assert.NoError(t, err)
	assert.Empty(t, imports)
}

func TestScanProtoDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/a.proto", []byte(`import "google/protobuf/any.proto";`), 0644)
	afero.WriteFile(fs, "/proto/sub/b.proto", []byte(`import "google/protobuf/empty.proto";`), 0644)
	afero.WriteFile(fs, "/proto/ignore.txt", []byte(`not a proto`), 0644)

	result, err := scanProtoDir(fs, "/proto")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, []string{"google/protobuf/any.proto"}, result["/proto/a.proto"])
	assert.Equal(t, []string{"google/protobuf/empty.proto"}, result["/proto/sub/b.proto"])
}

func TestScanProtoDir_NonRecursive(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/a.proto", []byte(`import "google/protobuf/any.proto";`), 0644)
	afero.WriteFile(fs, "/proto/sub/b.proto", []byte(`import "google/protobuf/empty.proto";`), 0644)

	result, err := scanProtoDirNonRecursive(fs, "/proto")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, "/proto/a.proto")
}

func TestDeduplicateImports(t *testing.T) {
	fileMap := map[string][]string{
		"/a.proto": {"google/protobuf/any.proto", "other/pkg.proto"},
		"/b.proto": {"google/protobuf/any.proto", "third/pkg.proto"},
	}
	unique := deduplicateImports(fileMap)
	assert.ElementsMatch(t, []string{
		"google/protobuf/any.proto",
		"other/pkg.proto",
		"third/pkg.proto",
	}, unique)
}

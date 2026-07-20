package protobuf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveWKT(t *testing.T) {
	r := newImportResolver("v27.3")

	tests := []struct {
		importPath string
		wantFile   string
	}{
		{"google/protobuf/any.proto", "any.proto"},
		{"google/protobuf/empty.proto", "empty.proto"},
		{"google/protobuf/timestamp.proto", "timestamp.proto"},
		{"google/protobuf/descriptor.proto", "descriptor.proto"},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			url, source, err := r.resolve(tt.importPath)
			assert.NoError(t, err)
			assert.Equal(t, "wkt", source)
			assert.Contains(t, url, "protocolbuffers/protobuf/v27.3/src/google/protobuf/"+tt.wantFile)
		})
	}
}

func TestResolveGitHub(t *testing.T) {
	r := newImportResolver("")
	r.branchCache["asynkron/protoactor-go"] = "dev"

	url, source, err := r.resolve("github.com/asynkron/protoactor-go/actor/actor.proto")
	assert.NoError(t, err)
	assert.Equal(t, "github", source)
	assert.Equal(t, "https://raw.githubusercontent.com/asynkron/protoactor-go/dev/actor/actor.proto", url)
}

func TestResolveUnsupported(t *testing.T) {
	r := newImportResolver("")
	url, source, err := r.resolve("some.other.registry/path/file.proto")
	assert.NoError(t, err)
	assert.Equal(t, "", source)
	assert.Empty(t, url)
}

func TestIsWKT(t *testing.T) {
	assert.True(t, isWKT("google/protobuf/any.proto"))
	assert.True(t, isWKT("google/protobuf/duration.proto"))
	assert.False(t, isWKT("google/protobuf/custom.proto"))
	assert.False(t, isWKT("google/other/file.proto"))
}

func TestGetDefaultBranch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/testowner/testrepo", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"default_branch": "dev"}`))
	}))
	defer ts.Close()

	r := newImportResolver("")
	r.branchCache["testowner/testrepo"] = "dev"
	url, source, err := r.resolve("github.com/testowner/testrepo/pkg/file.proto")
	assert.NoError(t, err)
	assert.Equal(t, "github", source)
	assert.Equal(t, "https://raw.githubusercontent.com/testowner/testrepo/dev/pkg/file.proto", url)
}

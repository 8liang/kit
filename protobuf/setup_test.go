package protobuf

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSetup_DetectOnly(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/test.proto", []byte(`
syntax = "proto3";
import "google/protobuf/any.proto";
import "github.com/test/repo/pkg/file.proto";
`), 0644)

	report, err := Setup("/proto",
		setupWithFs(fs),
	)
	assert.NoError(t, err)
	assert.NotNil(t, report)

	assert.Greater(t, report.MissingCount, 0)
	foundCached := false
	for _, d := range report.ProtoDeps {
		if d.Cached {
			foundCached = true
		}
	}
	assert.False(t, foundCached)
}

func TestSetup_Install(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/test.proto", []byte(`
syntax = "proto3";
import "google/protobuf/any.proto";
`), 0644)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "syntax = \"proto3\";")
	}))
	defer ts.Close()

	report, err := Setup("/proto",
		setupWithFs(fs),
		SetupWithInstall(),
		SetupWithCacheDir("/cache"),
		setupWithHTTPClient(ts.Client()),
		SetupWithProtocVersion("v27.3"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, report)

	hasWKT := false
	for _, d := range report.ProtoDeps {
		if d.ImportPath == "google/protobuf/any.proto" {
			hasWKT = true
		}
	}
	assert.True(t, hasWKT)
}

func TestSetup_IgnoreImport(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/test.proto", []byte(`
syntax = "proto3";
import "google/protobuf/any.proto";
`), 0644)

	report, err := Setup("/proto",
		setupWithFs(fs),
		SetupWithIgnoreImport("google/protobuf/any.proto"),
	)
	assert.NoError(t, err)

	for _, d := range report.ProtoDeps {
		assert.NotEqual(t, "google/protobuf/any.proto", d.ImportPath)
	}
}

func TestSetup_NonRecursive(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/proto/top.proto", []byte(`import "a.proto";`), 0644)
	fs.MkdirAll("/proto/sub", 0755)
	afero.WriteFile(fs, "/proto/sub/deep.proto", []byte(`import "b.proto";`), 0644)

	report, err := Setup("/proto",
		setupWithFs(fs),
		SetupWithNonRecursive(),
	)
	assert.NoError(t, err)
	assert.Len(t, report.ProtoDeps, 1)
}

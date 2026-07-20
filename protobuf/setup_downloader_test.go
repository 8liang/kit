package protobuf

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestDownloader_IsCached(t *testing.T) {
	fs := afero.NewMemMapFs()
	cacheDir := "/cache"
	importsDir := filepath.Join(cacheDir, "imports")
	fs.MkdirAll(filepath.Join(importsDir, "google/protobuf"), 0755)
	afero.WriteFile(fs, filepath.Join(importsDir, "google/protobuf/any.proto"), []byte("syntax = \"proto3\";"), 0644)

	d := &protoDownloader{cacheDir: cacheDir, fs: fs, client: http.DefaultClient}
	assert.True(t, d.isCached("google/protobuf/any.proto"))
	assert.False(t, d.isCached("google/protobuf/empty.proto"))
}

func TestDownloader_Download(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "syntax = \"proto3\";\npackage test;\n")
	}))
	defer ts.Close()

	fs := afero.NewMemMapFs()
	cacheDir := "/cache"
	fs.MkdirAll(cacheDir, 0755)

	d := &protoDownloader{cacheDir: cacheDir, fs: fs, client: ts.Client()}
	err := d.download("google/protobuf/any.proto", ts.URL)
	assert.NoError(t, err)

	expectedPath := filepath.Join(cacheDir, "imports", "google/protobuf/any.proto")
	data, err := afero.ReadFile(fs, expectedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "package test;")

	assert.True(t, d.isCached("google/protobuf/any.proto"))
}

func TestDownloader_DownloadError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	fs := afero.NewMemMapFs()
	cacheDir := "/cache"
	fs.MkdirAll(cacheDir, 0755)

	d := &protoDownloader{cacheDir: cacheDir, fs: fs, client: ts.Client()}
	err := d.download("google/protobuf/any.proto", ts.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDownloader_RealFS(t *testing.T) {
	if os.Getenv("SKIP_REAL_FS") != "" {
		t.Skip("skipping real filesystem test")
	}
	tmpDir := t.TempDir()
	d := newProtoDownloader(tmpDir, afero.NewOsFs(), http.DefaultClient, false)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "syntax = \"proto3\";")
	}))
	defer ts.Close()

	err := d.download("test/file.proto", ts.URL)
	assert.NoError(t, err)

	assert.True(t, d.isCached("test/file.proto"))
}

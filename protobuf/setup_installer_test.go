package protobuf

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func createFakeProtocZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("protoc/bin/protoc")
	assert.NoError(t, err)
	fmt.Fprint(f, "#!/bin/sh\necho libprotoc 27.3\n")
	zw.Close()
	return buf.Bytes()
}

func TestProtocInstaller_Install(t *testing.T) {
	zipData := createFakeProtocZip(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(zipData)
	}))
	defer ts.Close()

	fs := afero.NewMemMapFs()
	fs.MkdirAll("/cache", 0755)

	inst := &protocInstaller{
		cacheDir: "/cache",
		version:  "27.3",
		fs:       fs,
		client:   ts.Client(),
	}

	assert.False(t, inst.isInstalled())

	resp, err := ts.Client().Get(ts.URL + "/fake.zip")
	assert.NoError(t, err)
	defer resp.Body.Close()

	err = inst.extractProtoc(resp.Body)
	assert.NoError(t, err)
	assert.True(t, inst.isInstalled())
}

func TestProtocInstaller_OSArchString(t *testing.T) {
	s := osArchString()
	assert.NotEmpty(t, s)
}

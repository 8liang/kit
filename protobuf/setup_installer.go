package protobuf

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/afero"
)

// protocInstaller downloads and extracts the protoc binary.
// protocInstaller 下载并解压 protoc 二进制。
type protocInstaller struct {
	cacheDir string
	version  string
	fs       afero.Fs
	client   *http.Client
	verbose  bool
}

// newProtocInstaller creates a new protocInstaller.
// newProtocInstaller 创建新的 protocInstaller。
func newProtocInstaller(cacheDir, version string, client *http.Client, verbose bool) *protocInstaller {
	return &protocInstaller{
		cacheDir: cacheDir,
		version:  version,
		fs:       afero.NewOsFs(),
		client:   client,
		verbose:  verbose,
	}
}

// isInstalled checks if protoc already exists in the cache.
// isInstalled 检查 protoc 是否已在缓存中。
func (i *protocInstaller) isInstalled() bool {
	bin := i.protocBinPath()
	info, err := i.fs.Stat(bin)
	return err == nil && !info.IsDir()
}

// protocBinPath returns the expected protoc binary path in the cache.
// protocBinPath 返回缓存中 protoc 二进制的预期路径。
func (i *protocInstaller) protocBinPath() string {
	return filepath.Join(i.cacheDir, "bin", "protoc")
}

// protocReleaseURL returns the GitHub release download URL for the given version.
// protocReleaseURL 返回指定版本的 GitHub release 下载 URL。
func protocReleaseURL(version string) string {
	arch := osArchString()
	return fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
		version, version, arch)
}

// osArchString returns the protoc release asset name suffix like "linux-x86_64".
// osArchString 返回 protoc release 资源名后缀，如 "linux-x86_64"。
func osArchString() string {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	switch osName {
	case "darwin":
		osName = "osx"
	case "windows":
		osName = "win64"
	}

	switch archName {
	case "amd64":
		archName = "x86_64"
	case "arm64":
		archName = "aarch_64"
	case "386":
		archName = "x86_32"
	}

	if osName == "win64" {
		return osName
	}
	return osName + "-" + archName
}

// install downloads and extracts protoc into the cache.
// install 下载并解压 protoc 到缓存。
func (i *protocInstaller) install() (string, error) {
	if i.isInstalled() {
		return i.protocBinPath(), nil
	}

	url := protocReleaseURL(i.version)
	if i.verbose {
		fmt.Printf("  下载 protoc: %s\n", url)
	}

	resp, err := i.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download protoc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download protoc: HTTP %d from %s", resp.StatusCode, url)
	}

	if err := i.extractProtoc(resp.Body); err != nil {
		return "", fmt.Errorf("extract protoc: %w", err)
	}
	return i.protocBinPath(), nil
}

// extractProtoc extracts the protoc binary from a zip reader.
// extractProtoc 从 zip reader 解压 protoc 二进制。
func (i *protocInstaller) extractProtoc(r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read zip: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	binDir := filepath.Join(i.cacheDir, "bin")
	if err := i.fs.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}

	for _, f := range zr.File {
		name := filepath.ToSlash(f.Name)
		if !strings.HasSuffix(name, "bin/protoc") && !strings.HasSuffix(name, "bin/protoc.exe") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %s in zip: %w", f.Name, err)
		}
		defer rc.Close()

		outPath := filepath.Join(binDir, filepath.Base(name))
		out, err := i.fs.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", outPath, err)
		}
		defer out.Close()

		if _, err := io.Copy(out, rc); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		return nil
	}
	return fmt.Errorf("protoc binary not found in release zip")
}

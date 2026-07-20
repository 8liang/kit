package protobuf

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/spf13/afero"
)

// protoDownloader downloads proto files into the cache.
// protoDownloader 将 proto 文件下载到缓存。
type protoDownloader struct {
	cacheDir string
	fs       afero.Fs
	client   *http.Client
	verbose  bool
}

// newProtoDownloader creates a new protoDownloader.
// newProtoDownloader 创建新的 protoDownloader。
func newProtoDownloader(cacheDir string, fs afero.Fs, client *http.Client, verbose bool) *protoDownloader {
	return &protoDownloader{
		cacheDir: cacheDir,
		fs:       fs,
		client:   client,
		verbose:  verbose,
	}
}

// isCached checks if an import path already exists in the cache.
// isCached 检查 import 路径是否已在缓存中。
func (d *protoDownloader) isCached(importPath string) bool {
	target := filepath.Join(d.cacheDir, "imports", importPath)
	_, err := d.fs.Stat(target)
	return err == nil
}

// download fetches a proto file from url and saves it to the cache.
// download 从 url 下载 proto 文件并保存到缓存。
func (d *protoDownloader) download(importPath, url string) error {
	target := filepath.Join(d.cacheDir, "imports", importPath)

	// 确保父目录存在
	if err := d.fs.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	if d.verbose {
		fmt.Printf("  下载: %s\n     → %s\n", url, target)
	}

	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	f, err := d.fs.Create(target)
	if err != nil {
		return fmt.Errorf("create cache file %s: %w", target, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write cache file %s: %w", target, err)
	}
	return nil
}

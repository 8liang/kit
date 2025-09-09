package loader

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path"
	"path/filepath"
)

type Loader struct {
	normalTpls map[string]any
	singleTpls map[string]any
	assetsDir  string
	assets     fs.ReadDirFS
}

func New() *Loader {
	return &Loader{
		normalTpls: make(map[string]any),
		singleTpls: make(map[string]any),
	}
}

func (m *Loader) RegisterNormal(name string, tpl any) {
	m.normalTpls[name] = tpl
}

func (m *Loader) RegisterSingle(name string, tpl any) {
	m.singleTpls[name] = tpl
}

func (m *Loader) Load(assets fs.ReadDirFS, assetsDir string) (err error) {
	m.assets = assets
	m.assetsDir = assetsDir
	var files []fs.DirEntry
	if files, err = assets.ReadDir(assetsDir); err != nil {
		return fmt.Errorf("loading assets: %w", err)
	}
	for _, file := range files {
		if err = m.tryLoadFile(file); err != nil {
			return
		}
	}
	return
}

func (m *Loader) tryLoadFile(file fs.DirEntry) error {
	name := file.Name()
	if path.Ext(name) != ".json" {
		return nil
	}
	return m.loadFile(filepath.Join(m.assetsDir, name))
}

func (m *Loader) loadFile(p string) (err error) {
	var contents []byte
	if contents, err = m.readContent(p); err != nil {
		return fmt.Errorf("failed to open %q: %w", p, err)
	}
	if len(contents) == 0 {
		return
	}
	fileName := path.Base(p)
	name := fileName[0 : len(fileName)-5]
	if err = m.tryParseNormal(name, contents); err != nil {
		return
	}
	err = m.tryParseSingle(name, contents)
	return
}

func (m *Loader) readContent(p string) (contents []byte, err error) {
	var file fs.File
	if file, err = m.assets.Open(p); err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", p, err)
	}
	defer func(file fs.File) {
		if e := file.Close(); e != nil {
			slog.Error("manager close file error", "err", e)
		}
	}(file)
	if contents, err = io.ReadAll(file); err != nil {
		return nil, fmt.Errorf("cant read file %q: %w", p, err)
	}
	return
}
func (m *Loader) tryParseNormal(name string, contents []byte) error {
	if ins, ok := m.normalTpls[name]; ok {
		return json.Unmarshal(contents, ins)
	}
	return nil
}

func (m *Loader) tryParseSingle(name string, contents []byte) (err error) {
	if ins, ok := m.singleTpls[name]; ok {
		return parseSingle(contents, ins)
	}
	return nil
}

func parseSingle[T any](content []byte, ins T) (err error) {
	s := []T{ins}
	return json.Unmarshal(content, &s)
}

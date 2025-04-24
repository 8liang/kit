package excel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func Export(excelDir, clientDir, serverDir string, opts ...Option) (err error) {
	_Config = &config{
		excelDir:    excelDir,
		clientDir:   clientDir,
		serverDir:   serverDir,
		packageName: "templates",
	}
	for _, opt := range opts {
		opt(_Config)
	}
	return process()
}
func process() (err error) {
	var excelFiles []string
	var sheets []*Sheet
	if excelFiles, err = fetchExcelFiles(_Config.excelDir); err != nil {
		return
	}
	for _, fileName := range excelFiles {
		var _sheets []*Sheet
		if _sheets, err = parse(fileName); err != nil {
			return err
		}
		sheets = append(sheets, _sheets...)
	}
	if err = saveJsonFiles(sheets); err != nil {
		return
	}
	return
}

func fetchExcelFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var excelFiles []string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".xls") || strings.HasSuffix(file.Name(), ".xlsx")) && !strings.HasPrefix(file.Name(), "~") {
			excelFiles = append(excelFiles, filepath.Join(_Config.excelDir, file.Name()))
		}
	}

	if len(excelFiles) == 0 {
		return nil, fmt.Errorf("no excel files found in directory: %s", _Config.excelDir)
	}
	return excelFiles, nil
}

func shouldSheetExport(name string) (should bool, exportingName string, d Direction) {
	for _, direction := range []Direction{DirectionHorizontal, DirectionVertical} {
		prefix := fmt.Sprintf("%s-", direction)
		if should = strings.HasPrefix(name, prefix); should {
			d = direction
			exportingName = strings.TrimLeft(name, prefix)
			return
		}
	}
	return
}

func parse(fileName string) (sheets []*Sheet, err error) {
	var file *excelize.File
	if file, err = excelize.OpenFile(fileName); err != nil {
		return
	}
	var rows [][]string
	var sheet *Sheet
	var direction Direction
	for _, name := range file.GetSheetList() {
		var exportingName string
		var shouldExport bool
		if shouldExport, exportingName, direction = shouldSheetExport(name); !shouldExport {
			continue
		}
		if rows, err = file.GetRows(name); err != nil {
			return
		}
		if sheet, err = parseSheet(exportingName, direction, rows); err != nil {
			return
		}
		sheets = append(sheets, sheet)
	}
	return
}

func saveJsonFiles(sheets []*Sheet) (err error) {
	bf := bytes.Buffer{}
	for _, sheet := range sheets {
		for _, f := range []struct {
			shouldFieldDisplay func(f *Field) bool
			dirPath            string
		}{
			{shouldFieldDisplay: func(f *Field) bool { return strings.Contains(f.Mark, "c")}, dirPath: _Config.clientDir},
			{shouldFieldDisplay: func(f *Field) bool { return strings.Contains(f.Mark, "s") }, dirPath: _Config.serverDir},
		} {
			if err = saveJsonFileUsingSheet(sheet, f.dirPath, f.shouldFieldDisplay); err != nil {
				return
			}
		}
		if err = saveStructFileUsingSheet(sheet, _Config.serverDir); err != nil {
			return
		}
		var _bytes []byte
		if _bytes, err = genInterface(sheet, func(f *Field) bool {
			return strings.Contains(f.Mark, "c")
		}); err != nil {
			return
		}
		if _, err = bf.Write(_bytes); err != nil {
			return
		}
		if _, err = bf.WriteString("\n"); err != nil {
			return
		}
	}
	err = writeFile(filepath.Join(_Config.clientDir, "interfaceTpl.ts"), bf.Bytes())
	return
}

func saveJsonFileUsingSheet(sheet *Sheet, dirPath string, shouldFieldDisplay func(f *Field) bool) (err error) {
	var jsonByte []byte
	var jsonData []map[string]any
	if jsonData, err = sheet.ToJson(shouldFieldDisplay); err != nil {
		return err
	}
	if jsonByte, err = json.Marshal(jsonData); err != nil {
		return
	}
	err = writeFile(filepath.Join(dirPath, sheet.Name+".json"), jsonByte)
	return
}

func saveStructFileUsingSheet(sheet *Sheet, dirPath string) (err error) {
	var structByte []byte
	if structByte, err = genStruct(sheet, func(f *Field) bool {
		return strings.Contains(f.Mark, "s")
	}); err != nil {
		return
	}
	err = writeFile(filepath.Join(dirPath, sheet.Name+".go"), structByte)
	return
}

func writeFile(fileName string, data []byte) (err error) {
	if err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
		return
	}
	var file *os.File
	if file, err = os.Create(fileName); err != nil {
		return
	}
	defer file.Close()
	_, err = file.Write(data)
	return
}

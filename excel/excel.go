package excel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExportToJSON 是一个将数据导出为 JSON 格式的函数。
// ExportToJSON is a function that exports data to JSON format.
func ExportToJSON(excelDir string, opts ...Option) (err error) {
	cfg = &config{
		excelDir: excelDir,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if len(cfg.jsonConfigs) == 0 {
		cfg.jsonConfigs = []*schema{
			{
				outPath:           filepath.Join(excelDir, "json"),
				shouldExportField: ShouldExportAllField,
			},
		}
	}
	return process()
}

// process 是一个处理 Excel 文件并保存为 JSON 的函数。
// process is a function that processes Excel files and saves them as JSON.
func process() (err error) {
	var excelFiles []string
	var sheets []*Sheet
	if excelFiles, err = fetchExcelFiles(cfg.excelDir); err != nil {
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
			excelFiles = append(excelFiles, filepath.Join(cfg.excelDir, file.Name()))
		}
	}

	if len(excelFiles) == 0 {
		return nil, fmt.Errorf("no excel files found in directory: %s", cfg.excelDir)
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
		if sheet, err = parseSheet(fileName, exportingName, direction, rows); err != nil {
			if errors.Is(err, ErrEmptySheet) {
				err = nil
				continue
			} else {
				return
			}
		}
		sheets = append(sheets, sheet)
	}
	return
}

func saveJsonFiles(sheets []*Sheet) (err error) {
	// bf := bytes.Buffer{}
	for _, sheet := range sheets {
		for _, _cfg := range cfg.jsonConfigs {
			if err = exportJson(sheet, _cfg.outPath, _cfg.hashKey, _cfg.shouldExportField); err != nil {
				return
			}
		}

		for _, _cfg := range cfg.schemas {
			if err = exportSchema(sheet, _cfg.outPath, _cfg.schemaType, _cfg.shouldExportField, _cfg.extraArgs); err != nil {
				return
			}
		}
	}
	return
}

func exportJson(sheet *Sheet, dirPath string, asHash string, shouldFieldDisplay func(f *Field) bool) (err error) {
	var jsonByte []byte
	var jsonData []map[string]any
	if jsonData, err = sheet.ToJson(shouldFieldDisplay); err != nil {
		return err
	}
	if len(jsonData) == 0 {
		return
	}
	if asHash != "" {
		hashData := make(map[string]any)
		for _, item := range jsonData {
			id, ok := item[asHash]
			if !ok {
				return fmt.Errorf("json data must have '%s' field for hash export: %s|%s", asHash, sheet.FileName, sheet.Name)
			}
			switch _id := id.(type) {
			case int64:
				hashData[fmt.Sprintf("%d", _id)] = item
			case float64:
				hashData[fmt.Sprintf("%d", int64(_id))] = item
			case string:
				hashData[_id] = item
			default:
				return fmt.Errorf("invalid '%s' type for hash export: %s|%s", asHash, sheet.FileName, sheet.Name)
			}
		}
		jsonByte, err = json.Marshal(hashData)
	} else {
		jsonByte, err = json.Marshal(jsonData)
	}
	if err != nil {
		return
	}
	err = writeFile(filepath.Join(dirPath, sheet.Name+".json"), jsonByte)
	return
}

func exportSchema(sheet *Sheet, outPath string, schemaType SchemaType, shouldExportField ShouldExportField, args []string) (err error) {
	switch schemaType {
	case SchemaTypeGoStruct:
		err = exportGoStruct(sheet, outPath, args[0], shouldExportField)
	case SchemaTypeTsInterface:
		err = exportTsInterface(sheet, outPath, shouldExportField)

	}
	return
}

func exportGoStruct(sheet *Sheet, outPath, packageName string, shouldExportField ShouldExportField) (err error) {
	var structByte []byte
	if structByte, err = genStruct(sheet, packageName, shouldExportField); err != nil {
		return
	}
	err = writeFile(filepath.Join(outPath, sheet.Name+".go"), structByte)
	return
}

func exportTsInterface(sheet *Sheet, outPath string, shouldExportField ShouldExportField) (err error) {
	var _bytes []byte
	if _bytes, err = genInterface(sheet, shouldExportField); err != nil {
		return
	}
	err = writeFile(filepath.Join(outPath, sheet.Name+"Tpl.ts"), _bytes)
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

package excel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func Export(excelDir string, opts ...Option) (err error) {
	_Config = &config{
		excelDir: excelDir,
	}
	for _, opt := range opts {
		opt(_Config)
	}
	if len(_Config.jsonConfigs) == 0 {
		_Config.jsonConfigs = []*schema{
			{
				outPath:           filepath.Join(excelDir, "json"),
				shouldExportField: shouldExportAllField,
			},
		}
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
	// bf := bytes.Buffer{}
	for _, sheet := range sheets {
		for _, cfg := range _Config.jsonConfigs {
			if err = exportJson(sheet, cfg.outPath, cfg.shouldExportField); err != nil {
				return
			}
		}

		for _, cfg := range _Config.schemas {
			if err = exportSchema(sheet, cfg.outPath, cfg.schemaType, cfg.shouldExportField, cfg.extraArgs); err != nil {
				return
			}
		}

		// var _bytes []byte
		// if _bytes, err = genInterface(sheet, func(f *Field) bool {
		// 	return strings.Contains(f.Mark, "c")
		// }); err != nil {
		// 	return
		// }
		// if _, err = bf.Write(_bytes); err != nil {
		// 	return
		// }
		// if _, err = bf.WriteString("\n"); err != nil {
		// 	return
		// }
	}
	// err = writeFile(filepath.Join(_Config.clientDir, "interfaceTpl.ts"), bf.Bytes())
	return
}

func exportJson(sheet *Sheet, dirPath string, shouldFieldDisplay func(f *Field) bool) (err error) {
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

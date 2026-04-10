package excel

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/8liang/kit"
	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

const FieldNameLine = 3

const FieldTypeLine = 2

const FieldVisibleLine = 1

const FieldComment = 0

type Direction string

const (
	DirectionHorizontal Direction = "h"
	DirectionVertical             = "v"
)

type Sheet struct {
	OriginalName string
	FileName     string
	Name         string
	Fields       []*Field
	Rows         [][]string
	Direction    Direction
	File         *excelize.File
}

func (s *Sheet) ToJson(fieldsFilter func(f *Field) bool) (data []map[string]any, err error) {
	data = make([]map[string]any, 0, len(s.Rows)-FieldNameLine-1)
	for i := FieldNameLine + 1; i < len(s.Rows); i++ {
		var item map[string]any
		if item, err = s.ReadRowData(i, fieldsFilter); err != nil {
			return
		}
		if len(item) > 0 {
			data = append(data, item)
		}
	}
	return
}

func (s *Sheet) ReadRowData(i int, shouldFieldDisplay func(f *Field) bool) (item map[string]any, err error) {
	item = make(map[string]any)
	for j := 0; j < len(s.Fields); j++ {
		field := s.Fields[j]
		if !shouldFieldDisplay(field) {
			continue
		}
		if item[field.Name], err = field.ReadValue(s.Rows[i], field.Index, shouldFieldDisplay); err != nil {
			return
		}
	}
	return
}

func parseSheet(fileName, name string, file *excelize.File) (s *Sheet, err error) {
	shouldExport, exportingName, direction := shouldSheetExport(name)
	if !shouldExport {
		return
	}

	var rows [][]string
	if rows, err = file.GetRows(name); err != nil {
		return
	}
	s = &Sheet{
		Name:         exportingName,
		OriginalName: name,
		Rows:         rows,
		Direction:    direction,
		FileName:     filepath.Base(fileName),
		File:         file,
	}
	if s.Direction == DirectionVertical {
		s.Rows = s.transposeMatrix(rows)
	}
	err = s.parse()
	return
}

func (s *Sheet) parse() (err error) {
	if err = s.resolveFieldName(); err != nil {
		return err
	}
	if err = s.resolveFieldType(); err != nil {
		return err
	}
	s.resolveFieldVisible()
	err = s.parseTime()
	return
}

func (s *Sheet) parseTime() (err error) {
	for _, field := range s.Fields {
		if field.Type != FieldTypeTime {
			continue
		}
		for j, row := range s.Rows {
			if j <= FieldNameLine {
				continue
			}
			if row[field.Index], err = s.getCellTimeValue(j, field.Index); err != nil {
				return
			}
		}
	}
	return
}
func (s *Sheet) getCellTimeValue(col, row int) (value string, err error) {
	if s.Direction == DirectionHorizontal {
		col, row = row, col
	}
	var cell string
	var strValue string
	var t time.Time
	if cell, err = excelize.CoordinatesToCellName(col+1, row+1); err != nil {
		return
	}
	if strValue, err = s.File.GetCellValue(s.OriginalName, cell, excelize.Options{
		RawCellValue: true,
	}); err != nil {
		return
	}
	num, _ := strconv.ParseFloat(strValue, 64)
	if t, err = excelize.ExcelDateToTime(num, false); err != nil {
		return
	}
	return t.Format(time.DateTime), nil
}

func (s *Sheet) resolveFieldVisible() {
	for _, f := range s.Fields {
		f.resolveMark(s.Rows[FieldVisibleLine])
	}
}

func Read(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}
func (s *Sheet) resolveFieldName() (err error) {
	data := s.Rows[FieldNameLine]
	commands := s.Rows[FieldComment]
	s.Fields = []*Field{}
	for i := 0; i < len(data); i++ {
		field := ParseField(i, data[i], Read(commands, i))
		switch field.Type {
		case FieldTypeArray:
			i = field.tryCollectArray(data)
		case FieldTypeObjectArray:
			i = field.tryCollectObjectArray(data)
		}
		s.Fields = append(s.Fields, field)
	}
	_, err = lo.MapValuesErr(lo.CountValuesBy(s.Fields, func(f *Field) string {
		return f.Name
	}), func(value int, key string) (int, error) {
		if value > 1 {
			return value, fmt.Errorf("%w, field: %s", kit.ErrDuplicateExcelFieldName, key)
		}
		return value, nil
	})
	return
}

func (s *Sheet) resolveFieldType() (err error) {
	verticalData := s.transposeMatrix(s.Rows[FieldNameLine+1 : len(s.Rows)])
	if len(verticalData) == 0 {
		return ErrEmptySheet
	}
	for i := 0; i < len(s.Fields); i++ {
		field := s.Fields[i]

		if err = field.resolveType(s.Rows[FieldTypeLine], verticalData); err != nil {
			return
		}
	}
	return
}

func (s *Sheet) transposeMatrix(data [][]string) (verticalData [][]string) {
	if len(data) == 0 {
		return
	}

	// Initialize verticalData with the appropriate number of columns
	maxColumns := 0
	for _, row := range data {
		if len(row) > maxColumns {
			maxColumns = len(row)
		}
	}

	verticalData = make([][]string, maxColumns)
	for i := 0; i < len(data); i++ {
		row := data[i]
		for j := 0; j < len(row); j++ {
			if verticalData[j] == nil {
				verticalData[j] = []string{}
			}
			verticalData[j] = append(verticalData[j], row[j])
		}
	}
	return
}

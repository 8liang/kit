package excel

import "path/filepath"

const FieldNameLine = 3

const FieldTypeLine = 2

const FieldVisibleLine = 1

type Direction string

const (
	DirectionHorizontal Direction = "h"
	DirectionVertical             = "v"
)

type Sheet struct {
	FileName  string
	Name      string
	Fields    []*Field
	Rows      [][]string
	Direction Direction
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

func parseSheet(fileName, name string, direction Direction, rows [][]string) (s *Sheet, err error) {
	s = &Sheet{
		Name:      name,
		Rows:      rows,
		Direction: direction,
		FileName:  filepath.Base(fileName),
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
	return
}

func (s *Sheet) resolveFieldVisible() {
	for _, f := range s.Fields {
		f.resolveMark(s.Rows[FieldVisibleLine])
	}
}

func (s *Sheet) resolveFieldName() error {
	data := s.Rows[FieldNameLine]
	s.Fields = []*Field{}
	for i := 0; i < len(data); i++ {
		field := ParseField(i, data[i])
		switch field.Type {
		case FieldTypeArray:
			i = field.tryCollectArray(data, i)
		case FieldTypeObjectArray:
			i = field.tryCollectObjectArray(data, i)
		}
		s.Fields = append(s.Fields, field)
	}
	return nil
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

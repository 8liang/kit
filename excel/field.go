package excel

import (
	"strconv"
	"strings"
)

type FieldType string

const (
	FieldTypeString      FieldType = "string"
	FieldTypeInt                   = "int64"
	FieldTypeFloat                 = "float64"
	FieldTypeArray                 = "array"
	FieldTypeObjectArray           = "objectArray"
)

type Field struct {
	OrdinalName string
	Name        string
	Type        FieldType
	Index       int
	Rang        [2]int
	SubFields   []*Field
	Mark        string
}

func (f *Field) parseInt(row []string, index int) (value int64, err error) {
	if len(row) <= index || row[index] == "" {
		return 0, nil
	}
	value, err = strconv.ParseInt(row[index], 10, 64)
	return
}

func (f *Field) parseIFloat(row []string, index int) (value float64, err error) {
	if len(row) <= index || row[index] == "" {
		return 0, nil
	}
	value, err = strconv.ParseFloat(row[index], 64)
	return
}

func (f *Field) ReadValue(row []string, index int, shouldFieldDisplay func(f *Field) bool) (value any, err error) {
	switch f.Type {
	case FieldTypeInt:
		value, err = f.parseInt(row, index)
		return
	case FieldTypeFloat:
		value, err = f.parseIFloat(row, index)
		return
	case FieldTypeString:
		return row[index], nil
	case FieldTypeArray:
		return f.readArrayValue(row, index, shouldFieldDisplay)
	case FieldTypeObjectArray:
		return f.readObjectArrayValue(row, index, shouldFieldDisplay)
	default:
		return nil, nil
	}
}

func (f *Field) readObjectArrayValue(row []string, index int, shouldFieldDisplay func(f *Field) bool) (value []any, err error) {
	step := len(f.SubFields)
	for i := f.Rang[0]; i <= f.Rang[1]; i += step {
		_item := make(map[string]any)
		for j, sf := range f.SubFields {
			if !shouldFieldDisplay(sf) {
				continue
			}
			var v any
			if v, err = sf.ReadValue(row, i+j, shouldFieldDisplay); err != nil {
				return
			}
			_item[sf.Name] = v
		}
		value = append(value, _item)
	}
	return
}

func (f *Field) readArrayValue(row []string, index int, shouldFieldDisplay func(*Field) bool) (value []any, err error) {
	for i := f.Rang[0]; i <= f.Rang[1]; i++ {
		var v any
		if v, err = f.SubFields[0].ReadValue(row, i, shouldFieldDisplay); err != nil {
			return
		}
		value = append(value, v)
	}
	return
}

func (f *Field) resolveMark(row []string) {
	if len(row) <= f.Index {
		return
	}
	f.Mark = row[f.Index]
	if f.Type == FieldTypeArray {
		for _, sf := range f.SubFields {
			sf.Mark = f.Mark
		}
		return
	}
	if f.Type == FieldTypeObjectArray {
		seen := make(map[rune]bool)
		result := make([]rune, 0)
		for _, sf := range f.SubFields {
			sf.resolveMark(row)
			for _, ch := range sf.Mark {
				if !seen[ch] {
					seen[ch] = true
					result = append(result, ch)
				}
			}
		}
		f.Mark = string(result)
	}
}

func ParseField(index int, name string) *Field {
	f := &Field{
		Index:       index,
		Name:        name,
		OrdinalName: name,
		Rang:        [2]int{0, 0},
	}
	f.parse()
	return f
}

func (f *Field) resolveType(markRow []string, verticalData [][]string) {
	columnData := verticalData[f.Index]
	var typeMark string
	if len(markRow) > f.Index {
		typeMark = markRow[f.Index]
	}
	if f.Type == FieldTypeArray || f.Type == FieldTypeObjectArray {
		for _, sf := range f.SubFields {
			sf.resolveType(markRow, verticalData)
		}
		return
	}
	switch typeMark {
	case "int":
		f.Type = FieldTypeInt
		break
	case "float":
		f.Type = FieldTypeFloat
		break
	case "string":
		f.Type = FieldTypeString
		break
	default:
		f.detectTypeUsingData(columnData)
		break
	}
}

func isFloat(s string) bool {
	if !strings.Contains(s, ".") {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
func isInt(s string) bool {
	if s == "" {
		return true
	}
	if strings.Contains(s, ".") {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func (f *Field) detectTypeUsingData(data []string) {
	if len(data) == 0 {
		f.Type = FieldTypeString
		return
	}
	for _, d := range data {
		if isFloat(d) {
			f.Type = FieldTypeFloat
			continue
		}
		if f.Type != FieldTypeFloat {
			if isInt(d) {
				f.Type = FieldTypeInt
				continue
			}
			f.Type = FieldTypeString
			return
		}
	}
}

func (f *Field) tryCollectObjectArray(data []string, i int) int {
	for i := f.Index + 1; i < len(data); i++ {
		if !strings.Contains(data[i], ".") {
			return i - 1
		}
		tmp := strings.SplitN(data[i], ".", 2)
		if f.Name != tmp[0] {
			return i - 1
		}
		found := false
		for _, subField := range f.SubFields {
			if subField.Name == tmp[1] {
				found = true
				break
			}
		}
		if !found {
			f.SubFields = append(f.SubFields, &Field{
				Index:       i,
				Name:        tmp[1],
				OrdinalName: tmp[1],
			})
		}
		f.Rang[1] = i
	}
	return len(data) - 1
}

func (f *Field) tryCollectArray(data []string, i int) int {
	for i := f.Index + 1; i < len(data); i++ {
		if f.OrdinalName == data[i] {
			f.Rang[1] = i
		} else {
			return i - 1
		}
	}
	return len(data) - 1
}

func (f *Field) parse() {
	if idx := strings.Index(f.Name, "[]"); idx > -1 {
		f.Type = FieldTypeArray
		f.Name = strings.Replace(f.Name, "[]", "", 1)
		f.Rang[0] = f.Index
		f.SubFields = append(f.SubFields, &Field{
			Index:       f.Index,
			Name:        f.Name,
			OrdinalName: f.Name,
		})
		return
	}

	if idx := strings.Index(f.Name, "."); idx > -1 {
		f.Type = FieldTypeObjectArray
		f.Rang[0] = f.Index
		tmp := strings.SplitN(f.Name, ".", 2)
		f.Name = tmp[0]
		f.SubFields = append(f.SubFields, &Field{
			Index:       f.Index,
			Name:        tmp[1],
			OrdinalName: tmp[1],
		})
		return
	}
}

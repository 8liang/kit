package excel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenStruct(t *testing.T) {
	s := &Sheet{
		Name:        "Test",
		Fields: []*Field{
			{
				Name: "string1",
				Type: FieldTypeString,
			},
			{
				Name: "int2",
				Type: FieldTypeInt,
			},
			{
				Name: "float3",
				Type: FieldTypeFloat,
			},
			{
				Name: "array4",
				Type: FieldTypeArray,
				SubFields: []*Field{
					{
						Name: "string1",
						Type: FieldTypeString,
					},
				},
			},
			{
				Name: "objectArray5",
				Type: FieldTypeObjectArray,
				SubFields: []*Field{
					{
						Name: "string1",
						Type: FieldTypeString,
					},
					{
						Name: "int2",
						Type: FieldTypeInt,
					},
				},
			},
		},
	}

	result, err := genStruct(s,"tpl",  func(f *Field) bool {
		return true
	})
	assert.NoError(t, err)
	t.Log(string(result))

}

func TestGenInterface(t *testing.T) {
	s := &Sheet{
		Name:        "Test",
		Fields: []*Field{
			{
				Name: "string1",
				Type: FieldTypeString,
			},
			{
				Name: "int2",
				Type: FieldTypeInt,
			},
			{
				Name: "float3",
				Type: FieldTypeFloat,
			},
			{
				Name: "array4",
				Type: FieldTypeArray,
				SubFields: []*Field{
					{
						Name: "string1",
						Type: FieldTypeString,
					},
				},
			},
			{
				Name: "objectArray5",
				Type: FieldTypeObjectArray,
				SubFields: []*Field{
					{
						Name: "string1",
						Type: FieldTypeString,
					},
					{
						Name: "int2",
						Type: FieldTypeInt,
					},
				},
			},
		},
	}

	result, err := genInterface(s, func(f *Field) bool {
		return true
	})
	assert.NoError(t, err)
	t.Log(string(result))
}

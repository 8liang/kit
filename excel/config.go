package excel

var _Config *config

type SchemaType string

const (
	SchemaTypeTsInterface = "tsInterface"
	SchemaTypeGoStruct    = "goStruct"
)

type config struct {
	excelDir    string
	schemas     []*schema
	jsonConfigs []*schema
}

type Option func(*config)

type schema struct {
	outPath           string
	shouldExportField ShouldExportField
	schemaType        SchemaType
	extraArgs         []string
}

type ShouldExportField func(*Field) bool

var shouldExportAllField = func(_ *Field) bool { return true }

func WithJson(outPath string, shouldExportField ...ShouldExportField) Option {
	return func(c *config) {
		c.jsonConfigs = append(c.jsonConfigs, &schema{
			outPath:           outPath,
			shouldExportField: processShouldExportField(shouldExportField),
		})
	}
}

func WithSchemaExport(outPath string, schemaType SchemaType, shouldExportField ShouldExportField, args ...string) Option {
	return func(c *config) {
		cfg := &schema{
			outPath:           outPath,
			shouldExportField: shouldExportField,
			schemaType:        schemaType,
			extraArgs:         args,
		}
		c.schemas = append(c.schemas, cfg)
	}
}

func WithTsInterfaceExport(outPath string, shouldExportField ...ShouldExportField) Option {
	return WithSchemaExport(outPath, SchemaTypeTsInterface, processShouldExportField(shouldExportField))
}

func WithGoStructExport(outPath string, packageName string, shouldExportField ...ShouldExportField) Option {
	return WithSchemaExport(outPath, SchemaTypeGoStruct, processShouldExportField(shouldExportField), packageName)
}

func processShouldExportField(shouldExportField []ShouldExportField) ShouldExportField {
	if len(shouldExportField) == 0 {
		return shouldExportAllField
	}
	return shouldExportField[0]
}

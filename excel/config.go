package excel

var cfg *config

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

var ShouldExportAllField = func(_ *Field) bool { return true }

// WithJson is a function that configures JSON export settings.
// WithJson 是一个配置 JSON 导出设置的函数。
func WithJson(outPath string, shouldExportField ...ShouldExportField) Option {
	return func(c *config) {
		c.jsonConfigs = append(c.jsonConfigs, &schema{
			outPath:           outPath,
			shouldExportField: processShouldExportField(shouldExportField),
		})
	}
}

// WithSchemaExport is a function that configures schema export settings.
// WithSchemaExport 是一个配置模式导出设置的函数。
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

// WithTsInterfaceExport is a function that configures TypeScript interface export settings.
// WithTsInterfaceExport 是一个配置 TypeScript 接口导出设置的函数。
func WithTsInterfaceExport(outPath string, shouldExportField ...ShouldExportField) Option {
	return WithSchemaExport(outPath, SchemaTypeTsInterface, processShouldExportField(shouldExportField))
}

// WithGoStructExport is a function that configures Go struct export settings.
// WithGoStructExport 是一个配置 Go 结构体导出设置的函数。
func WithGoStructExport(outPath string, packageName string, shouldExportField ...ShouldExportField) Option {
	return WithSchemaExport(outPath, SchemaTypeGoStruct, processShouldExportField(shouldExportField), packageName)
}

func processShouldExportField(shouldExportField []ShouldExportField) ShouldExportField {
	if len(shouldExportField) == 0 {
		return ShouldExportAllField
	}
	return shouldExportField[0]
}

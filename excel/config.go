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
	hashKey           string
	shouldExportField ShouldExportField
	schemaType        SchemaType
	extraArgs         []string
	goIntType         string
}

type ShouldExportField func(*Field) bool

var ShouldExportAllField = func(_ *Field) bool { return true }

// WithJson is a function that configures JSON export settings.
// WithJson 是一个配置 JSON 导出设置的函数。
// outPath: 输出路径
// hashKey: 	是否以 Hash 导出
func WithJson(outPath string, opts ...SchemaOption) Option {
	return func(c *config) {
		s := &schema{
			outPath:           outPath,
			shouldExportField: ShouldExportAllField,
		}
		for _, opt := range opts {
			opt(s)
		}
		c.jsonConfigs = append(c.jsonConfigs, s)
	}
}

// WithSchemaExport is a function that configures schema export settings.
// WithSchemaExport 是一个配置模式导出设置的函数。
func WithSchemaExport(outPath string, schemaType SchemaType, opts ...SchemaOption) Option {
	return func(c *config) {
		s := &schema{
			outPath:           outPath,
			shouldExportField: ShouldExportAllField,
			schemaType:        schemaType,
			goIntType:         "int64", // default int type for Go struct
		}
		for _, opt := range opts {
			opt(s)
		}
		c.schemas = append(c.schemas, s)
	}
}

// WithTsInterfaceExport is a function that configures TypeScript interface export settings.
// WithTsInterfaceExport 是一个配置 TypeScript 接口导出设置的函数。
func WithTsInterfaceExport(outPath string, opts ...SchemaOption) Option {
	return WithSchemaExport(outPath, SchemaTypeTsInterface, opts...)
}

// WithGoStructExport is a function that configures Go struct export settings.
// WithGoStructExport 是一个配置 Go 结构体导出设置的函数。
func WithGoStructExport(outPath string, packageName string, opts ...SchemaOption) Option {
	opts = append(opts, WithExtraArgs([]string{packageName}))
	return WithSchemaExport(outPath, SchemaTypeGoStruct, opts...)
}

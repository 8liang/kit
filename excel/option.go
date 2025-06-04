package excel

type SchemaOption func(*schema)

func WithHashKey(hashKey string) SchemaOption {
	return func(s *schema) {
		s.hashKey = hashKey
	}
}

func WithShouldExportField(shouldExportField ShouldExportField) SchemaOption {
	return func(s *schema) {
		s.shouldExportField = shouldExportField
	}
}

func WithExtraArgs(args []string) SchemaOption {
	return func(s *schema) {
		s.extraArgs = args
	}
}

func WithGoIntType(goIntType string) SchemaOption {
	return func(s *schema) {
		s.goIntType = goIntType
	}
}

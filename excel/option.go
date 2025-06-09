package excel

type SchemaOption func(*schema)

func WithHashKey(hashKey string, hashTolerance ...bool) SchemaOption {
	return func(s *schema) {
		s.hashKey = hashKey
		if len(hashTolerance) > 0 {
			s.tolerantHashKeyError = hashTolerance[0]
		}
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

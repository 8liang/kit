package kit

import "errors"

var (
	ErrParameterInvalid = errors.New("parameter invalid")
	ErrPageOutOfRange   = errors.New("page out of range")
)

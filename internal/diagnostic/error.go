package diagnostic

import (
	"orische/internal/ast"
)

type Error struct {
	Message string
	Range   ast.Range
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

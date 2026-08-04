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

func (e *Error) UnWrap() error {
	return e.Err
}

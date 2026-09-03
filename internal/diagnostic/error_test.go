package diagnostic

import (
	"errors"
	"testing"
)

func TestErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("cause")
	err := &Error{Message: "diagnostic", Err: cause}

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is did not find the diagnostic cause")
	}
}

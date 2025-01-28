package socket

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrToListEmpty        = errors.New("to list is empty")
	ErrToExceptDuplicates = errors.New("the to list duplicates the value of the except list")
	ErrBinData            = errors.New("obj must be a non-nil pointer")
)

type EmitError struct {
	Member *Member
	Err    error
}

type EmitErrors []EmitError

func (e EmitErrors) Error() string {
	msg := make([]string, 0, len(e))

	for _, emitError := range e {
		m := fmt.Sprintf("member: %s, error: %s", emitError.Member.id.String(), emitError.Err.Error())
		msg = append(msg, m)
	}

	return strings.Join(msg, "\n")
}

func addEmitErr(src error, dst EmitError) EmitErrors {
	var err EmitErrors
	if errors.As(src, &err) {
		err = append(err, dst)
		return err
	} else {
		return EmitErrors{dst}
	}
}

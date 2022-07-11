package errors

import (
	stderrors "errors"
	"fmt"
)

var (
	// New is the same as errors.New.
	New = stderrors.New

	// As is the same as errors.As.
	As = stderrors.As
)

// Wrappable is a wrappable error.
type Wrappable interface {
	Error() string
	setWrapped(error)
}

// WrappableError is a wrappable struct that can be easily embedded in error
// types.
type WrappableError struct {
	err error
}

func (e *WrappableError) Error() string {
	return e.err.Error()
}

func (e *WrappableError) Unwrap() error {
	return e.err
}

func (e *WrappableError) setWrapped(err error) {
	e.err = err
}

// Errorf returns a wrappable typed error that can be type checked in tests.
func Errorf(err Wrappable, format string, a ...any) error {
	err.setWrapped(fmt.Errorf(format, a...))
	return err
}

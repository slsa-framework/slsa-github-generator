// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

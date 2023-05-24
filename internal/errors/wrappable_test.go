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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"testing"
)

func TestWrappable(t *testing.T) {
	t.Run("Is EOF", func(t *testing.T) {
		type errFoo struct {
			WrappableError
		}

		err := &errFoo{}
		err.setWrapped(io.EOF)

		if want, got := true, errors.Is(err, io.EOF); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		if want, got := false, errors.Is(err, io.ErrClosedPipe); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}
	})

	t.Run("As DNSError", func(t *testing.T) {
		type errFoo struct {
			WrappableError
		}

		err := &errFoo{}
		err.setWrapped(&net.DNSError{})

		var as *net.DNSError
		if want, got := true, errors.As(err, &as); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		var notAs *fs.PathError
		if want, got := false, errors.As(err, &notAs); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}
	})
}

func TestErrorf(t *testing.T) {
	t.Run("Is EOF", func(t *testing.T) {
		type errFoo struct {
			WrappableError
		}

		err := Errorf(&errFoo{}, "custom: %w", io.EOF)

		if want, got := true, errors.Is(err, io.EOF); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		if want, got := fmt.Sprintf("custom: %s", io.EOF), err.Error(); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		if want, got := false, errors.Is(err, io.ErrClosedPipe); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}
	})

	t.Run("As DNSError", func(t *testing.T) {
		type errFoo struct {
			WrappableError
		}

		dnsErr := &net.DNSError{
			Err: "foo",
		}
		err := Errorf(&errFoo{}, "custom: %w", dnsErr)

		var as *net.DNSError
		if want, got := true, errors.As(err, &as); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		if want, got := fmt.Sprintf("custom: %v", dnsErr), err.Error(); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}

		var notAs *fs.PathError
		if want, got := false, errors.As(err, &notAs); want != got {
			t.Errorf("unexpected result, want: %v, got: %v", want, got)
		}
	})
}

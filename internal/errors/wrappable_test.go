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

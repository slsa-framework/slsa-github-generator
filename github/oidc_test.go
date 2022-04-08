package github

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRequestOIDCToken(t *testing.T) {
	testCases := []struct {
		name     string
		audience string
		expected *OIDCToken
		raw      string
		err      error
	}{
		{
			name:     "basic token",
			audience: "hoge",
			expected: &OIDCToken{JobWorkflowRef: "hoge"},
		},
		{
			name:     "invalid response",
			audience: "hoge",
			raw:      `not json`,
			err:      errResponseJSON,
		},
		{
			name:     "invalid parts",
			audience: "hoge",
			raw:      `{"value": "part1"}`,
			err:      errInvalidToken,
		},
		{
			name:     "invalid base64",
			audience: "hoge",
			raw:      `{"value": "part1.part2.part3"}`,
			err:      errInvalidTokenB64,
		},
		{
			name:     "invalid json",
			audience: "hoge",
			raw:      fmt.Sprintf(`{"value": "part1.%s.part3"}`, base64.RawURLEncoding.EncodeToString([]byte("not json"))),
			err:      errInvalidTokenJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected != nil {
				_, stop := NewTestOIDCServer(tc.expected)
				defer stop()
			} else {
				_, stop := newRawTestOIDCServer(tc.raw)
				defer stop()
			}

			token, err := RequestOIDCToken(tc.audience)
			if !errors.Is(err, tc.err) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tc.err, cmpopts.EquateErrors()))
			}
			if want, got := tc.expected, token; !cmp.Equal(want, got) {
				t.Errorf("unexpected workflow ref\nwant: %#v\ngot:  %#v\ndiff: %#v", want, got, cmp.Diff(want, got))
			}
		})
	}
}

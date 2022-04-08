package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestRequestOIDCToken(t *testing.T) {
	testCases := []struct {
		name     string
		audience string
		expected *OIDCToken
		status   int
		raw      string
		err      error
	}{
		{
			name:     "basic token",
			audience: "hoge",
			expected: &OIDCToken{
				Audience:       "hoge",
				Expiry:         jsonTime(time.Now().Add(24 * time.Hour)),
				JobWorkflowRef: "hoge",
			},
		},
		{
			name:     "invalid response",
			audience: "hoge",
			raw:      `not json`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid parts",
			audience: "hoge",
			raw:      `{"value": "part1"}`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid base64",
			audience: "hoge",
			raw:      `{"value": "part1.part2.part3"}`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid json",
			audience: "hoge",
			raw:      fmt.Sprintf(`{"value": "part1.%s.part3"}`, base64.RawURLEncoding.EncodeToString([]byte("not json"))),
			status:   http.StatusOK,
			err:      &errToken{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected != nil {
				_, stop := NewTestOIDCServer(tc.expected)
				defer stop()
			} else {
				_, stop := newRawTestOIDCServer(tc.status, tc.raw)
				defer stop()
			}

			token, err := RequestOIDCToken(context.Background(), tc.audience)
			if err != nil {
				if tc.err != nil {
					if !errors.As(err, &tc.err) {
						t.Fatalf("unexpected error: %v", cmp.Diff(err, tc.err, cmpopts.EquateErrors()))
					}
				} else {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tc.err, cmpopts.EquateErrors()))
				}
			} else {
				if tc.err != nil {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tc.err, cmpopts.EquateErrors()))
				}
			}
			if want, got := tc.expected, token; !cmp.Equal(want, got) {
				t.Errorf("unexpected workflow ref\nwant: %#v\ngot:  %#v\ndiff:\n%v", want, got, cmp.Diff(want, got))
			}
		})
	}
}

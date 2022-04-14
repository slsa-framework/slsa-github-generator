package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// tokenEqual returns whether the tokens are functionally equal for the purposes of the test.
func tokenEqual(issuer string, wantToken, gotToken *OIDCToken) bool {
	if wantToken == nil && gotToken == nil {
		return true
	}

	if gotToken == nil || wantToken == nil {
		return false
	}

	// NOTE: don't check the wantToken issuer because it's not known until the
	// server is created and we can't use a dummy value because verification checks
	// it.
	if want, got := issuer, gotToken.Issuer; want != got {
		return false
	}

	if want, got := wantToken.Audience, gotToken.Audience; !compareStringSlice(want, got) {
		return false
	}

	if want, got := wantToken.Expiry, gotToken.Expiry; !want.Equal(got) {
		return false
	}

	if want, got := wantToken.JobWorkflowRef, gotToken.JobWorkflowRef; want != got {
		return false
	}

	return true
}

func TestRequestOIDCToken(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		audience []string
		expected *OIDCToken
		status   int
		raw      string
		err      error
	}{
		{
			name:     "basic token",
			audience: []string{"hoge"},
			expected: &OIDCToken{
				Audience:       []string{"hoge"},
				Expiry:         now.Add(1 * time.Hour),
				JobWorkflowRef: "hoge",
			},
		},
		{
			name:     "invalid response",
			audience: []string{"hoge"},
			raw:      `not json`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid parts",
			audience: []string{"hoge"},
			raw:      `{"value": "part1"}`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid base64",
			audience: []string{"hoge"},
			raw:      `{"value": "part1.part2.part3"}`,
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "invalid json",
			audience: []string{"hoge"},
			raw:      fmt.Sprintf(`{"value": "part1.%s.part3"}`, base64.RawURLEncoding.EncodeToString([]byte("not json"))),
			status:   http.StatusOK,
			err:      &errToken{},
		},
		{
			name:     "error response",
			audience: []string{"hoge"},
			raw:      "",
			status:   http.StatusServiceUnavailable,
			err:      &errRequestError{},
		},
		{
			name:     "redirect response",
			audience: []string{"hoge"},
			raw:      "",
			status:   http.StatusFound,
			err:      &errRequestError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s *httptest.Server
			var c *OIDCClient
			if tc.expected != nil {
				s, c = NewTestOIDCServer(t, now, tc.expected)
			} else {
				s, c = newRawTestOIDCServer(t, now, tc.status, tc.raw)
			}
			defer s.Close()

			token, err := c.Token(context.Background(), tc.audience)
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
			if want, got := tc.expected, token; !tokenEqual(s.URL, want, got) {
				t.Errorf("unexpected workflow ref\nwant: %#v\ngot:  %#v\ndiff:\n%v", want, got, cmp.Diff(want, got))
			}
		})
	}
}

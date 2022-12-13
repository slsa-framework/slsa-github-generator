package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"

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

func TestNewOIDCClient(t *testing.T) {
	// Tests that NewOIDCClient returns an error when the
	// ACTIONS_ID_TOKEN_REQUEST_URL env var is empty.
	t.Run("empty url", func(t *testing.T) {
		if os.Getenv(requestURLEnvKey) != "" {
			panic(fmt.Sprintf("expected %v to be empty", requestURLEnvKey))
		}

		_, err := NewOIDCClient()
		if err == nil {
			t.Fatalf("expected error")
		}
		if want, got := (&errURLError{}), err; !errors.As(got, &want) {
			t.Fatalf("unexpected error, want: %#v, got: %#v", want, got)
		}
	})
}

func TestToken(t *testing.T) {
	now := time.Date(2022, 4, 14, 12, 24, 0, 0, time.UTC)

	testCases := []struct {
		name     string     // 16
		raw      string     // 16
		err      error      // 16
		token    *OIDCToken // 8
		audience []string   // 24
		status   int        // 8
	}{
		{
			name:     "basic token",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"hoge"},
				Expiry:            now.Add(1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
		},
		{
			name:     "no repository id claim",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"hoge"},
				Expiry:            now.Add(1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
			err: &errClaims{},
		},
		{
			name:     "no workflow ref claim",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"hoge"},
				Expiry:            now.Add(1 * time.Hour),
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
			err: &errClaims{},
		},
		{
			name:     "no owner id claim",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:       []string{"hoge"},
				Expiry:         now.Add(1 * time.Hour),
				JobWorkflowRef: "pico",
				RepositoryID:   "1234",
				ActorID:        "4567",
			},
			err: &errClaims{},
		},
		{
			name:     "no actor id claim",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"hoge"},
				Expiry:            now.Add(1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
			},
			err: &errClaims{},
		},
		{
			name:     "expired token",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"hoge"},
				Expiry:            now.Add(-1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
			err: &errVerify{},
		},
		{
			name:     "bad audience",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Audience:          []string{"fuga"},
				Expiry:            now.Add(1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
			err: &errVerify{},
		},
		{
			name:     "bad issuer",
			audience: []string{"hoge"},
			token: &OIDCToken{
				Issuer:            "https://www.google.com/",
				Audience:          []string{"hoge"},
				Expiry:            now.Add(1 * time.Hour),
				JobWorkflowRef:    "pico",
				RepositoryID:      "1234",
				RepositoryOwnerID: "4321",
				ActorID:           "4567",
			},
			err: &errVerify{},
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
			if tc.token != nil {
				s, c = NewTestOIDCServer(t, now, tc.token)
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
				} else {
					// Successful response, as expected. Check token.
					if want, got := tc.token, token; !tokenEqual(s.URL, want, got) {
						t.Errorf("unexpected workflow ref\nwant: %#v\ngot:  %#v\ndiff:\n%v", want, got, cmp.Diff(want, got))
					}
				}
			}
		})
	}
}

func Test_compareStringSlice(t *testing.T) {
	testCases := []struct {
		name     string
		left     []string
		right    []string
		expected bool
	}{
		{
			name:     "empty",
			left:     []string{},
			right:    []string{},
			expected: true,
		},
		{
			name:     "nil",
			left:     nil,
			right:    nil,
			expected: true,
		},
		{
			name:     "left nil, right empty",
			left:     nil,
			right:    []string{},
			expected: true,
		},
		{
			name:     "left empty, right nil",
			left:     []string{},
			right:    nil,
			expected: true,
		},
		{
			name:     "equal",
			left:     []string{"hoge", "fuga"},
			right:    []string{"hoge", "fuga"},
			expected: true,
		},
		{
			name:     "unsorted",
			left:     []string{"hoge", "fuga"},
			right:    []string{"fuga", "hoge"},
			expected: true,
		},
		{
			name:     "left bigger",
			left:     []string{"hoge", "fuga", "pico"},
			right:    []string{"fuga", "hoge"},
			expected: false,
		},
		{
			name:     "right bigger",
			left:     []string{"hoge", "fuga"},
			right:    []string{"fuga", "hoge", "pico"},
			expected: false,
		},
		{
			name:     "diff value",
			left:     []string{"hoge", "fuga"},
			right:    []string{"fuga", "pico"},
			expected: false,
		},
		{
			name:     "left nil",
			left:     nil,
			right:    []string{"hoge", "fuga"},
			expected: false,
		},
		{
			name:     "right nil",
			left:     []string{"hoge", "fuga"},
			right:    nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if want, got := tc.expected, compareStringSlice(tc.left, tc.right); want != got {
				t.Errorf("unexpected result, want: %v, got: %v", want, got)
			}
		})
	}
}

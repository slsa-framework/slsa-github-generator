package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

// NewTestOIDCServer returns an httptest.Server that can be used as the OIDC
// server in tests and a cleanup function that can be used to stop and clean up
// the server. The server returns the given token when queried.
func NewTestOIDCServer(t *OIDCToken) (*httptest.Server, func()) {
	// FIXME: Fix creating a test server that can return tokens that can be verified.
	var issuerURL string
	s, stop := newTestOIDCServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Issuer = issuerURL
		b, err := json.Marshal(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(
			w,
			`{"value": "%s.%s.%s"}`,
			base64.RawURLEncoding.EncodeToString([]byte("{}")),
			base64.RawURLEncoding.EncodeToString(b),
			base64.RawURLEncoding.EncodeToString([]byte("{}")),
		)
	}))
	issuerURL = s.URL
	return s, stop
}

func newRawTestOIDCServer(status int, raw string) (*httptest.Server, func()) {
	return newTestOIDCServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a very basic 3-part JWT token.
		w.WriteHeader(status)
		fmt.Fprintln(w, raw)
	}))
}

func newTestOIDCServer(f http.HandlerFunc) (*httptest.Server, func()) {
	var issuerURL string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			f(w, r)
		case "/.well-known/openid-configuration":
			fmt.Fprintf(w, `{"issuer": %q}`, issuerURL)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	issuerURL = s.URL

	oldEnv, ok := os.LookupEnv(requestURLEnvKey)
	oldActionsURL := actionsProviderURL
	// NOTE: httptest.Server.URL has no trailing slash.
	os.Setenv(requestURLEnvKey, s.URL+"/")
	actionsProviderURL = s.URL
	return s, func() {
		s.Close()
		actionsProviderURL = oldActionsURL
		if ok {
			os.Setenv(requestURLEnvKey, oldEnv)
		} else {
			os.Unsetenv(requestURLEnvKey)
		}
	}
}

package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
)

// NewTestOIDCServer returns an httptest.Server that can be used as a test OIDC
// server.
func NewTestOIDCServer(t *OIDCToken) (*httptest.Server, func()) {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(`{"value": "par1.%s.part3"}`, base64.RawURLEncoding.EncodeToString(b)))
	}))
	oldEnv, ok := os.LookupEnv(requestURLEnvKey)
	// NOTE: httptest.Server.URL has no trailing slash.
	os.Setenv(requestURLEnvKey, s.URL+"/")
	return s, func() {
		s.Close()
		if ok {
			os.Setenv(requestURLEnvKey, oldEnv)
		} else {
			os.Unsetenv(requestURLEnvKey)
		}
	}
}

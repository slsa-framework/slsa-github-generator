package github

import (
	"reflect"
	"testing"
)

func TestRequestOIDCToken(t *testing.T) {
	wantToken := &OIDCToken{JobWorkflowRef: "hoge"}
	_, stop := NewTestOIDCServer(wantToken)
	defer stop()

	gotToken, err := RequestOIDCToken("fuga")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want, got := wantToken, gotToken; !reflect.DeepEqual(want, got) {
		t.Errorf("unexpected workflow ref, want: %#v, got: %#v", want, got)
	}
}

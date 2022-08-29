// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"context"
	"errors"
	"fmt"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/slsa-framework/slsa-github-generator/signing"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

type testAttestation struct {
	cert  []byte
	bytes []byte
}

func (a *testAttestation) Cert() []byte {
	return a.cert
}

func (a *testAttestation) Bytes() []byte {
	return a.bytes
}

type testSigner struct{}

func (s testSigner) Sign(context.Context, *intoto.Statement) (signing.Attestation, error) {
	return &testAttestation{}, nil
}

type tLogWithErr struct{}

var errTransparencyLog = errors.New("transparency log error")

func (tLogWithErr) Upload(context.Context, signing.Attestation) (signing.LogEntry, error) {
	fmt.Printf("Upload")
	return nil, errTransparencyLog
}

func TestGenerateProvenance_withErr(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	sha256 := "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2"
	_, err := GenerateProvenance("foo", sha256, "", "", "/home/foo", &testSigner{}, &tLogWithErr{}, &slsa.NilClientProvider{})
	if want, got := errTransparencyLog, err; want != got {
		t.Errorf("expected error, want: %v, got: %v", want, got)
	}
}

// Copyright 2022 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"context"
	"errors"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/slsa-framework/slsa-github-generator/signing"
)

// TestAttestation is a basic Attestation implementation.
type TestAttestation struct {
	CertVal  []byte
	BytesVal []byte
}

// Cert implements Attestation.Cert.
func (a *TestAttestation) Cert() []byte {
	return a.CertVal
}

// Bytes implements Attestation.Bytes.
func (a *TestAttestation) Bytes() []byte {
	return a.BytesVal
}

// TestSigner is a Signer implementation that returns the contained attestation.
type TestSigner struct {
	Att TestAttestation
}

// Sign implements Signer.Sign.
func (s TestSigner) Sign(context.Context, *intoto.Statement) (signing.Attestation, error) {
	return &s.Att, nil
}

// TransparencyLogWithErr is an implementation of TransparencyLog that returns an ErrTransparencyLog.
type TransparencyLogWithErr struct{}

// ErrTransparencyLog is returned by TransparencyLogWithErr.Upload.
var ErrTransparencyLog = errors.New("transparency log error")

// Upload implements TransparencyLog.Upload.
func (TransparencyLogWithErr) Upload(context.Context, signing.Attestation) (signing.LogEntry, error) {
	return nil, ErrTransparencyLog
}

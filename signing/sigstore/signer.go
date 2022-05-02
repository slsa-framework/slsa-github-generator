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

package sigstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/cmd/cosign/cli/rekor"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/cosign/pkg/providers"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/sigstore/pkg/signature/dsse"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
)

const (
	defaultFulcioAddr   = "https://v1.fulcio.sigstore.dev"
	defaultOIDCIssuer   = "https://oauth2.sigstore.dev/auth"
	defaultOIDCClientID = "sigstore"
	defaultRekorAddr    = "https://rekor.sigstore.dev"
)

// Signer is used to sign provenance statements and upload them to a
// transparency log.
type Signer struct {
	fulcioAddr   string
	rekorAddr    string
	oidcIssuer   string
	oidcClientID string
}

// Attestation is a signed attestation.
type Attestation struct {
	cert []byte
	att  []byte
}

// Bytes returns the signed attestation as an encoded DSSE JSON envelope.
func (a Attestation) Bytes() []byte {
	return a.att
}

// Cert returns the certificate used to sign the attestation.
func (a Attestation) Cert() []byte {
	return a.cert
}

// NewDefaultSigner creates a new signer with the default values.
func NewDefaultSigner() Signer {
	return NewSigner(defaultFulcioAddr, defaultRekorAddr, defaultOIDCIssuer, defaultOIDCClientID)
}

// NewSigner creates a new Signer.
func NewSigner(fulcioAddr, rekorAddr, oidcIssuer, oidcClientID string) Signer {
	return Signer{
		fulcioAddr:   fulcioAddr,
		rekorAddr:    rekorAddr,
		oidcIssuer:   oidcIssuer,
		oidcClientID: oidcClientID,
	}
}

// Sign signs the given provenance statement and returns the signed
// attestation.
func (s *Signer) Sign(ctx context.Context, p *intoto.ProvenanceStatement) (*Attestation, error) {
	// Get Fulcio signer
	if !providers.Enabled(ctx) {
		return nil, fmt.Errorf("no auth provider is enabled. Are you running outside of Github Actions?")
	}

	attBytes, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshalling json: %w", err)
	}

	fClient, err := fulcio.NewClient(s.fulcioAddr)
	if err != nil {
		return nil, fmt.Errorf("creating fulcio client: %w", err)
	}
	tok, err := providers.Provide(ctx, s.oidcClientID)
	if err != nil {
		return nil, fmt.Errorf("obtaining cosign provider: %w", err)
	}
	k, err := fulcio.NewSigner(ctx, tok, s.oidcIssuer, s.oidcClientID, "", fClient)
	if err != nil {
		return nil, fmt.Errorf("creating fulcio signer: %w", err)
	}
	wrappedSigner := dsse.WrapSigner(k, intoto.PayloadType)

	signedAtt, err := wrappedSigner.SignMessage(bytes.NewReader(attBytes))
	if err != nil {
		return nil, fmt.Errorf("signing message: %v", err)
	}

	return &Attestation{
		att:  signedAtt,
		cert: k.Cert,
	}, nil
}

// Upload uploads the signed attestation to the rekor transparency log.
func (s *Signer) Upload(ctx context.Context, att *Attestation) (*models.LogEntryAnon, error) {
	rekorClient, err := rekor.NewClient(s.rekorAddr)
	if err != nil {
		return nil, fmt.Errorf("creating rekor client: %w", err)
	}
	// TODO: Is it a bug that we need []byte(string(k.Cert)) or else we hit invalid PEM?
	logEntry, err := cosign.TLogUploadInTotoAttestation(ctx, rekorClient, att.att, []byte(string(att.cert)))
	if err != nil {
		return nil, fmt.Errorf("uploading attestation: %w", err)
	}

	return logEntry, nil
}

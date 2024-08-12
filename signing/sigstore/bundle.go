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
	"context"
	"encoding/json"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	sigstoreBundle "github.com/sigstore/sigstore-go/pkg/bundle"
	sigstoreSign "github.com/sigstore/sigstore-go/pkg/sign"
	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/signing"
)

// BundleSigner is used to produce Sigstore Bundles from provenance statements.
type BundleSigner struct {
}

type sigstoreBundleAtt struct {
	cert []byte
	att  []byte
}

// Cert returns the certificate used to sign the Bundle.
func (s *sigstoreBundleAtt) Cert() []byte {
	return s.cert
}

// attestation is a signed Sigstore Bundle.
func (s *sigstoreBundleAtt) Bytes() []byte {
	return s.att
}

// NewDefaultBundleSigner creates a new BundleSigner instance.
func NewDefaultBundleSigner() *BundleSigner {
	return &BundleSigner{}
}

// Sign signs the given provenance statement and returns the signed Sigstore Bundle.
func (s *BundleSigner) Sign(ctx context.Context, statement *intoto.Statement) (signing.Attestation, error) {
	// content to sign
	statementBytes, err := json.Marshal(*statement)
	if err != nil {
		return nil, err
	}
	content := &sigstoreSign.DSSEData{
		Data:        statementBytes,
		PayloadType: intoto.PayloadType,
	}

	// keypair for the certificate
	keypair, err := sigstoreSign.NewEphemeralKeypair(nil)
	if err != nil {
		return nil, err
	}

	// get the oidc token.
	oidcClient, err := github.NewOIDCClient()
	if err != nil {
		return nil, err
	}
	TokenStruct, err := oidcClient.Token(ctx, []string{"sigstore"})
	if err != nil {
		return nil, err
	}
	rawToken := TokenStruct.RawToken

	// signing opts.
	bundleOpts, err := getDefaultBundleOptsWithIdentityToken(&rawToken)

	// sign.
	innerBundle, err := sigstoreSign.Bundle(content, keypair, *bundleOpts)
	if err != nil {
		return nil, err
	}

	// marshall to json.
	bundleWrapper := &sigstoreBundle.ProtobufBundle{
		Bundle: innerBundle,
	}
	bundleBytes, err := bundleWrapper.MarshalJSON()
	if err != nil {
		return nil, err
	}
	bundleAtt := &sigstoreBundleAtt{
		cert: innerBundle.GetVerificationMaterial().GetCertificate().GetRawBytes(),
		att:  bundleBytes,
	}
	return bundleAtt, nil
}

// getDefaultBundleOptsWithIdentityToken provides the default opts for sigstoreSign.Bundle().
func getDefaultBundleOptsWithIdentityToken(identityToken *string) (*sigstoreSign.BundleOptions, error) {
	bundleOpts := &sigstoreSign.BundleOptions{}

	fulcioOpts := &sigstoreSign.FulcioOptions{
		BaseURL: "https://fulcio.sigstore.dev",
	}
	bundleOpts.CertificateProvider = sigstoreSign.NewFulcio(fulcioOpts)
	bundleOpts.CertificateProviderOptions = &sigstoreSign.CertificateProviderOptions{
		IDToken: *identityToken,
	}

	rekorOpts := &sigstoreSign.RekorOptions{
		BaseURL: "https://rekor.sigstore.dev",
	}
	bundleOpts.TransparencyLogs = append(bundleOpts.TransparencyLogs, sigstoreSign.NewRekor(rekorOpts))
	return bundleOpts, nil
}

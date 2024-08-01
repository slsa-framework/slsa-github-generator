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

package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/spf13/cobra"

	sigstoreBundle "github.com/sigstore/sigstore-go/pkg/bundle"
	sigstoreRoot "github.com/sigstore/sigstore-go/pkg/root"
	sigstoreSign "github.com/sigstore/sigstore-go/pkg/sign"
	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/builders/common"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/signing"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

// attestCmd returns the 'attest' command.
func attestCmd(provider slsa.ClientProvider, check func(error),
	signer signing.Signer, tlog signing.TransparencyLog,
) *cobra.Command {
	var attPath string
	var subjectsFilename string

	c := &cobra.Command{
		Use:   "attest",
		Short: "Create a signed SLSA provenance attestation from a Github Action",
		Long: `Generate and sign SLSA provenance from a Github Action to form an attestation
and upload to a Rekor transparency log. This command assumes that it is being
run in the context of a Github Actions workflow.`,

		Run: func(_ *cobra.Command, _ []string) {
			ghContext, err := github.GetWorkflowContext()
			check(err)

			varsContext, err := github.GetVarsContext()
			check(err)

			subjectsBytes, err := utils.SafeReadFile(subjectsFilename)
			check(err)
			parsedSubjects, err := parseSubjects(string(subjectsBytes))
			check(err)
			if len(parsedSubjects) == 0 {
				check(errors.New("expected at least one subject"))
			}

			// NOTE: The provenance file path is untrusted and should be
			// validated. This is done by CreateNewFileUnderCurrentDirectory.
			if attPath == "" {
				if len(parsedSubjects) == 1 {
					filename := path.Base(parsedSubjects[0].Name)
					attPath = fmt.Sprintf("%s.intoto.jsonl", filename)
				} else {
					// len(parsedSubjects) > 1
					attPath = "multiple.intoto.jsonl"
				}
			}

			// Verify the extension path and extension.
			err = utils.VerifyAttestationPath(attPath)
			check(err)

			ctx := context.Background()

			b := common.GenericBuild{
				GithubActionsBuild: slsa.NewGithubActionsBuild(parsedSubjects, &ghContext, varsContext),
				BuildTypeURI:       provenanceOnlyBuildType,
			}
			if provider != nil {
				b.WithClients(provider)
			} else if utils.IsPresubmitTests() {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				b.WithClients(&slsa.NilClientProvider{})
			}

			g := slsa.NewHostedActionsGenerator(&b)
			if provider != nil {
				g.WithClients(provider)
			} else if utils.IsPresubmitTests() {
				// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
				g.WithClients(&slsa.NilClientProvider{})
			}

			p, err := g.Generate(ctx)
			check(err)

			// Note: the path is validated within CreateNewFileUnderCurrentDirectory().
			var attBytes []byte
			if utils.IsPresubmitTests() {
				attBytes, err = json.Marshal(p)
				check(err)
			} else {
				// att, err := signer.Sign(ctx, &intoto.Statement{
				// 	StatementHeader: p.StatementHeader,
				// 	Predicate:       p.Predicate,
				// })
				// check(err)

				// _, err = tlog.Upload(ctx, att)
				// check(err)

				// attBytes = att.Bytes()

				att, err := makeSigstoreBundleAttestation(ctx, &intoto.Statement{
					StatementHeader: p.StatementHeader,
					Predicate:       p.Predicate,
				})
				check(err)

				attBytes = att.Bytes()
			}

			f, err := utils.CreateNewFileUnderCurrentDirectory(attPath, os.O_WRONLY)
			check(err)

			_, err = f.Write(attBytes)
			check(err)

			// Print the provenance name and sha256 so it can be used by the workflow.
			check(github.SetOutput("provenance-name", attPath))
			check(github.SetOutput("provenance-sha256", fmt.Sprintf("%x", sha256.Sum256(attBytes))))
		},
	}

	c.Flags().StringVarP(
		&attPath, "signature", "g", "",
		"Path to write the signed provenance.",
	)
	c.Flags().StringVarP(
		&subjectsFilename, "subjects-filename", "f", "",
		"Filename containing a formatted list of subjects in the same format as sha256sum (base64 encoded).",
	)
	return c
}

func makeSigstoreBundleAttestation(ctx context.Context, statement *intoto.Statement) (signing.Attestation, error) {
	fmt.Println("debug: running makeSigstoreBundle")
	statementBytes, err := json.Marshal(statement)
	if err != nil {
		return nil, err
	}
	content := &sigstoreSign.DSSEData{
		Data:        statementBytes,
		PayloadType: "application/vnd.in-toto+json",
	}

	keypair, err := sigstoreSign.NewEphemeralKeypair(nil)
	if err != nil {
		return nil, err
	}

	oidcClient, err := github.NewOIDCClient()
	if err != nil {
		return nil, err
	}
	TokenStruct, err := oidcClient.Token(ctx, []string{"sigstore"})
	if err != nil {
		return nil, err
	}
	rawToken := TokenStruct.RawToken

	bundleOpts, err := getDefaultBundleOptsWithIdentityToken(&rawToken)
	innerBundle, err := sigstoreSign.Bundle(content, keypair, *bundleOpts)
	if err != nil {
		return nil, err
	}
	outerBundle := &sigstoreBundle.ProtobufBundle{
		Bundle: innerBundle,
	}
	bundleBytes, err := outerBundle.MarshalJSON()
	if err != nil {
		return nil, err
	}

	bundleAtt := &sigstoreBundleAtt{
		cert:  innerBundle.GetVerificationMaterial().GetCertificate().GetRawBytes(),
		bytes: bundleBytes,
	}

	fmt.Println(fmt.Sprintf("debug: generated bundle attestation: %s", bundleAtt.Bytes()))
	return bundleAtt, nil
}

type sigstoreBundleAtt struct {
	cert  []byte
	bytes []byte
}

func (s *sigstoreBundleAtt) Cert() []byte {
	return s.cert
}
func (s *sigstoreBundleAtt) Bytes() []byte {
	return s.bytes
}

func getDefaultBundleOptsWithIdentityToken(identityToken *string) (*sigstoreSign.BundleOptions, error) {
	bundleOpts := &sigstoreSign.BundleOptions{}

	trustedRoot, err := sigstoreRoot.FetchTrustedRoot()
	if err != nil {
		return nil, err
	}
	bundleOpts.TrustedRoot = trustedRoot
	bundleOpts.TrustedRoot = nil

	fulcioOpts := &sigstoreSign.FulcioOptions{
		BaseURL: "https://fulcio.sigstore.dev",
		Timeout: time.Duration(30 * time.Second),
		Retries: 1,
	}
	bundleOpts.CertificateProvider = sigstoreSign.NewFulcio(fulcioOpts)
	bundleOpts.CertificateProviderOptions = &sigstoreSign.CertificateProviderOptions{
		IDToken: *identityToken,
	}

	tsaOpts := &sigstoreSign.TimestampAuthorityOptions{
		URL:     "https://timestamp.githubapp.com/api/v1/timestamp",
		Timeout: time.Duration(30 * time.Second),
		Retries: 1,
	}
	bundleOpts.TimestampAuthorities = append(bundleOpts.TimestampAuthorities, sigstoreSign.NewTimestampAuthority(tsaOpts))

	rekorOpts := &sigstoreSign.RekorOptions{
		BaseURL: "https://rekor.sigstore.dev",
		Timeout: time.Duration(90 * time.Second),
		Retries: 1,
	}
	bundleOpts.TransparencyLogs = append(bundleOpts.TransparencyLogs, sigstoreSign.NewRekor(rekorOpts))
	return bundleOpts, nil
}

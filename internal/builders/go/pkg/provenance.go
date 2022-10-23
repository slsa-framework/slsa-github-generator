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
	"encoding/hex"
	"fmt"
	"os"

	"github.com/slsa-framework/slsa-github-generator/signing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

const (
	buildConfigVersion int = 1
	buildType              = "https://github.com/slsa-framework/slsa-github-generator/go@v1"
)

type (
	step struct {
		Command    []string `json:"command"`
		Env        []string `json:"env"`
		WorkingDir string   `json:"workingDir"`
	}
	buildConfig struct {
		Version int    `json:"version"`
		Steps   []step `json:"steps"`
	}
)

type goProvenanceBuild struct {
	*slsa.GithubActionsBuild
	buildConfig buildConfig
}

// URI implements BuildType.URI.
func (b *goProvenanceBuild) URI() string {
	return buildType
}

// BuildConfig implements BuildType.BuildConfig.
func (b *goProvenanceBuild) BuildConfig(context.Context) (interface{}, error) {
	return b.buildConfig, nil
}

// GenerateProvenance translates github context into a SLSA provenance
// attestation.
// Spec: https://slsa.dev/provenance/v0.2
func GenerateProvenance(name, digest, command, envs, workingDir string, s signing.Signer, r signing.TransparencyLog, provider slsa.ClientProvider) ([]byte, error) {
	gh, err := github.GetWorkflowContext()
	if err != nil {
		return nil, err
	}

	if _, err := hex.DecodeString(digest); err != nil || len(digest) != 64 {
		return nil, fmt.Errorf("sha256 digest is not valid: %s", digest)
	}

	com, err := utils.UnmarshalList(command)
	if err != nil {
		return nil, err
	}

	env, err := utils.UnmarshalList(envs)
	if err != nil {
		return nil, err
	}

	var cmd []string
	if len(com) > 0 {
		cmd = []string{com[0], "mod", "vendor"}
	}

	b := goProvenanceBuild{
		GithubActionsBuild: slsa.NewGithubActionsBuild([]intoto.Subject{
			{
				Name: name,
				Digest: slsacommon.DigestSet{
					"sha256": digest,
				},
			},
		}, gh),
		buildConfig: buildConfig{
			Version: buildConfigVersion,
			Steps: []step{
				// Vendoring step.
				{
					// Note: vendoring and compilation are
					// performed in the same VM, so the compiler is
					// the same.
					Command:    cmd,
					WorkingDir: workingDir,
					// Note: No user-defined env set for this step.
				},
				// Compilation step.
				{
					Command:    com,
					Env:        env,
					WorkingDir: workingDir,
				},
			},
		},
	}

	// Pre-submit tests don't have access to write OIDC token.
	if provider != nil {
		b.WithClients(provider)
	} else {
		// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
		if utils.IsPresubmitTests() {
			b.GithubActionsBuild.WithClients(&slsa.NilClientProvider{})
		}
	}

	ctx := context.Background()
	g := slsa.NewHostedActionsGenerator(&b)
	// Pre-submit tests don't have access to write OIDC token.
	if provider != nil {
		g.WithClients(provider)
	} else {
		// TODO(github.com/slsa-framework/slsa-github-generator/issues/124): Remove
		if utils.IsPresubmitTests() {
			g.WithClients(&slsa.NilClientProvider{})
		}
	}
	p, err := g.Generate(ctx)
	if err != nil {
		return nil, err
	}

	// Set the architecture based on the runner. Architecture should be the
	// same for the provenance step where this is run and the build step if the
	// reusable workflow is used.
	//
	// NOTE: map is a reference so modifying invEnv modifies
	// p.Predicate.Invocation.Environment.
	invEnv := p.Predicate.Invocation.Environment.(map[string]interface{})
	invEnv["arch"] = os.Getenv("RUNNER_ARCH")
	invEnv["os"] = os.Getenv("ImageOS")

	// Add details about the runner's OS to the materials
	runnerMaterials := slsacommon.ProvenanceMaterial{
		// TODO: capture the digest here too
		URI: fmt.Sprintf("https://github.com/actions/virtual-environments/releases/tag/%s/%s", os.Getenv("ImageOS"), os.Getenv("ImageVersion")),
	}
	p.Predicate.Materials = append(p.Predicate.Materials, runnerMaterials)

	if utils.IsPresubmitTests() {
		fmt.Println("Pre-submit tests detected. Skipping signing.")
		return utils.MarshalToBytes(*p)
	}

	// Sign the provenance.
	att, err := s.Sign(ctx, &intoto.Statement{
		StatementHeader: p.StatementHeader,
		Predicate:       p.Predicate,
	})
	if err != nil {
		return nil, err
	}

	// Upload the signed attestation to rekor.
	logEntry, err := r.Upload(ctx, att)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Uploaded signed attestation to rekor with UUID %s.\n", logEntry.UUID())

	return att.Bytes(), nil
}

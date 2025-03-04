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

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/signing/sigstore"

	// Enable the GitHub OIDC auth provider.
	_ "github.com/sigstore/cosign/v2/pkg/providers/github"

	"github.com/slsa-framework/slsa-github-generator/internal/builders/go/pkg"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

func usage(p string) {
	panic(fmt.Sprintf(`Usage:
	 %s build [--dry] slsa-releaser.yml
	 %s provenance --binary-name $NAME --digest $DIGEST --command $COMMAND --env $ENV`, p, p))
}

func check(e error) {
	if e != nil {
		fmt.Fprint(os.Stderr, e.Error())
		os.Exit(1)
	}
}

func runBuild(dry bool, configFile, evalEnvs string) error {
	goc, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	cfg, err := pkg.ConfigFromFile(configFile)
	if err != nil {
		return err
	}
	fmt.Println(cfg)

	gobuild := pkg.GoBuildNew(goc, cfg)

	// Set env variables encoded as arguments.
	err = gobuild.SetArgEnvVariables(evalEnvs)
	if err != nil {
		return err
	}

	err = gobuild.Run(dry)
	if err != nil {
		return err
	}

	return nil
}

func runProvenanceGeneration(subject, digest, commands, envs, workingDir string) error {
	s := sigstore.NewDefaultBundleSigner()

	attBytes, err := pkg.GenerateProvenance(subject, digest,
		commands, envs, workingDir, s, nil)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.intoto.jsonl", subject)
	f, err := utils.CreateNewFileUnderCurrentDirectory(filename, os.O_WRONLY)
	if err != nil {
		return err
	}
	_, err = f.Write(attBytes)
	if err != nil {
		return err
	}

	if err := github.SetOutput("signed-provenance-name", filename); err != nil {
		return err
	}

	h, err := computeSHA256(attBytes)
	if err != nil {
		return err
	}

	return github.SetOutput("signed-provenance-sha256", h)
}

func main() {
	// Build command.
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildDry := buildCmd.Bool("dry", false, "dry run of the build without invoking compiler")

	// Provenance command.
	provenanceCmd := flag.NewFlagSet("provenance", flag.ExitOnError)
	provenanceName := provenanceCmd.String("binary-name", "", "untrusted binary name of the artifact built")
	provenanceDigest := provenanceCmd.String("digest", "", "sha256 digest of the untrusted binary")
	provenanceCommand := provenanceCmd.String("command", "", "command used to compile the binary")
	provenanceEnv := provenanceCmd.String("env", "", "env variables used to compile the binary")
	provenanceWorkingDir := provenanceCmd.String("workingDir", "", "working directory used to issue compilation commands")

	// Expect a sub-command.
	if len(os.Args) < 2 {
		usage(os.Args[0])
	}

	switch os.Args[1] {
	case buildCmd.Name():
		check(buildCmd.Parse(os.Args[2:]))
		if len(buildCmd.Args()) < 1 {
			usage(os.Args[0])
		}
		configFile := buildCmd.Args()[0]
		evaluatedEnvs := buildCmd.Args()[1]

		check(runBuild(*buildDry, configFile, evaluatedEnvs))

	case provenanceCmd.Name():
		check(provenanceCmd.Parse(os.Args[2:]))
		// Note: *provenanceEnv may be empty.
		if *provenanceName == "" || *provenanceDigest == "" ||
			*provenanceCommand == "" || *provenanceWorkingDir == "" {
			usage(os.Args[0])
		}

		err := runProvenanceGeneration(*provenanceName, *provenanceDigest,
			*provenanceCommand, *provenanceEnv, *provenanceWorkingDir)
		check(err)

	default:
		fmt.Println("expected 'build' or 'provenance' subcommands")
		os.Exit(1)
	}
}

func computeSHA256(data []byte) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(data)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

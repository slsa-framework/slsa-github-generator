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
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsav02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/slsa-framework/slsa-github-generator/internal/errors"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

func checkExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func checkTest(t *testing.T) func(err error) {
	return func(err error) {
		if err != nil {
			t.Fatalf("%v", err)
		}
	}
}

var (
	// shaCheck verifies a hash is has only hexadecimal digits and is 64
	// characters long.
	shaCheck = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

	// wsSplit is used to split lines in the subjects input.
	wsSplit = regexp.MustCompile(`[\t ]`)

	// provenanceOnlyBuildType is the URI for provenance only SLSA generation.
	provenanceOnlyBuildType = "https://github.com/slsa-framework/slsa-github-generator/generic@v1"
)

// errBase64 indicates a base64 error in the subject.
type errBase64 struct {
	errors.WrappableError
}

// errSha indicates a error in the hash format.
type errSha struct {
	errors.WrappableError
}

// errNoName indicates a missing subject name.
type errNoName struct {
	errors.WrappableError
}

// errDuplicateSubject indicates a duplicate subject name.
type errDuplicateSubject struct {
	errors.WrappableError
}

// errScan is an error scanning the SHA digest data.
type errScan struct {
	errors.WrappableError
}

// parseSubjects parses the value given to the subjects option.
func parseSubjects(b64str string) ([]intoto.Subject, error) {
	var parsed []intoto.Subject

	subjects, err := base64.StdEncoding.DecodeString(b64str)
	if err != nil {
		return nil, errors.Errorf(&errBase64{}, "error decoding subjects (is it base64 encoded?): %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(subjects))
	for scanner.Scan() {
		// Split by whitespace, and get values.
		parts := wsSplit.Split(strings.TrimSpace(scanner.Text()), 2)

		// Lowercase the sha digest to comply with the SLSA spec.
		shaDigest := strings.ToLower(strings.TrimSpace(parts[0]))
		if shaDigest == "" {
			// Ignore empty lines.
			continue
		}
		// Do a sanity check on the SHA to make sure it's a proper hex digest.
		if !shaCheck.MatchString(shaDigest) {
			return nil, errors.Errorf(&errSha{}, "unexpected sha256 hash format for %q", shaDigest)
		}

		// Check for the subject name.
		if len(parts) == 1 {
			return nil, errors.Errorf(&errNoName{}, "expected subject name for hash %q", shaDigest)
		}
		name := strings.TrimSpace(parts[1])

		for _, p := range parsed {
			if p.Name == name {
				return nil, errors.Errorf(&errDuplicateSubject{}, "duplicate subject %q", name)
			}
		}

		parsed = append(parsed, intoto.Subject{
			Name: name,
			Digest: slsav02.DigestSet{
				"sha256": shaDigest,
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Errorf(&errScan{}, "reading digest: %w", err)
	}

	return parsed, nil
}

type provenanceOnlyBuild struct {
	*slsa.GithubActionsBuild
}

// URI implements BuildType.URI.
func (b *provenanceOnlyBuild) URI() string {
	return provenanceOnlyBuildType
}

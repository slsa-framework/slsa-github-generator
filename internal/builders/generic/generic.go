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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
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

var (
	// errBase64 indicates a base64 error in the subject.
	errBase64 = errors.New("base64")

	// errSha indicates a error in the hash format.
	errSha = errors.New("sha")

	// errSubjectName indicates a subject name error.
	errSubjectName = errors.New("subject name")

	// errDuplicateSubject indicates a duplicate subject name.
	errDuplicateSubject = errors.New("duplicate subject")

	// errScan is an error scanning the SHA digest data.
	errScan = errors.New("subjects")
)

// parseSubjects parses the value given to the subjects option.
func parseSubjects(filename string) ([]intoto.Subject, error) {
	var parsed []intoto.Subject

	subjects_bytes, err := utils.SafeReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: error reading file", err)
	}
	subjects, err := base64.StdEncoding.DecodeString(string(subjects_bytes))
	if err != nil {
		return nil, fmt.Errorf("%w: error decoding subjects (is it base64 encoded?): %w", errBase64, err)
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
			return nil, fmt.Errorf("%w: unexpected sha256 hash format for %q", errSha, shaDigest)
		}

		// Check for the subject name.
		if len(parts) == 1 {
			return nil, fmt.Errorf("%w: expected subject name for hash %q", errSubjectName, shaDigest)
		}
		name := strings.TrimSpace(parts[1])

		for _, p := range parsed {
			if p.Name == name {
				return nil, fmt.Errorf("%w: %q", errDuplicateSubject, name)
			}
		}

		parsed = append(parsed, intoto.Subject{
			Name: name,
			Digest: slsacommon.DigestSet{
				"sha256": shaDigest,
			},
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: reading digest: %w", errScan, err)
	}

	return parsed, nil
}

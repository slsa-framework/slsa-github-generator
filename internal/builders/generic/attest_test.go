// Copyright 2023 SLSA Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"

	"github.com/slsa-framework/slsa-github-generator/internal/testutil"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

const (
	testHash = "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c  artifact1"
)

// TestParseSubjects tests the parseSubjects function.
func TestParseSubjects(t *testing.T) {
	errNoNameFunc := func(got error) {
		want := errSubjectName
		if !errors.Is(got, want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errShaFunc := func(got error) {
		want := errSha
		if !errors.Is(got, want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errDuplicateSubjectFunc := func(got error) {
		want := errDuplicateSubject
		if !errors.Is(got, want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errBase64Func := func(got error) {
		want := errBase64
		if !errors.Is(got, want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	testCases := []struct {
		name     string
		str      string
		err      func(error)
		expected []intoto.Subject
	}{
		{
			name: "single",
			// echo "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCg==",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsacommon.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
			},
		},
		{
			name: "name has spaces",
			// echo "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge fuga" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlIGZ1Z2EK",
			expected: []intoto.Subject{
				{
					Name: "hoge fuga",
					Digest: slsacommon.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
			},
		},
		{
			name: "extra whitespace",
			// echo -e "\t  2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \
			// \t hoge fuga  \t  " | base64 -w0
			str: "CSAgMmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3Y" +
				"mFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiAJIGhvZ2UgZnVnYSAgCSAgCg==",
			expected: []intoto.Subject{
				{
					Name: "hoge fuga",
					Digest: slsacommon.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
			},
		},

		{
			name: "multiple",
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \
			// hoge\ne712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCmU" +
				"3MTJhZmYzNzA1YWMzMTRiOWE4OTBlMGVjMjA4ZmFhMjAwNTRlZWU1MTRkODZhYjkxM2Q3NjhmOTRlMDEyNzkgZnVnYQo=",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsacommon.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
				{
					Name: "fuga",
					Digest: slsacommon.DigestSet{
						"sha256": "e712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279",
					},
				},
			},
		},
		{
			name:     "empty",
			str:      "",
			expected: nil,
		},
		{
			name: "blank lines",
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \
			// hoge\n\ne712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCgp" +
				"lNzEyYWZmMzcwNWFjMzE0YjlhODkwZTBlYzIwOGZhYTIwMDU0ZWVlNTE0ZDg2YWI5MTNkNzY4Zjk0ZTAxMjc5IGZ1Z2EK",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsacommon.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
				{
					Name: "fuga",
					Digest: slsacommon.DigestSet{
						"sha256": "e712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279",
					},
				},
			},
		},
		{
			name: "sha only",
			// echo "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMgo=",
			// err: &errNoName{},
			err: errNoNameFunc,
		},
		{
			name: "invalid hash",
			// echo "abcdef hoge" | base64 -w0
			str: "YWJjZGVmIGhvZ2UK",
			err: errShaFunc,
		},
		{
			name: "duplicate name",
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \
			// hoge\n2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dl" +
				"CjJlMDM5MGViMDI0YTUyOTYzZGI3Yjk1ZTg0YTljMmIxMmMwMDQwNTRhN2JhZDlhOTdlYzBjN2M4OWQ0NjgxZDIgaG9nZQo=",
			err: errDuplicateSubjectFunc,
		},
		{
			name: "not base64",
			str:  "this is not base64",
			err:  errBase64Func,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if s, err := parseSubjects(tc.str); err != nil {
				if tc.err != nil {
					tc.err(err)
				}
			} else {
				if tc.err != nil {
					tc.err(err)
				}

				if got, want := s, tc.expected; !cmp.Equal(want, got) {
					t.Errorf("unexpected subjects, got: %#v, want: %#v", got, want)
				}
			}
		})
	}
}

func createTmpFile(content string) (string, error) {
	file, err := os.CreateTemp(".", "test-")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.Write([]byte(content)); err != nil {
		return "", err
	}
	return file.Name(), nil
}

// Test_attestCmd tests the attest command.
func Test_attestCmd_default_single_artifact(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(testHash)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "artifact1.intoto.jsonl")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_attestCmd_default_multi_artifact(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(
		`b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c  artifact1
b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c  artifact2`)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "multiple.intoto.jsonl")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_attestCmd_custom_provenance_name(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(testHash)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
		"--signature", "custom.intoto.jsonl",
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the file exists.
	if _, err := os.Stat("custom.intoto.jsonl"); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_attestCmd_invalid_extension(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	// A custom check function that checks the error type is the expected error type.
	check := func(err error) {
		if err != nil {
			got, want := err, utils.ErrInvalidPath
			if !errors.Is(got, want) {
				t.Fatalf("expected error, got: %v, want: %v", got, want)
			}
			// Check should exit the program so we skip the rest of the test if we got the expected error.
			t.SkipNow()
		}
	}

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(testHash)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, check, &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
		"--signature", "invalid_name",
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// If no error occurs we catch it here. SkipNow will exit the test process so this code should be unreachable.
	t.Errorf("expected an error to occur.")
}

func Test_attestCmd_invalid_path(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	// A custom check function that checks the error type is the expected error type.
	check := func(err error) {
		if err != nil {
			got, want := err, utils.ErrInvalidPath
			if !errors.Is(got, want) {
				t.Fatalf("unexpected error, got: %v, want: %v", got, want)
			}
			// Check should exit the program so we skip the rest of the test if we got the expected error.
			t.SkipNow()
		}
	}

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(testHash)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, check, &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
		"--signature", "/provenance.intoto.jsonl",
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// If no error occurs we catch it here. SkipNow will exit the test process so this code should be unreachable.
	t.Errorf("expected an error to occur.")
}

// Test_attestCmd_subdirectory_artifact tests the attest command when provided
// subjects in subdirectories.
func Test_attestCmd_subdirectory_artifact(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")
	t.Setenv("VARS_CONTEXT", "{}")

	// Change to temporary dir
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	fn, err := createTmpFile(base64.StdEncoding.EncodeToString([]byte(testHash)))
	if err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	defer os.Remove(fn)
	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects-filename", fn,
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "artifact1.intoto.jsonl")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

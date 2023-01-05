package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsacommon "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
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
		want := &errNoName{}
		if !errors.As(got, &want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errShaFunc := func(got error) {
		want := &errSha{}
		if !errors.As(got, &want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errDuplicateSubjectFunc := func(got error) {
		want := &errDuplicateSubject{}
		if !errors.As(got, &want) {
			t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
		}
	}

	errBase64Func := func(got error) {
		want := &errBase64{}
		if !errors.As(got, &want) {
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

				if want, got := tc.expected, s; !cmp.Equal(want, got) {
					t.Errorf("unexpected subjects, want: %#v, got: %#v", want, got)
				}
			}
		})
	}
}

// Test_attestCmd tests the attest command.
func Test_attestCmd_default_single_artifact(t *testing.T) {
	t.Setenv("GITHUB_CONTEXT", "{}")

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
	defer os.Chdir(currentDir)

	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(testHash)),
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
	defer os.Chdir(currentDir)

	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(
			`b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c  artifact1
b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c  artifact2`)),
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
	defer os.Chdir(currentDir)

	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(testHash)),
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
	defer os.Chdir(currentDir)

	// A custom check function that checks the error type is the expected error type.
	check := func(err error) {
		if err != nil {
			errInvalidPath := &utils.ErrInvalidPath{}
			if !errors.As(err, &errInvalidPath) {
				t.Fatalf("expected %v but got %v", &utils.ErrInvalidPath{}, err)
			}
			// Check should exit the program so we skip the rest of the test if we got the expected error.
			t.SkipNow()
		}
	}

	c := attestCmd(&slsa.NilClientProvider{}, check, &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(testHash)),
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
	defer os.Chdir(currentDir)

	// A custom check function that checks the error type is the expected error type.
	check := func(err error) {
		if err != nil {
			errInvalidPath := &utils.ErrInvalidPath{}
			if !errors.As(err, &errInvalidPath) {
				t.Fatalf("expected %v but got %v", &utils.ErrInvalidPath{}, err)
			}
			// Check should exit the program so we skip the rest of the test if we got the expected error.
			t.SkipNow()
		}
	}

	c := attestCmd(&slsa.NilClientProvider{}, check, &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(testHash)),
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
	defer os.Chdir(currentDir)

	c := attestCmd(&slsa.NilClientProvider{}, checkTest(t), &testutil.TestSigner{}, &testutil.TestTransparencyLog{})
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{
		"--subjects", base64.StdEncoding.EncodeToString([]byte(testHash)),
	})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "artifact1.intoto.jsonl")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

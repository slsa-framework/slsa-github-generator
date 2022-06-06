package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsav02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

// TestParseSubjects tests the parseSubjects function.
func TestParseSubjects(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		expected []intoto.Subject
		err      error
	}{
		{
			name: "single",
			// echo "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCg==",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsav02.DigestSet{
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
					Digest: slsav02.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
			},
		},
		{
			name: "extra whitespace",
			// echo -e "\t  2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \t hoge fuga  \t  " | base64 -w0
			str: "CSAgMmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiAJIGhvZ2UgZnVnYSAgCSAgCg==",
			expected: []intoto.Subject{
				{
					Name: "hoge fuga",
					Digest: slsav02.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
			},
		},

		{
			name: "multiple",
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge\ne712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCmU3MTJhZmYzNzA1YWMzMTRiOWE4OTBlMGVjMjA4ZmFhMjAwNTRlZWU1MTRkODZhYjkxM2Q3NjhmOTRlMDEyNzkgZnVnYQo=",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsav02.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
				{
					Name: "fuga",
					Digest: slsav02.DigestSet{
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
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge\n\ne712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCgplNzEyYWZmMzcwNWFjMzE0YjlhODkwZTBlYzIwOGZhYTIwMDU0ZWVlNTE0ZDg2YWI5MTNkNzY4Zjk0ZTAxMjc5IGZ1Z2EK",
			expected: []intoto.Subject{
				{
					Name: "hoge",
					Digest: slsav02.DigestSet{
						"sha256": "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
					},
				},
				{
					Name: "fuga",
					Digest: slsav02.DigestSet{
						"sha256": "e712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279",
					},
				},
			},
		},
		{
			name: "sha only",
			// echo "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMgo=",
			err: &errNoName{},
		},
		{
			name: "invalid hash",
			// echo "abcdef hoge" | base64 -w0
			str: "YWJjZGVmIGhvZ2UK",
			err: &errSha{},
		},
		{
			name: "duplicate name",
			// echo -e "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge\n2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge" | base64 -w0
			str: "MmUwMzkwZWIwMjRhNTI5NjNkYjdiOTVlODRhOWMyYjEyYzAwNDA1NGE3YmFkOWE5N2VjMGM3Yzg5ZDQ2ODFkMiBob2dlCjJlMDM5MGViMDI0YTUyOTYzZGI3Yjk1ZTg0YTljMmIxMmMwMDQwNTRhN2JhZDlhOTdlYzBjN2M4OWQ0NjgxZDIgaG9nZQo=",
			err: &errDuplicateSubject{},
		},
		{
			name: "not base64",
			str:  "this is not base64",
			err:  &errBase64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if s, err := parseSubjects(tc.str); err != nil {
				if tc.err != nil && !errors.As(err, &tc.err) {
					t.Fatalf("unexpected error: %v", cmp.Diff(err, tc.err, cmpopts.EquateErrors()))
				}
			} else {
				if tc.err != nil {
					t.Fatalf("expected %#v but received %#v", tc.err, s)
				}

				if want, got := tc.expected, s; !cmp.Equal(want, got) {
					t.Errorf("unexpected subjects, want: %#v, got: %#v", want, got)
				}
			}
		})
	}
}

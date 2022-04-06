package main

import (
	"reflect"
	"testing"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsav02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
)

// TestParseSubjects tests the parseSubjects function.
func TestParseSubjects(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		expected []intoto.Subject
		err      bool
	}{
		{
			name: "single",
			str:  "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge",
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
			str:  "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge fuga",
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
			str:  "\t  2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 \t hoge fuga  \t  ",
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
			str: `2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge
e712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga`,
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
			str: `2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2 hoge

e712aff3705ac314b9a890e0ec208faa20054eee514d86ab913d768f94e01279 fuga`,
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
			str:  "2e0390eb024a52963db7b95e84a9c2b12c004054a7bad9a97ec0c7c89d4681d2",
			err:  true,
		},
		{
			name: "invalid hash",
			str:  "abcdef hoge",
			err:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if s, err := parseSubjects(tc.str); err != nil {
				if tc.err {
					// Error was expected.
					return
				}
				t.Fatalf("unexpected error: %v", err)
			} else {
				if want, got := tc.expected, s; !reflect.DeepEqual(want, got) {
					t.Errorf("unexpected subjects, want: %#v, got: %#v", want, got)
				}
			}
		})
	}
}

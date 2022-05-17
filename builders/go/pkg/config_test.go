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
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func Test_ConfigFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected error
	}{
		{
			name:     "valid releaser empty main",
			path:     "./testdata/releaser-valid-empty-main.yml",
			expected: nil,
		},
		{
			name:     "valid releaser no main",
			path:     "./testdata/releaser-valid-no-main.yml",
			expected: nil,
		},
		{
			name:     "valid releaser main",
			path:     "./testdata/releaser-valid-main.yml",
			expected: nil,
		},
		{
			name:     "valid releaser main with parent path",
			path:     "../pkg/testdata/releaser-valid-main.yml",
			expected: nil,
		},
		{
			name:     "missing version",
			path:     "./testdata/releaser-noversion.yml",
			expected: ErrorUnsupportedVersion,
		},
		{
			name:     "invalid version",
			path:     "./testdata/releaser-invalid-version.yml",
			expected: ErrorUnsupportedVersion,
		},
		{
			name:     "invalid envs",
			path:     "./testdata/releaser-invalid-envs.yml",
			expected: ErrorInvalidEnvironmentVariable,
		},
		{
			name:     "invalid main",
			path:     "./testdata/releaser-invalid-main.yml",
			expected: ErrorInvalidDirectory,
		},
		{
			name:     "invalid path",
			path:     "../testdata/releaser-invalid-main.yml",
			expected: ErrorInvalidDirectory,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ConfigFromFile(tt.path)
			if !errCmp(err, tt.expected) {
				t.Errorf(cmp.Diff(err, tt.expected))
			}
		})
	}
}

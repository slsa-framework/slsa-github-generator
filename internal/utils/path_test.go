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

package utils

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

func Test_PathIsUnderCurrentDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected error
	}{
		{
			name:     "valid same path",
			path:     "./",
			expected: nil,
		},
		{
			name:     "valid path no slash",
			path:     "./some/valid/path",
			expected: nil,
		},
		{
			name:     "valid path with slash",
			path:     "./some/valid/path/",
			expected: nil,
		},
		{
			name:     "valid path with no dot",
			path:     "some/valid/path/",
			expected: nil,
		},
		{
			name:     "some valid path",
			path:     "../utils/some/valid/path",
			expected: nil,
		},
		{
			name:     "parent invalid path",
			path:     "../invalid/path",
			expected: &ErrInvalidPath{},
		},
		{
			name:     "some invalid fullpath",
			path:     "/some/invalid/fullpath",
			expected: &ErrInvalidPath{},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := PathIsUnderCurrentDirectory(tt.path)
			if (err == nil && tt.expected != nil) ||
				(err != nil && tt.expected == nil) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}

			if err != nil && !errors.As(err, &tt.expected) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}
		})
	}
}

func Test_VerifyAttestationPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected error
	}{
		{
			name:     "valid file",
			path:     "./path/to/valid.intoto.jsonl",
			expected: nil,
		},
		{
			name:     "invalid path",
			path:     "../some/invalid/valid.intoto.jsonl",
			expected: &ErrInvalidPath{},
		},
		{
			name:     "invalid extension",
			path:     "some/file.ntoto.jsonl",
			expected: &ErrInvalidPath{},
		},
		{
			name:     "invalid not exntension",
			path:     "some/file.intoto.jsonl.",
			expected: &ErrInvalidPath{},
		},
		{
			name:     "invalid folder exntension",
			path:     "file.intoto.jsonl/file",
			expected: &ErrInvalidPath{},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := VerifyAttestationPath(tt.path)
			if (err == nil && tt.expected != nil) ||
				(err != nil && tt.expected == nil) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}

			if err != nil && !errors.As(err, &tt.expected) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}
		})
	}
}

func tempWD() (func() error, error) {
	// Set up a temporary working directory for the test.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tempwd, err := os.MkdirTemp("", "slsa-github-generator-tests")
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(tempwd); err != nil {
		return nil, err
	}
	return func() error {
		err := os.RemoveAll(tempwd)
		if err != nil {
			return err
		}
		err = os.Chdir(cwd)
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func Test_CreateNewFileUnderCurrentDirectory(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		existingPath bool
		expected     error
	}{
		{
			name:     "valid file cannot create",
			path:     "./path/to/validfile",
			expected: &ErrInvalidPath{},
		},
		{
			name:     "invalid path",
			path:     "../some/invalid/file",
			expected: &ErrInvalidPath{},
		},
		{
			name:         "existing file",
			path:         "existing_file",
			existingPath: true,
			expected:     &ErrInvalidPath{},
		},
		{
			name: "new file",
			path: "new_file",
		},
		{
			name:     "new file in sub-directory",
			path:     "dir/new_file",
			expected: &ErrInvalidPath{},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tempWD()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := cleanup()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}()

			if tt.existingPath {
				if _, err := CreateNewFileUnderCurrentDirectory(tt.path, os.O_WRONLY); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			_, err = CreateNewFileUnderCurrentDirectory(tt.path, os.O_WRONLY)
			if (err == nil && tt.expected != nil) ||
				(err != nil && tt.expected == nil) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}

			if err != nil && !errors.As(err, &tt.expected) {
				t.Fatalf("unexpected error: %v", cmp.Diff(err, tt.expected, cmpopts.EquateErrors()))
			}
		})
	}
}

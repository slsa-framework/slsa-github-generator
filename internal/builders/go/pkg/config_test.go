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
	"github.com/google/go-cmp/cmp/cmpopts"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func Test_ConfigFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		err    error
		config GoReleaserConfig
	}{
		{
			name: "valid releaser empty main",
			path: "./testdata/releaser-valid-empty-main.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0",
				},
			},
		},
		{
			name: "valid releaser no main",
			path: "./testdata/releaser-valid-no-main.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0",
				},
			},
		},
		{
			name: "valid releaser main",
			path: "./testdata/releaser-valid-main.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0",
				},
				Main: asPointer("./relative/main.go"),
			},
		},
		{
			name: "valid env var with no value",
			path: "./testdata/releaser-valid-envs-no-value.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0", "CGO_CFLAGS": "",
				},
			},
		},
		{
			name: "valid env var with multiple = signs",
			path: "./testdata/releaser-valid-envs-multiple-equal-signs.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0", "CGO_CFLAGS": "a=b=c",
				},
			},
		},
		{
			name: "missing version",
			path: "./testdata/releaser-noversion.yml",
			err:  &ErrUnsupportedVersion{},
		},
		{
			name: "invalid version",
			path: "./testdata/releaser-invalid-version.yml",
			err:  &ErrUnsupportedVersion{},
		},
		{
			name: "invalid envs",
			path: "./testdata/releaser-invalid-envs.yml",
			err:  &ErrInvalidEnvironmentVariable{},
		},
		{
			name: "invalid main",
			path: "./testdata/releaser-invalid-main.yml",
			err:  &ErrInvalidDirectory{},
		},
		{
			name: "invalid main path",
			path: "../testdata/releaser-invalid-main.yml",
			err:  &ErrInvalidDirectory{},
		},
		{
			name: "invalid dir path",
			path: "../testdata/releaser-invalid-dir.yml",
			err:  &ErrInvalidDirectory{},
		},
		{
			name: "valid dir path",
			path: "./testdata/releaser-valid-dir.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0",
				},
				Dir: asPointer("./path/to/dir"),
			},
		},
		{
			name: "valid config path with dots",
			// Resolves to "./testdata/releaser-valid-dir.yml".
			path: "./testdata/../testdata/./releaser-valid-dir.yml",
			config: GoReleaserConfig{
				Goos: "linux", Goarch: "amd64",
				Flags:   []string{"-trimpath", "-tags=netgo"},
				Ldflags: []string{"{{ .Env.VERSION_LDFLAGS }}"},
				Binary:  "binary-{{ .OS }}-{{ .Arch }}",
				Env: map[string]string{
					"GO111MODULE": "on", "CGO_ENABLED": "0",
				},
				Dir: asPointer("./path/to/dir"),
			},
		},
		{
			name: "invalid config path with dots",
			// Resolves to "../releaser-valid-dir.yml".
			path: "./testdata/../testdata/./foo/../../../releaser-valid-dir.yml",
			err:  &ErrInvalidDirectory{},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := ConfigFromFile(tt.path)

			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}
			if err != nil {
				return
			}

			if !cmp.Equal(*cfg, tt.config) {
				t.Errorf(cmp.Diff(*cfg, tt.config))
			}
		})
	}
}

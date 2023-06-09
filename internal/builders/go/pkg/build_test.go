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
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func errEnvVariableNameEmptyFunc(t *testing.T, got error) {
	want := errEnvVariableNameEmpty
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}

func errUnsupportedArgumentsFunc(t *testing.T, got error) {
	want := errUnsupportedArguments
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}

func errInvalidEnvArgumentFunc(t *testing.T, got error) {
	want := errInvalidEnvArgument
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}

func errEnvVariableNameNotAllowedFunc(t *testing.T, got error) {
	want := errEnvVariableNameNotAllowed
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}

func errInvalidFilenameFunc(t *testing.T, got error) {
	want := errInvalidFilename
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: %v", cmp.Diff(got, want, cmpopts.EquateErrors()))
	}
}

func Test_isAllowedEnvVariable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{
			name:     "BLA variable",
			variable: "BLA",
			expected: false,
		},
		{
			name:     "random variable",
			variable: "random",
			expected: false,
		},
		{
			name:     "GOSOMETHING variable",
			variable: "GOSOMETHING",
			expected: true,
		},
		{
			name:     "CGO_SOMETHING variable",
			variable: "CGO_SOMETHING",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := isAllowedEnvVariable(tt.variable)
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_getOutputBinaryPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  func(*testing.T, error)
		name string
		path string
	}{
		{
			name: "empty output",
			path: "",
			err:  errInvalidFilenameFunc,
		},
		{
			name: "not absolute",
			path: "./some/path/to/binary",
			err:  errInvalidFilenameFunc,
		},
		{
			name: "absolute path",
			path: "/to/absolute/path/to/binary",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := getOutputBinaryPath(tt.path)
			if tt.err != nil {
				tt.err(t, err)
			}

			if err != nil {
				return
			}

			if !cmp.Equal(r, tt.path) {
				t.Errorf(cmp.Diff(r, tt.path))
			}
		})
	}
}

func Test_isAllowedArg(t *testing.T) {
	t.Parallel()

	type test struct {
		name     string
		argument string
		expected bool
	}

	var tests []test

	for k := range allowedBuildArgs {
		tests = append(tests, test{
			name:     fmt.Sprintf("%s argument", k),
			argument: k,
			expected: true,
		}, test{
			name:     fmt.Sprintf("%sbla argument", k),
			argument: fmt.Sprintf("%sbla", k),
			expected: true,
		}, test{
			name:     fmt.Sprintf("bla %s argument", k),
			argument: fmt.Sprintf("bla%s", k),
			expected: false,
		}, test{
			name:     fmt.Sprintf("space %s argument", k),
			argument: fmt.Sprintf(" %s", k),
			expected: false,
		})
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := isAllowedArg(tt.argument)
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_generateOutputFilename(t *testing.T) {
	// Disable to avoid env clobbering between tests.
	// t.Parallel()

	tests := []struct {
		name     string
		filename string
		goos     string
		goarch   string
		envs     map[string]string
		argEnv   string
		expected struct {
			err func(*testing.T, error)
			fn  string
		}
	}{
		{
			name:     "invalid filename",
			filename: "../filename",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			name:     "valid filename",
			filename: "",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			name:     "filename arch",
			filename: "name-{{ .Arch }}",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errEnvVariableNameEmptyFunc,
			},
		},
		{
			name:     "filename os",
			filename: "name-{{ .Os }}",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errEnvVariableNameEmptyFunc,
			},
		},
		{
			name:     "filename invalid letter ^",
			filename: "Name-AB^",
			goarch:   "amd64",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			filename: "filename invalid letter $",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			name:     "filename os",
			filename: "name-{{ .Os }}",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errEnvVariableNameEmptyFunc,
			},
		},
		{
			name:     "filename linux os",
			filename: "name-{{ .Os }}",
			goos:     "linux",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-linux",
			},
		},
		{
			name:     "filename amd64 arch",
			filename: "name-{{ .Arch }}",
			goarch:   "amd64",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-amd64",
			},
		},
		{
			name:     "filename capital letter",
			filename: "Name-{{ .Arch }}",
			goarch:   "amd64",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "Name-amd64",
			},
		},
		{
			name:     "filename amd64/linux arch",
			filename: "name-{{ .Os }}-{{ .Arch }}",
			goarch:   "amd64",
			goos:     "linux",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-linux-amd64",
			},
		},
		{
			name:     "filename invalid arch",
			filename: "name-{{ .Arch }}",
			goarch:   "something/../../",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			name:     "filename invalid not supported",
			filename: "name-{{ .Bla }}",
			goarch:   "something/../../",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidEnvArgumentFunc,
			},
		},
		{
			name:     "filename amd64/linux v1.2.3",
			filename: "name-{{ .Os }}-{{ .Arch }}-{{ .Tag }}",
			goarch:   "amd64",
			goos:     "linux",
			envs: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-linux-amd64-v1.2.3",
			},
		},
		{
			name:     "filename twice v1.2.3",
			filename: "name-{{ .Tag }}-{{ .Tag }}",
			goarch:   "amd64",
			goos:     "linux",
			envs: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-v1.2.3-v1.2.3",
			},
		},
		{
			name:     "filename twice empty versions",
			filename: "name-{{ .Tag }}-{{ .Tag }}",
			goarch:   "amd64",
			goos:     "linux",
			envs: map[string]string{
				"GITHUB_REF_NAME": "",
			},
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  fmt.Sprintf("name-%s-%s", unknownTag, unknownTag),
			},
		},
		{
			name:     "invalid name with version",
			filename: "name-{{ .Tag }}/../bla",
			goarch:   "amd64",
			goos:     "linux",
			envs: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: errInvalidFilenameFunc,
			},
		},
		{
			name:     "filename twice unset versions",
			filename: "name-{{ .Tag }}-{{ .Tag }}",
			goarch:   "amd64",
			goos:     "linux",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  fmt.Sprintf("name-%s-%s", unknownTag, unknownTag),
			},
		},
		{
			name:     "filename envs",
			filename: "name-{{ .Env.VAR1 }}-{{ .Os }}-{{ .Arch }}-{{ .Env.VAR2 }}-{{ .Tag }}-end",
			goarch:   "amd64",
			goos:     "linux",
			envs: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			argEnv: "VAR1:var1, VAR2:var2",
			expected: struct {
				err func(*testing.T, error)
				fn  string
			}{
				err: nil,
				fn:  "name-var1-linux-amd64-var2-v1.2.3-end",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			// Note: disable parallelism to avoid env variable clobbering between tests.
			// t.Parallel()

			cfg := goReleaserConfigFile{
				Binary:  tt.filename,
				Version: 1,
				Goos:    tt.goos,
				Goarch:  tt.goarch,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}

			// Unset env variables, in case the workflow environment sets them.
			for _, k := range []string{"GITHUB_REF_NAME"} {
				os.Unsetenv(k)
			}

			// Set env variables.
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			b := GoBuildNew("go compiler", c)

			err = b.SetArgEnvVariables(tt.argEnv)
			if err != nil {
				t.Errorf("SetArgEnvVariables: %v", err)
			}

			fn, err := b.generateOutputFilename()
			if tt.expected.err != nil {
				tt.expected.err(t, err)
			}

			// Unset env variables, so that they don't
			// affect other tests.
			for k := range tt.envs {
				os.Unsetenv(k)
			}

			if err != nil {
				return
			}

			if fn != tt.expected.fn {
				t.Errorf(cmp.Diff(fn, tt.expected.fn))
			}
		})
	}
}

func Test_SetArgEnvVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expected struct {
			env map[string]string
			err func(*testing.T, error)
		}
		name   string
		argEnv string
	}{
		{
			name:   "valid arg envs",
			argEnv: "VAR1:value1, VAR2:value2",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				env: map[string]string{"VAR1": "value1", "VAR2": "value2"},
				err: nil,
			},
		},
		{
			name:   "empty arg envs",
			argEnv: "",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				env: map[string]string{},
				err: nil,
			},
		},
		{
			name:   "valid arg envs not space",
			argEnv: "VAR1:value1,VAR2:value2",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				env: map[string]string{"VAR1": "value1", "VAR2": "value2"},
				err: nil,
			},
		},
		{
			name:   "invalid arg empty 2 values",
			argEnv: "VAR1:value1,",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				err: errInvalidEnvArgumentFunc,
			},
		},
		{
			name:   "invalid arg empty 3 values",
			argEnv: "VAR1:value1,, VAR3:value3",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				err: errInvalidEnvArgumentFunc,
			},
		},
		{
			name:   "invalid arg uses equal",
			argEnv: "VAR1=value1",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				err: errInvalidEnvArgumentFunc,
			},
		},
		{
			name:   "valid single arg",
			argEnv: "VAR1:value1",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				env: map[string]string{"VAR1": "value1"},
				err: nil,
			},
		},
		{
			name:   "invalid valid single arg with empty",
			argEnv: "VAR1:value1:",
			expected: struct {
				env map[string]string
				err func(*testing.T, error)
			}{
				err: errInvalidEnvArgumentFunc,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}
			b := GoBuildNew("go compiler", c)

			err = b.SetArgEnvVariables(tt.argEnv)
			if tt.expected.err != nil {
				tt.expected.err(t, err)
			}

			if err != nil {
				return
			}

			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(b.argEnv, tt.expected.env, sorted) {
				t.Errorf(cmp.Diff(b.argEnv, tt.expected.env))
			}
		})
	}
}

func Test_generateEnvVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		goos     string
		goarch   string
		env      []string
		expected struct {
			err   func(*testing.T, error)
			flags []string
		}
	}{
		{
			name:   "empty flags",
			goos:   "linux",
			goarch: "x86",
			expected: struct {
				err   func(*testing.T, error)
				flags []string
			}{
				flags: []string{"GOOS=linux", "GOARCH=x86"},
				err:   nil,
			},
		},
		{
			name:   "empty goos",
			goarch: "x86",
			expected: struct {
				err   func(*testing.T, error)
				flags []string
			}{
				flags: []string{},
				err:   errEnvVariableNameEmptyFunc,
			},
		},
		{
			name: "empty goarch",
			goos: "windows",
			expected: struct {
				err   func(*testing.T, error)
				flags []string
			}{
				flags: []string{},
				err:   errEnvVariableNameEmptyFunc,
			},
		},
		{
			name:   "invalid flags",
			goos:   "windows",
			goarch: "amd64",
			env:    []string{"VAR1=value1", "VAR2=value2"},
			expected: struct {
				err   func(*testing.T, error)
				flags []string
			}{
				err: errEnvVariableNameNotAllowedFunc,
			},
		},
		{
			name:   "invalid flags",
			goos:   "windows",
			goarch: "amd64",
			env:    []string{"GOVAR1=value1", "GOVAR2=value2", "CGO_VAR1=val1", "CGO_VAR2=val2"},
			expected: struct {
				err   func(*testing.T, error)
				flags []string
			}{
				flags: []string{
					"GOOS=windows", "GOARCH=amd64",
					"GOVAR1=value1", "GOVAR2=value2",
					"CGO_VAR1=val1", "CGO_VAR2=val2",
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
				Goos:    tt.goos,
				Goarch:  tt.goarch,
				Env:     tt.env,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}
			b := GoBuildNew("go compiler", c)

			flags, err := b.generateCommandEnvVariables()

			if tt.expected.err != nil {
				tt.expected.err(t, err)
			}
			if err != nil {
				return
			}
			expectedFlags := tt.expected.flags
			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(flags, expectedFlags, sorted) {
				t.Errorf(cmp.Diff(flags, expectedFlags))
			}
		})
	}
}

func Test_generateLdflags(t *testing.T) {
	// Disable to avoid env clobbering between tests.
	// t.Parallel()

	tests := []struct {
		githubEnv  map[string]string
		name       string
		argEnv     string
		outldflags string
		err        func(*testing.T, error)
		inldflags  []string
	}{
		{
			name:       "version ldflags",
			argEnv:     "VERSION_LDFLAGS:value1",
			inldflags:  []string{"{{ .Env.VERSION_LDFLAGS }}"},
			outldflags: "value1",
		},
		{
			name:       "one value with text",
			argEnv:     "VAR1:value1, VAR2:value2",
			inldflags:  []string{"name-{{ .Env.VAR1 }}"},
			outldflags: "name-value1",
		},
		{
			name:       "two values with text",
			argEnv:     "VAR1:value1, VAR2:value2",
			inldflags:  []string{"name-{{ .Env.VAR1 }}-{{ .Env.VAR2 }}"},
			outldflags: "name-value1-value2",
		},
		{
			name:       "two values with text and not space between env",
			argEnv:     "VAR1:value1,VAR2:value2",
			inldflags:  []string{"name-{{ .Env.VAR1 }}-{{ .Env.VAR2 }}"},
			outldflags: "name-value1-value2",
		},
		{
			name:       "same two values with text",
			argEnv:     "VAR1:value1, VAR2:value2",
			inldflags:  []string{"name-{{ .Env.VAR1 }}-{{ .Env.VAR1 }}"},
			outldflags: "name-value1-value1",
		},
		{
			name:       "same value extremeties",
			argEnv:     "VAR1:value1, VAR2:value2",
			inldflags:  []string{"{{ .Env.VAR1 }}-name-{{ .Env.VAR1 }}"},
			outldflags: "value1-name-value1",
		},
		{
			name:       "two different value extremeties",
			argEnv:     "VAR1:value1, VAR2:value2",
			inldflags:  []string{"{{ .Env.VAR1 }}-name-{{ .Env.VAR2 }}"},
			outldflags: "value1-name-value2",
		},
		{
			name:      "undefined env variable",
			argEnv:    "VAR2:value2",
			inldflags: []string{"{{ .Env.VAR1 }}-name-{{ .Env.VAR1 }}"},
			err:       errEnvVariableNameEmptyFunc,
		},
		{
			name:      "undefined env variable 1",
			argEnv:    "VAR2:value2",
			inldflags: []string{"{{ .Env.VAR2 }}-name-{{ .Env.VAR1 }}"},
			err:       errEnvVariableNameEmptyFunc,
		},
		{
			name:      "empty env variable",
			argEnv:    "",
			inldflags: []string{"{{ .Env.VAR1 }}-name-{{ .Env.VAR1 }}"},
			err:       errEnvVariableNameEmptyFunc,
		},
		{
			name:   "several ldflags",
			argEnv: "VAR1:value1, VAR2:value2, VAR3:value3",
			inldflags: []string{
				"{{ .Env.VAR1 }}-name-{{ .Env.VAR2 }}",
				"{{ .Env.VAR1 }}-name-{{ .Env.VAR3 }}",
				"{{ .Env.VAR3 }}-name-{{ .Env.VAR1 }}",
				"{{ .Env.VAR3 }}-name-{{ .Env.VAR2 }}",
			},
			outldflags: "value1-name-value2 value1-name-value3 value3-name-value1 value3-name-value2",
		},
		{
			name:   "several ldflags with start/end",
			argEnv: "VAR1:value1, VAR2:value2, VAR3:value3",
			inldflags: []string{
				"start-{{ .Env.VAR1 }}-name-{{ .Env.VAR2 }}-end",
				"start-{{ .Env.VAR1 }}-name-{{ .Env.VAR3 }}-end",
				"start-{{ .Env.VAR3 }}-name-{{ .Env.VAR1 }}-end",
				"start-{{ .Env.VAR3 }}-name-{{ .Env.VAR2 }}-end",
			},
			outldflags: "start-value1-name-value2-end start-value1-name-value3-end " +
				"start-value3-name-value1-end start-value3-name-value2-end",
		},
		{
			name:   "several ldflags and tag",
			argEnv: "VAR1:value1, VAR2:value2, VAR3:value3",
			githubEnv: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			inldflags: []string{
				"start-{{ .Env.VAR1 }}-name-{{ .Env.VAR2 }}-{{ .Tag }}-end",
				"{{ .Env.VAR1 }}-name-{{ .Env.VAR3 }}",
				"{{ .Env.VAR3 }}-name-{{ .Env.VAR1 }}-{{ .Tag }}-{{ .Tag }}",
				"{{ .Env.VAR3 }}-name-{{ .Env.VAR2 }}-{{ .Tag }}-end",
			},
			outldflags: "start-value1-name-value2-v1.2.3-end value1-name-value3 " +
				"value3-name-value1-v1.2.3-v1.2.3 value3-name-value2-v1.2.3-end",
		},
		{
			name:   "several ldflags and Arch and Os",
			argEnv: "VAR1:value1, VAR2:value2, VAR3:value3",
			githubEnv: map[string]string{
				"GITHUB_REF_NAME": "v1.2.3",
			},
			inldflags: []string{
				"start-{{ .Env.VAR1 }}-name-{{ .Env.VAR2 }}-{{ .Tag }}-{{ .Arch }}-end",
				"{{ .Env.VAR1 }}-name-{{ .Env.VAR3 }}-{{ .Os }}-end",
			},
			outldflags: "start-value1-name-value2-v1.2.3-amd64-end value1-name-value3-linux-end",
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			// Disable to avoid env clobbering between tests.
			// t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
				Ldflags: tt.inldflags,
				Goarch:  "amd64",
				Goos:    "linux",
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}

			// Set GitHub env variables.
			for k, v := range tt.githubEnv {
				t.Setenv(k, v)
			}

			b := GoBuildNew("go compiler", c)

			err = b.SetArgEnvVariables(tt.argEnv)
			if err != nil {
				t.Errorf("SetArgEnvVariables: %v", err)
			}
			ldflags, err := b.generateLdflags()

			// Unset env variables, so that they don't
			// affect other tests.
			for k := range tt.githubEnv {
				os.Unsetenv(k)
			}

			if tt.err != nil {
				tt.err(t, err)
			}
			if err != nil {
				return
			}
			// Note: generated env variables contain the process's env variables too.
			if !cmp.Equal(ldflags, tt.outldflags) {
				t.Errorf(cmp.Diff(ldflags, tt.outldflags))
			}
		})
	}
}

func Test_generateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   func(*testing.T, error)
		flags []string
	}{
		{
			name:  "valid flags",
			flags: []string{"-race", "-x"},
			err:   nil,
		},
		{
			name:  "invalid -mod flags",
			flags: []string{"-mod=whatever", "-x"},
			err:   errUnsupportedArgumentsFunc,
		},
		{
			name: "invalid random flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath", "bla",
			},
			err: errUnsupportedArgumentsFunc,
		},
		{
			name: "valid all flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
				Flags:   tt.flags,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}
			b := GoBuildNew("gocompiler", c)

			flags, err := b.generateFlags()
			expectedFlags := append([]string{"gocompiler", "build", "-mod=vendor"}, tt.flags...)

			if tt.err != nil {
				tt.err(t, err)
			}
			if err != nil {
				return
			}
			// Note: generated env variables contain the process's env variables too.
			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(flags, expectedFlags, sorted) {
				t.Errorf(cmp.Diff(flags, expectedFlags))
			}
		})
	}
}

func Test_generateCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		flags []string
	}{
		{
			name:  "some flags",
			flags: []string{"-race", "-x"},
		},
		{
			name:  "some other flags",
			flags: []string{"-x"},
		},
		{
			name: "other random flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath", "bla",
			},
		},
		{
			name: "even more flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfgs := []*goReleaserConfigFile{
				{
					Version: 1,
					Flags:   tt.flags,
					Main:    asPointer("./some/path/main.go"),
				},
				{
					Version: 1,
					Flags:   tt.flags,
				},
			}

			for _, cfg := range cfgs {
				c, err := fromConfig(cfg)
				if err != nil {
					t.Errorf("fromConfig: %v", err)
				}
				b := GoBuildNew("gocompiler", c)

				command := b.generateCommand(tt.flags, "out-binary")
				expectedCommand := tt.flags
				expectedCommand = append(expectedCommand, "-o", "out-binary")
				if cfg.Main != nil {
					expectedCommand = append(expectedCommand, *cfg.Main)
				}

				sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
				if !cmp.Equal(command, expectedCommand, sorted) {
					t.Errorf(cmp.Diff(command, expectedCommand))
				}
			}
		})
	}
}

func asPointer(s string) *string {
	return &s
}

func TestGoBuild_Run(t *testing.T) {
	type fields struct {
		cfg    *GoReleaserConfig
		argEnv map[string]string
		goc    string
	}
	type args struct {
		dry bool
	}
	tests := []struct {
		name    string
		err     func(*testing.T, error)
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "dry run valid flags",
			fields: fields{
				cfg: &GoReleaserConfig{
					Goos:   "linux",
					Goarch: "amd64",
					Binary: "binary",
					Main:   asPointer("../builders/go/main.go"),
					Dir:    asPointer("../builders/go"),
					Ldflags: []string{
						"-X main.version=1.0.0",
					},
				},
			},
			args: args{
				dry: true,
			},
		},
		{
			name: "non-dry valid flags",
			fields: fields{
				cfg: &GoReleaserConfig{
					Goos:   "linux",
					Goarch: "amd64",
					Binary: "/tmp/binary",
					Main:   asPointer("main.go"),
					Dir:    asPointer("./testdata/go"),
					Ldflags: []string{
						"-X main.version=1.0.0",
					},
				},
			},
			args: args{
				dry: false,
			},
		},
		{
			name: "slash in the binary name",
			fields: fields{
				cfg: &GoReleaserConfig{
					Goos:   "linux",
					Goarch: "amd64",
					Binary: "tmp/binary",
					Main:   asPointer("../builders/go/main.go"),
					Dir:    asPointer("../builders/go"),
				},
			},
			args: args{
				dry: true,
			},
			wantErr: true,
			err:     errInvalidFilenameFunc,
		},
		{
			name: "dry run - invalid flags",
			fields: fields{
				cfg: &GoReleaserConfig{
					Goos:    "linux",
					Goarch:  "amd64",
					Binary:  "binary",
					Main:    asPointer("../builders/go/main.go"),
					Dir:     asPointer("../builders/go"),
					Ldflags: []string{},
				},
			},
			args: args{
				dry: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &GoBuild{
				cfg:    tt.fields.cfg,
				goc:    tt.fields.goc,
				argEnv: tt.fields.argEnv,
			}
			t.Setenv("OUTPUT_BINARY", tt.fields.cfg.Binary)
			// if the test is not dry run , then code has to look for golang binary
			if !tt.args.dry {
				path, err := exec.LookPath("go")
				if err != nil {
					t.Errorf("exec.LookPath: %v", err)
				}
				b.goc = path
			}
			err := b.Run(tt.args.dry)
			if err != nil != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.err != nil {
				if err == nil {
					t.Errorf("Run() error = nil, but wanted error")
				} else if tt.err != nil {
					tt.err(t, err)
				}
			}
		})
	}
}

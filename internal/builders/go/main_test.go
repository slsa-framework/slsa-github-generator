package main

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/slsa-framework/slsa-github-generator/internal/builders/go/pkg"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

func errCmp(e1, e2 error) bool {
	return errors.Is(e1, e2) || errors.Is(e2, e1)
}

func checkWorkingDir(t *testing.T, wd, expected string) {
	var expectedWd string
	var err error
	if expected != "" {
		expectedWd, err = filepath.Abs(expected)
		if err != nil {
			t.Errorf("Abs: %v", err)
		}
	} else {
		expectedWd, err = os.Getwd()
		if err != nil {
			t.Errorf("Getwd: %v", err)
		}
	}

	if expectedWd != wd {
		t.Errorf(cmp.Diff(wd, expectedWd))
	}
}

func Test_runBuild(t *testing.T) {
	t.Parallel()

	tests := []struct {
		subject    string
		name       string
		config     string
		evalEnvs   string
		workingDir string
		err        error
		commands   []string
		envs       []string
	}{
		{
			name:     "two ldflags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "two ldflags empty env",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags-emptyenv.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
			},
		},
		{
			name:     "two ldflags no env",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags-noenv.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
			},
		},
		{
			name:     "two ldflags empty flags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags-emptyflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "two ldflags no flags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags-noflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "one ldflags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/one-ldflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=something-else",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "no ldflags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/two-ldflags-noldflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "empty ldflags",
			subject:  "binary-linux-amd64",
			config:   "./testdata/emptyldflags.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-o",
				"binary-linux-amd64",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "valid main",
			subject:  "binary-linux-amd64",
			config:   "./testdata/valid-main.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
				"./path/to/main.go",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
		},
		{
			name:     "valid working dir",
			subject:  "binary-linux-amd64",
			config:   "./testdata/valid-working-dir.yml",
			evalEnvs: "VERSION_LDFLAGS:bla, ELSE:else",
			commands: []string{
				"-trimpath",
				"-tags=netgo",
				"-ldflags=bla something-else",
				"-o",
				"binary-linux-amd64",
				"main.go",
			},
			envs: []string{
				"GOOS=linux",
				"GOARCH=amd64",
				"GO111MODULE=on",
				"CGO_ENABLED=0",
			},
			workingDir: "./valid/path/",
		},
		// Below are the same tests we do in pkg/config_test.go
		{
			name:   "invalid main",
			config: "./pkg/testdata/releaser-invalid-main.yml",
			err:    pkg.ErrorInvalidDirectory,
		},
		{
			name:   "missing version",
			config: "./pkg/testdata/releaser-noversion.yml",
			err:    pkg.ErrorUnsupportedVersion,
		},
		{
			name:   "invalid version",
			config: "./pkg/testdata/releaser-invalid-version.yml",
			err:    pkg.ErrorUnsupportedVersion,
		},
		{
			name:   "invalid envs",
			config: "./pkg/testdata/releaser-invalid-envs.yml",
			err:    pkg.ErrorInvalidEnvironmentVariable,
		},
		{
			name:   "invalid path",
			config: "../pkg/testdata/releaser-invalid-main.yml",
			err:    pkg.ErrorInvalidDirectory,
		},
		{
			name:   "invalid dir path",
			config: "../pkg/testdata/releaser-invalid-dir.yml",
			err:    pkg.ErrorInvalidDirectory,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			// *** WARNING: do not enable t.Parallel(), because we're writing to  ***.

			file, err := os.CreateTemp("", "")
			if err != nil {
				t.Fatalf("unable to create a temp env file: %s", err)
			}
			defer os.Remove(file.Name())
			// http://craigwickesser.com/2015/01/capture-stdout-in-go/

			t.Setenv("GITHUB_OUTPUT", file.Name())

			err = runBuild(true,
				tt.config,
				tt.evalEnvs)

			if !errCmp(err, tt.err) {
				t.Errorf(cmp.Diff(err, tt.err, cmpopts.EquateErrors()))
			}

			if err != nil {
				return
			}

			file.Seek(0, 0)
			cmd, env, subject, wd, err := extract(file)
			if err != nil {
				t.Errorf("extract: %v", err)
			}

			goc, err := exec.LookPath("go")
			if err != nil {
				t.Errorf("exec.LookPath: %v", err)
			}

			if !cmp.Equal(subject, tt.subject) {
				t.Errorf(cmp.Diff(subject, tt.subject))
			}

			commands := append([]string{goc, "build", "-mod=vendor"}, tt.commands...)
			if !cmp.Equal(cmd, commands) {
				t.Errorf(cmp.Diff(cmd, commands))
			}

			checkWorkingDir(t, wd, tt.workingDir)

			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(env, tt.envs, sorted) {
				t.Errorf(cmp.Diff(env, tt.envs))
			}
		})
	}
}

func extract(file *os.File) ([]string, []string, string, string, error) {
	rsubject := regexp.MustCompile(`^go-binary-name=(.*)$`)
	rcmd := regexp.MustCompile(`^go-command=(.*)$`)
	renv := regexp.MustCompile(`^go-env=(.*)$`)
	rwd := regexp.MustCompile(`^go-working-dir=(.*)$`)
	var subject string
	var scmd string
	var senv string
	var wd string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		n := rsubject.FindStringSubmatch(scanner.Text())
		if len(n) > 1 {
			subject = n[1]
		}

		c := rcmd.FindStringSubmatch(scanner.Text())
		if len(c) > 1 {
			scmd = c[1]
		}

		e := renv.FindStringSubmatch(scanner.Text())
		if len(e) > 1 {
			senv = e[1]
		}

		w := rwd.FindStringSubmatch(scanner.Text())
		if len(w) > 1 {
			wd = w[1]
		}

		if subject != "" && scmd != "" && senv != "" && wd != "" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return []string{}, []string{}, "", "", err
	}

	cmd, err := utils.UnmarshalList(scmd)
	if err != nil {
		return []string{}, []string{}, "", "", err
	}

	env, err := utils.UnmarshalList(senv)
	if err != nil {
		return []string{}, []string{}, "", "", err
	}

	return cmd, env, subject, wd, nil
}

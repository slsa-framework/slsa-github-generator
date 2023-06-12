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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

func checkTest(t *testing.T) func(err error) {
	return func(err error) {
		if err != nil {
			t.Fatalf("%v", err)
		}
	}
}

func Test_generateCmd_default_predicate(t *testing.T) {
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
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	c := generateCmd(&slsa.NilClientProvider{}, checkTest(t))
	c.SetOut(new(bytes.Buffer))
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "predicate.json")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_generateCmd_custom_predicate(t *testing.T) {
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
	defer func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Errorf("unexpected failure: %v", err)
		}
	}()

	c := generateCmd(&slsa.NilClientProvider{}, checkTest(t))
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{"--predicate", "custom.json"})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "custom.json")); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_generateCmd_invalid_path(t *testing.T) {
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
				t.Fatalf("unexpected error, got: %v, want %v", got, want)
			}
			// Check should exit the program so we skip the rest of the test if we got the expected error.
			t.SkipNow()
		}
	}

	c := generateCmd(&slsa.NilClientProvider{}, check)
	c.SetOut(new(bytes.Buffer))
	c.SetArgs([]string{"--predicate", "/custom.json"})
	if err := c.Execute(); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, "custom.json")); err != nil {
		t.Errorf("error checking file: %v", err)
	}

	// If no error occurs we catch it here. SkipNow will exit the test process so this code should be unreachable.
	t.Errorf("expected an error to occur.")
}

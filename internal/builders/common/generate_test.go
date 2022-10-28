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

package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/errors"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
	"github.com/slsa-framework/slsa-github-generator/slsa"
)

func Test_Generate_default_predicate(t *testing.T) {
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

	name := "predicate.json"
	build := GenericBuild{
		GithubActionsBuild: slsa.NewGithubActionsBuild(nil, github.WorkflowContext{}),
	}
	if err := Generate(&slsa.NilClientProvider{}, &build, name); err != nil {
		t.Errorf("unexpected failure: %v", err)
	}

	// check that the expected file exists.
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Errorf("error checking file: %v", err)
	}
}

func Test_Generate_invalid_path(t *testing.T) {
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

	build := GenericBuild{
		GithubActionsBuild: slsa.NewGithubActionsBuild(nil, github.WorkflowContext{}),
	}
	err = Generate(&slsa.NilClientProvider{}, &build, "/invalid.json")
	errInvalidPath := &utils.ErrInvalidPath{}
	if !errors.As(err, &errInvalidPath) {
		t.Fatalf("expected %v but got %v", &utils.ErrInvalidPath{}, err)
	}
}

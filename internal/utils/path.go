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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
)

// ErrInternal indicates an internal error.
type ErrInternal struct {
	errors.WrappableError
}

// ErrInvalidPath indicates an invalid path.
type ErrInvalidPath struct {
	errors.WrappableError
}

func PathIsUnderCurrentDirectory(path string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Errorf(&ErrInternal{}, "os.Getwd(): %w", err)
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return errors.Errorf(&ErrInternal{}, "filepath.Abs(): %w", err)
	}

	if !strings.HasPrefix(p, wd+"/") &&
		wd != p {
		return errors.Errorf(&ErrInvalidPath{}, "invalid path: %q", path)
	}

	return nil
}

func VerifyAttestationPath(path string) error {
	if !strings.HasSuffix(path, "intoto.jsonl") {
		return errors.Errorf(&ErrInvalidPath{}, "invalid suffix: %q. Must be .intoto.jsonl", path)
	}
	if err := PathIsUnderCurrentDirectory(path); err != nil {
		return err
	}
	return nil
}

func CreateNewFileUnderCurrentDirectory(path string, flag int) (io.Writer, error) {
	if path == "-" {
		return os.Stdout, nil
	}

	if err := PathIsUnderCurrentDirectory(path); err != nil {
		return nil, err
	}

	// Ensure we never overwrite an existing file.
	return os.OpenFile(filepath.Clean(path), flag|os.O_CREATE|os.O_EXCL, 0o600)
}

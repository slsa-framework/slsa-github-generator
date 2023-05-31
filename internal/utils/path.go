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

// PathIsUnderCurrentDirectory checks whether the `path`
// is under the current working directory. Examples:
// ./file, ./some/path, ../<cwd>.file would return `nil`.
// `../etc/password` would return an error.
func PathIsUnderCurrentDirectory(path string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Errorf(&ErrInternal{}, "os.Getwd(): %w", err)
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return errors.Errorf(&ErrInternal{}, "filepath.Abs(): %w", err)
	}
	return checkPathUnderDir(p, wd)
}

// PathIsUnderDirectory checks to see if path is under the absolute
// directory specified.
func PathIsUnderDirectory(path, absoluteDir string) error {
	p, err := filepath.Abs(filepath.Join(absoluteDir, path))
	if err != nil {
		return errors.Errorf(&ErrInternal{}, "filepath.Abs(): %w", err)
	}

	return checkPathUnderDir(p, absoluteDir)
}

func checkPathUnderDir(p, dir string) error {
	if !strings.HasPrefix(p, dir+"/") &&
		dir != p {
		return errors.Errorf(&ErrInvalidPath{}, "invalid path: %q", p)
	}
	return nil
}

// VerifyAttestationPath verifies that the path of an attestation
// is valid. It checks that the path is under the current working directory
// and that the extension of the file is `intoto.jsonl`.
func VerifyAttestationPath(path string) error {
	if !strings.HasSuffix(path, "intoto.jsonl") {
		return errors.Errorf(&ErrInvalidPath{}, "invalid suffix: %q. Must be .intoto.jsonl", path)
	}
	return PathIsUnderCurrentDirectory(path)
}

// CreateNewFileUnderCurrentDirectory create a new file under the current directory
// and fails if the file already exists. The file is always created with the pemisisons
// `0o600`.
func CreateNewFileUnderCurrentDirectory(path string, flag int) (io.Writer, error) {
	if path == "-" {
		return os.Stdout, nil
	}

	if err := PathIsUnderCurrentDirectory(path); err != nil {
		return nil, err
	}

	// Ensure we never overwrite an existing file.
	fp, err := os.OpenFile(filepath.Clean(path), flag|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, errors.Errorf(&ErrInternal{}, "os.OpenFile(): %v", err)
	}

	return fp, nil
}

// CreateNewFileUnderDirectory create a new file under the current directory
// and fails if the file already exists. The file is always created with the pemisisons
// `0o600`. Ensures that the path does not exit out of the given directory.
func CreateNewFileUnderDirectory(path, dir string, flag int) (io.Writer, error) {
	if path == "-" {
		return os.Stdout, nil
	}

	if err := PathIsUnderDirectory(path, dir); err != nil {
		return nil, err
	}

	// Create the directory if it does not exist
	fullPath := filepath.Join(dir, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
	if err != nil {
		return nil, errors.Errorf(&ErrInternal{}, "os.MkdirAll(): %v", err)
	}

	// Ensure we never overwrite an existing file.
	fp, err := os.OpenFile(filepath.Clean(fullPath), flag|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, errors.Errorf(&ErrInternal{}, "os.OpenFile(): %v", err)
	}

	return fp, nil
}

// SafeReadFile checks for directory traversal before reading the given file.
func SafeReadFile(path string) ([]byte, error) {
	if err := PathIsUnderCurrentDirectory(path); err != nil {
		return nil, errors.Errorf(&ErrInternal{}, "PathIsUnderCurrentDirectory: %v", err)
	}
	return os.ReadFile(path)
}

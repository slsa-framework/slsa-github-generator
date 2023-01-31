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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/slsa-framework/slsa-github-generator/internal/errors"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

var supportedVersions = map[int]bool{
	1: true,
}

type goReleaserConfigFile struct {
	Main    *string  `yaml:"main"`
	Dir     *string  `yaml:"dir"`
	Goos    string   `yaml:"goos"`
	Goarch  string   `yaml:"goarch"`
	Binary  string   `yaml:"binary"`
	Env     []string `yaml:"env"`
	Flags   []string `yaml:"flags"`
	Ldflags []string `yaml:"ldflags"`
	Version int      `yaml:"version"`
}

// GoReleaserConfig tracks configuration for goreleaser.
type GoReleaserConfig struct {
	Env     map[string]string
	Main    *string
	Dir     *string
	Goos    string
	Goarch  string
	Binary  string
	Flags   []string
	Ldflags []string
}

// ErrUnsupportedVersion indicates an unsupported Go builder version.
type ErrUnsupportedVersion struct {
	errors.WrappableError
}

// ErrInvalidDirectory indicates an invalid directory.
type ErrInvalidDirectory struct {
	errors.WrappableError
}

// ErrInvalidEnvironmentVariable indicates  an invalid environment variable.
type ErrInvalidEnvironmentVariable struct {
	errors.WrappableError
}

func configFromString(b []byte) (*GoReleaserConfig, error) {
	var cf goReleaserConfigFile
	if err := yaml.Unmarshal(b, &cf); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return fromConfig(&cf)
}

// ConfigFromFile reads the file located at path and builds a GoReleaserConfig
// from it.
func ConfigFromFile(path string) (*GoReleaserConfig, error) {
	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("%q: %w", path, err)
	}

	cfg, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("%q: os.ReadFile: %w", path, err)
	}

	c, err := configFromString(cfg)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", path, err)
	}
	return c, nil
}

func fromConfig(cf *goReleaserConfigFile) (*GoReleaserConfig, error) {
	if err := validateVersion(cf); err != nil {
		return nil, err
	}

	if err := validateMain(cf); err != nil {
		return nil, err
	}

	if err := validateDir(cf); err != nil {
		return nil, err
	}

	cfg := GoReleaserConfig{
		Goos:    cf.Goos,
		Goarch:  cf.Goarch,
		Flags:   cf.Flags,
		Ldflags: cf.Ldflags,
		Binary:  cf.Binary,
		Main:    cf.Main,
		Dir:     cf.Dir,
	}

	if err := cfg.setEnvs(cf); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validatePath(path string) error {
	err := utils.PathIsUnderCurrentDirectory(path)
	if err != nil {
		return convertPathError(err, "PathIsUnderCurrentDirectory")
	}
	return nil
}

func validateDir(cf *goReleaserConfigFile) error {
	if cf.Dir == nil {
		return nil
	}
	return validatePath(*cf.Dir)
}

func validateMain(cf *goReleaserConfigFile) error {
	if cf.Main == nil {
		return nil
	}

	// Validate the main path is under the current directory.
	if err := utils.PathIsUnderCurrentDirectory(*cf.Main); err != nil {
		return convertPathError(err, "PathIsUnderCurrentDirectory")
	}
	return nil
}

func convertPathError(e error, msg string) error {
	if e != nil {
		var errInternal *utils.ErrInternal
		var errPath *utils.ErrInvalidPath
		if errors.As(e, &errInternal) ||
			errors.As(e, &errPath) {
			return &ErrInvalidDirectory{}
		}
		return fmt.Errorf("%s: %w", msg, e)
	}
	return e
}

func validateVersion(cf *goReleaserConfigFile) error {
	_, exists := supportedVersions[cf.Version]
	if !exists {
		return errors.Errorf(&ErrUnsupportedVersion{}, "version '%d' not supported", cf.Version)
	}

	return nil
}

func (r *GoReleaserConfig) setEnvs(cf *goReleaserConfigFile) error {
	m := make(map[string]string)
	for _, e := range cf.Env {
		name, value, present := strings.Cut(e, "=")
		if !present {
			return errors.Errorf(&ErrInvalidEnvironmentVariable{}, "'%s' contains no '='", e)
		}
		m[name] = value
	}

	if len(m) > 0 {
		r.Env = m
	}

	return nil
}

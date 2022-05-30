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
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrorInvalidEnvironmentVariable = errors.New("invalid environment variable")
	ErrorUnsupportedVersion         = errors.New("version not supported")
)

var supportedVersions = map[int]bool{
	1: true,
}

type goReleaserConfigFile struct {
	Goos    string   `yaml:"goos"`
	Goarch  string   `yaml:"goarch"`
	Env     []string `yaml:"env"`
	Flags   []string `yaml:"flags"`
	Ldflags []string `yaml:"ldflags"`
	Binary  string   `yaml:"binary`
	Version int      `yaml:"version"`
	Main    *string  `yaml:"main"`
	Dir     *string  `yaml:"dir"`
}

type GoReleaserConfig struct {
	Goos    string
	Goarch  string
	Main    *string
	Dir     *string
	Env     map[string]string
	Flags   []string
	Ldflags []string
	Binary  string
}

func configFromString(b []byte) (*GoReleaserConfig, error) {
	var cf goReleaserConfigFile
	if err := yaml.Unmarshal(b, &cf); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return fromConfig(&cf)
}

func ConfigFromFile(pathfn string) (*GoReleaserConfig, error) {
	if err := validatePath(pathfn); err != nil {
		return nil, err
	}

	cfg, err := os.ReadFile(pathfn)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	return configFromString(cfg)
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
	return pathIsUnderCurrentDirectory(path)
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
	return pathIsUnderCurrentDirectory(*cf.Main)
}

func validateVersion(cf *goReleaserConfigFile) error {
	_, exists := supportedVersions[cf.Version]
	if !exists {
		return fmt.Errorf("%w:%d", ErrorUnsupportedVersion, cf.Version)
	}

	return nil
}

func (r *GoReleaserConfig) setEnvs(cf *goReleaserConfigFile) error {
	m := make(map[string]string)
	for _, e := range cf.Env {
		es := strings.Split(e, "=")
		if len(es) != 2 {
			return fmt.Errorf("%w: %s", ErrorInvalidEnvironmentVariable, e)
		}
		m[es[0]] = es[1]
	}

	if len(m) > 0 {
		r.Env = m
	}

	return nil
}

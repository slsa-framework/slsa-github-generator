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
	"encoding/json"
	"fmt"
	"os"

	"github.com/slsa-framework/slsa-github-generator/internal/runner"
)

type slsaIntegration struct {
	Workspace  string
	Inputs     *slsaInputs
	OutputPath string
}

type slsaDryRunOutput map[string][]slsaDryMetadata

type slsaDryMetadata struct {
	Name    string                `json:"name"`
	Digests slsaDigests           `json:"digests"`
	Steps   []*runner.CommandStep `json:"steps"`
}

type slsaInputs struct {
	Version         uint              `json:"version"`
	BuilderPath     string            `json:"builderPath"`
	WorkflowInputs  map[string]string `json:"workflowInputs"`
	WorkflowSecrets map[string]string `json:"workflowSecrets"`
	Base64Extras    string            `json:"base64Extras"`
	DryRun          bool              `json:"dryRun"`
	Artifacts       []slsaArtifact    `json:"artifacts"`
}

type slsaArtifact struct {
	Path    string      `json:"path"`
	Digests slsaDigests `json:"digests"`
}

type slsaDigests map[string]string

func slsaIntegrationNew() (*slsaIntegration, error) {
	workspace, err := readEnvPath("SLSA_WORKSPACE", true)
	if err != nil {
		return nil, err
	}

	outputsPath, err := readEnvPath("SLSA_OUTPUTS_PATH", false)
	if err != nil {
		return nil, err
	}

	inputs, err := getInputs()
	if err != nil {
		return nil, err
	}

	return &slsaIntegration{
		Workspace:  workspace,
		Inputs:     inputs,
		OutputPath: outputsPath,
	}, nil
}

func getInputs() (*slsaInputs, error) {
	content, err := readFileContent("SLSA_INPUTS_PATH")
	if err != nil {
		return nil, fmt.Errorf("%w: os.ReadFile", err)
	}
	var inputs slsaInputs
	reader := bytes.NewReader(content)
	if err := json.NewDecoder(reader).Decode(&inputs); err != nil {
		return nil, fmt.Errorf("%w: json.NewDecoder", err)
	}

	if inputs.Version != 1 {
		return nil, fmt.Errorf("%w: version", errorInvalidField)
	}
	return &inputs, err
}

func readFileContent(name string) ([]byte, error) {
	fn, err := readEnvPath(name, true)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("%w: os.ReadFile", err)
	}
	return content, nil
}

func readEnvPath(name string, exists bool) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%w: %s", errorEmpty, name)
	}

	if exists {
		if _, err := os.Stat(value); err != nil {
			return "", fmt.Errorf("%w: os.Stat", err)
		}
	}

	return value, nil
}

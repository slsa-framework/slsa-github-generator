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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/slsa-framework/slsa-github-generator/github"
	"github.com/slsa-framework/slsa-github-generator/internal/runner"
	"github.com/slsa-framework/slsa-github-generator/internal/utils"
)

var unknownTag = "unknown"

// See `go build help`.
// `-asmflags`, `-n`, `-mod`, `-installsuffix`, `-modfile`,
// `-workfile`, `-overlay`, `-pkgdir`, `-toolexec`, `-o`,
// `-modcacherw`, `-work` not supported for now.

var allowedBuildArgs = map[string]bool{
	"-a": true, "-race": true, "-msan": true, "-asan": true,
	"-v": true, "-x": true, "-buildinfo": true,
	"-buildmode": true, "-buildvcs": true, "-compiler": true,
	"-gccgoflags": true, "-gcflags": true,
	"-ldflags": true, "-linkshared": true,
	"-tags": true, "-trimpath": true,
}

var allowedEnvVariablePrefix = map[string]bool{
	"GO": true, "CGO_": true,
}

var (
	errEnvVariableNameEmpty      = errors.New("variable name empty")
	errUnsupportedArguments      = errors.New("unsupported arguments")
	errInvalidEnvArgument        = errors.New("invalid env argument")
	errEnvVariableNameNotAllowed = errors.New("invalid variable name")
	errInvalidFilename           = errors.New("invalid filename")
)

// GoBuild implements building a Go application.
type GoBuild struct {
	cfg *GoReleaserConfig
	// Note: static env variables are contained in cfg.Env.
	argEnv map[string]string
	goc    string
}

// GoBuildNew returns a new GoBuild.
func GoBuildNew(goc string, cfg *GoReleaserConfig) *GoBuild {
	c := GoBuild{
		cfg:    cfg,
		goc:    goc,
		argEnv: make(map[string]string),
	}

	return &c
}

// Run executes the build.
func (b *GoBuild) Run(dry bool) error {
	// Get directory.
	dir, err := b.getDir()
	if err != nil {
		return err
	}
	// Set flags.
	flags, err := b.generateFlags()
	if err != nil {
		return err
	}

	// Generate env variables.
	envs, err := b.generateCommandEnvVariables()
	if err != nil {
		return err
	}

	// Generate ldflags.
	ldflags, err := b.generateLdflags()
	if err != nil {
		return err
	}

	// Add ldflags.
	if ldflags != "" {
		flags = append(flags, fmt.Sprintf("-ldflags=%s", ldflags))
	}

	// A dry run prints the information that is trusted, before
	// the compiler is invoked.
	if dry {
		// Generate filename.
		// Note: the filename uses the config file and is resolved if it contains env variables.
		// `OUTPUT_BINARY` is only used during the actual compilation, an is a trusted
		// variable hardcoded in the reusable workflow, to avoid weird looking name
		// that may interfere with the compilation.
		filename, err := b.generateOutputFilename()
		if err != nil {
			return err
		}

		// Generate the command.
		com := b.generateCommand(flags, filename)

		env, err := b.generateCommandEnvVariables()
		if err != nil {
			return err
		}

		r := runner.CommandRunner{
			Steps: []*runner.CommandStep{
				{
					Command:    com,
					Env:        env,
					WorkingDir: dir,
				},
			},
		}

		steps, err := r.Dry()
		if err != nil {
			return err
		}

		// There is a single command in steps given to the runner so we are
		// assured to have only one step.
		menv, err := utils.MarshalToString(steps[0].Env)
		if err != nil {
			return err
		}
		command, err := utils.MarshalToString(steps[0].Command)
		if err != nil {
			return err
		}

		// Share the resolved name of the binary.
		if err := github.SetOutput("go-binary-name", filename); err != nil {
			return err
		}

		// Share the command used.
		if err := github.SetOutput("go-command", command); err != nil {
			return err
		}

		// Share the env variables used.
		if err := github.SetOutput("go-env", menv); err != nil {
			return err
		}

		// Share working directory necessary for issuing the vendoring command.
		return github.SetOutput("go-working-dir", dir)
	}

	binary, err := getOutputBinaryPath(os.Getenv("OUTPUT_BINARY"))
	if err != nil {
		return err
	}

	// Generate the command.
	command := b.generateCommand(flags, binary)

	fmt.Println("dir", dir)
	fmt.Println("binary", binary)
	fmt.Println("command", command)
	fmt.Println("env", envs)

	r := runner.CommandRunner{
		Steps: []*runner.CommandStep{
			{
				Command:    command,
				Env:        envs,
				WorkingDir: dir,
			},
		},
	}

	// TODO: Add a timeout?
	_, err = r.Run(context.Background())
	return err
}

func getOutputBinaryPath(binary string) (string, error) {
	// Use the name provider via env variable for the compilation.
	// This variable is trusted and defined by the re-usable workflow.
	// It should be set to an absolute path value.
	abinary, err := filepath.Abs(binary)
	if err != nil {
		return "", fmt.Errorf("filepath.Abs: %w", err)
	}

	if binary == "" {
		return "", fmt.Errorf("%w: OUTPUT_BINARY not defined", errInvalidFilename)
	}

	if binary != abinary {
		return "", fmt.Errorf("%w: %v is not an absolute path", errInvalidFilename, binary)
	}

	return binary, nil
}

func (b *GoBuild) getDir() (string, error) {
	if b.cfg.Dir == nil {
		return os.Getenv("PWD"), nil
	}

	// Note: validation of the dir is done in config.go
	fp, err := filepath.Abs(*b.cfg.Dir)
	if err != nil {
		return "", err
	}

	return fp, nil
}

func (b *GoBuild) generateCommand(flags []string, binary string) []string {
	var command []string
	command = append(command, flags...)
	command = append(command, "-o", binary)

	// Add the entry point.
	if b.cfg.Main != nil {
		command = append(command, *b.cfg.Main)
	}
	return command
}

func (b *GoBuild) generateCommandEnvVariables() ([]string, error) {
	var env []string

	if b.cfg.Goos == "" {
		return nil, fmt.Errorf("%w: %s", errEnvVariableNameEmpty, "GOOS")
	}
	env = append(env, fmt.Sprintf("GOOS=%s", b.cfg.Goos))

	if b.cfg.Goarch == "" {
		return nil, fmt.Errorf("%w: %s", errEnvVariableNameEmpty, "GOARCH")
	}
	env = append(env, fmt.Sprintf("GOARCH=%s", b.cfg.Goarch))

	// Set env variables from config file.
	for k, v := range b.cfg.Env {
		if !isAllowedEnvVariable(k) {
			return env, fmt.Errorf("%w: %s", errEnvVariableNameNotAllowed, v)
		}

		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env, nil
}

// SetArgEnvVariables sets static environment variables.
func (b *GoBuild) SetArgEnvVariables(envs string) error {
	// Notes:
	// - I've tried running the re-usable workflow in a step
	// and set the env variable in a previous step, but found that a re-usable workflow is not
	// allowed to run in a step; they have to run as `job.uses`. Using `job.env` with `job.uses`
	// is not allowed.
	// - We don't want to allow env variables set in the workflow because of injections
	// e.g. LD_PRELOAD, etc.
	if envs == "" {
		return nil
	}

	for _, e := range strings.Split(envs, ",") {
		s := strings.Trim(e, " ")

		sp := strings.Split(s, ":")
		if len(sp) != 2 {
			return fmt.Errorf("%w: %s", errInvalidEnvArgument, s)
		}
		name := strings.Trim(sp[0], " ")
		value := strings.Trim(sp[1], " ")

		fmt.Printf("arg env: %s:%s\n", name, value)
		b.argEnv[name] = value
	}
	return nil
}

func (b *GoBuild) generateOutputFilename() (string, error) {
	// Note: the `.` is needed to accommodate the semantic version
	// as part of the name.
	const alpha = ".abcdefghijklmnopqrstuvwxyz1234567890-_"

	var name string

	// Special variables.
	name, err := b.resolveSpecialVariables(b.cfg.Binary)
	if err != nil {
		return "", err
	}

	// Dynamic env variables provided by caller.
	name, err = b.resolveEnvVariables(name)
	if err != nil {
		return "", err
	}

	for _, char := range name {
		if !strings.Contains(alpha, strings.ToLower(string(char))) {
			return "", fmt.Errorf("%w: found character '%c'", errInvalidFilename, char)
		}
	}

	if name == "" {
		return "", fmt.Errorf("%w: filename is empty", errInvalidFilename)
	}

	// Validate the path.
	if err := validatePath(name); err != nil {
		return "", err
	}

	return name, nil
}

func (b *GoBuild) generateFlags() ([]string, error) {
	// -x
	flags := []string{b.goc, "build", "-mod=vendor"}

	for _, v := range b.cfg.Flags {
		if !isAllowedArg(v) {
			return nil, fmt.Errorf("%w: %s", errUnsupportedArguments, v)
		}
		flags = append(flags, v)
	}
	return flags, nil
}

func isAllowedArg(arg string) bool {
	for k := range allowedBuildArgs {
		if strings.HasPrefix(arg, k) {
			return true
		}
	}
	return false
}

// Check if the env variable is allowed. We want to avoid
// variable injection, e.g. LD_PRELOAD, etc.
// See an overview in https://www.hale-legacy.com/class/security/s20/handout/slides-env-vars.pdf.
func isAllowedEnvVariable(name string) bool {
	for k := range allowedEnvVariablePrefix {
		if strings.HasPrefix(name, k) {
			return true
		}
	}
	return false
}

// TODO: maybe not needed if handled directly by go compiler.
func (b *GoBuild) generateLdflags() (string, error) {
	var a []string

	// Resolve variables.
	for _, v := range b.cfg.Ldflags {
		// Special variables.
		v, err := b.resolveSpecialVariables(v)
		if err != nil {
			return "", err
		}

		// Dynamic env variables provided by caller.
		v, err = b.resolveEnvVariables(v)
		if err != nil {
			return "", err
		}
		a = append(a, v)
	}

	if len(a) > 0 {
		return strings.Join(a, " "), nil
	}

	return "", nil
}

func (b *GoBuild) resolveSpecialVariables(s string) (string, error) {
	reVar := regexp.MustCompile(`{{ \.([A-Z][a-z]*) }}`)
	names := reVar.FindAllString(s, -1)
	for _, n := range names {
		name := strings.ReplaceAll(n, "{{ .", "")
		name = strings.ReplaceAll(name, " }}", "")

		switch name {
		case "Os":
			if b.cfg.Goos == "" {
				return "", fmt.Errorf("%w: {{ .Os }}", errEnvVariableNameEmpty)
			}
			s = strings.ReplaceAll(s, n, b.cfg.Goos)

		case "Arch":
			if b.cfg.Goarch == "" {
				return "", fmt.Errorf("%w: {{ .Arch }}", errEnvVariableNameEmpty)
			}
			s = strings.ReplaceAll(s, n, b.cfg.Goarch)

		case "Tag":
			tag := getTag()
			s = strings.ReplaceAll(s, n, tag)
		default:
			return "", fmt.Errorf("%w: %s", errInvalidEnvArgument, n)
		}
	}
	return s, nil
}

func (b *GoBuild) resolveEnvVariables(s string) (string, error) {
	reDyn := regexp.MustCompile(`{{ \.Env\.(\w+) }}`)
	names := reDyn.FindAllString(s, -1)
	for _, n := range names {
		name := strings.ReplaceAll(n, "{{ .Env.", "")
		name = strings.ReplaceAll(name, " }}", "")

		val, exists := b.argEnv[name]
		if !exists {
			return "", fmt.Errorf("%w: %s", errEnvVariableNameEmpty, n)
		}
		s = strings.ReplaceAll(s, n, val)
	}
	return s, nil
}

func getTag() string {
	tag := os.Getenv("GITHUB_REF_NAME")
	if tag == "" {
		return unknownTag
	}
	return tag
}

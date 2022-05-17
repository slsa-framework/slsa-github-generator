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
	"syscall"
)

var (
	errorEnvVariableNameEmpty      = errors.New("env variable empty or not set")
	errorUnsupportedArguments      = errors.New("argument not supported")
	errorInvalidEnvArgument        = errors.New("invalid env passed via argument")
	errorEnvVariableNameNotAllowed = errors.New("env variable not allowed")
	errorInvalidFilename           = errors.New("invalid filename")
	errorEmptyFilename             = errors.New("filename is not set")
)

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

type GoBuild struct {
	cfg *GoReleaserConfig
	goc string
	// Note: static env variables are contained in cfg.Env.
	argEnv  map[string]string
	ldflags string
}

func GoBuildNew(goc string, cfg *GoReleaserConfig) *GoBuild {
	c := GoBuild{
		cfg:    cfg,
		goc:    goc,
		argEnv: make(map[string]string),
	}

	return &c
}

func (b *GoBuild) Run(dry bool) error {
	// Set flags.
	flags, err := b.generateFlags()
	if err != nil {
		return err
	}

	// Generate env variables.
	envs, err := b.generateEnvVariables()
	if err != nil {
		return err
	}

	// Generate ldflags.
	ldflags, err := b.generateLdflags()
	if err != nil {
		return err
	}

	// Add ldflags.
	if len(ldflags) > 0 {
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

		// Share the resolved name of the binary.
		fmt.Printf("::set-output name=go-binary-name::%s\n", filename)
		command, err := marshallToString(com)
		if err != nil {
			return err
		}
		// Share the command used.
		fmt.Printf("::set-output name=go-command::%s\n", command)

		env, err := b.generateCommandEnvVariables()
		if err != nil {
			return err
		}

		menv, err := marshallToString(env)
		if err != nil {
			return err
		}

		// Share the env variables used.
		fmt.Printf("::set-output name=go-env::%s\n", menv)
		return nil
	}

	// Use the name provider via env variable for the compilation.
	// This variable is trusted and defined by the re-usable workflow.
	binary := os.Getenv("OUTPUT_BINARY")
	if binary == "" {
		return fmt.Errorf("OUTPUT_BINARY not defined")
	}

	// Generate the command.
	command := b.generateCommand(flags, binary)

	fmt.Println("binary", binary)
	fmt.Println("command", command)
	fmt.Println("env", envs)
	return syscall.Exec(b.goc, command, envs)
}

func (b *GoBuild) generateCommand(flags []string, binary string) []string {
	command := append(flags, []string{"-o", binary}...)

	// Add the entry point.
	if b.cfg.Main != nil {
		command = append(command, *b.cfg.Main)
	}
	return command
}

func (b *GoBuild) generateCommandEnvVariables() ([]string, error) {
	var env []string

	if b.cfg.Goos == "" {
		return nil, fmt.Errorf("%w: %s", errorEnvVariableNameEmpty, "GOOS")
	}
	env = append(env, fmt.Sprintf("GOOS=%s", b.cfg.Goos))

	if b.cfg.Goarch == "" {
		return nil, fmt.Errorf("%w: %s", errorEnvVariableNameEmpty, "GOARCH")
	}
	env = append(env, fmt.Sprintf("GOARCH=%s", b.cfg.Goarch))

	// Set env variables from config file.
	for k, v := range b.cfg.Env {
		if !isAllowedEnvVariable(k) {
			return env, fmt.Errorf("%w: %s", errorEnvVariableNameNotAllowed, v)
		}

		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env, nil
}

func (b *GoBuild) generateEnvVariables() ([]string, error) {
	env := os.Environ()

	cenv, err := b.generateCommandEnvVariables()
	if err != nil {
		return cenv, err
	}

	env = append(env, cenv...)

	return env, nil
}

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
			return fmt.Errorf("%w: %s", errorInvalidEnvArgument, s)
		}
		name := strings.Trim(sp[0], " ")
		value := strings.Trim(sp[1], " ")

		fmt.Printf("arg env: %s:%s\n", name, value)
		b.argEnv[name] = value

	}
	return nil
}

func (b *GoBuild) generateOutputFilename() (string, error) {
	const alpha = "abcdefghijklmnopqrstuvwxyz1234567890-_"

	var name string

	// Replace .Os variable.
	if strings.Contains(b.cfg.Binary, "{{ .Os }}") && b.cfg.Goos == "" {
		return "", fmt.Errorf("%w", errorEnvVariableNameEmpty)
	}
	name = strings.ReplaceAll(b.cfg.Binary, "{{ .Os }}", b.cfg.Goos)

	// Replace .Arch variable.
	if strings.Contains(name, "{{ .Arch }}") && b.cfg.Goarch == "" {
		return "", fmt.Errorf("%w", errorEnvVariableNameEmpty)
	}
	name = strings.ReplaceAll(name, "{{ .Arch }}", b.cfg.Goarch)

	for _, char := range name {
		if !strings.Contains(alpha, strings.ToLower(string(char))) {
			return "", fmt.Errorf("%w: found character '%c'", errorInvalidFilename, char)
		}
	}

	if name == "" {
		return "", fmt.Errorf("%w", errorEmptyFilename)
	}
	return name, nil
}

func (b *GoBuild) generateFlags() ([]string, error) {
	// -x
	flags := []string{b.goc, "build", "-mod=vendor"}

	for _, v := range b.cfg.Flags {
		if !isAllowedArg(v) {
			return nil, fmt.Errorf("%w: %s", errorUnsupportedArguments, v)
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

	for _, v := range b.cfg.Ldflags {
		var res string
		ss := "{{ .Env."
		es := "}}"
		found := false
		for {
			start := strings.Index(v, ss)
			if start == -1 {
				break
			}
			end := strings.Index(string(v[start+len(ss):]), es)
			if end == -1 {
				return "", fmt.Errorf("%w: %s", errorInvalidEnvArgument, v)
			}

			name := strings.Trim(string(v[start+len(ss):start+len(ss)+end]), " ")
			if name == "" {
				return "", fmt.Errorf("%w: %s", errorEnvVariableNameEmpty, v)
			}

			val, exists := b.argEnv[name]
			if !exists {
				return "", fmt.Errorf("%w: %s", errorEnvVariableNameEmpty, name)
			}
			res = fmt.Sprintf("%s%s%s", res, v[:start], val)
			found = true
			v = v[start+len(ss)+end+len(es):]
		}
		if !found {
			res = v
		}
		a = append(a, res)
	}
	if len(a) > 0 {
		return strings.Join(a, " "), nil
	}

	return "", nil
}

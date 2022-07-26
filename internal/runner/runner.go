package runner

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// CommandRunner runs commands and returns the build steps that were run.
type CommandRunner struct {
	// Env is global environment variables passed to all commands.
	Env []string

	// Steps are the steps to execute.
	Steps []*CommandStep
}

// CommandStep is a command that was executed by the builder.
type CommandStep struct {
	// Command is the command that was run.
	Command []string `json:"command"`

	// Env are the environment variables passed to the command.
	Env []string `json:"env"`

	// WorkingDir is the working directory the command was executed in.
	WorkingDir string `json:"workingDir"`
}

// Run excutes a series of commands and returns the steps that were executed
// successfully. Commands are run in sequence and are expected to return a zero
// exit status.
//
// Global environment variables are merged with steps environment variables in
// the returned steps. In the case of duplicates the last occurrence has precidence.
// Environment variables are *not* inherited from the current process.
//
// The returned CommandSteps should be included in the buildConfig provenance.
// These are *not* the same as the runner commands. Env vars are sanitized, pwd
// is changed to the absolute path, and only commands that executed
// successfully are returned.
func (r *CommandRunner) Run() (steps []*CommandStep, err error) {
	var originalwd string
	originalwd, err = os.Getwd()
	if err != nil {
		return // steps, err
	}
	defer func() {
		// Change directories back to the original working directory but only
		// return the error returned by Chdir if no other error occurred.
		if chDirErr := os.Chdir(originalwd); err == nil {
			// NOTE: err is returned by the function after the defer is called.
			err = chDirErr
		}
	}()

	for _, step := range r.Steps {
		if len(step.Command) == 0 {
			err = errors.New("command is empty")
			return // steps, err
		}

		cmd := step.Command[0]
		args := step.Command[1:]
		// TODO: Add some kind of LD_PRELOAD protection?
		env := dedupEnv(append(r.Env, step.Env...))

		// Set the POSIX PWD env var.
		var pwd string
		pwd, err = filepath.Abs(step.WorkingDir)
		if err != nil {
			return // steps, err
		}
		env = append(env, "PWD="+pwd)

		err = os.Chdir(pwd)
		if err != nil {
			return // steps, err
		}

		// NOTE: We use syscall.Exec directly because we are executing the
		// commands in sequence and we want full control over the environment.
		// The stdlib exec package deduplicates env vars and so it's hard to
		// know exactly what the command was run with.
		// NOTE: The executed command will inherit stdout/stderr.
		err = syscall.Exec(cmd, args, env)
		if err != nil {
			return // steps, err
		}

		steps = append(steps, &CommandStep{
			Command:    append([]string{cmd}, args...),
			Env:        env,
			WorkingDir: pwd,
		})
	}
	return // steps, err
}

// dedupEnv returns a copy of env with any duplicates removed, in favor of
// later values.
// Items not of the normal environment "key=value" form are preserved unchanged.
// NOTE: adapted from the stdlib exec package.
func dedupEnv(env []string) []string {
	// Construct the output in reverse order, to preserve the
	// last occurrence of each key.
	out := make([]string, 0, len(env))
	saw := make(map[string]bool, len(env))
	for n := len(env); n > 0; n-- {
		kv := env[n-1]

		i := strings.Index(kv, "=")
		if i == 0 {
			// We observe in practice keys with a single leading "=" on Windows.
			i = strings.Index(kv[1:], "=") + 1
		}
		if i < 0 {
			if kv != "" {
				// The entry is not of the form "key=value" (as it is required to be).
				// Leave it as-is for now.
				out = append(out, kv)
			}
			continue
		}
		k := kv[:i]
		if saw[k] {
			continue
		}

		saw[k] = true
		out = append(out, kv)
	}

	// Now reverse the slice to restore the original order.
	for i := 0; i < len(out)/2; i++ {
		j := len(out) - i - 1
		out[i], out[j] = out[j], out[i]
	}

	return out
}

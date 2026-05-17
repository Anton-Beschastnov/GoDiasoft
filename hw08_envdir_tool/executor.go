package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Start with current environment
	environ := os.Environ()

	// Build a map of current env vars for easy modification
	envMap := make(map[string]string)
	for _, e := range environ {
		if idx := strings.Index(e, "="); idx != -1 {
			envMap[e[:idx]] = e[idx+1:]
		}
	}

	// Apply changes from env
	for name, val := range env {
		if val.NeedRemove {
			delete(envMap, name)
		} else {
			envMap[name] = val.Value
		}
	}

	// Convert map back to slice
	newEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		newEnv = append(newEnv, k+"="+v)
	}
	command.Env = newEnv

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}

	return 0
}

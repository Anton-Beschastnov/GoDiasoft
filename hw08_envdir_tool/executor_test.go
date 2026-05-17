package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("empty command", func(t *testing.T) {
		code := RunCmd([]string{}, nil)
		require.Equal(t, 1, code)
	})

	t.Run("successful command", func(t *testing.T) {
		code := RunCmd([]string{"go", "version"}, nil)
		require.Equal(t, 0, code)
	})

	t.Run("command with exit code", func(t *testing.T) {
		code := RunCmd([]string{"go", "help", "nonexistent"}, nil)
		require.NotEqual(t, 0, code)
	})

	t.Run("set env variable", func(t *testing.T) {
		env := Environment{
			"TEST_VAR": {Value: "test_value"},
		}
		code := RunCmd([]string{"go", "version"}, env)
		require.Equal(t, 0, code)
	})

	t.Run("remove env variable", func(t *testing.T) {
		os.Setenv("TEST_TO_REMOVE", "value")
		defer os.Unsetenv("TEST_TO_REMOVE")

		env := Environment{
			"TEST_TO_REMOVE": {NeedRemove: true},
		}
		code := RunCmd([]string{"go", "version"}, env)
		require.Equal(t, 0, code)
	})

	t.Run("non-existent command", func(t *testing.T) {
		code := RunCmd([]string{"nonexistent-command-12345"}, nil)
		require.Equal(t, 1, code)
	})
}

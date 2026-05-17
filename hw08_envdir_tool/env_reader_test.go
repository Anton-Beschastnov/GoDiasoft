package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("testdata/env", func(t *testing.T) {
		env, err := ReadDir("testdata/env")
		require.NoError(t, err)

		require.Equal(t, "bar", env["BAR"].Value)
		require.False(t, env["BAR"].NeedRemove)

		require.Equal(t, "   foo\nwith new line", env["FOO"].Value)
		require.False(t, env["FOO"].NeedRemove)

		require.Equal(t, `"hello"`, env["HELLO"].Value)
		require.False(t, env["HELLO"].NeedRemove)

		require.Equal(t, "", env["EMPTY"].Value)
		require.False(t, env["EMPTY"].NeedRemove)

		require.True(t, env["UNSET"].NeedRemove)
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := ReadDir("non-existent-dir")
		require.Error(t, err)
	})

	t.Run("skip files with = in name", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "VALID"), []byte("value"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "INVALID=NAME"), []byte("value"), 0o644)
		require.NoError(t, err)

		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		require.Contains(t, env, "VALID")
		require.NotContains(t, env, "INVALID=NAME")
	})

	t.Run("skip directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.Mkdir(filepath.Join(tmpDir, "SUBDIR"), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "FILE"), []byte("value"), 0o644)
		require.NoError(t, err)

		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		require.Contains(t, env, "FILE")
		require.NotContains(t, env, "SUBDIR")
	})

	t.Run("trim trailing spaces and tabs", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "VAR"), []byte("value   \t  "), 0o644)
		require.NoError(t, err)

		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		require.Equal(t, "value", env["VAR"].Value)
	})
}

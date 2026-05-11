package main

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	// Place your code here.
}

func TestCopy_TestData(t *testing.T) {
	testDir := "testdata"
	input := filepath.Join(testDir, "input.txt")
	tempDir := t.TempDir()

	testCases := []struct {
		name     string
		offset   int64
		limit    int64
		expected string
	}{
		{"full copy", 0, 0, "out_offset0_limit0.txt"},
		{"limit 10", 0, 10, "out_offset0_limit10.txt"},
		{"limit 1000", 0, 1000, "out_offset0_limit1000.txt"},
		{"limit 10000", 0, 10000, "out_offset0_limit10000.txt"},
		{"offset 100 limit 1000", 100, 1000, "out_offset100_limit1000.txt"},
		{"offset 6000 limit 1000", 6000, 1000, "out_offset6000_limit1000.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			toPath := filepath.Join(tempDir, tc.expected)
			expectedPath := filepath.Join(testDir, tc.expected)

			err := Copy(input, toPath, tc.offset, tc.limit)
			require.NoError(t, err)

			actual, err := os.ReadFile(toPath)
			require.NoError(t, err)

			expected, err := os.ReadFile(expectedPath)
			require.NoError(t, err)

			require.Equal(t, expected, actual)
		})
	}
}

func TestCopy_Validation(t *testing.T) {
	tempDir := t.TempDir()
	fromPath := filepath.Join(tempDir, "small.txt")
	os.WriteFile(fromPath, []byte("small file"), 0644)

	t.Run("offset too big", func(t *testing.T) {
		err := Copy(fromPath, filepath.Join(tempDir, "out.txt"), 100, 0)
		require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
	})

	t.Run("limit bigger than file", func(t *testing.T) {
		toPath := filepath.Join(tempDir, "out_limit.txt")
		err := Copy(fromPath, toPath, 0, 500)
		require.NoError(t, err)

		res, _ := os.ReadFile(toPath)
		require.Equal(t, []byte("small file"), res)
	})
}

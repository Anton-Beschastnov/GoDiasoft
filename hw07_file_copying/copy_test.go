package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	testDir := "testdata"
	tempDir := t.TempDir()
	t.Run("TestData", func(t *testing.T) {
		testCases := []struct {
			name, exp string
			off, lim  int64
		}{
			{"full", "out_offset0_limit0.txt", 0, 0},
			{"limit10", "out_offset0_limit10.txt", 0, 10},
			{"limit1000", "out_offset0_limit1000.txt", 0, 1000},
			{"limit10000", "out_offset0_limit10000.txt", 0, 10000},
			{"off100_lim1000", "out_offset100_limit1000.txt", 100, 1000},
			{"off6000_lim1000", "out_offset6000_limit1000.txt", 6000, 1000},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				to := filepath.Join(tempDir, tc.exp)
				err := Copy(filepath.Join(testDir, "input.txt"), to, tc.off, tc.lim)
				require.NoError(t, err)
				act, _ := os.ReadFile(to)
				exp, _ := os.ReadFile(filepath.Join(testDir, tc.exp))
				if len(act) < len(exp) {
					exp = exp[:len(act)]
				}
				require.Equal(t, exp, act)
			})
		}
	})
	t.Run("Validation", func(t *testing.T) {
		from := filepath.Join(tempDir, "small.txt")
		os.WriteFile(from, []byte("small file"), 0o644)
		t.Run("offset_too_big", func(t *testing.T) {
			err := Copy(from, filepath.Join(tempDir, "out.txt"), 100, 0)
			require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
		})
		t.Run("limit_bigger_than_file", func(t *testing.T) {
			to := filepath.Join(tempDir, "out_limit.txt")
			require.NoError(t, Copy(from, to, 0, 500))
			res, _ := os.ReadFile(to)
			require.Equal(t, []byte("small file"), res)
		})
	})
}

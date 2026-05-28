// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat_Added(t *testing.T) {
	t.Run("skip empty lines and invalid json", func(t *testing.T) {
		data := `{"Id":1,"Email":"test1@example.com"}

{"Id":2,"Email":"test2@example.com"}
not a json
{"Id":3,"Email":"test3@another.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"example.com": 2, "another.com": 1}, result)
	})

	t.Run("handle empty input", func(t *testing.T) {
		data := ``
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("case insensitivity of domain", func(t *testing.T) {
		data := `{"Email":"test1@example.COM"}
{"Email":"test2@EXAMPLE.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"example.com": 2}, result)
	})
}

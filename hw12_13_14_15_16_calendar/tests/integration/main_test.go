package integration

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Here we can setup our test environment
	// For example, wait for services to be ready
	exitCode := m.Run()
	// And here we can tear it down
	os.Exit(exitCode)
}

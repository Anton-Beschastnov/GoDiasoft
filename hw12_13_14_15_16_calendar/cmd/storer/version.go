package main

import (
	"fmt"
)

var (
	version   = "develop"
	buildDate = ""
	gitHash   = ""
)

func printVersion() {
	fmt.Printf("calendar_storer version %s (built: %s, hash: %s)\n", version, buildDate, gitHash)
}

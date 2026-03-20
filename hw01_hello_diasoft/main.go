package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	const input = "Hello, DIASOFT!"
	fmt.Println(reverse.String(input))
}

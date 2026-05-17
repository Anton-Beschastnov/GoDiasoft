package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeoutStr := flag.String("timeout", "10s", "timeout for connection")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	timeout, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid timeout value: %v\n", err)
		os.Exit(1)
	}

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errChan := make(chan error, 2)

	go func() {
		errChan <- client.Send()
	}()

	go func() {
		errChan <- client.Receive()
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "\nBye-bye")
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "...Error: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	}
}

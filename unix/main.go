package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Get socket path from command line argument
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <socket_path>\n", os.Args[0])
		os.Exit(1)
	}
	socketPath := os.Args[1]

	// Listen on Unix domain socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on socket %s: %v\n", socketPath, err)
		os.Exit(1)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	// Print ready signal to stdout so parent knows we're listening
	fmt.Println("ready")

	// Accept single connection
	conn, err := listener.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to accept connection: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Handle messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())

		switch msg {
		case "ping":
			fmt.Fprintln(conn, "pong")
		case "quit":
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", msg)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading connection: %v\n", err)
		os.Exit(1)
	}
}
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Get port from command line argument
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <port>\n", os.Args[0])
		os.Exit(1)
	}
	port := os.Args[1]

	// Listen on TCP port
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on port %s: %v\n", port, err)
		os.Exit(1)
	}
	defer listener.Close()

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
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		
		switch msg {
		case "ping":
			fmt.Println("pong")
		case "quit":
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", msg)
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}
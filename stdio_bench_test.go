package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func BenchmarkStdio(b *testing.B) {
	// Create temporary directory for plugin binary
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "stdio-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./stdio")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Start the plugin process
	cmd := exec.Command(pluginPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		b.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatalf("Failed to start plugin: %v", err)
	}

	// Create scanner for reading responses
	scanner := bufio.NewScanner(stdout)

	// Reset timer after setup
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		// Send ping
		if _, err := fmt.Fprintln(stdin, "ping"); err != nil {
			b.Fatalf("Failed to send ping: %v", err)
		}

		// Read pong
		if !scanner.Scan() {
			b.Fatalf("Failed to read response")
		}
		response := scanner.Text()
		if response != "pong" {
			b.Fatalf("Unexpected response: %s", response)
		}
	}

	b.StopTimer()

	// Cleanup: send quit command
	if _, err := fmt.Fprintln(stdin, "quit"); err != nil {
		b.Errorf("Failed to send quit command: %v", err)
	}

	// Wait for process to exit
	if err := cmd.Wait(); err != nil {
		b.Errorf("Plugin process exited with error: %v", err)
	}
}

func TestStdioPingPong(t *testing.T) {
	// Create temporary directory for plugin binary
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "stdio-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./stdio")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Start the plugin process
	cmd := exec.Command(pluginPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start plugin: %v", err)
	}

	// Create scanner for reading responses
	scanner := bufio.NewScanner(stdout)

	// Test ping/pong
	for i := 0; i < 5; i++ {
		// Send ping
		if _, err := fmt.Fprintln(stdin, "ping"); err != nil {
			t.Fatalf("Failed to send ping: %v", err)
		}

		// Read pong
		if !scanner.Scan() {
			t.Fatalf("Failed to read response")
		}
		response := scanner.Text()
		if response != "pong" {
			t.Fatalf("Unexpected response: %s", response)
		}
	}

	// Test quit command
	if _, err := fmt.Fprintln(stdin, "quit"); err != nil {
		t.Errorf("Failed to send quit command: %v", err)
	}

	// Wait for process to exit
	if err := cmd.Wait(); err != nil {
		t.Errorf("Plugin process exited with error: %v", err)
	}
}
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkTCP(b *testing.B) {
	// Create temporary directory for plugin binary
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "tcp-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./tcp")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Find available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		b.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start the plugin process with port argument
	cmd := exec.Command(pluginPath, fmt.Sprintf("%d", port))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		b.Fatalf("Failed to start plugin: %v", err)
	}

	// Wait for plugin to be ready
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		b.Fatalf("Plugin did not signal ready")
	}

	// Connect to plugin
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		b.Fatalf("Failed to connect to plugin: %v", err)
	}
	defer conn.Close()

	connScanner := bufio.NewScanner(conn)

	// Reset timer after setup
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		// Send ping
		if _, err := fmt.Fprintln(conn, "ping"); err != nil {
			b.Fatalf("Failed to send ping: %v", err)
		}

		// Read pong
		if !connScanner.Scan() {
			b.Fatalf("Failed to read response")
		}
		response := connScanner.Text()
		if response != "pong" {
			b.Fatalf("Unexpected response: %s", response)
		}
	}

	b.StopTimer()

	// Cleanup: send quit command
	if _, err := fmt.Fprintln(conn, "quit"); err != nil {
		b.Errorf("Failed to send quit command: %v", err)
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			b.Errorf("Plugin process exited with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		cmd.Process.Kill()
		b.Errorf("Plugin process did not exit after quit command")
	}
}

func TestTCPPingPong(t *testing.T) {
	// Create temporary directory for plugin binary
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "tcp-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./tcp")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Find available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start the plugin process with port argument
	cmd := exec.Command(pluginPath, fmt.Sprintf("%d", port))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start plugin: %v", err)
	}

	// Wait for plugin to be ready
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		t.Fatalf("Plugin did not signal ready")
	}

	// Connect to plugin
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("Failed to connect to plugin: %v", err)
	}
	defer conn.Close()

	connScanner := bufio.NewScanner(conn)

	// Test ping/pong
	for i := 0; i < 5; i++ {
		// Send ping
		if _, err := fmt.Fprintln(conn, "ping"); err != nil {
			t.Fatalf("Failed to send ping: %v", err)
		}

		// Read pong
		if !connScanner.Scan() {
			t.Fatalf("Failed to read response")
		}
		response := connScanner.Text()
		if response != "pong" {
			t.Fatalf("Unexpected response: %s", response)
		}
	}

	// Test quit command
	if _, err := fmt.Fprintln(conn, "quit"); err != nil {
		t.Errorf("Failed to send quit command: %v", err)
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Plugin process exited with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		cmd.Process.Kill()
		t.Errorf("Plugin process did not exit after quit command")
	}
}
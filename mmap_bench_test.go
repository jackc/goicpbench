package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

const (
	pageSize = 4096
	cmdOffset = 0
	msgOffset = 64
)

func BenchmarkMmap(b *testing.B) {
	// Create temporary directory for plugin binary and shared memory
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "mmap-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./mmap")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Create shared memory file
	shmPath := filepath.Join(tmpDir, "shared.mem")
	shmFile, err := os.Create(shmPath)
	if err != nil {
		b.Fatalf("Failed to create shared memory file: %v", err)
	}

	// Resize file to page size
	if err := shmFile.Truncate(pageSize); err != nil {
		b.Fatalf("Failed to resize shared memory file: %v", err)
	}

	// Memory map the file
	data, err := syscall.Mmap(int(shmFile.Fd()), 0, pageSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		b.Fatalf("Failed to mmap file: %v", err)
	}
	defer syscall.Munmap(data)
	shmFile.Close()

	// Initialize shared memory to zeros
	for i := range data {
		data[i] = 0
	}

	// Create atomic pointer to the command byte
	cmdPtr := (*uint32)(unsafe.Pointer(&data[cmdOffset]))

	// Start the plugin process
	cmd := exec.Command(pluginPath, shmPath)
	if err := cmd.Start(); err != nil {
		b.Fatalf("Failed to start plugin: %v", err)
	}

	// Wait for plugin to be ready
	for i := 0; i < 100; i++ {
		if atomic.LoadUint32(cmdPtr) == 1 {
			msgBytes := data[msgOffset:msgOffset+32]
			if string(msgBytes[:5]) == "ready" {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		if i == 99 {
			b.Fatalf("Plugin did not signal ready")
		}
	}

	// Reset timer after setup
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		// Write ping command
		copy(data[msgOffset:], []byte("ping\x00"))
		// Signal command ready
		atomic.StoreUint32(cmdPtr, 2)

		// Wait for response
		for atomic.LoadUint32(cmdPtr) != 1 {
			// Busy wait
		}

		// Verify response
		msgBytes := data[msgOffset:msgOffset+32]
		if string(msgBytes[:4]) != "pong" {
			b.Fatalf("Unexpected response: %s", string(msgBytes[:4]))
		}
	}

	b.StopTimer()

	// Cleanup: send quit command
	copy(data[msgOffset:], []byte("quit\x00"))
	atomic.StoreUint32(cmdPtr, 2)

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

func TestMmapPingPong(t *testing.T) {
	// Create temporary directory for plugin binary and shared memory
	tmpDir, err := os.MkdirTemp("", "goipcbench-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build the plugin
	pluginPath := filepath.Join(tmpDir, "mmap-plugin")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, "./mmap")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build plugin: %v\nOutput: %s", err, output)
	}

	// Create shared memory file
	shmPath := filepath.Join(tmpDir, "shared.mem")
	shmFile, err := os.Create(shmPath)
	if err != nil {
		t.Fatalf("Failed to create shared memory file: %v", err)
	}

	// Resize file to page size
	if err := shmFile.Truncate(pageSize); err != nil {
		t.Fatalf("Failed to resize shared memory file: %v", err)
	}

	// Memory map the file
	data, err := syscall.Mmap(int(shmFile.Fd()), 0, pageSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		t.Fatalf("Failed to mmap file: %v", err)
	}
	defer syscall.Munmap(data)
	shmFile.Close()

	// Initialize shared memory to zeros
	for i := range data {
		data[i] = 0
	}

	// Create atomic pointer to the command byte
	cmdPtr := (*uint32)(unsafe.Pointer(&data[cmdOffset]))

	// Start the plugin process
	cmd := exec.Command(pluginPath, shmPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start plugin: %v", err)
	}

	// Wait for plugin to be ready
	for i := 0; i < 100; i++ {
		if atomic.LoadUint32(cmdPtr) == 1 {
			msgBytes := data[msgOffset:msgOffset+32]
			if string(msgBytes[:5]) == "ready" {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		if i == 99 {
			t.Fatalf("Plugin did not signal ready")
		}
	}

	// Test ping/pong
	for i := 0; i < 5; i++ {
		// Write ping command
		copy(data[msgOffset:], []byte("ping\x00"))
		// Signal command ready
		atomic.StoreUint32(cmdPtr, 2)

		// Wait for response
		for atomic.LoadUint32(cmdPtr) != 1 {
			time.Sleep(100 * time.Nanosecond)
		}

		// Verify response
		msgBytes := data[msgOffset:msgOffset+32]
		if string(msgBytes[:4]) != "pong" {
			t.Fatalf("Unexpected response: %s", string(msgBytes[:4]))
		}
	}

	// Test quit command
	copy(data[msgOffset:], []byte("quit\x00"))
	atomic.StoreUint32(cmdPtr, 2)

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
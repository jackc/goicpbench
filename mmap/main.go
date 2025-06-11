package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	pageSize = 4096
	cmdOffset = 0
	msgOffset = 64
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <shared_memory_file>\n", os.Args[0])
		os.Exit(1)
	}
	shmPath := os.Args[1]

	// Open the shared memory file
	file, err := os.OpenFile(shmPath, os.O_RDWR, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open shared memory file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, pageSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to mmap file: %v\n", err)
		os.Exit(1)
	}
	defer syscall.Munmap(data)

	// Create atomic pointers to the command byte
	cmdPtr := (*uint32)(unsafe.Pointer(&data[cmdOffset]))

	// Signal ready by writing to shared memory
	copy(data[msgOffset:], []byte("ready"))
	// Set command byte to indicate we've written
	atomic.StoreUint32(cmdPtr, 1)

	// Main loop
	for {
		// Wait for command from parent (busy wait with small sleep)
		for atomic.LoadUint32(cmdPtr) != 2 {
			time.Sleep(100 * time.Nanosecond)
		}

		// Read message
		msgBytes := data[msgOffset:msgOffset+32]
		msgLen := 0
		for i, b := range msgBytes {
			if b == 0 {
				msgLen = i
				break
			}
		}
		msg := string(msgBytes[:msgLen])

		switch msg {
		case "ping":
			// Write response
			copy(data[msgOffset:], []byte("pong"))
			// Signal response ready
			atomic.StoreUint32(cmdPtr, 1)
		case "quit":
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", msg)
			// Still signal we processed it
			atomic.StoreUint32(cmdPtr, 1)
		}
	}
}
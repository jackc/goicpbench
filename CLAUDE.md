# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go IPC Bench is a benchmarking tool for comparing the performance of various Inter-Process Communication (IPC) mechanisms in Go, particularly for plugin communication. The project is in its initial stage with no implementation code yet.

## Project Purpose

The benchmark tests three IPC mechanisms:
- stdin/stdout communication
- TCP sockets
- Memory sharing (e.g., mmap)

The test protocol is simple: the main process sends "ping" to a plugin process, which responds with "pong".

## Development Commands

Since this is an early-stage Go project with no implementation yet, standard Go commands will apply:

```bash
# Build the project (once main.go exists)
go build

# Run tests
go test ./...

# Run benchmarks
go test -bench=.

# Format code
go fmt ./...

# Vet code for common mistakes
go vet ./...
```

## Architecture Considerations

When implementing this benchmark:

1. **Process Architecture**: The main program will need to spawn child processes for plugins
2. **IPC Implementations**: Each IPC method (stdin/stdout, TCP, memory sharing) should be implemented as a separate module/package
3. **Benchmarking Framework**: Use Go's built-in `testing` package benchmark functionality
4. **Plugin Process**: Consider creating a separate executable or using Go's plugin system

## Module Information

- Module: `github.com/jackc/goipcbench`
- Go Version: 1.24.3
- No external dependencies yet (no go.sum file)
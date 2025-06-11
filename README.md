# Go IPC Bench

This is a simple benchmark of the speed of various IPC mechanisms in particular for plugins.

## Designs

The main program spawns a new process for the plugin. The communication methods to be tested are:

* stdin / stdout
* TCP
* Unix domain socket
* memory sharing such as with mmap

The message sent from the main process to the plugin process is "ping". The plugin process returns "pong". When the plugin process receives "quit", it terminates.

## AI Usage

This was also an experiment of using Claude Code to accelerate quick experiments. All code was written via Claude Code.

## Sample Results

```
jack@glados ~/dev/goipcbench ±master⚡ » go test -bench=.
goos: darwin
goarch: arm64
pkg: github.com/jackc/goipcbench
cpu: Apple M3 Max
BenchmarkMmap-16     	  510157	      2364 ns/op
BenchmarkStdio-16    	  139143	      8290 ns/op
BenchmarkTCP-16      	   63421	     16684 ns/op
BenchmarkUnix-16     	  266845	      4830 ns/op
PASS
ok  	github.com/jackc/goipcbench	14.295s
```

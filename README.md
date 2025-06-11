# Go IPC Bench

This is a simple benchmark of the speed of various IPC mechanisms in particular for plugins.

## Designs

The main program spawns a new process for the plugin. The communication methods to be tested are:

* stdin / stdout
* TCP
* Unix domain socket
* memory sharing such as with mmap

The message sent from the main process to the plugin process is "ping". The plugin process returns "pong". When the plugin process receives "quit", it terminates.

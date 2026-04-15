# Pebble Plugin System

Pebble is designed to be easily embeddable, making it an ideal choice for a plugin language in Go applications. This document outlines how to implement and use a plugin system with Pebble.

## 1. The Core Architecture

A Pebble-based plugin system consists of two parts:
- **The Host (Go)**: The application that provides the environment, executes scripts, and triggers events.
- **The Plugin (Pebble)**: A script (traditionally with a `.pb` extension) that implements custom logic.

## 2. Host Implementation (Go)

The simplest way to host Pebble is using the high-level API which provides a `State` manager (Lua-style).

### A. Initialize and Register Functions
Expose Go functions to the Pebble scripts so they can interact with your application.

```go
import "pebble"

L := pebble.New()

// Register a Go function as a builtin in Pebble
L.RegisterFunction("notify_host", func(args ...object.Object) object.Object {
    fmt.Printf("Plugin says: %s\n", args[0].Inspect())
    return object.NULL
})
```

### B. Load and Run Plugins
You can run strings or files directly. Running a file executes it globally, populating the environment with its functions and variables.

```go
// Load and execute a plugin file
_, err := L.RunFile("./plugins/my_plugin.pb")
```

### C. Trigger Hooks
When events occur in your Go application, you can call specific functions defined in the Pebble script.

```go
// Pass a Go map as an argument; it will be converted to a Pebble Hash
eventData := map[string]any{"user": "alice", "action": "login"}

result, err := L.Call("on_event", eventData)
if err != nil {
    // Function might not exist or encountered an error
}
```

## 3. Plugin Implementation (Pebble)

The Plugin script implements the logic that the host expects.

```go
// log_plugin.pb

// Global variables can be read by the host
plugin_author := "Pannaga"

// This is a "hook" called by the host
func on_event(event) {
    if event["action"] == "login" {
        notify_host("User logged in!")
        return true
    }
    return false
}
```

## 4. Running the Example Host

We provide a reference implementation in `examples/plugin_example/host_lua_style.go`.

### Build
```bash
go build -o plugin_host ./examples/plugin_example/host_lua_style.go
```

### Usage
By default, it searches for `.pb` files in `./examples/plugin_example/plugins`. You can specify a custom directory using the `-dir` flag:

```bash
./plugin_host -dir ./my_plugins
```

## 5. Why Use Pebble?

1. **Safety**: Scripts run in a sandboxed environment; they can only call functions you explicitly provide.
2. **Familiarity**: Go developers don't need to learn a new syntax.
3. **Pure Go**: No CGO dependencies, making cross-compilation and distribution seamless.

---

See the `examples/plugin_example/` directory for a full working implementation.

# Pebble Interpreter

Pebble is a simple, interpreted programming language written in Go. It features a C-like syntax, first-class functions, and a REPL.

## Features
- **Variables**: `var x = 10;`
- **Data Types**: Integers, Booleans, Strings, Functions.
- **Control Flow**: `if/else`, `while`.
- **Functions**: `fn(x) { return x + 1; }`
- **Built-ins**: `print()`, `len()`.
- **Embeddable**: Can be used as a scripting language for Go applications.

## Installation
Ensure you have Go installed.
```bash
git clone https://github.com/yourusername/pebble.git
cd pebble
```

## Usage

### USE THE EXECUTABLE (directly GO not required)
./pebble examples/demo.pb

### REPL
Start the interactive Read-Eval-Print-Loop:
```bash
go run cmd/pebble/main.go
```

### Running Scripts
Run a Pebble script file:
```bash
go run cmd/pebble/main.go examples/demo.pb
```

## Next Steps

Pebble is evolving! Here are the planned functionalities to transform it into a mid-level programming language:

### Core Language Features
- **Arrays & Hashes**: Support for ordered lists and key-value pair data structures.
- **Advanced Control Flow**: `for` loops, `switch` statements, and `break`/`continue` support.
- **Structs & Methods**: Custom data types and object-oriented patterns for better data modeling.
- **Modules & Imports**: Support for multi-file projects and code reuse.

### Advanced Capabilities
- **Concurrency**: Lightweight threads (fibers/goroutines) and channels for parallel execution.
- **FFI (Foreign Function Interface)**: Ability to call Go or C functions directly from Pebble.
- **Bytecode Compiler & VM**: Performance optimizations through compilation to bytecode and a dedicated Virtual Machine.

### Ecosystem & Tooling
- **Standard Library**: Expanded built-in functions for Networking (HTTP), JSON/YAML parsing, and Math utilities.
- **Package Manager**: A dedicated tool for managing dependencies and modules.
- **LSP Support**: Language Server Protocol implementation for IDE integration (VS Code, etc.).
- **Testing Framework**: Built-in support for unit and integration tests.
- **Improved Error Handling**: Detailed error messages with line and column information for easier debugging.


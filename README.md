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


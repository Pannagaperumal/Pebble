# Pebble Design Specification

## 1. Core Purpose & Value Proposition

**Pebble** is a high-performance, embeddable scripting language for Go applications that feels like a natural extension of Go.

### Why Pebble?
Go developers often need to embed logic that can change without recompilation (e.g., business rules, plugins, dynamic configurations). Existing solutions like Lua or Starlark force developers to switch mental models (1-based indexing, Python syntax, etc.). Pebble aims to eliminate this friction.

| Feature | Pebble | Lua | Starlark | Embedded Go (Yaegi) |
|---------|--------|-----|----------|---------------------|
| **Syntax** | Go-like | C-like/Unique | Python-like | Full Go |
| **Indexing**| 0-based | 1-based | 0-based | 0-based |
| **Safety** | Sandboxed | Sandboxed | Deterministic | Hard to Sandbox |
| **Speed** | VM-based | Very Fast (JIT) | Moderate | Interpreter overhead |
| **Interop** | Native Go | CGO/GopherLua | Native Go | Native Go |

---

## 2. Syntax Design (v1.0)

To ensure consistency and developer ergonomics for Go programmers, Pebble will follow these rules:

### 2.1 Variables
- **`var x = 10`**: Standard declaration.
- **`x := 10`**: Short declaration (Syntactic sugar for `var x = 10`).
- **NO `let`**: `let` is deprecated and will be removed to avoid confusion.

### 2.2 Functions
- **`func add(a, b) { ... }`**: Preferred function definition.
- **`add := func(a, b) { ... }`**: Anonymous function assigned to a variable.
- **`fn` is legacy**: Supported for backward compatibility but discouraged.

### 2.3 Control Flow
- **`if` / `else`**: No parentheses around conditions.
- **`for`**: Go-style loop. (Replaces `while`).
- **Parentheses**: Optional but discouraged for conditions.

### 2.4 Data Types
- `int`, `string`, `bool`, `array` (0-indexed).
- `map` (planned).

---

## 3. Practical Examples

### 3.1 Plugin System
*Use case: Transforming data in a pipeline.*
```go
// transform.pb
func transform(data) {
    if data.status == "pending" {
        data.priority := 1
    }
    return data
}
```

### 3.2 Rule Engine
*Use case: Fraud detection.*
```go
// rules.pb
func is_fraudulent(tx) {
    if tx.amount > 10000 {
        return true
    }
    if tx.location != tx.user_home {
        return true
    }
    return false
}
```

---

## 4. Comparison vs Alternatives (Honest Critique)

### vs Lua
- **Pros**: 0-based indexing (huge for Go devs), no CGO required for performance, syntax is instantly familiar.
- **Cons**: Lua is more mature, has a larger ecosystem, and the JIT (Luajit) is unbeatable in raw speed.

### vs Starlark
- **Pros**: Supports recursion and non-deterministic logic (often needed in general scripts), more familiar syntax for C-family devs.
- **Cons**: Starlark is safer by design (no recursion, deterministic), which is better for build systems (Bazel) but restrictive for general plugins.

---

## 5. Next Steps
1.  **Standardize Parser**: Fully migrate internal AST names to Go-style (e.g., `LetStatement` -> `VarStatement`).
2.  **Map Support**: Implement Go-like maps `{key: value}`.
3.  **Go Interop**: Create a high-level `pebble.Run(script)` helper that maps Go structs to Pebble values.
4.  **Error Handling**: Go-style `val, err := ...` patterns.

## 6. What NOT to Build (Yet)
- **Package Manager**: Overkill for an embedded language. Scripts should be small and self-contained.
- **Classes/OOP**: Go is composition-over-inheritance; Pebble should be too.
- **Generics**: Complexity doesn't justify the cost for scripting use cases.
- **JIT**: Focus on a fast, predictable bytecode VM first.

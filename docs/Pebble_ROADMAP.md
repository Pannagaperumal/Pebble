# Pebble Language Roadmap
---

## Vision

Pebble aims to become a **high-performance embeddable scripting language written in Go**, focused on:

- Performance-first architecture
- Clean and expressive syntax
- Strong Go embedding support
- Safe sandbox execution
- Production-grade plugin and configuration system

> A fast embeddable scripting and configuration engine for Go systems.

---

# Phase 1 – Core Performance Upgrade (Month 1)

## Goal
Replace AST tree-walk interpreter with Bytecode + Virtual Machine.

### Tasks

- [ ] Define instruction set  
- [ ] Implement constant pool  
- [ ] Design stack-based VM  
- [ ] AST → Bytecode compiler  
- [ ] Indexed local variable system  
- [ ] Tagged value system  
- [ ] Mark-and-sweep GC  
- [ ] Benchmark AST vs VM  

Success Criteria:
- 5x+ faster than AST interpreter

---

# Phase 2 – Embedding & Developer Experience (Month 2)

## Embedding API

```go
vm := pebble.NewVM()
vm.RegisterFunction("log", myLogger)
vm.ExecuteFile("config.pb")
value := vm.GetGlobal("result")
```

### Tasks

- [ ] Stable VM API  
- [ ] Register Go functions  
- [ ] Call Pebble from Go  
- [ ] Sandbox execution mode  
- [ ] Timeout and memory limits  
- [ ] REPL  
- [ ] Improved error diagnostics  

---

# Phase 3 – Modules & Standard Library (Month 3)

### Module System
- [ ] import syntax  
- [ ] Module resolution  
- [ ] Bytecode caching  
- [ ] Namespace isolation  

### Standard Library
- [ ] JSON  
- [ ] HTTP client  
- [ ] File I/O  
- [ ] Time utilities  
- [ ] Collection helpers  

### CLI
- [ ] pebble run  
- [ ] pebble fmt  
- [ ] pebble test  
- [ ] pebble build  

---

# Long-Term Vision

- Rules engine for Go services  
- Configuration DSL  
- Plugin system  
- Secure script execution engine  

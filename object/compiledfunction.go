package object

// CompiledFunction represents a compiled function in the VM
type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
}

// Type returns the type of the object
func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }

// Inspect returns a string representation of the compiled function
func (cf *CompiledFunction) Inspect() string { return "CompiledFunction" }

package pebble

import (
	"fmt"
	"io"
	"os"
	"pebble/internal/compiler"
	"pebble/internal/evaluator"
	"pebble/internal/lexer"
	"pebble/internal/parser"
	"pebble/internal/vm"
	"pebble/object"
)

// State represents the state of a Pebble interpreter instance.
// Similar to lua_State.
type State struct {
	Env *object.Environment
}

// New creates a new Pebble interpreter state with built-ins registered.
func New() *State {
	return &State{
		Env: object.NewEnvironment(),
	}
}

// Run executes the code using the Evaluator (Feature-complete, best for plugins).
func (s *State) Run(code string) (object.Object, error) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("parser errors: %v", p.Errors())
	}

	evaluator.RegisterBuiltins(s.Env)
	result := evaluator.Eval(program, s.Env)
	if result != nil && result.Type() == object.ERROR_OBJ {
		return nil, fmt.Errorf("runtime error: %s", result.Inspect())
	}

	return result, nil
}

// RunVM executes the code using the high-performance VM (Limited support for dynamic builtins).
func (s *State) RunVM(code string) (object.Object, error) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("parser errors: %v", p.Errors())
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		return nil, fmt.Errorf("compiler error: %s", err)
	}

	machine := vm.New(comp.Bytecode())
	err = machine.Run()
	if err != nil {
		return nil, fmt.Errorf("vm error: %s", err)
	}

	return s.GetLastPopped(machine), nil
}

// GetLastPopped is a helper to get the result from the VM
func (s *State) GetLastPopped(v *vm.VM) object.Object {
	return v.LastPoppedStackElem()
}

// RunFile reads and executes a Pebble script file.
func (s *State) RunFile(path string) (object.Object, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return s.Run(string(content))
}

// SetGlobal sets a global variable in the interpreter environment.
func (s *State) SetGlobal(name string, value any) error {
	obj := convertToPebbleObject(value)
	if obj == nil {
		return fmt.Errorf("unsupported type: %T", value)
	}
	s.Env.Set(name, obj)
	return nil
}

// GetGlobal retrieves a global variable from the interpreter environment.
func (s *State) GetGlobal(name string) (object.Object, bool) {
	return s.Env.Get(name)
}

// RegisterFunction registers a Go function as a Pebble builtin.
func (s *State) RegisterFunction(name string, fn func(args ...object.Object) object.Object) {
	builtin := &object.Builtin{Fn: fn}

	// Add to environment for Evaluator
	s.Env.Set(name, builtin)

	// Also add to the global Builtins for VM (Workaround until VM is decentralized)
	// We check if it's already there to avoid duplicates
	found := false
	for _, b := range object.Builtins {
		if b.Name == name {
			found = true
			break
		}
	}

	if !found {
		object.Builtins = append(object.Builtins, struct {
			Name    string
			Builtin *object.Builtin
		}{Name: name, Builtin: builtin})
	}
}

// Call calls a Pebble function defined in the state.
func (s *State) Call(funcName string, args ...any) (object.Object, error) {
	fnObj, ok := s.Env.Get(funcName)
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	pebbleArgs := make([]object.Object, len(args))
	for i, arg := range args {
		pebbleArgs[i] = convertToPebbleObject(arg)
	}

	result := evaluator.ApplyFunction(fnObj, pebbleArgs)
	if result.Type() == object.ERROR_OBJ {
		return nil, fmt.Errorf("runtime error: %s", result.Inspect())
	}

	return result, nil
}

// convertToPebbleObject converts standard Go types to Pebble objects.
func convertToPebbleObject(val any) object.Object {
	switch v := val.(type) {
	case object.Object:
		return v
	case int:
		return &object.Integer{Value: int64(v)}
	case int64:
		return &object.Integer{Value: v}
	case float64:
		return &object.Float{Value: v}
	case string:
		return &object.String{Value: v}
	case bool:
		if v {
			return object.TRUE
		}
		return object.FALSE
	case nil:
		return object.NULL
	case map[string]any:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, val := range v {
			key := &object.String{Value: k}
			pVal := convertToPebbleObject(val)
			pairs[key.HashKey()] = object.HashPair{Key: key, Value: pVal}
		}
		return &object.Hash{Pairs: pairs}
	case []any:
		elements := make([]object.Object, len(v))
		for i, val := range v {
			elements[i] = convertToPebbleObject(val)
		}
		return &object.Array{Elements: elements}
	default:
		return nil
	}
}

// RunREPL starts an interactive REPL session.
func RunREPL(in io.Reader, out io.Writer) {
	// Simple wrapper for existing REPL
	// This would need to be implemented or imported from package/repl
}

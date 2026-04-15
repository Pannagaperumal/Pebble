package moxy

import (
	"fmt"
	"io"
	"os"
	"github.com/pannagaperumal/moxy/internal/compiler"
	"github.com/pannagaperumal/moxy/internal/evaluator"
	"github.com/pannagaperumal/moxy/internal/lexer"
	"github.com/pannagaperumal/moxy/internal/parser"
	"github.com/pannagaperumal/moxy/internal/vm"
	"github.com/pannagaperumal/moxy/types"
)

// State represents the state of a Moxy interpreter instance.
// Similar to lua_State.
type State struct {
	Env *types.Environment
}

// New creates a new Moxy interpreter state with built-ins registered.
func New() *State {
	return &State{
		Env: types.NewEnvironment(),
	}
}

// Run executes the code using the Evaluator (Feature-complete, best for plugins).
func (s *State) Run(code string) (types.Object, error) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return nil, fmt.Errorf("parser errors: %v", p.Errors())
	}

	evaluator.RegisterBuiltins(s.Env)
	result := evaluator.Eval(program, s.Env)
	if result != nil && result.Type() == types.ERROR_OBJ {
		return nil, fmt.Errorf("runtime error: %s", result.Inspect())
	}

	return result, nil
}

// RunVM executes the code using the high-performance VM (Limited support for dynamic builtins).
func (s *State) RunVM(code string) (types.Object, error) {
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
func (s *State) GetLastPopped(v *vm.VM) types.Object {
	return v.LastPoppedStackElem()
}

// RunFile reads and executes a Moxy script file.
func (s *State) RunFile(path string) (types.Object, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return s.Run(string(content))
}

// SetGlobal sets a global variable in the interpreter environment.
func (s *State) SetGlobal(name string, value any) error {
	obj := convertToMoxyObject(value)
	if obj == nil {
		return fmt.Errorf("unsupported type: %T", value)
	}
	s.Env.Set(name, obj)
	return nil
}

// GetGlobal retrieves a global variable from the interpreter environment.
func (s *State) GetGlobal(name string) (types.Object, bool) {
	return s.Env.Get(name)
}

// RegisterFunction registers a Go function as a Moxy builtin.
func (s *State) RegisterFunction(name string, fn func(args ...types.Object) types.Object) {
	builtin := &types.Builtin{Fn: fn}

	// Add to environment for Evaluator
	s.Env.Set(name, builtin)

	// Also add to the global Builtins for VM (Workaround until VM is decentralized)
	// We check if it's already there to avoid duplicates
	found := false
	for _, b := range types.Builtins {
		if b.Name == name {
			found = true
			break
		}
	}

	if !found {
		types.Builtins = append(types.Builtins, struct {
			Name    string
			Builtin *types.Builtin
		}{Name: name, Builtin: builtin})
	}
}

// Call calls a Moxy function defined in the state.
func (s *State) Call(funcName string, args ...any) (types.Object, error) {
	fnObj, ok := s.Env.Get(funcName)
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	pebbleArgs := make([]types.Object, len(args))
	for i, arg := range args {
		pebbleArgs[i] = convertToMoxyObject(arg)
	}

	result := evaluator.ApplyFunction(fnObj, pebbleArgs)
	if result.Type() == types.ERROR_OBJ {
		return nil, fmt.Errorf("runtime error: %s", result.Inspect())
	}

	return result, nil
}

// convertToMoxyObject converts standard Go types to Moxy objects.
func convertToMoxyObject(val any) types.Object {
	switch v := val.(type) {
	case types.Object:
		return v
	case int:
		return &types.Integer{Value: int64(v)}
	case int64:
		return &types.Integer{Value: v}
	case float64:
		return &types.Float{Value: v}
	case string:
		return &types.String{Value: v}
	case bool:
		if v {
			return types.TRUE
		}
		return types.FALSE
	case nil:
		return types.NULL
	case map[string]any:
		pairs := make(map[types.HashKey]types.HashPair)
		for k, val := range v {
			key := &types.String{Value: k}
			pVal := convertToMoxyObject(val)
			pairs[key.HashKey()] = types.HashPair{Key: key, Value: pVal}
		}
		return &types.Hash{Pairs: pairs}
	case []any:
		elements := make([]types.Object, len(v))
		for i, val := range v {
			elements[i] = convertToMoxyObject(val)
		}
		return &types.Array{Elements: elements}
	default:
		return nil
	}
}

// RunREPL starts an interactive REPL session.
func RunREPL(in io.Reader, out io.Writer) {
	// Simple wrapper for existing REPL
	// This would need to be implemented or imported from package/repl
}

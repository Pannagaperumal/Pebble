package compiler

import (
	"fmt"

	"pebble/package/ast"
	"pebble/package/code"
	"pebble/package/object"
	"pebble/package/symbol"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	scope               CompilationScope
	symbolTable         *SymbolTable
	scopes              []CompilationScope
	scopeIndex          int
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	// Initialize built-in functions (commented out as object.Builtins is not defined)
	// for i, v := range object.Builtins {
	// 	symbolTable.DefineBuiltin(i, v.Name)
	// }

	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		scopes:       []CompilationScope{mainScope},
		scopeIndex:   0,
		symbolTable:  symbolTable,
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	var instructions code.Instructions
	if len(c.scopes) > 0 {
		instructions = c.scopes[c.scopeIndex].instructions
	}
	return &Bytecode{
		Instructions: instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.InfixExpression:
		if node.Operator == "=" {
			err := c.compileAssignment(node)
			if err != nil {
				return err
			}
		} else {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			err = c.Compile(node.Right)
			if err != nil {
				return err
			}

			switch node.Operator {
			case "+":
				c.emit(code.OpAdd)
			case "-":
				c.emit(code.OpSub)
			case "*":
				c.emit(code.OpMul)
			case "/":
				c.emit(code.OpDiv)
			case "%":
				c.emit(code.OpMod)
			case "==":
				c.emit(code.OpEqual)
			case "!=":
				c.emit(code.OpNotEqual)
			case ">":
				c.emit(code.OpGreaterThan)
			case "<":
				c.emit(code.OpLessThan)
			case ">=":
				c.emit(code.OpGreaterOrEqual)
			case "<=":
				c.emit(code.OpLessOrEqual)
			default:
				return fmt.Errorf("unknown operator %s", node.Operator)
			}
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.scope.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.scope.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		sym := c.symbolTable.Define(node.Name.Value)

		if sym.Scope == symbol.GlobalScope {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, elem := range node.Elements {
			err := c.Compile(elem)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		// Get free symbols and number of locals before leaving scope
		freeSymbols := []symbol.Symbol{}
		numLocals := 0
		if c.symbolTable != nil {
			freeSymbols = c.symbolTable.FreeSymbols
			// Get the number of local variables defined in this scope
			numLocals = c.symbolTable.NumDefinitions()
		}

		instructions := c.leaveScope()

		// Load all free variables
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		// Create compiled function
		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		// Add the compiled function to constants and emit closure
		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	case *ast.WhileExpression:
		// Save the position to jump back to
		loopStart := len(c.scope.instructions)

		// Compile the condition
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit a conditional jump that exits the loop if the condition is false
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		// Compile the loop body
		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Add an unconditional jump back to the condition
		c.emit(code.OpJump, loopStart)

		// Update the jump position to after the loop
		afterLoopPos := len(c.scope.instructions)
		c.changeOperand(jumpNotTruthyPos, afterLoopPos)

		// Emit a null value as the result of the while expression
		c.emit(code.OpNull)
	}

	return nil
}

func (c *Compiler) compileAssignment(node *ast.InfixExpression) error {
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return fmt.Errorf("left-hand side of assignment must be an identifier")
	}

	err := c.Compile(node.Right)
	if err != nil {
		return err
	}

	symbol, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("undefined variable %s", ident.Value)
	}

	if symbol.Scope == symbol.Scope{
		c.emit(code.OpSetGlobal, symbol.Index)
	} else {
		c.emit(code.OpSetLocal, symbol.Index)
	}

	return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.scope.instructions)
	updatedInstructions := append(c.scope.instructions, ins...)

	c.scope.instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scope.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scope.previousInstruction = previous
	c.scope.lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.scope.instructions) == 0 {
		return false
	}

	return c.scope.lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	c.scope.instructions = c.scope.instructions[:c.scope.lastInstruction.Position]
	c.scope.lastInstruction = c.scope.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.scope.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.scope.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex = len(c.scopes) - 1

	// Create new symbol table with outer scope
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.scopes[c.scopeIndex].instructions
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	// Restore outer symbol table
	if c.symbolTable.Outer != nil {
		c.symbolTable = c.symbolTable.Outer
	}

	return instructions
}

func (c *Compiler) loadSymbol(s symbol.Symbol) {
	switch s.Scope {
	case symbol.GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case symbol.LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case symbol.BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case symbol.FreeScope:
		c.emit(code.OpGetFree, s.Index)
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scope.lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scope.lastInstruction.Opcode = code.OpReturnValue
}

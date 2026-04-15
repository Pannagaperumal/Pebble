package compiler

import (
	"fmt"

	"pebble/ast"
	"pebble/internal/code"
	"pebble/object"
	"pebble/internal/symbol"
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

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

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
		return c.compileProgram(node)
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(node)
	case *ast.InfixExpression:
		return c.compileInfixExpression(node)
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
		return c.compilePrefixExpression(node)
	case *ast.IfExpression:
		return c.compileIfExpression(node)
	case *ast.BlockStatement:
		return c.compileBlockStatement(node)
	case *ast.VarStatement:
		return c.compileVarStatement(node)
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
		return c.compileArrayLiteral(node)
	case *ast.HashLiteral:
		return c.compileHashLiteral(node)
	case *ast.IndexExpression:
		return c.compileIndexExpression(node)
	case *ast.FunctionLiteral:
		return c.compileFunctionLiteral(node)
	case *ast.ReturnStatement:
		return c.compileReturnStatement(node)
	case *ast.CallExpression:
		return c.compileCallExpression(node)
	case *ast.ForStatement:
		return c.compileForStatement(node)
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

	sym, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("undefined variable %s", ident.Value)
	}

	if sym.Scope == symbol.GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
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
	posNewInstruction := len(c.scopes[c.scopeIndex].instructions)
	updatedInstructions := append(c.scopes[c.scopeIndex].instructions, ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.scopes[c.scopeIndex].instructions) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	c.scopes[c.scopeIndex].instructions = c.scopes[c.scopeIndex].instructions[:c.scopes[c.scopeIndex].lastInstruction.Position]
	c.scopes[c.scopeIndex].lastInstruction = c.scopes[c.scopeIndex].previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.scopes[c.scopeIndex].instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.scopes[c.scopeIndex].instructions[opPos])
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
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

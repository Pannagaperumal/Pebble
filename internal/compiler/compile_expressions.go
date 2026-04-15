package compiler

import (
	"fmt"
	"pebble/ast"
	"pebble/internal/code"
	"pebble/object"
	"pebble/internal/symbol"
)

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) error {
	if node.Operator == "=" {
		return c.compileAssignment(node)
	}

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
	return nil
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpression) error {
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
	return nil
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
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

	afterConsequencePos := len(c.scopes[c.scopeIndex].instructions)
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

	afterAlternativePos := len(c.scopes[c.scopeIndex].instructions)
	c.changeOperand(jumpPos, afterAlternativePos)
	return nil
}

func (c *Compiler) compileArrayLiteral(node *ast.ArrayLiteral) error {
	for _, elem := range node.Elements {
		err := c.Compile(elem)
		if err != nil {
			return err
		}
	}

	c.emit(code.OpArray, len(node.Elements))
	return nil
}

func (c *Compiler) compileHashLiteral(node *ast.HashLiteral) error {
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
	return nil
}

func (c *Compiler) compileIndexExpression(node *ast.IndexExpression) error {
	err := c.Compile(node.Left)
	if err != nil {
		return err
	}

	err = c.Compile(node.Index)
	if err != nil {
		return err
	}

	c.emit(code.OpIndex)
	return nil
}

func (c *Compiler) compileFunctionLiteral(node *ast.FunctionLiteral) error {
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
	return nil
}

func (c *Compiler) compileCallExpression(node *ast.CallExpression) error {
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
	return nil
}

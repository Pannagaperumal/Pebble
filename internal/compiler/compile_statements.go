package compiler

import (
	"pebble/ast"
	"pebble/internal/code"
	"pebble/internal/symbol"
)

func (c *Compiler) compileProgram(node *ast.Program) error {
	for _, s := range node.Statements {
		err := c.Compile(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileExpressionStatement(node *ast.ExpressionStatement) error {
	err := c.Compile(node.Expression)
	if err != nil {
		return err
	}

	// If the last instruction was an assignment (SetGlobal/SetLocal),
	// it already popped the value, so we don't need another OpPop.
	if c.lastInstructionIs(code.OpSetGlobal) || c.lastInstructionIs(code.OpSetLocal) {
		return nil
	}

	c.emit(code.OpPop)
	return nil
}

func (c *Compiler) compileBlockStatement(node *ast.BlockStatement) error {
	for _, s := range node.Statements {
		err := c.Compile(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileVarStatement(node *ast.VarStatement) error {
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
	return nil
}

func (c *Compiler) compileReturnStatement(node *ast.ReturnStatement) error {
	err := c.Compile(node.ReturnValue)
	if err != nil {
		return err
	}

	c.emit(code.OpReturnValue)
	return nil
}

func (c *Compiler) compileForStatement(node *ast.ForStatement) error {
	if node.Init != nil {
		err := c.Compile(node.Init)
		if err != nil {
			return err
		}
	}

	loopStart := len(c.scopes[c.scopeIndex].instructions)

	var jumpNotTruthyPos int = -1
	if node.Condition != nil {
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotTruthyPos = c.emit(code.OpJumpNotTruthy, 9999)
	}

	err := c.Compile(node.Body)
	if err != nil {
		return err
	}

	if node.Post != nil {
		err := c.Compile(node.Post)
		if err != nil {
			return err
		}
	}

	c.emit(code.OpJump, loopStart)

	afterLoopPos := len(c.scopes[c.scopeIndex].instructions)
	if jumpNotTruthyPos != -1 {
		c.changeOperand(jumpNotTruthyPos, afterLoopPos)
	}

	c.emit(code.OpNull)
	return nil
}

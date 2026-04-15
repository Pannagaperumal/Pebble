package parser

import (
	"pebble/ast"
	"pebble/internal/token"
)

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET, token.VAR: // Support both let and var for variable declarations
		return p.parseVarStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.FUNCTION:
		if p.peekTokenIs(token.IDENT) {
			return p.parseNamedFunctionStatement()
		}
		return p.parseExpressionStatement()
	default:
		// Check for short declaration: IDENT := EXPRESSION
		if p.curToken.Type == token.IDENT && p.peekToken.Type == token.DECLARE_ASSIGN {
			return p.parseShortDeclareStatement()
		}
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseShortDeclareStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken} // We reuse VarStatement
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // move to :=
	stmt.Token = p.curToken

	p.nextToken() // move to expression
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	p.nextToken()

	// Parse the first part. It could be an init statement or just a condition.
	firstPart := p.parseStatement()

	if p.curTokenIs(token.SEMICOLON) {
		// It's 'for init; ...'
		stmt.Init = firstPart
		p.nextToken()

		if !p.curTokenIs(token.SEMICOLON) {
			stmt.Condition = p.parseExpression(LOWEST)
			p.nextToken() // move to semicolon
		}

		if !p.curTokenIs(token.SEMICOLON) {
			p.peekError(token.SEMICOLON)
			return nil
		}
		p.nextToken() // move past semicolon

		if !p.peekTokenIs(token.LBRACE) {
			stmt.Post = p.parseStatement()
		}
	} else {
		// It's 'for condition {'
		if es, ok := firstPart.(*ast.ExpressionStatement); ok {
			stmt.Condition = es.Expression
		} else {
			p.errors = append(p.errors, "expected condition in for loop")
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseNamedFunctionStatement() ast.Statement {
	// Current token is 'func' or 'fn'
	p.nextToken() // move to identifier

	stmt := &ast.VarStatement{Token: token.Token{Type: token.LET, Literal: "let"}}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIs(token.LPAREN) {
		return nil
	}

	// We move back to have 'func' as curToken for parseFunctionLiteral to work if needed?
	// Actually, let's just parse the FunctionLiteral manually or reuse it.
	
	p.nextToken() // move to '('
	function := &ast.FunctionLiteral{Token: token.Token{Type: token.FUNCTION, Literal: "func"}}
	function.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	function.Body = p.parseBlockStatement()
	stmt.Value = function

	return stmt
}

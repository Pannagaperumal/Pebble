package ast

import (
	"bytes"
	"pebble/internal/token"
)

type ForStatement struct {
	Token     token.Token // the 'for' token
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for ")
	if fs.Init != nil {
		out.WriteString(fs.Init.String() + " ")
	}
	if fs.Condition != nil {
		out.WriteString(fs.Condition.String() + " ")
	}
	if fs.Post != nil {
		out.WriteString("; " + fs.Post.String())
	}
	out.WriteString(fs.Body.String())

	return out.String()
}

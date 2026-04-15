package evaluator

import (
	"fmt"
	"github.com/pannagaperumal/moxy/ast"
	"github.com/pannagaperumal/moxy/types"
)

var (
	NULL  = &types.Null{}
	TRUE  = &types.Boolean{Value: true}
	FALSE = &types.Boolean{Value: false}
)

func Eval(node ast.Node, env *types.Environment) types.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &types.ReturnValue{Value: val}

	case *ast.VarStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.ForStatement:
		return evalForStatement(node, env)

	// Expressions
	case *ast.IntegerLiteral:
		return &types.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &types.Float{Value: node.Value}

	case *ast.StringLiteral:
		return &types.String{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		if node.Operator == "=" {
			return evalAssignmentExpression(node, env)
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &types.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return ApplyFunction(function, args)

	case *ast.WhileExpression:
		return evalWhileExpression(node, env)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &types.Array{Elements: elements}

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}
	return nil
}

func ApplyFunction(fn types.Object, args []types.Object) types.Object {
	switch fn := fn.(type) {
	case *types.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *types.Builtin:
		return fn.Fn(args...)

	case *types.Closure:
		// To call a VM-compiled function from the evaluator, we need to bridge it.
		// For now, we return an error or implement a quick VM run.
		return &types.Error{Message: fmt.Sprintf("cannot call VM closure from evaluator. Use State.Run() instead.")}

	default:
		return newError("not a function: %s", fn.Type())
	}
}

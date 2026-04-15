package evaluator

import (
	"github.com/pannagaperumal/moxy/ast"
	"github.com/pannagaperumal/moxy/types"
)

func evalProgram(program *ast.Program, env *types.Environment) types.Object {
	var result types.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *types.ReturnValue:
			return result.Value
		case *types.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *types.Environment) types.Object {
	var result types.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == types.RETURN_VALUE_OBJ || rt == types.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalWhileExpression(we *ast.WhileExpression, env *types.Environment) types.Object {
	var result types.Object = NULL

	for {
		condition := Eval(we.Condition, env)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			break
		}
		result = Eval(we.Body, env)
		if isError(result) {
			return result
		}
		if result != nil && result.Type() == types.RETURN_VALUE_OBJ {
			return result
		}
	}
	return result
}

func evalForStatement(fs *ast.ForStatement, env *types.Environment) types.Object {
	// Create a new environment for the for loop if it has an init statement
	var evaluationEnv *types.Environment
	if fs.Init != nil {
		evaluationEnv = types.NewEnclosedEnvironment(env)
		Eval(fs.Init, evaluationEnv)
	} else {
		evaluationEnv = env
	}

	var result types.Object = NULL

	for {
		if fs.Condition != nil {
			condition := Eval(fs.Condition, evaluationEnv)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		result = Eval(fs.Body, evaluationEnv)
		if isError(result) {
			return result
		}
		if result != nil && result.Type() == types.RETURN_VALUE_OBJ {
			return result
		}

		if fs.Post != nil {
			Eval(fs.Post, evaluationEnv)
		}
	}

	return result
}

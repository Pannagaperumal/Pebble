package evaluator

import (
	"pebble/ast"
	"pebble/object"
)

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalWhileExpression(we *ast.WhileExpression, env *object.Environment) object.Object {
	var result object.Object = NULL

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
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}
	return result
}

func evalForStatement(fs *ast.ForStatement, env *object.Environment) object.Object {
	// Create a new environment for the for loop if it has an init statement
	var evaluationEnv *object.Environment
	if fs.Init != nil {
		evaluationEnv = object.NewEnclosedEnvironment(env)
		Eval(fs.Init, evaluationEnv)
	} else {
		evaluationEnv = env
	}

	var result object.Object = NULL

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
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}

		if fs.Post != nil {
			Eval(fs.Post, evaluationEnv)
		}
	}

	return result
}

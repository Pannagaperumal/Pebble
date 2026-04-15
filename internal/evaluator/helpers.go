package evaluator

import (
	"fmt"

	"github.com/pannagaperumal/moxy/ast"
	"github.com/pannagaperumal/moxy/types"
)

func nativeBoolToBooleanObject(input bool) *types.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj types.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *types.Error {
	return &types.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj types.Object) bool {
	if obj != nil {
		return obj.Type() == types.ERROR_OBJ
	}
	return false
}

func evalExpressions(exps []ast.Expression, env *types.Environment) []types.Object {
	var result []types.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []types.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func unwrapReturnValue(obj types.Object) types.Object {
	if returnValue, ok := obj.(*types.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func extendFunctionEnv(fn *types.Function, args []types.Object) *types.Environment {
	env := types.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

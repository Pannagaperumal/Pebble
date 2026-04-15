package evaluator

import (
	"github.com/pannagaperumal/moxy/ast"
	"github.com/pannagaperumal/moxy/types"
)

func evalPrefixExpression(operator string, right types.Object) types.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right types.Object) types.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right types.Object) types.Object {
	if right.Type() != types.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*types.Integer).Value
	return &types.Integer{Value: -value}
}

func evalInfixExpression(operator string, left, right types.Object) types.Object {
	switch {
	case left.Type() == types.INTEGER_OBJ && right.Type() == types.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == types.FLOAT_OBJ && right.Type() == types.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == types.FLOAT_OBJ && right.Type() == types.INTEGER_OBJ:
		rightFloat := &types.Float{Value: float64(right.(*types.Integer).Value)}
		return evalFloatInfixExpression(operator, left, rightFloat)
	case left.Type() == types.INTEGER_OBJ && right.Type() == types.FLOAT_OBJ:
		leftFloat := &types.Float{Value: float64(left.(*types.Integer).Value)}
		return evalFloatInfixExpression(operator, leftFloat, right)
	case left.Type() == types.STRING_OBJ && right.Type() == types.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right types.Object) types.Object {
	leftVal := left.(*types.Integer).Value
	rightVal := right.(*types.Integer).Value

	switch operator {
	case "+":
		return &types.Integer{Value: leftVal + rightVal}
	case "-":
		return &types.Integer{Value: leftVal - rightVal}
	case "*":
		return &types.Integer{Value: leftVal * rightVal}
	case "/":
		return &types.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right types.Object) types.Object {
	leftVal := left.(*types.String).Value
	rightVal := right.(*types.String).Value

	switch operator {
	case "+":
		return &types.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalAssignmentExpression(node *ast.InfixExpression, env *types.Environment) types.Object {
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		return newError("left side of assignment must be an identifier")
	}

	val := Eval(node.Right, env)
	if isError(val) {
		return val
	}

	_, ok = env.Update(ident.Value, val)
	if !ok {
		return newError("identifier not found: %s", ident.Value)
	}

	return val
}

func evalIfExpression(ie *ast.IfExpression, env *types.Environment) types.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalIdentifier(node *ast.Identifier, env *types.Environment) types.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	return newError("identifier not found: " + node.Value)
}

func evalHashLiteral(node *ast.HashLiteral, env *types.Environment) types.Object {
	pairs := make(map[types.HashKey]types.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(types.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = types.HashPair{Key: key, Value: value}
	}

	return &types.Hash{Pairs: pairs}
}

func evalIndexExpression(left, index types.Object) types.Object {
	switch {
	case left.Type() == types.ARRAY_OBJ && index.Type() == types.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == types.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index types.Object) types.Object {
	arrayObject := array.(*types.Array)
	idx := index.(*types.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashIndexExpression(hash, index types.Object) types.Object {
	hashObject := hash.(*types.Hash)

	key, ok := index.(types.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalFloatInfixExpression(operator string, left, right types.Object) types.Object {
	leftVal := left.(*types.Float).Value
	rightVal := right.(*types.Float).Value

	switch operator {
	case "+":
		return &types.Float{Value: leftVal + rightVal}
	case "-":
		return &types.Float{Value: leftVal - rightVal}
	case "*":
		return &types.Float{Value: leftVal * rightVal}
	case "/":
		return &types.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

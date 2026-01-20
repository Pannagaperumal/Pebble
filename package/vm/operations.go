package vm

import (
	"fmt"
	"pebble/package/object"
)

func (vm *VM) executeBinaryOperation(op Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case OpAdd:
		result = leftVal + rightVal
	case OpSub:
		result = leftVal - rightVal
	case OpMul:
		result = leftVal * rightVal
	case OpDiv:
		if rightVal == 0 {
			return fmt.Errorf("division by zero")
		}
		result = leftVal / rightVal
	case OpMod:
		if rightVal == 0 {
			return fmt.Errorf("modulo by zero")
		}
		result = leftVal % rightVal
	case OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	case OpLessThan:
		return vm.push(nativeBoolToBooleanObject(leftVal < rightVal))
	case OpGreaterOrEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal >= rightVal))
	case OpLessOrEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal <= rightVal))
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op Opcode, left, right object.Object) error {
	if op != OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case object.TRUE:
		return vm.push(object.FALSE)
	case object.FALSE, object.NULL:
		return vm.push(object.TRUE)
	default:
		return vm.push(object.FALSE)
	}
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return vm.push(object.NULL)
	}

	return vm.push(arrayObject.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(object.NULL)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.CompiledFunction:
		return vm.callFunction(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function: %s", callee.Type())
	}
}

func (vm *VM) callFunction(fn *object.CompiledFunction, numArgs int) error {
	if numArgs != fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(fn, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + fn.NumLocals

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(object.NULL)
	}

	return nil
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

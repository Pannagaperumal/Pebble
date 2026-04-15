package vm

import (
	"encoding/binary"
	"fmt"

	"github.com/pannagaperumal/moxy/types"
)

func (vm *VM) executeBinaryOperation(op Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == types.INTEGER_OBJ && rightType == types.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == types.STRING_OBJ && rightType == types.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op Opcode, left, right types.Object) error {
	leftVal := left.(*types.Integer).Value
	rightVal := right.(*types.Integer).Value

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

	return vm.push(&types.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op Opcode, left, right types.Object) error {
	if op != OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*types.String).Value
	rightVal := right.(*types.String).Value

	return vm.push(&types.String{Value: leftVal + rightVal})
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != types.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*types.Integer).Value
	return vm.push(&types.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case types.TRUE:
		return vm.push(types.FALSE)
	case types.FALSE, types.NULL:
		return vm.push(types.TRUE)
	default:
		return vm.push(types.FALSE)
	}
}

func (vm *VM) executeIndexExpression(left, index types.Object) error {
	switch {
	case left.Type() == types.ARRAY_OBJ && index.Type() == types.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == types.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index types.Object) error {
	arrayObject := array.(*types.Array)
	idx := index.(*types.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return vm.push(types.NULL)
	}

	return vm.push(arrayObject.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index types.Object) error {
	hashObject := hash.(*types.Hash)

	key, ok := index.(types.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(types.NULL)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *types.Closure:
		return vm.callFunction(callee, numArgs)
	case *types.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function: %s", callee.Type())
	}
}

func (vm *VM) callFunction(cl *types.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (vm *VM) executeArrayLiteral() error {
	numElements := int(binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:]))
	vm.currentFrame().ip += 2

	array := make([]types.Object, numElements)
	for i := numElements - 1; i >= 0; i-- {
		array[i] = vm.pop()
	}

	return vm.push(&types.Array{Elements: array})
}

func (vm *VM) executeHashLiteral() error {
	numPairs := int(binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:]))
	vm.currentFrame().ip += 2

	hash := make(map[types.HashKey]types.HashPair)

	for i := 0; i < numPairs; i++ {
		value := vm.pop()
		key := vm.pop()

		pair := types.HashPair{Key: key, Value: value}

		hashKey, ok := key.(types.Hashable)
		if !ok {
			return fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hash[hashKey.HashKey()] = pair
	}

	return vm.push(&types.Hash{Pairs: hash})
}

func (vm *VM) callBuiltin(builtin *types.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(types.NULL)
	}

	return nil
}

func nativeBoolToBooleanObject(input bool) *types.Boolean {
	if input {
		return types.TRUE
	}
	return types.FALSE
}

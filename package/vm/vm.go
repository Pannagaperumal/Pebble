package vm

import (
	"encoding/binary"
	"errors"
	"fmt"

	"pebble/package/code"
	"pebble/package/compiler"
	"pebble/package/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var (
	ErrStackOverflow   = errors.New("stack overflow")
	ErrUndefinedGlobal = errors.New("undefined global variable")
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames     []*Frame
	frameIndex int
}

type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),

		frames:     frames,
		frameIndex: 1,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) Run() error {
	for vm.currentFrame().ip < len(vm.currentFrame().fn.Instructions)-1 {
		vm.currentFrame().ip++

		top := vm.currentFrame().fn.Instructions[vm.currentFrame().ip]

		// Lookup the opcode to validate it (even though we don't use the definition yet)
		_, err := Lookup(top)
		if err != nil {
			return err
		}

		switch Opcode(top) {
		case OpConstant:
			constIndex := binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpEqual, OpNotEqual, OpGreaterThan, OpLessThan, OpGreaterOrEqual, OpLessOrEqual:
			err := vm.executeBinaryOperation(Opcode(top))
			if err != nil {
				return err
			}

		case OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case OpTrue:
			err := vm.push(object.TRUE)
			if err != nil {
				return err
			}

		case OpFalse:
			err := vm.push(object.FALSE)
			if err != nil {
				return err
			}

		case OpNull:
			err := vm.push(object.NULL)
			if err != nil {
				return err
			}

		case OpJumpNotTruthy:
			pos := int(binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case OpJump:
			pos := int(binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip = pos - 1

		case OpSetGlobal:
			globalIndex := binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case OpGetGlobal:
			globalIndex := binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case OpSetLocal:
			localIndex := int(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		case OpGetLocal:
			localIndex := int(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++

			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+localIndex])
			if err != nil {
				return err
			}

		case OpArray:
			numElements := int(binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			array := make([]object.Object, numElements)
			for i := numElements - 1; i >= 0; i-- {
				array[i] = vm.pop()
			}

			err := vm.push(&object.Array{Elements: array})
			if err != nil {
				return err
			}

		case OpHash:
			numPairs := int(binary.BigEndian.Uint16(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			hash := make(map[object.HashKey]object.HashPair)

			for i := 0; i < numPairs; i++ {
				value := vm.pop()
				key := vm.pop()

				pair := object.HashPair{Key: key, Value: value}

				hashKey, ok := key.(object.Hashable)
				if !ok {
					return fmt.Errorf("unusable as hash key: %s", key.Type())
				}

				hash[hashKey.HashKey()] = pair
			}

			err := vm.push(&object.Hash{Pairs: hash})
			if err != nil {
				return err
			}

		case OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case OpCall:
			numArgs := int(vm.currentFrame().fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}

		case OpReturnValue:
			returnValue := vm.pop()
			vm.popFrame()
			vm.pop() // Pop function from stack
			vm.push(returnValue)

		case OpReturn:
			vm.popFrame()
			vm.pop() // Pop function from stack
			vm.push(object.NULL)

		case OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return ErrStackOverflow
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

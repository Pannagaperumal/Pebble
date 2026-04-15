package vm

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/pannagaperumal/moxy/internal/code"
	"github.com/pannagaperumal/moxy/internal/compiler"
	"github.com/pannagaperumal/moxy/types"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var (
	ErrStackOverflow   = errors.New("stack overflow")
	ErrUndefinedGlobal = errors.New("undefined global variable")
)

type VM struct {
	constants    []types.Object
	instructions code.Instructions

	stack []types.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []types.Object

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &types.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &types.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]types.Object, StackSize),
		sp:    0,

		globals: make([]types.Object, GlobalsSize),

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
	for vm.currentFrame().ip < len(vm.currentFrame().cl.Fn.Instructions)-1 {
		vm.currentFrame().ip++

		top := vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip]

		_, err := Lookup(top)
		if err != nil {
			return err
		}

		op := Opcode(top)
		switch op {
		case OpConstant:
			constIndex := binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpEqual, OpNotEqual, OpGreaterThan, OpLessThan, OpGreaterOrEqual, OpLessOrEqual:
			err := vm.executeBinaryOperation(op)
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
			vm.push(types.TRUE)
		case OpFalse:
			vm.push(types.FALSE)
		case OpNull:
			vm.push(types.NULL)

		case OpJumpNotTruthy:
			pos := int(binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case OpJump:
			pos := int(binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip = pos - 1

		case OpSetGlobal:
			globalIndex := binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case OpGetGlobal:
			globalIndex := binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2
			vm.push(vm.globals[globalIndex])

		case OpSetLocal:
			localIndex := int(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		case OpGetLocal:
			localIndex := int(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++
			frame := vm.currentFrame()
			vm.push(vm.stack[frame.basePointer+localIndex])

		case OpArray:
			err := vm.executeArrayLiteral()
			if err != nil {
				return err
			}

		case OpHash:
			err := vm.executeHashLiteral()
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
			numArgs := int(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1])
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
			vm.push(types.NULL)

		case OpGetBuiltin:
			builtinIndex := vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1]
			vm.currentFrame().ip++
			definition := types.Builtins[builtinIndex]
			vm.push(definition.Builtin)

		case OpGetFree:
			freeIndex := int(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1])
			vm.currentFrame().ip++
			vm.push(vm.currentFrame().cl.FreeVariables[freeIndex])

		case OpClosure:
			constIndex := binary.BigEndian.Uint16(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+1:])
			numFree := int(vm.currentFrame().cl.Fn.Instructions[vm.currentFrame().ip+3])
			vm.currentFrame().ip += 3
			err := vm.pushClosure(int(constIndex), numFree)
			if err != nil {
				return err
			}

		case OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*types.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]types.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &types.Closure{Fn: function, FreeVariables: free}
	return vm.push(closure)
}

func isTruthy(obj types.Object) bool {
	switch obj := obj.(type) {
	case *types.Boolean:
		return obj.Value
	case *types.Null:
		return false
	default:
		return true
	}
}

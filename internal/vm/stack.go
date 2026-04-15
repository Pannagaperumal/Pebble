package vm

import "github.com/pannagaperumal/moxy/types"

func (vm *VM) push(o types.Object) error {
	if vm.sp >= StackSize {
		return ErrStackOverflow
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() types.Object {
	if vm.sp == 0 {
		return types.NULL
	}
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() types.Object {
	return vm.stack[vm.sp]
}

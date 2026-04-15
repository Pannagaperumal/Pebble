package vm

import "github.com/pannagaperumal/moxy/types"

type Frame struct {
	cl          *types.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *types.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
	}
}

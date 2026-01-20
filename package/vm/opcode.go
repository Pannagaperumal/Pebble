package vm

import "fmt"

type Opcode byte

const (
	OpConstant       Opcode = iota // Load constant
	OpAdd                          // Add two numbers
	OpSub                          // Subtract two numbers
	OpMul                          // Multiply two numbers
	OpDiv                          // Divide two numbers
	OpMod                          // Modulo operation
	OpEqual                        // Equality comparison
	OpNotEqual                     // Inequality comparison
	OpGreaterThan                  // Greater than comparison
	OpLessThan                     // Less than comparison
	OpGreaterOrEqual               // Greater than or equal comparison
	OpLessOrEqual                  // Less than or equal comparison
	OpMinus                        // Unary minus
	OpBang                         // Logical NOT
	OpTrue                         // Push true
	OpFalse                        // Push false
	OpNull                         // Push null
	OpJumpNotTruthy                // Conditional jump if false
	OpJump                         // Unconditional jump
	OpSetGlobal                    // Set global variable
	OpGetGlobal                    // Get global variable
	OpSetLocal                     // Set local variable
	OpGetLocal                     // Get local variable
	OpArray                        // Create array
	OpHash                         // Create hash
	OpIndex                        // Index operation
	OpCall                         // Function call
	OpReturnValue                  // Return with value
	OpReturn                       // Return (implicit nil)
	OpPop                          // Pop value from stack
)

type Definition struct {
	Name          string
	OperandWidths []int // Number of bytes each operand takes
}

var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}}, // 2 bytes for constant index (65536 max constants)
	OpAdd:            {"OpAdd", []int{}},
	OpSub:            {"OpSub", []int{}},
	OpMul:            {"OpMul", []int{}},
	OpDiv:            {"OpDiv", []int{}},
	OpMod:            {"OpMod", []int{}},
	OpEqual:          {"OpEqual", []int{}},
	OpNotEqual:       {"OpNotEqual", []int{}},
	OpGreaterThan:    {"OpGreaterThan", []int{}},
	OpLessThan:       {"OpLessThan", []int{}},
	OpGreaterOrEqual: {"OpGreaterOrEqual", []int{}},
	OpLessOrEqual:    {"OpLessOrEqual", []int{}},
	OpMinus:          {"OpMinus", []int{}},
	OpBang:           {"OpBang", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpNull:           {"OpNull", []int{}},
	OpJumpNotTruthy:  {"OpJumpNotTruthy", []int{2}},
	OpJump:           {"OpJump", []int{2}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpSetLocal:       {"OpSetLocal", []int{1}}, // 1 byte for local index (256 max locals)
	OpGetLocal:       {"OpGetLocal", []int{1}},
	OpArray:          {"OpArray", []int{2}}, // 2 bytes for element count
	OpHash:           {"OpHash", []int{2}},  // 2 bytes for pair count * 2
	OpIndex:          {"OpIndex", []int{}},
	OpCall:           {"OpCall", []int{1}}, // 1 byte for argument count
	OpReturnValue:    {"OpReturnValue", []int{}},
	OpReturn:         {"OpReturn", []int{}},
	OpPop:            {"OpPop", []int{}},
}

// Lookup returns the definition for the given opcode
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

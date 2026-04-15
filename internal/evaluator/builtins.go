package evaluator

import (
	"fmt"

	"github.com/pannagaperumal/moxy/types"
)

var Builtins = map[string]*types.Builtin{
	"print": {
		Fn: func(args ...types.Object) types.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"str": {
		Fn: func(args ...types.Object) types.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			return &types.String{Value: args[0].Inspect()}
		},
	},
	"len": {
		Fn: func(args ...types.Object) types.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			switch arg := args[0].(type) {
			case *types.String:
				return &types.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
}

func RegisterBuiltins(env *types.Environment) {
	for name, builtin := range Builtins {
		env.Set(name, builtin)
	}
}

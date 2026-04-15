package object

type Closure struct {
	Fn            *CompiledFunction
	FreeVariables []Object
}

func (c *Closure) Type() ObjectType { return "CLOSURE" }
func (c *Closure) Inspect() string  { return "Closure" }

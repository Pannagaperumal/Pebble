package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"pebble/package/evaluator"
	"pebble/package/lexer"
	"pebble/package/object"
	"pebble/package/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	evaluator.RegisterBuiltins(env)

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func RunFile(filename string, out io.Writer) {
	content, err := io.ReadAll(openFile(filename))
	if err != nil {
		fmt.Fprintf(out, "Error reading file: %s\n", err)
		return
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(out, p.Errors())
		return
	}

	env := object.NewEnvironment()
	evaluator.RegisterBuiltins(env)

	// Evaluate the program
	evaluated := evaluator.Eval(program, env)

	// Check for evaluation errors
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		fmt.Fprintf(out, "Runtime error: %s\n", evaluated.Inspect())
	}

	// Note: We don't print the result of the last expression for script files
	// as it's the standard behavior to only show explicit print statements
}

func openFile(filename string) io.Reader {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return f
}

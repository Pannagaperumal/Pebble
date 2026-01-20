package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"pebble/package/lexer"
	"pebble/package/parser"
	"pebble/package/repl"
)

var (
	verbose = flag.Bool("v", false, "Enable verbose output")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) > 0 {
		filename := args[0]
		if *verbose {
			runVerbose(filename)
		} else {
			repl.RunFile(filename, os.Stdout)
		}
		return
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Pebble programming language!\n",
		user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}

func runVerbose(filename string) {
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	// Show the input
	fmt.Println("=== Input ===")
	fmt.Println(string(content))
	fmt.Println("\n=== Lexer Output ===")

	// Lex the input
	l := lexer.New(string(content))
	for {
		tok := l.NextToken()
		fmt.Printf("%+v\n", tok)
		if tok.Type == "EOF" {
			break
		}
	}

	// Parse the input
	fmt.Println("\n=== Parser Output ===")
	l = lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	// Print parser errors if any
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
	}

	// Print the parsed program
	fmt.Println("\n=== AST ===")
	fmt.Println(program.String())
}

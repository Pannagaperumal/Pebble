package main

import (
	"fmt"
	"os"
	"os/user"
	"pebble/package/repl"
)

func main() {
	if len(os.Args) > 1 {
		repl.RunFile(os.Args[1], os.Stdout)
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

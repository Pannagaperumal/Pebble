package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"pebble" // Using the high-level API
	"pebble/object"
)

func main() {
	pluginsDir := flag.String("dir", "./examples/plugin_example/plugins", "Directory to search for .pb plugins")
	flag.Parse()

	fmt.Printf("=== Pebble Lua-style Plugin Host (Directory: %s) ===\n", *pluginsDir)

	// 1. Discover plugins
	files, err := ioutil.ReadDir(*pluginsDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".pb" {
			path := filepath.Join(*pluginsDir, file.Name())

			// 2. Load and run like Lua
			L := pebble.New()

			// Register host functions
			L.RegisterFunction("notify_host", func(args ...object.Object) object.Object {
				if len(args) > 0 {
					fmt.Printf("  [Host Notification]: %s\n", args[0].Inspect())
				}
				return object.NULL
			})

			// Run the script
			_, err := L.RunFile(path)
			if err != nil {
				fmt.Printf("Error running %s: %v\n", path, err)
				continue
			}

			// 3. Call a hook if it exists
			fmt.Printf("Triggering 'on_event' in %s...\n", file.Name())
			event := map[string]any{
				"message": "This is a very long message that should trigger the notification",
			}

			result, err := L.Call("on_event", event)
			if err != nil {
				fmt.Printf("  Failed to call on_event: %v\n", err)
			} else {
				fmt.Printf("  Result: %s\n", result.Inspect())
			}
		}
	}
}

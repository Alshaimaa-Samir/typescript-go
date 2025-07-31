package main

import (
	"os"
)

func main() {
	os.Exit(runMain())
}

var SupportedExtensions = map[string]bool{".ts": true, ".tsx": true, ".d.ts": true, ".js": true, ".jsx": true}

func runMain() int {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--parse":
			return runParse(args[1:])
		case "--resolve":
			return runResolve(args[1:])
		}
	}
	return 1
}

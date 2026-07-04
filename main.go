package main

import (
	"lem-in/internal/parser"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, input file missing")
		os.Exit(1)
	}

	filePath := os.Args[1]
	graph, err := parser.Parse(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}

	// paths, err := graph.FindPaths(graph)
	// if err != nil {
	//
	// }
}

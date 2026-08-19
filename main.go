package main

import (
	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/simulation"

	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, input file missing")
		os.Exit(1)
	}

	filePath := os.Args[1]

	raw, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}

	g, err := parser.Parse(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}

	paths, err := graph.FindPaths(g)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}

	turns, err := simulation.Simulate(paths, g.NumAnts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}

	fmt.Print(strings.TrimRight(string(raw), "\n") + "\n")
	fmt.Println()
	fmt.Println(strings.Join(turns, "\n"))
}

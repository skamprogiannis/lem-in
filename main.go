package main

import (
	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/simulation"

	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format,", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("input file missing")
	}

	filePath := args[0]
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	g, err := parser.Parse(filePath)
	if err != nil {
		return err
	}

	paths, err := graph.FindPaths(g)
	if err != nil {
		return err
	}

	turns, err := simulation.Simulate(paths, g.NumAnts)
	if err != nil {
		return err
	}

	if _, err = stdout.Write(raw); err != nil {
		return err
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		if _, err = fmt.Fprintln(stdout); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintln(stdout); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, strings.Join(turns, "\n"))
	return err
}

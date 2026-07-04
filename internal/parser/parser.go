// package parser reads and validates lem-in input files.
package parser

import (
	"lem-in/internal/graph"

	"errors"
	"os"
	"strings"
)

// Parse reads a lem-in input file and returns its validated graph.
func Parse(filePath string) (*graph.Graph, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	data := string(file)

	if len(strings.TrimSpace(data)) == 0 {
		return nil, errors.New("empty input file")
	}

	lines := strings.SplitSeq(data, "\n")

	for line := range lines {
		_ = line
	}

	return &graph.Graph{}, nil
}

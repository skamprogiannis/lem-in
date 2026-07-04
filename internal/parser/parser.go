// package parser reads and validates lem-in input files.
package parser

import (
	"lem-in/internal/graph"

	"errors"
	"os"
	"strconv"
	"strings"
)

// Parse reads a lem-in input file and returns its validated graph.
func Parse(filePath string) (*graph.Graph, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	data := string(file)
	normalizedData := strings.TrimSpace(strings.ReplaceAll(data, "\r\n", "\n"))

	if len(normalizedData) == 0 {
		return nil, errors.New("empty input file")
	}

	lines := strings.SplitSeq(normalizedData, "\n")
	g := &graph.Graph{}
	seenStart, seenEnd := false, false

	for line := range lines {
		tokens := strings.Fields(strings.TrimSpace(line))
		if len(tokens) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if seenStart {
					return nil, errors.New("more than one start location")
				}
				seenStart = true
			case "##end":
				if seenEnd {
					return nil, errors.New("more than one end location")
				}
				seenEnd = true
			}
			continue
		}

		if g.Start == "" && seenStart {
			g.Start = tokens[0]
		}

		if g.End == "" && seenEnd {
			g.End = tokens[0]
		}

		if g.Start == "" && g.NumAnts == 0 {
			g.NumAnts, err = strconv.Atoi(line)
			if err != nil {
				return nil, errors.New("number of ants not provided")
			}
		}
	}

	if g.NumAnts < 1 {
		return nil, errors.New("need at least one ant")
	}
	if !seenStart || !seenEnd {
		return nil, errors.New("did not find both a seenStart and an end")
	}

	return g, nil
}

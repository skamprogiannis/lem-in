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
	var pending *string

	for line := range lines {
		line = strings.TrimSpace(line)
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if seenStart || pending != nil {
					return nil, errors.New("more than one start location")
				}
				seenStart = true
				pending = &g.Start
			case "##end":
				if seenEnd || pending != nil {
					return nil, errors.New("more than one end location")
				}
				seenEnd = true
				pending = &g.End
			}
			continue
		}

		if pending != nil {
			if len(tokens) != 3 {
				return nil, errors.New("start or end not followed by a room")
			}
			*pending = tokens[0]
			pending = nil
			continue
		}

		if g.NumAnts == 0 {
			if len(tokens) != 1 {
				return nil, errors.New("number of ants not provided")
			}
			g.NumAnts, err = strconv.Atoi(tokens[0])
			if err != nil {
				return nil, errors.New("number of ants not provided")
			}
			continue
		}
	}

	if g.NumAnts < 1 {
		return nil, errors.New("need at least one ant")
	}
	if g.Start == "" || g.End == "" {
		return nil, errors.New("did not find both a start and an end")
	}

	return g, nil
}

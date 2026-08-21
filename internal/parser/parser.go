// Package parser reads and validates lem-in input files.
package parser

import (
	"lem-in/internal/graph"

	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Parse reads a lem-in input file and returns its validated graph.
func Parse(filePath string) (*graph.Graph, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("input file not found")
	}
	data := string(file)
	normalizedData := strings.ReplaceAll(data, "\r\n", "\n")

	if len(strings.TrimSpace(normalizedData)) == 0 {
		return nil, errors.New("empty input file")
	}

	lines := strings.Split(normalizedData, "\n")
	seenStart, seenEnd, parsingLinks := false, false, false
	var pending *string
	g := &graph.Graph{
		Rooms: make(map[string]*graph.Room),
		Links: make(map[string][]string),
	}

	for lineNumber, line := range lines {
		line = strings.TrimSpace(line)
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") {
			switch line {
			case "##start":
				if seenStart || pending != nil {
					return nil, errors.New("multiple start rooms are not allowed")
				}
				seenStart = true
				pending = &g.Start
			case "##end":
				if seenEnd || pending != nil {
					return nil, errors.New("multiple end rooms are not allowed")
				}
				seenEnd = true
				pending = &g.End
			}
			continue
		}

		if g.NumAnts == 0 {
			if len(tokens) != 1 {
				return nil, errors.New("invalid number in ant count entry")
			}
			g.NumAnts, err = strconv.Atoi(tokens[0])
			if err != nil {
				return nil, errors.New("invalid number in ant count entry")
			}
			continue
		}

		if len(tokens) == 1 && strings.Contains(line, "-") {
			parsingLinks = true
			roomNames := strings.Split(line, "-")

			if len(roomNames) != 2 || roomNames[0] == "" || roomNames[1] == "" {
				return nil, fmt.Errorf("line %d: malformed tunnel entry", lineNumber)
			}

			if roomNames[0] == roomNames[1] {
				return nil, fmt.Errorf("line %d: room cannot link to itself", lineNumber)
			}

			if g.Rooms[roomNames[0]] == nil || g.Rooms[roomNames[1]] == nil {
				return nil,
					fmt.Errorf("line %d: room referenced by tunnel does not exist", lineNumber)
			}

			if slices.Contains(g.Links[roomNames[0]], roomNames[1]) {
				return nil, fmt.Errorf("line %d: duplicate tunnel", lineNumber)
			}

			g.Links[roomNames[0]] = append(g.Links[roomNames[0]], roomNames[1])
			g.Links[roomNames[1]] = append(g.Links[roomNames[1]], roomNames[0])
			continue
		}

		if parsingLinks {
			return nil, fmt.Errorf("line %d: room declared after tunnels", lineNumber)
		}

		room, err := parseRoom(tokens, lineNumber)
		if err != nil {
			return nil, err
		}

		if _, ok := g.Rooms[room.Name]; ok {
			return nil, fmt.Errorf("line %d: room %s already exists", lineNumber, room.Name)
		}
		g.Rooms[room.Name] = room

		if pending != nil {
			*pending = room.Name
			pending = nil
		}
	}

	// Failure conditions checked after the whole file has been read.
	if g.NumAnts < 1 {
		return nil, errors.New("invalid ant value provided")
	}
	if g.Start == "" {
		return nil, errors.New("start room entry missing")
	}
	if g.End == "" {
		return nil, errors.New("end room entry missing")
	}

	return g, nil
}

func parseRoom(tokens []string, lineNumber int) (*graph.Room, error) {
	if len(tokens) != 3 {
		return nil, fmt.Errorf("line %d: room lines must be of format roomName x y", lineNumber)
	}

	if strings.HasPrefix(tokens[0], "L") {
		return nil, fmt.Errorf("line %d: room names must not start with L", lineNumber)
	}

	x, err := strconv.Atoi(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("line %d: x value is malformed", lineNumber)
	}

	y, err := strconv.Atoi(tokens[2])
	if err != nil {
		return nil, fmt.Errorf("line %d: y value is malformed", lineNumber)
	}

	return &graph.Room{Name: tokens[0], X: x, Y: y}, nil
}

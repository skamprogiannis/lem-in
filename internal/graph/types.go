// Package graph defines the lem-in ant farm model.
package graph

// Room describes one room in the ant farm.
type Room struct {
	Name string
	X, Y int
}

// Graph stores a validated ant farm.
type Graph struct {
	Rooms   map[string]*Room
	Links   map[string][]string
	NumAnts int
	Start   string
	End     string
}

// Path is an ordered list of room names running from Start to End.
type Path []string

// package graph defiens the lem-in ant farm model.
package graph

// Room describes on room in the ant-farm.
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

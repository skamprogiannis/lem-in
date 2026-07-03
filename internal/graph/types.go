package graph

type Room struct {
	Name string
	X, Y int
}

type Graph struct {
	Rooms   map[string]*Room
	Links   map[string][]string
	NumAnts int
	Start   string
	End     string
}

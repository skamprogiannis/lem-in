package graph

import (
	"math/rand"
	"testing"
)

func TestAuditAgainstExhaustiveSmallFarms(t *testing.T) {
	const rooms = 7
	rng := rand.New(rand.NewSource(20260902))

	for sample := 0; sample < 3000; sample++ {
		g := randomAuditFarm(rooms, rng)
		all := enumerateAuditPaths(g, g.Start, nil, map[string]bool{})
		if len(all) == 0 {
			continue
		}

		for _, ants := range []int{1, 2, 3, 5, 8, 13} {
			g.NumAnts = ants
			got, err := FindPaths(g)
			if err != nil {
				t.Fatalf("sample %d, ants %d: %v", sample, ants, err)
			}
			gotTurns := auditCapacityTurns(got, ants)
			wantTurns := exhaustiveAuditTurns(all, ants)
			if gotTurns != wantTurns {
				t.Fatalf("sample %d, ants %d: got %d turns with %v, want %d; links=%v",
					sample, ants, gotTurns, got, wantTurns, g.Links)
			}
		}
	}
}

func randomAuditFarm(roomCount int, rng *rand.Rand) *Graph {
	names := make([]string, roomCount)
	for i := range names {
		names[i] = string(rune('a' + i))
	}

	g := &Graph{
		Rooms: make(map[string]*Room, roomCount),
		Links: make(map[string][]string, roomCount),
		Start: names[0],
		End:   names[len(names)-1],
	}
	for _, name := range names {
		g.Rooms[name] = &Room{Name: name}
	}
	for a := 0; a < roomCount; a++ {
		for b := a + 1; b < roomCount; b++ {
			if rng.Intn(100) >= 38 {
				continue
			}
			g.Links[names[a]] = append(g.Links[names[a]], names[b])
			g.Links[names[b]] = append(g.Links[names[b]], names[a])
		}
	}
	return g
}

func enumerateAuditPaths(g *Graph, room string, path Path, visited map[string]bool) []Path {
	path = append(path, room)
	if room == g.End {
		return []Path{append(Path(nil), path...)}
	}

	visited[room] = true
	defer delete(visited, room)

	var paths []Path
	for _, next := range g.Links[room] {
		if !visited[next] {
			paths = append(paths, enumerateAuditPaths(g, next, path, visited)...)
		}
	}
	return paths
}

func exhaustiveAuditTurns(paths []Path, ants int) int {
	best := 0
	var choose func(int, []Path, map[string]bool)
	choose = func(at int, selected []Path, occupied map[string]bool) {
		if len(selected) > 0 {
			turns := auditCapacityTurns(selected, ants)
			if best == 0 || turns < best {
				best = turns
			}
		}
		for i := at; i < len(paths); i++ {
			path := paths[i]
			compatible := true
			for _, room := range path[1 : len(path)-1] {
				if occupied[room] {
					compatible = false
					break
				}
			}
			if !compatible {
				continue
			}
			for _, room := range path[1 : len(path)-1] {
				occupied[room] = true
			}
			choose(i+1, append(selected, path), occupied)
			for _, room := range path[1 : len(path)-1] {
				delete(occupied, room)
			}
		}
	}
	choose(0, nil, make(map[string]bool))
	return best
}

func auditCapacityTurns(paths []Path, ants int) int {
	for turns := 1; ; turns++ {
		capacity := 0
		for _, path := range paths {
			if delivered := turns - (len(path) - 1) + 1; delivered > 0 {
				capacity += delivered
			}
		}
		if capacity >= ants {
			return turns
		}
	}
}

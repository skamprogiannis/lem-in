package graph

import (
	"errors"
	"sort"
)

// FindPaths returns the paths from Start to End that move g.NumAnts ants in the
// fewest turns. It expects a validated graph. No two paths share a room in the
// middle, since a room holds one ant at a time.
func FindPaths(g *Graph) ([]Path, error) {
	net := newNetwork(g)

	// More paths is not automatically better: one long path can raise the turn
	// count. So score every size and keep the winner.
	var best []Path
	bestTurns := 0

	for net.augment() {
		candidate := net.decompose(g.Start)

		cost := turnsNeeded(candidate, g.NumAnts)
		if len(best) == 0 || cost < bestTurns {
			best, bestTurns = candidate, cost
		}
	}

	if len(best) == 0 {
		return nil, errors.New("no path between start and end")
	}
	sort.SliceStable(best, func(i, j int) bool {
		return tunnelCount(best[i]) < tunnelCount(best[j])
	})
	return best, nil
}

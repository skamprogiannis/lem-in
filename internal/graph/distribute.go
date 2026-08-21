package graph

// AssignAnts hands ants 1..numAnts out one at a time, each to the path where it
// would arrive soonest, and returns every path's ants in the order they queue
// on it. It is the same greedy rule FindPaths already uses (via antsPerPath) to
// score candidate path sets, so Simulate can reuse it to learn which ant rides
// which path instead of re-deriving the distribution.
func AssignAnts(paths []Path, numAnts int) [][]int {
	queues := make([][]int, len(paths))
	if len(paths) == 0 {
		return queues
	}

	for ant := 1; ant <= numAnts; ant++ {
		choice := 0
		soonest := arrivalTurn(paths[0], len(queues[0]))

		for p := 1; p < len(paths); p++ {
			if arrival := arrivalTurn(paths[p], len(queues[p])); arrival < soonest {
				choice, soonest = p, arrival
			}
		}
		queues[choice] = append(queues[choice], ant)
	}
	return queues
}

// tunnelCount is how many moves a path takes: one less than its room count.
func tunnelCount(p Path) int { return len(p) - 1 }

// arrivalTurn is the turn an ant reaches the end of a path, given how many ants
// are queued ahead of it: one turn per tunnel, plus one turn per ant waited on.
func arrivalTurn(p Path, queuedAhead int) int {
	return tunnelCount(p) + queuedAhead
}

// antsPerPath reports how many ants AssignAnts put on each path.
func antsPerPath(paths []Path, ants int) []int {
	queues := AssignAnts(paths, ants)
	counts := make([]int, len(queues))
	for p, q := range queues {
		counts[p] = len(q)
	}
	return counts
}

// turnsNeeded reports how many turns a set of paths takes for a number of ants.
// The paths run side by side, so the slowest one decides when the farm is empty.
func turnsNeeded(paths []Path, ants int) int {
	if len(paths) == 0 {
		return 0 // nothing to walk, and antsPerPath would index paths[0]
	}

	slowest := 0
	for p, count := range antsPerPath(paths, ants) {
		if count == 0 {
			continue // a path nobody was sent down costs nothing
		}
		slowest = max(slowest, arrivalTurn(paths[p], count-1))
	}
	return slowest
}

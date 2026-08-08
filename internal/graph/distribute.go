package graph

// tunnelCount is how many moves a path takes: one less than its room count.
func tunnelCount(p Path) int { return len(p) - 1 }

// arrivalTurn is the turn an ant reaches the end of a path, given how many ants
// are queued ahead of it: one turn per tunnel, plus one turn per ant waited on.
func arrivalTurn(p Path, queuedAhead int) int {
	return tunnelCount(p) + queuedAhead
}

// antsPerPath hands the ants out one at a time, each to the path where it would
// arrive soonest, and reports how many ended up on each path.
func antsPerPath(paths []Path, ants int) []int {
	carried := make([]int, len(paths))

	for i := 0; i < ants; i++ {
		choice := 0
		soonest := arrivalTurn(paths[0], carried[0])

		for p := 1; p < len(paths); p++ {
			if arrival := arrivalTurn(paths[p], carried[p]); arrival < soonest {
				choice, soonest = p, arrival
			}
		}
		carried[choice]++
	}
	return carried
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

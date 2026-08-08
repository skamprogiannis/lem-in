package graph

// tunnelCount is how many moves a path takes: one less than its room count.
func tunnelCount(p Path) int { return len(p) - 1 }

// antsPerPath hands the ants out one at a time, each to the path where it would
// arrive soonest, and reports how many ended up on each path.
//
// A path of L tunnels already carrying c ants delivers the next ant on turn L+c,
// since that ant waits for the c queued ahead of it. Smallest L+c wins.
func antsPerPath(paths []Path, ants int) []int {
	carried := make([]int, len(paths))

	for i := 0; i < ants; i++ {
		choice := 0
		soonest := tunnelCount(paths[0]) + carried[0]

		for p := 1; p < len(paths); p++ {
			if arrival := tunnelCount(paths[p]) + carried[p]; arrival < soonest {
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
		if finish := tunnelCount(paths[p]) + count - 1; finish > slowest {
			slowest = finish
		}
	}
	return slowest
}

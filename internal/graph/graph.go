package graph

import "errors"

// FindPaths returns the paths from Start to End that move g.NumAnts ants in the
// fewest turns. No two paths share a room in the middle, since a room holds one
// ant at a time.
func FindPaths(g *Graph) ([]Path, error) {
	if err := validateInput(g); err != nil {
		return nil, err
	}

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
	return best, nil
}

// validateInput rejects graphs the search cannot make sense of.
func validateInput(g *Graph) error {
	switch {
	case g == nil:
		return errors.New("no graph provided")
	case g.Start == "":
		return errors.New("start room entry missing")
	case g.End == "":
		return errors.New("end room entry missing")
	case g.Start == g.End:
		return errors.New("start and end must be different rooms")
	case g.NumAnts < 1:
		return errors.New("invalid ant value provided")
	}
	return nil
}

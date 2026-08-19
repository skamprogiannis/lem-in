// package simulation turns a set of disjoint paths and an ant count into the
// turn-by-turn moves that carry every ant from start to end.
package simulation

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"lem-in/internal/graph"
)

// move is one ant entering one room on some turn.
type move struct {
	ant  int
	room string
}

// Simulate distributes numAnts ants over paths and returns one line per turn,
// each listing every ant that moved that turn as "Lant-room" in ascending ant
// order, ants separated by single spaces.
//
// FindPaths guarantees the paths share no intermediate room, so each path acts
// as an independent single-lane corridor: ant i (0-indexed by queue position on
// its path) advances exactly one room every turn, always one turn behind the
// ant ahead of it, and enters the path's last room on turn i+tunnels — the same
// arrival turn graph.FindPaths already scores candidate path sets by. That
// closed form is what this function evaluates for every ant, rather than
// tracking room occupancy turn by turn: it can never place two ants in the same
// room or send two ants down the same tunnel on the same turn, because the
// paths it was given are disjoint and every ant on a path is exactly one turn
// apart from its neighbours.
func Simulate(paths []graph.Path, numAnts int) ([]string, error) {
	if len(paths) == 0 {
		return nil, errors.New("no paths to simulate")
	}
	if numAnts < 1 {
		return nil, errors.New("invalid ant value provided")
	}

	queues := graph.AssignAnts(paths, numAnts)
	totalTurns := lastArrival(paths, queues)

	turnMoves := make([][]move, totalTurns)
	for p, queue := range queues {
		path := paths[p]
		tunnels := len(path) - 1

		for i, ant := range queue {
			for turn := i + 1; turn <= i+tunnels; turn++ {
				room := path[turn-i]
				turnMoves[turn-1] = append(turnMoves[turn-1], move{ant, room})
			}
		}
	}

	return formatTurns(turnMoves), nil
}

// lastArrival is the turn the slowest path's last ant reaches the end: one turn
// per tunnel, plus one turn per ant queued ahead of it on that path.
func lastArrival(paths []graph.Path, queues [][]int) int {
	last := 0
	for p, queue := range queues {
		if len(queue) == 0 {
			continue // nobody sent down this path, so it sets no deadline
		}
		if arrival := len(paths[p]) - 1 + len(queue) - 1; arrival > last {
			last = arrival
		}
	}
	return last
}

// formatTurns renders each turn's moves as "Lant-room", ascending by ant
// number, space-separated.
func formatTurns(turnMoves [][]move) []string {
	lines := make([]string, len(turnMoves))

	for turn, moves := range turnMoves {
		sort.Slice(moves, func(a, b int) bool { return moves[a].ant < moves[b].ant })

		parts := make([]string, len(moves))
		for i, m := range moves {
			parts[i] = fmt.Sprintf("L%d-%s", m.ant, m.room)
		}
		lines[turn] = strings.Join(parts, " ")
	}
	return lines
}

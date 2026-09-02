# Lem-in

Lem-in is a deterministic Go command-line program that routes an ant colony
through a graph in as few turns as possible. It parses and validates a farm,
finds useful room-disjoint paths, balances ants across them, and prints a legal
turn-by-turn movement transcript.

The project was built at Zone01 Athens using only the Go standard library.

## Highlights

- Models the one-ant-per-room rule with a split-node flow network.
- Uses residual breadth-first searches to discover room-disjoint routes.
- Scores each successive route set for the given ant count instead of assuming
  that more paths are always faster.
- Produces stable output by ordering graph construction and ant movements.
- Validates the official examples at the CLI boundary, including room and
  tunnel occupancy invariants.
- Returns contextual errors without panicking or printing partial results.

## Quick start

Requirements: Go 1.26 or newer.

```bash
go run . examples/example00.txt
```

Build a reusable binary:

```bash
go build -o lem-in .
./lem-in examples/example05.txt
```

The program accepts exactly one farm file. Invalid input is reported on
standard error with the required prefix:

```text
ERROR: invalid data format, <reason>
```

## Input and output

A farm declares a positive ant count, rooms, one start room, one end room, and
undirected tunnels:

```text
4
##start
0 0 3
2 2 5
3 4 0
##end
1 8 3
0-2
2-3
3-1
```

The CLI prints the original farm unchanged, one blank line, and one movement
line per turn. `L2-3`, for example, means that ant 2 entered room 3.

```text
L1-2
L1-3 L2-2
L1-1 L2-3 L3-2
L2-1 L3-3 L4-2
L3-1 L4-3
L4-1
```

Start and end can hold any number of ants. Every other room can hold one ant,
each ant moves at most once per turn, and each tunnel is used at most once per
turn.

## How it works

1. `internal/parser` reads the file and validates ants, endpoint commands,
   rooms, coordinates, and tunnels.
2. `internal/graph` splits each room into entry and exit nodes connected by a
   unit-capacity edge. This turns room occupancy into a flow constraint.
3. Repeated residual BFS passes produce candidate sets of room-disjoint paths.
   Each set is scored against the actual ant count, and the fastest set is kept.
4. `internal/simulation` balances ants by projected arrival time and generates
   the movement lines without mutable global state.
5. `main.go` preserves the input bytes and writes the complete result.

The optional graphical visualizer bonus is not included; this repository
focuses on the pathfinding and simulation requirements.

## Verification

```bash
go test ./...
go test -race ./...
go vet ./...
test -z "$(gofmt -l .)"
```

The integration suite uses the official Zone01 examples and checks:

- examples 00–05 meet their required turn ceilings;
- the two official invalid examples fail cleanly;
- every move follows a real tunnel;
- ants never collide or move twice in one turn;
- a tunnel is never reused in the same turn;
- every ant reaches the end; and
- repeated runs produce identical output.

On the latest local compiled-binary audit (Linux amd64, Go 1.26.5), the 100-ant
and 1,000-ant examples completed in approximately 0.002 s and 0.004 s. Their
official limits are 90 s and 150 s respectively; exact timings vary by machine.

## Team and contributions

- **Stefanos Kamprogiannis (`skamprogiannis`)** — parser, CLI and error-path
  integration, graph refinements, official audit coverage, and portfolio
  documentation.
- **George Tzimokas (`gtzimoka`)** — split-room flow network, path search,
  distribution arithmetic, algorithm tests, and initial examples.
- **`ebimai`** — ant assignment interface, simulation engine and tests, and
  initial CLI integration.
- **Daniel Tymoshenko (`dtymoshe`)** — repository setup and the original
  requirements and architecture plan.

The commit history preserves the original authorship of each contribution.

## Project context

This implementation follows the official
[01-edu Lem-in subject](https://github.com/01-edu/public/tree/master/subjects/lem-in)
and its functional audit. `PRD.md` preserves the team’s original implementation
plan and algorithm discussion as historical design documentation.

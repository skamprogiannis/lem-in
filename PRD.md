# lem-in — Product Requirements Document

> **Historical design document.** This was the team's implementation plan.
> The finished architecture, commands, verification results, and contribution
> record are documented in `README.md`.

**Project:** `lem-in` (digital ant farm / pathfinding)
**Language:** Go (standard library only)
**Team:** 4 people
**Goal of this doc:** give every team member a complete understanding of what we are building, why, and exactly how to do their part.

---

## Table of contents

1. What this project is
2. The problem in plain language
3. Goals and non-goals
4. Glossary
5. Input format (full spec)
6. Output format (full spec)
7. The algorithm (the heart of the project)
8. Architecture, shared types, and contracts
9. Git workflow for the team
10. Milestones and issues (detailed)
11. Acceptance criteria (the audit, as a checklist)
12. Testing strategy
13. Definition of done
14. Suggested timeline

---

## 1. What this project is

`lem-in` is a command-line program that simulates an **ant farm**. We are given a file that describes:

- a number of ants,
- a set of **rooms**,
- a set of **tunnels** (links) connecting rooms,
- one special **start** room and one special **end** room.

All the ants begin in the start room. Our job is to move **every ant to the end room in as few turns as possible**, then print each turn's movements.

We run it like this:

```
go run . example00.txt
```

It reads the file, computes the fastest way to move the ants, and prints the colony followed by the moves.

---

## 2. The problem in plain language

Think of the rooms as small waiting areas and the tunnels as **one-lane corridors** between them. The rules:

- A room holds **only one ant at a time** — except start and end, which hold as many as needed.
- A tunnel can be used **only once per turn**.
- On each turn, an ant may move **at most once**, and only into a room that is **empty** at the end of that turn.

If we only ever used the single shortest path, the ants would form a queue and bump into each other — a traffic jam. So the real challenge is **not** "find the shortest path." It is:

> Move many ants through a network with one-ant-per-room corridors, in the fewest total turns.

The fastest solution usually sends ants down **several different paths at once**, and those paths must not overlap (they cannot share an intermediate room, or two ants would collide there). We also have to decide **how many ants to send down each path** so they all finish as early as possible.

That is the entire puzzle. Everything in this document supports solving it correctly and quickly.

---

## 3. Goals and non-goals

**Goals**

- Correctly move all ants from start to end with the **minimum number of turns**.
- Handle every malformed input with the message `ERROR: invalid data format` and **never crash, panic, or hang**.
- Be fast: 100 ants in under 1.5 minutes, 1000 ants in under 2.5 minutes (on the audit's example06 / example07).
- Use **only Go standard packages**.
- Follow good Go practices (gofmt clean, `go vet` clean, sensible package layout).
- Include unit tests and integration tests.

**Non-goals (unless we do the bonus)**

- A graphical or 3D visualizer.
- More specific error strings (e.g. `, no start room found`) — nice to have, not required.

The bonus items earn extra credit but are not needed to pass.

---

## 4. Glossary

| Term | Meaning |
|---|---|
| Room | A node in the graph. Defined as `name x y`. Holds one ant (except start/end). |
| Tunnel / link | An undirected edge between two rooms. Defined as `name1-name2`. |
| Ant | A unit to move from start to end. Numbered `1..N`. |
| Turn | One simulation step. Multiple ants can move in a single turn. |
| Move | One ant entering one room this turn, printed as `Lx-y`. |
| Start / end | The `##start` and `##end` rooms. |
| Path | An ordered list of rooms from start to end. |
| Disjoint paths | Paths that share no **intermediate** rooms (they may all touch start and end, but nothing in between). |

---

## 5. Input format (full spec)

The file is plain text, read line by line. Order matters.

### 5.1 Number of ants

The **first meaningful line** is the ant count: a single positive integer.

- Reject `0`, negatives, and anything that is not a number.

### 5.2 Rooms

A room line looks like:

```
name coord_x coord_y
```

Examples: `Room 1 2`, `nameoftheroom 1 6`, `4 6 7`.

Rules:

- The name **must not start with `L` or `#`** and **must not contain spaces**.
- `coord_x` and `coord_y` are **integers** (coordinates are only used by the bonus visualizer, but they must still be valid).
- **No duplicate room names.**

### 5.3 Commands

Lines beginning with `#` are commands or comments:

- `##start` — the **next** room line is the start room.
- `##end` — the **next** room line is the end room.
- Any other line starting with `#` (e.g. `#comment`) is a comment and is ignored.
- Any **unknown** command is ignored (do not error on it).

There must be **exactly one** start and **exactly one** end.

### 5.4 Links

A link line looks like:

```
name1-name2
```

Examples: `1-2`, `2-5`.

Rules:

- Both rooms **must already exist**.
- A room **cannot link to itself** (`a-a` is invalid).
- The **same pair cannot appear twice**.
- Links are **undirected** — store the connection in both directions.

### 5.5 Worked example

```
##start
1 23 3
2 16 7
#comment
3 16 3
4 16 5
5 9 3
6 1 5
7 4 8
##end
0 9 5
0-4
0-6
1-3
4-3
5-2
3-5
4-2
2-1
7-6
7-2
7-4
6-5
```

Reading it: 3 ants are implied by the example output (the count would be line 1 in a real file). `1` is the start room, `0` is the end room. `#comment` is skipped. The links wire the rooms into a graph with multiple possible routes from `1` to `0`.

---

## 6. Output format (full spec)

The program prints to standard output in this exact order:

1. The **original input file, unchanged** (number of ants, the rooms, the links).
2. A **single blank line**.
3. One line **per turn**, each line listing every ant that moved that turn.

A move is written `Lx-y`:

- `x` is the ant number (`1..N`).
- `y` is the room the ant moved **into** this turn.
- Moves on the same turn are separated by **single spaces**.
- Only ants that actually moved appear on the line.

### 6.1 Example (example00)

```
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

L1-2
L1-3 L2-2
L1-1 L2-3 L3-2
L2-1 L3-3 L4-2
L3-1 L4-3
L4-1
```

This finishes in 6 turns, which is the required maximum for this map.

### 6.2 Movement rules the output must respect

- An ant moves **at most once per turn**.
- The room an ant moves into must be **empty** at the end of the turn (start and end excepted).
- A tunnel is used **at most once per turn**.
- At the end, **all ants are in the end room**.
- Running the same input multiple times must produce the **same result** (deterministic).

---

## 7. The algorithm (the heart of the project)

This is the most important section. Read it fully before writing M2 or M3.

### 7.1 Why shortest path alone is wrong

If all ants follow one shortest path, only one ant can occupy each room per turn, so they queue up single file. With many ants this is slow. We need ants flowing through **multiple paths in parallel**.

### 7.2 The two requirements for the paths

1. **Disjoint:** the chosen paths must not share any intermediate room. If two paths crossed at room `X`, two ants could need `X` on the same turn — illegal (one ant per room).
2. **Balanced for the ant count:** more paths is not always better. Extra paths are often longer, and forcing more disjoint paths can lengthen the short ones. We must pick the **set of paths that minimizes turns for our specific number of ants**.

### 7.3 Modeling it as a flow problem

Because each room holds one ant, we give each room **capacity 1**. The standard trick:

- **Split every room** into an "in" node and an "out" node, joined by an internal edge of **capacity 1**. All tunnels connect `out(a) -> in(b)` and `out(b) -> in(a)`, each with capacity 1.
- Now finding the maximum number of room-disjoint paths from start to end is a **maximum-flow** problem where every capacity is 1. (Start and end are not capacity-limited.)

### 7.4 Finding the paths: successive shortest augmenting paths

Use BFS-based augmenting (Edmonds-Karp style):

1. Build the residual graph from the split-node model.
2. **BFS** from start to end to find the **shortest** augmenting path.
3. Push one unit of flow along it; reverse those edges in the residual graph (this lets a later path "undo" part of an earlier one if that produces a better overall set).
4. Repeat until no augmenting path remains.

Because each augmentation uses the **shortest** remaining path, after `k` augmentations the flow corresponds to the **`k` disjoint paths with the smallest total length** — exactly what we want.

### 7.5 Choosing how many paths to use

We do not blindly use the maximum number of paths. After **each** augmentation (flow value `k = 1, 2, 3, ...`), we:

1. Decompose the current flow into a concrete set of `k` disjoint paths.
2. Compute how many **turns** it would take to move all `N` ants using those `k` paths (formula below).
3. Remember the `k` that gives the **fewest turns**.

When augmenting stops, we keep the best set found. This guarantees optimal turn count.

> Simpler fallback: if time is short, just take the maximum-flow set and distribute ants over it. This passes most maps; use the "best-k" method above if any turn-count check fails.

### 7.6 Turns formula and ant distribution

For a single path with `L` edges carrying `a` ants one after another, the last ant arrives on turn `L + a - 1`.

To distribute `N` ants over a set of paths so the **slowest** path finishes earliest:

```
sort paths by length (shortest first)
assign each of the N ants, one at a time, to the path that currently
minimizes (path_length + ants_already_assigned_to_it)
total_turns = max over paths of (path_length + ants_on_path - 1)
```

This greedy assignment is provably optimal for a fixed set of paths.

### 7.7 Putting it together (high level)

```
parse file              -> Graph
find best path set      -> []Path        (sections 7.3 - 7.6)
distribute ants + run   -> []string moves (section 7.6 + simulation)
print file + moves
```

---

## 8. Architecture, shared types, and contracts

### 8.1 Package layout

```
lem-in/
  go.mod
  main.go                 # CLI entry point (M4)
  internal/
    parser/               # M1
      parser.go
      parser_test.go
    graph/                # M2
      graph.go
      graph_test.go
    simulation/           # M3
      simulation.go
      simulation_test.go
  examples/               # test data files
  README.md               # M4
```

(Package names and folders can be adjusted in the kickoff, but agree on them **before** anyone branches.)

### 8.2 Shared types — `types.go`

Define these together on Day 0 so all four people code against the same shapes. They can live in the `graph` package or a small shared package.

```go
type Room struct {
    Name string
    X, Y int
}

type Graph struct {
    Rooms   map[string]*Room
    Links   map[string][]string // adjacency: room name -> neighbour names
    NumAnts int
    Start   string
    End     string
}

type Path []string // ordered room names from Start to End
```

### 8.3 Contracts (function signatures)

These are the **only** things the milestones share. As long as everyone honors them, all four can work in parallel against stubs.

```go
// M1 — parser package
func Parse(filePath string) (*Graph, error)

// M2 — graph package
func FindPaths(g *Graph) ([]Path, error) // uses g.NumAnts to pick the best set

// M3 — simulation package
func Simulate(paths []Path, numAnts int) ([]string, error) // returns the turn lines

// M4 — main wires them and also prints the raw file + blank line + lines
```

> Stubbing tip: until M1 is ready, M2 and M3 can hand-write a small `Graph` literal and a couple of `Path` values to test against. That is the whole point of agreeing on these types first.

---

## 9. Git workflow for the team

- One feature branch per milestone
- A shared integration branch: `dev`.
- Each person works on their branch, opens a PR into `dev`.
- **Run the integration tests before every merge into `dev`.**
- Merge `dev` into `main` only when all examples pass and the audit checklist is green.
- **Start M2 first** — it is the bottleneck and everything else is lighter.
- Resolve merge conflicts in person/over call when two people touch the shared `types.go`.

---

## 10. Milestones and issues (detailed)

Each issue below has **What / Why / How / Done when**.

---

### Day 0 — Kickoff (all 4, ~15 min)

Agree on the package layout (section 8.1), write `types.go` (section 8.2), and confirm the contracts (section 8.3). Push it to `dev` so everyone branches from the same base. Then split off.

---

### Milestone 1 — Parsing and validation
**Owner: Stefanos Kamprogiannis (`skamprogiannis`) · branch `feat/parser`**

Goal: turn the input file into a validated `*Graph`, and reject every malformed file cleanly.

#### Issue 1.1 — Read the file and handle missing/empty input
- **What:** Read the filename from `os.Args[1]` and load the whole file.
- **Why:** Everything downstream needs the file contents; a missing argument or empty file is an immediate error.
- **How:**
  1. Check `len(os.Args) >= 2`; if not, return the error.
  2. Read the file (`os.ReadFile`). If it fails or the content is empty, return the error.
  3. Split into lines and trim trailing whitespace/`\r` so the other steps get clean lines.
- **Done when:** Missing arg and empty file both produce the standard error, and a normal file yields a slice of lines.

#### Issue 1.2 — Parse and validate the ant count
- **What:** Read the first meaningful line as the number of ants.
- **Why:** The simulation needs `N`, and `N` must be a positive integer.
- **How:**
  1. Take the first non-comment line.
  2. `strconv.Atoi` it.
  3. Reject if it errors, or the value is `<= 0`.
  4. Store in `Graph.NumAnts`.
- **Done when:** `0`, negatives, and non-numbers error; valid counts are stored.

#### Issue 1.3 — Parse rooms (names, coordinates, duplicates)
- **What:** Parse every `name x y` line into a `Room`.
- **Why:** Rooms are the graph's nodes; invalid rooms must be caught.
- **How:**
  1. Split the line into exactly 3 fields.
  2. Reject names that start with `L` or `#`, or contain spaces.
  3. `strconv.Atoi` both coordinates; reject non-integers.
  4. Reject a name that already exists.
  5. Add the `Room` to `Graph.Rooms`.
- **Done when:** Bad names, non-int coords, and duplicates all error; valid rooms are stored.

#### Issue 1.4 — Handle `##start` and `##end`
- **What:** Apply the start/end commands to the room line that follows.
- **Why:** The algorithm needs to know where ants begin and end.
- **How:**
  1. When you see `##start`, set a flag so the **next** room line is recorded in `Graph.Start`.
  2. Same for `##end` -> `Graph.End`.
  3. Treat other `#...` lines as comments (skip). Ignore unknown commands.
  4. Require exactly one start and one end; error otherwise.
- **Done when:** Exactly one start and one end are required; comments and unknown commands are ignored without error.

#### Issue 1.5 — Parse and validate links
- **What:** Parse every `a-b` line into an undirected edge.
- **Why:** Links are the graph's edges; pathfinding walks them.
- **How:**
  1. Split on `-` into exactly two room names.
  2. Reject if either room does not exist.
  3. Reject self-links (`a-a`).
  4. Reject duplicate pairs (check both directions).
  5. Append to `Graph.Links` for **both** rooms.
- **Done when:** Unknown rooms, self-links, and duplicates error; valid links appear in both directions.

#### Issue 1.6 — Emit `ERROR: invalid data format` on any failure
- **What:** A single, consistent error path.
- **Why:** The audit checks the exact message and that the program never crashes.
- **How:**
  1. Every validation failure returns an `error`.
  2. `main` prints exactly `ERROR: invalid data format` to stderr and exits.
  3. Optionally append a reason (e.g. `, no start room found`) — bonus, not required.
  4. **Never** `panic` or print a stack trace.
- **Done when:** Every invalid input prints the message and exits cleanly.

#### Issue 1.7 — Parser unit tests
- **What:** `parser_test.go` with table-driven tests.
- **Why:** Parsing has the most edge cases; tests prevent regressions.
- **How:** One test case per invalid scenario (bad count, dup room, unknown link, missing start/end, self-link) plus one fully valid file whose resulting `Graph` you assert.
- **Done when:** All invalid cases return errors and the valid case builds the expected `Graph`.

---

### Milestone 2 — Graph and pathfinding
**Owner: George Tzimokas (`gtzimoka`) · this is the hardest milestone**

Goal: from a `*Graph`, produce the set of disjoint paths that minimizes turns for `g.NumAnts`. Read section 7 first.

#### Issue 2.1 — Build the working adjacency map
- **What:** A clean neighbour map for traversal.
- **Why:** Keeps the algorithm independent of the parser's structures.
- **How:** Convert `Graph.Links` into a `map[string][]string` (or your residual-graph structure). Keep it inside the `graph` package.
- **Done when:** You can list every neighbour of any room.

#### Issue 2.2 — BFS shortest path
- **What:** Find one shortest path from start to end.
- **Why:** It is the building block for augmenting paths.
- **How:** Standard BFS with a `parent` map; rebuild the path by walking parents back from end. Return `Path` (room names) or empty if none.
- **Done when:** Returns a correct shortest path on the examples, empty when none exists.

#### Issue 2.3 — Node splitting (vertex capacity 1)
- **What:** Enforce one-ant-per-room in the flow model.
- **Why:** Without it, "disjoint" paths could still collide inside a room.
- **How:** Split each room into `in`/`out` with a capacity-1 internal edge; tunnels connect `out -> in` both ways with capacity 1. Start and end are exempt from the capacity-1 limit.
- **Done when:** No two returned paths share an intermediate room.

#### Issue 2.4 — Augmenting paths (max flow)
- **What:** Find disjoint paths via successive shortest augmenting paths.
- **Why:** This yields, at each step `k`, the `k` shortest-total disjoint paths.
- **How:** Repeatedly BFS the residual graph, push one unit of flow, reverse the used edges; stop when no augmenting path remains. (Edmonds-Karp on unit capacities.)
- **Done when:** The maximum number of disjoint paths is found and reversing works.

#### Issue 2.5 — Decompose flow and pick the best set for N
- **What:** Turn the flow into concrete paths and choose how many to use.
- **Why:** Fewest turns, not most paths, is the objective (section 7.5).
- **How:**
  1. After each augmentation, decompose the flow into a `[]Path`.
  2. Compute `turns(N, paths)` using the formula in 7.6.
  3. Track the path set with the minimum turns.
  4. Return that best set.
- **Done when:** `FindPaths` returns the path set giving the fewest turns for `g.NumAnts`.

#### Issue 2.6 — Handle the no-path case
- **What:** Fail gracefully when start and end are disconnected.
- **Why:** It is a valid input that must not loop or panic.
- **How:** If the first BFS finds no path, return an error (or let `main` print the standard message).
- **Done when:** A disconnected map errors cleanly; tested.

#### Issue 2.7 — Pathfinding unit tests
- **What:** `graph_test.go`.
- **Why:** The algorithm is subtle; tests lock in correctness.
- **How:** Run on test0/test1/example maps. Assert: each path starts at start and ends at end, no two share an intermediate room, and the count matches expectations. Add a disconnected-map case.
- **Done when:** All assertions pass on the sample maps.

---

### Milestone 3 — Simulation and output
**Owner: `ebimai` · branch `feat/m3-simulation`**

Goal: from `[]Path` and `N`, produce the exact turn-by-turn output lines.

#### Issue 3.1 — Distribute ants across paths (greedy)
- **What:** Decide which ant takes which path.
- **Why:** Correct distribution is what achieves the minimum turn count.
- **How:** Sort paths shortest-first. For each ant `1..N`, assign it to the path that currently minimizes `path_length + ants_already_on_it` (section 7.6).
- **Done when:** Distribution matches the optimal turn count for the sample maps.

#### Issue 3.2 — Simulate turns (one ant per room, one use per tunnel)
- **What:** Advance ants turn by turn.
- **Why:** This produces the moves and enforces the rules.
- **How:**
  1. Track each ant's position (index along its path).
  2. Each turn, iterate ants (e.g. the ants on a path move front-to-back) and advance one step **only if** the next room is empty.
  3. Mark a room occupied as soon as an ant enters it this turn; start/end are unlimited.
  4. Ensure each tunnel is used at most once per turn and each ant moves at most once.
  5. Record `(ant, room)` for everything that moved.
  6. Repeat until all ants are at end.
- **Done when:** Ants never share a non-terminal room, tunnels are used once per turn, and all ants reach end.

#### Issue 3.3 — Format the moves
- **What:** Build the `Lx-y` lines.
- **Why:** The audit checks exact formatting.
- **How:** For each turn, join all moves as `Lant-room` separated by single spaces; one line per turn; omit ants that did not move.
- **Done when:** Output matches the example formatting exactly.

#### Issue 3.4 — Print file content, then moves
- **What:** Assemble the final stdout (this overlaps with M4 — coordinate).
- **Why:** The full required format is file + blank line + moves.
- **How:** Print the original file verbatim, a single blank line, then the move lines.
- **Done when:** Output byte-matches the example outputs.

#### Issue 3.5 — Simulation unit tests
- **What:** `simulation_test.go`.
- **Why:** Lock in turn count and move correctness.
- **How:** Feed known path sets + ant counts; assert turn count and moves. Include an edge case where one short path carries all ants, and verify determinism (same input -> same output).
- **Done when:** Turn counts and moves match expectations across runs.

---

### Milestone 4 — Integration, CLI, performance, and bonus
**Owner: Stefanos Kamprogiannis (`skamprogiannis`) · integration/refinement branch**

Goal: wire everything into a working binary, make it robust and fast, and (optionally) build the visualizer. This milestone also owns the **audit-gap items**.

#### Issue 4.1 — Wire the `main.go` pipeline
- **What:** Connect parser -> graph -> simulation -> print.
- **Why:** This is the actual program.
- **How:** `main` reads args, calls `Parse`, then `FindPaths`, then `Simulate`, then prints the file + blank line + lines. Keep `main` thin.
- **Done when:** `go run . example00.txt` produces correct output end to end.

#### Issue 4.2 — End-to-end error handling, no panics, no leaks
- **What:** Make the whole program crash-proof.
- **Why:** The audit fails the project on any crash, hang, or leak.
- **How:** Every stage returns errors; `main` checks each and prints the standard message. Stress-test with the badexamples, empty input, huge ant counts, and disconnected maps.
- **Done when:** No input causes a panic, crash, or hang.

#### Issue 4.3 — Module, layout, gofmt, vet
- **What:** Repo hygiene and good practices.
- **Why:** Required by the audit ("good practices", standard packages only).
- **How:** `go mod init`, organize packages (parser/graph/simulation), ensure `go build ./...` works, run `gofmt -l .` and `go vet ./...` and fix everything. Confirm no non-standard imports.
- **Done when:** Build is clean, gofmt and vet report nothing, only stdlib is used.

#### Issue 4.4 — Performance benchmarks (audit gap)
- **What:** Verify the speed requirements.
- **Why:** 100 ants must finish < 1.5 min; 1000 ants < 2.5 min.
- **How:** Time `example06` with 100 ants and `example07` with 1000 ants (`time go run . ...` or a Go benchmark). If too slow, profile: avoid rebuilding the graph each search, reuse buffers, prefer iterative BFS.
- **Done when:** Both runs are comfortably under their limits.

#### Issue 4.5 — Robustness, determinism, and tunnel-once checks (audit gap)
- **What:** Explicit checks the auditor performs separately.
- **Why:** "ants alone in each room", "each tunnel used once per turn", and "results always correct" are distinct audit questions.
- **How:** Add integration assertions that, for several valid maps: no two ants occupy the same non-terminal room on a turn, no tunnel is used twice in a turn, all ants end at end, and running the same input multiple times gives identical output.
- **Done when:** All four properties hold and are tested.

#### Issue 4.6 — Integration tests on the examples
- **What:** Run the binary on every example file and diff output.
- **Why:** Catches bugs at the seams between milestones.
- **How:** A test (or script) that runs each example and compares the turn count / output to the expected. Run before every merge into `dev`.
- **Done when:** Examples 00–05 meet their turn limits and badexamples print the error.

#### Issue 4.7 — README
- **What:** Short, practical docs.
- **Why:** Required, and helps the team and the auditor.
- **How:** Cover what it does, `go run . file.txt`, the input format, the error message, and how to run the tests.
- **Done when:** A new reader can run and test the project in minutes.

#### Issue 4.8 — Bonus: ant farm visualizer (optional)
- **What:** Animate ants moving through the colony.
- **Why:** Extra credit (and the only place coordinates are used).
- **How:** A second program that reads the move output and renders the ants using room coordinates: `./lem-in map.txt | ./visualizer`. 3D is further extra credit.
- **Done when:** Ants are visibly moving start -> end. Only attempt after the core passes.

---

## 11. Acceptance criteria (verified)

Self-check this before submitting. Each line is something the auditor verifies.

**Functional**
- [x] Only standard packages are used.
- [x] `go run . example00.txt` reads the colony correctly.
- [x] Only `##start` and `##end` are accepted as commands (others ignored).
- [x] Output format is correct: file, blank line, one line per turn, moves as `Lx-y`.
- [x] example00 solves in ≤ 6 turns.
- [x] example01 solves in ≤ 8 turns.
- [x] example02 solves in ≤ 11 turns.
- [x] example03 solves in ≤ 6 turns.
- [x] example04 solves in ≤ 6 turns.
- [x] example05 solves in ≤ 8 turns.
- [x] badexample00 prints `ERROR: invalid data format`.
- [x] badexample01 prints `ERROR: invalid data format`.
- [x] example06 with 100 ants runs in < 1.5 minutes.
- [x] example07 with 1000 ants runs in < 2.5 minutes.
- [x] Ants are alone in each room (except start/end).
- [x] Each tunnel is used only once per turn.
- [x] All ants end in the end room.
- [x] Results are always correct across repeated runs (deterministic).
- [x] All error cases produce a message.
- [x] No empty or incomplete work; builds without crashes or leaks.

**Bonus (extra credit, not required)**
- [ ] Visualizer that shows ants moving.
- [ ] More specific error output.
- [ ] 3D visualizer.

**Basic**
- [x] Runs quickly and efficiently (no unnecessary work).
- [x] Has test files covering each case.
- [x] Code follows good practices.

---

## 12. Testing strategy

- **Unit tests** per package: `parser_test.go`, `graph_test.go`, `simulation_test.go`.
- **Integration tests** (M4) run the full binary on every example and diff output.
- **Robustness tests**: badexamples, empty input, disconnected maps, huge ant counts — assert no crash/hang.
- **Determinism test**: run the same input several times, assert identical output.
- **Performance**: time example06 (100 ants) and example07 (1000 ants).

---

## 13. Definition of done

The project is done when:

1. Every box in section 11 (Functional + Basic) is checked.
2. `gofmt` and `go vet` are clean and only the standard library is used.
3. The program never crashes, hangs, or leaks on any input.
4. All tests pass and `dev` is merged into `main`.

Bonus items are encouraged but not part of the core definition of done.

---

## 14. Suggested timeline (speedrun)

| Phase | Who | Focus |
|---|---|---|
| Day 0 | All 4 | Kickoff: layout, `types.go`, contracts. |
| Phase 1 | A, B, C, D in parallel | A: parsing. B: pathfinding (start immediately — bottleneck). C: simulation against stub paths. D: scaffold main + repo + test harness. |
| Phase 2 | All 4 | Integrate on `dev`; run examples; fix seams. |
| Phase 3 | D (+ help) | Performance pass on example06/07; determinism and robustness. |
| Phase 4 | Anyone free | Bonus visualizer. |

Start M2 first. Keep `dev` green. Merge often.

# A tiny graph and ReBAC tutorial in Go

This small project teaches one idea:

> ReBAC checks whether an allowed relationship path connects a user to a
> resource. Path found → allow. No path → deny.

Along the way you build the two skills the project exists for: implementing a
graph in plain Go, and seeing how authorization becomes a graph search. No
advanced graph theory, no external dependencies.

## Run it

```bash
go run ./cmd/demo
go test ./...
```

The demo prints the decision *and* the path that explains it:

```text
Can Alice edit the roadmap? true
allowed by relationship path: member_of -> editor_of -> contains
  user:alice --member_of--> team:engineering
  team:engineering --editor_of--> folder:product
  folder:product --contains--> document:roadmap
```

## Learn it in three short sessions

The material lives in [the graph guide](docs/GRAPH-FOR-REBAC.md), which pairs
every concept with the actual Go code in this repo and traces the code by
hand. One lesson per session is plenty.

### Session 1: How is a graph stored in Go?

Read **Lesson 1** of the guide, then the top of
[`internal/graph/graph.go`](internal/graph/graph.go) through `Neighbors`.
The whole graph is one `map[NodeID][]Edge`.

```bash
go test -v ./internal/graph -run TestNeighbors
```

Then stop. You have learned how the graph is stored.

### Session 2: How do we move through a graph?

Read **Lesson 2** of the guide — it traces BFS and DFS step by step. Then
read `BFS`, `DFS`, and `FindPath` in
[`internal/graph/graph.go`](internal/graph/graph.go).

```bash
go run ./cmd/traversal
go test -v ./internal/graph
```

Then stop. You have learned graph traversal, including why cycles are safe.

### Session 3: How does ReBAC use the graph?

Read **Lesson 3** of the guide, then
[`internal/rebac/evaluator.go`](internal/rebac/evaluator.go) — it is under 70
lines.

```bash
go test -v ./internal/rebac
go run ./cmd/demo
```

You now have the complete picture: graph storage, traversal, and a working
(if deliberately tiny) ReBAC evaluator on top.

## When the first pass feels comfortable

- [Go idioms used here](docs/GO-IDIOMS.md) explains every Go-specific trick
  in the codebase (map zero values, slice-as-queue, recursive closures, …).
  Keep it open beside the code if the *syntax* rather than the *graph idea*
  is ever what's blocking you.
- [Trees and graphs](docs/TREES-VS-GRAPHS.md) answers "isn't this just a
  tree?" — short, worth reading whenever the question occurs to you.
- [Deeper ReBAC graph ideas](docs/DEEPER-REBAC-GRAPHS.md) is five small
  experiments: multiple paths, nested groups, cycles, search cost, and what
  production systems add. Intentionally separate so day one stays small.

## Files

```text
internal/graph/         Graph storage, traversal, and tests
internal/rebac/         ReBAC path checker and tests
cmd/demo/main.go        Runnable end-to-end example
cmd/traversal/main.go   Visible BFS, DFS, and path example
docs/                   The lessons and optional deeper experiments
```

## Try these exercises

Do one at a time:

1. Before running `cmd/traversal`, predict its BFS and DFS output on paper.
2. In `cmd/demo`, add Carol as a direct `viewer_of` the roadmap and give the
   `view` permission a rule that allows it.
3. Write a test proving Carol can view but cannot edit.

Remember: if no allowed relationship path exists, access is denied.

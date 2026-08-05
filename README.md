# A tiny Graph and ReBAC tutorial in Go

This small project teaches one idea:

> ReBAC checks whether an allowed relationship path connects a user to a resource.

You do not need advanced graph theory. The project has no external dependencies.

## Run it

```bash
go run ./cmd/demo
go test ./...
```

The demo prints why Alice can edit a document:

```text
user:alice --member_of--> team:engineering
team:engineering --editor_of--> folder:product
folder:product --contains--> document:roadmap
```

## Learn it in three short sessions

### Session 1: What is a graph?

Read [the graph guide](docs/GRAPH-FOR-REBAC.md), through “Lesson 1.” Then read
the top of [`internal/graph/graph.go`](internal/graph/graph.go) through `Neighbors`.

Run:

```bash
go test -v ./internal/graph -run TestNeighbors
```

Then stop. You have learned how the graph is stored.

### Session 2: How do we move through a graph?

Read “Lesson 2” in [the graph guide](docs/GRAPH-FOR-REBAC.md). Then read `BFS`,
`DFS`, and `FindPath` in [`internal/graph/graph.go`](internal/graph/graph.go).

Run:

```bash
go run ./cmd/traversal
go test -v ./internal/graph
```

Then stop. You have learned graph traversal.

### Session 3: How does ReBAC use the graph?

Read “Lesson 3” in [the graph guide](docs/GRAPH-FOR-REBAC.md). Then read
[`internal/rebac/evaluator.go`](internal/rebac/evaluator.go).

Run:

```bash
go test -v ./internal/rebac
go run ./cmd/demo
```

You now have the graph foundation used by this small ReBAC evaluator.

## When the first pass feels comfortable

Read [Deeper ReBAC graph ideas](docs/DEEPER-REBAC-GRAPHS.md). It uses five small
experiments to cover multiple paths, nested groups, cycles, search cost, and the
limits of this learning evaluator. It is intentionally separate so you do not
need to absorb everything on day one.

## Files

```text
internal/graph/         Graph storage, traversal, and tests
internal/rebac/         ReBAC path checker and tests
cmd/demo/main.go        Runnable example
cmd/traversal/main.go   Visible BFS, DFS, and path example
docs/                   Short lessons and optional deeper experiments
```

## Try these exercises

Do one at a time:

1. Before running `cmd/traversal`, predict its BFS and DFS output.
2. Add Carol as a direct viewer of the roadmap.
3. Write a test proving Carol can view but cannot edit.

Remember: if no allowed relationship path exists, access is denied.

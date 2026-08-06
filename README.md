# A tiny Graph and ReBAC tutorial in Go

This small project teaches one idea:

> ReBAC checks whether an allowed relationship path connects a user to a resource.

You do not need advanced graph theory. Sessions 1-3 have no external
dependencies; Session 4 adds one small YAML library so policy files can be
plain, readable YAML.

## Vocabulary

One consistent set of words is used everywhere in this repo — code,
comments, and docs. If a term below is new, its session teaches it in
context; this table is just a map back to where.

| Term | Meaning | Taught in |
|---|---|---|
| **node** | A thing: a user, team, folder, document, org, account. | Session 1 |
| **edge** | A directed, labelled connection between two nodes. | Session 1 |
| **relation** | The label on an edge, e.g. `member_of`, `accessor`, `admin`. | Session 1 |
| **relationship** | One fact, as a triple: `(subject, relation, resource)`. It is exactly one edge — `subject --relation--> resource`. Not called a "fact" or "tuple" elsewhere in this repo, even though other systems use those words. | Session 1, [Session 4](docs/POLICY-RULES.md#relationships-subject-relation-resource) |
| **subject** | The node asking to do something — usually a user. The left side of a relationship. | Session 1 |
| **resource** | The node being acted on. The right side of a relationship. | Session 1 |
| **path** | A sequence of edges connecting one node to another. | Session 2 |
| **traversal** | Following edges to explore or search a graph (BFS, DFS). | Session 2 |
| **action** | The question `Check(subject, action, resource)` asks: can subject do this to resource? Not a grant — nobody "has" an action the way a subject has a relation; it is answered fresh from the relationships that currently exist. Also called a "permission" in some other ReBAC/RBAC systems — this repo avoids that word because it invites confusing an action with a relationship. | Session 3, [Session 4](docs/POLICY-RULES.md#rules-two-ways-to-satisfy-an-action) |
| **rule** | One way to satisfy an action. Sessions 1-3 write a rule as an exact, fixed-length list of relations in Go. Session 4 writes the same idea as YAML, as either `this` or `via_relationship`. | Session 3, [Session 4](docs/POLICY-RULES.md#rules-two-ways-to-satisfy-an-action) |
| **type** | The part of a node ID before its `:`, e.g. `bankaccount` in `bankaccount:daytoday`. Selects which rules apply. | [Session 4](docs/POLICY-RULES.md) |
| **policy** | The full set of types, relations, and actions, loaded from `policy.yaml`. | [Session 4](docs/POLICY-RULES.md) |
| **`this` rule** | Direct check: does `subject --relation--> resource` exist? | [Session 4](docs/POLICY-RULES.md#rules-two-ways-to-satisfy-an-action) |
| **`via_relationship` rule** | Indirect check: does the resource have a relation to some parent, and does the subject separately have a relation to that *same* parent? | [Session 4](docs/POLICY-RULES.md#via_relationship-access-granted-through-another-object) |

Session 4's doc also covers two [modeling
conventions](docs/POLICY-RULES.md#modeling-conventions) — which node an
edge belongs on (child points at parent) and noun-style vs verb-style
relation names — and includes a [full mapping to Google's Zanzibar
paper](docs/POLICY-RULES.md#mapping-to-the-zanzibar-paper), the design
most production ReBAC systems (OpenFGA, SpiceDB, Ory Keto) are based on,
if you want to connect these words to that vocabulary too.

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

## Learn it in four short sessions

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

### Session 4: Rules and relationships as data, not Go code

Sessions 1-3 hard-code the allowed relation path in Go. Real policies
usually need to be declared and changed without recompiling, and sometimes
need indirection ("an org admin can access every account in that org")
that a fixed-length path can't express. Read
[Session 4: Rules as data](docs/POLICY-RULES.md), then look at
[`internal/policy`](internal/policy) and the example files in
[`examples/bank`](examples/bank).

Run:

```bash
go test -v ./internal/policy
go run ./cmd/policy-demo
```

## When the first pass feels comfortable

Read [Deeper ReBAC graph ideas](docs/DEEPER-REBAC-GRAPHS.md). It uses five small
experiments to cover multiple paths, nested groups, cycles, search cost, and the
limits of this learning evaluator. It is intentionally separate so you do not
need to absorb everything on day one.

## Files

```text
internal/graph/         Graph storage, traversal, and tests
internal/rebac/         ReBAC path checker and tests (fixed relation paths in Go)
internal/policy/        YAML-defined relationships and rules, and their evaluator
examples/bank/          Example relationships.yaml and policy.yaml for internal/policy
cmd/demo/main.go        Runnable example
cmd/traversal/main.go   Visible BFS, DFS, and path example
cmd/policy-demo/main.go Runnable example using YAML relationships and policy
docs/                   Short lessons and optional deeper experiments
```

## Try these exercises

Do one at a time:

1. Before running `cmd/traversal`, predict its BFS and DFS output.
2. Add Carol as a direct viewer of the roadmap.
3. Write a test proving Carol can view but cannot edit.

Remember: if no allowed relationship path exists, access is denied.

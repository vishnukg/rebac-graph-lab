# Graphs for ReBAC, from the ground up

This guide teaches two things at once:

1. how to build a graph in Go, and
2. how a ReBAC authorization check is really just a graph search.

It has three lessons. Each one explains an idea, then shows you the exact Go
code in this repository that implements it. Read one lesson at a time — each
ends with a way to check your understanding before moving on.

If a line of code ever confuses you for *Go* reasons rather than *graph*
reasons — `map[NodeID][]Edge`, the two-value map read, `var visit func(...)` —
look it up in [the Go idioms list](GO-IDIOMS.md) and come back.

The single idea behind everything here:

> ReBAC answers "can this user do this?" by searching for an allowed
> relationship path from the user to the resource. Path found → allow.
> No path → deny.

---

## Lesson 1: Storing a graph in Go

### The idea

A graph is just **things** and **relationships between things**:

- A **node** is a thing: a user, a team, a document.
- An **edge** is a relationship between two nodes.

```text
user:alice --member_of--> team:engineering
```

That is a graph with two nodes and one edge. Two properties of that edge
matter a lot for authorization:

- It is **directed**: the arrow points from Alice to Engineering.
  "Alice is a member of Engineering" is true; "Engineering is a member of
  Alice" is not. Direction carries meaning.
- It is **labelled**: `member_of` says *what kind* of relationship this is.
  A `viewer_of` edge must never grant the same access as an `editor_of` edge,
  so the label has to travel with the edge.

### How to store it: an adjacency list

The most common way to store a graph is an **adjacency list**: for every node,
keep a list of its outgoing edges.

```text
user:alice        -> [ (member_of, team:engineering) ]
team:engineering  -> [ (editor_of, document:roadmap) ]
document:roadmap  -> [ ]
```

### The Go code

In Go, "look up a list by name" is exactly what a map of slices gives you.
Here are the core types from [`internal/graph/graph.go`](../internal/graph/graph.go):

```go
type NodeID string

type Edge struct {
    To       NodeID
    Relation string
}

type Graph struct {
    edges map[NodeID][]Edge
}
```

Read `map[NodeID][]Edge` out loud: *for each node, a slice of its outgoing
edges*. That one map **is** the whole graph. There is no separate node
struct — a node is just a key in the map.

Why this shape?

- A map looks up any node's edges by name instantly, no matter how large the
  graph grows — and traversal (Lesson 2) constantly asks "what are this
  node's edges?"
- A slice is the natural Go container for "a few edges per node".

Adding an edge is small enough to read in one breath:

```go
func (g *Graph) AddEdge(from NodeID, relation string, to NodeID) {
    g.AddNode(from)
    g.AddNode(to)

    // Avoid storing the exact same relationship twice.
    for _, edge := range g.edges[from] {
        if edge.To == to && edge.Relation == relation {
            return
        }
    }

    g.edges[from] = append(g.edges[from], Edge{To: to, Relation: relation})
}
```

Two details worth noticing:

- `AddEdge` registers both endpoints first, so callers never have to call
  `AddNode` themselves. Convenience like this keeps the demo code clean.
- The duplicate check compares **both** `To` and `Relation`. Alice can be
  both a `viewer_of` and an `editor_of` the same document — those are two
  different edges between the same two nodes.

A fair question: why does `AddNode` exist at all, if `AddEdge` adds nodes
automatically? Because **a thing can exist before it has any relationships**.
A brand-new user who has not joined a team yet, or a fresh document nothing
links to, is still a node. Registering it lets the graph tell apart two
situations that should behave differently:

- *a known node with no edges yet* — traversal starts there and finds just
  that one node;
- *a name the graph has never heard of* — traversal returns nothing at all.

You will rarely call `AddNode` yourself, but that distinction is why it is
part of the API.

One more idiom: `Neighbors` returns a **copy** of the edge slice, so a caller
who modifies the returned slice cannot corrupt the graph by accident. Handing
out your internal slice is a classic Go leak; the copy prevents it.

### Check yourself

- In `g.edges`, what is the key and what is the value?
- Why does an edge need a label *and* a direction in an authorization system?
- What happens if you call `AddEdge` twice with identical arguments?

Then run:

```bash
go test -v ./internal/graph -run TestNeighbors
```

That is all of Lesson 1: a graph in Go is one map from node name to outgoing
edges.

---

## Lesson 2: Moving through the graph

### The idea

**Traversal** means starting at a node and following edges to visit others.
All the traversals in this project run on graphs like this one, from
[`internal/graph/graph_test.go`](../internal/graph/graph_test.go):

```text
      A
     / \
    B   C
     \ /
      D
      |
      +--> A   (D points back to A: a cycle)
```

Every edge points downward except the last one, which loops back up.

### Breadth-first search (BFS): explore in rings

First, the goal — before any code. Group the example graph's nodes by their
distance from A:

```text
distance 0:  A        (the start)
distance 1:  B, C     (one edge away)
distance 2:  D        (two edges away)
```

BFS promises to visit nodes in exactly that order — A, then B and C, then D —
like ripples spreading from a stone dropped in water. Never a far node before
a nearer one.

To keep that promise, BFS maintains two pieces of bookkeeping:

- a **to-do line** (a *queue*): new nodes join the **back**, and the next
  node to visit is taken from the **front**, like a checkout line;
- a **seen set** (called *visited* in the code): every node that has ever
  joined the line, so no node gets in line twice.

Then it repeats one move until the line is empty: **take the front node,
record it, and put its never-seen neighbors at the back.** Walk it by hand,
starting with A in line:

```text
                                                 line (front first)   recorded order
start: A is in line                              [A]
1. Take A. Neighbors B, C are new — line up      [B, C]               A
2. Take B. Neighbor D is new — line it up        [C, D]               A, B
3. Take C. Neighbor D is already seen — skip     [D]                  A, B, C
4. Take D. Neighbor A is already seen — skip     []                   A, B, C, D
```

The line is empty, so the result is `[A B C D]`. The level-by-level order
falls out for one reason only: new discoveries join the **back** of the line,
so everything at distance 1 is taken before anything at distance 2.

Step 4 is also the cycle defence in action. D's edge points back to A, but A
is in the seen set, so it is skipped. **That single check is the entire
protection** — without it, the loop A → B → D → A would run forever.

Now the real code from `graph.go` — the same moves in Go spelling:

```go
visited := map[NodeID]bool{start: true}
queue := []NodeID{start}          // the to-do line
order := make([]NodeID, 0)

for len(queue) > 0 {
    current := queue[0]           // take from the front...
    queue = queue[1:]             // ...of the line
    order = append(order, current)

    for _, edge := range g.edges[current] {
        if !visited[edge.To] {
            visited[edge.To] = true          // mark as seen when lined up,
            queue = append(queue, edge.To)   // then join the back
        }
    }
}
```

(Go has no queue type; taking `queue[0]` and re-slicing with `queue[1:]` is
the standard small-scale substitute.) One subtlety: a node is marked seen
when it **joins** the line, not when it is taken off. Look at step 3 of the
hand-trace — B and C both point at D, and marking early is what stops D from
being lined up twice.

### Depth-first search (DFS): one path to the end, then back up

DFS explores the way you would explore a maze: at every junction, take the
first unexplored corridor and keep going **deeper**. When you hit a dead end,
walk back to the previous junction and try its next corridor. On the example
graph, starting at A:

```text
A's first edge goes to B ............ go deeper    recorded: A, B
B's only edge goes to D ............. go deeper    recorded: A, B, D
D's edge goes to A — already seen ... dead end, walk back to B
B has no more edges ................. walk back to A
A's next edge goes to C ............. go deeper    recorded: A, B, D, C
C's edge goes to D — already seen ... dead end, nothing left anywhere
```

Result: `[A B D C]`. The entire B-branch is finished before C is ever
touched — that is the "depth-first" in the name.

Now notice what "walk back to the previous junction" requires: something must
*remember* every junction you passed and which corridors are still untried.
Recursion provides that memory for free. Each function call remembers where
it was standing; when a call returns, you are automatically back at the
previous node, in front of its next edge:

```go
var visit func(NodeID)
visit = func(current NodeID) {
    visited[current] = true
    order = append(order, current)

    for _, edge := range g.edges[current] {
        if !visited[edge.To] {
            visit(edge.To)   // go deeper; returning = walking back
        }
    }
}
visit(start)
```

(The two-step `var visit func(NodeID)` declaration is how a Go closure calls
itself — see [the Go idioms list](GO-IDIOMS.md).) Here is the same trace
again, written as the calls nest — each indent is one step deeper into the
maze, and un-indenting is the walk back:

```text
visit A          order: A
  visit B        order: A, B         (A's first edge)
    visit D      order: A, B, D      (B's only edge)
      D -> A     skipped, already seen — the cycle defence again
  visit C        order: A, B, D, C   (back at A, its second edge)
    C -> D       skipped, already seen
```

**BFS and DFS side by side:** both visit the same four nodes, and both rely
on the same seen-set to survive the cycle. The only difference is the order:
BFS spreads outward in rings (driven by a queue), DFS probes one full path at
a time (driven by recursion).

### From "visit everything" to "find a path"

`FindPath` is BFS with one addition: a `previous` map that records how each
node was reached (`previous[D] = B` means "we reached D from B"). When the
search dequeues the target, `buildPath` walks the `previous` chain backwards —
target to start — and reverses it. Because BFS reaches every node in the
fewest possible edges, the reconstructed path is a shortest path.

See both runs live:

```bash
go run ./cmd/traversal
go test -v ./internal/graph
```

### Check yourself

- Cover the traces above and predict `BFS("A")` and `DFS("A")` yourself.
- In BFS, why is a node marked visited when it is *queued* rather than when
  it is *dequeued*? (Hint: two nodes both point at D.)
- What exactly stops traversal from looping around the D → A edge forever?

That is all of Lesson 2: a queue gives you BFS, recursion gives you DFS, a
visited set makes cycles safe, and a `previous` map turns "visited" into "a
path".

---

## Lesson 3: ReBAC is a path search with rules about labels

### The idea

**ReBAC** (Relationship-Based Access Control) stores facts like these:

```text
user:alice        --member_of--> team:engineering
team:engineering  --editor_of--> folder:product
folder:product    --contains-->  document:roadmap
```

Each fact is one labelled, directed edge — the data *is* a graph, the same
structure you built in Lesson 1. A permission is then defined as a **rule**:
a sequence of edge labels that counts as an allowed path. In this project:

```text
"edit" is granted by the path:  member_of -> editor_of -> contains
```

So the authorization question

```text
Can alice edit document:roadmap?
```

becomes a graph question:

```text
Starting at user:alice, can we reach document:roadmap by following
a member_of edge, then an editor_of edge, then a contains edge?
```

### The Go code: traversal that must match labels

Plain BFS follows *any* edge. ReBAC traversal may only follow the edge label
the rule demands at each step. `FindPathByRelations` in `graph.go` does this
with recursion, where `relationIndex` tracks how much of the rule is left:

```go
walk = func(current NodeID, relationIndex int, path []Step) ([]Step, bool) {
    if relationIndex == len(relations) {
        return path, current == target   // rule used up: are we there?
    }

    wantedRelation := relations[relationIndex]
    for _, edge := range g.edges[current] {
        if edge.Relation != wantedRelation {
            continue                     // wrong label: not allowed to follow
        }

        step := Step{From: current, Relation: edge.Relation, To: edge.To}
        if result, found := walk(edge.To, relationIndex+1, append(path, step)); found {
            return result, true
        }
    }
    return nil, false
}
```

Trace it for alice / `["member_of", "editor_of", "contains"]`:

```text
walk(user:alice, 0)        want member_of  -> follow to team:engineering
walk(team:engineering, 1)  want editor_of  -> follow to folder:product
walk(folder:product, 2)    want contains   -> follow to document:roadmap
walk(document:roadmap, 3)  rule used up    -> current == target? yes -> found
```

If any step finds no edge with the wanted label, that branch returns
`false` and the search backtracks to try other edges. Note there is no
`visited` map here — none is needed, because every recursive call consumes
one label from a finite rule. The search cannot run longer than the rule is
long, even if the graph contains cycles.

### The evaluator: rules + search + explanation

[`internal/rebac/evaluator.go`](../internal/rebac/evaluator.go) wraps this in
an authorization API. It stores rules per permission, and `Check` tries each
rule until one produces a path:

```go
for _, rule := range rules {
    path, found := e.graph.FindPathByRelations(subject, resource, rule)
    if found {
        return Decision{Allowed: true, Path: path, Reason: ...}
    }
}
return Decision{Reason: "denied: no allowed relationship path was found"}
```

Three properties of this loop are the heart of ReBAC:

- **Any one rule is enough.** A permission may have several rules (direct
  `viewer_of` *or* inherited through a team); rules are OR-ed together.
- **Deny is the default.** No rule matched — including "nobody ever defined
  a rule for this permission" — means denied. The evaluator never guesses.
- **Decisions carry evidence.** The returned `Decision` includes the actual
  path, so the demo can print *why* Alice can edit, step by step. Real
  authorization systems care deeply about this kind of explainability.

Run it:

```bash
go test -v ./internal/rebac
go run ./cmd/demo
```

### Check yourself

- Bob has a `viewer_of` edge to the roadmap. Which rule in
  [`evaluator_test.go`](../internal/rebac/evaluator_test.go) lets him view
  it, and why can he still not edit it?
- Why does `FindPathByRelations` not need a `visited` map when BFS does?
- What does `Check` return for a permission nobody added rules for?

---

## Four things to remember

1. ReBAC data is a directed, labelled graph — in Go, one `map[NodeID][]Edge`.
2. Traversal (BFS/DFS) searches that graph; a visited set makes cycles safe.
3. A permission rule is a sequence of edge labels; checking access means
   searching for a path that matches one rule.
4. No matching path means deny. Always.

Wondering how trees fit into this? Read [Trees and graphs](TREES-VS-GRAPHS.md).
When all three lessons feel comfortable, continue with
[Deeper ReBAC graph ideas](DEEPER-REBAC-GRAPHS.md).

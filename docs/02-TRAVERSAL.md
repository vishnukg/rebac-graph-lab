# Session 2: Traversal finds paths

**Traversal** means following edges from one node to another. It is the
skill this whole project builds on: Session 3 will show that a ReBAC
check — "can Alice edit this document?" — is nothing more than a traversal
that succeeds or fails. This session teaches the traversal itself, on
small abstract graphs, so the mechanics are clear before authorization
vocabulary arrives.

```text
    A
   / \
  B   C
   \ /
    D
```

There are two common ways to traverse a graph. Both start at one node and
eventually visit everything reachable from it; they differ only in the
*order* they visit, and that order comes from one data structure choice:
a queue or a stack.

## Breadth-first search (BFS)

BFS visits nearby nodes first: the start, then everything one edge away,
then everything two edges away. Nodes the same number of edges from the
start form a **level**:

```text
    A        level 0: the start
   / \
  B   C      level 1: one edge away
   \ /
    D        level 2: two edges away
```

BFS finishes a whole level before starting the next. To do that it keeps a
**queue** — first in, first out, like a line at a shop — of nodes waiting
to be visited:

1. Put the start node in the queue.
2. Take the *front* node out, record it, and queue its unvisited neighbors.
3. Repeat until the queue is empty.

Tracing it on the diamond graph above:

| Step | Visit | Queue after | Why |
|------|-------|-------------|-----|
| 1 | A | B, C | A's neighbors line up |
| 2 | B | C, D | D waits behind C, which was found first |
| 3 | C | D | C's neighbor D is already queued |
| 4 | D | *(empty)* | nothing new — done |

Result: `A, B, C, D` — level 0, then level 1, then level 2. Because BFS
never visits a level-2 node while a level-1 node is still waiting, the
*first* time it reaches any node is via a path with the fewest edges. That
guarantee is why `FindPath` below is built on BFS.

## Depth-first search (DFS)

DFS follows one branch as far as it can, then backs up to the most recent
fork and tries the next branch. It uses a **stack** — last in, first out —
and the tidiest way to get one is recursion: each nested call pushes a
node, each return pops it ("backing up"). Tracing it on the same graph:

| Step | Visit | Stack after | Why |
|------|-------|-------------|-----|
| 1 | A | A | start |
| 2 | B | A, B | follow A's first edge before its second |
| 3 | D | A, B, D | follow B's first edge; D has nothing unvisited |
| — | | A | D and B are finished — pop back to the fork at A |
| 4 | C | A, C | try A's next edge; C's neighbor D is already visited |

Result: `A, B, D, C` — the deep node D comes before the shallow node C.
Notice what the stack holds at step 3: `A, B, D` is exactly the path from
the start to the current node. That is the useful property of DFS — it
always knows *how it got here*. Path-finding with backtracking needs that,
and so does an authorization system when it has to explain *why* access
was granted: the answer is the path. (The `Step` type in
`internal/graph/graph.go` exists for exactly that explanation.)

## Choosing between them

Same nodes, different order — so when does the order matter?

- **BFS** when distance matters: shortest path, "who is closest",
  exploring level by level.
- **DFS** when you are hunting for *any* path and want to try one route
  fully before another. `FindPathByRelations` at the end of this session is
  DFS-shaped: it walks one candidate path deep, and backtracks when the
  labels stop matching.
- For a plain "visit everything" walk, either works. Both take time
  proportional to the graph's size — every node and edge is handled at
  most once, `O(nodes + edges)`.

They also hold different things in memory while running: the BFS queue
holds a whole level at once (large when the graph is *wide*), while the DFS
stack holds one path from the start (large when the graph is *deep*). For
graphs the size of this project's, neither matters — but it explains why
the two traversals feel different in the trace tables above.

## Cycles

A graph can contain a loop, called a **cycle**:

```text
A --> B --> C
^           |
|-----------|
```

Both traversals keep a `visited` set and skip any node already in it. When
DFS in the cycle above reaches C and sees the edge back to A, A is already
marked, so the edge is ignored and the walk ends instead of looping
forever. The same set also stops a node reachable along two branches (like
D in the diamond graph) from being visited twice.

## Check yourself

`cmd/traversal` runs both traversals on a graph with a cycle in it:

```text
    A
   / \
  B   C
  |   |
  D   E
  |
  +----> A  (cycle)
```

Before running it, trace the two orders yourself the way the tables above
do — where does the cycle edge get skipped? Then check your answer:

```sh
go run ./cmd/traversal
```

## Paths

A **path** is the sequence of edges used to reach a node. When at least one
path from node X to node Y exists, Y is said to be **reachable** from X —
and asking "is Y reachable from X?" is called a **reachability** question.
That word is worth keeping: it is the standard name, in graph theory and in
ReBAC literature alike, for exactly the question an authorization check
asks. In a directed graph reachability is one-way — D may be reachable from
A while A is not reachable from D.

`FindPath` uses BFS to find a short path. `FindPathByRelations` finds a
path whose labels appear in a specific order — that second function is the
one ReBAC is built on, as Session 3 shows.

## What to remember

1. BFS explores level by level with a queue, so its first arrival at any
   node is a path with the fewest edges.
2. DFS explores deepest-first with a stack, and the stack always holds the
   path from the start to the current node.
3. A `visited` set is what makes cycles safe.
4. A path is a sequence of edges; labels can constrain which paths count.
5. "Is there any path from X to Y?" is a **reachability** question — the
   shape of question authorization checks ask.

That is enough for Session 2. Next: [Session 3 — the ReBAC check](03-REBAC-CHECK.md).

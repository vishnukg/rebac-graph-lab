# Session 2: Traversal finds paths

**Traversal** means following edges from one node to another.

```text
    A
   / \
  B   C
   \ /
    D
```

There are two common ways to traverse a graph.

**Breadth-first search (BFS)** visits nearby nodes first:

```text
A, B, C, D
```

It uses a queue. It is useful for finding a path with the fewest edges.

**Depth-first search (DFS)** follows one branch deeply before returning:

```text
A, B, D, C
```

It uses recursion or a stack.

## Cycles

A graph can contain a loop, called a **cycle**:

```text
A --> B --> C
^           |
|-----------|
```

BFS and DFS remember visited nodes so they do not loop forever.

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

1. BFS explores nearest-first with a queue; DFS explores deepest-first.
2. A `visited` set is what makes cycles safe.
3. A path is a sequence of edges; labels can constrain which paths count.
4. "Is there any path from X to Y?" is a **reachability** question — the
   shape of question authorization checks ask.

That is enough for Session 2. Next: [Session 3 — the ReBAC check](03-REBAC-CHECK.md).

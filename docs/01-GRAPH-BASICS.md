# Session 1: A graph stores things and relationships

A graph has **nodes** and **edges**.

- A node is a thing, such as a user, team, or document.
- An edge is a relationship between two things.

```text
user:alice --member_of--> team:engineering
```

This example has two nodes and one edge. The edge is:

- **directed**: the arrow goes from Alice to Engineering;
- **labelled**: `member_of` explains what the connection means.

Direction and labels matter. `Alice member_of Engineering` does not mean
`Engineering member_of Alice`. Likewise, `viewer_of` must not grant the same
access as `editor_of`.

## Two nodes can share several edges

Because edges are labelled, the same pair of nodes can be connected more
than once — one edge per distinct relationship:

```text
user:alice --viewer_of--> document:roadmap
user:alice --editor_of--> document:roadmap
```

Alice is both a viewer and an editor of the same document, and each fact is
its own edge. If you have seen a textbook definition of a graph that allows
at most one edge between any two nodes, note that this project (and every
ReBAC system) relaxes it: a graph permitting parallel edges is called a
**multigraph**. You never need that word again here — just don't be
surprised that adding a second relationship between the same two nodes is
normal and does not replace the first.

## How the graph is stored

The graph stores each node beside its outgoing edges. This is called an
**adjacency list**:

```text
user:alice:
  (member_of, team:engineering)

team:engineering:
  (editor_of, document:roadmap)
```

In this project, `internal/graph.Graph` stores that list and `AddEdge` adds a
relationship.

Keep the adjacency list in mind — it is not just an implementation detail.
It means a node knows its own *outgoing* edges cheaply, and knows nothing
about edges pointing *at* it. Session 4 turns that asymmetry into a rule
about which direction to write relationships in.

## Isn't this just a tree?

If your data looks like a folder hierarchy, you may wonder why we need a
general graph at all. Short version: a tree allows exactly one path between
any two nodes, and permission data routinely has several (direct access
*and* inherited access to the same document). The longer answer, with
pictures, is in [Trees and graphs](TREES-VS-GRAPHS.md) — optional reading.

## What to remember

1. A node is a thing; a directed, labelled edge is a relationship.
2. Direction and labels carry the meaning — never ignore them.
3. The same two nodes can share several edges — one per relationship.
4. The graph is an adjacency list: each node lists its own outgoing edges.

That is enough for Session 1. Next: [Session 2 — traversal](02-TRAVERSAL.md).

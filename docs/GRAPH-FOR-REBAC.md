# The graph knowledge needed for this ReBAC project

This guide has three short lessons. Read one lesson at a time.

## Lesson 1: A graph stores things and relationships

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

That is enough for Lesson 1.

## Lesson 2: Traversal finds paths

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

A graph can contain a loop, called a **cycle**:

```text
A --> B --> C
^           |
|-----------|
```

BFS and DFS remember visited nodes so they do not loop forever.

A **path** is the sequence of edges used to reach a node. `FindPath` uses BFS to
find a short path. `FindPathByRelations` finds a path whose labels appear in a
specific order.

That is enough for Lesson 2.

## Lesson 3: ReBAC looks for an allowed path

ReBAC means **Relationship-Based Access Control**.

Two words matter from here on, because the rest of this project uses them
precisely:

- A **relationship** is one fact: `subject --relation--> resource`. It is
  one edge, like `user:alice --member_of--> team:engineering`.
- An **action** is the question being asked: "can this subject do this to
  this resource?" — the middle argument of `Check(subject, action,
  resource)`, like `"edit"` below. An action is not itself a fact and
  nobody "has" one; it is answered fresh, each time, by searching the
  relationships for an allowed path.

Here is an inherited-access path:

```text
Alice
  --member_of--> Engineering
  --editor_of--> Product Folder
  --contains--> Roadmap Document
```

The edit rule says these labels form an allowed path:

```text
member_of -> editor_of -> contains
```

The question:

```text
Can Alice edit the Roadmap?
```

becomes:

```text
Can we start at Alice and reach the Roadmap by following
member_of, then editor_of, then contains?
```

The evaluator checks each step:

1. Start at Alice and find a `member_of` edge.
2. From that team, find an `editor_of` edge.
3. From that folder, find a `contains` edge.
4. Confirm that the final node is the Roadmap.

If every step matches, access is allowed. Otherwise, access is denied.

A direct relationship is simply a one-edge path:

```text
Bob --viewer_of--> Roadmap
```

One action may have several allowed paths. For example, viewing might be
granted by a direct `viewer_of` path or by an inherited team path. Finding any
one allowed path is enough.

That is the central idea behind this learning evaluator.

## Four things to remember

1. ReBAC data forms a directed, labelled graph.
2. Traversal searches that graph.
3. Edge labels decide whether a path grants an action.
4. No allowed path means deny.

For the tree question, read [Trees and graphs](TREES-VS-GRAPHS.md).

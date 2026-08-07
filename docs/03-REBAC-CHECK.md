# Session 3: ReBAC looks for an allowed path

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

That is the central idea behind this learning evaluator. In the vocabulary
of Session 2: a ReBAC check is a **reachability** question with an extra
condition on the labels — not just "can Alice reach the Roadmap?" but "can
she reach it along a path whose labels the rule allows?"

## Where this is going

This session's evaluator answers every check with one general tool: a walk
that follows an arbitrary-length label sequence, backtracking through the
graph until it reaches the target (`FindPathByRelations`). Hold on to that
picture. In Session 5 you will see that once rules are written as data,
each rule pins the path down so tightly that almost no searching is left —
the same walk, shrunk to one or two edge lookups. That shrunken form is
called a **degenerate** graph walk, and it is why real ReBAC checks can be
fast.

## What to remember

1. ReBAC data forms a directed, labelled graph.
2. Traversal searches that graph.
3. Edge labels decide whether a path grants an action.
4. No allowed path means deny.

That is enough for Session 3. Next:
[Session 4 — relationships as data](04-RELATIONSHIPS-AS-DATA.md).

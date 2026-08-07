# Deeper ReBAC graph ideas

Read this only after the five short sessions in the README. Each section is a
small experiment, not a new chapter to memorize.

## 1. One resource can have several paths

Alice might have direct and inherited access at the same time:

```text
Alice --------viewer_of----------------> Roadmap
  |
  +--member_of--> Team --viewer_of-----> Roadmap
```

A `view` action can allow either path. The evaluator tries its rules one at
a time and stops when one succeeds. This is an **OR** choice: one valid path is
enough.

Try it:

```bash
go test -v ./internal/rebac -run TestAnyAllowedPath
```

## 2. A relationship can be inherited through several levels

A team can belong to another team:

```text
Alice --member_of--> App Team
      --member_of--> Engineering
      --editor_of--> Roadmap
```

That path has three edges. Our small evaluator can follow it when the rule lists
all three labels:

```text
member_of -> member_of -> editor_of
```

Try it:

```bash
go test -v ./internal/rebac -run TestNestedGroup
```

The learning evaluator uses fixed-length rules. A production evaluator often
supports recursive group membership, where the number of group levels is not
known in advance.

## 3. Cycles must not cause infinite searching

Relationships can form a loop:

```text
A --> B --> C
^           |
+-----------+
```

Plain BFS and DFS keep a `visited` set. Once a node has been visited, they do not
visit it again. Without that check, they could follow the loop forever.

Try it:

```bash
go run ./cmd/traversal
go test -v ./internal/graph -run Cycles
```

`FindPathByRelations` is also safe in this lab because every step consumes one
label from a finite rule. A more flexible recursive evaluator needs both cycle
detection and a maximum search depth.

## 4. BFS and DFS do different work

BFS explores nearby nodes first. That makes it useful when you want the path
with the fewest edges. DFS explores one branch first and can be simple when you
only need to find any matching path. ([Session 2](02-TRAVERSAL.md) traces
both orders step by step — this section adds only the production angle.)

For `V` visited nodes and `E` examined edges, ordinary BFS and DFS take roughly
`V + E` work. You do not need to calculate this. The useful idea is simply:
larger and more connected graphs require more searching.

A real authorization evaluator therefore uses limits, deadlines, and caching.
If it cannot prove access is allowed, it should deny access rather than guess.

## 5. This evaluator teaches the core, not every policy feature

This project's model is:

```text
relationship data -> labelled graph
action rule       -> allowed label sequence
authorization     -> search for a matching path
```

Full ReBAC systems can also express:

- **OR**: either path may grant access;
- **AND**: two conditions must both be true;
- **NOT**: one relationship can exclude access;
- recursive groups and folder hierarchies.

Those features combine graph searches with logical rules. The graph foundation
does not change: nodes represent things, labelled edges represent relationships,
and evaluation searches for evidence connecting a subject to a resource.
Each of these features — and the operational machinery real systems add
around them — is documented in
[What production ReBAC adds](PRODUCTION-FEATURES.md).

## You have understood it when...

You can look at a small relationship diagram and explain:

1. which nodes and labelled edges it contains;
2. which paths connect a user to a resource;
3. which path matches the requested action;
4. why a cycle does not make traversal run forever;
5. why no matching path must result in denial.

There is no need to memorize anything beyond that. Change an edge, predict the
decision, and run the tests—the prediction is the learning.

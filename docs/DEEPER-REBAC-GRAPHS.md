# Deeper ReBAC graph ideas

Read this after the three lessons in [the main guide](GRAPH-FOR-REBAC.md).
Each section is a small experiment with a test you can run and tinker with —
not a chapter to memorize.

## 1. One resource can have several paths

Alice might have direct and inherited access at the same time:

```text
Alice --------viewer_of----------------> Roadmap
  |
  +--member_of--> Team --viewer_of-----> Roadmap
```

The `view` permission accepts either path. The evaluator tries its rules one
at a time and stops at the first success — rules are an **OR**: any one valid
path is enough.

Try it:

```bash
go test -v ./internal/rebac -run TestAnyAllowedPath
```

Then experiment: in `TestAnyAllowedPathIsEnough`, delete the direct
`viewer_of` edge and predict the decision before rerunning. Access should
still be allowed — through the other path.

## 2. Access can be inherited through several levels

A team can belong to another team:

```text
Alice --member_of--> App Team
      --member_of--> Engineering
      --editor_of--> Roadmap
```

That path has three edges, and our evaluator follows it when the rule lists
all three labels:

```text
member_of -> member_of -> editor_of
```

Try it:

```bash
go test -v ./internal/rebac -run TestNestedGroup
```

Notice the limitation you just bumped into: the rule hard-codes *exactly two*
levels of team nesting. Add a third team in the middle and the test fails
until you also change the rule. Production evaluators solve this with
recursive rules ("member of a team, at any depth"), where the number of
levels is not known in advance.

## 3. Cycles must not cause infinite searching

Relationships can form a loop:

```text
A --> B --> C
^           |
+-----------+
```

The two traversals in this project stay safe in two different ways, and it is
worth being able to say why:

- **BFS and DFS** keep a `visited` set. A node that has been seen once is
  never entered again, so the loop is walked at most once.
- **`FindPathByRelations`** has no visited set at all — it doesn't need one,
  because every recursive step consumes one label from a finite rule. The
  search is over after at most `len(rule)` steps no matter how the edges
  loop.

A more powerful recursive evaluator (see experiment 2) loses that second
guarantee and needs cycle detection plus a maximum depth.

Try it:

```bash
go run ./cmd/traversal
go test -v ./internal/graph -run Cycles
```

## 4. Search has a cost, and real systems respect it

For `V` visited nodes and `E` examined edges, BFS and DFS do roughly `V + E`
work. You do not need to calculate this precisely; the useful intuition is
just: **bigger, more connected graphs take more searching**, and an
authorization check runs on every request.

That is why real evaluators add limits, deadlines, and caching on top of the
search. And when a check cannot finish — timeout, depth limit hit — the safe
answer is deny. An authorization system must never guess "probably fine".

## 5. What this evaluator teaches, and what it leaves out

This project's whole model fits in three lines:

```text
relationship data -> labelled graph
permission rule   -> allowed label sequence
authorization     -> search for a matching path
```

Full ReBAC systems (Google Zanzibar and its descendants, for example) can
also express:

- **OR** across paths — this project has that, as rule lists;
- **AND** — two conditions must both hold;
- **NOT** — a relationship (like `banned_from`) that removes access;
- recursive groups and folder hierarchies of unknown depth.

Those features combine graph search with logic on top. The graph foundation
underneath never changes: nodes are things, labelled edges are relationships,
and evaluation searches for evidence connecting a subject to a resource.

## You have understood it when...

You can look at a small relationship diagram and explain:

1. which nodes and labelled edges it contains;
2. which paths connect a user to a resource;
3. which path matches the requested permission;
4. why a cycle does not make traversal run forever;
5. why no matching path must result in denial.

Nothing beyond that needs memorizing. Change an edge, predict the decision,
run the tests — the prediction is the learning.

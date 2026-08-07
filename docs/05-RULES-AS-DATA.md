# Session 5: Rules as data

Session 4 loaded the facts. This session loads their meaning: the rules in
`policy.yaml` that decide which paths through those facts grant which
actions.

Sessions 1-3 hard-coded the allowed relation path in Go:

```go
evaluator.AddRule("edit", "member_of", "editor_of", "contains")
```

That only works when the path has a fixed number of steps, all in one
direction. Real authorization usually needs something more: "an org admin
can access every bank account in that org" is not a fixed path from a
specific user to a specific account — it depends on which org the account
belongs to, discovered at check time.

## Rules: two ways to satisfy an action

A policy groups relations and actions by object **type** (the part of an
ID before the `:`). Each action lists rules; the first rule that matches
grants access. Every rule is still just a description of a path to look
for — the same idea as `FindPathByRelations` in Session 3, just data
instead of a Go argument list.

A quick word on naming, because it trips people up: an **action** (like
`access` below) is *not* a grant. Nobody "has" the `access` action the way
Ben "has" the `admin` relation. `admin` is stored data — a fact about the
world, written down in `relationships.yaml`. `access` is a rule for
*reading* that data — the answer to "given the relationships that exist
right now, can this subject do this?", recomputed every time `Check` runs.
If your instinct is to call `admin` a "permission" because it's something
Ben was granted, you're right — that instinct is correct, it's just that
this project calls that a **relationship**, and reserves the word
**action** for the question, not the grant.

### `this`: access granted directly

**`this`** is a one-edge path: check a direct relation between subject and
resource.

```yaml
- this: accessor
```

Read as: is there a `subject --accessor--> resource` edge? This is exactly
`FindPathByRelations(subject, resource, []string{"accessor"})`.

### `via_relationship`: access granted through another object

`this` can only ever ask "is there a direct edge from the subject to the
resource?" That covers Ed, who is a direct `accessor` of the account. It
cannot cover Ben, who is never directly connected to the account at all —
Ben is an `admin` of the *org*, and the account merely *belongs to* that
org. The action has to be computed by combining two separate relationships
that neither party holds alone.

This pattern — "anyone with role X on the parent gets access to everything
inside it" — is one of the most common shapes of policy anywhere: folders
inherit from drives, documents from folders, accounts from orgs. The
Zanzibar paper calls the rule for it `tuple_to_userset` (see the
[Zanzibar mapping](ZANZIBAR-MAPPING.md) for that story); this project
names it for what it does in our domain language — **`via_relationship`**:

```yaml
- via_relationship:
    through: org
    requires: admin
```

It is still a two-edge path, but the second edge's *starting point* isn't
the subject-to-resource edge — it's a node the first edge lands on. Read it
as two hops, starting from opposite ends and meeting in the middle:

1. From the **resource**, follow its own `through` relation to find its
   parent: `resource --org--> parent`.
2. From the **subject**, check for a `requires` relation to that *same*
   parent: `subject --admin--> parent`.

Both hops must land on the same parent node — that's what makes "admin of
*this account's* org" different from "admin of *some* org." In graph
terms: does an `org`-edge from the resource and an `admin`-edge from the
subject converge on one node?

```text
bankaccount:daytoday --org--> org:acme <--admin-- user:ben
                              ^^^^^^^^
                         same node both edges must reach
```

If Ben were an admin of a *different* org, the first hop would still land
on `org:acme`, the second hop would look for an `admin` edge from Ben to
`org:acme` specifically, not find one, and the rule would not match —
exactly the case Exercise 3 below asks you to prove.

See [`examples/bank/policy.yaml`](../examples/bank/policy.yaml).

## The evaluator

[`internal/policy/evaluator.go`](../internal/policy/evaluator.go) does
nothing exotic: `Check` finds the resource's type, looks up the action's
rules, and tries each one in order. The first rule that matches grants
access; if none match, the answer is deny. No BFS, no recursion, no cycle
tracking — because both rule types name their relations exactly, there is
only ever one or two edges to check, and the code below shows why.

## Every rule is a path search — a degenerate one

This deserves to be said explicitly, because it is the idea that connects
this session back to everything before it. Session 3's evaluator answered
every check with one general tool:

```go
// The general walk: follow an arbitrary-length label sequence,
// backtracking through the graph until the target is reached.
path, found := g.FindPathByRelations(subject, resource,
	[]string{"member_of", "editor_of", "contains"})
```

Both rule types in this session are that same search, shrunk until no
searching is left. A search this small is called **degenerate**: still
technically a graph walk, but with so few steps that the loop disappears.

A `this` rule is the walk with a one-label sequence. Nothing to explore,
nothing to backtrack — the walk is a single edge lookup:

```go
// this: accessor — a "path search" exactly one edge long.
//
//	subject --accessor--> resource        (found? allowed)
found := e.hasEdge(subject, "accessor", resource)
```

A `via_relationship` rule is two of those one-edge walks, launched from
opposite ends, that must meet on the same node:

```go
// via_relationship {through: org, requires: admin} — two one-edge
// walks that must converge:
//
//	resource --org-->   parent            (walk 1: child to parent)
//	subject  --admin--> parent            (walk 2: subject to SAME parent)
for _, edge := range e.graph.Neighbors(resource) {
	if edge.Relation != "org" {
		continue // not the edge the rule follows
	}
	parent := edge.To
	if e.hasEdge(subject, "admin", parent) {
		return allowed // both walks reached the same node
	}
}
```

Compare the three: the general walk loops over a label sequence and
recurses; `this` has no loop at all; `via_relationship` has one loop of
(usually) one iteration. Same graph, same edges, same question — "does an
allowed path connect subject to resource?" — with progressively less
freedom about where the path may go. That lack of freedom is exactly what
makes the policy evaluator fast and predictable, and it is what you give
up first when rules need recursion again (see Exercise 4).

You can watch the degenerate walks happen: `cmd/policy-demo` prints each
edge the evaluator walked, in order —

```text
check(user:ben, access, bankaccount:daytoday) = true
  walked: bankaccount:daytoday --org--> org:acme
  walked: user:ben --admin--> org:acme
```

One step for Ed's direct rule, two converging steps for Ben's, no steps
for Carol's denial (nothing was found to walk).

Run it:

```bash
go run ./cmd/policy-demo
go test -v ./internal/policy
```

## What to remember

1. A `this` rule is a one-edge path check; `via_relationship` is two
   one-edge checks that must converge on the same parent node.
2. Every rule is still a path search — just a degenerate one, with the
   searching squeezed out. That is why checks are fast and predictable.
3. An action is the question, never a grant; the grant is a relationship.

## Try these exercises

1. Run the two experiments from Session 4's "Two files, two jobs": first
   delete the `via_relationship` rule from `policy.yaml` and run the demo
   (Ben should flip to denied while his `admin` edge still exists), then
   restore it and delete Ben's `admin` edge from `relationships.yaml`
   instead (same flip, opposite cause). Predict each output before
   running.
2. Add a `viewer` relation and a `view` action that any org employee
   satisfies (hint: `through: org`, `requires: employee`).
3. Add a second bank account belonging to a different org, and a rule
   proving an admin of Acme cannot access it.
4. `via_relationship` only follows one hop. Sketch (in comments, no code
   needed) what a *second* level of indirection would require — for
   example, an admin role inherited from a parent company that owns
   several orgs. (Hint: this is the same shape of problem Session 3's
   nested-group exercise solved with a longer relation path — think about
   why a fixed path doesn't work here, and what would have to become a
   search instead of a lookup.)

If you go on to read the Zanzibar paper or a production system built on it
(OpenFGA, SpiceDB, Ory Keto), the
[Zanzibar mapping](ZANZIBAR-MAPPING.md) translates this repo's vocabulary
to theirs.

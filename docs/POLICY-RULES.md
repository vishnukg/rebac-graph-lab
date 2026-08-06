# Session 4: Rules as data, not code

Sessions 1-3 hard-coded the allowed relation path in Go:

```go
evaluator.AddRule("edit", "member_of", "editor_of", "contains")
```

That only works when the path has a fixed number of steps, all in one
direction. Real authorization usually needs something more: "an org admin
can access every bank account in that org" is not a fixed path from a
specific user to a specific account — it depends on which org the account
belongs to, discovered at check time.

`internal/policy` teaches that with two ideas: **relationships as data**
and **rules as data**. It is still the same graph from Session 1 — nodes,
directed labelled edges, and a walk over them. Nothing new in graph theory,
only in how the rules that guide the walk are written down.

## Two files, two jobs

Session 4 splits authorization into two YAML files, and the split is the
lesson:

- **`relationships.yaml` holds what is true.** Who relates to whom, right
  now: Ed is an accessor of the account, Ben is an admin of Acme, the
  account belongs to Acme. Each line is one graph edge and nothing more.
  Relationships carry no opinion about access — an `admin` edge does not
  say what an admin may do.
- **`policy.yaml` holds what it means.** Which relations exist per type,
  and which paths through them satisfy each action. The policy contains no
  users, no accounts, no orgs — it never mentions Ed or Acme. It only
  describes path *shapes*: "a direct `accessor` edge grants `access`,"
  "an `admin` edge into the account's org grants `access`."

The graph walk needs both, and they answer different questions. The
relationships are the **terrain**: the nodes and edges the walker can
actually step on. The policy is the **map legend**: which of the many
possible paths across that terrain count as proof of access. A walk with
no terrain finds nothing; a walk with no legend cannot tell an
authorizing path (`accessor`) from an irrelevant one (`follows`).

The proof that they are genuinely separate: change one without the other
and watch `check` flip.

- Delete the `via_relationship` rule from the policy. Ben still "has the
  admin role" — his `admin` edge is untouched in `relationships.yaml` —
  but `check(user:ben, access, ...)` now denies. The fact survived; its
  *meaning* was revoked.
- Delete Ben's `admin` edge instead. The policy still says org admins get
  access, but the walk finds no edge to stand on, so it denies. The
  meaning survived; the fact is gone.

Access exists only where the two meet: a real path in the terrain whose
shape the legend accepts.

The split also matches who changes each file, and how often. Relationships
change constantly and are written by the application at runtime — every
account opened, employee hired, or role granted adds or removes edges.
Policy changes rarely and deliberately — it is authorization *design*,
reviewed like code, and one rule edit changes what every matching edge in
the graph means. Keeping them separate means the fast-changing data never
requires re-reviewing the rules, and a rule change never requires touching
the data.

## Relationships: subject, relation, resource

A relationship is the same thing you already met in Session 1, written as
a triple instead of a diagram: `(subject, relation, resource)`. It becomes
one graph edge, `subject --relation--> resource`, exactly like before —
just loaded from YAML instead of written as `AddEdge` calls:

```yaml
relationships:
  - subject: user:ed
    relation: accessor
    resource: bankaccount:daytoday
```

See [`examples/bank/relationships.yaml`](../examples/bank/relationships.yaml).
Loaded, that file is this graph:

```text
user:ed      --employee--> org:acme
user:ed      --accessor--> bankaccount:daytoday
bankaccount:daytoday --org--> org:acme
user:ben     --admin-->    org:acme
user:carol   --employee--> org:acme
```

Draw it as a picture and the shape of the problem becomes visible:

```text
user:ed ----accessor----> bankaccount:daytoday
                                  |
                                 org
                                  v
user:ed    ---employee-----> org:acme
user:ben   ---admin----------^
user:carol ---employee-------^
```

Ed reaches the account directly. Ben does *not* have an edge to the account
at all — he reaches it only by way of the org. Carol has no path to the
account by any relation.

## Modeling conventions

Before writing rules, two decisions come up every time you add a
relationship: **which node gets the edge**, and **what to call the
relation**. Neither is enforced by the code — the graph will happily store
anything — so the conventions below are what keep the data usable.

### Convention 1: put the edge on the child, pointing at its parent

Some relationships are between an actor and a thing (`user:ed --accessor-->
bankaccount:daytoday`). Their direction is obvious: from the one *doing*
to the one *being done to*.

But some relationships are between a thing and its **container**: an
account and its org, a document and its folder, a folder and its drive.
Call the contained thing the **child** and the container the **parent**.
The fact "the day-to-day account belongs to Acme" involves no actor at
all, so which side should be the subject of the triple?

The convention: **the child is the subject, and its edge points at the
parent.**

```yaml
# Yes: the child (account) holds an edge pointing at its parent (org).
- subject: bankaccount:daytoday
  relation: org
  resource: org:acme

# No: the parent holding edges to each of its children.
# - subject: org:acme
#   relation: owns
#   resource: bankaccount:daytoday
```

Here is the reason, and it is mechanical, not stylistic. Recall from
Lesson 1 that the graph is an **adjacency list**: each node stores its own
*outgoing* edges. That means the evaluator can cheaply answer one shape of
question — "what does this node point at?" — and cannot cheaply answer the
reverse — "what points at this node?" — without scanning the entire graph.

Now look at what `via_relationship` needs to do, starting from a
`check(user:ben, access, bankaccount:daytoday)` call. It stands on the
**account** and asks: "which org does this account belong to?" If the edge
lives on the account (child), that is one lookup in the account's own
edge list — follow `org`, arrive at `org:acme`, done. If instead the edge
lived on the org (parent), the evaluator standing on the account would
find *no outgoing edge at all* — the fact exists in the graph, but it is
recorded somewhere the evaluator, starting from the account, cannot reach.
Answering it would mean asking the reverse question, "which org points at
me?", i.e. scanning every node in the graph looking for an `owns` edge
that happens to end at this account.

So "put the edge on the child" is really: **store the edge on the node
that evaluation starts from.** Checks always start at the resource (and
the subject) — never at the parent — because the parent is precisely the
thing the check does not know yet and is trying to find.

Two footnotes worth knowing:

- One parent, many children makes this natural: an org with ten thousand
  accounts would otherwise hold ten thousand `owns` edges, while each
  account holds exactly one `org` edge.
- Production systems (Zanzibar, OpenFGA, SpiceDB) *can* answer reverse
  questions ("list everything this user can access") — they do it by
  maintaining a second, reverse index of every relationship. This lab has
  one forward index, so the convention is load-bearing here; in real
  systems it is still the way relationships are written, with the reverse
  index built automatically underneath.

### Convention 2: noun style or verb style for relation names

There are two common ways to name a relation, and both appear in this
repo — deliberately, so you see each in action.

**Verb style** (Sessions 1-3: `member_of`, `editor_of`, `contains`) names
the relation so the whole triple reads aloud as a sentence:

```text
user:alice --member_of--> team:engineering   "Alice is a member of Engineering"
folder:product --contains--> document:roadmap "Product contains Roadmap"
```

Its strength is that every relationship is self-explaining — nothing to
teach, just read it.

**Noun style** (Session 4: `accessor`, `admin`, `employee`, `org`) names
the relation as if it were a *field on the resource*, read right-to-left:

```text
user:ed --accessor--> bankaccount:daytoday   "an accessor of the account: Ed"
bankaccount:daytoday --org--> org:acme       "the account's org: Acme"
```

Its strength shows up in the policy file. A rule names relations without
their endpoints, and noun style keeps it readable there:

```yaml
- via_relationship:
    through: org        # "through the account's org..."
    requires: admin     # "...require admin on it"
```

Noun style is also the Zanzibar/OpenFGA/SpiceDB convention (`parent`,
`owner`, `viewer`), so it is what you will meet in production schemas.

**The preference:** either is fine — the only real rule is to pick one
per schema and stay consistent. This repo uses verb style in the early
sessions because self-reading sentences teach best, and noun style in the
bank example because that is what real policy files look like. A quick
test for whichever you choose: say the triple out loud in that style's
reading order, and if the sentence is false (like `account --owns--> org`,
"the account owns Acme"), the name or the direction is wrong.

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
[mapping table](#mapping-to-the-zanzibar-paper) for that story); this
project names it for what it does in our domain language —
**`via_relationship`**:

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

### Every rule is a path search — a degenerate one

This deserves to be said explicitly, because it is the idea that connects
Session 4 back to everything before it. Session 3's evaluator answered
every check with one general tool:

```go
// The general walk: follow an arbitrary-length label sequence,
// backtracking through the graph until the target is reached.
path, found := g.FindPathByRelations(subject, resource,
	[]string{"member_of", "editor_of", "contains"})
```

Both Session 4 rule types are that same search, shrunk until no searching
is left. A search this small is called **degenerate**: still technically a
graph walk, but with so few steps that the loop disappears.

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

## Mapping to the Zanzibar paper

This project renames things to keep one consistent vocabulary — subject,
relation, resource, relationship — instead of mixing in Zanzibar's own
words. If you go on to read the [Zanzibar
paper](https://research.google/pubs/pub48190/) or a real system built on
it (OpenFGA, SpiceDB, Ory Keto), here is the translation table:

| This repo             | Zanzibar paper                          | Notes |
|------------------------|------------------------------------------|-------|
| relationship            | relation tuple                          | Written `object#relation@user` in the paper, e.g. `bankaccount:daytoday#accessor@user:ed`. Same three parts, reverse order: object=resource, user=subject. |
| subject                 | user (or userset)                       | Zanzibar's "user" can also be a *userset* — another object's relation, e.g. `group:eng#member` — so a tuple can point at "everyone in a group" instead of one person. This repo's subjects are always a single leaf ID; `via_relationship` is the one place we approximate a userset reference. |
| resource                 | object                                  | The thing access is being checked against. |
| relation                 | relation                                | Same word, same meaning: the label on the edge/tuple. |
| type (the `org` / `bankaccount` prefix) | namespace / object type | Zanzibar groups relation and rewrite-rule definitions under a named "namespace config" per object type — this repo's `types:` block in `policy.yaml`. |
| policy.yaml              | namespace config                        | The declaration of which relations and actions exist per type, and how each action is computed. |
| action (`access`, `view`, `edit`)  | relation with a userset rewrite rule | Zanzibar does not distinguish a stored fact from a computed answer by name — both are just "relations," some backed by stored tuples, some by a rewrite rule. This repo splits that in two for teaching clarity: `relations:` are stored (matches Zanzibar's plain, tuple-backed relations), `actions:` are computed (matches Zanzibar's rewrite-rule relations). Note this is *not* the everyday sense of "permission" as a grant — see the note above the rules section. |
| `this` rule              | `_this` rewrite rule                    | Same name in spirit, same meaning: "use the tuples stored directly for this relation," rather than computing it. |
| `via_relationship` rule  | `tuple_to_userset` rewrite rule          | Follow a tupleset (a group of tuples matching the resource, e.g. all its `org` edges) and rewrite the check into a userset lookup on whatever those tuples point to. |
| `through`                | `tupleset.relation`                     | Which of the resource's own relations to follow to find the parent object. |
| `requires`                | `computed_userset.relation`             | Which relation the subject must hold on that parent object. |
| `Check(subject, action, resource)` | `Check(object#relation@user)` API | Same question, different argument order: "does this user have this relation to this object?" |

What Zanzibar has that this project deliberately leaves out, to keep the
first pass small: `union`/`intersection`/`exclusion` combinators for rules
(an action here is always an implicit union — the first matching rule
wins), userset subjects in `via_relationship` (so no "is a member of the
subject's group" style checks beyond the one built-in indirection), and
namespace config validation. Those are natural next steps once `this` and
`via_relationship` feel familiar.

## Try these exercises

1. Run the two experiments from "Two files, two jobs": first delete the
   `via_relationship` rule from `policy.yaml` and run the demo (Ben should
   flip to denied while his `admin` edge still exists), then restore it
   and delete Ben's `admin` edge from `relationships.yaml` instead (same
   flip, opposite cause). Predict each output before running.
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

# What production ReBAC adds

Reference, not a lesson — read it after Session 5. This lab teaches the
core: authorization is a search for an allowed relationship path. A
production system (Zanzibar and its descendants: OpenFGA, SpiceDB, Ory
Keto) keeps that core unchanged and adds the features below around it.
Each section says what the feature is, shows a concrete case this lab
cannot express, and explains — in this repo's graph vocabulary — what it
would take.

None of these change the foundation. Nodes are still things, labelled
edges are still relationships, and every check is still "does an allowed
path connect subject to resource?"

## 1. Rule combinators: AND and NOT

This lab's rules combine only one way: an action lists rules, and the
first match grants access. That is an implicit **OR** — any one allowed
path is enough.

Production policies also need:

- **AND (intersection)**: two conditions must both hold. "You can share a
  document only if you are an `editor_of` it *and* a `member_of` its
  owning org."
- **NOT (exclusion)**: one relationship removes access. "Every `employee`
  can view the report, *except* subjects with a `blocked` edge to it."

In graph terms, OR is cheap: run path searches until one succeeds. AND
means every listed search must succeed. NOT is the expensive one — it
requires *proving a path does not exist*, so the search must run to
exhaustion instead of stopping at the first hit, and denial can no longer
be the cheap default case. NOT also makes rule evaluation
order-sensitive, which is why production systems define combinators as an
explicit expression tree (Zanzibar's `union` / `intersection` /
`exclusion`) rather than a first-match-wins list like `policy.yaml`.

## 2. Userset subjects: one edge grants to a whole group

Every subject in this lab is a single leaf node: `user:ed`, `user:ben`.
To make a whole team viewers of a document, you would need one
relationship per member.

Zanzibar lets a relationship's *subject* be a **userset** — a reference
to another object's relation, written `group:eng#member`:

```text
document:roadmap <--viewer-- group:eng#member
        "the viewers of roadmap include: whoever is a member of eng"
```

One edge now grants access to a set whose membership changes elsewhere.
The cost lands on `Check`: finding a `viewer` edge is no longer enough,
because the edge may point at a set. The evaluator must then ask a second
question — "is this subject a `member` of `group:eng`?" — and that
question can itself hit another userset, recursively. The
`via_relationship` rule in Session 5 is this lab's one fixed-shape
approximation of the idea.

## 3. Recursion: hierarchies of unknown depth

This lab's rules are fixed-length label sequences. Session 3 handles a
group inside a group by writing a longer rule (`member_of -> member_of ->
editor_of`), and Session 5's Exercise 4 shows where that breaks: real
hierarchies — folder trees, nested groups, orgs owning orgs — have depth
that is not known when the policy is written.

Production schemas say "a folder's `viewer` includes the `viewer` of its
*parent* folder", a rule that refers to itself. Evaluating it is a true
graph search of unbounded depth, and everything Session 2 taught comes
back for real:

- a **visited set**, because org charts and group graphs contain cycles;
- a **depth limit and deadline**, because a rule that can recurse can
  also run away;
- a defined answer when the limit is hit — deny, never guess (see §6).

This is the moment the "degenerate walk" of Session 5 un-degenerates: the
loop and the recursion that `this` and `via_relationship` squeezed out
return, and BFS/DFS stop being background theory.

## 4. Reverse queries: "list everything Alice can access"

`Check` answers one (subject, action, resource) triple. Products also
need the reverse shapes:

- **list objects**: every document Alice can view (to render her home
  screen);
- **list subjects**: every user who can access this account (to render a
  sharing dialog, or for an audit).

Session 1 explains why this lab cannot do that: the adjacency list stores
each node's *outgoing* edges, so "what points at this node?" means
scanning the whole graph — and answering "everything Alice can access"
honestly would mean running `Check` against every resource. Production
systems maintain a **reverse index** — every relationship written
forwards is also indexed backwards — and walk rules from both ends.
Session 4's "store the edge where evaluation starts" convention is the
one-index version of this idea; the reverse index is what makes the other
direction affordable.

## 5. Consistency: the new-enemy problem

Relationships change while checks are running, and the order in which
changes become visible can break security. The Zanzibar paper's name for
this is the **new-enemy problem**:

1. Alice removes Bob from the `team:acquisition` team.
2. Alice then adds a sensitive document to that team's folder.
3. Bob's next check must not be answered from a state that has seen
   change 2 but not change 1 — that stale mixture shows the new document
   to a removed member, an answer that was never correct at any moment.

A single in-memory graph like this lab's cannot exhibit the problem; a
replicated, cached production store must actively prevent it. Zanzibar's
answer is a snapshot token (a **zookie**): writes return one, and a check
presenting it is guaranteed an answer at least as fresh as that write.
The lesson survives even without the machinery: an authorization result
is only meaningful *relative to a version of the relationship data*.

## 6. Operational safety: limits, caching, deny-on-uncertainty

`DEEPER-REBAC-GRAPHS.md` §4 states the cost model: roughly `V + E` work
per search. At production scale — billions of relationships, millions of
checks per second, checks on every page load — that forces:

- **deadlines and depth limits** on every check, so no single question
  can consume the service;
- **caching** of hot check results and group memberships, plus
  precomputed indexes for pathologically deep or wide sets (Zanzibar's
  "Leopard" index for nested groups);
- **fail-closed semantics**: when a check cannot be completed — timeout,
  partial data, depth limit — the answer is deny. An authorization
  system must never guess "allow".

The lab already models the last rule in miniature: every evaluator here
returns deny when no allowed path is *found*, which is not the same claim
as "no path exists" — and production systems keep that humility
deliberately.

## 7. Explaining decisions

This lab's `Decision` returns the exact path that granted access — run
`cmd/policy-demo` and you can watch every edge walked. Production systems
formalize that into an **expand** API: given a resource and relation,
return the full tree of usersets and rules that produce its member set.
Support teams use it to answer "why can this user see this?", which at
scale is asked as often as `Check` itself. If you extend this lab, the
`Path` field on `Decision` (built from `graph.Step`) is the seed of that
feature.

## 8. Schema management

`policy.yaml` here is trusted as written. A production policy is a schema
that evolves for years, so systems add: validation (a rule may only name
relations that exist on the right types), versioned migrations (renaming
a relation without breaking live checks against old data), and testing
tools that replay recorded checks against a proposed policy before it
ships. The reflex to keep: Session 4 already treats policy as
authorization *design*, reviewed like code — production tooling exists
because a policy mistake is an outage or a breach, not a bug.

## Summary

| Feature | Extends which lab idea | Zanzibar name |
|---|---|---|
| AND / NOT combinators | first-matching-rule OR (Session 5) | `intersection`, `exclusion` |
| Userset subjects | single-leaf subjects (Session 1) | userset, `object#relation` |
| Recursive rules | fixed-length rule paths (Sessions 3, 5) | self-referential userset rewrite |
| Reverse queries | forward-only adjacency list (Sessions 1, 4) | reverse index; Read/Expand APIs |
| Consistency tokens | single in-memory graph | zookie; new-enemy problem |
| Limits, caching, fail-closed | `V + E` cost model (Deeper §4) | deadlines, Leopard index |
| Decision explanation | `Decision.Path` (Sessions 3, 5) | `Expand` API |
| Schema management | trusted `policy.yaml` (Session 5) | namespace config validation |

Read the features in any order, but notice what none of them changed:
the check is still a search for an allowed path. Everything above is
about doing that search correctly at scale — combining searches (§1-3),
running them backwards (§4), pinning them to a moment in time (§5),
bounding them (§6), and explaining or evolving them (§7-8).

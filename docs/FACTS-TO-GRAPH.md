# From facts to a graph — and where facts live in production

Companion to [Session 4](04-RELATIONSHIPS-AS-DATA.md). Read it when either
of these questions nags:

- "Which way is the arrow supposed to point, and how does a pile of facts
  turn into a graph?"
- "Surely the facts don't live in a YAML file in real life?"

## The one mechanical rule

**The arrow always goes from `subject` to `resource`. Always.** Converting
a fact to an edge involves no judgment call — one entry becomes exactly one
arrow, drawn left-to-right from the triple:

```yaml
- subject: user:ed                 # tail of the arrow (where the edge is stored)
  relation: accessor               # the label on the arrow
  resource: bankaccount:daytoday   # head of the arrow (what it points at)
```

becomes, with zero interpretation:

```text
user:ed --accessor--> bankaccount:daytoday
```

## Watch five facts become the graph

The conversion is dumber than you might expect, and that is the point.
Take [`examples/bank/relationships.yaml`](../examples/bank/relationships.yaml)
and process it one fact at a time:

```text
Fact 1: (user:ed, employee, org:acme)
        New nodes: user:ed, org:acme
        Draw:  user:ed --employee--> org:acme

Fact 2: (user:ed, accessor, bankaccount:daytoday)
        user:ed already exists — reuse it. New node: bankaccount:daytoday
        Draw:  user:ed --accessor--> bankaccount:daytoday

Fact 3: (bankaccount:daytoday, org, org:acme)
        Both nodes already exist — just draw the arrow between them.
        Draw:  bankaccount:daytoday --org--> org:acme

Fact 4: (user:ben, admin, org:acme)
        New node: user:ben
        Draw:  user:ben --admin--> org:acme

Fact 5: (user:carol, employee, org:acme)
        New node: user:carol
        Draw:  user:carol --employee--> org:acme
```

That is the entire algorithm: **any ID ever mentioned becomes a node —
mentioning the same ID in two facts is what connects them — and each fact
becomes one arrow.** No inference, no merging, no direction-flipping. The
graph *is* the pile of facts, drawn. In code it is five `AddEdge` calls.

## Two rules that look alike but do different jobs

Here is the subtlety that causes most of the arrow confusion, so it gets
its own section. There are **two separate rules**, and they operate at
different moments:

| | The arrow rule | The child-points-at-parent rule |
|---|---|---|
| What it governs | Converting a written fact into an edge | *Writing* the fact in the first place |
| When it runs | Every fact, mechanically | Only when you author a containment fact |
| Choices involved | **None** — subject always gets the tail | **One** — which node goes in the `subject` slot |
| Can it be violated? | No — it just runs | Yes — and the graph won't stop you |

**The arrow rule is a converter with no choices in it.** Whatever you put
in the `subject` slot gets the tail of the arrow, period. It cannot be
"gotten wrong" — only fed good or bad input.

**The child-points-at-parent rule is about what to feed it.** When a fact
has an actor ("Ed accesses the account"), the subject slot picks itself —
the doer. But a containment fact ("the account belongs to Acme") has no
actor, so *both* of these are legal inputs to the converter:

```yaml
# Option A: child in the subject slot
- subject: bankaccount:daytoday      # → arrow: account --org--> org:acme
  relation: org
  resource: org:acme

# Option B: parent in the subject slot
- subject: org:acme                  # → arrow: org:acme --owns--> account
  relation: owns
  resource: bankaccount:daytoday
```

Notice the arrow rule works flawlessly in **both** cases — each produces a
perfectly valid arrow that faithfully records the fact. The graph does not
reject Option B. The *check* does: subject = tail = the node the edge is
stored on (the adjacency-list entry — or, later in this doc, the DynamoDB
partition key). A check stands on the **account** asking "which org do I
belong to?" With Option A, the account's own edge list answers in one
lookup. With Option B, the fact is stored under the org's key — the
account has no outgoing edge, and the fact is invisible from where the
walk stands. Recorded, true, and unreachable.

So the two rules compose like this:

```text
child-points-at-parent  =  "for container facts, put the CHILD in the subject slot"
          ↓ feed into
subject → resource      =  "the subject gets the tail of the arrow"
          ↓ result
the arrow points UP, from child to parent — stored exactly where
the check will stand
```

And the naming convention (Session 4, Convention 2) is the third piece of
the same chain: once the subject slot is decided, the relation name must
read truthfully *in that direction* — which is what the read-aloud test
checks. "The account's org: Acme" ✓ confirms the slots are right; "the
account owns Acme" ✗ tells you the slots got swapped.

**One principle sits underneath all three:** the `subject` slot is where
the fact will be *looked up from*, so it must be the node a check stands
on when it needs that fact. Arrow direction, edge placement, and name
readability all fall out of that single decision. When in doubt, ask: "at
check time, which node will be standing here, needing this fact?" — that
node is your subject.

## What an arrow means (and does not mean)

Do not read arrows as "who has power over whom" or "which way access
flows" — they mean neither. An arrow means exactly one thing: **"standing
on the tail node, this fact is discoverable."** The walk only ever moves
*with* arrows, never against them.

Watch Ben's check with that reading:

```text
bankaccount:daytoday --org--> org:acme <--admin-- user:ben
```

Both arrows point *at* `org:acme` — neither points at Ben, neither points
at the account, and nothing walks backwards. The walk starts at both ends
(resource and subject), each follows its own outgoing arrow one step, and
access is granted because they *converge* on the same node.

**Quick self-test** for any fact: say it aloud in its naming style — verb
style left-to-right ("Ed is an accessor of the account" ✓), noun style
right-to-left ("the account's org: Acme" ✓). If the sentence comes out
false ("the account owns Acme" ✗), subject and resource are swapped.

## In production, facts live in a database

`relationships.yaml` is a stand-in. It is YAML only so this lab has no
database dependency. In a real system (Zanzibar, OpenFGA, SpiceDB all work
this way) the facts are rows in a datastore, shaped exactly like the
triple:

```text
| subject              | relation | resource             |
|----------------------|----------|----------------------|
| user:ed              | accessor | bankaccount:daytoday |
| bankaccount:daytoday | org      | org:acme             |
| user:ben             | admin    | org:acme             |
```

The **policy stays a file** even in production — YAML or a schema DSL. It
is authorization *design*: it lives in git, goes through code review, and
ships like code. The facts cannot work that way; they change on every
business event and are written by machines.

And here is the part that dissolves "how do all the facts become a graph"
for good: **in production the graph is never actually built.** There is no
load-everything-into-memory step. The graph exists only conceptually — the
evaluator walks it by issuing point queries against the store, one hop at
a time. Look at what the degenerate walk (Session 5) actually demands per
check — at most two questions:

1. **Prefix query:** "give me the outgoing edges of node X with relation
   R" (hop 1 of `via_relationship`).
2. **Point lookup:** "does the exact triple (X, R, Y) exist?" (`this`, and
   hop 2).

No joins, no scans, no recursion. Each "follow an arrow" is one indexed
lookup. Any store that can answer those two questions quickly can hold the
facts — a SQL table indexed on `(subject, relation)`, or a key-value store.

### Example: the facts in DynamoDB

The triple maps straight onto a partition-key/sort-key design:

```text
Table: relationships
  PK (partition key) = subject
  SK (sort key)      = relation#resource

| PK                   | SK                            |
|----------------------|-------------------------------|
| user:ed              | accessor#bankaccount:daytoday |
| user:ed              | employee#org:acme             |
| bankaccount:daytoday | org#org:acme                  |
| user:ben             | admin#org:acme                |
```

The satisfying part: **an item collection — all items sharing one
partition key — *is* the adjacency list from Session 1.** Not an analogy;
the same structure. `Query PK = "user:ed"` returns exactly what
`graph.Neighbors(user:ed)` returns in the lab code. Ben's check becomes:

```text
Hop 1: Query    PK = "bankaccount:daytoday", SK begins_with "org#"
       → org#org:acme
Hop 2: GetItem  PK = "user:ben", SK = "admin#org:acme"
       → found → allow
```

Two fast point operations per rule, cost independent of graph size — the
degenerate walk cashed out in database calls.

Two things a key-value store makes *more* true, not less:

- **The arrow-direction convention becomes a hard constraint.** In SQL,
  a backwards fact degrades to a slow query you could still run. In a
  key-value store there is no slow fallback: if the edge lives under the
  parent's partition key, the check standing on the child *cannot find
  it*. "Store the edge where evaluation starts" is enforced by the
  database's own physics.
- **The reverse index gets a concrete name.** Session 4's footnote said
  production systems answer reverse questions ("list everything Ben can
  access") with a second index. Here that is literally a secondary index
  with the keys flipped — `PK = resource, SK = relation#subject`. Not
  needed for `check`; added the day someone asks for a "who has access to
  this?" screen.

## The write path: a relationship API

Product teams never touch the table directly and never see the policy.
They call a small write API on business events — account opened, role
granted, employee offboarded:

```text
POST   /relationships { subject: "user:ed",  relation: "accessor", resource: "bankaccount:daytoday" }
DELETE /relationships { subject: "user:ben", relation: "admin",    resource: "org:acme" }
```

This maps 1:1 to Zanzibar's `Write` RPC and OpenFGA's write API. The API's
job is validation — does the policy even declare an `accessor` relation on
type `bankaccount`? is the triple well-formed? — it writes facts but
assigns them no meaning. Meaning stays in the policy, exactly as Session
4's two-files split teaches.

## The full production loop

```text
product team ──(business event)──> Relationship API ──> facts store   (rows = edges, changes constantly)
platform team ──(git PR, review)──> policy file                       (path shapes, changes rarely)

app asks:   check(user:ben, access, bankaccount:daytoday)
evaluator:  reads policy → picks rules → 1-2 point queries per rule → allow/deny
```

The lab and production differ only in where the facts sit — an in-memory
adjacency list here, a database there. The model is identical: a fact is
one triple, one row, one arrow; the policy names which path shapes count;
and a check is a degenerate walk, one indexed lookup per hop.

## What to remember

1. The arrow always goes subject → resource; converting facts to a graph
   is mechanical, with no judgment calls.
2. The judgment happens when *writing* a fact: the arrow rule and the
   child-points-at-parent rule are different rules at different moments —
   the converter has no choices; choosing the subject slot is the choice.
3. One principle decides that choice: the subject is the node a check
   will stand on when it needs this fact (the doer, or the child).
4. An arrow means "discoverable from the tail node" — nothing about power
   or access flow, and walks never go backwards.
5. In production, facts are database rows behind a write API; only the
   policy stays a reviewed file; the graph is walked via point queries,
   never built.

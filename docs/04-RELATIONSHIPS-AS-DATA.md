# Session 4: Relationships as data

Sessions 1-3 hard-coded both the graph (`AddEdge` calls) and the allowed
relation path (`AddRule("edit", "member_of", "editor_of", "contains")`) in
Go. Real authorization usually needs both to be declared and changed
without recompiling. `internal/policy` does that with two YAML files, and
this session covers the first of them: **relationships as data** — what
the facts are, how to write them down, and the conventions that keep the
resulting graph usable. [Session 5](05-RULES-AS-DATA.md) covers the second
file, the rules.

It is still the same graph from Session 1 — nodes, directed labelled
edges, and a walk over them. Nothing new in graph theory, only in how the
data is written down.

## Two files, two jobs

The split into two files is itself the lesson:

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

- Delete the rule that grants org admins access from the policy. Ben
  still "has the admin role" — his `admin` edge is untouched in
  `relationships.yaml` — but `check(user:ben, access, ...)` now denies.
  The fact survived; its *meaning* was revoked.
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
account by any relation. Session 5 writes the rules that turn those shapes
into allow/deny answers.

If the arrow directions feel confusing, or you want to see the five facts
above become the graph one edge at a time — and learn where the facts live
in production (a database behind a write API, not YAML) — read the
companion doc [From facts to a graph](FACTS-TO-GRAPH.md). Its first half
is worth reading now; its production half lands best after Session 5.

## Modeling conventions

Two decisions come up every time you add a relationship: **which node gets
the edge**, and **what to call the relation**. Neither is enforced by the
code — the graph will happily store anything — so the conventions below
are what keep the data usable.

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
Session 1 that the graph is an **adjacency list**: each node stores its own
*outgoing* edges. That means the evaluator can cheaply answer one shape of
question — "what does this node point at?" — and cannot cheaply answer the
reverse — "what points at this node?" — without scanning the entire graph.

Now look at what an indirect check needs to do, starting from a
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

Note this convention is *not* a second arrow rule, and it does not bend
the first one. The arrow still goes subject → resource, mechanically, for
every fact. This convention decides which node to put in the `subject`
slot *before* that mechanical rule runs. The distinction is subtle enough
that the companion doc gives it a full section —
["Two rules that look alike but do different jobs"](FACTS-TO-GRAPH.md#two-rules-that-look-alike-but-do-different-jobs)
— including what goes wrong (a true fact that checks cannot find) when
the parent is put in the subject slot instead.

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

**Noun style** (Sessions 4-5: `accessor`, `admin`, `employee`, `org`) names
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

## What to remember

1. `relationships.yaml` holds facts (terrain); `policy.yaml` holds meaning
   (map legend). Access exists only where they meet.
2. A relationship is one triple, one edge: `(subject, relation, resource)`.
3. Container facts go on the child, pointing at the parent — store the
   edge where evaluation starts.
4. Pick one relation-naming style per schema and read every triple aloud
   to check it.

Run it (loads both YAML files and prints, for each check, exactly which of
these edges the evaluator walked — Session 5 explains the rules driving
those walks):

```bash
go run ./cmd/policy-demo
```

Next: [Session 5 — rules as data](05-RULES-AS-DATA.md), which gives these
facts their meaning.

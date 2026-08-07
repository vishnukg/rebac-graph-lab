# Mapping to the Zanzibar paper

Reference, not a lesson — read it after Session 5, when you want to
connect this repo's vocabulary to the wider ReBAC world.

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
| action (`access`, `view`, `edit`)  | relation with a userset rewrite rule | Zanzibar does not distinguish a stored fact from a computed answer by name — both are just "relations," some backed by stored tuples, some by a rewrite rule. This repo splits that in two for teaching clarity: `relations:` are stored (matches Zanzibar's plain, tuple-backed relations), `actions:` are computed (matches Zanzibar's rewrite-rule relations). Note this is *not* the everyday sense of "permission" as a grant — see the naming note in [Session 5](05-RULES-AS-DATA.md). |
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

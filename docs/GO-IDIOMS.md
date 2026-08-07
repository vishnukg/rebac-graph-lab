# The Go idioms used in this project

Read this if a line of code is confusing for *Go* reasons rather than *graph*
reasons. Each entry shows an idiom exactly as it appears in this repository,
says what it means, and why the code uses it. Skim now, return when needed.

## 1. A named type for a plain string

```go
type NodeID string
```

`NodeID` is just a `string` with a new name. Two reasons to bother:

- **Documentation.** `func (g *Graph) BFS(start NodeID)` tells you the string
  is a node name, not any old text.
- **Safety.** The compiler stops you passing some unrelated string where a
  node ID belongs, unless you convert it on purpose with `NodeID("...")`.

## 2. A map of slices — and its friendly zero values

```go
edges map[NodeID][]Edge
```

Read it inside-out: `[]Edge` is "a slice (growable list) of edges";
`map[NodeID][]Edge` is "for each node ID, such a list."

The idiom that makes the traversal code short is what happens on a **miss**.
Reading a key that is not in a map returns the value type's *zero value* —
for a slice, `nil`. And looping over a `nil` slice with `range` simply runs
zero times. So this line in `BFS`:

```go
for _, edge := range g.edges[current] {
```

is safe even for a node with no outgoing edges — no `if` needed. Likewise
`visited[edge.To]` on a map of `map[NodeID]bool` returns `false` for any node
never marked, which is exactly what "not visited yet" should mean.

## 3. The comma-ok test: "is this key present?"

```go
if _, exists := g.edges[start]; !exists {
    return nil
}
```

A map read can return two values: the value and a bool saying whether the key
was present. Use this form when you must distinguish "key absent" from "key
present with a zero value" — here, "unknown node" versus "known node with no
edges". The `_` discards the value because only the bool matters.

## 4. Multiple return values instead of null

```go
path, found := g.FindPath("A", "E")
```

Where Python returns `None` and TypeScript returns `null`, Go functions
return the result *and* a bool. The signature `([]NodeID, bool)` announces
that the search can fail, and the caller has to receive that bool — failure
is visible, not something you remember to check. In this project that
matters: "no path" is the deny case, the most important case in
authorization.

## 5. A slice as a queue

```go
current := queue[0]      // peek at the front
queue = queue[1:]        // drop the front
queue = append(queue, edge.To)  // push onto the back
```

Go's standard library has no queue type, and for small programs nobody
misses it: take from index 0, re-slice to drop it, `append` to the back.
It is not the fastest queue in the world, but every step of the mechanism is
visible — which is the point in a teaching codebase.

## 6. Copy on the way out, copy on the way in

Two defensive copies guard the data structures:

```go
// Neighbors: copy on the way OUT
result := make([]Edge, len(g.edges[id]))
copy(result, g.edges[id])
return result
```

If `Neighbors` returned the internal slice directly, a caller could write
into it and silently corrupt the graph. Handing out a copy makes that
impossible.

```go
// AddRule: copy on the way IN
rule := append(Rule(nil), relations...)
```

`append(Rule(nil), xs...)` is a compact "clone this slice" — start from an
empty slice, append every element. Without it, the evaluator would keep a
reference to the *caller's* slice, and the caller could mutate a rule after
adding it.

The shared idea: slices are references to shared underlying memory, so a
struct that wants to own its data must copy at the boundary.

## 7. A recursive closure needs a two-step declaration

```go
var visit func(NodeID)
visit = func(current NodeID) {
    ...
    visit(edge.To)   // calls itself
}
```

This looks odd in every language that isn't Go. A function literal cannot
refer to itself by name while it is being defined, so you first *declare* the
variable (`var visit func(NodeID)`), then *assign* the function to it — by
the time the body runs, `visit` exists and the recursive call works. Purely
a Go quirk; the DFS algorithm underneath is the same as anywhere else.

## 8. Pointer receivers: methods that change the struct

```go
func (g *Graph) AddEdge(...)
```

The `(g *Graph)` part makes `AddEdge` a method on `*Graph`, a *pointer* to a
Graph. The pointer matters: the method modifies the actual graph rather than
a throwaway copy of it. This project uses pointer receivers on every method
for consistency, which is the common Go convention once any method needs one.

## 9. Zero values make deny-by-default free

```go
type Decision struct {
    Allowed bool
    Path    []graph.Step
    Reason  string
}

return Decision{Reason: "denied: no allowed relationship path was found"}
```

Any field not mentioned in a struct literal gets its zero value — `false`
for bools, `nil` for slices. So a `Decision` that only sets `Reason` is
automatically `Allowed: false` with no path. The language default and the
security default line up: you cannot forget to deny.

---

That is every idiom this codebase relies on. If a line is still confusing
after this list, the confusion is probably about the graph algorithm — which
is the session docs' job to fix, starting from
[Session 1](01-GRAPH-BASICS.md).

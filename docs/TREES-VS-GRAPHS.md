# Trees and graphs

A common question while learning this material: "isn't this just a tree?"

## The short answer

A tree is a special, restricted kind of graph.

```text
Every tree is a graph.
Not every graph is a tree.
```

## What makes a tree a tree

```text
        Company
       /       \
 Engineering   Sales
    /    \
 Alice   Bob
```

A tree obeys three rules:

- no loops;
- everything is connected;
- there is exactly **one** path between any two nodes.

That last rule is the important one here. Folder hierarchies usually qualify:

```text
root folder
└── product folder
    └── roadmap document
```

Every document lives in exactly one place, so there is exactly one path from
the root to it.

## What real permission data looks like

Now look at the graph from this project's tests:

```text
Alice --member_of--> Engineering
  |                       |
viewer_of             editor_of
  |                       |
  +-----> Roadmap <-------+
```

Alice reaches the Roadmap two different ways: directly as a viewer, and
indirectly through her team. Two paths between the same two nodes — so this
is not a tree, and no amount of rearranging will make it one.

This is the normal case, not the exception. A user joins several teams. A
team can edit several folders. A document has both direct and inherited
access. Relationships overlap, and overlapping relationships are exactly what
a general graph can express and a tree cannot.

## Both can be true at once

One *part* of a system may still be tree-shaped — the folder hierarchy in
this project is a tree on its own. But folders plus users plus teams plus
relationships, taken together, form a graph. That is why the code in
`internal/graph` implements a graph and not a tree: the general structure
handles the tree-shaped parts for free, but not the other way around.

> A tree describes one hierarchy.
> A graph describes many overlapping relationships.

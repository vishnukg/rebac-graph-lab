# Trees and graphs

## The short answer

A tree is a special kind of graph.

```text
Every tree is a graph.
Not every graph is a tree.
```

## A tree

```text
        Company
       /       \
 Engineering   Sales
    /    \
 Alice   Bob
```

In a tree:

- there are no loops;
- everything is connected;
- there is only one path between two nodes.

Folder hierarchies are often trees:

```text
root folder
└── product folder
    └── roadmap document
```

## A general graph

```text
Alice --member_of--> Engineering
  |                       |
viewer_of             editor_of
  |                       |
  +-----> Roadmap <-------+
```

This is not a tree. Alice has two different paths to the Roadmap.

A general graph may have several paths, crossed connections, or loops.

## Why ReBAC uses a graph

A user can join several teams. A team can access several documents. A document
can have both direct and inherited access. These overlapping relationships do
not fit into one tree.

One part of a system, such as its folders, may still be tree-shaped. But all the
users, teams, resources, and relationships together form a graph.

Remember:

> A tree describes one hierarchy. A graph can describe many overlapping relationships.

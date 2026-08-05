// Command traversal makes BFS, DFS, and path-finding visible.
package main

import (
	"fmt"

	"example.com/rebac-graph-lab/internal/graph"
)

func main() {
	g := graph.New()
	g.AddEdge("A", "connects", "B")
	g.AddEdge("A", "connects", "C")
	g.AddEdge("B", "connects", "D")
	g.AddEdge("C", "connects", "E")
	g.AddEdge("D", "connects", "A") // This edge creates a cycle.

	fmt.Println("Graph:")
	fmt.Println("    A")
	fmt.Println("   / \\")
	fmt.Println("  B   C")
	fmt.Println("  |   |")
	fmt.Println("  D   E")
	fmt.Println("  |")
	fmt.Println("  +----> A  (cycle)")
	fmt.Println()

	fmt.Println("BFS from A:", g.BFS("A"))
	fmt.Println("DFS from A:", g.DFS("A"))

	path, found := g.FindPath("A", "E")
	fmt.Println("Path from A to E:", path, "found:", found)
}

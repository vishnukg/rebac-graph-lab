package main

import (
	"fmt"

	"example.com/rebac-graph-lab/internal/graph"
	"example.com/rebac-graph-lab/internal/rebac"
)

func main() {
	g := graph.New()

	// Build this relationship graph:
	// Alice --member_of--> Engineering --editor_of--> Product --contains--> Roadmap
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "folder:product")
	g.AddEdge("folder:product", "contains", "document:roadmap")

	evaluator := rebac.NewEvaluator(g)
	evaluator.AddRule("edit", "member_of", "editor_of", "contains")

	decision := evaluator.Check("user:alice", "edit", "document:roadmap")
	fmt.Printf("Can Alice edit the roadmap? %v\n", decision.Allowed)
	fmt.Println(decision.Reason)
	for _, step := range decision.Path {
		fmt.Printf("  %s --%s--> %s\n", step.From, step.Relation, step.To)
	}
}

package main

import (
	"fmt"
	"log"

	"example.com/rebac-graph-lab/internal/graph"
	"example.com/rebac-graph-lab/internal/policy"
)

func main() {
	g, _, err := policy.LoadRelationships("examples/bank/relationships.yaml")
	if err != nil {
		log.Fatal(err)
	}

	p, err := policy.LoadPolicy("examples/bank/policy.yaml")
	if err != nil {
		log.Fatal(err)
	}

	evaluator := policy.NewEvaluator(g, p)

	checks := []struct {
		subject  graph.NodeID
		action   string
		resource graph.NodeID
	}{
		{"user:ed", "access", "bankaccount:daytoday"},    // direct accessor
		{"user:ben", "access", "bankaccount:daytoday"},   // admin of the org
		{"user:carol", "access", "bankaccount:daytoday"}, // employee only, no path
	}

	for _, c := range checks {
		decision := evaluator.Check(c.subject, c.action, c.resource)
		fmt.Printf("check(%s, %s, %s) = %v\n", c.subject, c.action, c.resource, decision.Allowed)
		fmt.Printf("  %s\n", decision.Reason)
		for _, step := range decision.Path {
			fmt.Printf("  walked: %s --%s--> %s\n", step.From, step.Relation, step.To)
		}
		fmt.Println()
	}
}

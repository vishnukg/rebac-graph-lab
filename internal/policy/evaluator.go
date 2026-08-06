package policy

import (
	"fmt"
	"strings"

	"example.com/rebac-graph-lab/internal/graph"
)

// Evaluator answers check(subject, action, resource) by walking the graph
// according to the rules in a Policy.
type Evaluator struct {
	graph  *graph.Graph
	policy *Policy
}

// Decision includes both the answer and, when allowed, the exact edges the
// evaluator walked to reach it. For a "this" rule the path is one step; for
// a "via_relationship" rule it is two steps that end on the same parent
// node. Printing the steps makes the graph walk visible.
type Decision struct {
	Allowed bool
	Path    []graph.Step
	Reason  string
}

// NewEvaluator pairs a relationship graph with the policy that interprets it.
func NewEvaluator(g *graph.Graph, p *Policy) *Evaluator {
	return &Evaluator{graph: g, policy: p}
}

// Check asks: "Can subject perform action on resource?"
//
// The resource's type (the part of its ID before ":") selects which type
// definition in the policy applies. The action name must match an action
// defined on that type. Each rule for that action is tried in order; the
// first one that matches grants access.
func (e *Evaluator) Check(subject graph.NodeID, action string, resource graph.NodeID) Decision {
	resourceType := typeOf(resource)

	typeDef, ok := e.policy.Types[resourceType]
	if !ok {
		return Decision{Reason: fmt.Sprintf("denied: policy has no type %q", resourceType)}
	}

	rules, ok := typeDef.Actions[action]
	if !ok {
		return Decision{Reason: fmt.Sprintf("denied: type %q has no %q action", resourceType, action)}
	}

	for _, rule := range rules {
		if decision, matched := e.tryRule(subject, rule, resource); matched {
			return decision
		}
	}

	return Decision{Reason: fmt.Sprintf("denied: no rule for %q connects %s to %s", action, subject, resource)}
}

// tryRule evaluates one rule and reports whether it granted access.
func (e *Evaluator) tryRule(subject graph.NodeID, rule Rule, resource graph.NodeID) (Decision, bool) {
	if rule.This != "" {
		if e.hasEdge(subject, rule.This, resource) {
			return Decision{
				Allowed: true,
				Path:    []graph.Step{{From: subject, Relation: rule.This, To: resource}},
				Reason:  fmt.Sprintf("allowed: %s --%s--> %s", subject, rule.This, resource),
			}, true
		}
	}

	if rule.ViaRelationship != nil {
		via := rule.ViaRelationship
		// Step 1: find the resource's parent object, e.g.
		// bankaccount:daytoday --org--> org:acme.
		for _, edge := range e.graph.Neighbors(resource) {
			if edge.Relation != via.Through {
				continue
			}
			parent := edge.To

			// Step 2: check whether the subject has the required relation
			// to that same parent, e.g. user:ben --admin--> org:acme.
			if e.hasEdge(subject, via.Requires, parent) {
				steps := []graph.Step{
					{From: resource, Relation: via.Through, To: parent},
					{From: subject, Relation: via.Requires, To: parent},
				}
				return Decision{
					Allowed: true,
					Path:    steps,
					Reason: fmt.Sprintf("allowed via: %s --%s--> %s and %s --%s--> %s",
						resource, via.Through, parent, subject, via.Requires, parent),
				}, true
			}
		}
	}

	return Decision{}, false
}

// hasEdge reports whether an exact relationship exists in the graph.
func (e *Evaluator) hasEdge(from graph.NodeID, relation string, to graph.NodeID) bool {
	for _, edge := range e.graph.Neighbors(from) {
		if edge.Relation == relation && edge.To == to {
			return true
		}
	}
	return false
}

// typeOf returns the part of a node ID before its first colon, e.g.
// typeOf("bankaccount:daytoday") == "bankaccount".
func typeOf(id graph.NodeID) string {
	s := string(id)
	if i := strings.Index(s, ":"); i >= 0 {
		return s[:i]
	}
	return s
}

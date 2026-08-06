// Package rebac implements a deliberately small relationship-based access
// control evaluator on top of the graph package.
package rebac

import (
	"fmt"
	"strings"

	"example.com/rebac-graph-lab/internal/graph"
)

// Rule is a relationship path that grants an action.
// Example: member_of -> editor_of grants "edit".
type Rule []string

// Evaluator answers authorization questions by looking for allowed paths.
type Evaluator struct {
	graph *graph.Graph
	rules map[string][]Rule
}

// Decision includes both the answer and the graph path that explains it.
type Decision struct {
	Allowed bool
	Path    []graph.Step
	Reason  string
}

// NewEvaluator creates an evaluator for a relationship graph.
func NewEvaluator(g *graph.Graph) *Evaluator {
	return &Evaluator{
		graph: g,
		rules: make(map[string][]Rule),
	}
}

// AddRule says that an action is granted by this exact relationship path.
// Multiple rules may grant the same action.
func (e *Evaluator) AddRule(action string, relations ...string) {
	// Copy the slice so a caller cannot change the rule after adding it.
	rule := append(Rule(nil), relations...)
	e.rules[action] = append(e.rules[action], rule)
}

// Check asks: "Can this subject perform this action on this resource?"
func (e *Evaluator) Check(subject graph.NodeID, action string, resource graph.NodeID) Decision {
	rules := e.rules[action]
	if len(rules) == 0 {
		return Decision{Reason: fmt.Sprintf("denied: action %q has no rules", action)}
	}

	for _, rule := range rules {
		path, found := e.graph.FindPathByRelations(subject, resource, rule)
		if found {
			return Decision{
				Allowed: true,
				Path:    path,
				Reason:  "allowed by relationship path: " + strings.Join(rule, " -> "),
			}
		}
	}

	return Decision{Reason: "denied: no allowed relationship path was found"}
}

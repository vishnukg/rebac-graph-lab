// Package policy adds declarative, YAML-defined authorization rules on top
// of the graph package. Where internal/rebac hard-codes an exact relation
// path in Go, this package reads which paths are allowed from a policy
// file, so the rules can change without recompiling anything.
package policy

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ViaRelationship means: follow this resource's own Through relation to a
// parent object, then check whether the subject has the Requires relation
// to that same parent object.
//
// Example: a bankaccount --org--> org edge, plus a subject --admin--> org
// edge, together grant access. That is:
//
//	through:  org
//	requires: admin
type ViaRelationship struct {
	Through  string `yaml:"through"`
	Requires string `yaml:"requires"`
}

// Rule is one way to satisfy an action. Exactly one of its fields is set.
//
//   - This checks a direct relationship between the subject and the resource.
//   - ViaRelationship checks an indirect relationship through a parent object.
type Rule struct {
	This            string           `yaml:"this,omitempty"`
	ViaRelationship *ViaRelationship `yaml:"via_relationship,omitempty"`
}

// TypeDef describes one object type: the relations it can hold, and the
// actions computed from those relations. An action is allowed if any one
// of its rules is satisfied.
//
// An action is not a grant — nobody "has" it the way a subject has a
// relation. It is the named question Check's action argument asks: given
// the relationships that exist right now, can this subject do this? What
// looks like a granted permission in everyday language — "Ben has the
// admin role" — is a Relationship in this project, not an action; actions
// are the rules that read those relationships, not the relationships
// themselves.
type TypeDef struct {
	Relations []string          `yaml:"relations"`
	Actions   map[string][]Rule `yaml:"actions"`
}

// Policy is the full set of object types known to the evaluator.
type Policy struct {
	Types map[string]TypeDef `yaml:"types"`
}

// LoadPolicy reads a policy file such as examples/bank/policy.yaml.
func LoadPolicy(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

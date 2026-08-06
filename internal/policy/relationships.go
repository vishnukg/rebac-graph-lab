package policy

import (
	"os"

	"gopkg.in/yaml.v3"

	"example.com/rebac-graph-lab/internal/graph"
)

// Relationship is the basic unit of ReBAC data: a subject has a relation to
// a resource. Example: {Subject: "user:ed", Relation: "accessor", Resource: "bankaccount:daytoday"}
//
// This is the domain language, on purpose: not "tuple," not "fact" — a
// relationship, because that is what it represents and what the R in ReBAC
// stands for.
type Relationship struct {
	Subject  graph.NodeID `yaml:"subject"`
	Relation string       `yaml:"relation"`
	Resource graph.NodeID `yaml:"resource"`
}

// relationshipsFile mirrors the shape of a relationships.yaml file.
type relationshipsFile struct {
	Relationships []Relationship `yaml:"relationships"`
}

// LoadRelationships reads a relationships file such as
// examples/bank/relationships.yaml and turns each relationship into a
// directed, labelled graph edge: Subject --Relation--> Resource.
func LoadRelationships(path string) (*graph.Graph, []Relationship, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var f relationshipsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, nil, err
	}

	g := graph.New()
	for _, rel := range f.Relationships {
		g.AddEdge(rel.Subject, rel.Relation, rel.Resource)
	}
	return g, f.Relationships, nil
}

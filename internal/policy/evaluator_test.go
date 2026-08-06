package policy

import (
	"os"
	"reflect"
	"testing"

	"example.com/rebac-graph-lab/internal/graph"
)

// Each test follows the Arrange-Act-Assert pattern: first build the graph
// and policy, then make exactly one call under test, then check the result.

func bankPolicy() *Policy {
	return &Policy{
		Types: map[string]TypeDef{
			"org": {Relations: []string{"employee", "admin"}},
			"bankaccount": {
				Relations: []string{"accessor", "org"},
				Actions: map[string][]Rule{
					"access": {
						{This: "accessor"},
						{ViaRelationship: &ViaRelationship{Through: "org", Requires: "admin"}},
					},
				},
			},
		},
	}
}

func bankRelationships() *graph.Graph {
	g := graph.New()
	g.AddEdge("user:ed", "employee", "org:acme")
	g.AddEdge("user:ed", "accessor", "bankaccount:daytoday")
	g.AddEdge("bankaccount:daytoday", "org", "org:acme")
	g.AddEdge("user:ben", "admin", "org:acme")
	g.AddEdge("user:carol", "employee", "org:acme")
	return g
}

func TestDirectAccessorIsAllowed(t *testing.T) {
	e := NewEvaluator(bankRelationships(), bankPolicy())

	decision := e.Check("user:ed", "access", "bankaccount:daytoday")

	if !decision.Allowed {
		t.Fatalf("expected direct accessor to be allowed: %s", decision.Reason)
	}
	want := []graph.Step{
		{From: "user:ed", Relation: "accessor", To: "bankaccount:daytoday"},
	}
	if !reflect.DeepEqual(decision.Path, want) {
		t.Fatalf("Path = %v, want the single direct edge %v", decision.Path, want)
	}
}

func TestOrgAdminIsAllowedByIndirection(t *testing.T) {
	e := NewEvaluator(bankRelationships(), bankPolicy())

	decision := e.Check("user:ben", "access", "bankaccount:daytoday")

	if !decision.Allowed {
		t.Fatalf("expected org admin to be allowed via via_relationship: %s", decision.Reason)
	}
	want := []graph.Step{
		{From: "bankaccount:daytoday", Relation: "org", To: "org:acme"},
		{From: "user:ben", Relation: "admin", To: "org:acme"},
	}
	if !reflect.DeepEqual(decision.Path, want) {
		t.Fatalf("Path = %v, want the two converging edges %v", decision.Path, want)
	}
}

func TestPlainEmployeeIsDenied(t *testing.T) {
	e := NewEvaluator(bankRelationships(), bankPolicy())

	decision := e.Check("user:carol", "access", "bankaccount:daytoday")

	if decision.Allowed {
		t.Fatalf("plain employee must not have account access: %s", decision.Reason)
	}
}

func TestAdminOfADifferentOrgIsDenied(t *testing.T) {
	// Dana is an admin, but of Globex — not of the org this account
	// belongs to. via_relationship must require both edges to land on the
	// same org node.
	g := bankRelationships()
	g.AddEdge("user:dana", "admin", "org:globex")
	e := NewEvaluator(g, bankPolicy())

	decision := e.Check("user:dana", "access", "bankaccount:daytoday")

	if decision.Allowed {
		t.Fatalf("admin of an unrelated org must not have access: %s", decision.Reason)
	}
}

func TestViaRelationshipSkipsUnrelatedResourceEdges(t *testing.T) {
	// The account's first edge is not the "through" relation. The rule
	// must skip past it and still find the org edge.
	g := graph.New()
	g.AddEdge("bankaccount:daytoday", "branch", "branch:downtown")
	g.AddEdge("bankaccount:daytoday", "org", "org:acme")
	g.AddEdge("user:ben", "admin", "org:acme")
	e := NewEvaluator(g, bankPolicy())

	decision := e.Check("user:ben", "access", "bankaccount:daytoday")

	if !decision.Allowed {
		t.Fatalf("unrelated resource edges must not block the org rule: %s", decision.Reason)
	}
}

func TestUnknownTypeIsDenied(t *testing.T) {
	e := NewEvaluator(bankRelationships(), bankPolicy())

	decision := e.Check("user:ed", "access", "vault:secret")

	if decision.Allowed {
		t.Fatalf("unknown resource type must not be allowed: %s", decision.Reason)
	}
}

func TestUnknownActionIsDenied(t *testing.T) {
	e := NewEvaluator(bankRelationships(), bankPolicy())

	decision := e.Check("user:ed", "delete", "bankaccount:daytoday")

	if decision.Allowed {
		t.Fatalf("undefined action must not be allowed: %s", decision.Reason)
	}
}

func TestTypeOfSplitsAtTheFirstColon(t *testing.T) {
	tests := []struct {
		id   graph.NodeID
		want string
	}{
		{id: "bankaccount:daytoday", want: "bankaccount"},
		{id: "justaname", want: "justaname"},
	}

	for _, tt := range tests {
		got := typeOf(tt.id)

		if got != tt.want {
			t.Fatalf("typeOf(%s) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestLoadRelationshipsAndPolicyFromYAML(t *testing.T) {
	g, relationships, err := LoadRelationships("../../examples/bank/relationships.yaml")
	if err != nil {
		t.Fatalf("LoadRelationships: %v", err)
	}
	p, err := LoadPolicy("../../examples/bank/policy.yaml")
	if err != nil {
		t.Fatalf("LoadPolicy: %v", err)
	}
	e := NewEvaluator(g, p)

	decision := e.Check("user:ben", "access", "bankaccount:daytoday")

	if len(relationships) == 0 {
		t.Fatal("expected at least one relationship")
	}
	if !decision.Allowed {
		t.Fatalf("expected admin access from loaded files: %s", decision.Reason)
	}
}

func TestLoadRejectsMissingFiles(t *testing.T) {
	_, _, relErr := LoadRelationships("does-not-exist.yaml")
	_, polErr := LoadPolicy("does-not-exist.yaml")

	if relErr == nil {
		t.Fatal("expected an error for a missing relationships file")
	}
	if polErr == nil {
		t.Fatal("expected an error for a missing policy file")
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	bad := t.TempDir() + "/bad.yaml"
	if err := os.WriteFile(bad, []byte(":\tthis is not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, relErr := LoadRelationships(bad)
	_, polErr := LoadPolicy(bad)

	if relErr == nil {
		t.Fatal("expected an error for malformed relationships YAML")
	}
	if polErr == nil {
		t.Fatal("expected an error for malformed policy YAML")
	}
}

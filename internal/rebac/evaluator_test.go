package rebac

import (
	"testing"

	"example.com/rebac-graph-lab/internal/graph"
)

// Each test follows the Arrange-Act-Assert pattern: first build the graph
// and rules, then make exactly one call under test, then check the result.

func documentEvaluator() *Evaluator {
	g := graph.New()

	// Alice inherits edit access through her team and the folder.
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "folder:product")
	g.AddEdge("folder:product", "contains", "document:roadmap")

	// Bob has a direct viewer relationship with the document.
	g.AddEdge("user:bob", "viewer_of", "document:roadmap")

	e := NewEvaluator(g)
	e.AddRule("view", "viewer_of")
	e.AddRule("view", "member_of", "editor_of", "contains")
	e.AddRule("edit", "editor_of")
	e.AddRule("edit", "member_of", "editor_of", "contains")
	return e
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name      string
		subject   graph.NodeID
		action    string
		want      bool
		wantSteps int
	}{
		{name: "inherited editor can edit", subject: "user:alice", action: "edit", want: true, wantSteps: 3},
		{name: "inherited editor can also view", subject: "user:alice", action: "view", want: true, wantSteps: 3},
		{name: "direct viewer can view", subject: "user:bob", action: "view", want: true, wantSteps: 1},
		{name: "viewer cannot edit", subject: "user:bob", action: "edit", want: false},
		{name: "unrelated user is denied", subject: "user:carol", action: "view", want: false},
		{name: "unknown action is denied", subject: "user:alice", action: "delete", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := documentEvaluator()

			decision := e.Check(tt.subject, tt.action, "document:roadmap")

			if decision.Allowed != tt.want {
				t.Fatalf("Allowed = %v, want %v; reason: %s", decision.Allowed, tt.want, decision.Reason)
			}
			if len(decision.Path) != tt.wantSteps {
				t.Fatalf("len(Path) = %d, want %d; path: %v", len(decision.Path), tt.wantSteps, decision.Path)
			}
		})
	}
}

func TestWrongRelationDoesNotGrantAccess(t *testing.T) {
	g := graph.New()
	g.AddEdge("user:mallory", "follows", "document:roadmap")
	e := NewEvaluator(g)
	e.AddRule("view", "viewer_of")

	decision := e.Check("user:mallory", "view", "document:roadmap")

	if decision.Allowed {
		t.Fatalf("follows must not grant view access: %+v", decision)
	}
}

func TestAnyAllowedPathIsEnough(t *testing.T) {
	g := graph.New()
	g.AddEdge("user:alice", "viewer_of", "document:roadmap")
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "viewer_of", "document:roadmap")
	e := NewEvaluator(g)
	e.AddRule("view", "viewer_of")
	e.AddRule("view", "member_of", "viewer_of")

	decision := e.Check("user:alice", "view", "document:roadmap")

	if !decision.Allowed {
		t.Fatalf("expected either valid path to allow access: %+v", decision)
	}
	if len(decision.Path) != 1 {
		t.Fatalf("expected the first matching (direct) path, got %v", decision.Path)
	}
}

func TestNestedGroup(t *testing.T) {
	g := graph.New()
	g.AddEdge("user:alice", "member_of", "team:app")
	g.AddEdge("team:app", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "document:roadmap")
	e := NewEvaluator(g)
	e.AddRule("edit", "member_of", "member_of", "editor_of")

	decision := e.Check("user:alice", "edit", "document:roadmap")

	if !decision.Allowed {
		t.Fatalf("expected nested team membership to allow access: %+v", decision)
	}
	if len(decision.Path) != 3 {
		t.Fatalf("expected a three-step path, got %v", decision.Path)
	}
}

package graph

import (
	"reflect"
	"testing"
)

func learningGraph() *Graph {
	g := New()
	g.AddEdge("A", "knows", "B")
	g.AddEdge("A", "knows", "C")
	g.AddEdge("B", "knows", "D")
	g.AddEdge("C", "knows", "D")
	g.AddEdge("D", "knows", "A") // A cycle: D points back to A.
	return g
}

func TestNeighbors(t *testing.T) {
	g := learningGraph()

	want := []Edge{{To: "B", Relation: "knows"}, {To: "C", Relation: "knows"}}
	if got := g.Neighbors("A"); !reflect.DeepEqual(got, want) {
		t.Fatalf("Neighbors(A) = %v, want %v", got, want)
	}
}

func TestBFSVisitsLevelByLevelAndHandlesCycles(t *testing.T) {
	got := learningGraph().BFS("A")
	want := []NodeID{"A", "B", "C", "D"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BFS(A) = %v, want %v", got, want)
	}
}

func TestDFSExploresOneBranchFirstAndHandlesCycles(t *testing.T) {
	got := learningGraph().DFS("A")
	want := []NodeID{"A", "B", "D", "C"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DFS(A) = %v, want %v", got, want)
	}
}

func TestFindPathReturnsAShortestPath(t *testing.T) {
	got, found := learningGraph().FindPath("A", "D")
	want := []NodeID{"A", "B", "D"}

	if !found {
		t.Fatal("FindPath(A, D) did not find a path")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindPath(A, D) = %v, want %v", got, want)
	}
}

func TestFindPathByRelationsOnlyFollowsTheRequestedLabels(t *testing.T) {
	g := New()
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "document:roadmap")
	g.AddEdge("user:alice", "follows", "document:roadmap")

	got, found := g.FindPathByRelations(
		"user:alice",
		"document:roadmap",
		[]string{"member_of", "editor_of"},
	)
	want := []Step{
		{From: "user:alice", Relation: "member_of", To: "team:engineering"},
		{From: "team:engineering", Relation: "editor_of", To: "document:roadmap"},
	}

	if !found {
		t.Fatal("expected a matching relationship path")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("path = %v, want %v", got, want)
	}
}

func TestUnknownStartHasNoTraversal(t *testing.T) {
	g := learningGraph()
	if got := g.BFS("missing"); got != nil {
		t.Fatalf("BFS(missing) = %v, want nil", got)
	}
}

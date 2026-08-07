package graph

import (
	"reflect"
	"testing"
)

// Each test follows the Arrange-Act-Assert pattern: first build a graph,
// then make exactly one call under test, then check the result.

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

	got := g.Neighbors("A")

	want := []Edge{{To: "B", Relation: "knows"}, {To: "C", Relation: "knows"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Neighbors(A) = %v, want %v", got, want)
	}
}

func TestAddEdgeIgnoresExactDuplicates(t *testing.T) {
	g := New()

	g.AddEdge("A", "knows", "B")
	g.AddEdge("A", "knows", "B")

	if got := g.Neighbors("A"); len(got) != 1 {
		t.Fatalf("Neighbors(A) = %v, want exactly one edge", got)
	}
}

func TestBFSVisitsLevelByLevelAndHandlesCycles(t *testing.T) {
	g := learningGraph()

	got := g.BFS("A")

	want := []NodeID{"A", "B", "C", "D"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BFS(A) = %v, want %v", got, want)
	}
}

func TestDFSExploresOneBranchFirstAndHandlesCycles(t *testing.T) {
	g := learningGraph()

	got := g.DFS("A")

	want := []NodeID{"A", "B", "D", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DFS(A) = %v, want %v", got, want)
	}
}

// A deeper graph shows the orders diverging: BFS visits the shallow C
// before the deep D and E, while DFS dives to E before touching C.
//
//	A --> B --> D --> E
//	 \
//	  --> C
func deepGraph() *Graph {
	g := New()
	g.AddEdge("A", "knows", "B")
	g.AddEdge("A", "knows", "C")
	g.AddEdge("B", "knows", "D")
	g.AddEdge("D", "knows", "E")
	return g
}

func TestBFSVisitsShallowNodesBeforeDeepOnes(t *testing.T) {
	g := deepGraph()

	got := g.BFS("A")

	want := []NodeID{"A", "B", "C", "D", "E"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BFS(A) = %v, want %v", got, want)
	}
}

func TestDFSFinishesABranchBeforeStartingTheNext(t *testing.T) {
	g := deepGraph()

	got := g.DFS("A")

	want := []NodeID{"A", "B", "D", "E", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DFS(A) = %v, want %v", got, want)
	}
}

func TestTraversalOfAnEdgelessNodeVisitsJustThatNode(t *testing.T) {
	g := New()
	g.AddNode("lonely")

	bfs := g.BFS("lonely")
	dfs := g.DFS("lonely")

	want := []NodeID{"lonely"}
	if !reflect.DeepEqual(bfs, want) {
		t.Fatalf("BFS(lonely) = %v, want %v", bfs, want)
	}
	if !reflect.DeepEqual(dfs, want) {
		t.Fatalf("DFS(lonely) = %v, want %v", dfs, want)
	}
}

func TestFindPathReturnsAShortestPath(t *testing.T) {
	g := learningGraph()

	got, found := g.FindPath("A", "D")

	want := []NodeID{"A", "B", "D"}
	if !found {
		t.Fatal("FindPath(A, D) did not find a path")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindPath(A, D) = %v, want %v", got, want)
	}
}

func TestFindPathReportsWhenNoRouteExists(t *testing.T) {
	g := New()
	g.AddEdge("A", "knows", "B")
	g.AddEdge("C", "knows", "D") // C and D are not reachable from A.

	path, found := g.FindPath("A", "D")

	if found {
		t.Fatalf("FindPath(A, D) = %v, want no path", path)
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

func TestFindPathByRelationsRejectsAWrongLabel(t *testing.T) {
	g := New()
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "document:roadmap")

	path, found := g.FindPathByRelations("user:alice", "document:roadmap",
		[]string{"viewer_of"})

	if found {
		t.Fatalf("wrong label must not match, got %v", path)
	}
}

func TestFindPathByRelationsRejectsAWrongFinalNode(t *testing.T) {
	g := New()
	g.AddEdge("user:alice", "member_of", "team:engineering")
	g.AddEdge("team:engineering", "editor_of", "document:roadmap")

	path, found := g.FindPathByRelations("user:alice", "document:other",
		[]string{"member_of", "editor_of"})

	if found {
		t.Fatalf("wrong final node must not match, got %v", path)
	}
}

func TestUnknownStartHasNoTraversal(t *testing.T) {
	g := learningGraph()

	bfs := g.BFS("missing")
	dfs := g.DFS("missing")
	_, found := g.FindPath("missing", "A")

	if bfs != nil {
		t.Fatalf("BFS(missing) = %v, want nil", bfs)
	}
	if dfs != nil {
		t.Fatalf("DFS(missing) = %v, want nil", dfs)
	}
	if found {
		t.Fatal("FindPath(missing, A) must not find a path")
	}
}

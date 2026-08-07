// Package graph contains a small directed, labelled graph.
//
// Directed means an edge has a direction: A -> B is not the same as B -> A.
// Labelled means an edge says what the relationship is, such as "member_of".
package graph

// NodeID is the name of a vertex (node) in the graph.
// Examples: "user:alice", "team:engineering", "document:roadmap".
type NodeID string

// Edge connects the current node to To. Relation describes the connection.
type Edge struct {
	To       NodeID
	Relation string
}

// Step is one complete hop through the graph. It is useful when explaining
// why a path was found.
type Step struct {
	From     NodeID
	Relation string
	To       NodeID
}

// Graph stores each node's outgoing edges in an adjacency list.
//
// For example:
//
//	"user:alice" -> [{To: "team:engineering", Relation: "member_of"}]
type Graph struct {
	edges map[NodeID][]Edge
}

// New creates an empty graph.
func New() *Graph {
	return &Graph{edges: make(map[NodeID][]Edge)}
}

// AddNode registers a node that has no relationships yet, such as a new user
// who has not joined any team. It also lets traversal distinguish a known
// edgeless node from a name the graph has never seen. AddEdge adds its
// endpoint nodes, so most callers do not need to call AddNode themselves.
func (g *Graph) AddNode(id NodeID) {
	if _, exists := g.edges[id]; !exists {
		g.edges[id] = nil
	}
}

// AddEdge adds a one-way relationship from one node to another.
func (g *Graph) AddEdge(from NodeID, relation string, to NodeID) {
	g.AddNode(from)
	g.AddNode(to)

	// Avoid storing the exact same relationship twice.
	for _, edge := range g.edges[from] {
		if edge.To == to && edge.Relation == relation {
			return
		}
	}

	g.edges[from] = append(g.edges[from], Edge{To: to, Relation: relation})
}

// Neighbors returns the outgoing edges of a node. The returned slice is a
// copy, so callers cannot accidentally modify the graph.
func (g *Graph) Neighbors(id NodeID) []Edge {
	result := make([]Edge, len(g.edges[id]))
	copy(result, g.edges[id])
	return result
}

// BFS visits nodes breadth-first: first the start, then every node one edge
// away, then every node two edges away, and so on. Nodes the same number of
// edges from the start form a "level", and BFS finishes each level before
// starting the next.
//
// The queue (first in, first out) is what makes it breadth-first: nodes
// discovered early are visited before nodes discovered later. On the graph
// A->B, A->C, B->D, C->E, D->A the queue evolves like this:
//
//	visit A   queue [B C]   A's neighbors join the queue
//	visit B   queue [C D]   D lines up behind C, which was found first
//	visit C   queue [D E]
//	visit D   queue [E]     D's edge back to A is skipped: A is visited
//	visit E   queue []      done: [A B C D E], level by level
//
// Every node is queued at most once and every edge is examined at most once,
// so the work grows linearly with graph size: O(nodes + edges).
func (g *Graph) BFS(start NodeID) []NodeID {
	if _, exists := g.edges[start]; !exists {
		return nil
	}

	visited := map[NodeID]bool{start: true}
	queue := []NodeID{start}
	order := make([]NodeID, 0)

	for len(queue) > 0 {
		// Take the oldest waiting node from the front of the queue.
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, edge := range g.edges[current] {
			if !visited[edge.To] {
				// Mark when queued, not when visited. Otherwise two nodes
				// that share a neighbor could both queue it before either
				// visit happens, and it would appear in the result twice.
				visited[edge.To] = true
				queue = append(queue, edge.To)
			}
		}
	}

	return order
}

// DFS visits nodes depth-first: it follows one branch as far as it can, then
// backs up to the most recent fork and tries that fork's next branch.
//
// Where BFS uses a queue, DFS uses a stack (last in, first out). This
// implementation does not build one by hand: each recursive visit call pushes
// a frame onto Go's call stack, and the "back up" step is simply the function
// returning. On the graph A->B, A->C, B->D, C->E, D->A:
//
//	visit A         start
//	 visit B        follow A's first edge
//	  visit D       follow B's first edge
//	                D's edge back to A is skipped: A is visited
//	 visit C        D and B are exhausted, so back up and try A's next edge
//	  visit E
//
// Result: [A B D C E]. Compare BFS's [A B C D E] — the same nodes, but DFS
// reaches the deep node D before the shallow node C.
//
// Like BFS, DFS handles every node and edge at most once: O(nodes + edges).
func (g *Graph) DFS(start NodeID) []NodeID {
	if _, exists := g.edges[start]; !exists {
		return nil
	}

	visited := make(map[NodeID]bool)
	order := make([]NodeID, 0)

	var visit func(NodeID)
	visit = func(current NodeID) {
		// Marking on entry means a node reachable along two branches, or
		// through a cycle back to an ancestor, is visited exactly once.
		visited[current] = true
		order = append(order, current)

		for _, edge := range g.edges[current] {
			if !visited[edge.To] {
				visit(edge.To)
			}
		}
	}

	visit(start)
	return order
}

// FindPath uses BFS to find a shortest path (fewest edges) between two nodes.
func (g *Graph) FindPath(start, target NodeID) ([]NodeID, bool) {
	if _, exists := g.edges[start]; !exists {
		return nil, false
	}

	visited := map[NodeID]bool{start: true}
	queue := []NodeID{start}
	previous := make(map[NodeID]NodeID)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == target {
			return buildPath(previous, start, target), true
		}

		for _, edge := range g.edges[current] {
			if !visited[edge.To] {
				visited[edge.To] = true
				previous[edge.To] = current
				queue = append(queue, edge.To)
			}
		}
	}

	return nil, false
}

func buildPath(previous map[NodeID]NodeID, start, target NodeID) []NodeID {
	path := []NodeID{target}
	for current := target; current != start; {
		current = previous[current]
		path = append(path, current)
	}

	// We constructed target -> start, so reverse it to start -> target.
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

// FindPathByRelations follows an exact sequence of relationship labels.
// This is the bridge from ordinary graph traversal to ReBAC evaluation.
//
// Example: ["member_of", "editor_of", "contains"] means:
// user -> team -> folder -> document.
func (g *Graph) FindPathByRelations(start, target NodeID, relations []string) ([]Step, bool) {
	var walk func(current NodeID, relationIndex int, path []Step) ([]Step, bool)
	walk = func(current NodeID, relationIndex int, path []Step) ([]Step, bool) {
		if relationIndex == len(relations) {
			return path, current == target
		}

		wantedRelation := relations[relationIndex]
		for _, edge := range g.edges[current] {
			if edge.Relation != wantedRelation {
				continue
			}

			step := Step{From: current, Relation: edge.Relation, To: edge.To}
			if result, found := walk(edge.To, relationIndex+1, append(path, step)); found {
				return result, true
			}
		}
		return nil, false
	}

	return walk(start, 0, nil)
}

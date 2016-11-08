package resolver

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/simple"
	"github.com/mdy/melody/resolver/types"
	"sort"
	"strings"
)

// Graph of dependencies
type Graph struct {
	*simple.DirectedGraph
	namesToID map[string]int
}

func NewGraph() *Graph {
	simpleGraph := simple.NewDirectedGraph(1, 0)
	return &Graph{simpleGraph, map[string]int{}}
}

func (g *Graph) isEmpty() bool {
	return len(g.namesToID) == 0
}

func (g *Graph) Dup() *Graph {
	newGraph := NewGraph()

	// Dup all nodes by value
	for _, n := range g.Nodes() {
		vOld := n.(*Vertex)
		newGraph.AddNode(vOld.Dup())
		newGraph.namesToID[vOld.Name] = vOld.ID()
	}

	// Dup all edges with requirement
	for _, oldSrc := range g.Nodes() {
		for _, oldDst := range g.From(oldSrc) {
			newSrc := newGraph.Node(oldSrc.ID()).(*Vertex)
			newDst := newGraph.Node(oldDst.ID()).(*Vertex)
			oldReq := g.Edge(oldSrc, oldDst).(*Edge).Requirement
			newGraph.SetEdge(&Edge{newSrc, newDst, oldReq})
		}
	}

	return newGraph
}

func (g *Graph) DetachVertexNamed(name string) {
	g.detachVertexNamed(name)
}

func (g *Graph) detachVertexNamed(name string) {
	if id, ok := g.namesToID[name]; ok {
		delete(g.namesToID, name)
		byeNode := g.Node(id)

		// Save all outgoing nodes
		outgoing := g.From(byeNode)

		// Ditch node/edges
		g.RemoveNode(byeNode)

		// Remove any loose leafs
		for _, n := range outgoing {
			vertex := n.(*Vertex)
			if !vertex.Root && len(g.To(n)) == 0 {
				g.detachVertexNamed(vertex.Name)
			}
		}
	}
}

func (g *Graph) addVertex(name string, payload types.Specification, root bool) *Vertex {
	vertex := g.vertexNamed(name)
	if vertex == nil {
		vertex = &Vertex{id: g.NewNodeID(), Name: name}
		g.namesToID[name] = vertex.ID()
		g.AddNode(vertex)
	}

	if vertex.Payload == nil {
		vertex.Payload = payload
	}
	vertex.Root = vertex.Root || root
	return vertex
}

func (g *Graph) addChildVertex(name string, payload types.Specification, parents []string, req types.Requirement) (*Vertex, error) {
	vertex := g.addVertex(name, payload, false)
	for _, pName := range parents {
		if pName == "" {
			vertex.Root = true
		} else {
			pNode := g.vertexNamed(pName)
			if g.hasPath(vertex, pNode) {
				return nil, &CircularDependencyError{vertex, pNode}
			}
			g.SetEdge(&Edge{pNode, vertex, req})
		}
	}
	return vertex, nil
}

// Check if path exists between nodes (to prevent circular deps)
func (g *Graph) hasPath(src graph.Node, dst graph.Node) bool {
	if src.ID() == dst.ID() {
		return true
	}
	for _, n := range g.From(src) {
		if g.hasPath(n, dst) {
			return true
		}
	}
	return false
}

func (g *Graph) vertexNamed(name string) *Vertex {
	if id, ok := g.namesToID[name]; ok {
		return g.Node(id).(*Vertex)
	}
	return nil
}

func (g *Graph) rootVertexNamed(name string) *Vertex {
	if vertex := g.vertexNamed(name); vertex != nil && vertex.Root {
		return vertex
	}
	return nil
}

func (g *Graph) PayloadFor(name string) types.Specification {
	if vertex := g.vertexNamed(name); vertex != nil {
		return vertex.Payload
	}
	return nil
}

func (g *Graph) requirementsFor(name string) []types.Requirement {
	vertex := g.vertexNamed(name)
	requirements := vertex.ExplicitRequirements
	for _, neighbor := range g.To(vertex) {
		edge := g.Edge(neighbor, vertex).(*Edge)
		requirements = append(requirements, edge.Requirement)
	}
	return requirements
}

// Sorting of Vertices by name
type verticesByName []graph.Node

func (n verticesByName) Len() int      { return len(n) }
func (n verticesByName) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n verticesByName) Less(i, j int) bool {
	return n[i].(*Vertex).Name < n[j].(*Vertex).Name
}

// Activated vertices (with payload) vertices by name
func (g *Graph) ActivatedByName() map[string]types.Specification {
	active := make(map[string]types.Specification, len(g.namesToID))
	for _, node := range g.Nodes() {
		if v := node.(*Vertex); v.Payload != nil {
			active[v.Name] = v.Payload
		}
	}
	return active
}

// Return a sorted list of all specifications for testing
func (g *Graph) Specifications() []types.Specification {
	nodes := verticesByName(g.Nodes())
	sort.Sort(nodes)

	var specs []types.Specification
	for _, node := range nodes {
		if v := node.(*Vertex); v.Payload != nil {
			specs = append(specs, v.Payload)
		}
	}

	return specs
}

// List all specs for dumping and debugging
func (g *Graph) String() string {
	nodes := verticesByName(g.Nodes())
	var info []string
	sort.Sort(nodes)

	for _, node := range nodes {
		if v := node.(*Vertex); v.Payload == nil {
			info = append(info, "NoSpec("+v.Name+")")
		} else {
			info = append(info, v.Payload.String())
		}
	}

	return "Graph(" + strings.Join(info, " ") + ")"
}

// Implements graph.Edge interface
type Edge struct {
	from, to    *Vertex
	Requirement types.Requirement
}

func (e *Edge) From() graph.Node {
	return e.from
}

func (e *Edge) To() graph.Node {
	return e.to
}

func (e *Edge) Weight() float64 {
	return 1.0
}

// Implements graph.Node interface
type Vertex struct {
	id                   int
	Root                 bool
	Name                 string
	Payload              types.Specification
	ExplicitRequirements types.Requirements
}

func (v *Vertex) ID() int {
	return v.id
}

func (v *Vertex) Dup() *Vertex {
	vNew := &Vertex{}
	*vNew = *v // Shallow copy
	//vNew.ExplicitRequirements = v.ExplicitRequirements.Dup()
	return vNew
}

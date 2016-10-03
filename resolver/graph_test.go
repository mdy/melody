package resolver

import (
	gnum "github.com/gonum/graph"
	c "gopkg.in/check.v1"
)

func (s *MySuite) Test_Graph_General(t *c.C) {
	graph := NewGraph()
	graph.addVertex("Root2", nil, true)
	root1 := graph.addVertex("Root", nil, true)
	child, _ := graph.addChildVertex("Child", nil, []string{"Root"}, nil)

	// Test #rootVertexNamed
	t.Assert(graph.rootVertexNamed("Root"), c.Equals, root1)
	t.Assert(graph.rootVertexNamed("Child"), c.IsNil)
	t.Assert(graph.rootVertexNamed("Noop"), c.IsNil)

	// Test #vertexNamed
	t.Assert(graph.vertexNamed("Root"), c.Equals, root1)
	t.Assert(graph.vertexNamed("Child"), c.Equals, child)
	t.Assert(graph.vertexNamed("Noop"), c.IsNil)

	// Test #Dup (full copy)
	graph2 := graph.Dup()
	t.Assert(graph2, c.Not(c.Equals), graph)
	t.Assert(graph2.vertexNamed("Root"), c.Not(c.Equals), root1)
	t.Assert(graph2.vertexNamed("Root"), c.DeepEquals, root1)
	t.Assert(graph2.vertexNamed("Child"), c.Not(c.Equals), child)
	t.Assert(graph2.vertexNamed("Child"), c.DeepEquals, child)
	t.Assert(graph2.vertexNamed("Noop"), c.IsNil)
}

func (s *MySuite) Test_Graph_Circular(t *c.C) {
	graph := NewGraph()
	graph.addVertex("Root", nil, true)
	_, err := graph.addChildVertex("Foo", nil, []string{"Root"}, nil)
	t.Assert(err, c.IsNil)
	_, err = graph.addChildVertex("Bar", nil, []string{"Foo"}, nil)
	t.Assert(err, c.IsNil)
	_, err = graph.addChildVertex("Foo", nil, []string{"Bar"}, nil)
	t.Assert(err, c.FitsTypeOf, &CircularDependencyError{})
}

func (s *MySuite) Test_Graph_Detatch(t *c.C) {
	var graph *Graph
	var root, root2, child *Vertex

	// Detaches a root vertex without successors
	graph = NewGraph()
	root = graph.addVertex("root", nil, true)
	graph.detachVertexNamed(root.Name)
	t.Assert(graph.vertexNamed(root.Name), c.IsNil)
	t.Assert(graph.isEmpty(), c.Equals, true)

	// Detaches a root vertex with successors
	graph = NewGraph()
	root = graph.addVertex("root", nil, true)
	child, _ = graph.addChildVertex("child", nil, []string{"root"}, nil)
	graph.detachVertexNamed(root.Name)
	t.Assert(graph.vertexNamed(root.Name), c.IsNil)
	t.Assert(graph.vertexNamed(child.Name), c.IsNil)
	t.Assert(graph.isEmpty(), c.Equals, true)

	// Detaches a root vertex with successors with other parents
	graph = NewGraph()
	root = graph.addVertex("root", nil, true)
	root2 = graph.addVertex("root2", nil, true)
	child, _ = graph.addChildVertex("child", nil, []string{"root", "root2"}, nil)
	graph.detachVertexNamed(root.Name)
	t.Assert(graph.vertexNamed(root.Name), c.IsNil)
	t.Assert(graph.vertexNamed(child.Name), c.Equals, child)
	t.Assert(graph.To(child), c.DeepEquals, []gnum.Node{root2})

	// Detaches a vertex with predecessors
	graph = NewGraph()
	root = graph.addVertex("root", nil, true)
	child, _ = graph.addChildVertex("child", nil, []string{"root"}, nil)
	graph.detachVertexNamed(child.Name)
	t.Assert(graph.vertexNamed(child.Name), c.IsNil)
	t.Assert(graph.vertexNamed(root.Name), c.Equals, root)
	t.Assert(len(graph.From(root)), c.Equals, 0)
}

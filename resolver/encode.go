package resolver

import (
	"encoding/json"
	"github.com/melodysh/melody/resolver/rubygem"
	"github.com/melodysh/melody/resolver/types"
	"sort"
	"strings"
)

type GraphEncoder interface {
	Encode(interface{}) error
}

type GraphDecoder interface {
	NewSpec(*GraphItem) (types.Specification, error)
	Decode(interface{}) error
}

// Public API for encoding/decoding graphs
type EncodedGraph struct {
	Project  *encodedItem   `toml:"project"`
	Packages []*encodedItem `toml:"packages,omitempty"`
	Version  string         `toml:"_lockFormatVersion,omitempty"`
}

type GraphItem struct {
	Name    string `toml:"name,omitempty"`
	Version string `toml:"version,omitempty"`
	Release string `toml:"release,omitempty"`
}

func (i *GraphItem) id() string {
	return i.Name + " " + i.Version
}

///////////////////////////////////////////////////
type revisioned interface {
	Revision() string
}

type released interface {
	ReleaseSpec() types.Specification
}

///////////////////////////////////////////////////

type encodedItem struct {
	GraphItem             // Keep here to preserve encoding order
	Dependencies []string `toml:"dependencies,omitempty"`
}

func DecodeGraph(decoder GraphDecoder) (*Graph, error) {
	eGraph := &EncodedGraph{}

	// Populate graph from our decoder (TOML, JSON, whatever)
	if err := decoder.Decode(&eGraph); err != nil {
		return nil, err
	}

	// ID to lockGraphItem mapping
	itemMap := map[string]*encodedItem{}
	for _, p := range eGraph.Packages {
		itemMap[p.id()] = p
	}

	// Recuse through all items and populate graph
	// Each graph item is only added once from itemMap
	var addFunc func(graph *Graph, item *encodedItem, root bool) error
	addFunc = func(graph *Graph, item *encodedItem, root bool) error {
		for _, depID := range item.Dependencies {
			childItem := itemMap[depID]
			parents := []string{""} // root
			if !root {
				parents = []string{item.Name}
			}

			// Will add each item once, and append parent list later
			spec, err := decoder.NewSpec(&childItem.GraphItem)
			if err != nil {
				return err
			}

			// Add specification to the graph
			graph.addChildVertex(childItem.Name, spec, parents, nil)

			// Add dependent release specification, if present
			if r, ok := spec.(released); ok {
				release, parent := r.ReleaseSpec(), []string{spec.Name()}
				graph.addChildVertex(release.Name(), release, parent, nil)
			}

			// Add recurse to traverse graph via dependencies
			if err := addFunc(graph, childItem, false); err != nil {
				return err
			}
		}
		return nil
	}

	graph := NewGraph()
	err := addFunc(graph, eGraph.Project, true)
	return graph, err
}

func (g *Graph) Encode(encoder GraphEncoder) error {
	eGraph := EncodedGraph{Project: &encodedItem{}}
	eGraph.Packages = []*encodedItem{}

	// Iterate through the nodes to generate packages
	itemMap := map[*Vertex]*encodedItem{}
	for _, n := range g.Nodes() {
		v, s := n.(*Vertex), n.(*Vertex).Payload
		gItem := GraphItem{Name: s.Name(), Version: s.Version()}
		item := &encodedItem{GraphItem: gItem}
		item.Dependencies = []string{}
		itemMap[v] = item

		if v.Root {
			eGraph.Project.Dependencies = append(eGraph.Project.Dependencies, item.id())
		}

		if !strings.HasPrefix(s.Name(), "repo://") {
			eGraph.Packages = append(eGraph.Packages, item)
		}
	}

	// Generate dependeny lists from edges
	for _, e := range g.Edges() {
		t := e.To().(*Vertex)
		f := e.From().(*Vertex)
		tI, fI := itemMap[t], itemMap[f]

		if !strings.HasPrefix(t.Payload.Name(), "repo://") {
			fI.Dependencies = append(fI.Dependencies, tI.id())
			continue
		}

		// If it's a release, use revision from Version or Release
		fI.Release = strings.TrimPrefix(t.Payload.Name(), "repo://")
		if r, ok := t.Payload.(revisioned); ok {
			fI.Release += "#" + r.Revision()
		} else if r, ok := f.Payload.(revisioned); ok {
			fI.Release += "#" + r.Revision()
		}
	}

	// Sort/normalize package everything
	sort.Sort(encodedItemSort{eGraph.Packages})

	// Sort/normalize all item dependency IDs
	sort.Sort(sort.StringSlice(eGraph.Project.Dependencies))
	for _, item := range eGraph.Packages {
		sort.Sort(sort.StringSlice(item.Dependencies))
	}

	return encoder.Encode(&eGraph)
}

// Sorting lockfile objects
type encodedItemSort struct {
	s []*encodedItem
}

func (s encodedItemSort) Len() int {
	return len(s.s)
}

func (s encodedItemSort) Swap(i, j int) {
	s.s[i], s.s[j] = s.s[j], s.s[i]
}

func (s encodedItemSort) Less(i, j int) bool {
	return s.s[i].Name < s.s[j].Name
}

///////////////////////////////////////////////////
// JSON encoding for reading Molinillo testdata
///////////////////////////////////////////////////

type jsonGraphItem struct {
	Name         string
	Version      string
	Dependencies []*jsonGraphItem
}

// Marshalling/unmarshalling of JSON testdata
func (g *Graph) UnmarshalJSON(data []byte) error {
	var graphItems []*jsonGraphItem
	if err := json.Unmarshal(data, &graphItems); err != nil {
		return err
	}

	// Recursively add all graph items to Graph
	var addFunc func(graph *Graph, ary []*jsonGraphItem, root bool)
	addFunc = func(graph *Graph, ary []*jsonGraphItem, root bool) {
		for _, r := range ary {
			graph.addVertex(r.Name, rubygem.NewSpec(r.Name, r.Version), root)
			addFunc(graph, r.Dependencies, false)
		}
	}

	newGraph := NewGraph()
	addFunc(newGraph, graphItems, true)
	*g = *newGraph
	return nil
}

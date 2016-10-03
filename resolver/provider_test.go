package resolver

import (
	"encoding/json"
	"github.com/melody-sh/melody/resolver/rubygem"
	"github.com/melody-sh/melody/resolver/types"
	c "gopkg.in/check.v1"
	"io/ioutil"
	"sort"
)

// SpecificationProvider loaded from JSON
type testSpecProvider struct {
	BaseProvider
	Index map[string][]*rubygem.Specification
}

func (s *MySuite) jsonProvider(name string) *testSpecProvider {
	raw, err := ioutil.ReadFile("testdata/index/" + name + ".json")
	checkSuiteError(err, "Cannot open index json: ", name)
	provider := &testSpecProvider{}
	err = json.Unmarshal(raw, &provider.Index)
	checkSuiteError(err, "Cannot parse index json: ", name)
	return provider
}

func (p *testSpecProvider) SearchFor(dep types.Requirement) []types.Specification {
	specs := []types.Specification{}
	for _, s := range p.Index[dep.Name()] {
		if p.IsRequirementSatisfiedBy(dep, nil, s) {
			specs = append(specs, s)
		}
	}

	SortSpecs(specs, rubygem.VersionParser)
	return specs
}

func (p *testSpecProvider) DependenciesFor(spec types.Specification) types.Requirements {
	return spec.Requirements()
}

func (p *testSpecProvider) IsRequirementSatisfiedBy(d types.Requirement, _ *Graph, spec types.Specification) bool {
	ok, err := d.SatisfiedBy(spec)
	checkSuiteError(err, "Cannot check version ", spec.Version(), " vs. ", d)
	return ok
}

// Name-based sorter for SpecProvider.SortDependencies because
// JSON files store "dependencies" as maps, not arrays and we
// need to normalize to check original sort order
type depsNameSorter struct {
	types.Requirements
}

func (s depsNameSorter) Less(i, j int) bool {
	slice := s.Requirements
	return slice[i].Name() < slice[j].Name()
}

// Name-based sorter for SpecProvider.SortDependencies
func (s *MySuite) Test_SpecProvider_SortDependencies(t *c.C) {
	provider := s.jsonProvider("awesome")
	deps := provider.Index["rails"][0].Requirements()
	sort.Sort(&depsNameSorter{deps})
	t.Assert(len(deps), c.Equals, 5)

	nameSortedOrder := []string{
		"actionmailer", "actionpack", "activerecord",
		"activesupport", "railties",
	}

	// Check whether all the dependencies are here
	checkDepsOrder(t, deps, nameSortedOrder)

	// Sorting empty graph/conflict retains order
	activated, conflicts := NewGraph(), &Conflicts{}
	deps2 := provider.SortDependencies(deps, activated, *conflicts)
	checkDepsOrder(t, deps2, nameSortedOrder)

	// Prioritize dependencies with conflicts
	(*conflicts)["railties"] = &Conflict{}
	deps2 = provider.SortDependencies(deps, activated, *conflicts)
	checkDepsOrder(t, deps2, []string{
		"railties", "actionmailer", "actionpack",
		"activerecord", "activesupport",
	})

	// Prioritize activated dependencies w/ payloads
	activated, conflicts = NewGraph(), &Conflicts{}
	activated.addVertex("railties", rubygem.NewSpec("", ""), false)
	deps2 = provider.SortDependencies(deps, activated, *conflicts)
	checkDepsOrder(t, deps2, []string{
		"railties", "actionmailer", "actionpack",
		"activerecord", "activesupport",
	})

	// Check that original dependencies array is untouched
	checkDepsOrder(t, deps, nameSortedOrder)
}

// Check that the dependencies are ordered by name
func checkDepsOrder(t *c.C, deps types.Requirements, order []string) {
	for i, name := range order {
		t.Assert(deps[i].Name(), c.Equals, name)
	}
}

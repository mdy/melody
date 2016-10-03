package resolver

import (
	"encoding/json"
	"github.com/melody-sh/melody/resolver/rubygem"
	"github.com/melody-sh/melody/resolver/types"
	c "gopkg.in/check.v1"
	"io/ioutil"
	"sort"
)

type resolverTestCase struct {
	Name      string
	Index     string
	Base      *Graph
	Requested rubygem.Requirements
	Resolved  *Graph
	Conflicts []string
}

func (s *MySuite) resolverCase(name string) *resolverTestCase {
	raw, err := ioutil.ReadFile("testdata/case/" + name + ".json")
	checkSuiteError(err, "Cannot open case json: "+name)

	testCase := &resolverTestCase{}
	err = json.Unmarshal(raw, &testCase)
	checkSuiteError(err, "Cannot open case json: ", name)
	sort.Sort(sort.StringSlice(testCase.Conflicts))
	return testCase
}

// Basic no-requirements test
func (s *MySuite) TestResolverNoRequirements(t *c.C) {
	provider := s.jsonProvider("awesome")
	resolver := NewResolver(provider, NewStdoutUI())
	out, _ := resolver.Resolve(types.Requirements{}, nil)
	expected := NewGraph().String()
	t.Assert(out.String(), c.Equals, expected)
}

// Read individual test cases and resolve them
var caseTests = []string{
	"circular",
	"complex_conflict",
	"conflict",
	"conflict_on_child",
	"deep_complex_conflict",
	"pruned_unresolved_orphan",
	"root_conflict_on_child",
	"simple",
	"simple_with_base",
	"simple_with_dependencies",
	"simple_with_shared_dependencies",
	"three_way_conflict",
	"unresolvable_child",
}

func (s *MySuite) TestResolverForCase(t *c.C) {
	for _, caseID := range caseTests {
		caseObj := s.resolverCase(caseID)
		indexID := caseObj.Index
		if indexID == "" {
			indexID = "awesome"
		}
		provider := s.jsonProvider(indexID)

		t.Log("Resolving case: ", caseObj.Name)
		resolver := NewResolver(provider, NewStdoutUI())
		out, outErr := resolver.Resolve(caseObj.Requested, caseObj.Base)

		expected, actual := caseObj.Resolved, out
		t.Log("Expected graph: ", expected.String())
		t.Log("Resolved graph: ", actual.String())

		t.Log("Expected conflicts: ", caseObj.Conflicts)
		t.Log("Resolved conflicts: ", outErr)
		t.Assert(actual.String(), c.Equals, expected.String())

		// Compare conflicts
		if len(caseObj.Conflicts) == 0 {
			t.Assert(outErr, c.IsNil)
		} else {
			t.Assert(outErr, c.NotNil) // Sanity check
			if cErr, ok := outErr.(*CircularDependencyError); ok {
				cNames := []string{cErr.Src.Name, cErr.Dst.Name}
				sort.Sort(sort.StringSlice(cNames))
				t.Assert(cNames, c.DeepEquals, caseObj.Conflicts)

			} else if vErr, ok := outErr.(*VersionConflictError); ok {
				conflictNames := []string{}
				for n := range Conflicts(*vErr) {
					conflictNames = append(conflictNames, n)
				}

				sort.Sort(sort.StringSlice(conflictNames))
				t.Assert(conflictNames, c.DeepEquals, caseObj.Conflicts)
			}
		}
	}

	//t.Assert(1234, c.IsNil) // FORCE DUMP LOGS
}

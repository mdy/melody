package resolver

import (
	"github.com/melody-sh/melody/resolver/flex"
	"github.com/melody-sh/melody/resolver/types"
	c "gopkg.in/check.v1"
)

func (s *MySuite) Test_PossibilityState(t *c.C) {
	ds := emptyResolutionState()
	ds.Name = "Test-Value"
	ds.Possibilities = []types.Specification{
		flex.NewSpec("possibility1", "1.2.3"),
		flex.NewSpec("possibility2", "4.5.6"),
	}

	// Just the last possibility in one array
	lastP := ds.Possibilities[len(ds.Possibilities)-1:]

	ds.Type = possibilityState
	t.Assert(ds.popPossibilityState(), c.IsNil)

	ds.Type = resolutionState
	t.Assert(ds.popPossibilityState(), c.IsNil)

	ds.Type = dependencyState
	ps := ds.popPossibilityState()
	t.Assert(ps.Name, c.Equals, ds.Name)
	t.Assert(ps.Type, c.Equals, possibilityState)
	t.Assert(ps.Requirement, c.Equals, ds.Requirement)

	// Should pop possibility from parent task
	t.Assert(len(ds.Possibilities), c.Equals, 1)

	// These are slices... they're ok being the same
	t.Assert(ps.Requirements, c.DeepEquals, ds.Requirements)

	// Should be copied, but not the same object
	t.Assert(&ps.Conflicts, c.Not(c.Equals), &ds.Conflicts)
	t.Assert(ps.Conflicts, c.DeepEquals, ds.Conflicts)
	t.Assert(ps.Activated, c.Not(c.Equals), ds.Activated)
	t.Assert(ps.Activated, c.DeepEquals, ds.Activated)

	// Pop will just populate one last possibility
	t.Assert(ps.Possibilities, c.DeepEquals, lastP)
	t.Assert(ps.Depth, c.Equals, ds.Depth+1)
}

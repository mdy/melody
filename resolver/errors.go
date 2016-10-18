package resolver

import (
	"fmt"
	"github.com/melody-sh/melody/resolver/types"
)

// Resolver error to indicate a circular dependency
type CircularDependencyError struct {
	Src *Vertex
	Dst *Vertex
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("CircularDependencyError(%s, %s)", e.Src.Name, e.Dst.Name)
}

// Resolver error to indicate version conflict
type VersionConflictError Conflicts

func (e *VersionConflictError) Error() string {
	s := "VersionConflictError: \n"
	for name, c := range Conflicts(*e) {
		s += "  Could not find compatible versions for"
		s += " \"" + name + "\":\n    " + c.Requirement.String() + "\n"
		for _, branch := range c.RequirementTrees {
			if len(branch) == 0 {
				continue
			}
			prefix := "  "
			s += "\n"
			for _, r := range branch {
				s += prefix + r.String()
				if spec, ok := c.ActivatedByName[r.Name()]; ok && r.Name() != name {
					s += " was resolved to " + spec.Version()
				}
				prefix += "  "
				s += "\n"
			}
		}
	}
	return s //fmt.Sprintf("VersionConflictError: %s", Conflicts(*e))
}

// Conflicts by dependency name
type Conflicts map[string]*Conflict

// Copy map keeping Conflict objects
func (c Conflicts) Dup() Conflicts {
	newMap := make(Conflicts, len(c))
	for k, v := range c {
		newMap[k] = v
	}
	return newMap
}

// Conflicting requirements that cannot be both met
type Conflict struct {
	Requirement       types.Requirement
	Existing          types.Specification
	Possibility       types.Specification
	LockedRequirement types.Requirement
	RequirementTrees  [][]types.Requirement
	ActivatedByName   map[string]types.Specification
	Requirements      map[string][]types.Requirement
}

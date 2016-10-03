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
	return fmt.Sprintf("VersionConflictError: %s", Conflicts(*e))
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

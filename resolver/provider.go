package resolver

import (
	"github.com/mdy/melody/resolver/types"
	"sort"
)

// Provider interface for package index
type SpecificationProvider interface {
	AllowMissing(types.Requirement) bool
	SearchFor(types.Requirement) []types.Specification
	DependenciesFor(types.Specification) types.Requirements
	SortDependencies(types.Requirements, *Graph, Conflicts) types.Requirements
	NameForExplicitDependencySource() string
	NameForLockingDependencySource() string
	IsRequirementSatisfiedBy(types.Requirement, *Graph, types.Specification) bool
}

// Basic implementation for some methods
type BaseProvider struct {
}

func (p *BaseProvider) SearchFor(dep types.Requirement) []types.Specification {
	return []types.Specification{}
}

func (p *BaseProvider) DependenciesFor(spec types.Specification) types.Requirements {
	return types.Requirements{}
}

func (p *BaseProvider) NameForExplicitDependencySource() string {
	return "user-specified dependency"
}

func (p *BaseProvider) NameForLockingDependencySource() string {
	return "Lockfile"
}

func (p *BaseProvider) AllowMissing(dep types.Requirement) bool {
	return false
}

// Sort dependencies so that the ones that are easiest to resolve are first.
func (p *BaseProvider) SortDependencies(deps types.Requirements, activated *Graph, conflicts Conflicts) types.Requirements {
	output := append(types.Requirements{}, deps...)
	sort.Stable(&depsPrioritySorter{output, activated, conflicts})
	return output
}

// Dependency sorter for SpecProvider.SortDependencies
type depsPrioritySorter struct {
	types.Requirements
	activated *Graph
	conflicts Conflicts
}

func (s depsPrioritySorter) Less(i, j int) bool {
	nameI, nameJ := s.Requirements[i].Name(), s.Requirements[j].Name()
	hasLoadI := s.activated.PayloadFor(nameI) != nil
	hasLoadJ := s.activated.PayloadFor(nameJ) != nil
	_, hasConfI := s.conflicts[nameI]
	_, hasConfJ := s.conflicts[nameJ]
	return (hasLoadI && !hasLoadJ) || (hasConfI && !hasConfJ)
}

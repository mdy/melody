package resolver

import (
	"github.com/mdy/melody/resolver/types"
)

type StateType int

const (
	resolutionState  StateType = iota
	possibilityState StateType = iota
	dependencyState  StateType = iota
)

type State struct {
	Type          StateType
	Depth         int
	Name          string
	Activated     *Graph
	Requirement   types.Requirement
	Requirements  types.Requirements
	Possibilities []types.Specification
	Conflicts     Conflicts
}

func emptyResolutionState() *State {
	return &State{
		Type:          resolutionState,
		Activated:     NewGraph(),
		Conflicts:     Conflicts{},
		Requirements:  []types.Requirement{},
		Possibilities: []types.Specification{},
	}
}

func (s *State) popPossibilityState() *State {
	if s.Type != dependencyState {
		return nil
	}

	// Pop last possibility or use nil
	possibilites := []types.Specification{nil}
	if l := len(s.Possibilities); l > 0 {
		possibilites[0] = s.Possibilities[l-1]
		s.Possibilities = s.Possibilities[:l-1]
	}

	return &State{
		Type:          possibilityState,
		Name:          s.Name,
		Depth:         s.Depth + 1,
		Requirement:   s.Requirement,
		Requirements:  s.Requirements.Dup(),
		Activated:     s.Activated.Dup(),
		Conflicts:     s.Conflicts.Dup(),
		Possibilities: possibilites,
	}
}

func (s *State) hasRequirement(req types.Requirement) bool {
	allReqs := append([]types.Requirement{s.Requirement}, s.Requirements...)
	for _, r := range allReqs {
		if r == req { // FIXME
			return true
		}
	}
	return false
}

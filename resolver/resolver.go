package resolver

import (
	"github.com/melody-sh/melody/resolver/types"
	"time"
)

type Resolver struct {
	provider SpecificationProvider
	ui       UI
}

func NewResolver(provider SpecificationProvider, ui UI) *Resolver {
	return &Resolver{provider: provider, ui: ui}
}

func (r *Resolver) Resolve(requested []types.Requirement, base *Graph) (*Graph, error) {
	if base == nil {
		base = NewGraph()
	}
	return (&Resolution{
		SpecProvider:      r.provider,
		OriginalRequested: requested,
		Base:              base,
		UI:                r.ui,
	}).Resolve()
}

type Resolution struct {
	// Provider to retrieve dependencies, requirements, etc
	SpecProvider SpecificationProvider

	// UI to communicate back to the user
	UI UI

	// Base dependency graph to lock dependencies
	Base *Graph

	// Explicitly requested updates
	OriginalRequested []types.Requirement

	// Internal processing
	iterationCounter int
	progressAt       time.Time
	startedAt        time.Time
	endedAt          time.Time
	states           []*State
}

func (r *Resolution) Resolve() (*Graph, error) {
	r.startResolution()
	defer r.endResolution()

	for len(r.states) > 0 {
		r.debug("ITERATION: %d STATES: %d", r.iterationCounter, len(r.states))
		r.indicateProgress()

		// FIXME: REMOVE ITERATION LIMIT
		if r.iterationCounter > 20000 {
			panic("REACHED ITERATION COUNT")
		}

		state := r.state() // r.states[last]
		if len(state.Requirements) == 0 && state.Requirement == nil {
			break
		}

		if state := state.popPossibilityState(); state != nil {
			pCount := len(state.Possibilities)
			r.debug("Creating possibility state for %s (%d remaining)", state.Requirement, pCount)
			r.states = append(r.states, state)
		}

		if err := r.processTopmostState(); err != nil {
			return NewGraph(), err
		}
	}

	if len(r.state().Conflicts) == 0 {
		return r.state().Activated, nil
	}

	//err := fmt.Errorf("Unexpected conflicts: %s", r.state().Conflicts)
	return r.state().Activated, nil
}

func (r *Resolution) startResolution() {
	r.startedAt = time.Now()
	r.progressAt = r.startedAt
	r.handleMissingOrPushDependencyState(r.initialState())
	r.debug("Starting resolution (%s)", r.startedAt)
	r.UI.BeforeResolution()
}

func (r *Resolution) endResolution() {
	r.endedAt = time.Now()
	r.UI.AfterResolution()
	r.debug("Finished resolution (%d steps in %s)",
		r.iterationCounter, r.endedAt.Sub(r.startedAt))
}

func (r *Resolution) processTopmostState() error {
	if r.possibility() != nil {
		return r.attemptToActivate()
	}

	if r.state().Type == possibilityState {
		r.createConflict()
	}

	for !(r.possibility() != nil && r.state().Type == dependencyState) {
		if err := r.unwindForConflict(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolution) possibility() types.Specification {
	if s := r.state(); s != nil && len(s.Possibilities) > 0 {
		return s.Possibilities[len(s.Possibilities)-1]
	}
	return nil
}

func (r *Resolution) state() *State {
	if l := len(r.states); l > 0 {
		return r.states[l-1]
	}
	return emptyResolutionState()
}

func (r *Resolution) initialState() *State {
	graph := NewGraph()
	for _, dep := range r.OriginalRequested {
		vertex := graph.addVertex(dep.Name(), nil, true)
		vertex.ExplicitRequirements = append(vertex.ExplicitRequirements, dep)
	}

	state := &State{
		Activated: graph,
		Conflicts: Conflicts{},
		Type:      dependencyState,
	}

	// Sort with an empty Graph and Conflicts
	requirements := r.SpecProvider.SortDependencies(r.OriginalRequested, graph, state.Conflicts)

	if len(requirements) == 0 {
		state.Possibilities = []types.Specification{}
	} else {
		initialRequirement, requirements := requirements[0], requirements[1:]
		state.Possibilities = r.SpecProvider.SearchFor(initialRequirement)
		state.Name = initialRequirement.Name()
		state.Requirement = initialRequirement
		state.Requirements = requirements
	}

	return state
}

func (r *Resolution) unwindForConflict() error {
	r.debug("Unwinding for conflict: %s", r.state().Requirement)
	conflicts := r.state().Conflicts

	i := r.stateIndexForUnwind()
	r.debug("stateIndexForUnwind: %d of %d", i, len(r.states))
	r.states = r.states[:i+1]

	if len(r.states) == 0 {
		err := VersionConflictError(conflicts)
		return &err
	}

	r.state().Conflicts = conflicts
	return nil
}

func (r *Resolution) stateIndexForUnwind() int {
	existing := r.requirementForExistingState(r.state())
	current := r.state().Requirement

	if i := r.stateIndexToMeetDependency(current); i >= 0 {
		return i
	}
	return r.stateIndexToMeetDependency(existing)
}

func (r *Resolution) stateIndexToMeetDependency(req types.Requirement) int {
	for ; req != nil; req = r.parentOf(req) {
		state := r.findStateFor(req)
		if state != nil && len(state.Possibilities) > 0 {
			for i, s := range r.states {
				if s == state {
					return i
				}
			}
		}
	}
	return -1
}

func (r *Resolution) parentOf(requirement types.Requirement) types.Requirement {
	if requirement == nil {
		return nil
	}
	for seen, i := false, len(r.states)-1; i >= 0; i-- {
		state := r.states[i]
		seen = seen || state.hasRequirement(requirement)
		if seen && !state.hasRequirement(requirement) {
			return state.Requirement
		}
	}
	return nil
}

func (r *Resolution) requirementForExistingState(s *State) types.Requirement {
	if s == nil || r.state().Activated.PayloadFor(s.Name) == nil {
		return nil
	}

	for i := len(r.states) - 1; i >= 0; i-- {
		if r.states[i].Activated.PayloadFor(s.Name) == nil {
			return r.states[i].Requirement
		}
	}
	return nil
}

func (r *Resolution) findStateFor(requirement types.Requirement) *State {
	if requirement != nil {
		for i := len(r.states) - 1; i >= 0; i-- {
			if s := r.states[i]; s.Type == dependencyState && s.Requirement == requirement {
				return s
			}
		}
	}
	return nil
}

func (r *Resolution) createConflict() {
	state, provider := r.state(), r.SpecProvider
	vertex := state.Activated.vertexNamed(state.Name)
	requirements := map[string][]types.Requirement{}

	if er := vertex.ExplicitRequirements; len(er) > 0 {
		requirements[provider.NameForExplicitDependencySource()] = er
	}

	lockedReq := r.lockedRequirementNamed(state.Name)
	if lockedReq != nil {
		key := provider.NameForLockingDependencySource()
		requirements[key] = []types.Requirement{lockedReq}
	}

	state.Conflicts[state.Name] = &Conflict{
		LockedRequirement: lockedReq,
		Requirements:      requirements,
		Requirement:       state.Requirement,
		Existing:          r.possibility(),
		RequirementTrees:  r.requirementTrees(),
		ActivatedByName:   state.Activated.ActivatedByName(),
	}
}

func (r *Resolution) requirementTrees() [][]types.Requirement {
	state := r.state()
	requirements := state.Activated.requirementsFor(state.Name)
	out := make([][]types.Requirement, len(requirements))
	for _, req := range requirements {
		tree := []types.Requirement{} // requirementTreeFor
		for ; req != nil; req = r.parentOf(req) {
			tree = append([]types.Requirement{req}, tree...)
		}
		out = append(out, tree)
	}
	return out
}

func (r *Resolution) indicateProgress() {
	now := time.Now()
	r.iterationCounter++
	if now.Sub(r.progressAt) >= r.UI.ProgressRate() {
		r.UI.IndicateProgress()
		r.progressAt = now
	}
}

func (r *Resolution) attemptToActivate() error {
	r.debug("Attempting to activate %s", r.possibility())
	state := r.state()
	if v := state.Activated.vertexNamed(state.Name); v.Payload != nil {
		r.debug("Found existing spec %s", v.Payload)
		return r.attemptToActivateExistingSpec(v)
	}
	return r.attemptToActivateNewSpec()
}

func (r *Resolution) attemptToActivateExistingSpec(node *Vertex) error {
	s, existingSpec := r.state(), node.Payload
	if r.isRequirementSatisfiedBy(s.Requirement, s.Activated, existingSpec) {
		r.debug("Requirement %s satisfied by %s", s.Requirement, existingSpec)
		r.pushStateForRequirements(s.Requirements.Dup(), false, nil)
	} else if ok, err := r.attemptToSwapPossibility(); err != nil {
		r.debug("Attempting to swap possibility")
		return err
	} else if !ok {
		r.debug("Unsatisfied by existing spec (%s)", existingSpec)
		r.createConflict()
		return r.unwindForConflict()
	}

	return nil
}

func (r *Resolution) attemptToSwapPossibility() (bool, error) {
	s, p := r.state(), r.possibility()
	swapped := s.Activated.Dup()
	swapped.vertexNamed(s.Name).Payload = p

	for _, req := range swapped.requirementsFor(s.Name) {
		if !r.isRequirementSatisfiedBy(req, swapped, p) {
			return false, nil
		}
	}

	if !r.isNewSpecSatisfied() {
		return false, nil
	}

	actualVertex := s.Activated.vertexNamed(s.Name)
	actualVertex.Payload = p
	r.fixSwappedChildren(actualVertex)
	return true, r.activateSpec()
}

func (r *Resolution) fixSwappedChildren(vertex *Vertex) {
	deps := r.SpecProvider.DependenciesFor(vertex.Payload)
	state, depSet := r.state(), map[string]bool{}

	// Vertex dependencies as { name => bool } map
	for _, dep := range deps {
		depSet[dep.Name()] = true
	}

	// Clean graph of orphaned kids for new vertex
	for _, sVec := range state.Activated.From(vertex) {
		succ, upDeps := sVec.(*Vertex), state.Activated.To(sVec)
		if !depSet[succ.Name] && !succ.Root && len(upDeps) == 1 && upDeps[0].ID() == vertex.ID() {
			r.debug("Removing orphaned spec %s after swapping %s", succ.Name, state.Name)
			state.Activated.detachVertexNamed(succ.Name)
			newReqs := []types.Requirement{}
			for _, r := range state.Requirements {
				if r.Name() != succ.Name {
					newReqs = append(newReqs, r)
				}
			}
			state.Requirements = newReqs
		}
	}
}

func (r *Resolution) attemptToActivateNewSpec() error {
	if r.isNewSpecSatisfied() {
		return r.activateSpec()
	}
	r.createConflict()
	return r.unwindForConflict()
}

func (r *Resolution) isNewSpecSatisfied() bool {
	state, p := r.state(), r.possibility()
	lockedReq := r.lockedRequirementNamed(state.Name)
	r.debug("CHECKING POSSIBILITY AGAINST LOCKED %s and %s", lockedReq, state.Requirement)
	reqOK := r.isRequirementSatisfiedBy(state.Requirement, state.Activated, p)
	lockedOK := lockedReq == nil || r.isRequirementSatisfiedBy(lockedReq, state.Activated, p)
	if !reqOK {
		r.debug("Unsatisfied by requested spec")
	}
	if !lockedOK {
		r.debug("Unsatisfied by locked spec")
	}
	return reqOK && lockedOK
}

func (r *Resolution) lockedRequirementNamed(name string) types.Requirement {
	if spec := r.Base.PayloadFor(name); spec != nil {
		return &lockedRequirement{spec}
	}
	return nil
}

type lockedRequirement struct {
	types.Specification
}

func (l *lockedRequirement) SatisfiedBy(spec types.Specification) (bool, error) {
	return SpecEqual(l, spec), nil
}

func (r *Resolution) activateSpec() error {
	state, possibility := r.state(), r.possibility()
	delete(state.Conflicts, state.Name)
	r.debug("Activated %s at %s", state.Name, possibility)
	vertex := state.Activated.vertexNamed(state.Name)
	vertex.Payload = possibility
	r.debug("ACTIVATED %s %s", possibility.Name, possibility.Version)
	return r.requireNestedDependenciesFor(possibility)
}

func (r *Resolution) requireNestedDependenciesFor(spec types.Specification) error {
	s, nestedDeps := r.state(), r.SpecProvider.DependenciesFor(spec)
	r.debug("Requiring nested dependencies (%s)", nestedDeps)
	specNames := []string{spec.Name()}

	// Populate dependencies in graph
	for _, d := range nestedDeps {
		_, err := s.Activated.addChildVertex(d.Name(), nil, specNames, d)
		if err != nil {
			return err
		}
	}

	r.pushStateForRequirements(
		append(s.Requirements, nestedDeps...),
		len(nestedDeps) > 0,
		nil,
	)

	return nil
}

func (r *Resolution) pushStateForRequirements(reqs []types.Requirement, requiresSort bool, newActivated *Graph) {
	state, provider := r.state(), r.SpecProvider
	if newActivated == nil {
		newActivated = r.state().Activated.Dup()
	}

	if requiresSort {
		reqs = provider.SortDependencies(reqs, newActivated, state.Conflicts)
	}

	// Shift new requirement/requirementsState
	newState := &State{
		Type:      dependencyState,
		Depth:     state.Depth,
		Activated: newActivated,
		Conflicts: state.Conflicts.Dup(),
	}

	if len(reqs) > 0 {
		r, reqs := reqs[0], reqs[1:]
		newState.Name = r.Name()
		newState.Requirement = r
		newState.Requirements = reqs
		newState.Possibilities = provider.SearchFor(r)
	} else {
		newState.Requirements = []types.Requirement{}
		newState.Possibilities = []types.Specification{}
	}

	r.handleMissingOrPushDependencyState(newState)
}

func (r *Resolution) handleMissingOrPushDependencyState(state *State) {
	if state.Requirement != nil && len(state.Possibilities) == 0 && r.allowMissing(state.Requirement) {
		r.debug("Pushing state for requirements", state.Name)
		state.Activated.detachVertexNamed(state.Name)
		r.pushStateForRequirements(state.Requirements.Dup(), false, state.Activated)
	} else {
		r.debug("Adding this state dependency state: %s", state.Name)
		r.states = append(r.states, state)
	}
}

// ==== Proxy methods to SpecificationProvider ====State
func (r *Resolution) isRequirementSatisfiedBy(req types.Requirement, graph *Graph, p types.Specification) bool {
	//  return req != nil && r.SpecProvider.IsRequirementSatisfiedBy(req, graph, p)
	return r.SpecProvider.IsRequirementSatisfiedBy(req, graph, p)
}

func (r *Resolution) allowMissing(req types.Requirement) bool {
	return r.SpecProvider.AllowMissing(req)
}

func (r *Resolution) debug(args ...interface{}) {
	if s := r.state(); s != nil {
		r.UI.Debug(s.Depth, args...)
	} else {
		r.UI.Debug(0, args...)
	}
}

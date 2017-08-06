package melody

import (
	"github.com/mdy/melody/provider"
	"github.com/mdy/melody/resolver"
	"github.com/mdy/melody/resolver/flex"
	"github.com/mdy/melody/resolver/types"

	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ============== Allows to do revision matching ================
type Revisioned interface {
	Revision() string
}

// ============== Builder for Graph decode/encode ===============
type Builder struct{}

func (b *Builder) NewSpec(i *resolver.GraphItem) (types.Specification, error) {
	spec := &melodySpec{Specification: *(flex.NewSpec(i.Name, i.Version))}

	if j := strings.Index(i.Release, "#"); j >= 0 {
		name, rev := i.Release[0:j], i.Release[j+1:]
		releaseSpec := flex.NewSpec(name, i.Version)
		url := fmt.Sprintf(melodyReleaseURL, name, rev)
		spec.Release = &melodyRelease{*releaseSpec, rev, url}
	}

	return spec, nil
}

// ============== JSON RESPONSE AND RESOLVER ================

type melodySpec struct {
	flex.Specification
	Release        *melodyRelease
	DependencyList melodyRequirements
}

func (ms *melodySpec) Requirements() types.Requirements {
	return append(types.Requirements(ms.DependencyList), ms.Release)
}

// Revisioned interface
func (ms *melodySpec) Revision() string {
	return ms.Release.Revision
}

// Implement resolver.Released interface
func (ms *melodySpec) ReleaseSpec() provider.ReleaseSpec {
	return ms.Release
}

// Melody release acts as its own Requirement & Specification.  This prevents
// makes sure that packages from the same repo resolve to the same version
type melodyRelease struct {
	flex.Specification
	Revision string
	URL      string
}

// Unique name from a corresponding package
func (r *melodyRelease) Name() string {
	return "repo://" + r.NameStr
}

// Name displayed and queried by user via UI, as well
// as the name written to project (Melody.*) files
func (r *melodyRelease) ExternalName() string {
	return r.NameStr
}

// Subdirectory to for installation into project
func (r *melodyRelease) InstallPath() string {
	return filepath.Clean(r.NameStr)
}

// Releases are always leafs -- just to be sure
func (r *melodyRelease) Requirements() types.Requirements {
	return types.Requirements{}
}

// fmt.Revisioned ...
func (r *melodyRelease) String() string {
	return fmt.Sprintf("Release(%s %s)", r.NameStr, r.VersStr)
}

// Only matches itself (same name and version)
func (r *melodyRelease) SatisfiedBy(spec types.Specification) (bool, error) {
	mSpec, ok := spec.(*melodyRelease)
	return ok && resolver.SpecEqual(mSpec, r), nil
}

// Used for initializing Graph when loading Melody.lock
func (p *Melody) NewRequirement(n, v string) types.Requirement {
	return &melodyRequirement{flex.NewDependency(n, v)}
}

// Melody requirement that allows "head" meaning "latest release or beta"
type melodyRequirement struct {
	*flex.Dependency
}

func (s *melodyRequirement) SatisfiedBy(spec types.Specification) (bool, error) {
	if s.NameStr != spec.Name() {
		return false, nil
	}

	if s.RangeStr == "head" || s.RangeStr == "**" {
		return true, nil
	}

	if strings.HasPrefix(s.RangeStr, "#") {
		if r, ok := spec.(Revisioned); ok {
			return s.RangeStr[1:] == r.Revision(), nil
		}
	}

	return s.Dependency.SatisfiedBy(spec)
}

// Melody requirement marshalling and initialization
type melodyRequirements types.Requirements

func (mr *melodyRequirements) UnmarshalJSON(data []byte) error {
	depList := []struct {
		Name  string `json:"name"`
		Range string `json:"versionRange"`
	}{}

	if err := json.Unmarshal(data, &depList); err != nil {
		return err
	}

	out := melodyRequirements{}
	for _, req := range depList {
		fDep := flex.NewDependency(req.Name, req.Range)
		out = append(out, &melodyRequirement{fDep})
	}

	*mr = out
	return nil
}

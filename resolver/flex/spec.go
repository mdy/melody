package flex

import (
	"fmt"
	"github.com/mdy/melody/resolver/types"
	"github.com/mdy/melody/version"
)

func NewSpec(n, v string) *Specification {
	return &Specification{NameStr: n, VersStr: v}
}

func NewDependency(n, r string) *Dependency {
	return &Dependency{n, r}
}

// Interface for a full name/version/dependency spec
type Specification struct {
	NameStr      string `json:"name"`
	VersStr      string `json:"version"`
	Dependencies types.Requirements
}

func (s *Specification) Name() string {
	return s.NameStr
}

func (s *Specification) Version() string {
	return s.VersStr
}

func (s *Specification) Requirements() types.Requirements {
	return types.Requirements(s.Dependencies)
}

func (s *Specification) String() string {
	return fmt.Sprintf("Spec(%s %s)", s.NameStr, s.VersStr)
}

// For resolver's lockedRequirementNamed to lock a requirement
func (s *Specification) SatisfiedBy(spec types.Specification) (bool, error) {
	return types.SpecEqual(s, spec), nil
}

// This runs against Molinillo testdata
type Dependency struct {
	NameStr  string
	RangeStr string
}

func (s *Dependency) Name() string {
	return s.NameStr
}

func (s *Dependency) String() string {
	return fmt.Sprintf("FlexDependency(%s %s)", s.NameStr, s.RangeStr)
}

func (s *Dependency) SatisfiedBy(spec types.Specification) (bool, error) {
	if s.NameStr != spec.Name() {
		return false, nil
	}

	flex, err := ParseVersion(spec.Version())
	if err != nil {
		return false, err
	}

	return version.SatisfiesRange(flex, s.RangeStr)
}

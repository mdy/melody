package rubygem

import (
	"encoding/json"
	"fmt"
	"github.com/melodysh/melody/resolver/flex"
	"github.com/melodysh/melody/resolver/types"
	"github.com/melodysh/melody/version"
)

func NewSpec(n, v string) *Specification {
	return &Specification{*flex.NewSpec(n, v), nil}
}

type Specification struct {
	flex.Specification
	Dependencies Requirements
}

func (gs *Specification) String() string {
	v, _ := VersionParser.Parse(gs.VersStr)
	return fmt.Sprintf("Spec(%s %s)", gs.NameStr, v.String())
}

func (gs *Specification) Requirements() types.Requirements {
	return types.Requirements(gs.Dependencies)
}

// To help serialization for Molinillo test data
type Requirements []types.Requirement

func (sd *Requirements) UnmarshalJSON(data []byte) error {
	depMap := map[string]string{}
	if err := json.Unmarshal(data, &depMap); err != nil {
		return err
	}

	out := []types.Requirement{}
	for name, req := range depMap {
		svDep := flex.NewDependency(name, req)
		out = append(out, &Dependency{*svDep})
	}

	*sd = Requirements(out)
	return nil
}

// This runs against Molinillo testdata
type Dependency struct {
	flex.Dependency
}

func (s *Dependency) String() string {
	return fmt.Sprintf("GemDependency(%s %s)", s.NameStr, s.RangeStr)
}

// FIXME: Describe how this is different from SemVer
func (s *Dependency) SatisfiedBy(spec types.Specification) (bool, error) {
	if s.NameStr != spec.Name() {
		return false, nil
	}

	flex, err := flex.ParseVersion(spec.Version())
	if err != nil {
		return false, err
	}

	gemVer := Version{flex}
	return version.SatisfiesRange(gemVer, s.RangeStr)
}

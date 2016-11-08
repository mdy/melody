package provider

import (
	"github.com/mdy/melody/provider/melody"
	"github.com/mdy/melody/resolver"
	"github.com/mdy/melody/resolver/types"
)

type Provider interface {
	NewRequirement(string, string) types.Requirement
	InstallToDir(string, []types.Specification) error
	resolver.SpecificationProvider
}

func NewMelody(base *resolver.Graph) *melody.Melody {
	return melody.New(base)
}

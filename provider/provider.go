package provider

import (
	"github.com/melody-sh/melody/provider/melody"
	"github.com/melody-sh/melody/resolver"
	"github.com/melody-sh/melody/resolver/types"
)

type Provider interface {
	NewRequirement(string, string) types.Requirement
	InstallToDir(string, []types.Specification) error
	resolver.SpecificationProvider
}

func NewMelody(base *resolver.Graph) *melody.Melody {
	return melody.New(base)
}

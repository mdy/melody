package provider

import (
	"github.com/mdy/melody/resolver"
	"github.com/mdy/melody/resolver/types"
)

// Interface to fetch remote/local specs
type Provider interface {
	NewRequirement(string, string) types.Requirement
	InstallToDir(string, []types.Specification) error
	resolver.SpecificationProvider
}

// Specification for package version
type VersionSpec interface {
	ReleaseSpec() ReleaseSpec
	types.Specification
}

// Specification for repository release
type ReleaseSpec interface {
	ExternalName() string
	InstallPath() string
	types.Specification
}

package project

import (
	log "github.com/Sirupsen/logrus"
	"github.com/melody-sh/melody/provider"
	"github.com/melody-sh/melody/resolver"
	"github.com/melody-sh/melody/resolver/types"
	"os"
	"path/filepath"
)

// Resolve project specifications while locking a dependency graph
// This can be used for 3 different ways to use it:
// 1. Lock everything in lockfile (install)
// 2. Lock nothing, update everything (update)
// 3. Lock some packages (update <pkg>)
func (p *Project) Resolve(src provider.Provider, base *resolver.Graph) (*resolver.Graph, error) {
	// Convert Project.Config to Requested
	rDeps := []types.Requirement{}
	for name, r := range p.Config.Dependencies {
		rDeps = append(rDeps, src.NewRequirement(name, r))
	}

	// Resolve dependencies
	log.Info("Dependencies", rDeps)
	res := resolver.NewResolver(src, resolver.NewStdoutUI())
	return res.Resolve(rDeps, base)
}

// Resolve project specifications and install them in ./vendor
func (p *Project) UpdateWithBase(src provider.Provider, base *resolver.Graph) error {
	// Resolve dependencies
	out, outErr := p.Resolve(src, base)
	if outErr != nil {
		return outErr
	}

	// Save state
	p.Locked = out
	log.Info("Saving lockfile: ", p.Save())

	// Start package installation
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Install packages to destination
	target := filepath.Join(dir, "vendor")
	err = src.InstallToDir(target, p.Locked.Specifications())
	return err
}

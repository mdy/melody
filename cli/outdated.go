package cli

import (
	"github.com/mdy/melody/project"
	"github.com/mdy/melody/provider"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"fmt"
	"os"
)

func outdated(c *cli.Context) error {
	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	log.Info("Project", project, " -- ", err)
	if err != nil {
		return err
	}

	// Load or resolve current specs
	source := project.Provider()
	if project.Locked == nil {
		project.Locked, err = project.Resolve(source, nil)
		if err != nil {
			return err
		}
	}

	fmt.Printf("♫ Resolving ...")

	// Check for new versions of all current specification
	outdated := map[string]outdatedRelease{}
	for _, oldSpec := range project.Locked.Specifications() {
		if _, ok := oldSpec.(provider.ReleaseSpec); ok {
			continue
		}

		fmt.Printf(".")
		req := source.NewRequirement(oldSpec.Name(), "> "+oldSpec.Version())
		specs := source.SearchFor(req)
		if len(specs) == 0 {
			continue
		}

		if ver, isVer := specs[len(specs)-1].(provider.VersionSpec); isVer {
			release := ver.ReleaseSpec()
			outdated[release.Name()] = outdatedRelease{
				Name:       release.ExternalName(),
				OldVersion: oldSpec.Version(),
				NewVersion: release.Version(),
			}
		}
	}
	fmt.Printf(" done.\n")

	if len(outdated) == 0 {
		fmt.Println("♫ All packages are up to date!")
		return nil
	}

	fmt.Println("♫ Outdated repositories in the project:")
	for _, or := range outdated {
		fmt.Printf("  * %s (newest %s, installed %s)\n",
			or.Name, or.NewVersion, or.OldVersion,
		)
	}

	return nil
}

type outdatedRelease struct {
	Name       string
	NewVersion string
	OldVersion string
}

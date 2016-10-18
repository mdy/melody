package cli

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/melodysh/melody/project"
	"github.com/melodysh/melody/provider"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func outdated(c *cli.Context) error {
	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	log.Info("Project", project, " -- ", err)
	if err != nil {
		return err
	}

	// Load or resolve current specs
	source := provider.NewMelody(project.Locked)
	if project.Locked == nil {
		project.Locked, err = project.Resolve(source, nil)
		if err != nil {
			return err
		}
	}

	// Check for new versions of all current specification
	hasOutdated := false
	for _, oldSpec := range project.Locked.Specifications() {
		if strings.HasPrefix(oldSpec.Name(), "repo://") {
			continue
		}

		req := source.NewRequirement(oldSpec.Name(), "> "+oldSpec.Version())
		specs := source.SearchFor(req)
		if len(specs) == 0 {
			continue
		}

		if !hasOutdated {
			fmt.Println("♫ Outdated packages in the project:")
			hasOutdated = true
		}

		newSpec := specs[len(specs)-1]
		fmt.Printf("  * %s (newest %s, installed %s)\n",
			oldSpec.Name(),
			newSpec.Version(),
			oldSpec.Version(),
		)
	}

	if !hasOutdated {
		fmt.Println("♫ All packages are up to date!")
	}

	return nil
}

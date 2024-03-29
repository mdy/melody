package cli

import (
	"github.com/mdy/melody/project"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"fmt"
	"os"
)

func install(c *cli.Context) error {
	if len(c.Args()) != 0 {
		return fmt.Errorf("`install` command takes no arguments. See '%s install --help'.", c.App.Name)
	}

	wDir, _ := os.Getwd()
	return runInstall(wDir, nil)
}

// This helper will load project and lockfile, run any mutations that the user
// specified via command line and run install (including saving the lockfile)
func runInstall(dir string, mutate func(*project.Project) error) error {
	project, err := project.Load(dir)
	log.Info("Project", project, " -- ", err)
	if err != nil {
		return err
	}

	// Perform mutation (add, remove, etc)
	if mutate != nil {
		if err := mutate(project); err != nil {
			return err
		}
	}

	// Convert Project.Config to Requested
	return project.UpdateWithBase(project.Provider(), project.Locked)
}

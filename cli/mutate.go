package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mdy/melody/project"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func addDependency(c *cli.Context) error {
	if len(c.Args()) < 1 {
		log.Fatalf("`add` command takes one or more packages. See '%s add --help'.", c.App.Name)
	}

	wDir, _ := os.Getwd()
	return runInstall(wDir, func(p *project.Project) error {
		for _, pkgName := range c.Args() {
			version := "" // Auto version, if unspecified
			if i := strings.Index(pkgName, "@"); i >= 0 {
				pkgName, version = pkgName[:i], pkgName[i+1:]
			}

			if err := p.AddDependency(pkgName, version); err != nil {
				return err
			}
		}
		return nil
	})
}

func removeDependency(c *cli.Context) error {
	if len(c.Args()) < 1 {
		log.Fatalf("`remove` command takes one or more packages. See '%s remove --help'.", c.App.Name)
	}

	wDir, _ := os.Getwd()
	return runInstall(wDir, func(p *project.Project) error {
		for _, pkgName := range c.Args() {
			if err := p.RemoveDependency(pkgName); err != nil {
				return err
			}
		}
		return nil
	})
}

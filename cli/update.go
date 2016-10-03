package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/melody-sh/melody/project"
	"github.com/melody-sh/melody/provider"
	"github.com/melody-sh/melody/resolver"
	"github.com/urfave/cli"
	"os"
)

func update(c *cli.Context) error {
	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	log.Info("Project", project, " -- ", err)

	var baseGraph *resolver.Graph
	if len(c.Args()) == 0 {
		baseGraph = resolver.NewGraph()
	} else {
		baseGraph = project.Locked.Dup()
		for _, name := range c.Args() {
			baseGraph.DetachVertexNamed(name)
		}
	}

	// Convert Project.Config to Requested
	source := provider.NewMelody(project.Locked)
	return project.UpdateWithBase(source, baseGraph)
}

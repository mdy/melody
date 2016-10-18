package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/melodysh/melody/project"
	"github.com/melodysh/melody/provider"
	"github.com/melodysh/melody/resolver"
	"github.com/urfave/cli"
	"os"
)

func update(c *cli.Context) error {
	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	log.Info("Project", project, " -- ", err)
	if err != nil {
		return err
	}

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

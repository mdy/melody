package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gobwas/glob"
	"github.com/mdy/melody/project"
	"github.com/mdy/melody/provider"
	"github.com/mdy/melody/resolver"
	"github.com/urfave/cli"
	"os"
	"strings"
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
		g, err := glob.Compile("{" + strings.Join(c.Args(), ",") + "}")
		if err != nil {
			return err
		}
		baseGraph = project.Locked.Dup()
		for _, spec := range project.Locked.Specifications() {
			if name := spec.Name(); g.Match(name) {
				baseGraph.DetachVertexNamed(name)
			}
		}
	}

	// Convert Project.Config to Requested
	source := provider.NewMelody(project.Locked)
	return project.UpdateWithBase(source, baseGraph)
}

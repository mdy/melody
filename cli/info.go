package cli

import (
	"github.com/mdy/melody/project"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strings"
	"text/template"
)

func info(c *cli.Context) error {
	infoFmt := "{{.Name}} {{.Version}}\n"

	if len(c.Args()) != 0 {
		infoFmt = strings.Join(c.Args(), " ")
	}

	// Load project info
	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	log.Info("Project", project, " -- ", err)
	if err != nil {
		return err
	}

	// Prepare output template
	t, err := template.New("info").Parse(infoFmt)
	if err != nil {
		return err
	}

	// Print info to STDOUT
	return t.Execute(os.Stdout, project.Config)
}

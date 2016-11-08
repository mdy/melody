package cli

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/mdy/melody/project"
	"github.com/mdy/melody/provider"
	"github.com/mdy/melody/resolver/types"
	"github.com/urfave/cli"
	"os"
)

// TODO: This should be a supported concept at resolver level
type released interface {
	ReleaseSpec() types.Specification
}

func list(c *cli.Context) error {
	if len(c.Args()) != 0 {
		log.Fatalf("`list` command takes no arguments. See '%s list --help'.", c.App.Name)
	}

	wDir, _ := os.Getwd()
	project, err := project.Load(wDir)
	if err != nil {
		return err
	}

	// Resolve if not locked
	if project.Locked == nil {
		source := provider.NewMelody(nil)
		project.Locked, err = project.Resolve(source, nil)
		if err != nil {
			return err
		}
	}

	specs := project.Locked.Specifications()
	if len(specs) == 0 {
		fmt.Println("♫ Simply no dependencies!")
		return nil
	}

	fmt.Println("♫ Dependencies for this project:")
	for _, s := range specs {
		if _, isPkg := s.(released); isPkg {
			fmt.Printf("  - %s\n", s.Name())
		}
	}

	return nil
}

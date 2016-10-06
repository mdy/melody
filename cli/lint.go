package cli

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/melody-sh/melody/project"
	//	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

func lint(c *cli.Context) error {
	if len(c.Args()) > 1 {
		log.Fatalf("`lint` command takes at most one argument. See '%s lint --help'.", c.App.Name)
	}

	// No arguments will return a blank path which becomes current working dir
	projectDir, err := filepath.Abs(c.Args().Get(0))
	if err != nil {
		return err
	}

	if !isGOPATHSubdir(projectDir) {
		return fmt.Errorf("Project should be in a $GOPATH/src subdirectory")
	}

	fmt.Println("♫ Checking Melody.toml")
	configPath := filepath.Join(projectDir, "Melody.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("Melody.toml does not exist in %s", projectDir)
	}

	proj, err := project.Load(projectDir)
	if err != nil {
		return fmt.Errorf(" Error reading Melody.toml: %s", err)
	}

	initConfig, err := initProjectConfig(projectDir)
	if err != nil {
		return err
	}

	if a, e := proj.Config.Name, initConfig.Name; a != e {
		fmt.Printf("  Name is \"%s\", should be \"%s\"\n", a, e)
	}

	return nil
}

// Check if the project is in a $GOPATH/src subdirectory
func isGOPATHSubdir(dir string) bool {
	dir = filepath.FromSlash(filepath.Clean("/" + dir))
	for _, p := range build.Default.SrcDirs() {
		if strings.HasPrefix(dir, p) {
			return true
		}
	}
	return false
}

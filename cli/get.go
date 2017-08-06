package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mdy/melody/project"
	"github.com/mdy/melody/provider"
	"github.com/mdy/melody/resolver/types"
	"github.com/urfave/cli"

	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func getPackages(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return fmt.Errorf("`get` command takes one or more packages. See '%s get --help'.", c.App.Name)
	}

	shouldUpdate := c.Bool("u")
	downloadOnly := c.Bool("d")
	goPathSrc := ""

	for _, dir := range build.Default.SrcDirs() {
		if !strings.HasPrefix(dir, runtime.GOROOT()) {
			goPathSrc = dir
		}
	}

	if goPathSrc == "" {
		return fmt.Errorf("Please set your GOPATH")
	}

	source := (&project.Project{}).Provider()
	for _, pkgName := range c.Args() {
		specs := source.SearchFor(source.NewRequirement(pkgName, "head"))
		if len(specs) == 0 {
			return fmt.Errorf("Package not found: %s", pkgName)
		}

		// Specs are ordered by version
		verSpec, ok := specs[len(specs)-1].(provider.VersionSpec)
		if !ok {
			return fmt.Errorf("Internal error. Unreleaseable spec.")
		}

		spec := verSpec.ReleaseSpec()

		// Check presence of directory and if it's a Melody install
		installPath := filepath.Join(goPathSrc, spec.InstallPath())
		if iStat, err := os.Stat(installPath); os.IsNotExist(err) {
			shouldUpdate = true // Force update if missing
		} else if iStat.IsDir() {
			_, err1 := os.Stat(filepath.Join(installPath, ".melody.ver"))
			_, err2 := os.Stat(filepath.Join(installPath, "Melody.toml"))
			if os.IsNotExist(err1) || os.IsNotExist(err2) {
				return fmt.Errorf("%s was previously installed without Melody", pkgName)
			}
		}

		// Update project to latest "head" from melodyAPI
		if shouldUpdate {
			log.Infof("Installing %s to %s", spec.Name(), installPath)
			err := source.InstallToDir(goPathSrc, []types.Specification{spec})
			if err != nil {
				return err
			}
		}

		// Save/restore current working directory
		if savedWd, err := os.Getwd(); err == nil {
			defer os.Chdir(savedWd)
		}

		// Initialize/load project
		os.Chdir(installPath)
		initProjectByName(installPath)
		project, err := project.Load(installPath)
		if err != nil {
			return err
		}

		// Force update on "-u" or new install
		if shouldUpdate {
			project.Locked = nil
		}

		// Install locked or update dependencies
		source = project.Provider()
		if err := project.UpdateWithBase(source, project.Locked); err != nil {
			return err
		}

		// Download-only flag
		if downloadOnly {
			continue
		}

		// Run 'go install'
		fmt.Printf("♫ Building and installing %s ...\n", pkgName)
		cmd := exec.Command("go", "install", ".", "./cmd/...")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

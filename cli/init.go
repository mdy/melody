package cli

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/melody-sh/melody/project"
	"github.com/melody-sh/melody/templates"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func initProject(c *cli.Context) error {
	if len(c.Args()) > 1 {
		log.Fatalf("`init` command takes at most one argument. See '%s init --help'.", c.App.Name)
	}

	// No arguments will return a blank path which becomes current working dir
	projectDir, err := filepath.Abs(c.Args().Get(0))
	if err != nil {
		return err
	}

	fmt.Printf("♫ Writing new Melody.toml to %s\n", projectDir)
	configPath := filepath.Join(projectDir, "Melody.toml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("Melody.toml already exists")
	}

	raw, err := templates.Asset("Melody.toml.tt")
	if err != nil {
		return errors.Wrap(err, "Internal error")
	}

	fm := template.FuncMap{"toml": tomlTemplateFunc}
	tmpl, err := template.New("config").Funcs(fm).Parse(string(raw))
	if err != nil {
		return errors.Wrap(err, "Internal error")
	}

	config, err := initProjectConfig(projectDir)
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Funcs(fm).Execute(file, config)
}

func tomlTemplateFunc(v interface{}) (string, error) {
	var buff bytes.Buffer
	err := toml.NewEncoder(&buff).Encode(v)
	return strings.TrimSpace(buff.String()), err
}

// TODO: Move this to project package w/ all things config
func initProjectConfig(dir string) (*project.Config, error) {
	pkg, err := build.Default.ImportDir(dir, build.FindOnly)
	config := &project.Config{Version: "0.1.0"}
	config.Name = pkg.ImportPath
	return config, err
}

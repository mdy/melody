package cli

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/melodysh/melody/internal/extract"
	"github.com/melodysh/melody/templates"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func initProject(c *cli.Context) error {
	if len(c.Args()) > 1 {
		log.Fatalf("`init` command takes at most one argument. See '%s init --help'.", c.App.Name)
	}

	// No arguments will return a blank path which uses current dir
	projectDir, err := mkdirForProject(c.Args().Get(0))
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

	config, err := extract.ProjectConfig(projectDir)
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

// Figure out best possible project directory from argument
func mkdirForProject(name string) (string, error) {
	dir, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	// If named directory exists as is, let's use that
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return dir, nil
	}

	// If it looks like an ImportPath, let's put it under $GOPATH/src
	name = filepath.Clean(name)
	if gopath := os.Getenv("GOPATH"); gopath != "" && isPossibleImportPath(name) {
		dir = filepath.Join(filepath.SplitList(gopath)[0], "src", name)
	}

	// Maybe the user meant that we should create the directory too
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return "", err
	}

	return dir, nil
}

// TODO: Move this to project package w/ all things config
func isPossibleImportPath(path string) bool {
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	return strings.Contains(path[:i], ".")
}

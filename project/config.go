package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/melodysh/melody/resolver"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	melodyFile = "Melody.toml"
	lockedFile = "Melody.lock"
)

type Project struct {
	// Melody.toml
	Config Config

	// Locked dependencies graph
	Locked *resolver.Graph

	// Root directory
	root string
}

type Config struct {
	Name         string            `toml:"name"`
	Version      string            `toml:"version"`
	Authors      []string          `toml:"authors"`
	Dependencies map[string]string `toml:"dependencies,omitempty"`
}

type Locked struct {
	Speficiations map[string]string
}

func Load(root string) (*Project, error) {
	tomlConfig := &tomlRootConfig{}
	cPath := filepath.Join(root, melodyFile)
	if err := loadTOMLFile(cPath, tomlConfig); err != nil {
		return nil, err
	}

	// LEGACY/DEPRECATED CONFIG
	if tomlConfig.Package != nil {
		tomlConfig.Project = *tomlConfig.Package
	}

	project := &Project{root: root}
	project.Config = tomlConfig.Project
	project.Config.Dependencies = tomlConfig.Dependencies

	lPath := filepath.Join(root, lockedFile)
	if info, err := os.Stat(lPath); os.IsNotExist(err) {
		return project, nil
	} else if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", lockedFile)
	} else if err != nil {
		return nil, err
	}

	project.Locked = resolver.NewGraph()
	if err := project.LoadLockfile(lPath); err != nil {
		return nil, err
	}

	return project, nil
}

func (p *Project) Save() error {
	lPath := filepath.Join(p.root, lockedFile)
	return p.SaveLockfile(lPath)
}

// === DEBUG HELPERS ===
func (p *Project) dumpGraph() string {
	// Marshal specifications
	specs := p.Locked.Specifications()
	raw, _ := json.Marshal(&specs)

	var pretty bytes.Buffer
	json.Indent(&pretty, raw, "", "  ")
	return string(pretty.Bytes())
}

// === CONFIGURATION MARSHAL/UNMARSHAL ===
func loadTOMLFile(path string, data interface{}) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(raw, data); err != nil {
		return err
	}
	return nil
}

// Unmarshal Config by splitting out [package] info
type tomlRootConfig struct {
	Project      Config
	Dependencies map[string]string

	// DEPRECATED: Use Project
	Package *Config
}

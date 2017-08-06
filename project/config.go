package project

import (
	"github.com/BurntSushi/toml"
	"github.com/mdy/melody/provider"
	"github.com/mdy/melody/provider/melody"
	"github.com/mdy/melody/resolver"

	"bytes"
	"encoding/json"
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
	Config      Config
	configData  []byte
	configDirty bool

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
	raw, err := ioutil.ReadFile(filepath.Join(root, melodyFile))
	if err != nil {
		return nil, err
	}

	project := &Project{root: root, configData: raw}
	if err := project.parseConfig(); err != nil {
		return nil, err
	}

	project.Locked = resolver.NewGraph()
	err = project.LoadLockfile(filepath.Join(root, lockedFile))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return project, nil
}

// Initialize Specification provider for this project
func (p *Project) Provider() provider.Provider {
	return melody.New(p.Locked)
}

func (p *Project) parseConfig() error {
	tomlConfig := &tomlRootConfig{}
	if err := toml.Unmarshal(p.configData, tomlConfig); err != nil {
		return err
	}

	// LEGACY/DEPRECATED CONFIG
	if tomlConfig.Package != nil {
		tomlConfig.Project = *tomlConfig.Package
	}

	p.Config = tomlConfig.Project
	p.Config.Dependencies = tomlConfig.Dependencies
	return nil
}

func (p *Project) Save() error {
	if err := p.saveConfig(); err != nil {
		return err
	}
	return p.saveLockfile()
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

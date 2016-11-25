package project

import (
	"fmt"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"io/ioutil"
	"path/filepath"
)

func (p *Project) AddDependency(name, version string) error {
	return p.mutateDeps(func(buf []byte, t *ast.Table) ([]byte, error) {
		// Validate package existence and generate Melody.toml line
		bytes, err := p.lookupDependency(name, version)
		if err != nil {
			return buf, err
		}

		// Start a new dependencies section, if needed
		deps, ok := t.Fields["dependencies"].(*ast.Table)
		if !ok {
			bytes = append([]byte("[dependencies]"), bytes...)
			return append(buf, bytes...), nil
		}

		// Add or replace dependency if it already exists
		if old, ok := deps.Fields[name].(*ast.KeyValue); !ok {
			at := deps.Position.End // End of dependencies section
			buf = append(buf[:at], append(bytes, buf[at:]...)...)
		} else if version != "" {
			buf = replaceLine(buf, old.Line, bytes)
		}

		return buf, nil
	})
}

func (p *Project) RemoveDependency(name string) error {
	return p.mutateDeps(func(buf []byte, t *ast.Table) ([]byte, error) {
		deps, ok := t.Fields["dependencies"].(*ast.Table)
		if !ok {
			return nil, fmt.Errorf("No [dependencies] in %s", melodyFile)
		}

		old, ok := deps.Fields[name].(*ast.KeyValue)
		if !ok {
			return nil, fmt.Errorf("Didn't find %s dependency", name)
		}

		// Remove dependency if it already exists
		return replaceLine(buf, old.Line, []byte{}), nil
	})
}

func (p *Project) saveConfig() error {
	if !p.configDirty {
		return nil
	}
	cPath := filepath.Join(p.root, melodyFile)
	return ioutil.WriteFile(cPath, p.configData, 0644)
}

func (p *Project) mutateDeps(mutate func([]byte, *ast.Table) ([]byte, error)) error {
	t, err := toml.Parse(p.configData)
	if err != nil {
		return err
	}

	// Execute mutation on config
	buf, err := mutate(p.configData, t)
	if err != nil {
		return err
	}

	// Save and reparse
	p.configData = buf
	p.configDirty = true
	return p.parseConfig()
}

// Generates a line to insert into "[dependencies]" of Melody.toml
// TODO: This should check existence of dependency and generate range
func (p *Project) lookupDependency(name, version string) ([]byte, error) {
	if version == "" {
		version = "head"
	}

	return []byte(fmt.Sprintf("\n\"%s\"=\"%s\"", name, version)), nil
}

// Replace a specific numbered line in buf with line.  If the
// line number doesn't exist, line is appended to the end
func replaceLine(buf []byte, num int, line []byte) []byte {
	curLine, startAt, endAt := 1, len(buf), len(buf)

	for pos, char := range string(buf) {
		if char != '\n' {
			continue
		}

		curLine++
		if curLine == num {
			startAt = pos
		} else if curLine == num+1 {
			endAt = pos
			break
		}
	}

	return append(buf[:startAt], append(line, buf[endAt:]...)...)
}

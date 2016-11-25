package project

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mdy/melody/provider/melody"
	"github.com/mdy/melody/resolver"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	lockFileVersion  = "0.1.0"
	lockFilePreamble = "# AUTO-GENERATED: Do not modify\n"
)

// Marshalling/unmarshalling of JSON testdata
func (p *Project) LoadLockfile(path string) error {
	if info, err := os.Stat(path); err != nil {
		return err
	} else if info.IsDir() {
		return fmt.Errorf("%s is a directory", lockedFile)
	}

	builder := &LockEncoderDecoder{path: path}
	graph, err := resolver.DecodeGraph(builder)
	p.Locked = graph
	return err
}

func (p *Project) saveLockfile() error {
	path := filepath.Join(p.root, lockedFile)
	encoder := &LockEncoderDecoder{path: path, config: &p.Config}
	return p.Locked.Encode(encoder)
}

// Graph encoder/decoder
type LockEncoderDecoder struct {
	path           string  // Lockfile path
	config         *Config // Project config
	melody.Builder         // provides NewSpec(...)
}

func (l *LockEncoderDecoder) Decode(v interface{}) error {
	return loadTOMLFile(l.path, v)
}

func (l *LockEncoderDecoder) Encode(v interface{}) error {
	if g, ok := v.(*resolver.EncodedGraph); ok {
		g.Project.Name = l.config.Name
		g.Project.Version = l.config.Version
		g.Version = lockFileVersion
	}

	var output bytes.Buffer
	output.WriteString(lockFilePreamble)
	err := toml.NewEncoder(&output).Encode(v)
	if err != nil {
		return err
	}

	fixed := multiLineDependencies(output.String())
	return ioutil.WriteFile(l.path, []byte(fixed), 0644)
}

// Make all TOML dependency arrays multi-line
// FIXME: Write a MarshalTOML when that's supported
func multiLineDependencies(src string) string {
	re := regexp.MustCompile(`(?m)^(\s*)dependencies\s*=\s*\[.*\]\s*$`) //\s*=\s*\[.*\]\s*$
	return re.ReplaceAllStringFunc(string(src), func(str string) string {
		indent := strings.Repeat(" ", strings.IndexRune(str, 'd'))
		out := strings.Replace(str, `["`, "[\n  "+indent+`"`, -1)
		out = strings.Replace(out, `", "`, "\",\n  "+indent+`"`, -1)
		out = strings.Replace(out, `"]`, "\"\n"+indent+`]`, -1)
		return out
	})
}

package rubygem

import (
	"github.com/mdy/melody/resolver/flex"
	"github.com/mdy/melody/version"
	"strings"
)

// Global version parser
var VersionParser version.Parser = &Version{}

// RubyGems compatible version class that loosens
// beta version matching from Flex versioning
type Version struct {
	flex.Version
}

func (v Version) Parse(str string) (version.Version, error) {
	vFlex, err := flex.VersionParser.Parse(str)
	return Version{vFlex.(flex.Version)}, err
}

func (v Version) Compare(w version.Version) int {
	out := v.Version.Compare(w.(Version).Version)
	if out < 0 {
		return -1
	} else if out > 0 {
		return 1
	}
	return out
}

func (v Version) MajorBump() version.Version {
	bump := v.Version.MajorBump()
	return Version{bump.(flex.Version)}
}

func (v Version) MinorBump() version.Version {
	var bump version.Version
	if len(v.Main) < 3 {
		bump = v.Version.MajorBump()
	} else {
		bump = v.Version.MinorBump()
	}
	return Version{bump.(flex.Version)}
}

// Convert to string for testing
func (v Version) String() string {
	fullStr := v.Version.String()
	if len(v.Main) >= 3 {
		return fullStr
	}

	// Find the ending of Main part
	i := strings.IndexAny(fullStr, "-+")
	if i < 0 {
		i = len(fullStr)
	}

	// Pad with ".0" to 3 main components
	pad := strings.Repeat(".0", 3-len(v.Main))
	return fullStr[0:i] + pad + fullStr[i:]
}

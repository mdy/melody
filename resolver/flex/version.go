package flex

import (
	"bytes"
	"github.com/mdy/melody/version"
	"strconv"
	"strings"
	tscanner "text/scanner"
)

// Global version parser
var VersionParser version.Parser = &Version{}

// Flex version (SemVer-compatible)
type Version struct {
	Main  []part
	Pre   []part
	Build []part
}

func ParseVersion(src string) (Version, error) {
	s := (&tscanner.Scanner{}).Init(strings.NewReader(src))
	s.Mode = tscanner.ScanIdents | tscanner.ScanInts
	parts := [][]part{{}, {}, {}}

	for i := 0; ; {
		tok, text := s.Scan(), s.TokenText()
		if tok == tscanner.EOF {
			break
		} else if tok == '-' && i < 1 {
			i = 1
		} else if tok == '+' && i < 2 {
			i = 2
		} else if tok == '-' || tok == '+' || tok == '.' {
			continue
		} else if tok == tscanner.Int {
			num, _ := strconv.ParseUint(text, 10, 64)
			parts[i] = append(parts[i], part{num: num, isNum: true})
		} else {
			parts[i] = append(parts[i], part{str: text, isNum: false})
		}
	}

	return Version{parts[0], parts[1], parts[2]}, nil
}

// MustParse is like Parse with panic on error
func MustParseVersion(s string) (v version.Version) {
	var err error
	if v, err = ParseVersion(s); err == nil {
		return v
	}
	panic(`flex.ParseVersion(` + s + `): ` + err.Error())
}

// Allow Version to fulfill Parser interface
func (v Version) Parse(str string) (version.Version, error) {
	return ParseVersion(str)
}

// Fulfill Version interface for version Range
func (v Version) Compare(w version.Version) int {
	o, ok := w.(Version)
	if !ok {
		return -99
	}

	// Pad these to be the same length
	mainV := padParts(v.Main, len(o.Main))
	mainO := padParts(o.Main, len(v.Main))

	// Compare main components, then beta components
	if mainC := compareParts(mainV, mainO); mainC != 0 {
		return mainC * 2 // Main versions differ
	} else if len(v.Pre) == 0 && len(o.Pre) > 0 {
		return 1
	} else if len(o.Pre) == 0 && len(v.Pre) > 0 {
		return -1
	}

	return compareParts(v.Pre, o.Pre)
}

// Bump major version for caret comparison
func (v Version) IsPrerelease() bool {
	return len(v.Pre) != 0
}

// Stricter check whether it's a SemVer
func (v Version) IsSemver() bool {
	for _, p := range v.Main {
		if !p.isNum {
			return false
		}
	}

	return len(v.Main) == 3
}

// Bump major version for caret comparison
func (v Version) MajorBump() version.Version {
	major := padParts(v.Main, 1)[0].num + 1
	main := []part{{num: major, isNum: true}}
	return Version{main, v.Pre, v.Build}
}

// Bump major version for tilde comparison
func (v Version) MinorBump() version.Version {
	main := append([]part(nil), padParts(v.Main, 2)[:2]...)
	main[1] = part{num: main[1].num + 1, isNum: true}
	return Version{main, v.Pre, v.Build}
}

// Convert to string for debugging. This will likely not be
// the same as original parsed string (different separators)
func (v Version) String() string {
	buffer := bytes.NewBufferString("")
	stringAppendParts(buffer, "", v.Main)
	stringAppendParts(buffer, "-", v.Pre)
	stringAppendParts(buffer, "+", v.Build)
	return buffer.String()
}

func stringAppendParts(buffer *bytes.Buffer, prefix string, parts []part) {
	for i := 0; i < len(parts); i++ {
		if i == 0 {
			buffer.WriteString(prefix)
		} else {
			buffer.WriteString(".")
		}
		if p := parts[i]; p.isNum {
			buffer.WriteString(strconv.FormatUint(p.num, 10))
		} else {
			buffer.WriteString(p.str)
		}
	}
}

// Version segment (either int or string)
type part struct {
	str   string
	num   uint64
	isNum bool
}

// Zero-pad version []part to target length
func padParts(v []part, tgt int) []part {
	pad := part{num: 0, isNum: true}
	for len(v) < tgt {
		v = append(v, pad)
	}
	return v
}

// Compare array of segments/parts
func compareParts(v, o []part) int {
	for i := 0; i < len(v) && i < len(o); i++ {
		vI, oI := v[i], o[i]

		if vI.isNum && !oI.isNum {
			return -1
		} else if !vI.isNum && oI.isNum {
			return 1
		} else if vI.isNum && oI.isNum {
			if vI.num == oI.num {
				continue
			} else if vI.num > oI.num {
				return 1
			} else {
				return -1
			}
		} else { // both are strings
			if vI.str == oI.str {
				continue
			} else if vI.str > oI.str {
				return 1
			} else {
				return -1
			}
		}
	}

	// If prefix is the same, shortest wins
	if len(v) == len(o) {
		return 0
	} else if len(v) < len(o) {
		return -1
	} else {
		return 1
	}
}

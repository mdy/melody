package version

import (
	"fmt"
	"strings"
)

type comparator func(Version, Version) bool

var comparators = map[Token]comparator{
	EQ: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) == 0
	},
	NEQ: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) != 0
	},
	GT: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) > 0
	},
	GTE: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) >= 0
	},
	LT: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) < 0
	},
	LTE: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) <= 0
	},
	CARET: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) >= 0 && v1.Compare(v2.MajorBump()) < 0
	},
	TILDE: func(v1 Version, v2 Version) bool {
		return v1.Compare(v2) >= 0 && v1.Compare(v2.MinorBump()) < 0
	},
}

// Version object that can also parse
type ParsingVersion interface {
	Version
	Parser
}

// SatisfiesRange parses a range and checks Version against it
func SatisfiesRange(v ParsingVersion, rStr string) (bool, error) {
	if rStr = strings.Trim(rStr, " "); rStr == "" {
		return !v.IsPrerelease(), nil
	} else if r, err := ParseRange(rStr, v); err == nil {
		return r(v), nil
	} else {
		return false, err
	}
}

// ParseRange parses a range and returns a Range.
// If the range could not be parsed an error is returned.
func ParseRange(str string, verPar Parser) (Range, error) {
	var tok Token
	var lit string
	var outputFn, andFn Range
	currentOp := EQ // Default if omitted

	// Rewrite shortcuts into valid ranges
	if rewrite, ok := shortcuts[str]; ok {
		str = rewrite
	}

	// Use text/scanner to parse versions and operators
	for s := newRangeScanner(strings.NewReader(str)); tok != EOF; {
		tok, _, lit = s.Scan()

		if tok == INVALID {
			return nil, fmt.Errorf("Illegal token: %s", lit)
		} else if tok == WS {
			continue
		} else if tok == OR {
			outputFn, andFn = outputFn.OR(andFn), nil
		} else if tok != VERSION {
			currentOp = tok
		} else { // tok == VERSION
			version, err := verPar.Parse(lit)
			if err != nil {
				return nil, err
			}

			// Convert operator into comparator
			comp, ok := comparators[currentOp]
			if !ok {
				return nil, fmt.Errorf("Invalid comparator: %s", currentOp)
			}

			// Convert comparator into a Range
			andFn = andFn.AND(comp.toRangeFunc(version))
			currentOp = EQ // Reset for next section
		}
	}

	return outputFn.OR(andFn), nil
}

// Create a Range for a comparator/version pair
func (c comparator) toRangeFunc(v2 Version) Range {
	return Range(func(v1 Version) bool {
		pre1, pre2 := v1.IsPrerelease(), v2.IsPrerelease()
		if pre1 && pre2 { // Both are prerelease
			diff := v1.Compare(v2)
			return diff > -2 && diff < 2 && c(v1, v2)
		}
		return (pre2 || !pre1) && c(v1, v2)
	})
}

// Range matching function
type Range func(Version) bool

// Logical OR between two ranges
func (rf Range) OR(f Range) Range {
	if rf == nil {
		return f
	}
	return Range(func(v Version) bool {
		return rf(v) || f(v)
	})
}

// Logical OR between two ranges
func (rf Range) AND(f Range) Range {
	if rf == nil {
		return f
	}
	return Range(func(v Version) bool {
		return rf(v) && f(v)
	})
}

// Shortcut range helpers
var shortcuts = map[string]string{
	"*": ">=0.0.0",
	"":  ">=0.0.0",
}

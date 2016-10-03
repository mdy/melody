package version

import (
	"fmt"
)

// Version interface
type Version interface {
	Compare(Version) int
	MajorBump() Version
	MinorBump() Version
	IsPrerelease() bool
	fmt.Stringer
}

type Parser interface {
	Parse(string) (Version, error)
}

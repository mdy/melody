package types

import (
	"fmt"
)

// Interface for name/version/dependencies
type Specification interface {
	Name() string
	Version() string
	Requirements() Requirements
	fmt.Stringer
}

// Comparison helpers
func SpecEqual(s, t Specification) bool {
	return s.Name() == t.Name() && s.Version() == t.Version()
}

// Interface for Spec[name,version] requirement
type Requirement interface {
	SatisfiedBy(Specification) (bool, error)
	Name() string
	fmt.Stringer
}

// List of name/requirement pairs
type Requirements []Requirement

func (r Requirements) Dup() Requirements {
	return append(Requirements{}, r...)
}

// Len() for sort.Interface
func (r Requirements) Len() int {
	return len(r)
}

// Swap() for sort.Interface
func (r Requirements) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

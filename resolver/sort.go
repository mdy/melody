package resolver

import (
	"github.com/melody-sh/melody/resolver/types"
	"github.com/melody-sh/melody/version"
	"sort"
)

// Comparison helpers
func SpecEqual(s, t types.Specification) bool {
	return s.Name() == t.Name() && s.Version() == t.Version()
}

// FIXME: Should not need a parser if Version()
// returns a version.Version object, not string
func SortSpecs(v []types.Specification, p version.Parser) {
	sort.Sort(specsVesionSort{v, p})
}

// Sorting versioned objects
type specsVesionSort struct {
	s []types.Specification
	p version.Parser
}

func (s specsVesionSort) Len() int      { return len(s.s) }
func (s specsVesionSort) Swap(i, j int) { s.s[i], s.s[j] = s.s[j], s.s[i] }
func (s specsVesionSort) Less(i, j int) bool {
	vI, _ := s.p.Parse(s.s[i].Version())
	vJ, _ := s.p.Parse(s.s[j].Version())
	return vI.Compare(vJ) < 0
}

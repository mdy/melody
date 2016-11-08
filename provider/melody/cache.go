package melody

import (
	"github.com/mdy/melody/resolver"
	"github.com/mdy/melody/resolver/flex"
	"github.com/mdy/melody/resolver/types"
)

// External function used to fetch available specifications
type CacheFetchFunc func(string) ([]types.Specification, error)

// The cache!
type Cache struct {
	content   map[string][]types.Specification
	fetched   map[string]struct{}
	fetchFunc CacheFetchFunc
}

func NewCache(fetchFunc CacheFetchFunc) *Cache {
	fetched := map[string]struct{}{}
	content := map[string][]types.Specification{}
	return &Cache{content, fetched, fetchFunc}
}

func (c *Cache) Fetch(name string) ([]types.Specification, error) {
	if _, ok := c.fetched[name]; !ok {
		specs, err := c.fetchFunc(name)
		if err != nil {
			return nil, err
		}

		c.fetched[name] = struct{}{}
		c.Append(name, specs)
	}

	return c.content[name], nil
}

// Append new specifications, sort and dedupe
func (c *Cache) Append(name string, s []types.Specification) {
	specs := append(c.content[name], s...)
	if specs == nil || len(specs) == 0 {
		return
	}

	resolver.SortSpecs(specs, flex.VersionParser)
	c.content[name] = specs[:0]
	for i, s := range specs {
		if i == 0 || s.Version() != specs[i-1].Version() {
			c.content[name] = append(c.content[name], s)
		}
	}
}

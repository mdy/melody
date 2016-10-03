package rubygem

import (
	"testing"
)

func TestVersionString(t *testing.T) {
	verKlass := &Version{}
	tests := []struct {
		i string
		j string
	}{
		// Releasable versions
		{"1", "1.0.0"},
		{"1.2", "1.2.0"},
		{"1.2.3", "1.2.3"},
		// Beta versions
		{"1-beta.1", "1.0.0-beta.1"},
		{"1.2-beta.1", "1.2.0-beta.1"},
		{"1.2.3-beta.1", "1.2.3-beta.1"},
		{"1.2-beta.1+3", "1.2.0-beta.1+3"},
		// Build versions
		{"1+abc123", "1.0.0+abc123"},
		{"1.2+abc123", "1.2.0+abc123"},
		{"1.2.3+abc123", "1.2.3+abc123"},
	}

	for _, tc := range tests {
		r, err := verKlass.Parse(tc.i)
		if err != nil {
			t.Errorf("Error parsing version %q: %s", tc.i, err)
		} else if str := r.String(); str != tc.j {
			t.Errorf("Invalid String(): %s vs. %s", str, tc.j)
		}
	}
}

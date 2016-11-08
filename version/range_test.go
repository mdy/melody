package version_test

import (
	"github.com/mdy/melody/resolver/flex"
	"github.com/mdy/melody/version"
	"testing"
)

func TestParseRange(t *testing.T) {
	verParser := flex.VersionParser
	type tv struct {
		v string
		b bool
	}
	tests := []struct {
		i string
		t []tv
	}{
		// Simple expressions
		{">1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", false},
			{"1.2.4", true},
			{"1.2.5-beta", false},
		}},
		{">=1.2.3", []tv{
			{"1.2.3", true},
			{"1.2.4", true},
			{"1.2.2", false},
		}},
		{"<1.2.3", []tv{
			{"1.2.2", true},
			{"1.2.3", false},
			{"1.2.4", false},
		}},
		{"<=1.2.3", []tv{
			{"1.2.2", true},
			{"1.2.3", true},
			{"1.2.4", false},
		}},
		{"1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
		}},
		{"=1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
		}},
		{"==1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
		}},
		{"!=1.2.3", []tv{
			{"1.2.2", true},
			{"1.2.3", false},
			{"1.2.4", true},
		}},
		{"!1.2.3", []tv{
			{"1.2.2", true},
			{"1.2.3", false},
			{"1.2.4", true},
		}},
		{"^1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", true},
			{"2.0.0", false},
		}},
		{"~1.2.3", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", true},
			{"1.3.0", false},
		}},

		// Shortcut Expressions
		{"", []tv{
			{"0.4.3", true},
			{"0.4.4-beta", false},
		}},

		{"*", []tv{
			{"0.4.3", true},
			{"0.4.4-beta", false},
		}},

		// Prerelease Expression
		{">=1.4.2-beta.2", []tv{
			{"1.3.2", false},
			{"1.4.3", true},
			{"1.4.0-beta", false},
			{"1.4.2-beta", false},
			{"1.4.2-beta.3", true},
			{"1.4.3-beta", false},
		}},

		// Simple Expression errors
		{">>1.2.3", nil},
		{"!1.2.3", nil},
		{"1.0", nil},
		{"string", nil},

		// AND Expressions
		{">1.2.2 <1.2.4", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
		}},
		{"<1.2.2 <1.2.4", []tv{
			{"1.2.1", true},
			{"1.2.2", false},
			{"1.2.3", false},
			{"1.2.4", false},
		}},
		{">1.2.2 <1.2.5 !=1.2.4", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
			{"1.2.5", false},
		}},
		{">1.2.2 <1.2.5 !1.2.4", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
			{"1.2.5", false},
		}},
		// OR Expressions
		{">1.2.2 || <1.2.4", []tv{
			{"1.2.2", true},
			{"1.2.3", true},
			{"1.2.4", true},
		}},
		{"<1.2.2 || >1.2.4", []tv{
			{"1.2.2", false},
			{"1.2.3", false},
			{"1.2.4", false},
		}},
		// Combined Expressions
		{">1.2.2 <1.2.4 || >=2.0.0", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
			{"2.0.0", true},
			{"2.0.1", true},
		}},
		{">1.2.2 <1.2.4 || >=2.0.0 <3.0.0", []tv{
			{"1.2.2", false},
			{"1.2.3", true},
			{"1.2.4", false},
			{"2.0.0", true},
			{"2.0.1", true},
			{"2.9.9", true},
			{"3.0.0", false},
		}},
	}

	for _, tc := range tests {
		r, err := version.ParseRange(tc.i, verParser)
		if err != nil && tc.t != nil {
			t.Errorf("Error parsing range %q: %s", tc.i, err)
			continue
		}
		for _, tvc := range tc.t {
			v, _ := verParser.Parse(tvc.v)
			if res := r(v); res != tvc.b {
				t.Errorf("Invalid for case %q matching %q: Expected %t, got: %t", tc.i, tvc.v, tvc.b, res)
			}
		}

	}
}

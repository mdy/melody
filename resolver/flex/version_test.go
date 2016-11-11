package flex

import (
	"reflect"
	"testing"
)

func TestVersionParse(t *testing.T) {
	type tv struct {
		v string
		b bool
	}
	tests := []struct {
		i string
		j bool
		k *Version
	}{
		// Releasable versions
		{"1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{}, []part{}}},
		{"1.2", false, &Version{[]part{
			{num: 1, isNum: true},
			{num: 2, isNum: true},
		}, []part{}, []part{}}},
		{"1.2.3", true, &Version{[]part{
			{num: 1, isNum: true},
			{num: 2, isNum: true},
			{num: 3, isNum: true},
		}, []part{}, []part{}}},
		{"1.2.3.4", false, &Version{[]part{
			{num: 1, isNum: true},
			{num: 2, isNum: true},
			{num: 3, isNum: true},
			{num: 4, isNum: true},
		}, []part{}, []part{}}},

		// Beta versions
		{"1-1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{num: 1, isNum: true},
		}, []part{}}},
		{"1-beta", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{str: "beta", isNum: false},
		}, []part{}}},
		{"1-beta1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{str: "beta1", isNum: false},
		}, []part{}}},
		{"1-beta.1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{str: "beta", isNum: false},
			{num: 1, isNum: true},
		}, []part{}}},
		{"1-beta-rc", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{str: "beta", isNum: false},
			{str: "rc", isNum: false},
		}, []part{}}},

		// Build versions
		{"1+1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{}, []part{
			{num: 1, isNum: true},
		}}},
		{"1-beta+hi", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{
			{str: "beta", isNum: false},
		}, []part{
			{str: "hi", isNum: false},
		}}},
		{"1+hi.1", false, &Version{[]part{
			{num: 1, isNum: true},
		}, []part{}, []part{
			{str: "hi", isNum: false},
			{num: 1, isNum: true},
		}}},
	}

	for _, tc := range tests {
		r, err := VersionParser.Parse(tc.i)
		if err != nil && tc.k != nil {
			t.Errorf("Error parsing version %q: %s", tc.i, err)
		} else if r.(Version).IsSemver() != tc.j {
			t.Errorf("%q IsSemVer() should return: %t", tc.i, tc.j)
		} else if !reflect.DeepEqual(r, *(tc.k)) {
			t.Errorf("Invalid parts for %s, %v vs. %v", tc.i, r, *(tc.k))
		}
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		i string
		j string
		k int
	}{
		// Releasable
		{"1", "1", 0},
		{"1", "1.2", -2},
		{"1", "1.2.3", -2},
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -2},
		{"1.2.4", "1.2.3", 2},
		{"1.2.3", "1.2", 2},
		{"1.2.3", "1", 2},
		// Auto-padding
		{"1", "1.0", 0},
		{"1.0", "1.0.0", 0},
		// Beta vs. release
		{"1-beta", "1.2.3", -2},
		{"1.2-beta", "1.2.3", -2},
		{"1.2.3-beta", "1.2.3", -1},
		{"1.2.4-beta", "1.2.3", 2},
		// Beta vs. beta
		{"1-beta", "1.0-beta", 0},
		{"1.2-beta", "1.3-beta", -2},
		{"1.2.4-beta", "1.2.3-beta", 2},
		{"1.2-beta.1942", "1.2-beta.1941", 1},
		{"1.2.4-beta", "1.2.3", 2},
		// Build versions
		{"1+1", "1+2", 0},
		{"1.2.3+hi", "1.2.3+9", 0},
		{"1.2+99", "1.1+999", 2},
	}

	for _, tc := range tests {
		if v1, err := ParseVersion(tc.i); err != nil {
			t.Errorf("Error parsing version %q: %s", tc.i, err)
		} else if v2, err := ParseVersion(tc.j); err != nil {
			t.Errorf("Error parsing version %q: %s", tc.j, err)
		} else if k := v1.Compare(v2); k != tc.k {
			t.Errorf("Failed compare (%s, %s) => (%d, %d)", tc.i, tc.j, k, tc.k)
		}
	}
}

func TestVersionPadding(t *testing.T) {
	v1, _ := ParseVersion("1")
	len1 := len(v1.Main)
	if v1.Compare(MustParseVersion("1.0.0")); len(v1.Main) != len1 {
		t.Errorf("Compare modified Version by padding")
	}
}

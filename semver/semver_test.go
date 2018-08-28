package semver_test

import (
	"testing"

	"github.com/lpar/goup/semver"
)

func TestSemVerParse(t *testing.T) {

	type TestCase struct {
		SemVer string
		X      int
		Y      int
		Z      int
	}

	data := []TestCase{
		{"0.1", 0, 1, 0},
		{"2.1", 2, 1, 0},
		{"0.1.0", 0, 1, 0},
		{"1.0.0", 1, 0, 0},
		{"1.2.0", 1, 2, 0},
		{"1.2.3", 1, 2, 3},
		{"11.22.33", 11, 22, 33},
		{"01.02.03", 1, 2, 3}, // no octal!
		{"thing11.22.33beta4", 11, 22, 33},
	}

	for _, tc := range data {
		s := semver.NewSemVer(tc.SemVer)
		if s.Major != tc.X || s.Minor != tc.Y || s.Patch != tc.Z {
			t.Errorf("failed to parse semver %s: expected %d.%d.%d, got %d.%d.%d", tc.SemVer, tc.X, tc.Y, tc.Z, s.Major, s.Minor, s.Patch)
		}
	}
}

func TestStringer(t *testing.T) {
	data := []string{
		"0.1.0",
		"1.2.3",
		"10.22.33",
	}
	for _, sx := range data {
		sv := semver.NewSemVer(sx)
		nx := sv.String()
		if sx != nx {
			t.Errorf("failed to String() semver: expected %s got %s", sx, nx)
		}
	}
}

func TestSemVerOrder(t *testing.T) {

	type TestCase struct {
		SemVerA string
		SemVerB string
		AltB    bool
		AeqB    bool
	}

	data := []TestCase{
		{"0.9.9", "1.0.0", true, false},
		{"2.0.0", "2.0.1", true, false},
		{"2.0.0", "2.1.0", true, false},
		{"2.1.2", "3.0.0", true, false},
		// Reverse of the above
		{"1.0.0", "0.9.9", false, false},
		{"2.0.1", "2.0.0", false, false},
		{"2.1.0", "2.0.0", false, false},
		{"3.0.0", "2.1.2", false, false},
		// And check equality works
		{"1.2.3", "1.2.3", false, true},
	}

	for _, tc := range data {
		a := semver.NewSemVer(tc.SemVerA)
		b := semver.NewSemVer(tc.SemVerB)
		gt := a.GreaterThan(b)
		lt := a.LessThan(b)
		eq := a.Equals(b)
		if lt != tc.AltB {
			t.Errorf("%s > %s expected %v got %v", tc.SemVerA, tc.SemVerB, tc.AltB, lt)
		}
		if eq != tc.AeqB {
			t.Errorf("%s == %s expected %v got %v", tc.SemVerA, tc.SemVerB, tc.AeqB, eq)
		}
		if !eq && gt == lt {
			t.Errorf("%s both < and > %s!", tc.SemVerA, tc.SemVerB)
		}
	}

}

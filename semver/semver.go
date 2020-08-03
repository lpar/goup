// Package semver implements semver-style version numbers, with support for
// greater-than, less-than and equals comparisons.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVer represents a SemVer-standard version number.
type SemVer struct {
	Major int
	Minor int
	Patch int
}

var semverRE = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

// parse parses a version string into a SemVer struct.
func parse(sv string) SemVer {
	v := semverRE.FindString(sv)
	xx := strings.Split(v+".0.0", ".")
	xa, _ := strconv.Atoi(xx[0])
	xb, _ := strconv.Atoi(xx[1])
	xc, _ := strconv.Atoi(xx[2])
	return SemVer{
		Major: xa,
		Minor: xb,
		Patch: xc,
	}
}

// NewSemVer creates a SemVer struct from a version string.
func NewSemVer(x string) SemVer {
	return parse(x)
}

func (a SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", a.Major, a.Minor, a.Patch)
}

func (a SemVer) Equals(b SemVer) bool {
	if a.Major == b.Major &&
			a.Minor == b.Minor &&
			a.Patch == b.Patch {
		return true
	}
	return false
}

func (a SemVer) LessThan(b SemVer) bool {
	if a.Major < b.Major {
		return true
	}
	if a.Major > b.Major {
		return false
	}
	if a.Minor < b.Minor {
		return true
	}
	if a.Minor > b.Minor {
		return false
	}
	if a.Patch < b.Patch {
		return true
	}
	if a.Patch > b.Patch {
		return false
	}
	return false // they're equal
}

func (a SemVer) GreaterThan(b SemVer) bool {
	if a.Major > b.Major {
		return true
	}
	if a.Major < b.Major {
		return false
	}
	if a.Minor > b.Minor {
		return true
	}
	if a.Minor < b.Minor {
		return false
	}
	if a.Patch > b.Patch {
		return true
	}
	if a.Patch < b.Patch {
		return false
	}
	return false // they're equal
}

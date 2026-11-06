package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// parseSemver splits a version such as "v3.2.0-rc.32+build.7" into its numeric
// core and its prerelease identifiers. Build metadata is ignored, as semantic
// versioning requires for precedence. It reports false for anything that is not
// MAJOR.MINOR.PATCH, such as "dev" or an empty string.
func parseSemver(v string) (core [3]int, pre []string, ok bool) {
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		pre = strings.Split(v[i+1:], ".")
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return core, nil, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return core, nil, false
		}
		core[i] = n
	}
	return core, pre, true
}

// compareSemver orders two parsed versions by semantic-versioning precedence:
// the numeric core first, then a release outranks its prereleases, then the
// prerelease identifiers left to right.
func compareSemver(aCore [3]int, aPre []string, bCore [3]int, bPre []string) int {
	for i := range aCore {
		if aCore[i] != bCore[i] {
			if aCore[i] < bCore[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(aPre) == 0 && len(bPre) == 0:
		return 0
	case len(aPre) == 0:
		return 1
	case len(bPre) == 0:
		return -1
	}
	for i := 0; i < len(aPre) && i < len(bPre); i++ {
		if c := comparePreIdent(aPre[i], bPre[i]); c != 0 {
			return c
		}
	}
	switch {
	case len(aPre) < len(bPre):
		return -1
	case len(aPre) > len(bPre):
		return 1
	}
	return 0
}

// comparePreIdent orders one prerelease identifier. Numeric identifiers compare
// numerically and sort below alphanumeric ones; alphanumeric identifiers compare
// lexically.
func comparePreIdent(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	switch {
	case aErr == nil && bErr == nil:
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		}
		return 0
	case aErr == nil:
		return -1
	case bErr == nil:
		return 1
	}
	return strings.Compare(a, b)
}

// versionSkewWarning reports whether the project's recorded template version is
// newer than the installed binary. A sync in that state would deploy older rules
// over newer ones, so the caller warns and does not block. Versions that do not
// parse as MAJOR.MINOR.PATCH (a dev build, an empty or unset record) never warn.
func versionSkewWarning(binaryVersion, projectVersion string) (string, bool) {
	bCore, bPre, bOK := parseSemver(binaryVersion)
	pCore, pPre, pOK := parseSemver(projectVersion)
	if !bOK || !pOK || compareSemver(pCore, pPre, bCore, bPre) <= 0 {
		return "", false
	}
	return fmt.Sprintf("project rules were deployed by moai %s, but this binary is %s: this sync would downgrade them; update the binary first (moai update --binary)", projectVersion, binaryVersion), true
}

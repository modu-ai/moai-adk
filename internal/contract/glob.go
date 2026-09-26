package contract

import (
	"path"
	"strings"
)

// MatchGlob reports whether the forward-slash, repo-relative path name
// matches pattern under design.md § Glob Semantics: `*` matches within one
// path segment, `**` (as a whole segment) matches zero or more whole
// segments, and `?` matches one non-separator character. Each non-`**`
// segment is matched with the standard library's path.Match, so no
// doublestar dependency is needed. A malformed pattern matches nothing.
//
// @MX:ANCHOR: [AUTO] The one glob rule for contract ownership sets.
// @MX:REASON: Verify uses it for the contract.yaml coverage rule, and the
// downstream ownership detectors (A2) reuse it for write/never/scratch and
// the frozen-files set; two matchers would let the verifier and the detector
// disagree about which paths a signed contract covers.
func MatchGlob(pattern, name string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			rest := pat[1:]
			for len(rest) > 0 && rest[0] == "**" {
				rest = rest[1:]
			}
			for i := 0; i <= len(name); i++ {
				if matchSegments(rest, name[i:]) {
					return true
				}
			}
			return false
		}
		if len(name) == 0 {
			return false
		}
		ok, err := path.Match(pat[0], name[0])
		if err != nil || !ok {
			return false
		}
		pat, name = pat[1:], name[1:]
	}
	return len(name) == 0
}

// relativeClean reports whether p is a non-empty relative path with no `..`
// segment: no leading `/` or `\` (which also rejects `//server` and
// `\\server` UNC forms), no Windows drive prefix, and no `..` segment under
// either separator.
func relativeClean(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) {
		return false
	}
	if len(p) >= 2 && p[1] == ':' && isASCIILetter(p[0]) {
		return false
	}
	for _, seg := range strings.FieldsFunc(p, func(r rune) bool { return r == '/' || r == '\\' }) {
		if seg == ".." {
			return false
		}
	}
	return true
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// validGlob reports whether g is a well-formed ownership glob: relative, no
// `..` segment, and every segment a syntactically valid path.Match pattern
// (a malformed pattern could never match, silently disarming the glob).
func validGlob(g string) bool {
	if !relativeClean(g) {
		return false
	}
	for _, seg := range strings.Split(g, "/") {
		if seg == "**" {
			continue
		}
		if _, err := path.Match(seg, ""); err != nil {
			return false
		}
	}
	return true
}

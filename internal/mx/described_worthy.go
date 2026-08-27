package mx

import "strings"

// IsDescribedWorthy reports whether a repo-relative slash path names a file a
// curated architecture document could describe: Go source, excluding tests and
// anything living under a testdata directory (REQ-GFC-002).
//
// Pure by construction — no configuration, no filesystem access — so the two
// call sites that feed it (git output and a directory walk) stay honest about
// supplying comparable path forms. "testdata" is matched as a whole path
// segment, never as a substring: internal/foo/testdatax/a.go is admitted.
func IsDescribedWorthy(relPath string) bool {
	if !strings.HasSuffix(relPath, ".go") {
		return false
	}
	if strings.HasSuffix(relPath, "_test.go") {
		return false
	}
	for _, seg := range strings.Split(relPath, "/") {
		if seg == "testdata" {
			return false
		}
	}
	return true
}

//go:build !linux && !darwin

// ccversion_other.go — the running-version probe is unsupported on platforms
// without a reader (SPEC-SESSION-CC-VERSION-001 C.2: darwin reads lsof,
// linux reads /proc/<pid>/exe, every other platform degrades). Reporting
// unsupported rather than guessing keeps the view on its documented
// degradation: the value renders unknown, never an inferred one
// (REQ-SCV-003).
package session

// platformReadProcessMapping always reports unsupported.
func platformReadProcessMapping(pid int) (string, bool) {
	_ = pid
	return "", false
}

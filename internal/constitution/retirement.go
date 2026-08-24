package constitution

import "strings"

// retirementMarkerPrefix is the marker that retires a registry entry.
//
// A clause beginning with this prefix records a clause that used to be in force
// and has since been withdrawn. The entry stays in the registry as an audit
// record ("this clause was retired") instead of being deleted, and validate
// stops checking it against the source tree — the source text is gone by
// definition, so a verbatim check can only ever fail.
//
// The marker is deliberately a PREFIX rather than a substring: a live clause may
// legitimately mention the marker in its own text (the session-handoff clause
// instructing that superseded memory entries be marked with it), and treating
// that as retirement would silently disable a real check.
//
// canary_gate:false is NOT a retirement marker. It is the documented default for
// every Evolvable entry, so honoring it as one would disable the drift check for
// most of the registry. Retired entries carry the prefix; the canary invariant is
// relaxed for them as a consequence of retirement, not as a second marker.
const retirementMarkerPrefix = "[SUPERSEDED"

// IsRetiredClause reports whether a registry clause carries the retirement
// marker — a leading "[SUPERSEDED" (optionally followed by a reason, e.g.
// "[SUPERSEDED by CONST-V3R6-001]"). Leading whitespace is ignored.
//
// The marker must be a COMPLETE token: the prefix has to end at the bracket or
// at a space, and the bracket has to close. A bare prefix match would classify
// "[SUPERSEDEDLY] …" and an unterminated "[SUPERSEDED …" as retired, and a
// retired entry skips drift, canary-gate, and source-file validation — so a
// live or malformed clause would silently switch its own checks off.
func IsRetiredClause(clause string) bool {
	trimmed := strings.TrimSpace(clause)
	if !strings.HasPrefix(trimmed, retirementMarkerPrefix) {
		return false
	}
	rest := trimmed[len(retirementMarkerPrefix):]
	if rest == "" || (rest[0] != ']' && rest[0] != ' ') {
		return false
	}
	return strings.Contains(rest, "]")
}

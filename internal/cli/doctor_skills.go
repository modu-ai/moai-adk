// SPEC-V3R3-HARNESS-001 / T-M3-01
// Skills allowlist check for the doctor diagnostic.
//
// SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 (#1088): the "known moai-* core skill" set
// is DERIVED from the embedded template manifest, not a hand-maintained static
// slice. The former staticCoreAllowlist drifted from the templates (27 embedded
// skills vs 22 listed → 10 false "unknown" warnings + 5 stale entries); a
// derived set is drift-free by construction.

package cli

import (
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/template"
)

// @MX:NOTE: [AUTO] SSOT for the known-skill set is the embedded template FS
// (template.EmbeddedMoaiSkillNames), the exact bytes `moai update` installs.
// Do NOT reintroduce a hand-maintained static slice — it will silently drift
// from the manifest and reproduce the #1088 false "unknown" warnings.
var (
	coreAllowlistOnce sync.Once
	coreAllowlist     map[string]bool
)

// knownCoreSkills returns the set of known MoAI-ADK core skill names derived
// from the embedded template manifest. Computed once and cached. A manifest
// read failure yields an EMPTY set; callers MUST treat empty as "manifest
// unavailable" and degrade gracefully (no spurious WARN) rather than flagging
// every moai-* skill as unknown.
func knownCoreSkills() map[string]bool {
	coreAllowlistOnce.Do(func() {
		coreAllowlist = make(map[string]bool)
		names, err := template.EmbeddedMoaiSkillNames()
		if err != nil {
			return // empty set → graceful degradation (REQ-DFS-005)
		}
		for _, n := range names {
			coreAllowlist[n] = true
		}
	})
	return coreAllowlist
}

// @MX:ANCHOR: [AUTO] known-set-from-manifest invariant — the Skills Allowlist
// classification contract (closes #1088, REQ-DFS-005/006). classifySkill is the
// false-warning locus; deriving "known" from the embedded manifest makes a
// template-fresh project report zero unknowns by construction.
// @MX:REASON: [AUTO] the retired staticCoreAllowlist was a hand-maintained slice
// that `moai update` could not touch, so it drifted from the templates and
// reported 10 just-installed skills as strangers. The "known" set MUST come from
// the same authoritative embedded source the binary installs.
//
// classifySkill returns a classification string for a single skill directory name.
//
// Classification rules (REQ-HARNESS-003, REQ-DFS-005):
//   - Name in the manifest-derived known set → "PASS"
//   - Name has "moai-" prefix, NOT known      → "WARN" (unknown moai- skill)
//     (graceful fallback: when the derived set is empty — manifest unavailable —
//     a moai-* name is NOT warned, to avoid spurious warnings)
//   - Name has "hns-" / "harness-" / "my-harness-" prefix → "INFO" (user customization)
//   - Anything else                           → "INFO" (non-moai skill, no enforcement)
func classifySkill(name string) string {
	known := knownCoreSkills()
	if known[name] {
		return "PASS"
	}

	// moai- prefix not in the known set → unknown / warn.
	if strings.HasPrefix(name, "moai-") {
		// Graceful fallback: an empty derived set means the manifest could not
		// be read; do NOT spuriously WARN every moai-* skill.
		if len(known) == 0 {
			return "INFO"
		}
		return "WARN"
	}

	// User customization area.
	// SPEC-HNS-PREFIX-RENAME-001: recognize hns-* (canonical) plus harness-*
	// and my-harness-* (legacy generations, REQ-HNS-005 backward-compat).
	if strings.HasPrefix(name, "hns-") || strings.HasPrefix(name, "harness-") || strings.HasPrefix(name, "my-harness-") {
		return "INFO"
	}

	// Everything else (third-party, moai unified dir, etc.) → INFO.
	return "INFO"
}

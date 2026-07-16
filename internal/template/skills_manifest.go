// SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 / Defect B
// Embedded skills manifest reader — the authoritative SSOT for the doctor's
// Skills Allowlist "known moai-* skill" set (REQ-DFS-005).

package template

import (
	"io/fs"
	"strings"
)

// moaiSkillPrefix is the directory-name prefix identifying a MoAI-ADK core
// skill. The trailing dash is significant: it excludes the bare `moai` unified
// skill directory (no trailing dash) from the core-skill set.
const moaiSkillPrefix = "moai-"

// @MX:NOTE: [AUTO] SSOT source for the doctor Skills Allowlist known-set.
// Deriving the set from the embedded FS (the exact bytes `moai update` installs)
// makes the allowlist drift-free by construction — do NOT reintroduce a
// hand-maintained static slice (closes #1088, REQ-DFS-005).
//
// EmbeddedMoaiSkillNames returns the sorted-by-FS set of `moai-*` skill
// directory names present under the embedded templates' .claude/skills/. It is
// the authoritative manifest of core skills the running binary installs, so a
// template-fresh project reports zero unknown skills by construction.
//
// A read failure (unexpected in a compiled binary) returns the error and an
// empty slice; callers MUST treat an empty derived set as "manifest
// unavailable" and degrade gracefully rather than classifying every moai-*
// skill as unknown.
func EmbeddedMoaiSkillNames() ([]string, error) {
	sub, err := EmbeddedTemplates()
	if err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(sub, ".claude/skills")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), moaiSkillPrefix) {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

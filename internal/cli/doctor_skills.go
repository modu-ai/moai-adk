// SPEC-V3R3-HARNESS-001 / T-M3-01
// Skills allowlist check for the doctor diagnostic.

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// staticCoreAllowlist is the canonical set of 23 MoAI-ADK core skill names.
// Skills in this list are maintained by moai update and are expected to be present.
// POST-D-1 FIX: moai-workflow-research and moai-workflow-pencil-integration removed.
var staticCoreAllowlist = []string{
	// foundation (4)
	"moai-foundation-cc", "moai-foundation-core", "moai-foundation-quality", "moai-foundation-thinking",
	// workflow (10)
	"moai-workflow-ddd", "moai-workflow-design-context", "moai-workflow-design-import", "moai-workflow-gan-loop",
	"moai-workflow-loop", "moai-workflow-project", "moai-workflow-spec", "moai-workflow-tdd",
	"moai-workflow-testing", "moai-workflow-worktree",
	// ref (5)
	"moai-ref-api-patterns", "moai-ref-git-workflow", "moai-ref-owasp-checklist",
	"moai-ref-react-patterns", "moai-ref-testing-pyramid",
	// design (1)
	"moai-design-system",
	// FROZEN domain (2)
	"moai-domain-brand-design", "moai-domain-copywriting",
	// meta (1)
	"moai-meta-harness",
}

// classifySkill returns a classification string for a single skill directory name.
//
// Classification rules (REQ-HARNESS-003):
//   - Name in staticCoreAllowlist   → "PASS"
//   - Name has "moai-" prefix, NOT in allowlist → "WARN" (unknown moai- skill)
//   - Name has "my-harness-" prefix → "INFO" (user customization detected)
//   - Anything else                 → "INFO" (non-moai skill, no enforcement)
func classifySkill(name string) string {
	// Check allowlist first
	for _, allowed := range staticCoreAllowlist {
		if name == allowed {
			return "PASS"
		}
	}

	// moai- prefix not in allowlist → unknown / warn
	if strings.HasPrefix(name, "moai-") {
		return "WARN"
	}

	// User customization area
	if strings.HasPrefix(name, "my-harness-") {
		return "INFO"
	}

	// Everything else (third-party, moai unified dir, etc.) → INFO
	return "INFO"
}

// runSkillsCheck inspects .claude/skills/ in projectRoot and classifies each
// entry by its directory name. Returns a DiagnosticCheck summarising results.
//
// Status mapping:
//   - Any WARN classification → CheckWarn for the aggregate check
//   - Otherwise              → CheckOK with counts detail line
func runSkillsCheck(projectRoot string) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Skills Allowlist"}

	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			check.Status = CheckOK
			check.Message = ".claude/skills/ not found (run 'moai init')"
			return check
		}
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("cannot read .claude/skills/: %v", err)
		return check
	}

	var passCount, warnCount, infoCount int
	var warnNames []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		class := classifySkill(entry.Name())
		switch class {
		case "PASS":
			passCount++
		case "WARN":
			warnCount++
			warnNames = append(warnNames, entry.Name())
		default:
			infoCount++
		}
	}

	if warnCount > 0 {
		check.Status = CheckWarn
		check.Message = fmt.Sprintf("%d unknown moai- skill(s) outside static core: %s",
			warnCount, strings.Join(warnNames, ", "))
		check.Detail = "Unknown moai- skills may be stale. Run 'moai update' to sync or remove manually."
		return check
	}

	check.Status = CheckOK
	check.Message = fmt.Sprintf("skills OK — %d core, %d user/info", passCount, infoCount)
	return check
}

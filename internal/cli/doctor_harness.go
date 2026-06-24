// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-03
// Doctor 5-Layer harness diagnosis.

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness"
)

// runHarnessCheck performs the 5-Layer harness diagnosis (REQ-PH-009 indirect).
// Returns a single aggregated DiagnosticCheck whose Detail summarises the per-
// layer status (L1 ~ L5) and any prefix conflicts.
func runHarnessCheck(projectRoot string) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Harness 5-Layer"}

	harnessDir := filepath.Join(projectRoot, ".moai", "harness")
	if _, err := os.Stat(harnessDir); os.IsNotExist(err) {
		check.Status = CheckOK
		check.Message = ".moai/harness/ not present (no harness configured)"
		return check
	}

	var statuses []string
	var failures []string

	// L1: harness-* skills have triggers section.
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	l1, l1Detail := checkLayer1Triggers(skillsDir)
	statuses = append(statuses, "L1:"+l1)
	if l1 == "FAIL" {
		failures = append(failures, "L1 "+l1Detail)
	}

	// L2: workflow.yaml has harness section.
	wfYAML := filepath.Join(projectRoot, ".moai", "config", "sections", "workflow.yaml")
	l2, l2Detail := checkLayer2Workflow(wfYAML)
	statuses = append(statuses, "L2:"+l2)
	if l2 == "FAIL" {
		failures = append(failures, "L2 "+l2Detail)
	}

	// L3: CLAUDE.md marker block paired.
	l3, l3Detail := checkLayer3Marker(filepath.Join(projectRoot, "CLAUDE.md"))
	statuses = append(statuses, "L3:"+l3)
	if l3 == "FAIL" {
		failures = append(failures, "L3 "+l3Detail)
	}

	// L4: 4 workflow files contain @.moai/harness/*-extension.md import.
	l4, l4Detail := checkLayer4ImportLines(filepath.Join(projectRoot, ".claude", "skills", "moai", "workflows"))
	statuses = append(statuses, "L4:"+l4)
	if l4 == "FAIL" {
		failures = append(failures, "L4 "+l4Detail)
	}

	// L5: .moai/harness/ baseline files exist.
	l5, l5Detail := checkLayer5Files(harnessDir)
	statuses = append(statuses, "L5:"+l5)
	if l5 == "FAIL" {
		failures = append(failures, "L5 "+l5Detail)
	}

	// L6 (SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001 smoke gate): generated-agent
	// frontmatter self-activation checks (REQ-HAW-012/013/013b). Iterates
	// .claude/agents/harness/*.md and FAILs on an empty description, a missing
	// skills: key, or a dangling harness-* skills: reference. No-op when no
	// generated agents exist. Reuses skillsDir (resolved above) for reference
	// resolution. Preserves L1-L5 semantics (AC-HAW-014) — additive layer only.
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "harness")
	l6, l6Detail := checkLayer6AgentActivation(agentsDir, skillsDir)
	statuses = append(statuses, "L6:"+l6)
	if l6 == "FAIL" {
		failures = append(failures, "L6 "+l6Detail)
	}

	// Prefix conflicts (warn, not fail)
	var warnSuffix string
	conflicts, _ := harness.DetectPrefixConflicts(skillsDir)
	if len(conflicts) > 0 {
		var lines []string
		for _, c := range conflicts {
			lines = append(lines, fmt.Sprintf("WARN: %s conflicts with %s (%s)",
				c.MyHarnessSkill, c.MoaiSkill, c.Reason))
		}
		warnSuffix = "\n  " + strings.Join(lines, "\n  ")
	}

	switch {
	case len(failures) > 0:
		check.Status = CheckFail
		check.Message = strings.Join(statuses, " ")
		check.Detail = strings.Join(failures, "; ") + warnSuffix
	case warnSuffix != "":
		check.Status = CheckWarn
		check.Message = strings.Join(statuses, " ") + " (with prefix conflicts)"
		check.Detail = warnSuffix
	default:
		check.Status = CheckOK
		check.Message = strings.Join(statuses, " ")
	}
	return check
}

// checkLayer1Triggers verifies that every harness-*/SKILL.md has the
// required triggers frontmatter section.
func checkLayer1Triggers(skillsDir string) (string, string) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "PASS", "no skills dir"
		}
		return "FAIL", err.Error()
	}
	var problems []string
	for _, e := range entries {
		// SPEC-V3R6-HARNESS-NAMESPACE-V2-001: recognize both harness-* (canonical)
		// and my-harness-* (legacy, REQ-HNS-005 backward-compat).
		if !e.IsDir() || (!strings.HasPrefix(e.Name(), "harness-") && !strings.HasPrefix(e.Name(), "my-harness-")) {
			continue
		}
		skillPath := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err != nil {
			problems = append(problems, e.Name()+": SKILL.md missing")
			continue
		}
		if err := harness.VerifyTriggers(skillPath); err != nil {
			problems = append(problems, e.Name()+": "+err.Error())
		}
	}
	if len(problems) > 0 {
		return "FAIL", strings.Join(problems, "; ")
	}
	return "PASS", "ok"
}

// checkLayer2Workflow verifies workflow.yaml contains a harness section.
func checkLayer2Workflow(yamlPath string) (string, string) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "FAIL", "workflow.yaml missing"
		}
		return "FAIL", err.Error()
	}
	// Light text check (avoids re-parsing cost): look for `harness:` under workflow.
	if !regexp.MustCompile(`(?m)^\s*harness:\s*$`).Match(data) {
		return "FAIL", "harness: section not found"
	}
	return "PASS", "ok"
}

// checkLayer3Marker verifies CLAUDE.md has a paired marker block (1 start + 1 end).
func checkLayer3Marker(claudeMdPath string) (string, string) {
	data, err := os.ReadFile(claudeMdPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "FAIL", "CLAUDE.md missing"
		}
		return "FAIL", err.Error()
	}
	startCount := strings.Count(string(data), "<!-- moai:harness-start")
	endCount := strings.Count(string(data), "<!-- moai:harness-end -->")
	if startCount != 1 || endCount != 1 {
		return "FAIL", fmt.Sprintf("marker not paired (%d start / %d end)", startCount, endCount)
	}
	return "PASS", "ok"
}

// checkLayer4ImportLines verifies plan/run/sync/design.md each contain
// the static @.moai/harness/*-extension.md import line.
func checkLayer4ImportLines(workflowsDir string) (string, string) {
	required := []string{"plan.md", "run.md", "sync.md", "design.md"}
	var missing []string
	for _, f := range required {
		path := filepath.Join(workflowsDir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			missing = append(missing, f+" (read error)")
			continue
		}
		if !strings.Contains(string(data), "@.moai/harness/") {
			missing = append(missing, f+" (no @.moai/harness/ import)")
		}
	}
	if len(missing) > 0 {
		return "FAIL", strings.Join(missing, "; ")
	}
	return "PASS", "ok"
}

// checkLayer5Files verifies the 7 baseline files exist in .moai/harness/.
func checkLayer5Files(harnessDir string) (string, string) {
	required := []string{
		"main.md",
		"plan-extension.md",
		"run-extension.md",
		"sync-extension.md",
		"chaining-rules.yaml",
		"interview-results.md",
		"README.md",
	}
	var missing []string
	for _, f := range required {
		if _, err := os.Stat(filepath.Join(harnessDir, f)); err != nil {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return "FAIL", "missing: " + strings.Join(missing, ", ")
	}
	return "PASS", "ok"
}

// checkLayer6AgentActivation is the SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
// Phase-6 smoke gate (REQ-HAW-012/013/013b). For each generated
// .claude/agents/harness/*.md agent it asserts the self-activation frontmatter
// contract:
//
//   - REQ-HAW-012: the `description:` field is non-empty.
//   - REQ-HAW-013b: the `skills:` frontmatter key is present (a `skills:`-less
//     agent would otherwise pass silently and reproduce the auto-discovery
//     failure mode the SPEC exists to close).
//   - REQ-HAW-013: each `skills:` entry that names a `harness-*` skill
//     resolves to an existing .claude/skills/<name>/ directory (dangling
//     references FAIL). Template-distributed `moai-*` references are NOT
//     resolved against disk and are never dangling (EC-4).
//
// No-op (PASS) when the .claude/agents/harness/ directory is absent or contains
// no *.md agents — the contract applies only to generated agents.
func checkLayer6AgentActivation(agentsDir, skillsDir string) (string, string) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "PASS", "no generated agents"
		}
		return "FAIL", err.Error()
	}

	var problems []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		agentName := strings.TrimSuffix(e.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(agentsDir, e.Name()))
		if err != nil {
			problems = append(problems, agentName+": read error: "+err.Error())
			continue
		}
		fm := extractFrontmatter(string(data))

		// REQ-HAW-012: non-empty description.
		if strings.TrimSpace(frontmatterScalar(fm, "description")) == "" {
			problems = append(problems, agentName+": empty description frontmatter field")
		}

		// REQ-HAW-013b: skills: key present at all.
		skillRefs, hasSkillsKey := frontmatterList(fm, "skills")
		if !hasSkillsKey {
			problems = append(problems, agentName+": missing skills: frontmatter key (no companion skill preload)")
			continue
		}

		// REQ-HAW-013: harness-* references must resolve on disk (EC-4:
		// moai-* template skills are not resolved here).
		// SPEC-V3R6-HARNESS-NAMESPACE-V2-001: recognize both harness-* (canonical)
		// and my-harness-* (legacy, REQ-HNS-005 backward-compat).
		for _, ref := range skillRefs {
			if !strings.HasPrefix(ref, "harness-") && !strings.HasPrefix(ref, "my-harness-") {
				continue
			}
			if _, err := os.Stat(filepath.Join(skillsDir, ref)); err != nil {
				problems = append(problems, agentName+": dangling skills: reference "+ref+" (skill dir absent)")
			}
		}
	}

	if len(problems) > 0 {
		return "FAIL", strings.Join(problems, "; ")
	}
	return "PASS", "ok"
}

// extractFrontmatter returns the YAML frontmatter block (between the leading
// `---` fences) of a markdown agent file, or "" when no frontmatter is present.
func extractFrontmatter(content string) string {
	trimmed := strings.TrimLeft(content, " \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return ""
	}
	rest := trimmed[len("---"):]
	// Find the closing fence at the start of a line.
	if idx := strings.Index(rest, "\n---"); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// frontmatterScalar returns the trimmed scalar value of a top-level `key:` line
// in the frontmatter block (empty string when absent or value-less).
func frontmatterScalar(fm, key string) string {
	for _, line := range strings.Split(fm, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if v, ok := strings.CutPrefix(trimmed, key+":"); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// frontmatterList parses a top-level `key:` list from the frontmatter block.
// It supports both the block-list form (`skills:` then `  - item` lines) and
// the inline-flow form (`skills: [a, b]`). The second return value reports
// whether the key was present at all (distinguishing an absent key from an
// empty list — REQ-HAW-013b needs to FAIL on absence specifically).
func frontmatterList(fm, key string) ([]string, bool) {
	lines := strings.Split(fm, "\n")
	for i, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		v, ok := strings.CutPrefix(line, key+":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		// Inline-flow form: skills: [a, b]
		if strings.HasPrefix(v, "[") {
			inner := strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
			var out []string
			for _, item := range strings.Split(inner, ",") {
				item = strings.TrimSpace(strings.Trim(item, "\"'"))
				if item != "" {
					out = append(out, item)
				}
			}
			return out, true
		}
		// Block-list form: subsequent `  - item` lines until dedent.
		var out []string
		for _, sub := range lines[i+1:] {
			s := strings.TrimRight(sub, "\r")
			trimmed := strings.TrimSpace(s)
			if trimmed == "" {
				continue
			}
			// A non-indented, non-list line ends the block.
			if !strings.HasPrefix(s, " ") && !strings.HasPrefix(s, "\t") {
				break
			}
			if item, isItem := strings.CutPrefix(trimmed, "- "); isItem {
				out = append(out, strings.TrimSpace(strings.Trim(item, "\"'")))
			} else {
				// Indented but not a list item → end of this key's block.
				break
			}
		}
		return out, true
	}
	return nil, false
}

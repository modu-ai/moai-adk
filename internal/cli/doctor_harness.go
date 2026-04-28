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

	// L1: my-harness-* skills have triggers section.
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

// checkLayer1Triggers verifies that every my-harness-*/SKILL.md has the
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
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "my-harness-") {
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

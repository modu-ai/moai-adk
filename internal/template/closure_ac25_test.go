package template

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// closureMarkerBlock extracts the lines between the literal
// moai:closure-second-review:start/end markers of one file.
func closureMarkerBlock(t *testing.T, path string) (string, int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(data)
	const start = "<!-- moai:closure-second-review:start -->"
	const end = "<!-- moai:closure-second-review:end -->"
	if got := strings.Count(text, start); got != 1 {
		t.Fatalf("%s carries %d start markers, want exactly 1", path, got)
	}
	if got := strings.Count(text, end); got != 1 {
		t.Fatalf("%s carries %d end markers, want exactly 1", path, got)
	}
	s := strings.Index(text, start)
	e := strings.Index(text, end)
	if e < s {
		t.Fatalf("%s: end marker before start marker", path)
	}
	return text[s+len(start) : e], s
}

// TestAC_CLOSURE_025 — template neutrality of the auditor instructions.
func TestAC_CLOSURE_025(t *testing.T) {
	copies := []string{
		"templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md",
		"templates/.claude/agents/moai/sync-auditor.md",
	}
	forbidden := regexp.MustCompile(`SPEC-[A-Z]|\bt[0-9]{3,5}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}|A-Q[0-9]|\b[0-9a-f]{9,40}\b`)
	for _, rel := range copies {
		t.Run(rel, func(t *testing.T) {
			block, _ := closureMarkerBlock(t, rel)
			if forbidden.MatchString(block) {
				t.Fatalf("marker block of %s carries internal content (SPEC id, card id, date, decision id, or SHA)", rel)
			}
			for _, want := range []string{"card_id", "baseBranch", "governed paths"} {
				if !strings.Contains(block, want) {
					t.Fatalf("marker block of %s does not instruct %q", rel, want)
				}
			}
		})
	}

	// The emitted Codex copy of the auditor mirrors the block (make
	// agents-emit-check is the shell check; this asserts the committed
	// artifact directly).
	data, err := os.ReadFile("templates/.codex/agents/moai/sync-auditor.toml")
	if err != nil {
		t.Fatalf("read codex copy: %v", err)
	}
	if !strings.Contains(string(data), "closure-second-review") {
		t.Fatalf("codex copy of sync-auditor does not carry the second-review block")
	}
}

// TestAC_CLOSURE_025_AgentsEmitCheck shells out to the AC's shell check on
// Unix hosts with make available; skipped elsewhere (CI runs the go-level
// golden test in every environment).
func TestAC_CLOSURE_025_AgentsEmitCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("make agents-emit-check requires a POSIX host")
	}
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not available")
	}
	cmd := exec.Command("make", "agents-emit-check")
	cmd.Dir = "../.." // the Makefile lives at the repository root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make agents-emit-check failed: %v\n%s", err, out)
	}
}

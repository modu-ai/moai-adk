package homestate

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

// Card t2007 (PLN-01) guard: the audit-ready block the shipped plan workflow
// tells a plan to write must satisfy the kickoff predicate. Before this card
// the docs documented only `plan_complete_at` + `plan_status: audit-ready`,
// while auditReadyRecorded accepts only a bare `audit_ready: true` line, so a
// plan completed exactly as documented failed the Kickoff readiness input.

// fencedLines returns the non-empty lines inside the first ``` fence of text.
func fencedLines(text string) []string {
	_, after, ok := strings.Cut(text, "```")
	if !ok {
		return nil
	}
	_, after, _ = strings.Cut(after, "\n")
	body, _, _ := strings.Cut(after, "```")
	var out []string
	for _, l := range strings.Split(body, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func writeE1(t *testing.T, lines []string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "progress.md")
	body := "# Progress\n\n## §E.1 Plan-phase Audit-Ready Signal\n\n" + strings.Join(lines, "\n") + "\n\n## §E.2 Run-phase Evidence\n"
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDocumentedAuditReadyBlockSatisfiesKickoff(t *testing.T) {
	fsys, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	read := func(p string) string {
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		return string(raw)
	}

	_, sig, ok := strings.Cut(read(".claude/skills/moai/workflows/plan.md"), "\n## Audit-Ready Signal\n")
	if !ok {
		t.Fatal("plan.md: '## Audit-Ready Signal' not found")
	}
	sig, _, _ = strings.Cut(sig, "\n## ")

	_, crit, ok := strings.Cut(read(".claude/skills/moai/workflows/plan/spec-assembly.md"), "**Audit-ready signal**")
	if !ok {
		t.Fatal("spec-assembly.md: '**Audit-ready signal**' not found")
	}
	crit, _, _ = strings.Cut(crit, "\n---")

	for name, text := range map[string]string{"plan.md": sig, "spec-assembly.md": crit} {
		lines := fencedLines(text)
		if len(lines) == 0 {
			t.Errorf("%s: documented signal block has no fenced lines", name)
			continue
		}
		if ready, ignored := auditReadyRecorded(writeE1(t, lines)); !ready {
			t.Errorf("%s: documented block %q is not accepted by auditReadyRecorded (ignored look-alike %q)", name, lines, ignored)
		}
	}
}

func TestAuditReadyRecordedNegativeFixtures(t *testing.T) {
	cases := map[string][]string{
		"hyphen key":            {"audit-ready: true"},
		"false value":           {"audit_ready: false"},
		"suffixed key":          {"audit_ready_note: x"},
		"plan_status alone":     {"- plan_complete_at: 2026-11-01T00:00:00Z", "- plan_status: audit-ready"},
		"true then conflicting": {"audit_ready: true", "audit_ready: false"},
	}
	for name, lines := range cases {
		if ready, _ := auditReadyRecorded(writeE1(t, lines)); ready {
			t.Errorf("%s: %q accepted, want not ready", name, lines)
		}
	}
	// Missing §E.1 section entirely.
	p := filepath.Join(t.TempDir(), "progress.md")
	if err := os.WriteFile(p, []byte("# Progress\n\naudit_ready: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if ready, _ := auditReadyRecorded(p); ready {
		t.Error("missing §E.1 section accepted, want not ready")
	}
}

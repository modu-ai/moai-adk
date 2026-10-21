package homestate

import (
	"path/filepath"
	"strings"
	"testing"
)

// t1854 — a markdown bullet before `audit_ready` is the same signal; a
// look-alike key is not, and the refusal names it.
func TestAuditReadyBulletForms(t *testing.T) {
	sec := func(body string) string {
		return "# progress\n\n## §E.1 Plan-phase Audit-Ready Signal\n\n" + body + "\n## §E.2 Run-phase Evidence\n\n_pending_\n"
	}
	cases := []struct {
		name     string
		body     string
		accepted bool
		reason   string // substring the refusal must carry; "" = none required
	}{
		{"bare line", "audit_ready: true\n", true, ""},
		{"dash bullet", "- audit_ready: true\n", true, ""},
		{"star bullet", "* audit_ready: true\n", true, ""},
		{"bullet inside fenced yaml", "```yaml\n- audit_ready: true\n```\n", true, ""},
		{"bullet conflicting with bare false", "- audit_ready: true\naudit_ready: false\n", false, ""},
		{"bullet false", "- audit_ready: false\n", false, ""},
		{"look-alike suffix key", "- audit_ready_note: true\n", false, `ignored look-alike key "audit_ready_note"`},
		{"look-alike hyphen key", "audit-ready: true\n", false, `ignored look-alike key "audit-ready"`},
		{"look-alike capitalised key", "Audit_Ready: true\n", false, `ignored look-alike key "Audit_Ready"`},
		{"no signal at all", "plan done\n", false, "no audit-ready status"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFDAKickoff(t, "b")
			frWrite(t, filepath.Join(f.repo.Dir, ".moai", "specs", frSpecID, "progress.md"), sec(tc.body))
			// Re-stamp the verdict so its plan_artifact_hash covers the rewritten progress.md.
			frWriteVerdict(t, f.repo.Dir, f.card.CardID, "plan-audit.md", "PASS", f.repo.Commit)
			_, err := f.approve(DeciderAudit, QueueHoldClear)
			if tc.accepted {
				if err != nil {
					t.Fatalf("want acceptance, got refusal: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want refusal, got acceptance")
			}
			if tc.reason != "" && !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("refusal %q does not name %q", err.Error(), tc.reason)
			}
		})
	}
}

package hook

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// clearKanbanEnv unsets every kanban variable so each case starts from a
// known-absent state. t.Setenv registers the restore, so the process env is
// returned to its prior value when the test ends.
func clearKanbanEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanSettingsInjected,
		config.EnvMoaiKanbanLeadAddr,
	} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestKanbanBootstrapNoticeSilentForOrdinarySession is the case that matters
// most for blast radius: a session that is not part of a kanban run must be
// completely unaffected.
func TestKanbanBootstrapNoticeSilentForOrdinarySession(t *testing.T) {
	clearKanbanEnv(t)

	if got := kanbanBootstrapNotice(langEnglish); got != "" {
		t.Errorf("non-kanban session got a notice: %q", got)
	}
}

// TestKanbanBootstrapNoticeLead asserts the lead branch names the run and
// carries one launch command per companion role, each now carrying -k
// (AC-FB-015: the companion launch lines are kanban-membership commands, not
// the bare --name form the prior-art notice printed).
func TestKanbanBootstrapNoticeLead(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjlgt1")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-tjlgt1")

	got := kanbanBootstrapNotice(langEnglish)
	if got == "" {
		t.Fatal("lead session got no notice")
	}
	if !strings.Contains(got, "tjlgt1") {
		t.Errorf("notice omits the run id: %q", got)
	}
	for _, role := range kanban.CompanionRoles {
		want := "moai cc -k --name " + role + "-tjlgt1"
		if !strings.Contains(got, want) {
			t.Errorf("notice omits %q:\n%s", want, got)
		}
	}
	// Manual bootstrap is stated, not left to be discovered.
	if !strings.Contains(got, "cannot launch another session") {
		t.Errorf("notice does not explain why bootstrap is manual:\n%s", got)
	}
	// The GLM substitution is one line, not a presumed per-role backend split.
	if strings.Count(got, "moai glm") != 1 {
		t.Errorf("expected exactly one 'moai glm' line:\n%s", got)
	}
}

// TestKanbanBootstrapNoticeCompanion asserts the companion branch is role-less
// (AC-FB-016): it names the run, does NOT name the role, and never prints the
// launch block.
func TestKanbanBootstrapNoticeCompanion(t *testing.T) {
	for _, role := range kanban.CompanionRoles {
		t.Run(role, func(t *testing.T) {
			clearKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanbanLabel, kanban.CompanionLabel(role, "tjlgt1"))

			got := kanbanBootstrapNotice(langEnglish)
			if !strings.Contains(got, "tjlgt1") {
				t.Errorf("companion notice = %q, want it to name run %q", got, "tjlgt1")
			}
			// AC-FB-016: the prior-art role clause ("as the X companion") and the
			// word "companion" must NOT appear. (A role name like "run" can appear
			// as part of "joined run <id>" — that is the allowed exception.)
			if strings.Contains(got, "as the ") {
				t.Errorf("companion notice carries the prior-art role clause:\n%s", got)
			}
			if strings.Contains(got, "companion") {
				t.Errorf("companion notice contains the word \"companion\" (role-less per AC-FB-016):\n%s", got)
			}
			if strings.Contains(got, "--name") {
				t.Errorf("companion printed the launch block (recursive bootstrap):\n%s", got)
			}
		})
	}
}

// TestKanbanBootstrapNoticeFailsOpen asserts an unresolvable value degrades to
// emitting nothing rather than emitting something wrong or failing the session
// start, matching the surrounding hook code.
func TestKanbanBootstrapNoticeFailsOpen(t *testing.T) {
	t.Run("lead without a run id", func(t *testing.T) {
		clearKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")

		if got := kanbanBootstrapNotice(langEnglish); got != "" {
			t.Errorf("emitted a notice with no run id: %q", got)
		}
	})

	t.Run("unparseable companion label", func(t *testing.T) {
		clearKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanbanLabel, "oauth-migration")

		if got := kanbanBootstrapNotice(langEnglish); got != "" {
			t.Errorf("emitted a notice for a non-companion label: %q", got)
		}
	})
}

// TestKanbanBootstrapNoticeLabelWinsOverKanban pins the branch order: a
// session carrying both variables is treated as a companion, so it can never
// print the launch block.
func TestKanbanBootstrapNoticeLabelWinsOverKanban(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjlgt1")
	t.Setenv(config.EnvMoaiKanbanLabel, "plan-tjlgt1")

	got := kanbanBootstrapNotice(langEnglish)
	if strings.Contains(got, "--name") {
		t.Errorf("a labelled session printed the launch block:\n%s", got)
	}
	if !strings.Contains(got, "tjlgt1") {
		t.Errorf("expected the companion notice to name the run, got: %q", got)
	}
}

// ── M3 ACs ──

// TestKanbanLeadNoticeFullContent is AC-FB-013: the lead notice carries, in
// order: (a) run id; (b) four companion lines each matching
// `moai cc -k --name (plan|run|review|sync)-<run-id>`; (c) a non-empty leader
// socket path line; (d) an inbound-automation notice; (e) the SPEC identifier.
func TestKanbanLeadNoticeFullContent(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-FOO-001")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-abc123")
	t.Setenv(config.EnvMoaiKanbanSettingsInjected, "1")

	got := kanbanLeadNotice("abc123", langEnglish)

	// (a) run id
	if !strings.Contains(got, "abc123") {
		t.Errorf("notice omits run id abc123:\n%s", got)
	}

	// (b) four companion lines carrying -k
	lineRe := regexp.MustCompile(`(?m)^moai cc -k --name (plan|run|review|sync)-abc123$`)
	matches := lineRe.FindAllString(got, -1)
	if len(matches) != 4 {
		t.Errorf("expected 4 companion lines matching the regex, got %d:\n%s", len(matches), got)
	}

	// (c) leader socket path — a non-empty path-shaped line
	if !strings.Contains(got, "/tmp/moai-kanban-abc123") {
		t.Errorf("notice omits the leader socket path:\n%s", got)
	}

	// (d) inbound-automation notice
	if !strings.Contains(strings.ToLower(got), "auto-accept") {
		t.Errorf("notice lacks the inbound-automation line:\n%s", got)
	}

	// (e) SPEC identifier
	if !strings.Contains(got, "SPEC-FOO-001") {
		t.Errorf("notice omits the SPEC identifier:\n%s", got)
	}

	// Ordering: run id must appear before the companion lines, which must
	// appear before the SPEC identifier.
	runIDIdx := strings.Index(got, "abc123")
	specIdx := strings.Index(got, "SPEC-FOO-001")
	if runIDIdx < 0 || specIdx < 0 || runIDIdx >= specIdx {
		t.Errorf("ordering: run id must precede SPEC id (runIDIdx=%d, specIdx=%d):\n%s", runIDIdx, specIdx, got)
	}
}

// TestKanbanLeadNoticeOmitsSPECWhenUnset is AC-FB-014: the SPEC line is
// omitted entirely when MOAI_KANBAN_SPEC is unset (not printed as an empty
// placeholder).
func TestKanbanLeadNoticeOmitsSPECWhenUnset(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "abc123")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-abc123")

	got := kanbanLeadNotice("abc123", langEnglish)
	// No line should contain a SPEC- prefixed identifier.
	if strings.Contains(got, "SPEC-") {
		t.Errorf("notice contains a SPEC- identifier when MOAI_KANBAN_SPEC is unset:\n%s", got)
	}
}

// TestKanbanLeadNoticeCompanionLinesCarryF is AC-FB-015: each companion line
// matches `^moai (cc|glm) -k --name (plan|run|review|sync)-<run-id>$`.
func TestKanbanLeadNoticeCompanionLinesCarryF(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "xyz789")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-xyz789")

	got := kanbanLeadNotice("xyz789", langEnglish)
	re := regexp.MustCompile(`(?m)^moai (cc|glm) -k --name (plan|run|review|sync)-xyz789$`)
	matches := re.FindAllString(got, -1)
	if len(matches) < 4 {
		t.Errorf("expected ≥4 lines matching the companion-launch regex, got %d:\n%s", len(matches), got)
	}
	// The bare --name form (no -k) must NOT appear.
	bareRe := regexp.MustCompile(`(?m)^moai cc --name (plan|run|review|sync)-xyz789$`)
	if bareRe.FindString(got) != "" {
		t.Errorf("bare --name companion line found (prior-art form; must carry -k):\n%s", got)
	}
}

// TestKanbanCompanionNoticeRoleless is AC-FB-016: the companion notice is
// join-only and role-less — it names the run and does NOT contain the word
// "companion" or the prior-art "as the X companion" clause. (A role name like
// "run" can appear as part of "joined run <id>" — the allowed exception per
// AC-FB-016's "other than as part of the run id" qualifier.)
func TestKanbanCompanionNoticeRoleless(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "run-abc123")

	got := kanbanCompanionNotice("run-abc123", langEnglish)
	if !strings.Contains(got, "abc123") {
		t.Errorf("companion notice does not name the run: %q", got)
	}
	if strings.Contains(got, "companion") {
		t.Errorf("companion notice contains \"companion\": %q", got)
	}
	if strings.Contains(got, "as the ") {
		t.Errorf("companion notice carries the prior-art role clause: %q", got)
	}
}

// TestKanbanCompanionNoticeFailOpen is AC-FB-016a: when MOAI_KANBAN_LABEL is
// unset OR SplitCompanionLabel returns ok=false, the notice is the empty
// string (no notice emitted, no error raised, the launch proceeds).
func TestKanbanCompanionNoticeFailOpen(t *testing.T) {
	t.Run("empty label", func(t *testing.T) {
		clearKanbanEnv(t)
		if got := kanbanCompanionNotice("", langEnglish); got != "" {
			t.Errorf("empty label produced a notice: %q", got)
		}
	})
	t.Run("malformed label (empty run-id portion)", func(t *testing.T) {
		clearKanbanEnv(t)
		if got := kanbanCompanionNotice("run-", langEnglish); got != "" {
			t.Errorf("malformed label produced a notice: %q", got)
		}
	})
	t.Run("non-companion label", func(t *testing.T) {
		clearKanbanEnv(t)
		if got := kanbanCompanionNotice("oauth-migration", langEnglish); got != "" {
			t.Errorf("non-companion label produced a notice: %q", got)
		}
	})
}

// TestKanbanCompanionNoticeJoinOnly is AC-FB-017: the companion notice is a
// single line acknowledging the join (matching "joined run <id>") and does NOT
// print the four-companion launch block.
func TestKanbanCompanionNoticeJoinOnly(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "run-abc123")

	got := kanbanCompanionNotice("run-abc123", langEnglish)
	if !strings.Contains(got, "joined run abc123") {
		t.Errorf("companion notice does not match \"joined run <id>\": %q", got)
	}
	if strings.Contains(got, "--name") {
		t.Errorf("companion notice printed the launch block: %q", got)
	}
	if strings.Count(got, "\n") != 0 {
		t.Errorf("companion notice should be a single line, got:\n%s", got)
	}
}

// TestKanbanLeadNoticeOperatorSettingsAdvisory is AC-FB-012. When the launcher
// did NOT inject its own settings (the operator supplied --settings, or a write
// failure degraded to fail-open), the lead notice prints an advisory instructing
// the operator to verify crossSessionInbound: accept is present.
func TestKanbanLeadNoticeOperatorSettingsAdvisory(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjlgt1")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-tjlgt1")
	// EnvMoaiKanbanSettingsInjected is UNSET → the launcher did not inject.
	t.Setenv(config.EnvMoaiKanbanSettingsInjected, "")
	_ = os.Unsetenv(config.EnvMoaiKanbanSettingsInjected)

	got := kanbanLeadNotice("tjlgt1", langEnglish)
	if got == "" {
		t.Fatal("expected a lead notice, got empty string")
	}
	lowered := strings.ToLower(got)
	if !strings.Contains(lowered, "verify") || !strings.Contains(lowered, "crosssessioninbound") || !strings.Contains(lowered, "accept") {
		t.Errorf("lead notice lacks the verify-crossSessionInbound-accept advisory:\n%s", got)
	}
}

// TestKanbanLeadNoticeInjectedSettingsAutoAccept is the complementary case to
// AC-FB-012: when the launcher DID inject (EnvMoaiKanbanSettingsInjected=1),
// the notice states cross-session messages are auto-accepted (not a verify
// advisory). This is AC-FB-013(d).
func TestKanbanLeadNoticeInjectedSettingsAutoAccept(t *testing.T) {
	clearKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanID, "tjlgt1")
	t.Setenv(config.EnvMoaiKanbanLeadAddr, "/tmp/moai-kanban-tjlgt1")
	t.Setenv(config.EnvMoaiKanbanSettingsInjected, "1")

	got := kanbanLeadNotice("tjlgt1", langEnglish)
	lowered := strings.ToLower(got)
	if !strings.Contains(lowered, "auto-accept") {
		t.Errorf("lead notice lacks the auto-accept notice:\n%s", got)
	}
}

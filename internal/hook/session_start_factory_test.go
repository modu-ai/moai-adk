package hook

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// clearFactoryEnv unsets every factory variable so each case starts from a
// known-absent state. t.Setenv registers the restore, so the process env is
// returned to its prior value when the test ends.
func clearFactoryEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{config.EnvMoaiFactory, config.EnvMoaiFactoryID, config.EnvMoaiFactoryLabel} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
}

// TestFactoryBootstrapNoticeSilentForOrdinarySession is the case that matters
// most for blast radius: a session that is not part of a factory run must be
// completely unaffected.
func TestFactoryBootstrapNoticeSilentForOrdinarySession(t *testing.T) {
	clearFactoryEnv(t)

	if got := factoryBootstrapNotice(); got != "" {
		t.Errorf("non-factory session got a notice: %q", got)
	}
}

// TestFactoryBootstrapNoticeLead asserts the lead branch names the run and
// carries one launch command per companion role.
func TestFactoryBootstrapNoticeLead(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactory, "1")
	t.Setenv(config.EnvMoaiFactoryID, "tjlgt1")

	got := factoryBootstrapNotice()
	if got == "" {
		t.Fatal("lead session got no notice")
	}
	if !strings.Contains(got, "tjlgt1") {
		t.Errorf("notice omits the run id: %q", got)
	}
	for _, role := range factory.CompanionRoles {
		want := "moai cc --name " + role + "-tjlgt1"
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
	// No companion may carry the factory token; the notice must not suggest it.
	if strings.Contains(got, "--factory") || strings.Contains(got, "moai cc -f") {
		t.Errorf("notice invites a companion to carry the chain seed:\n%s", got)
	}
}

// TestFactoryBootstrapNoticeCompanion asserts the companion branch confirms its
// role and run — and, critically, never prints the launch block, which would
// have four sessions each inviting four more.
func TestFactoryBootstrapNoticeCompanion(t *testing.T) {
	for _, role := range factory.CompanionRoles {
		t.Run(role, func(t *testing.T) {
			clearFactoryEnv(t)
			t.Setenv(config.EnvMoaiFactoryLabel, factory.CompanionLabel(role, "tjlgt1"))
			t.Setenv(config.EnvMoaiFactoryID, "tjlgt1")

			got := factoryBootstrapNotice()
			if !strings.Contains(got, role) || !strings.Contains(got, "tjlgt1") {
				t.Errorf("companion notice = %q, want it to name role %q and run %q", got, role, "tjlgt1")
			}
			if strings.Contains(got, "--name") {
				t.Errorf("companion printed the launch block (recursive bootstrap):\n%s", got)
			}
			if strings.Count(got, "\n") != 0 {
				t.Errorf("companion notice should be a single line, got:\n%s", got)
			}
		})
	}
}

// TestFactoryBootstrapNoticeFailsOpen asserts an unresolvable value degrades to
// emitting nothing rather than emitting something wrong or failing the session
// start, matching the surrounding hook code.
func TestFactoryBootstrapNoticeFailsOpen(t *testing.T) {
	t.Run("lead without a run id", func(t *testing.T) {
		clearFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactory, "1")

		if got := factoryBootstrapNotice(); got != "" {
			t.Errorf("emitted a notice with no run id: %q", got)
		}
	})

	t.Run("unparseable companion label", func(t *testing.T) {
		clearFactoryEnv(t)
		t.Setenv(config.EnvMoaiFactoryLabel, "oauth-migration")

		if got := factoryBootstrapNotice(); got != "" {
			t.Errorf("emitted a notice for a non-companion label: %q", got)
		}
	})
}

// TestFactoryBootstrapNoticeLabelWinsOverFactory pins the branch order: a
// session carrying both variables is treated as a companion, so it can never
// print the launch block.
func TestFactoryBootstrapNoticeLabelWinsOverFactory(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactory, "1")
	t.Setenv(config.EnvMoaiFactoryID, "tjlgt1")
	t.Setenv(config.EnvMoaiFactoryLabel, "plan-tjlgt1")

	got := factoryBootstrapNotice()
	if strings.Contains(got, "--name") {
		t.Errorf("a labelled session printed the launch block:\n%s", got)
	}
	if !strings.Contains(got, "companion") {
		t.Errorf("expected the companion notice, got: %q", got)
	}
}

// TestFactoryLeadNoticeOperatorSettingsAdvisory is AC-FB-012. When the launcher
// did NOT inject its own settings (the operator supplied --settings, or a write
// failure degraded to fail-open), the lead notice prints an advisory instructing
// the operator to verify crossSessionInbound: accept is present.
func TestFactoryLeadNoticeOperatorSettingsAdvisory(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactory, "1")
	t.Setenv(config.EnvMoaiFactoryID, "tjlgt1")
	// EnvMoaiFactorySettingsInjected is UNSET → the launcher did not inject.
	t.Setenv(config.EnvMoaiFactorySettingsInjected, "")
	_ = os.Unsetenv(config.EnvMoaiFactorySettingsInjected)

	got := factoryLeadNotice("tjlgt1")
	if got == "" {
		t.Fatal("expected a lead notice, got empty string")
	}
	lowered := strings.ToLower(got)
	if !strings.Contains(lowered, "verify") || !strings.Contains(lowered, "crosssessioninbound") || !strings.Contains(lowered, "accept") {
		t.Errorf("lead notice lacks the verify-crossSessionInbound-accept advisory:\n%s", got)
	}
}

// TestFactoryLeadNoticeInjectedSettingsAutoAccept is the complementary case to
// AC-FB-012: when the launcher DID inject (EnvMoaiFactorySettingsInjected=1),
// the notice states cross-session messages are auto-accepted (not a verify
// advisory). This is AC-FB-013(d).
func TestFactoryLeadNoticeInjectedSettingsAutoAccept(t *testing.T) {
	clearFactoryEnv(t)
	t.Setenv(config.EnvMoaiFactory, "1")
	t.Setenv(config.EnvMoaiFactoryID, "tjlgt1")
	t.Setenv(config.EnvMoaiFactorySettingsInjected, "1")

	got := factoryLeadNotice("tjlgt1")
	lowered := strings.ToLower(got)
	if !strings.Contains(lowered, "auto-accept") {
		t.Errorf("lead notice lacks the auto-accept notice:\n%s", got)
	}
}

package cli

// kanban_help_test.go verifies AC-FB-018 (lead entry documented) and
// AC-FB-019 (companion entry documented) for the cc.go and glm.go CLI help
// text. AC-FB-020 (no cmd.Flags() registration) is a SHOULD AC verified by
// grep at verification time.

import (
	"strings"
	"testing"
)

// TestACFB018_HelpDocumentsLeadEntry verifies AC-FB-018: the cc and glm help
// text contains -k / --kanban and states it seeds a plan->run->verify->sync chain.
func TestACFB018_HelpDocumentsLeadEntry(t *testing.T) {
	t.Parallel()
	for _, cmd := range []struct {
		name string
		long string
	}{
		{"cc", ccCmd.Long},
		{"glm", glmCmd.Long},
	} {
		t.Run(cmd.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(cmd.long, "-k") && !strings.Contains(cmd.long, "--kanban") {
				t.Errorf("%s help does not document -k / --kanban", cmd.name)
			}
			lowered := strings.ToLower(cmd.long)
			if !strings.Contains(lowered, "plan") || !strings.Contains(lowered, "run") ||
				!strings.Contains(lowered, "verify") || !strings.Contains(lowered, "sync") {
				t.Errorf("%s help does not state the plan->run->verify->sync chain", cmd.name)
			}
		})
	}
}

// TestACFB019_HelpDocumentsCompanionEntry verifies AC-FB-019: the cc and glm
// help text contains -k --name <role> and enumerates the four roles, plus the
// collision-bump rule the bare-role naming policy (card t56) adds.
func TestACFB019_HelpDocumentsCompanionEntry(t *testing.T) {
	t.Parallel()
	for _, cmd := range []struct {
		name string
		long string
	}{
		{"cc", ccCmd.Long},
		{"glm", glmCmd.Long},
	} {
		t.Run(cmd.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(cmd.long, "-k --name") {
				t.Errorf("%s help does not document the companion entry form -k --name", cmd.name)
			}
			for _, role := range []string{"plan", "run", "review", "sync"} {
				if !strings.Contains(cmd.long, role) {
					t.Errorf("%s help does not enumerate role %q", cmd.name, role)
				}
			}
			// The bare-role form is the documented launch shape, and a live-held
			// name bumping to the next free number is stated so the operator is
			// not surprised by a session named plan-1.
			if strings.Contains(cmd.long, "<role>-<run-id>") {
				t.Errorf("%s help still documents the retired run-id companion form", cmd.name)
			}
			if !strings.Contains(cmd.long, "bumped to the next free number") {
				t.Errorf("%s help does not state the companion name collision bump", cmd.name)
			}
		})
	}
}

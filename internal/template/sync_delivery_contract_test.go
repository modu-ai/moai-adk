// sync_delivery_contract_test.go — card t2009: the sync workflow documents must
// agree with the delivery and SPEC-ownership contracts they sit beside.
//
//   - sync/delivery.md Route A may not promise a direct commit on main
//     regardless of strategy, and its direct-push step must be gated by the
//     active mode's push_to_remote value (AGENTS.md §3: main advances through
//     PR merges; delivery-policy.md: phase agents do not push by default).
//   - sync/doc-execution.md may not tell the sync phase to rewrite approved
//     SPEC bodies: manager-docs owns only the status/updated frontmatter and
//     returns a blocker report so manager-spec edits the body.
package template

import (
	"strings"
	"testing"
)

func readTemplateDoc(t *testing.T, path string) string {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	data, err := fsReadAll(fsys, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// sectionBetween returns the text from the first occurrence of start up to the
// next occurrence of end after it, failing the test when either marker is gone.
func sectionBetween(t *testing.T, doc, start, end string) string {
	t.Helper()
	i := strings.Index(doc, start)
	if i < 0 {
		t.Fatalf("marker %q not found", start)
	}
	rest := doc[i:]
	j := strings.Index(rest[len(start):], end)
	if j < 0 {
		t.Fatalf("end marker %q not found after %q", end, start)
	}
	return rest[:len(start)+j]
}

func TestSyncDeliveryRouteAIsStrategyQualified(t *testing.T) {
	t.Parallel()
	doc := readTemplateDoc(t, ".claude/skills/moai/workflows/sync/delivery.md")

	var routeA string
	for line := range strings.SplitSeq(doc, "\n") {
		if strings.HasPrefix(line, "- **Route A") {
			routeA = line
			break
		}
	}
	if routeA == "" {
		t.Fatal("SYNC_ROUTE_A_MISSING: delivery.md carries no Route A bullet")
	}
	for _, token := range []string{"agentic-trunk", "git-flow", "WT-*"} {
		if !strings.Contains(routeA, token) {
			t.Errorf("SYNC_ROUTE_A_UNQUALIFIED: Route A bullet does not name %q; it promises a direct commit on main that the %s route contradicts", token, token)
		}
	}
}

func TestSyncDeliveryDirectPushHonoursPushToRemote(t *testing.T) {
	t.Parallel()
	doc := readTemplateDoc(t, ".claude/skills/moai/workflows/sync/delivery.md")

	block := sectionBetween(t, doc, "**Main branch** (direct commit):", "**Worktree context**")
	if !strings.Contains(block, "push_to_remote") {
		t.Errorf("SYNC_DIRECT_PUSH_UNGATED: the github-flow main-branch direct push does not consult git_strategy.{mode}.push_to_remote:\n%s", block)
	}
}

// TestSyncDeliveryCIMirrorDoesNotMaskFailure (S04): the local CI mirror example
// must keep tool presence apart from the check's exit status, and must read
// every parallel build's own status instead of ending on a bare `wait`.
func TestSyncDeliveryCIMirrorDoesNotMaskFailure(t *testing.T) {
	t.Parallel()
	doc := readTemplateDoc(t, ".claude/skills/moai/workflows/sync/delivery.md")

	section := sectionBetween(t, doc, "##### Step 3.1.5.2: Run Local Equivalents in Parallel", "##### Step 3.1.5.3")
	if strings.Contains(section, `|| echo "SKIP`) {
		t.Error("SYNC_CI_LINT_MASKED: `|| echo \"SKIP...\"` turns an installed linter's non-zero exit into a skip")
	}
	for line := range strings.SplitSeq(section, "\n") {
		if strings.TrimSpace(line) == "wait" {
			t.Error("SYNC_CI_BARE_WAIT: a bare `wait` returns zero whatever the individual builds did")
		}
	}
	for _, required := range []string{`wait "$pid"`, "NOT_RUN"} {
		if !strings.Contains(section, required) {
			t.Errorf("SYNC_CI_STATUS_UNSEPARATED: the CI mirror example carries no %q", required)
		}
	}

	eval := sectionBetween(t, doc, "##### Step 3.1.5.4: Evaluate Results", "##### Step 3.1.5.5")
	if !strings.Contains(eval, "NOT_RUN") {
		t.Error("SYNC_CI_NOTRUN_AS_PASS: Step 3.1.5.4 does not keep a check that did not run apart from a passing one")
	}
}

// TestSyncStopHookDescriptionMatchesTheHook (S07): sync.md must describe what
// sync-phase-quality-gate.sh does today — fast vet/build structural checks,
// blocking by default with MOAI_SYNC_GATE_BLOCKING=0 as the opt-out — and must
// not claim it runs lint, tests, or a coverage delta.
func TestSyncStopHookDescriptionMatchesTheHook(t *testing.T) {
	t.Parallel()
	doc := readTemplateDoc(t, ".claude/skills/moai/workflows/sync.md")

	var line string
	for l := range strings.SplitSeq(doc, "\n") {
		if strings.Contains(l, "sync-phase-quality-gate.sh") && strings.Contains(l, "Stop hook") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatal("SYNC_HOOK_DESCRIPTION_MISSING: sync.md carries no sentence describing the sync-phase-quality-gate.sh Stop hook")
	}
	for _, stale := range []string{"lint + test + coverage", "(MOAI_SYNC_GATE_BLOCKING=1)", "coverage regression"} {
		if strings.Contains(line, stale) {
			t.Errorf("SYNC_HOOK_DESCRIPTION_STALE: the Stop hook sentence still says %q", stale)
		}
	}
	for _, required := range []string{"vet", "build", "MOAI_SYNC_GATE_BLOCKING=0"} {
		if !strings.Contains(line, required) {
			t.Errorf("SYNC_HOOK_DESCRIPTION_INCOMPLETE: the Stop hook sentence does not name %q", required)
		}
	}
}

func TestSyncDocExecutionPreservesApprovedSpecBody(t *testing.T) {
	t.Parallel()
	doc := readTemplateDoc(t, ".claude/skills/moai/workflows/sync/doc-execution.md")

	section := sectionBetween(t, doc, "##### Step 2.2.1: SPEC Document Update", "#### Step 2.2.5")
	for _, banned := range []string{
		"Update spec.md requirements to reflect actual implementation",
		"Update plan.md with actual implementation steps taken",
		"Update acceptance.md with new acceptance criteria",
		"Append \"Implementation Notes\" section to spec.md",
	} {
		if strings.Contains(section, banned) {
			t.Errorf("SYNC_SPEC_BODY_REWRITE: Step 2.2.1 still instructs the sync phase to edit an approved SPEC body: %q", banned)
		}
	}
	for _, required := range []string{"manager-spec", "blocker report"} {
		if !strings.Contains(section, required) {
			t.Errorf("SYNC_SPEC_BODY_UNROUTED: Step 2.2.1 does not route body changes through %q", required)
		}
	}
}

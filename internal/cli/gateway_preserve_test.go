package cli

// Card t708 (SPEC-GATEWAY-ENVELOPE-REPAIR-001 M4): the mechanical
// no-weakening lock (AC-EVR-011). The card's own diff — merge-base
// develop..HEAD, recomputed at evaluation time and never a pinned base —
// must touch no file under internal/gateway/receipt/ and none of the §D
// PRESERVE surfaces of plan.md. The range collapses once the card merges
// into develop, so the live assertion is pre-merge-only and skips loudly
// rather than passing vacuously (the t543 merge-base lesson); the
// discriminator itself is tested in both directions so its failure mode is
// observed on a known violating input, not just on a green diff.

import (
	"os/exec"
	"strings"
	"testing"
)

// preservedGatewayFiles is plan.md §D's PRESERVE file list for this card.
const preservedGatewayFiles = "" +
	"internal/gateway/receipt/core.go\n" +
	"internal/gateway/receipt/projection.go\n" +
	"internal/gateway/receipt/store.go\n" +
	"internal/gateway/translate/receipt_history.go\n" +
	"internal/gateway/translate/request.go\n" +
	"internal/gateway/translate/response.go\n" +
	"internal/gateway/translate/stream.go\n" +
	"internal/gateway/translate/reasoning.go\n" +
	"internal/gateway/opaque/codec.go\n" +
	"internal/gateway/conversation/family.go\n" +
	"internal/cli/gateway_factory.go\n"

// preservedDiffViolations returns the changed files that weaken a preserved
// surface: anything under the receipt package, or any exact §D path.
func preservedDiffViolations(changed []string) []string {
	var out []string
	for _, f := range changed {
		if strings.HasPrefix(f, "internal/gateway/receipt/") {
			out = append(out, f)
			continue
		}
		for _, p := range strings.Split(preservedGatewayFiles, "\n") {
			if p != "" && f == p {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

func gitOutput(t *testing.T, args ...string) (string, bool) {
	t.Helper()
	cmd := exec.Command("git", args...)
	raw, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(raw)), true
}

func TestGatewayRepairCardDiffTouchesNoPreservedFile(t *testing.T) {
	base, ok := gitOutput(t, "merge-base", "develop", "HEAD")
	if !ok {
		base, ok = gitOutput(t, "merge-base", "origin/develop", "HEAD")
	}
	if !ok {
		t.Skip("no develop ref available for the merge-base computation")
	}
	head, ok := gitOutput(t, "rev-parse", "HEAD")
	if !ok {
		t.Skip("not a git checkout")
	}
	if base == head {
		t.Skip("card is merged into develop; the diff-scope assertion is pre-merge evaluation only")
	}
	raw, ok := gitOutput(t, "diff", "--name-only", base+".."+head)
	if !ok {
		t.Skipf("git diff unavailable: merge-base %s", base)
	}
	changed := strings.Split(raw, "\n")
	if len(changed) == 0 || (len(changed) == 1 && changed[0] == "") {
		t.Fatalf("card diff measured empty against merge-base %s — unmeasurable, not clean", base)
	}
	if violations := preservedDiffViolations(changed); len(violations) > 0 {
		t.Fatalf("card diff weakens preserved gateway surfaces: %v", violations)
	}
}

// The discriminator's failure mode is observed on a synthetic violating
// diff, so the lock is capable of failing and its red is on record.
func TestPreservedDiffDiscriminatorRejectsViolations(t *testing.T) {
	synthetic := []string{
		"internal/cli/gateway_repair.go",
		"internal/gateway/receipt/store.go",
		"internal/gateway/translate/receipt_history.go",
		"internal/gateway/conversation/family.go",
		"README.md",
	}
	got := preservedDiffViolations(synthetic)
	want := []string{
		"internal/gateway/receipt/store.go",
		"internal/gateway/translate/receipt_history.go",
		"internal/gateway/conversation/family.go",
	}
	if len(got) != len(want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("violations[%d] = %s, want %s", i, got[i], want[i])
		}
	}
	if clean := preservedDiffViolations([]string{"internal/cli/gateway_repair.go", "docs/x.md"}); len(clean) != 0 {
		t.Fatalf("clean diff flagged: %v", clean)
	}
}

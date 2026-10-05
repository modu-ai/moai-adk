package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-GSA-002 / AC-GSA-003 — the CLI boundary for an object-present,
// non-ancestor stamp: exit 2, a recovery message that names the actual fix,
// and NO numeric freshness row anywhere on stdout.
//
// The four tokens are checked separately because each one fails a different
// wrong implementation: dropping "unreachable" hides which of the two
// comparability failures occurred; dropping "freshness unmeasured" lets the
// reader take it for a stale verdict; dropping "regenerate" or "stamp" leaves
// only half the recovery, and the half-recovery (a bare re-stamp over an
// untouched body) is exactly what the freshness contract forbids.
func TestGraphCheckCmd_ExistingNonAncestorStampExitsTwoWithRecovery(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)
	baseBranch := checkFixtureGit(t, root, "symbolic-ref", "--short", "HEAD")
	checkFixtureGit(t, root, "switch", "-q", "-c", "stamp-side")
	if err := os.WriteFile(filepath.Join(root, "internal", "side.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "internal/side.go")
	checkFixtureGit(t, root, "commit", "-q", "-m", "side stamp")
	stamp := checkFixtureGit(t, root, "rev-parse", "HEAD")
	checkFixtureGit(t, root, "switch", "-q", baseBranch)
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "provenance.json"),
		marshalCodemapsProvenance(t, root, stamp), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runGraphCheck(t, root)

	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 2 {
		t.Fatalf("existing non-ancestor stamp must exit 2; err=%v stdout=%q stderr=%q", err, out, errOut)
	}
	surface := strings.ToLower(errOut)
	for _, token := range []string{"unreachable", "freshness unmeasured", "regenerate", "stamp"} {
		if !strings.Contains(surface, token) {
			t.Errorf("stderr missing recovery token %q: %s", token, errOut)
		}
	}
	if strings.Contains(out, "metric=described-source-diff value=") {
		t.Errorf("unmeasured error rendered a numeric freshness row on stdout: %s", out)
	}
	if strings.Contains(surface, "measured from:") || strings.Contains(surface, "contribution:") {
		t.Errorf("unmeasured error rendered attribution it never computed: %s", errOut)
	}
}

// AC-GSA-002 mutant guard — the recovery block is specific to the unreachable
// state. A different system error (an unresolvable stamp) must NOT inherit the
// regenerate-and-restamp advice: its fix is history, not regeneration.
func TestGraphCheckCmd_UnresolvableStampOmitsRegenerationAdvice(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "provenance.json"),
		marshalCodemapsProvenance(t, root, strings.Repeat("0", 40)), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runGraphCheck(t, root)

	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 2 {
		t.Fatalf("an unresolvable stamp must exit 2; err=%v stdout=%q stderr=%q", err, out, errOut)
	}
	if strings.Contains(strings.ToLower(errOut), "unreachable stamp") {
		t.Errorf("an unresolvable stamp was labelled unreachable — the two comparability failures must stay distinguishable: %s", errOut)
	}
}

// runGraphCheck executes `graph check --root <root>` against a fixture and
// returns stdout, stderr and the command error.
func runGraphCheck(t *testing.T, root string) (string, string, error) {
	t.Helper()
	cmd := newGraphCmd()
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

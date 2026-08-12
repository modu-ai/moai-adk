package fix

// nonoverlap_test.go — M3.5 non-overlap grep guard (SPEC-NAVIGATOR-SYNC-005,
// AC-NS5-006). Carries forward the M0/M1/M2 nonoverlap_test.go pattern
// (internal/navigator/{sync,detect,route}/nonoverlap_test.go) so the Fix
// layer's consumer-only + non-overlap contract is mechanically observable.
//
// Per AC-NS5-006 the guard asserts:
//
//  1. The ONLY matches for the M0/M1/M2 read surfaces (work-items.json,
//     nav-graph.json, the detect JSONL path, capability-map.md,
//     audit-report.{md,json}, capability-symbols.{md,json}) in the production
//     source are in a READ context (os.ReadFile / os.Open).
//  2. There are ZERO write targets (os.WriteFile / os.Rename) for tiers.json,
//     blueprint/, decisions/, or work-items.md anywhere in the production
//     source.
//
// apply.go's post-approval writes to live doc surfaces ARE the
// approved-subtree patches (the intended exception, governed by AC-NS5-008c —
// excluded from the write-target grep per AC-NS5-006 text: "apply.go is
// excluded from this grep because its writes ARE the approved-subtree
// patches"). The READ-context check below DOES scan apply.go (it reads
// nav-graph / work-items / live docs as inputs), which is permitted.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readSurfaceLiterals names the M0/M1/M2/M4 outputs + live doc surfaces the
// Fix layer CONSUMES read-only (REQ-NS5-005). Any production-source match for
// one of these literals MUST be in a READ context (os.ReadFile / os.Open).
var readSurfaceLiterals = []string{
	"work-items.json",
	"nav-graph.json",
	"navigator-detect", // the M1 detect state directory
	"capability-map.md",
	"audit-report.md",
	"audit-report.json",
	"capability-symbols.md",
	"capability-symbols.json",
}

// forbiddenWriteTargets names the surfaces the Fix layer MUST NEVER write to
// (REQ-NS5-005 / AC-NS5-006). These belong to the M0/M1/M2/M4 predecessor
// chains (tiers.json from M4, blueprint/ + decisions/ from the 001/003 chains,
// work-items.md — distinct from the M2 work-items.json the Fix layer reads).
// A production-source line that names one of these on a WriteFile/Rename line
// is the regression this guard catches.
var forbiddenWriteTargets = []string{
	"tiers.json",
	"blueprint/",
	"decisions/",
	"work-items.md",
}

// TestNonOverlap_ReadSurfacesAreReadOnly exercises AC-NS5-006 read-context
// half: every production-source mention of a read surface literal MUST be on
// a line carrying os.ReadFile or os.Open (a READ context). A write verb
// targeting a read surface is the regression (the Fix layer must not become a
// producer of M0/M1/M2/M4 outputs).
//
// apply.go's atomic-rename of an APPROVED subtree to a live doc surface IS a
// permitted write to capability-map.md / audit-report.json /
// capability-symbols.json (the intended exception, AC-NS5-008c). The guard
// therefore does NOT flag those three surfaces when the write is in apply.go.
// It DOES flag any other write context, and it flags writes to the remaining
// read surfaces (work-items.json, nav-graph.json, detect JSONL) which the Fix
// layer must never write at all.
func TestNonOverlap_ReadSurfacesAreReadOnly(t *testing.T) {
	t.Parallel()
	pkgDir := "."

	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		base := filepath.Base(m)
		if strings.HasSuffix(base, "_test.go") {
			continue // tests name the literals as the assertion target
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, line := range strings.Split(src, "\n") {
			if !strings.Contains(line, "WriteFile") && !strings.Contains(line, "os.Rename") {
				continue
			}
			for _, lit := range readSurfaceLiterals {
				if !strings.Contains(line, lit) {
					continue
				}
				// Permitted exception: apply.go writing the approved-subtree
				// patch to a live doc surface (AC-NS5-008c). The live doc
				// surfaces are capability-map.md, audit-report.{md,json},
				// capability-symbols.{md,json}. The other read surfaces
				// (work-items.json, nav-graph.json, detect state) are NEVER
				// writable by the Fix layer.
				if base == "apply.go" && isLiveDocSurface(lit) {
					continue
				}
				t.Errorf("%s: write verb targets read surface %q (Fix layer is consumer-only, REQ-NS5-005): %s",
					base, lit, strings.TrimSpace(line))
			}
		}
	}
}

// TestNonOverlap_NoForbiddenWriteTargets exercises AC-NS5-006 write-target
// half: the production source MUST NOT reference tiers.json, blueprint/,
// decisions/, or work-items.md as a write target (WriteFile / os.Rename).
// These belong to the M4 tiers chain + the 001/003 predecessor chains and are
// out of the Fix layer's scope entirely.
func TestNonOverlap_NoForbiddenWriteTargets(t *testing.T) {
	t.Parallel()
	pkgDir := "."

	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		base := filepath.Base(m)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, line := range strings.Split(src, "\n") {
			if !strings.Contains(line, "WriteFile") && !strings.Contains(line, "os.Rename") {
				continue
			}
			for _, lit := range forbiddenWriteTargets {
				if strings.Contains(line, lit) {
					t.Errorf("%s: write verb targets forbidden surface %q (out of Fix layer scope, AC-NS5-006): %s",
						base, lit, strings.TrimSpace(line))
				}
			}
		}
	}
}

// TestNonOverlap_ProducerChainScriptsUntouched is a structural reminder
// (AC-NS5-006 scope): the three predecessor chain scripts
// (.claude/skills/moai-workflow-project/scripts/navigator-{audit,regen,enrich}.sh)
// are byte-unchanged by the M3 run-phase diff (they live outside this package).
// This test asserts the Fix layer's production source does NOT import, invoke,
// or exec those scripts — the Fix layer reads the OUTPUTS (work-items.json +
// detect JSONL + nav-graph.json) directly via os.ReadFile, not by running the
// producer chains (REQ-NS5-005b).
func TestNonOverlap_ProducerChainScriptsUntouched(t *testing.T) {
	t.Parallel()
	pkgDir := "."

	matches, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	forbiddenExecRefs := []string{
		"navigator-audit.sh",
		"navigator-regen.sh",
		"navigator-enrich.sh",
	}
	for _, m := range matches {
		base := filepath.Base(m)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		b, err := os.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}
		src := string(b)
		for _, ref := range forbiddenExecRefs {
			if strings.Contains(src, ref) {
				t.Errorf("%s: Fix layer references predecessor chain script %q (Fix reads the outputs directly, not by exec'ing the producers, REQ-NS5-005b)",
					base, ref)
			}
		}
	}
}

// isLiveDocSurface returns true for the live doc surface filenames the apply
// step is permitted to atomic-rename an approved-subtree patch into (the
// AC-NS5-008c exception to the read-surface grep).
func isLiveDocSurface(lit string) bool {
	switch lit {
	case "capability-map.md", "audit-report.md", "audit-report.json",
		"capability-symbols.md", "capability-symbols.json":
		return true
	}
	return false
}

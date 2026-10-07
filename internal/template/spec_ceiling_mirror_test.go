package template_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resolves the repository root from this file's location, so the
// deployed paths in the table below are repo-root-relative regardless of the
// test binary's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// TestSPECEditedRegionsSynced is AC-ACE-014's regression guard
// (SPEC-AUDIT-CEILING-001 REQ-ACE-014): every deployed file this SPEC edits
// carries its SPEC-edited regions byte-identical in the template mirror —
// the mirror is what go:embed ships, so drift there means user projects
// receive a stale contract.
//
// Region granularity: a region is an h2 section ("## ...") or, for
// plan-auditor.md's Retry Loop Contract, the exact paragraphs this SPEC's
// B1 prose correction authored. That file is the named known-FAIL
// carve-out: its whole-file drift (measured DIFF at f2f815008 and
// re-measured DIFF at 69a085b2d, research.md §3) reaches inside the Retry
// Loop Contract section — the mirror's pre-A6 ceiling paragraphs pre-date
// this SPEC and are follow-up repair material — so the criterion checks the
// edited hunks and passes only when the residual diff lies wholly outside
// them, never encoding the unrelated drift as expected state.
var specEditedRegions = []struct {
	file    string
	mirror  string
	regions []string
}{
	{
		file:    ".moai/docs/audit-artifact-convention.md",
		mirror:  "internal/template/templates/.moai/docs/audit-artifact-convention.md",
		regions: []string{"## What"},
	},
	{
		file:   ".claude/agents/moai/plan-auditor.md",
		mirror: "internal/template/templates/.claude/agents/moai/plan-auditor.md",
		regions: []string{
			"## Verdict File Machine Lines",
			"**STOP escalation on score regression.**",
			"**Ceiling policy (hard limit).**",
		},
	},
}

// extractRegion returns one region: an h2 section when anchor starts with
// "## " (from the heading line to just before the next h2 heading), else the
// paragraph beginning with the anchor prefix (to the next blank line).
func extractRegion(raw, anchor string) string {
	lines := strings.Split(raw, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, anchor) || strings.TrimSpace(l) == anchor {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	if strings.HasPrefix(anchor, "## ") {
		for j := start + 1; j < len(lines); j++ {
			if strings.HasPrefix(lines[j], "## ") {
				return strings.Join(lines[start:j], "\n")
			}
		}
		return strings.Join(lines[start:], "\n")
	}
	for j := start + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) == "" {
			return strings.Join(lines[start:j], "\n")
		}
	}
	return strings.Join(lines[start:], "\n")
}

func TestSPECEditedRegionsSynced(t *testing.T) {
	root := repoRoot(t)
	for _, tc := range specEditedRegions {
		deployed, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tc.file)))
		if err != nil {
			t.Fatalf("read deployed %s: %v", tc.file, err)
		}
		mirrored, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(tc.mirror)))
		if err != nil {
			t.Fatalf("read mirror %s: %v", tc.mirror, err)
		}
		for _, anchor := range tc.regions {
			a := extractRegion(string(deployed), anchor)
			b := extractRegion(string(mirrored), anchor)
			if a == "" {
				t.Fatalf("deployed %s carries no region %q", tc.file, anchor)
			}
			if b == "" {
				t.Fatalf("mirror %s carries no region %q", tc.mirror, anchor)
			}
			if a != b {
				t.Errorf("SPEC-EDITED-REGION-DRIFT: %s region %q differs from its template mirror %s", tc.file, anchor, tc.mirror)
			}
		}
	}
}

// sync_4dim_spec_guard_test.go: CI guard for card t1686 — the sync-audit-4dim
// workflow must refuse explicitly when launched WITHOUT a usable spec_id
// (the Class B no-SPEC card shape) instead of fabricating the ghost SPEC
// 'SPEC-UNKNOWN' and spawning five agents on it. Observed live in t1667: with
// a ghost SPEC id the Context Explore agent's discretion adopted an unrelated
// SPEC (SPEC-ZONE-SHELL-PARSING-001) as the audit surface, so the four judges
// scored a delta the card never touched.
//
// Both copies are asserted — the deployed workflow and its template twin —
// because the template twin is the distribution SSOT: a fix applied only to
// the deployed copy is lost on the next `moai update`, and a fix applied only
// to the template is invisible to the running repo.
//
// This is a source-inspection guard, not an execution test: the workflow
// script runs only inside the dynamic-workflows runtime (global `agent` /
// `parallel` / `phase` / `args` + top-level return), so the Go suite asserts
// the three guard elements and their ORDER (guard before the Context phase
// spawn). Sentinel on failure: SYNC_4DIM_SPEC_GUARD_MISSING.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sync4dimWorkflowRelPaths lists both copies of the workflow script, relative
// to the project root.
var sync4dimWorkflowRelPaths = []string{
	".claude/workflows/sync-audit-4dim.js",
	"internal/template/templates/.claude/workflows/sync-audit-4dim.js",
}

// Required guard elements (card t1686 repair):
const (
	// pre-spawn refusal branch: the no-surface verdict carries a dedicated
	// `missing` entry naming the absent audit surface.
	guardMarkerMissingSurface = "missing: ['audit-surface']"
	// the ghost-marker rejection: the script must treat 'SPEC-UNKNOWN' as a
	// refusal condition, not only as a fallback assignment. The marker is the
	// EXECUTABLE comparison form — a comment quoting the ghost marker must
	// never satisfy this assertion (codex card-review round 1, P2).
	guardMarkerGhostReject = "SPEC_ID.toUpperCase() === 'SPEC-UNKNOWN'"
	// the Context-prompt discretion block: when a spec_id IS injected but its
	// SPEC directory does not exist, the agent must not adopt a different SPEC.
	promptMarkerNoAdoption = "do NOT adopt or substitute"
	// the spawn point the guard must precede.
	contextSpawnMarker = "phase('Context')"
)

// codeOnlySource strips full-line JS comments so a marker hidden in prose can
// never satisfy a guard assertion. The workflow scripts carry no block
// comments before the guarded region, so dropping lines whose first
// non-blank token is `//` is a complete comment filter here.
func codeOnlySource(src string) string {
	var kept []string
	for _, ln := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "//") {
			continue
		}
		kept = append(kept, ln)
	}
	return strings.Join(kept, "\n")
}

// TestSyncAudit4DimRefusesSpecIdlessRun asserts both copies of
// sync-audit-4dim.js carry the pre-spawn surface guard, the ghost-marker
// rejection, and the Context-prompt no-adoption block, with the guard placed
// BEFORE the Context phase spawn (an early return must prevent the agent
// fan-out).
func TestSyncAudit4DimRefusesSpecIdlessRun(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for _, rel := range sync4dimWorkflowRelPaths {
		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			src, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
			if err != nil {
				t.Fatalf("workflow script unreadable: %v", err)
			}
			// All assertions run over the COMMENT-STRIPPED source: a comment
			// sitting earlier in the file than the executable guard must not
			// satisfy any marker (codex card-review round 1, P2).
			source := codeOnlySource(string(src))

			assertContains := func(marker, what string) int {
				idx := strings.Index(source, marker)
				if idx < 0 {
					t.Errorf("SYNC_4DIM_SPEC_GUARD_MISSING: %s is absent from %s "+
						"(card t1686: the no-spec_id launch must refuse explicitly instead of "+
						"spawning agents on a ghost SPEC surface)", what, rel)
				}
				return idx
			}

			idxMissing := assertContains(guardMarkerMissingSurface, "pre-spawn refusal branch (missing: ['audit-surface'])")
			idxGhost := assertContains(guardMarkerGhostReject, "ghost-marker rejection ('SPEC-UNKNOWN' treated as a refusal condition)")
			idxPrompt := assertContains(promptMarkerNoAdoption, "Context-prompt no-adoption block (do NOT adopt or substitute)")
			idxSpawn := strings.Index(source, contextSpawnMarker)
			if idxSpawn < 0 {
				t.Fatalf("workflow structure unexpected: %s lost its %s marker", rel, contextSpawnMarker)
			}
			t.Logf("%s: guard markers at refusal=%d ghost=%d prompt=%d spawn=%d",
				rel, idxMissing, idxGhost, idxPrompt, idxSpawn)

			// Ordering: the refusal branch and the ghost rejection must both sit
			// BEFORE the Context spawn — otherwise the guard runs after agents
			// are already launched. (The prompt marker naturally sits after the
			// spawn call in source order; only the guard ordering is asserted.)
			if idxMissing >= 0 && idxMissing > idxSpawn {
				t.Errorf("SYNC_4DIM_SPEC_GUARD_MISSING: refusal branch found at %d, AFTER the %s marker at %d — "+
					"the guard must run before any agent spawns", idxMissing, contextSpawnMarker, idxSpawn)
			}
			if idxGhost >= 0 && idxGhost > idxSpawn {
				t.Errorf("SYNC_4DIM_SPEC_GUARD_MISSING: ghost-marker rejection found at %d, AFTER the %s marker at %d — "+
					"'SPEC-UNKNOWN' must be a refusal condition evaluated before the Context spawn", idxGhost, contextSpawnMarker, idxSpawn)
			}
		})
	}
}

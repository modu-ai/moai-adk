package hook

// AC-AE-008 (SPEC-AUTONOMY-ESCALATION-001, REQ-AE-007): with frozen-files in
// the contract's invariants, a write by any caller — not only the harness
// learner — to a path matching the A1-derived frozen_files list (registry
// Frozen targets ∪ **/CLAUDE.md ∪ ownership.never) trips invariant-violation
// (frozen-file), with the glob taken from the list cached at arming. Without
// frozen-files, the same writes trip no frozen-file record.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/escalation/escalationtest"
)

const frozenTargetRel = ".claude/rules/moai/core/frozen-target.md"

func frozenFixture(t *testing.T, withFrozenInvariant bool) *escalationtest.Worktree {
	t.Helper()
	w := escalationHookFixture(t, "")
	w.Write(frozenTargetRel, "# frozen\n")
	w.WriteRegistry(escalationtest.RegistryEntry{ID: "CONST-V3R2-001", Zone: "Frozen", File: frozenTargetRel})
	w.AddSpec("SPEC-A-001", escalationtest.SpecOptions{Edit: func(d string) string {
		d = strings.Replace(d, `    - "internal/fixture/**"`, `    - "internal/foo/**"`, 1)
		d = strings.Replace(d, `    - "internal/x/**"`, `    - "internal/foo/secret/**"`, 1)
		if !withFrozenInvariant {
			d = strings.Replace(d, "  - frozen-files\n", "", 1)
		}
		return d
	}})
	return w
}

// frozenRecords returns the frozen-file invariant-violation records' bodies.
func frozenRecords(t *testing.T, w *escalationtest.Worktree) []string {
	t.Helper()
	var out []string
	for _, name := range escalationRecordNames(t, w) {
		if !strings.HasPrefix(name, escalation.ClassInvariantViolation+"-") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(escalation.RecordDir(w.Root, w.Card), name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "frozen-file") {
			out = append(out, string(data))
		}
	}
	return out
}

func TestFrozenFileUnionTrips(t *testing.T) {
	targets := map[string]string{
		"docs/CLAUDE.md":           "**/CLAUDE.md",
		frozenTargetRel:            frozenTargetRel,
		"internal/foo/secret/y.go": "internal/foo/secret/**",
	}
	write := func(t *testing.T, pre Handler, w *escalationtest.Worktree, rel string) {
		in := escalationInput(w, "Write", map[string]string{"file_path": w.Path(rel), "content": "x\n"})
		in.AgentType = "manager-develop" // any identity other than the harness learner
		if _, err := pre.Handle(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("with-frozen-files", func(t *testing.T) {
		w := frozenFixture(t, true)
		pre, _ := escalationHandlers(t, w)
		for rel := range targets {
			write(t, pre, w, rel)
		}
		recs := frozenRecords(t, w)
		if len(recs) != 3 {
			t.Fatalf("frozen-file records = %d (%v), want 3", len(recs), escalationRecordNames(t, w))
		}
		files, _ := escalation.CardFilesFor(w.Root, w.Card)
		st, _, err := escalation.ReadCardState(files.State)
		if err != nil || st.Armed == nil {
			t.Fatalf("state = %+v, %v", st, err)
		}
		for rel, glob := range targets {
			found := false
			for _, r := range recs {
				if strings.Contains(r, rel) && strings.Contains(r, glob) {
					found = true
				}
			}
			if !found {
				t.Errorf("no frozen-file record for %s naming glob %s", rel, glob)
			}
			cached := false
			for _, g := range st.Armed.FrozenFiles {
				cached = cached || g == glob
			}
			if !cached {
				t.Errorf("glob %s is not in the frozen_files cached at arming: %v", glob, st.Armed.FrozenFiles)
			}
		}
	})

	t.Run("without-frozen-files", func(t *testing.T) {
		w := frozenFixture(t, false)
		pre, _ := escalationHandlers(t, w)
		for rel := range targets {
			write(t, pre, w, rel)
		}
		if recs := frozenRecords(t, w); len(recs) != 0 {
			t.Errorf("frozen-file records without the invariant: %d", len(recs))
		}
		if !escalationCardLog(t, w).Armed() {
			t.Error("fixture did not arm")
		}
	})
}

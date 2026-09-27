package hook

import (
	"slices"
	"testing"
)

// TestFrozenInstructionFiles guards the harness-learner frozen instruction set
// (REQ-IFU-013). The set must hold exactly the four instruction basenames, and
// each entry must be denied to the harness-learner wherever it sits in the
// tree — the set matches on basename, so a nested path is the case that proves
// it. One sub-case per entry, so a dropped entry fails by name.
func TestFrozenInstructionFiles(t *testing.T) {
	want := []string{"AGENTS.local.md", "AGENTS.md", "CLAUDE.local.md", "CLAUDE.md"}

	if got := slices.Sorted(slices.Values(frozenInstructionFiles)); !slices.Equal(got, want) {
		t.Fatalf("frozenInstructionFiles = %v, want exactly %v", frozenInstructionFiles, want)
	}

	h := &preToolHandler{}
	for _, name := range want {
		t.Run(name, func(t *testing.T) {
			for _, path := range []string{name, "nested/dir/" + name} {
				sentinel, reason := h.checkHarnessFrozenZone(harnessLearnerIdentity, path)
				if sentinel != SentinelHarnessFrozenInstruction {
					t.Errorf("harness-learner write to %q: sentinel = %q, want %q", path, sentinel, SentinelHarnessFrozenInstruction)
				}
				if reason == "" {
					t.Errorf("harness-learner write to %q: empty deny reason", path)
				}
				if s, _ := h.checkHarnessFrozenZone("manager-develop", path); s != "" {
					t.Errorf("non-learner write to %q: sentinel = %q, want none", path, s)
				}
			}
		})
	}
}

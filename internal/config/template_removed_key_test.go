package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// TestTemplateRemovedKeySurvivesUserConfig (AC-CKH-014, REQ-CKH-011) asserts the
// report-once, delete-never posture over the merge engine for a config key that
// M1 classified D and a future template removed. The merge engine itself is
// owned by SPEC-UPDATE-YAML-PRESERVE-001 (plan §D3); this test pins the posture
// for keys this SPEC removes.
//
// The D-classified key used here is design.evolution.max_active_learnings
// (spec.md §A.5: a key whose stated effect is decided by two hardcoded
// constants; M1 inventory marks it class D). The test simulates a future
// template that has dropped the key while the user's file still carries a
// user-set value for it.
//
// Three assertions:
//
//  (a) the user-set key AND its user-set value survive the merge (delete-never);
//  (b) the removal IS surfaced — the retained-key advisory names the key
//      (reported at least once, so the user is notified, not silently dropped);
//  (c) no other user-set key is dropped (a hand-added unrelated user key is
//      also retained).
//
// Deferred sub-clause (AC-CKH-014 part b literal "emits no further report on a
// second merge over the same tree"): the current merge engine emits the
// retained-key advisory on every call that finds an old-only key and keeps no
// cross-call "already reported" state, so a second merge over the same tree
// re-emits the advisory. Cross-call report-once idempotency is a merge-engine
// behaviour owned by sibling SPEC-UPDATE-YAML-PRESERVE-001 and is out of scope
// for M6 (plan §F M6 = template-neutrality + E5 handoff, NOT merge-engine
// modification). This test pins the load-bearing delete-never + reported-at-
// least-once posture M6 can guarantee; the cross-call idempotency is recorded
// as a deferred gap in progress.md §E.2.
func TestTemplateRemovedKeySurvivesUserConfig(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll sections: %v", err)
	}

	// template-v1 shipped the D key; template-v2 removed it; the user retains a
	// custom value (200, not the default 50) plus a hand-added user key.
	const sectionFile = "design.yaml"
	templateV1 := []byte("evolution:\n  max_active_learnings: 50\n  max_evolution_rate_per_week: 3\n")
	templateV2 := []byte("evolution:\n  max_evolution_rate_per_week: 3\n")
	userFile := []byte("evolution:\n  max_active_learnings: 200\n  max_evolution_rate_per_week: 3\nuser_custom_key: keep-me\n")

	userPath := filepath.Join(sectionsDir, sectionFile)
	if err := os.WriteFile(userPath, userFile, 0o644); err != nil {
		t.Fatalf("WriteFile user config: %v", err)
	}

	// Capture the retained-key advisory that MergeYAML3Way writes to the sink.
	var advisory bytes.Buffer
	restore := backup.SetRetainedKeySinkForTest(&advisory)
	defer restore()

	merged, err := backup.MergeYAML3Way(templateV2, userFile, templateV1)
	if err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if err := os.WriteFile(userPath, merged, 0o644); err != nil {
		t.Fatalf("WriteFile merged result: %v", err)
	}

	result, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatalf("ReadFile merged result: %v", err)
	}

	// (a) delete-never: the D key AND its user-set value survive.
	if !strings.Contains(string(result), "max_active_learnings: 200") {
		t.Errorf("(a) delete-never FAILED: D key max_active_learnings:200 missing from merged output:\n%s", result)
	}

	// (b) reported-at-least-once: the advisory names the retained key.
	if !strings.Contains(advisory.String(), "max_active_learnings") {
		t.Errorf("(b) reported-at-least-once FAILED: retained-key advisory did not name max_active_learnings; got:\n%s", advisory.String())
	}

	// (c) no-other-key-dropped: the hand-added unrelated user key survives too.
	if !strings.Contains(string(result), "user_custom_key: keep-me") {
		t.Errorf("(c) no-other-key-dropped FAILED: user_custom_key missing from merged output:\n%s", result)
	}
}

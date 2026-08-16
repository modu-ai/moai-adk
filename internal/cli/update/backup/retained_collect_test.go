// t63 — retained-key collection tests for the 3-way merge / restore path.
//
// The production restore path (RestoreMoaiConfigRetained →
// MergeYAML3WayRetained) must COLLECT retained keys instead of writing the
// legacy per-key advisory text to the stderr sink, so the update render layer
// owns the output surface (one TUI summary line by default, key list under
// --verbose). The legacy entry points (MergeYAML3Way / RestoreMoaiConfig)
// keep the REQ-UYP-007 advisory-text-on-sink contract byte-identically for
// their existing callers and contract tests.

package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// seedRestoreTree builds the minimal 3-way restore layout:
//
//	root/.moai/config/sections/<file>            — deployed NEW template data
//	backupDir/sections/<file>                    — user's OLD data (extra key)
//	backupDir/.template-defaults/sections/<file> — BASE (previous template)
//
// and returns the project root. The old file carries one key absent from the
// new file so the merge retains it.
func seedRestoreTree(t *testing.T, file, newData, oldData, baseData string) string {
	t.Helper()
	root := t.TempDir()
	backupDir := filepath.Join(root, "backup")

	for _, content := range map[string]struct{ path, data string }{
		"target": {filepath.Join(root, defs.MoAIDir, defs.ConfigSubdir, "sections", file), newData},
		"old":    {filepath.Join(backupDir, "sections", file), oldData},
		"base":   {filepath.Join(backupDir, ".template-defaults", "sections", file), baseData},
	} {
		if err := os.MkdirAll(filepath.Dir(content.path), defs.DirPerm); err != nil {
			t.Fatalf("MkdirAll %s: %v", content.path, err)
		}
		if err := os.WriteFile(content.path, []byte(content.data), defs.FilePerm); err != nil {
			t.Fatalf("WriteFile %s: %v", content.path, err)
		}
	}
	return root
}

// withSwappedSink swaps retainedKeySink for a buffer and restores it on
// cleanup, returning the buffer. Tests using it MUST NOT run in parallel:
// the sink is a package global (same discipline as the existing
// REQ-UYP-007 contract tests, which swap it with defer, unparallelized).
func withSwappedSink(t *testing.T) *bytes.Buffer {
	t.Helper()
	var sink bytes.Buffer
	restore := SetRetainedKeySinkForTest(&sink)
	t.Cleanup(restore)
	return &sink
}

// The production merge entry collects the retained keys AND writes nothing to
// the advisory text sink — no raw stderr line may fire while the update
// progress line is mid-redraw.
func TestMergeYAML3WayRetained_CollectsWithoutSinkText(t *testing.T) {
	sink := withSwappedSink(t)

	newData := []byte("shared: new_val\n")
	oldData := []byte("shared: new_val\nremoved: keep-me\n")
	baseData := []byte("shared: base_val\n")

	merged, keys, err := MergeYAML3WayRetained(newData, oldData, baseData)
	if err != nil {
		t.Fatalf("MergeYAML3WayRetained: %v", err)
	}
	if !strings.Contains(string(merged), "removed: keep-me") {
		t.Errorf("retained key must still survive the merge (delete-never), got:\n%s", merged)
	}
	if len(keys) != 1 || keys[0] != "removed" {
		t.Errorf("retained keys = %v, want [removed]", keys)
	}
	if sink.Len() != 0 {
		t.Errorf("Retained entry must not write advisory text to the sink, got:\n%s", sink.String())
	}
}

// The legacy entry keeps the REQ-UYP-007 contract: advisory text naming the
// retained key reaches the sink.
func TestMergeYAML3Way_LegacySinkTextPreserved(t *testing.T) {
	sink := withSwappedSink(t)

	newData := []byte("shared: new_val\n")
	oldData := []byte("shared: new_val\nremoved: keep-me\n")
	baseData := []byte("shared: base_val\n")

	if _, err := MergeYAML3Way(newData, oldData, baseData); err != nil {
		t.Fatalf("MergeYAML3Way: %v", err)
	}
	if !strings.Contains(sink.String(), "removed") {
		t.Errorf("legacy entry must keep the advisory text contract (REQ-UYP-007), got:\n%s", sink.String())
	}
}

// The production restore entry aggregates retained keys across section files
// (section-relative path + dotted key) without touching the text sink.
func TestRestoreMoaiConfigRetained_CollectsRefsAndStaysSilentOnSink(t *testing.T) {
	root := seedRestoreTree(t, "design.yaml",
		"evolution:\n  max_rate: 3\n",
		"evolution:\n  max_rate: 3\n  max_active_learnings: 200\n",
		"evolution:\n  max_rate: 3\n")
	backupDir := filepath.Join(root, "backup")

	sink := withSwappedSink(t)
	refs, err := RestoreMoaiConfigRetained(root, backupDir, nil)
	if err != nil {
		t.Fatalf("RestoreMoaiConfigRetained: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("refs = %+v, want exactly 1", refs)
	}
	if refs[0].Section != "design.yaml" || refs[0].Key != "evolution.max_active_learnings" {
		t.Errorf("ref = {%s %s}, want {design.yaml evolution.max_active_learnings}", refs[0].Section, refs[0].Key)
	}
	if sink.Len() != 0 {
		t.Errorf("Retained restore must not write advisory text to the sink, got:\n%s", sink.String())
	}
}

// The legacy RestoreMoaiConfig wrapper re-emits the per-key advisory text for
// callers not yet routed through the TUI renderer (clean-reinstall,
// --restore-config): the stderr reporting those paths have today survives
// byte-identically.
func TestRestoreMoaiConfig_LegacyWrapperReemitsAdvisoryText(t *testing.T) {
	root := seedRestoreTree(t, "design.yaml",
		"evolution:\n  max_rate: 3\n",
		"evolution:\n  max_rate: 3\n  max_active_learnings: 200\n",
		"evolution:\n  max_rate: 3\n")
	backupDir := filepath.Join(root, "backup")

	sink := withSwappedSink(t)
	if err := RestoreMoaiConfig(root, backupDir, nil); err != nil {
		t.Fatalf("RestoreMoaiConfig: %v", err)
	}
	if !strings.Contains(sink.String(), "advisory: retained key") {
		t.Errorf("legacy wrapper must re-emit the advisory text, got:\n%s", sink.String())
	}
	if !strings.Contains(sink.String(), "evolution.max_active_learnings") {
		t.Errorf("legacy wrapper advisory must name the retained key, got:\n%s", sink.String())
	}
}

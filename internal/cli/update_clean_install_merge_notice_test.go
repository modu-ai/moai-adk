// Package cli — update_clean_install_merge_notice_test.go
//
// Regression test for issue #1243: the clean-reinstall path passed a nil
// MergeFallbackRecorder to backup.RestoreMoaiConfig, so a 3-way merge that
// fell back to the 2-way merge produced NO signal on any verbosity level —
// even though `moai update --help` documents `--verbose` as emitting "3-way
// merge fallback notices". The user's config was silently rewritten from
// template defaults with no way to notice.
//
// The observable the test asserts is the merge-history ledger
// (.moai/cache/merge-history.json), which recordMergeFallback writes on every
// fallback. With a nil recorder the ledger is never created; with the recorder
// wired the fallback is counted. This makes the assertion falsifiable without
// capturing os.Stderr.

package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// makeMergeFallbackFixture builds a v2 project whose backed-up user.yaml is
// syntactically invalid YAML (tab indentation is illegal in YAML). That forces
// MergeYAML3Way to fail during restore, which is exactly the fallback branch
// that must be reported.
func makeMergeFallbackFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// v2 signals so runCleanReinstall activates.
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired\n")

	// Invalid YAML (tab indentation) -> 3-way merge fails -> fallback fires.
	writeTestFile(t, root, ".moai/config/sections/user.yaml",
		"user:\n\tname: GOOS-CUSTOM-NAME\n")

	writeTestFile(t, root, ".claude/settings.json", `{"model": "opus"}`+"\n")
	return root
}

// TestCleanReinstall_MergeFallbackIsRecorded asserts the clean-reinstall path
// routes 3-way merge fallbacks through the same noise-suppression ledger the
// normal `moai update` path uses. Regression guard for issue #1243: a nil
// recorder here made the documented --verbose fallback notice structurally
// unreachable on this path.
func TestCleanReinstall_MergeFallbackIsRecorded(t *testing.T) {
	root := makeMergeFallbackFixture(t)

	deployer := &overwritingDeployer{}
	migrate := &stubMigrateRunner{}
	if _, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}

	ledgerPath := filepath.Join(root, filepath.FromSlash(mergeHistoryLedgerRelPath))
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("merge-history ledger not written at %s: %v\n"+
			"the clean-reinstall path is not reporting 3-way merge fallbacks "+
			"(issue #1243: nil MergeFallbackRecorder)", mergeHistoryLedgerRelPath, err)
	}

	ledger := map[string]mergeHistoryEntry{}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatalf("unmarshal merge-history ledger: %v (content: %s)", err, data)
	}
	entry, ok := ledger["user.yaml"]
	if !ok {
		t.Fatalf("merge-history ledger has no entry for user.yaml; entries: %v", ledger)
	}
	if entry.FallbackCount < 1 {
		t.Errorf("user.yaml fallback_count = %d; want >= 1 (the fallback happened but was not counted)",
			entry.FallbackCount)
	}
}

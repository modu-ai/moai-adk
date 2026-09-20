package backup

// mcp_snapshot_test.go pins the staging/canonical data model of the .mcp.json
// base snapshot (card t1029), the sibling of the .claude/settings.json snapshot
// t656 shipped: where the two copies live, when a copy is valid, that recording
// only happens when the deploy actually wrote the file, and that a .mcp.json
// failure never reads as a settings.json failure.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

const (
	mcpCanonicalRel = ".moai/cache/template-snapshot/mcp.json"
	mcpPendingRel   = ".moai/cache/template-snapshot/mcp.json.pending"
	mcpLiveRel      = ".mcp.json"
)

// deployMCPRender simulates a deploy that wrote render to .mcp.json: the file
// lands at the live path and the manifest records it as template-managed with
// the render hash, exactly as template.deployer does after a real write.
func deployMCPRender(t *testing.T, root string, m manifest.Manager, render string) {
	t.Helper()
	writeFileRel(t, root, mcpLiveRel, render)
	if err := m.Track(mcpLiveRel, manifest.TemplateManaged, manifest.HashBytes([]byte(render))); err != nil {
		t.Fatalf("track: %v", err)
	}
}

// The two copies sit under the shared cache root as their own sibling
// namespace, next to sections/ and claude/ — never inside either.
func TestMCPSnapshotPaths(t *testing.T) {
	root := t.TempDir()
	if got, want := MCPSnapshotPath(root), filepath.Join(root, filepath.FromSlash(mcpCanonicalRel)); got != want {
		t.Errorf("MCPSnapshotPath = %s, want %s", got, want)
	}
	if got, want := MCPSnapshotPendingPath(root), filepath.Join(root, filepath.FromSlash(mcpPendingRel)); got != want {
		t.Errorf("MCPSnapshotPendingPath = %s, want %s", got, want)
	}
	// The cache must never hold a file named .mcp.json — a dot-prefixed copy
	// is the shape a tool scanning for MCP config could discover, which is the
	// same constraint that made the settings sibling drop its leading dot.
	if strings.Contains(MCPSnapshotPath(root), string(filepath.Separator)+".mcp.json") {
		t.Errorf("MCPSnapshotPath keeps the leading dot: %s", MCPSnapshotPath(root))
	}
}

// A canonical copy is usable only when it exists, reads, and decodes as a JSON
// object; anything else keeps the derived base.
func TestLoadMCPSnapshot_Validity(t *testing.T) {
	cases := []struct {
		name  string
		plant func(t *testing.T, root string)
		want  bool
	}{
		{"object", func(t *testing.T, root string) { writeFileRel(t, root, mcpCanonicalRel, `{"a":1}`) }, true},
		{"absent", func(*testing.T, string) {}, false},
		{"directory", func(t *testing.T, root string) {
			if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(mcpCanonicalRel)), 0o755); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"invalid_json", func(t *testing.T, root string) { writeFileRel(t, root, mcpCanonicalRel, `{"a":`) }, false},
		{"array", func(t *testing.T, root string) { writeFileRel(t, root, mcpCanonicalRel, `[1]`) }, false},
		{"null", func(t *testing.T, root string) { writeFileRel(t, root, mcpCanonicalRel, `null`) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.plant(t, root)
			data, ok := LoadMCPSnapshot(root)
			if ok != tc.want {
				t.Fatalf("LoadMCPSnapshot ok = %v, want %v", ok, tc.want)
			}
			if ok && string(data) != `{"a":1}` {
				t.Errorf("LoadMCPSnapshot data = %s", data)
			}
		})
	}
}

// Staging is gated on the deployer's own manifest record, never on the file
// alone: a deploy that skipped an existing user file records nothing.
func TestStageDeployedMCPSnapshot_ManifestGate(t *testing.T) {
	t.Run("deploy_wrote_it", func(t *testing.T) {
		root := t.TempDir()
		m := loadedManager(t, root)
		deployMCPRender(t, root, m, `{"mcpServers":{"a":1}}`)
		StageDeployedMCPSnapshot(root, m, nil)
		assertFileRel(t, root, mcpPendingRel, `{"mcpServers":{"a":1}}`)
	})
	t.Run("user_created_file", func(t *testing.T) {
		root := t.TempDir()
		m := loadedManager(t, root)
		writeFileRel(t, root, mcpLiveRel, `{"mcpServers":{"mine":1}}`)
		if err := m.Track(mcpLiveRel, manifest.UserCreated, ""); err != nil {
			t.Fatalf("track: %v", err)
		}
		StageDeployedMCPSnapshot(root, m, nil)
		assertAbsentRel(t, root, mcpPendingRel)
	})
	t.Run("hash_mismatch_after_rewrite", func(t *testing.T) {
		root := t.TempDir()
		m := loadedManager(t, root)
		deployMCPRender(t, root, m, `{"mcpServers":{"a":1}}`)
		// A later rewrite of the live file must not be recorded as template
		// content — the hash no longer matches what the deployer wrote.
		writeFileRel(t, root, mcpLiveRel, `{"mcpServers":{"a":2}}`)
		StageDeployedMCPSnapshot(root, m, nil)
		assertAbsentRel(t, root, mcpPendingRel)
	})
	t.Run("no_manifest", func(t *testing.T) {
		root := t.TempDir()
		StageDeployedMCPSnapshot(root, nil, nil)
		assertAbsentRel(t, root, mcpPendingRel)
	})
}

// Settle promotes the staging copy unless the flow's merge preserved the user's
// file, in which case the live file does not reflect the render and the staging
// copy is discarded.
func TestSettleMCPSnapshot(t *testing.T) {
	t.Run("promote", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		writeFileRel(t, root, mcpPendingRel, `{"new":true}`)
		SettleMCPSnapshot(root, false, nil)
		assertFileRel(t, root, mcpCanonicalRel, `{"new":true}`)
		assertAbsentRel(t, root, mcpPendingRel)
	})
	t.Run("discard_when_preserved", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		writeFileRel(t, root, mcpPendingRel, `{"new":true}`)
		SettleMCPSnapshot(root, true, nil)
		assertFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		assertAbsentRel(t, root, mcpPendingRel)
	})
	t.Run("no_staging_copy_is_a_noop", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		SettleMCPSnapshot(root, false, nil)
		assertFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
	})
}

// A staging copy an interrupted flow left behind is promoted only when the live
// file still equals it; any intervening write discards it.
func TestJudgeLeftoverMCPSnapshot(t *testing.T) {
	t.Run("live_matches_leftover_promotes", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, mcpLiveRel, `{"new":true}`)
		writeFileRel(t, root, mcpPendingRel, `{"new":true}`)
		JudgeLeftoverMCPSnapshot(root, nil)
		assertFileRel(t, root, mcpCanonicalRel, `{"new":true}`)
		assertAbsentRel(t, root, mcpPendingRel)
	})
	t.Run("live_reverted_discards", func(t *testing.T) {
		root := t.TempDir()
		writeFileRel(t, root, mcpLiveRel, `{"reverted":true}`)
		writeFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		writeFileRel(t, root, mcpPendingRel, `{"new":true}`)
		JudgeLeftoverMCPSnapshot(root, nil)
		assertFileRel(t, root, mcpCanonicalRel, `{"old":true}`)
		assertAbsentRel(t, root, mcpPendingRel)
	})
}

// A .mcp.json snapshot failure must never read as a settings.json failure. The
// two prefixes are the only thing separating them in the operator's terminal,
// which is the same reason sections and settings were split.
func TestMCPSnapshotWarningPrefixesAreDistinct(t *testing.T) {
	pairs := [][2]string{
		{MCPSnapshotWriteFailedPrefix, SettingsSnapshotWriteFailedPrefix},
		{MCPSnapshotPromoteFailedPrefix, SettingsSnapshotPromoteFailedPrefix},
	}
	for _, p := range pairs {
		if p[0] == p[1] {
			t.Errorf("mcp and settings share the prefix %q", p[0])
		}
		if p[0] == "" {
			t.Error("mcp prefix is empty")
		}
	}
	if MCPSnapshotWriteFailedPrefix == MCPSnapshotPromoteFailedPrefix {
		t.Error("mcp write and promote prefixes are identical")
	}
}

// A promotion failure warns once with the .mcp.json prefix and does not block.
func TestSettleMCPSnapshot_PromoteFailureWarnsWithMCPPrefix(t *testing.T) {
	root := t.TempDir()
	writeFileRel(t, root, mcpPendingRel, `{"new":true}`)
	// A directory at the canonical path makes the rename fail.
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(mcpCanonicalRel)), 0o755); err != nil {
		t.Fatalf("mkdir canonical-as-dir: %v", err)
	}
	var warn strings.Builder
	SettleMCPSnapshot(root, false, &warn)

	got := warn.String()
	if !strings.HasPrefix(got, MCPSnapshotPromoteFailedPrefix) {
		t.Errorf("warning = %q, want prefix %q", got, MCPSnapshotPromoteFailedPrefix)
	}
	if strings.Contains(got, SettingsSnapshotPromoteFailedPrefix) {
		t.Errorf("warning reads as a settings failure: %q", got)
	}
	if n := strings.Count(got, "\n"); n != 1 {
		t.Errorf("warning line count = %d, want 1: %q", n, got)
	}
}

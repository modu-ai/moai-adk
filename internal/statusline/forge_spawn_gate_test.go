package statusline

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The spawn gate (SPEC-STATUSLINE-PROFILE-RESPECT-001 REQ-001/REQ-002). Both
// opt-out levers already suppressed the RENDER of the forge pair; these tests
// pin that they also stop the detached refresh child — the polling, not the
// pixels, was the rate-limit motivation of the operator's 2026-08-17 order.
//
// Assertions are made against the githubSpawnProbe counter (acceptance §D,
// kickoff decision D3), never against a real process: under `go test` the
// isSelfInvocable guard blocks the exec anyway, so the counter is the only
// instrument that can distinguish "gated before the spawn" from "spawn blocked
// by the guard".
//
// Tests in this file are deliberately NOT parallel: they swap a package-level
// probe in and out.

// countSpawns installs githubSpawnProbe for the test's lifetime and returns a
// function reporting how many spawn attempts reached the probe.
func countSpawns(t *testing.T) func() int {
	t.Helper()
	var n int
	githubSpawnProbe = func(string) { n++ }
	t.Cleanup(func() { githubSpawnProbe = nil })
	return func() int { return n }
}

// TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing (AC-002, REQ-002):
// an explicit `forge:` value naming no forge resolves Suppressed from the config
// read alone and never reaches the spawn point — regardless of cache state,
// because a child could only re-derive the suppression the config already
// states.
func TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing(t *testing.T) {
	spawns := countSpawns(t)

	tests := []struct {
		name  string
		forge string
		cache bool // seed a stale cache alongside the override?
	}{
		{name: "forge none, absent cache", forge: "none"},
		{name: "forge off, absent cache", forge: "off"},
		{name: "forge none, stale cache still spawns nothing", forge: "none", cache: true},
		{name: "a typo names no forge either", forge: "githbu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeForgeOverride(t, root, tt.forge)
			if tt.cache {
				seedStaleCache(t, root, 7, 3)
			}

			maybeRefreshGitHubCounts(root)

			if got := resolveGitHubCounts(root); !got.Suppressed {
				t.Errorf("counts = %+v, want Suppressed from the config read alone", got)
			}
			if n := spawns(); n != 0 {
				t.Errorf("spawn attempts = %d, want 0 — an explicit no-forge override must gate the refresh child", n)
			}
		})
	}
}

// TestMaybeRefreshGitHubCounts_UnsetOverrideStillSpawns pins the negative space
// of REQ-002: with NO explicit override the auto-detect path stays with the
// child, which owns the `git remote` cost. Gating on a child-written Suppressed
// verdict instead would latch the opt-out semantics onto detection and break
// the not-a-latch contract (installing the CLI must restore the pair).
func TestMaybeRefreshGitHubCounts_UnsetOverrideStillSpawns(t *testing.T) {
	spawns := countSpawns(t)

	root := t.TempDir()
	// No statusline.yaml. The cache carries a CHILD-written suppression verdict
	// (no forge recognised) and is stale — exactly the state where a verdict
	// conflation would wrongly gate.
	stale := GitHubCounts{Suppressed: true, FetchedAt: time.Now().Add(-time.Hour).Unix()}
	if err := writeGitHubCache(root, stale); err != nil {
		t.Fatal(err)
	}

	maybeRefreshGitHubCounts(root)

	if n := spawns(); n != 1 {
		t.Errorf("spawn attempts = %d, want 1 — an unset override must leave detection to the child", n)
	}
}

// TestMaybeRefreshGitHubCounts_NoConfigSpawnsOnce (AC-003, REQ-003
// characterization): a stale cache with no opt-out anywhere still asks for one
// refresh. This is the all-enabled fallback the gates must not narrow.
func TestMaybeRefreshGitHubCounts_NoConfigSpawnsOnce(t *testing.T) {
	spawns := countSpawns(t)

	root := t.TempDir()
	seedStaleCache(t, root, 7, 3)

	maybeRefreshGitHubCounts(root)

	if n := spawns(); n != 1 {
		t.Errorf("spawn attempts = %d, want 1 — the no-config fallback must keep refreshing", n)
	}
}

// builderStdin renders one Build pass against the fixture root, wired so the
// repo segment can appear (workspace.repo present) and the board root resolves
// to the fixture (workspace.current_dir). No network is involved: the counts
// come from the cache fixture alone.
func builderStdin(root string) string {
	return `{"session_id":"t293","workspace":{"current_dir":` + quoteJSON(root) +
		`,"repo":{"host":"github.com","owner":"modu-ai","name":"moai-adk"}}}`
}

func quoteJSON(s string) string {
	// Temp-dir paths carry no characters JSON escapes beyond nothing we need to
	// handle; keep the helper honest anyway.
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// fakeGitProvider keeps the Builder test off any real git invocation: the
// repo+branch segment needs branch data to render at all, and a stub provider
// is the cheapest honest source of it.
type fakeGitProvider struct{}

func (fakeGitProvider) CollectGitStatus(_ context.Context) (*GitStatusData, error) {
	return &GitStatusData{Branch: "main", Available: true}, nil
}

// buildOnce runs a Builder render over stdinJSON and returns the statusline.
func buildOnce(t *testing.T, segmentConfig map[string]bool, stdinJSON string) string {
	t.Helper()
	b := New(Options{
		GitProvider:   fakeGitProvider{},
		RootDir:       t.TempDir(), // no repo to auto-open; the stub provides git data
		NoColor:       true,
		SegmentConfig: segmentConfig,
	})
	out, err := b.Build(context.Background(), strings.NewReader(stdinJSON))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return out
}

// TestBuilder_SegmentGateSuppressesPairAndSpawn (AC-001, REQ-001): with
// `segments.github: false` the Builder renders no forge pair AND asks for no
// refresh child. The paired enabled case proves the test bites: the same
// fixture with the segment on reaches the spawn point exactly once.
func TestBuilder_SegmentGateSuppressesPairAndSpawn(t *testing.T) {
	spawns := countSpawns(t)

	root := t.TempDir()
	seedStaleCache(t, root, 7, 3)
	stdin := builderStdin(root)

	gated := buildOnce(t, map[string]bool{SegmentGitHub: false}, stdin)
	if n := spawns(); n != 0 {
		t.Errorf("gated render spawn attempts = %d, want 0 (REQ-001: the segment gate must reach the spawn, not just the render)", n)
	}
	if strings.Contains(gated, "7/3") || strings.Contains(gated, "-/-") {
		t.Errorf("gated render %q must not carry a forge pair", gated)
	}

	enabled := buildOnce(t, map[string]bool{SegmentGitHub: true}, stdin)
	if n := spawns(); n != 1 {
		t.Errorf("enabled render spawn attempts = %d, want 1 — the gate must be the only delta", n)
	}
	if !strings.Contains(enabled, "7/3") {
		t.Errorf("enabled render %q must serve the cached pair", enabled)
	}
}

// TestBuilder_NilSegmentsPreservesSpawn (AC-003, REQ-003): a nil segments map
// is the all-enabled default; the Builder still refreshes. Characterization —
// no regression from the M2 gate.
func TestBuilder_NilSegmentsPreservesSpawn(t *testing.T) {
	spawns := countSpawns(t)

	root := t.TempDir()
	seedStaleCache(t, root, 7, 3)

	_ = buildOnce(t, nil, builderStdin(root))

	if n := spawns(); n != 1 {
		t.Errorf("spawn attempts = %d, want 1 — nil segments must keep the all-enabled fallback", n)
	}
}

// TestForgeOptOut_TwoWayRevert (AC-004, REQ-004): the paired regression the
// lead mandated. Removing the opt-out (`none` → `github`) brings the pair back
// within one refresh cycle with NO cache deletion — the operator flips a config
// key, not the state directory.
func TestForgeOptOut_TwoWayRevert(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub forge CLIs are exercised on unix; windows covered by GOOS=windows build/vet")
	}
	spawns := countSpawns(t)

	root := t.TempDir()
	writeForgeOverride(t, root, "none")

	// The opted-out steady state: suppressed from config, no child asked for.
	maybeRefreshGitHubCounts(root)
	if n := spawns(); n != 0 {
		t.Fatalf("pre-revert spawn attempts = %d, want 0", n)
	}
	got := resolveGitHubCounts(root)
	if !got.Suppressed {
		t.Fatalf("pre-revert counts = %+v, want Suppressed", got)
	}

	// The revert: rewrite the SAME file to a recognised forge. No cache
	// deletion — the cache file is never removed, only re-written by the child.
	writeForgeOverride(t, root, "github")
	cachePath := githubCachePath(root)
	if _, err := os.Stat(cachePath); err == nil {
		t.Fatal("precondition: this fixture must run from an absent cache to observe creation-in-place; got one")
	}

	stubDir := writeForgeStub(t, "gh", 42, 17, false)
	t.Setenv("PATH", stubDir+":"+os.Getenv("PATH"))

	if err := RefreshGitHubCounts(context.Background(), root); err != nil {
		t.Fatalf("RefreshGitHubCounts after revert: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache must exist after the refresh without any deletion: %v", err)
	}

	// One cycle later the render path serves the fetched pair and the spawn
	// gate is open again.
	after := resolveGitHubCounts(root)
	if after.Suppressed {
		t.Fatalf("post-revert counts = %+v, want Suppressed cleared", after)
	}
	if after.OpenIssues != 42 || after.OpenPRs != 17 {
		t.Errorf("post-revert counts = %d/%d, want 42/17", after.OpenIssues, after.OpenPRs)
	}
	maybeRefreshGitHubCounts(root)
	if n := spawns(); n != 0 {
		t.Errorf("post-revert fresh-cache spawn attempts = %d, want 0 (fresh cache, not gate)", n)
	}

	// And with a stale cache the spawn gate is open again — one TTL later the
	// pair refreshes on its own, no cache deletion involved.
	seedStaleCache(t, root, 42, 17)
	maybeRefreshGitHubCounts(root)
	if n := spawns(); n != 1 {
		t.Errorf("post-revert stale-cache spawn attempts = %d, want 1 — the reverted override must not stay gated", n)
	}
}

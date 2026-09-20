package config_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// wantSettingsAxisCleanup is the canonical 14-key settings-axis cleanup view in
// its declared order (SPEC-GLM-CLEANUP-SSOT-001 REQ-1). It is spelled here as a
// literal on purpose: the test's job is to pin what the production declaration
// says, so a future addition to that declaration shows up as a deliberate diff
// on this list rather than sliding in unobserved.
var wantSettingsAxisCleanup = []string{
	"MOAI_BACKUP_AUTH_TOKEN",
	"ANTHROPIC_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_FABLE_MODEL",
	"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
	"API_TIMEOUT_MS",
	"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
	"CLAUDE_CODE_TEAMMATE_DISPLAY",
	"MOAI_STATUSLINE_CONTEXT_SIZE",
	"CLAUDE_CODE_AUTO_COMPACT_WINDOW",
	"CLAUDE_CODE_MAX_CONTEXT_TOKENS",
}

// wantSettingsAxisLegacyTail is the 5-key legacy tail: the members of the
// cleanup view that no live producer can write today. Cleanup of these keys is
// legacy-residue cleanup of a file an older binary wrote.
var wantSettingsAxisLegacyTail = []string{
	"ANTHROPIC_DEFAULT_FABLE_MODEL",
	"API_TIMEOUT_MS",
	"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
	"CLAUDE_CODE_TEAMMATE_DISPLAY",
	"MOAI_STATUSLINE_CONTEXT_SIZE",
}

// wantSettingsAxisLive is the 9-key live view: the keys a live producer
// (ensureGLMCredentials) can put into settings.local.json's env, plus the
// OAuth-backup key the injection path parks a pre-existing token in.
var wantSettingsAxisLive = []string{
	"ANTHROPIC_AUTH_TOKEN",
	"MOAI_BACKUP_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
	"CLAUDE_CODE_AUTO_COMPACT_WINDOW",
	"CLAUDE_CODE_MAX_CONTEXT_TOKENS",
}

func sliceToSet(keys []string) map[string]int {
	set := make(map[string]int, len(keys))
	for _, k := range keys {
		set[k]++
	}
	return set
}

// assertSameMembership compares two key slices as multisets and reports the
// exact asymmetric difference, so a failure names the drifted key rather than
// only the cardinality.
func assertSameMembership(t *testing.T, viewName string, got, want []string) {
	t.Helper()

	gotSet, wantSet := sliceToSet(got), sliceToSet(want)
	for k, n := range gotSet {
		if wantSet[k] != n {
			t.Errorf("%s: unexpected key %q (got %d occurrence(s), want %d)", viewName, k, n, wantSet[k])
		}
	}
	for k, n := range wantSet {
		if gotSet[k] != n {
			t.Errorf("%s: missing key %q (got %d occurrence(s), want %d)", viewName, k, gotSet[k], n)
		}
	}
}

// TestSettingsAxisCleanupViewMembership pins the cleanup view's exact
// membership, its cardinality (14), and its declared order.
func TestSettingsAxisCleanupViewMembership(t *testing.T) {
	t.Parallel()

	got := config.SettingsAxisCleanupKeys()

	// Emptiness guard: an accessor that returned nothing would make every
	// membership comparison below pass vacuously on the "missing key" side only
	// if the want list were also empty — this guard makes the vacuous case an
	// explicit, named failure instead.
	if len(got) == 0 {
		t.Fatal("SettingsAxisCleanupKeys() returned no keys; a settings-axis view can never be empty")
	}

	if len(got) != 14 {
		t.Errorf("SettingsAxisCleanupKeys() cardinality = %d, want 14: %v", len(got), got)
	}

	assertSameMembership(t, "cleanup view", got, wantSettingsAxisCleanup)

	// Stable order: the declaration is ordered so that a future addition is a
	// readable one-line diff rather than a reshuffle.
	for i := range wantSettingsAxisCleanup {
		if i >= len(got) {
			t.Errorf("cleanup view: position %d missing, want %q", i, wantSettingsAxisCleanup[i])
			continue
		}
		if got[i] != wantSettingsAxisCleanup[i] {
			t.Errorf("cleanup view: position %d = %q, want %q", i, got[i], wantSettingsAxisCleanup[i])
		}
	}
}

// TestSettingsAxisLiveViewMembership pins the live view's exact membership and
// its cardinality (9).
func TestSettingsAxisLiveViewMembership(t *testing.T) {
	t.Parallel()

	got := config.SettingsAxisLiveKeys()

	if len(got) == 0 {
		t.Fatal("SettingsAxisLiveKeys() returned no keys; a settings-axis view can never be empty")
	}

	if len(got) != 9 {
		t.Errorf("SettingsAxisLiveKeys() cardinality = %d, want 9: %v", len(got), got)
	}

	assertSameMembership(t, "live view", got, wantSettingsAxisLive)
}

// TestSettingsAxisViewsShareOneDeclaration pins the set relation that makes the
// two views derivations of ONE declaration rather than two literals kept in step
// by hand: cleanup == live ∪ legacy-tail, the tail being exactly 5 keys, and the
// live view appearing inside the cleanup view in the same relative order.
func TestSettingsAxisViewsShareOneDeclaration(t *testing.T) {
	t.Parallel()

	cleanup := config.SettingsAxisCleanupKeys()
	live := config.SettingsAxisLiveKeys()

	if len(cleanup) == 0 || len(live) == 0 {
		t.Fatalf("empty view(s): cleanup=%d live=%d; the set relation is not measurable", len(cleanup), len(live))
	}

	liveSet := sliceToSet(live)

	// Derive the tail from the two views, then compare it against the expected
	// 5 keys. Deriving rather than re-listing is what makes this a check of the
	// relation and not a third literal.
	var gotTail []string
	for _, k := range cleanup {
		if liveSet[k] == 0 {
			gotTail = append(gotTail, k)
		}
	}

	if len(gotTail) != 5 {
		t.Errorf("legacy tail cardinality = %d, want 5: %v", len(gotTail), gotTail)
	}
	assertSameMembership(t, "legacy tail", gotTail, wantSettingsAxisLegacyTail)

	// Every live key is a cleanup key (live ⊆ cleanup).
	cleanupSet := sliceToSet(cleanup)
	for _, k := range live {
		if cleanupSet[k] == 0 {
			t.Errorf("live key %q is absent from the cleanup view; the views do not derive from one declaration", k)
		}
	}

	if len(live)+len(gotTail) != len(cleanup) {
		t.Errorf("live(%d) + legacy tail(%d) != cleanup(%d)", len(live), len(gotTail), len(cleanup))
	}

	// The live view preserves the declaration's relative order, which is what a
	// filter over one ordered declaration produces. A hand-written second
	// literal would not be required to.
	var filtered []string
	for _, k := range cleanup {
		if liveSet[k] > 0 {
			filtered = append(filtered, k)
		}
	}
	for i := range filtered {
		if i >= len(live) {
			t.Errorf("live view: position %d missing, want %q", i, filtered[i])
			continue
		}
		if live[i] != filtered[i] {
			t.Errorf("live view: position %d = %q, want %q (declaration order)", i, live[i], filtered[i])
		}
	}
}

// TestSettingsAxisViewsExcludeUserOwnedStandIn pins that CUSTOM_VAR — the
// stand-in key the cleanup guards use for "a key the user owns" — is a member of
// NEITHER view. The preservation assertions elsewhere rest on that fact, so it
// is asserted here rather than assumed there.
func TestSettingsAxisViewsExcludeUserOwnedStandIn(t *testing.T) {
	t.Parallel()

	const standIn = "CUSTOM_VAR"

	cleanup := config.SettingsAxisCleanupKeys()
	live := config.SettingsAxisLiveKeys()

	if len(cleanup) == 0 || len(live) == 0 {
		t.Fatalf("empty view(s): cleanup=%d live=%d; absence of %q is not measurable", len(cleanup), len(live), standIn)
	}

	if sliceToSet(cleanup)[standIn] != 0 {
		t.Errorf("%q must not be a member of the cleanup view", standIn)
	}
	if sliceToSet(live)[standIn] != 0 {
		t.Errorf("%q must not be a member of the live view", standIn)
	}
}

// TestGLMEnvVarSetUnchangedBySettingsAxis pins that the settings-axis
// declaration was added as a SIBLING of GLMEnvVarSet rather than by widening it.
// GLMEnvVarSet answers a different question (the 3-key inject↔clear parity
// anchor of SPEC-CLIFIX-HYGIENE-001 REQ-HYG-001-003).
func TestGLMEnvVarSetUnchangedBySettingsAxis(t *testing.T) {
	t.Parallel()

	want := []string{
		config.EnvClaudeCodeDisableExperimentalBetas,
		config.EnvClaudeCodeDisableNonessentialTraffic,
		config.EnvClaudeCodeTeammateDisplay,
	}

	got := config.GLMEnvVarSet()
	if len(got) != 3 {
		t.Fatalf("GLMEnvVarSet() cardinality = %d, want 3: %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GLMEnvVarSet() position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestSettingsAxisViewsAreNotAliasedToCallers pins that a caller mutating a
// returned slice cannot corrupt the declaration for the next caller — the same
// obligation GLMEnvVarSet's doc comment states, made mechanical here.
func TestSettingsAxisViewsAreNotAliasedToCallers(t *testing.T) {
	t.Parallel()

	first := config.SettingsAxisCleanupKeys()
	if len(first) == 0 {
		t.Fatal("SettingsAxisCleanupKeys() returned no keys; aliasing is not measurable")
	}
	original := first[0]
	first[0] = "MUTATED_BY_CALLER"

	if second := config.SettingsAxisCleanupKeys(); second[0] != original {
		t.Errorf("cleanup view is aliased: after a caller mutation, position 0 = %q, want %q", second[0], original)
	}

	firstLive := config.SettingsAxisLiveKeys()
	if len(firstLive) == 0 {
		t.Fatal("SettingsAxisLiveKeys() returned no keys; aliasing is not measurable")
	}
	originalLive := firstLive[0]
	firstLive[0] = "MUTATED_BY_CALLER"

	if secondLive := config.SettingsAxisLiveKeys(); secondLive[0] != originalLive {
		t.Errorf("live view is aliased: after a caller mutation, position 0 = %q, want %q", secondLive[0], originalLive)
	}
}

package permission

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveConflict_SpecificityWins verifies that the more specific pattern wins.
// Related to T-RT002-22 and AC-12.
func TestResolveConflict_SpecificityWins(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		{
			Pattern: "Bash(git push:*)",      // less specific -> deny.
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a-settings.json",
		},
		{
			Pattern: "Bash(git push origin main)", // more specific -> allow.
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "b-settings.json",
		},
	}

	winner := resolveConflict(rules, "Bash", "git push origin main")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	if winner.Action != DecisionAllow {
		t.Errorf("resolveConflict() winner.Action = %v, want Allow (more specific pattern)", winner.Action)
	}
	if winner.Pattern != "Bash(git push origin main)" {
		t.Errorf("resolveConflict() winner.Pattern = %q, want 'Bash(git push origin main)'", winner.Pattern)
	}
}

// TestResolveConflict_FsOrderTiebreak verifies fs-order (later Origin) wins when specificity ties.
// Related to T-RT002-22 and AC-12.
func TestResolveConflict_FsOrderTiebreak(t *testing.T) {
	t.Parallel()

	// Same pattern specificity, different Origin.
	rules := []*PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a-settings.json", // earlier (lexicographically first).
		},
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "z-settings.json", // later (lexicographically last) -> wins.
		},
	}

	winner := resolveConflict(rules, "Bash", "curl https://example.com")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	// fs-order later -> the allow rule from z-settings.json wins.
	if winner.Origin != "z-settings.json" {
		t.Errorf("resolveConflict() winner.Origin = %q, want 'z-settings.json' (fs-order tiebreak)", winner.Origin)
	}
}

// TestResolveConflict_SingleMatchNoLog verifies single-match returns without a conflict log.
// Related to T-RT002-22.
func TestResolveConflict_SingleMatchNoLog(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		{
			Pattern: "Bash(go test:*)",
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "settings.json",
		},
	}

	winner := resolveConflict(rules, "Bash", "go test ./...")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil for single rule")
	}
	if winner.Action != DecisionAllow {
		t.Errorf("resolveConflict() winner.Action = %v, want Allow", winner.Action)
	}
}

// TestResolveConflict_LogPath verifies that conflict resolution completes without error.
// Related to T-RT002-22 — logConflict invocation completes without panic or error.
func TestResolveConflict_LogPath(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		{
			Pattern: "Bash(rm:*)",
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a.json",
		},
		{
			Pattern: "Bash(rm /tmp:*)",
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "b.json",
		},
	}

	// Completes without panic, including the logConflict invocation.
	winner := resolveConflict(rules, "Bash", "rm /tmp/test.txt")
	if winner == nil {
		t.Fatal("resolveConflict() should not return nil for 2 rules")
	}
}

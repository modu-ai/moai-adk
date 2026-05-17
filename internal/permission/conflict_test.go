package permission

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestResolveConflict_SpecificityWins 더 구체적인 패턴이 우선함을 검증한다.
// T-RT002-22, AC-12 관련.
func TestResolveConflict_SpecificityWins(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		{
			Pattern: "Bash(git push:*)",      // 덜 구체적 → deny.
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a-settings.json",
		},
		{
			Pattern: "Bash(git push origin main)", // 더 구체적 → allow.
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

// TestResolveConflict_FsOrderTiebreak specificity 동점 시 fs-order (Origin 나중) 우선 검증.
// T-RT002-22, AC-12 관련.
func TestResolveConflict_FsOrderTiebreak(t *testing.T) {
	t.Parallel()

	// 같은 패턴 specificity, 다른 Origin.
	rules := []*PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a-settings.json", // 먼저 (lexicographic 앞).
		},
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "z-settings.json", // 나중 (lexicographic 뒤) → 우선.
		},
	}

	winner := resolveConflict(rules, "Bash", "curl https://example.com")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	// fs-order 나중 → z-settings.json 의 allow 가 우선.
	if winner.Origin != "z-settings.json" {
		t.Errorf("resolveConflict() winner.Origin = %q, want 'z-settings.json' (fs-order tiebreak)", winner.Origin)
	}
}

// TestResolveConflict_SingleMatchNoLog 단일 매칭 시 충돌 로그 없이 반환 검증.
// T-RT002-22 관련.
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

// TestResolveConflict_LogPath 충돌 발생 시 오류 없이 완료됨을 검증한다.
// T-RT002-22 관련 — logConflict 호출이 패닉/오류 없이 완료.
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

	// logConflict 호출 포함하여 패닉 없이 완료.
	winner := resolveConflict(rules, "Bash", "rm /tmp/test.txt")
	if winner == nil {
		t.Fatal("resolveConflict() should not return nil for 2 rules")
	}
}

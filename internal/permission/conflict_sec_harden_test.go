package permission

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-SEC-HARDEN-001 §M2 — Permission conflict resolution: deny wins on tie + audit log written.
//
// reproduction-first 계약:
//   - AC-SEC-M2-001 (RED): 픽스 전 equal-specificity allow+deny tie가 Origin-order로
//     ALLOW를 선택함을 입증 (deny가 이기지 않음).
//   - AC-SEC-M2-002 (GREEN): 픽스 후 equal-specificity tie에서 deny가 이긴다.
//   - AC-SEC-M2-003/004 (NO-REG): all-allow tie는 Origin order 보존; 높은 specificity는
//     action 무관하게 이긴다 (deny-precedence는 동일 specificity tie에서만 적용).
//   - AC-SEC-M2-005 (RED): 픽스 전 conflict log 미기록.
//   - AC-SEC-M2-006 (GREEN): 픽스 후 conflict log 기록.
//   - AC-SEC-M2-007 (NO-REG): unwritable log dir여도 결정 불변 (best-effort).

// withConflictLogDir 는 conflictLogDir 패키지 변수를 임시로 교체하고 복원하는 헬퍼다.
// 테스트는 실제 프로젝트 트리에 쓰지 않고 t.TempDir() 루트만 사용한다 (CLAUDE.local §6).
func withConflictLogDir(t *testing.T, dir string) {
	t.Helper()
	prev := conflictLogDir
	conflictLogDir = dir
	t.Cleanup(func() { conflictLogDir = prev })
}

// TestResolveConflict_DenyWinsOnTie 는 AC-SEC-M2-002 (GREEN) 다.
// equal-specificity allow + deny tie는 Origin 순서와 무관하게 deny가 이긴다.
func TestResolveConflict_DenyWinsOnTie(t *testing.T) {
	t.Parallel()

	// allow의 Origin이 lexicographically 더 늦음(z) → 픽스 전이라면 Origin tiebreak로 allow가 이겼음.
	rules := []*PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionDeny,
			Source:  config.SrcLocal,
			Origin:  "a-settings.json", // 더 이른 Origin
		},
		{
			Pattern: "Bash(curl:*)",
			Action:  DecisionAllow,
			Source:  config.SrcLocal,
			Origin:  "z-settings.json", // 더 늦은 Origin (픽스 전엔 이김)
		},
	}

	winner := resolveConflict(rules, "Bash", "curl https://example.com")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	if winner.Action != DecisionDeny {
		t.Errorf("resolveConflict() winner.Action = %v, want Deny (deny wins on equal-specificity tie)", winner.Action)
	}
	if winner.Origin != "a-settings.json" {
		t.Errorf("resolveConflict() winner.Origin = %q, want 'a-settings.json' (the deny rule)", winner.Origin)
	}
}

// TestResolveConflict_DenyWinsOnTie_OrderIndependent 는 deny-precedence가 슬라이스 순서에
// 무관함을 확인한다 (deny가 먼저 오든 나중에 오든 deny가 이긴다).
func TestResolveConflict_DenyWinsOnTie_OrderIndependent(t *testing.T) {
	t.Parallel()

	// allow 먼저, deny 나중 (Origin도 deny가 더 늦음 → 어느 tiebreak로도 deny여야 함을 약화하지 않도록
	// allow Origin을 더 늦게 두어 deny-precedence가 진짜로 동작함을 입증)
	rules := []*PermissionRule{
		{Pattern: "Bash(rm:*)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "z-allow.json"},
		{Pattern: "Bash(rm:*)", Action: DecisionDeny, Source: config.SrcLocal, Origin: "a-deny.json"},
	}
	winner := resolveConflict(rules, "Bash", "rm -rf /tmp/x")
	if winner == nil || winner.Action != DecisionDeny {
		t.Fatalf("resolveConflict() = %v, want Deny winner regardless of slice/Origin order", winner)
	}
}

// TestResolveConflict_AllAllowTiePreservesOrigin 는 AC-SEC-M2-003 (NO-REG) 다.
// deny가 없는 all-allow equal-specificity tie는 기존 Origin-order tiebreak를 보존한다.
func TestResolveConflict_AllAllowTiePreservesOrigin(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		{Pattern: "Bash(curl:*)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "a-settings.json"},
		{Pattern: "Bash(curl:*)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "z-settings.json"},
	}
	winner := resolveConflict(rules, "Bash", "curl https://example.com")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	if winner.Origin != "z-settings.json" {
		t.Errorf("resolveConflict() winner.Origin = %q, want 'z-settings.json' (Origin-order tiebreak preserved for all-allow)", winner.Origin)
	}
}

// TestResolveConflict_HigherSpecificityWinsRegardlessOfAction 는 AC-SEC-M2-004 (NO-REG) 다.
// deny-precedence는 동일 specificity tie에서만 적용된다. 더 높은 specificity를 가진 allow는
// 더 낮은 specificity의 deny를 이긴다.
func TestResolveConflict_HigherSpecificityWinsRegardlessOfAction(t *testing.T) {
	t.Parallel()

	rules := []*PermissionRule{
		// 낮은 specificity deny (wildcard 많음)
		{Pattern: "Bash(git push:*)", Action: DecisionDeny, Source: config.SrcLocal, Origin: "a.json"},
		// 높은 specificity allow (정확 매칭, wildcard 없음 + 더 긺)
		{Pattern: "Bash(git push origin main)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "b.json"},
	}
	winner := resolveConflict(rules, "Bash", "git push origin main")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil")
	}
	if winner.Action != DecisionAllow {
		t.Errorf("resolveConflict() winner.Action = %v, want Allow (higher specificity wins regardless of action)", winner.Action)
	}
	if winner.Pattern != "Bash(git push origin main)" {
		t.Errorf("resolveConflict() winner.Pattern = %q, want the higher-specificity allow", winner.Pattern)
	}
}

// TestLogConflict_WritesAuditRecord 는 AC-SEC-M2-005 (RED) + AC-SEC-M2-006 (GREEN) 다.
// 픽스 전: permission.log가 기록되지 않음 (RED). 픽스 후: conflict record가 append됨 (GREEN).
func TestLogConflict_WritesAuditRecord(t *testing.T) {
	dir := t.TempDir()
	withConflictLogDir(t, dir)

	rules := []*PermissionRule{
		{Pattern: "Bash(curl:*)", Action: DecisionDeny, Source: config.SrcLocal, Origin: "a-settings.json"},
		{Pattern: "Bash(curl:*)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "z-settings.json"},
	}

	// resolveConflict 내부에서 logConflict가 호출된다.
	_ = resolveConflict(rules, "Bash", "curl https://example.com")

	logPath := filepath.Join(dir, "permission.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("conflict log not written at %s: %v (AC-SEC-M2-006)", logPath, err)
	}
	content := string(data)
	// 기록된 record에는 candidate origin과 action이 포함되어야 한다.
	if !strings.Contains(content, "a-settings.json") || !strings.Contains(content, "z-settings.json") {
		t.Errorf("conflict log missing candidate origins; got: %q", content)
	}
	if !strings.Contains(content, string(DecisionDeny)) || !strings.Contains(content, string(DecisionAllow)) {
		t.Errorf("conflict log missing candidate actions; got: %q", content)
	}
}

// TestLogConflict_UnwritableDirDoesNotChangeDecision 는 AC-SEC-M2-007 (NO-REG) 다.
// 로그 디렉토리에 쓸 수 없어도 결정은 불변이며 에러가 caller로 surface되지 않는다.
func TestLogConflict_UnwritableDirDoesNotChangeDecision(t *testing.T) {
	// 존재하는 파일을 디렉토리 경로로 지정하여 MkdirAll/OpenFile이 실패하도록 유도한다.
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "not-a-dir")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// conflictLogDir를 파일 하위 경로로 설정 → MkdirAll 실패.
	withConflictLogDir(t, filepath.Join(blocker, "logs"))

	rules := []*PermissionRule{
		{Pattern: "Bash(curl:*)", Action: DecisionDeny, Source: config.SrcLocal, Origin: "a-settings.json"},
		{Pattern: "Bash(curl:*)", Action: DecisionAllow, Source: config.SrcLocal, Origin: "z-settings.json"},
	}

	// 쓰기 불가 로그 경로에서도 deny-precedence 결정은 동일해야 하고 panic/에러가 없어야 한다.
	winner := resolveConflict(rules, "Bash", "curl https://example.com")
	if winner == nil {
		t.Fatal("resolveConflict() returned nil under unwritable log dir")
	}
	if winner.Action != DecisionDeny {
		t.Errorf("resolveConflict() winner.Action = %v, want Deny (decision unaffected by log write failure)", winner.Action)
	}
}

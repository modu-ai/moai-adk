// SPEC-STOP-EVIDENCE-GATE-001: session-ledger reader (IMP-03) + evidence
// evaluator (IMP-02 판단 로직). 기존 telemetry.LoadBySession read 경로 위에
// read-only 로 구축되며, 새 원장 저장소 파일을 만들지 않는다 (REQ-SEG-001).
package hook

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// runEvidenceGate 는 이미 적재 가능한 세션 원장을 읽어 "성공 주장 + 관측된 이진
// 증거 없음" 조합을 advisory 로 표면화한다. Stop 이벤트를 절대 차단하지 않으며
// (fail-open, REQ-SEG-005), 모든 출력은 stderr 로 간다 (REQ-SEG-006). error 는
// swallow 한다 (pruning/reflection 과 동일한 best-effort).
//
// telemetry.LoadBySession 만 호출한다 — 새 I/O / network / test 재실행 없음
// (REQ-SEG-008, ≤5s). AskUserQuestion / mcp__askuser 미호출 (REQ-SEG-007, C-HRA-008).
//
// 현재 텔레메트리 스트림에 대해 knowingly-dormant 다: 유일한 production writer
// logSkillUsage 가 Outcome=unknown / Phase="none" 를 하드코딩하므로 본 게이트는
// 현재 발화하지 않는다. 가치는 후속 SPEC-STOP-EVIDENCE-WRITER-001 에 blocked (spec.md §A.3).
func runEvidenceGate(projectRoot, sessionID string) {
	if projectRoot == "" {
		return // 기존 pruning/reflection 과 동일 가드
	}
	records, err := telemetry.LoadBySession(projectRoot, sessionID) // REUSE — REQ-SEG-001
	if err != nil || len(records) == 0 {
		return // 평가할 것 없음; fail-open
	}
	finding := evaluateEvidence(buildSessionLedger(records))
	if finding == nil {
		return
	}
	// advisory: slog.Warn (stderr 라우팅) + human-readable 한 줄. stdout 미접촉.
	slog.Warn("evidence-gate: unbacked success claim", finding.slogArgs()...)
	fmt.Fprintln(os.Stderr, finding.HumanReadable())
}

// @MX:ANCHOR: [AUTO] 세션 원장 판단 로직의 단일 진입점 — runEvidenceGate 와 단위 테스트가 호출
// @MX:REASON: [AUTO] fan_in>=3 (buildSessionLedger → inferPathKind → evaluateEvidence), 게이트 정확성의 핵심
// SessionLedger 는 한 세션의 텔레메트리 레코드에 대한 read-only in-memory 뷰다.
// telemetry.LoadBySession 결과로부터 구축되며 어떤 저장소 파일도 생성·기록하지
// 않는다 (REQ-SEG-001).
type SessionLedger struct {
	SessionID     string                  // 세션 식별자
	Records       []telemetry.UsageRecord // 원본 레코드 (LoadBySession 결과)
	PathKind      string                  // "docs-only" | "code-change" | "unknown"
	SuccessClaims int                     // Outcome in {success, partial} 카운트
	BinaryPass    bool                    // 어느 레코드든 IsTestPass == true
	BinaryFail    bool                    // 어느 레코드든 IsTestFail == true
}

// buildSessionLedger 는 LoadBySession 이 반환한 레코드 슬라이스로부터 세션 원장을
// 집계한다. 새 I/O 없음 — 입력 슬라이스만 in-memory 로 순회한다 (REQ-SEG-008).
func buildSessionLedger(records []telemetry.UsageRecord) SessionLedger {
	ledger := SessionLedger{
		Records:  records,
		PathKind: inferPathKind(records),
	}
	for _, r := range records {
		if ledger.SessionID == "" && r.SessionID != "" {
			ledger.SessionID = r.SessionID
		}
		if r.Outcome == telemetry.OutcomeSuccess || r.Outcome == telemetry.OutcomePartial {
			ledger.SuccessClaims++
		}
		if r.IsTestPass {
			ledger.BinaryPass = true
		}
		if r.IsTestFail {
			ledger.BinaryFail = true
		}
	}
	return ledger
}

// inferPathKind 는 세션을 고정 taxonomy 의 한 path-kind 버킷으로 분류한다
// (design.md §2.2, REQ-SEG-011).
//
//  1. 명시 PathKind 우선 (Approach A — 신규 omitempty 필드).
//  2. 추론 fallback (Approach B — 레거시 레코드): Phase/AgentType 신호.
//     혼합 세션은 code-change 신호가 하나라도 있으면 code-change 우선 (보수적 —
//     코드 변경이 있으면 증거를 요구한다).
//  3. 추론 불가 → unknown (REQ-SEG-010 conservative fallback).
func inferPathKind(records []telemetry.UsageRecord) string {
	// (1) explicit signal first.
	for _, r := range records {
		if r.PathKind != "" {
			return r.PathKind
		}
	}

	// (2) inference fallback. code-change 신호가 docs 신호보다 우선한다 (보수적).
	hasCodeSignal := false
	hasDocsSignal := false
	for _, r := range records {
		switch {
		case r.Phase == "run" || r.Phase == "plan" || r.AgentType == "manager-develop":
			hasCodeSignal = true
		case r.Phase == "sync" || r.AgentType == "manager-docs":
			hasDocsSignal = true
		}
	}
	if hasCodeSignal {
		return telemetry.PathKindCodeChange
	}
	if hasDocsSignal {
		return telemetry.PathKindDocsOnly
	}

	// (3) ambiguous / absent.
	return telemetry.PathKindUnknown
}

// Finding 은 게이트가 표면화하는 advisory(권고) 발견 사항이다. 차단하지 않으며
// stderr/slog 로만 출력된다 (REQ-SEG-005, REQ-SEG-006).
type Finding struct {
	SessionID     string
	PathKind      string
	SuccessClaims int
	BinaryPass    bool
	BinaryFail    bool
}

// evaluateEvidence 는 세션 원장으로부터 "성공 주장 + 관측된 이진 증거 없음" 조합을
// 판단한다. 부수효과 없는 pure 함수다 (design.md §0.4).
//
// 판단 규칙:
//   - docs-only → nil          (REQ-SEG-003: 문서 작업 면제)
//   - unknown   → nil          (REQ-SEG-010: 보수적, 추론 불가)
//   - success claim 없음 → nil  (주장이 없으면 unbacked-claim 도 없음)
//   - success claim + 이진 신호 관측 불가 → nil (REQ-SEG-010: 부재 ≠ 실패)
//   - success claim + 이진 신호 관측 + pass 없음 → Finding (표적 결함 형태)
//   - 그 외 (pass 관측됨) → nil  (성공이 관측된 pass 로 뒷받침됨)
func evaluateEvidence(ledger SessionLedger) *Finding {
	switch ledger.PathKind {
	case telemetry.PathKindDocsOnly, telemetry.PathKindUnknown:
		return nil
	}

	if ledger.SuccessClaims == 0 {
		return nil
	}

	binaryObservable := ledger.BinaryPass || ledger.BinaryFail
	if !binaryObservable {
		// 이진 신호 부재는 "관측 불가"이지 "검증 실패"가 아니다 (REQ-SEG-010).
		return nil
	}

	if ledger.BinaryPass {
		// pass 가 하나라도 관측되면 backed 로 처리 (edge case #6, 일부 통과 = 검증 관측).
		return nil
	}

	// success claim + 이진 신호 관측됨 + pass 없음 → advisory finding.
	return &Finding{
		SessionID:     ledger.SessionID,
		PathKind:      ledger.PathKind,
		SuccessClaims: ledger.SuccessClaims,
		BinaryPass:    ledger.BinaryPass,
		BinaryFail:    ledger.BinaryFail,
	}
}

// HumanReadable 은 advisory finding 의 사람이 읽을 수 있는 한 줄 요약을 반환한다
// (design.md §0.5). stderr 로 출력된다.
func (f *Finding) HumanReadable() string {
	return fmt.Sprintf(
		"[evidence-gate] session %s claimed success without observed binary evidence "+
			"(path-kind=%s, success claims=%d, binary pass observed=%t, binary fail observed=%t); "+
			"verify the claimed completion was actually observed",
		f.SessionID, f.PathKind, f.SuccessClaims, f.BinaryPass, f.BinaryFail,
	)
}

// slogArgs 는 slog.Warn 에 넘길 구조화된 key/value 인자 슬라이스를 반환한다
// (design.md §0.5). 짝수 길이의 key, value, key, value, ... 형태다.
func (f *Finding) slogArgs() []any {
	return []any{
		"session_id", f.SessionID,
		"path_kind", f.PathKind,
		"success_claims", f.SuccessClaims,
		"binary_pass_observed", f.BinaryPass,
		"binary_fail_observed", f.BinaryFail,
	}
}

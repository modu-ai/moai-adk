package session

import "fmt"

// HydrateForPrompt는 프롬프트 조립에 필요한 세 컨텍스트 조각을 고정된 순서로 반환합니다.
//
// cache-prefix: DO NOT REORDER
// (systemPrompt, userContext, systemContext) 순서는 Anthropic 캐시 키의 구성 요소입니다.
// 순서를 바꾸면 캐시 미스가 발생하여 추론 비용이 증가합니다.
// 새 SPEC 없이 절대로 순서를 변경하지 마세요.
//
// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-004 cache-prefix discipline (P-C05 closure)
// @MX:REASON: (systemPrompt, userContext, systemContext) assembly order is frozen here.
// DO NOT REORDER without a new SPEC — reordering causes Anthropic cache misses.
func HydrateForPrompt(phase Phase, specID string, store SessionStore) (systemPrompt, userContext, systemContext string, err error) {
	// cache-prefix: DO NOT REORDER — 반환 순서를 변경하면 캐시 키가 달라집니다.
	state, err := store.Hydrate(phase, specID)
	if err != nil {
		return "", "", "", fmt.Errorf("hydrate phase state: %w", err)
	}
	if state == nil {
		// checkpoint가 없는 경우 빈 문자열 반환 (오케스트레이터가 새 세션으로 처리)
		return "", "", "", nil
	}

	// 1. systemPrompt: checkpoint 메타데이터 (phase, specID, 상태 요약)
	systemPrompt = fmt.Sprintf("phase=%s spec=%s status=%s", state.Phase, state.SPECID, checkpointStatus(state))

	// 2. userContext: SPECID 기반 사용자 컨텍스트
	userContext = fmt.Sprintf("spec_id=%s", state.SPECID)

	// 3. systemContext: provenance 정보
	systemContext = fmt.Sprintf("provenance_source=%s origin=%s", state.Provenance.Source, state.Provenance.Origin)

	return systemPrompt, userContext, systemContext, nil
}

// checkpointStatus는 checkpoint에서 간단한 상태 문자열을 추출합니다.
func checkpointStatus(state *PhaseState) string {
	if state.Checkpoint == nil {
		return "no_checkpoint"
	}
	switch cp := state.Checkpoint.(type) {
	case *PlanCheckpoint:
		return cp.Status
	case *RunCheckpoint:
		return cp.Status
	case *SyncCheckpoint:
		// SyncCheckpoint는 Status 필드가 없음 — docs_synced 기반 상태 반환
		if cp.DocsSynced {
			return "synced"
		}
		return "pending"
	default:
		return "unknown"
	}
}

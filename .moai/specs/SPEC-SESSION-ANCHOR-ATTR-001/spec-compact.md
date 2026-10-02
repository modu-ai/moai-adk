# spec-compact.md — SPEC-SESSION-ANCHOR-ATTR-001

## REQ 요약 (GEARS, 14건)

- REQ-SAA-001 [W1] `WorktreeGuardRefusal` 로그 행에 session_id + 해상 cwd + 거부 인용 트리 경로 포함
- REQ-SAA-002 [W1] 해상 불가 시 `unknown` 마커로 행 유지 (유실 금지)
- REQ-SAA-003 [W2] 모든 `RelocateSession`에 감사 행(session, from, to, trigger hook, timestamp)
- REQ-SAA-004 [W2] 목표 트리 lock이 다른 생존 카드 지시 시 advisory 플래그 + 진행
- REQ-SAA-005 [W2] opt-in `workflow.anchor_relocation_guard.enabled` 시 플래그 재배치 거부 + 기록
- REQ-SAA-006 [불변] 두-패스 재배치 후보 순서 + fail-open 불변 (REQ-RAR-002/003/004)
- REQ-SAA-007 [W3] `MOAI_ANCHOR_TRACE=1` 시 의사결정별 상세 로그(session_id+pid+cwd+timestamp)
- REQ-SAA-008 [W3] 미게이트 시 출력 0, 오버헤드 env 조회 1회
- REQ-SAA-009 [W3] env 상수는 envkeys.go 선언
- REQ-SAA-010 [불변] `worktreeGuardAnchor` 문구 + 살생 변이 2건 불변
- REQ-SAA-011 [불변] `LiveAnchoredSessions` 단일 소스화 금지, 처분 가드 시맨틱 불변 (R1 게이트)
- REQ-SAA-012 [불변] Seam A 분리 유지 (input.CWD git-context / CLAUDE_PROJECT_DIR 감사 로그)
- REQ-SAA-013 [불변] 계측기는 실행 지점에만(훅 사본 금지), `RefuseMutationFromNonCanonicalTree` 불변
- REQ-SAA-014 [W4] worktree-integration-ops.md 4요소 문서화 (복구 절차·런타임 경계·t741 무음 위험·정정 사고 기록)

## AC 요약

- AC-001/002: W1 행 필드 + unknown 마커 (실측 선택자: `TestClassifyError_GuardRefusal|TestFormatMessage_GuardRefusal` — 8 RUN 관측)
- AC-003/004/005: W2 감사 행 + REQ-SAA-004 소유 규칙 4케이스(self/other-live/unreadable/registry-only) + opt-in blocking 두 팔 (선택자: `TestRelocateSession` — 3 RUN 관측)
- AC-006: 재배치 순서·fail-open 회귀 게이트 (기존 스위트 무수정 GREEN)
- AC-007/008: W3 트레이스 on/off (AC-008은 AC-007 양성 대조 필수)
- AC-009: 불변식 회귀 (변이 살생·Seam A·env 상수·LiveAnchoredSessions 6 RUN·RefuseMutation 1 RUN)
- AC-010: W4 문서 4요소 + 템플릿-로컬 diff 0 + neutrality audit
- 신규 동작 AC는 §D.0 채택표(RED-now/green-path) 두-셀 규칙 적용; AC-006·009는 regression-guard; `[no tests to run]` 출력은 실패로 처리하는 빈-스윕 가드 전 AC 공통

## Files to Modify

- `[MODIFY]` internal/hook/post_tool_failure.go
- `[MODIFY]` internal/hook/failure_observer.go
- `[MODIFY]` internal/session/anchor.go
- `[MODIFY]` internal/hook/cwd_changed_relocate.go
- `[MODIFY]` internal/hook/branch_guard.go (트레이스 계측점)
- `[MODIFY]` internal/config/defaults.go (`workflow.anchor_relocation_guard.enabled`)
- `[MODIFY]` internal/config/envkeys.go (`MOAI_ANCHOR_TRACE` 상수)
- `[MODIFY]` internal/template/templates/.claude/rules/moai/workflow/worktree-integration-ops.md (+ 로컬 동기)
- `[MODIFY]` internal/hook/worktree_guard_refusal_test.go (행 필드 단정 추가 — 앵커 문구 불변)
- `[EXISTING]` internal/session/anchor_lock.go — 동작 불변

## Exclusions

- Claude Code 런타임 앵커 내부 (수리·단정 모두 제외 — per-agent anchor registration 도크트린 거절)
- 처분 가드 시맨틱 변경 / `LiveAnchoredSessions` 단일 소스화 (R1 게이트)
- 세션 종료 시점 메모리 갱신 (run-phase 후속)
- `worktreeGuardAnchor` 거부 문구 변경
- stale 레지스트리 항목 청소 (t1369 sweep 소관)

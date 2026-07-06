# Progress — SPEC-HANDOFF-THRESHOLD-001

> Lifecycle 3-phase: plan → run → sync. phase별 audit-ready signal + evidence 누적. §E.1은 plan-phase(manager-spec), §E.2/§E.3은 run-phase(manager-develop), §E.4는 sync-phase(manager-docs) 소관.

## §E.1 Plan-phase Audit-Ready Signal

- 작성일: 2026-07-06 (Epic Handoff-v2 M4/4, 마지막 마일스톤)
- 산출물 (Tier M 3-artifact + design/research + progress skeleton = 5 + 1):
  - `spec.md` — frontmatter 12-field + tier:M/era:V3R6/related_specs, GEARS REQ-THRESHOLD-001..018, Out-of-Scope h3 6개
  - `plan.md` — M1(config+stage)/M2(state-file)/M3(독트린+화해) milestone-split + 6 blocker 해소(4 C-axis + D1/D2) + AC 바인딩, Out-of-Scope h3
  - `acceptance.md` — AC-THRESHOLD-001..018 (REQ 1:1), edge cases, quality gate, DoD, 잔여 위험(concurrent empty-id 포함)
  - `design.md` — stage enum + 하드 상한 공식(clamp) + state-file 스키마(+writer_pid) + Guide/Mode 화해 + 호출부 + §D.4a/§D.4b concurrent-empty-id
  - `research.md` — autoCompactThreshold 위치/호출부/atomic 선례/template mirror drift 실측
- SPEC ID self-check: `decomposition: SPEC ✓ | HANDOFF ✓ | THRESHOLD ✓ | 001 ✓ → PASS`
- 2 LOCKED 결정 준수: 기존 HandoffConfig 필드만 소비 / 밴드 경계 defaults.go 상수 하드코딩
- 6 blocker 해소 (4 C-axis + plan-auditor iter-1 D1/D2):
  - B1 하드 상한 unreachable → `min(95, getAutoCompactThreshold()+10)` + `hard<soft` clamp + reachability 문서(REQ-005). autoCompactThreshold=memory.go 동일 패키지(open question 아님)
  - B2 write 호출부 → `builder.Build`(collectAll 직후, session_id+Memory 동시 스코프)
  - B3 session_id guard → last-writer-wins 스탬프 + reader 불일치 stale
  - B4 fallback-UUID → session_id 부재 시 captured_at freshness(single-session 생존)
  - **D1 (iter-1 SHOULD-FIX) template drift** → template mirror 256K 행 부재(실측 `grep -c`=0) vs LIVE 존재(=1). D4가 template에 256K 행 ADD(parity, REQ-017/AC-017) + Detection 절 section-level 편집(BOTH), full-sync 금지(LIVE 256K 삭제 회귀 방지). AC-016 `<doctrine>`=LIVE 명시.
  - **D2 (iter-1 SHOULD-FIX) concurrent empty-id hole** → B4 empty-id fallback이 B3 guard 재개방(UUID-less 2+ concurrent). `writer_pid` discriminator(REQ-018/AC-018) Go 헬퍼 레벨 기계 차단. 잔여: 독트린-only reader 미비교(보수적 폴백), 후속 Go reader가 완전 폐쇄. Tier M 비례(trigger 드묾).
- 실측 정정: Detection 독트린 **template mirror 존재**(`internal/template/templates/...`) → task 전제("template 밖") 오기. 256K 행 **LIVE만 존재(=1), template mirror 부재(=0)** → M1 drift, D4가 parity 회복.
- M1 무회귀 불변식 = AC-THRESHOLD-006 (statusline suffix config 무관, default guide=false에서 soft 유지)
- plan-auditor iter-1: PASS-WITH-DEBT 0.84 (4 MUST-PASS pass, no BLOCKING); D1/D2 SHOULD-FIX 반영 → REQ/AC 16→18
- plan_status: **audit-ready** (plan-auditor Tier M 임계 0.80; iter-1 0.84 ≥ 0.80, D1/D2 정정 완료)
- plan_complete_at: 2026-07-06

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs>_

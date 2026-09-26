# Progress — SPEC-HANDOFF-NEUTRAL-001 (card t1273)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-26T20:10:00+09:00
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L 5)
plan_commit: (이 커밋에 동반)
baseline: worktree t1273, branch WT-handoff-neutral, develop 분기점 1b7a88d78

## §F Phase 4 Mode Selection

- tier: L / scope: internal/cli·internal/codexadapter·internal/worktree (신규 파일 2-3 + 수정 3-4) / domain count: 3 (CLI·어댑터·워크트리) / language mix: Go + Markdown / concurrency benefit: LOW (직렬 의존: RED→show→시딩→LIVE→판정)
- direct: not selected — 다중 파일·다중 도메인
- fanout: not selected — 코딩 중심(Anthropic coding-task parallelism caveat), 단계 간 의존
- sweep: not selected — ~30 파일 미만·비기계적
- agent-team: not selected — 명시 요청 없음
- **Decision: serial**
- Justification: 구현 체인이 직렬(RED 관측 → show → 시딩 → LIVE → P1 판정)이고 각 단계가 앞 단계 산출을 소비. 단일 manager-develop 위임이 동기 부여·의존 관리 모두에서 단순.

## 진행 기록

- 2026-09-26 research.md — 공식 문서 4건 검증(§F), codexadapter 갭 발견, 쿼터 관측(F-7). 커밋 36326286f·f6c96d862·fb51c8a59.
- 2026-09-26 design.md — P3 기본+P1 조건부(전제 2개: 워크트리 시딩·관문 b), D2.5 옵션 A(materializer 시딩+런처 보완, 리드 조정 반영). 커밋 cc463138a.
- 2026-09-26 spec.md·acceptance.md·plan.md 작성 — plan-audit 대기.
- 2026-09-26 **plan-audit 1차 FAIL 0.75** (임계 0.85, `.moai/reports/t1273/plan-audit.md`). 방향(P3+조건부 P1·LIVE 관문·codexadapter 갭·워크트리 부재)은 감사가 소스 재현으로 확인; 결함 5건은 전부 명세 수준 — 수리: ① 저장모델 부정합(pending.json→factory.db `resume_handoffs` 상태머신 — SavePending이 pending.json을 삭제함을 pending.go:84-113·homestate/factory.go:70·handoff.go:211-220 실측 정정) ② 시딩 재사용 소재(`update_codex_wiring.go` "creates nothing" → `internal/codexwiring` wire.go) ③ REQ-HN-011 미커버(AC-HN-011 문서화 AC 신설) ④ AC-HN-010 허위 셀렉터(실존 TestHandoffSave_*/TestHandoffClear/TestHandoffCmdRegistered로 교체) ⑤ red 셀 SHA 계약(첫 줄 HEAD SHA 기록). acceptance fixture 계약을 DB 세팅으로 전환(공허 초록 방지). 재심사는 결함 델타 범위로 요청.

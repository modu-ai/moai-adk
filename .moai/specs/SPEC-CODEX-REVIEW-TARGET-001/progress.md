# SPEC-CODEX-REVIEW-TARGET-001 — 진행 기록

카드 **t399** · 이슈 **modu-ai/moai-adk#1632** · 브랜치 `WT-codex-native-branch`

## §E.1 Plan-phase Audit-Ready Signal

- 산출: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 산출물 집합 + progress).
- Tier: **M**. 근거는 AC 예산 — AC 는 iter1 수리 후 11건(001·002·003·004·005·006·006b·007·008·009·010)으로 Tier S 상한(8)을 초과한다. REQ 는 8건으로 Tier M 상한(16) 안. LOC·파일 수만 보면 S 로도 읽히므로 이 분류는 예산 기준의 판단이다.
- 측정 원천: `.moai/reports/t399/discovery.md`, `.moai/reports/t399/schema/v2/ReviewStartParams.json`, 트리 `442da4f06`.
- 카드 전제 정정 1건: "`coerceCodexReviewTarget` 을 덮는 테스트가 없다" → 행동 기준으로는 거짓. 리프트는 이미 검사되고 있으며 검사 지점이 하필 리프트가 옳은 유일한 variant 다(spec.md §A.4).
- 축 2(백엔드 참여 노출)는 t284 소관으로 확정 — 축의 존재·처분은 `.moai/reports/t229/succession.md`, 카드 번호는 큐 판독(spec.md §E).

### plan-audit iter1 → iter2 수리 기록

- iter1 판정: **FAIL 0.75** (Tier M 임계 0.80). must-pass 7/7 통과. 보고서 `.moai/reports/t399/plan-audit.md`.
- blocking 6건(D1~D6) + optional 4건(D7~D10) **전부 반영**. 각 결함의 처리는 spec.md HISTORY v0.2.0 에 항목별로 기록.
- 감사자가 지적한 중심 결함은 "자기가 세운 재사용 규율을 자기 검증 표면에는 적용하지 않았다" — AC-CRT-010(라이브 왕복)이 그 자리를 닫는다.
- 감사자의 3대 실측 주장은 이 세션이 **독립적으로 재확인**했다: 라이브 `review/start` 선례 존재(`codex_live_protocol_probe_test.go:507`), skip 3조건 확립본(`codex_review_gate_live_test.go:33`), `toolErr` 의 `IsError: true` 조기 반환(`mcp_server.go:868`). D2 의 트리 우연 일치(`worktree_base_branch: develop` ↔ `origin/HEAD → origin/develop`)도 직접 측정.
- **운영자 결정 대기 1건**: plan.md §B 의 (가)/(나) — M1 에서 고정. 권장은 (가).
- **운영자 검토 권유 1건**: D2 를 **결정 (a) 정렬**로 닫았다(spec.md §A.7). `worktree_base_branch` 를 기준으로 삼아야 한다면 그것은 GLM 경로까지 함께 바꾸는 별도 카드다.

### 이 세션이 관측하지 않은 것

- 어떤 테스트도 실행하지 않았다. AC-CRT-003 의 "변경 전 초록" 전제는 미확인이며 M1 에서 확인한다.
- 라이브 codex 왕복을 시도하지 않았다. AC-CRT-010 의 RED 예측은 스키마 판독 기반 추론이며, 관측은 M2b 의 몫이다.
- 감사 보고서의 점수·차원 판정은 재계산하지 않았다. 재확인한 것은 그 판정이 근거로 든 파일 사실들이다.

_run-phase 진입 전. 아래 §E.2 ~ §E.4 는 자리표시._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

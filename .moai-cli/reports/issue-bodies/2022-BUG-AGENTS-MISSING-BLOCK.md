<!-- moai-template:moai-bugfix@0.3.0; 2026-11-08 -->
```yaml
logical_key: BUG-AGENTS-MISSING-BLOCK
milestone: 2026.11.0-rc.1
parent_key: EP-02
work_kind: bugfix
related: [01-B, 03-C]
claim: {state: unclaimed}
evidence_status: not_run
```
# [RC1][결함] `agents-plan`이 관리 구간이 삭제된 AGENTS.md를 `unchanged`로 판정한다

## 재현(2026-11-08, 리더 실측)
1. `AGENTS.md`에서 `<!-- moai:managed:begin id=moai-workflow -->…end` 블록을 통째로 지운다(사용자 삭제 상황).
2. `echo '{"protocol_version":1,"language":"en","workflow_reference":"skills/moai/workflows/","local_file":"tracked"}' | moai-cli agents-plan .`
3. 기대: 삭제를 감지해 `conflict`(detail: "managed block missing or removed")로 보고하거나, 명시 재채택 요청이 있을 때만 `append`. 실제: `AGENTS.md unchanged` — 삭제가 조용히 무시되고 apply도 `state: unchanged`.

## 원인(코드 읽기)
`src/agents.rs` `decide(baseline, current, desired)`가 `desired == baseline`이면 무조건 `KeepCurrent`를 반환한다. 블록이 없을 때 `current=None`이지만 렌더 텍스트가 baseline과 같으면 이 분기에 먼저 걸려 `section_plan`의 "missing block" 분기(`Decision::Conflict` + detail)에 도달하지 못한다. 기존 fixture는 "렌더 변경 없음 + 블록 존재" 경우만 다룬다.

## 수정 방향(이 Issue 범위)
- `section_plan`: `present && block_digest.is_none() && baseline.is_some()` → `action: "conflict"`, detail `"managed block removed after baseline; user deletion preserved (re-adopt requires explicit request)"`. `decide()` 자체는 건드리지 않는다(01-B #1989가 apply.rs로 이관 예정 — 소유 Issue에 통지).
- fixture 1개 추가: 블록 삭제 후 plan이 conflict, apply가 파일을 바꾸지 않음.
- `docs/rust-product-path.md` AGENTS 절에 "삭제된 블록은 복원하지 않고 conflict로 보고" 한 줄.

## 인수 기준
- AC-BUG-1: 위 재현 시 `agents-plan` 출력 action == `conflict`, detail 포함 · 검증: `… cargo test --offline --locked --test product_cli agents_plan_reports_removed_block_as_conflict` · 결과: not_run
- AC-BUG-2: 기존 15+ agents fixture 전부 통과(동작 보존) · 검증: `… cargo test --offline --locked` · 결과: not_run
- AC-BUG-3: 독립 AUDIT PASS → 에이전틱 게이트로 통합 · 결과: not_run

## 범위
- 변경 가능: `src/agents.rs`(section_plan 분기 1곳), `tests/product_cli.rs`, `docs/rust-product-path.md`
- 변경 금지: `decide()` 시그니처·의미, 렌더 텍스트, manifest 스키마

## HOLD
- "재채택(re-adopt)" 명시 요청 플래그(`AgentsRequest.readopt_missing: bool`)를 둘지 · 상태: HOLD — 01-B/03-C에서 결정

상태 규칙: label·assignee는 lock이 아니다. 실행하지 않은 검증은 not_run.

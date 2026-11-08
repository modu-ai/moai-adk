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

## 원인(코드 읽기, r2 정정 — agent-1 지적)
`src/agents.rs` `decide(baseline, current, desired)`가 `desired == baseline`이면 `KeepCurrent`를 반환하므로, 블록이 삭제돼 `current=None`이어도 렌더 텍스트가 baseline과 같으면 `unchanged`가 된다. **이 동작은 의도된 것**이다: 기존 fixture `agents_user_edits_are_preserved_or_reported_as_conflict`(tests/product_cli.rs:433-440)가 "블록 삭제 + 템플릿 동일 → unchanged, apply 성공(사용자 삭제 보존)"을 단언한다(UPDATE 계약 "사용자 삭제 보존"). 결함은 **판정이 아니라 보고**다: 삭제 사실이 plan의 어디에도 나타나지 않아 운영자가 "관리 구간이 없다"는 것을 모른 채 지나간다(리더 실측 2026-11-08).

## 수정 방향(이 Issue 범위, r2)
- action은 `unchanged` 그대로(삭제 보존·apply no-op 유지, `decide()` 불변). `section_plan`에서 `present && block_digest.is_none() && baseline.is_some()`이면 `detail: "managed block absent; user deletion preserved (re-adopt requires an explicit request)"`를 채우고, `AgentsPlan.warnings`에 같은 문구를 1건 추가한다.
- 기존 fixture는 단언을 **확장**(unchanged 유지 + detail·warning 존재); 신규 fixture 불필요하면 생략.
- `docs/rust-product-path.md` AGENTS 절에 "삭제된 블록은 복원하지 않으며 plan detail/warning으로 보고" 한 줄. 재채택 플래그는 HOLD(아래).

## 인수 기준
- AC-BUG-1: 위 재현 시 `agents-plan` 출력 action == `unchanged`이고 `files[].detail`과 `warnings[]`에 "managed block absent" 문구가 있으며 apply는 파일을 바꾸지 않는다 · 검증: `… cargo test --offline --locked --test product_cli agents_user_edits_are_preserved_or_reported_as_conflict`(단언 확장) · 결과: not_run
- AC-BUG-2: 기존 agents fixture 전부 통과(위 단언 확장 외 무수정) · 검증: `… cargo test --offline --locked` · 결과: not_run
- AC-BUG-3: 독립 AUDIT PASS → 에이전틱 게이트로 통합 · 결과: not_run

## 범위
- 변경 가능: `src/agents.rs`(section_plan 분기 1곳), `tests/product_cli.rs`, `docs/rust-product-path.md`
- 변경 금지: `decide()` 시그니처·의미, 렌더 텍스트, manifest 스키마

## HOLD
- "재채택(re-adopt)" 명시 요청 플래그(`AgentsRequest.readopt_missing: bool`)를 둘지 · 상태: HOLD — 01-B/03-C에서 결정

상태 규칙: label·assignee는 lock이 아니다. 실행하지 않은 검증은 not_run.

---
id: SPEC-20261124-model-value-trim
title: "The fast-mode pair check trims the model value before it matches the Opus 4.8 family"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7952 of card C-436 (ref note:436:30859b1a563929e9623b460d21d0bb53, 2026-11-23T22:18:20Z, under the user's delegation of 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 7951 with option (a), approving revision 1 at head 9c1004de04; not approved by the lane", digest: "sha256:f84cb80984e8c3f454a954c9708e6a434d52326089b904e78c352382a6cd519a"}
---

# The fast-mode pair check trims the model value before it matches the Opus 4.8 family

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-436 의 본문과 인수 기준 1, 그리고 C-353 의 `AUDIT #353` 델타 노트 N-a 가 정했어요.

## Intent

Related: SPEC-20261122-model-fast-pair (REQ-3, D-2). 그 SPEC 의 경계 수리(감사 B1, 2026-11-23)가 `is_opus_48` 의 일치를 "끝이거나 `-`, `[`, `@` 가 뒤따를 때"로 좁히면서, 끝 공백이 붙은 `claude-opus-4-8 ` 은 fast mode 키 옆에서도 더는 막히지 않아요(수리 전의 `contains` 는 막았어요). `one_line` 은 끝 공백을 받아들이므로(`src/claude_settings_merge.rs:86`) 그 값은 병합 후보로 남아요. Claude Code 가 모델 값을 다듬는지는 확인되지 않았어요(unverified). 값이 다듬어진다면 이 쌍은 2026-11-30 부터 API 가 거절하는 바로 그 쌍이에요. 그래서 일치 전에 값을 다듬어서 닫아요: 추정이 틀려도 해가 없는 방어적 변경이에요.

원 출처: 카드 C-353 의 `AUDIT #353 (model fast-mode pair delta)` 노트의 비차단 항목 N-a, 후속 카드 C-436.

## Non-goals

- 병합이 내보내는 값(`carry`)은 바꾸지 않아요: 다듬는 것은 일치 판정용 사본뿐이고, 병합 후보의 원문은 그대로예요.
- `one_line` 의 규칙(끝 공백 허용)과 다른 행의 검사는 건드리지 않아요.
- Claude Code 가 값을 다듬는지는 이 SPEC 이 정하지 않아요(unverified 로 남겨요).
- 경계 규칙(D-2 of SPEC-20261122-model-fast-pair)은 그대로예요: `opus-4-80`, `opus-4.85` 는 여전히 다른 모델이에요.

## Requirements

- REQ-1: When a model value beside a fast-mode key has leading or trailing whitespace, the settings merge shall match the Opus 4.8 family on the trimmed value · failure: `claude-opus-4-8 ` beside `fastMode` stays mergeable.
- REQ-2 (REQ-1): IF the trimmed model value names another model, THEN the settings merge shall leave the model row mergeable · failure: a near id such as `claude-opus-4-80 ` beside a fast-mode key is blocked.
- REQ-3: The settings merge shall carry the model value exactly as read · failure: a trimmed copy reaches the `carry` of a mergeable row.

## Acceptance

- AC-1 (REQ-1): a model value with a leading space, a trailing space or both (`" claude-opus-4-8"`, `"claude-opus-4-8 "`, `" opus-4.8 "`, `"Claude-OPUS-4-8[1m] "`) beside `fastMode` and beside `fastModePerSessionOptIn` gets its `model` row `blocked` with the fixed reason and no carried value, written first as a failing test (RED) beside the existing near-id test, and a guard mutant that drops the trim makes it fail · verify: `cargo nextest run --lib -E 'test(fast_pair)'`
- AC-2 (REQ-2): a near id with a trailing space (`"claude-opus-4-80 "`, `"claude-opus-4.85 "`) beside `fastMode` and beside `fastModePerSessionOptIn` stays `mergeable` · verify: `cargo nextest run --lib -E 'test(fast_pair)'`
- AC-3 (REQ-3): the same value with no fast-mode key stays `mergeable` and its `carry` equals the original text with its spaces · verify: `cargo nextest run --lib -E 'test(fast_pair)'`

## Scope

- write: `.moai/specs/SPEC-20261124-model-value-trim/spec.md`
- write: `src/claude_settings_merge.rs` (`is_opus_48` and its unit tests in `mod tests`)
- read-only: `.moai/specs/SPEC-20261122-model-fast-pair/spec.md` (its Result is not edited; this SPEC carries its own)
- forbidden: every other path, `CLAUDE.md` included

## Plan

1. SPEC 초안을 커밋하고 리더 승인을 기다려요(스스로 승인하지 않아요).
2. RED: AC-1 의 값들을 기존 근처 id 테스트 옆의 새 테스트에 넣고 실패를 관찰해요.
3. GREEN: `is_opus_48` 이 `model.trim()` 의 소문자 사본으로 일치를 판정해요.
4. 가드 뮤턴트: trim 을 빼고 같은 테스트가 실패(exit 100)하는지 보고 되돌려요.
5. `moai check affected` 를 맨몸으로 한 번 돌리고 READY 에 영수증을 실어요.

## Risks

- 다듬기가 너무 넓으면(예: 내부 공백까지) 다른 모델을 막을 수 있어요. 앞뒤 공백(`str::trim`)만 다듬고 경계 규칙은 그대로라서, 근처 id 테스트가 그 선을 지켜요.
- 두 번째 가드: AC-2 가 원문이 `carry` 에 그대로 남는지 못박아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 일치 판정에만 `trim()` 을 쓰고 `carry` 와 `one_line` 은 그대로 둬요 | 병합 원문을 바꾸지 않고 막는 범위만 넓히는 가장 작은 변경이에요(C-436 본문) | `is_opus_48` 의 `.trim()` 한 호출을 빼요 |
| D-2 | 별도 SPEC 을 만들어요 | 브랜치 이름 `spec/model-value-trim-20261124` 에 맞는 날짜의 SPEC 이 없으면 훅이 코드 쓰기를 막아요(C-436 배차) | SPEC-20261122-model-fast-pair 의 Result 에 옮겨 적고 이 SPEC 을 닫아요 |

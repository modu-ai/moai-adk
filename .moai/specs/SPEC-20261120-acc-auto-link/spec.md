---
id: SPEC-20261120-acc-auto-link
title: "Agents and tools add @cc: comments and link them automatically"
status: draft
revision: 1
created: 2026-11-20
---

# Agents and tools add @cc: comments and link them automatically

## Intent

사용자 결정 2026-11-20(리더 대화): 2026-11-17 ACC 결정 행의 "에이전트는 자동 생성하지 않는다" 절과 `sdd.md` 규칙 9 의 "Agents never add them on their own" 를 대체해요. 앞으로 에이전트와 도구가 `@cc:` 코멘트를 달고 자동으로 연결해요.

그대로인 것: 이름 ACC, 표식 `@cc:`, 종류 `later`·`careful`·`proves`, 옛 `@MX:`·`@MOAI:` 는 읽기만 해요. 표식의 연결 문법과 종류 확장은 이 SPEC 이 정하지 않아요(ACC 재설계가 정해요).

이 SPEC 은 규칙 문장만 바꿔요: 결정 기록 한 행, 규칙 문서·에이전트 정의·워크플로의 해당 문장, 옛 문장을 못 박은 시험. 도구가 실제로 코멘트를 달고 잇는 코드는 이 SPEC 에 없어요.

## Non-goals

- 표식 문법(연결 방식)·새 종류·옛 태그 이전은 정하지 않아요: ACC 재설계 SPEC 이 정해요.
- `@cc:` 를 달거나 잇는 도구 코드를 만들지 않아요.
- 코드 파일에 `@cc:` 코멘트를 실제로 더하지 않아요.
- `moai-impl` 이 어떤 변경에 어떤 종류를 달지의 기준 표는 만들지 않아요.

## Requirements

- REQ-1: The decisions log shall carry a row dated 2026-11-20 that names the 2026-11-17 ACC row and supersedes its clause that agents do not generate `@cc:` tags.
- REQ-2: The SDD rule on `@cc:` comments shall state that agents and tools add `@cc:` comments and link them automatically.
- REQ-3: The documentation agent definition shall allow adding `@cc:` comments in the code files its scope names.
- REQ-4: The implementation agent definition shall tell the agent to add `@cc:later`, `@cc:careful` and `@cc:proves` comments where its change calls for them.
- REQ-5: The SYNC workflow shall state that SYNC keeps the `@cc:` comments true to the code and adds those the change calls for.
- REQ-6: The chains index and the quality guide shall describe `@cc:` comments as added by agents and tools.
- REQ-7: The rule text shall keep the name ACC, the marker `@cc:`, the kinds `later`, `careful` and `proves` and the read-only rule for the old `@MX:` and `@MOAI:` tags.
- REQ-8: The guidance test shall pin the new wording in the SDD rule, the SYNC workflow and the documentation agent definition and reject the old sentences.

## Acceptance

- AC-1 (REQ-1): 결정 기록에 11-20 행이 있고 11-17 ACC 행을 이름으로 가리키며 그 절을 대체한다고 적어요.
  verify: `cargo test --locked --offline --test sdd_guidance the_decisions_log_supersedes_the_acc_clause`
- AC-2 (REQ-2, REQ-5, REQ-6, REQ-7): `sdd.md` 규칙 9·`sync.md`·`chains/index.md`·`quality.md` 가 새 문장을 말하고 이름·표식·종류·옛 태그 읽기 전용은 그대로예요.
  verify: `cargo test --locked --offline --test sdd_guidance preserve_then_change_and_the_cc_rule_are_stated`
- AC-3 (REQ-3, REQ-4): `agents/docs.md` 는 `@cc:` 추가를 허용하고 `agents/impl.md` 는 달도록 말해요.
  verify: `cargo test --locked --offline --test sdd_guidance the_agents_may_add_cc_comments`
- AC-4 (REQ-8): 옛 세 문장("Agents never add them on their own", "`@cc:` comment lines, never add one", "it never adds new ones")이 패키지·에이전트·규칙 문서 어디에도 없어요.
  verify: `cargo test --locked --offline --test sdd_guidance the_old_cc_sentences_are_gone`

## Scope

- write: `.moai/specs/SPEC-20261120-acc-auto-link/spec.md`, `docs/decisions-2026-11-08.md`(한 행), `skills/moai/references/internal/sdd.md`, `skills/moai/references/internal/quality.md`, `skills/moai/references/chains/index.md`, `skills/moai/workflows/sync.md`, `agents/docs.md`, `agents/impl.md`, `tests/sdd_guidance.rs`, `skills/moai/references/source-manifest.json`
- read-only: `.moai/specs/SPEC-20261117-sdd-guidance/spec.md`(옛 규칙의 출처), `skills/moai/references/internal/sdd.md` 의 나머지 규칙
- forbidden: 표식 문법 정하기, 도구 코드, 코드 파일에 `@cc:` 더하기

## Plan

1. SPEC 을 먼저 커밋해요(이 문서).
2. RED: `tests/sdd_guidance.rs` 의 `preserve_then_change_and_the_cc_rule_are_stated` 를 새 문장으로 바꾸고 세 시험(`the_decisions_log_supersedes_the_acc_clause`, `the_agents_may_add_cc_comments`, `the_old_cc_sentences_are_gone`)을 더해 옛 문장에서 실패하는 것을 눈으로 봐요.
3. GREEN: 결정 기록은 `decision-append`(미리보기 뒤 적용)로 한 행만 더하고 나머지는 해당 문장만 고쳐요. 스킬 파일을 고치므로 `python3 -I tools/source_manifest.py write` 로 매니페스트를 갱신해요.
4. 확인: 위 verify 명령, `sdd_guidance` 전체, `product_cli skill_package`, `tests.source_manifest_test`·`tests.skill_references_test`·`tests.repo_scans_test`, `repo_agents_md`, 에이전트 정의를 읽는 시험(`agent_defs_package`·`update_user_layer_cli`), fmt, clippy `--all-targets`.

## Risks

- "자동으로 연결"하는 도구가 아직 없어요: 규칙 문장은 하려는 동작을 말하고, 도구는 ACC 재설계 SPEC 이 만들어요. 그 전에는 에이전트가 손으로 단 `@cc:` 만 있어요.
- 에이전트가 근거 없이 `@cc:` 를 달아 코드가 어지러워질 수 있어요: 문장은 "변경이 필요로 하는 곳에" 달라고 말하고, 감사가 SPEC·코드·`@cc:` 사이 어긋남을 계속 FAIL 로 봐요(`sdd.md` 규칙 8 은 그대로예요).
- 규칙 파일 다섯 곳과 시험이 같은 문장을 못 박아서 한 곳만 고치면 시험이 깨져요: 같은 커밋에서 함께 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 대체는 결정 기록의 새 행으로 하고 옛 11-17 행은 고치지 않는다 | 기록 규칙: 규칙 변경은 대체하는 기록을 이름으로 밝혀요 | 새 행에 되돌림을 적기 |
| D-2 | 결정 기록 행은 `decision-append` 로 더한다 | 기존 행을 건드리지 않는 도구예요 | 행 지우기는 사람이 |
| D-3 | 추가 시험 하나는 옛 세 문장이 없음을 저장소의 규칙 문서와 에이전트 정의에서 찾아 확인한다 | 한 곳만 남겨 두는 것을 막아요 | 시험 지우기 |

## Result

(구현 뒤에 채워요.)

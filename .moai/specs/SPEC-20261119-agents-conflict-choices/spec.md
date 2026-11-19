---
id: SPEC-20261119-agents-conflict-choices
title: "An AGENTS.md conflict comes with append, skip and a CLAUDE.md import pick"
status: draft
revision: 1
created: 2026-11-19
---

# An AGENTS.md conflict comes with append, skip and a CLAUDE.md import pick

## Intent

카드 C-57의 기준 1·2·3이에요. `/moai project`를 사람이 쓴 `AGENTS.md`(표식 없음) 위에서 돌렸더니(mo.ai.kr, 2026-11-16) 생성이 `managed block missing or removed; baseline mismatch` 충돌로 **통째로** 멈췄고, 같은 단계의 `REVIEW.md`와 `.gitignore` 블록도 적용되지 못했고, `CLAUDE.md`가 `AGENTS.md`를 가리는 것은 말만 나왔어요. 고를 수 있는 것이 없는 막다른 길이에요.

tip(d25ecb1ba1)의 `section_plan`은 매니페스트에 기준선이 없는 표식 없는 파일을 이미 `append`로 계획해요. 충돌은 기준선이 있고, 표식이 사라졌고, 렌더 글이 기준선과 달라진 경우에만 나요. 이 SPEC은 그 충돌과 그에 묶인 두 가지에 사용자가 고르는 길을 더해요. 선택은 요청의 `picks` 목록으로 들어가고, 고르지 않으면 지금처럼 충돌로 멈춰요(기본값은 바뀌지 않아요).

| 선택 id | 하는 일 |
| --- | --- |
| `append-managed` | 표식이 사라진 `AGENTS.md`의 끝에 관리 구역을 덧붙이고 나머지 바이트는 그대로 둬요 |
| `skip-conflicts` | 충돌한 파일은 건드리지 않고(`skip`), 충돌하지 않은 파일(`REVIEW.md`, `.gitignore` 블록 등)은 적용해요 |
| `claude-import` | `AGENTS.md`를 가리는 `CLAUDE.md`에 `@AGENTS.md` 한 줄을 덧붙여요 |

## Non-goals

- 단계 JSON(`project-update-plan`의 `generation` 단계)에 선택지와 명령을 싣기, 그리고 `generation-plan`·`generation-apply` 요청이 `picks`를 `agents` 계획으로 넘기기: `GenerationRequest`를 만드는 곳(`src/project_update.rs:593`)이 C-107 예약이라 조각 2b로 미뤄요. 이 조각의 `picks`는 `agents-plan`·`agents-apply`에서만 받아요.
- 워크플로 글(`skills/moai/**`)에 선택지 문장 넣기(카드 기준 5): manifest 예약이 풀린 뒤 별도 SPEC.
- 편집된 관리 블록(`managed block edited after baseline`)과 깨진 표식(`malformed managed markers`)의 충돌: `append-managed`가 풀지 않아요(`skip-conflicts`만 적용돼요).
- `CLAUDE.local.md`·`AGENTS.override.md` 쓰기, 번들 풀기, `codex-hooks` 충돌.
- CLI 플래그 추가: 선택은 stdin JSON 요청의 `picks`로만 들어가요(`src/main.rs`·`src/help.rs`는 건드리지 않아요).

## Requirements

- REQ-1: The agents plan shall accept an optional `picks` list in the request with the ids `append-managed`, `skip-conflicts` and `claude-import`.
- REQ-2 (REQ-1): IF `picks` holds any other id, THEN the plan shall refuse the request as invalid input and write nothing.
- REQ-3: WHEN `append-managed` is picked and `AGENTS.md` holds no managed block while the manifest holds a baseline for it, the plan shall report the action `append` for `AGENTS.md` with its own digest and write nothing.
- REQ-4: WHEN the `append` plan is applied, `AGENTS.md` shall keep every earlier byte and hold exactly one begin marker and one end marker after them. · failure: none
- REQ-5 (REQ-3): IF an apply carries the digest of the plan made without the pick, THEN the apply shall refuse it as a stale plan and write nothing.
- REQ-6 (REQ-3): IF the conflict of `AGENTS.md` is an edited block or malformed markers, THEN `append-managed` shall leave the row as `conflict`.
- REQ-7: WHEN the receipt of the `append` apply is recovered, `AGENTS.md` shall return to the bytes it held before the apply. · failure: none
- REQ-8: WHEN `skip-conflicts` is picked, the plan shall report each conflicting file as `skip` and leave the action of every other file as it is. · failure: none
- REQ-9: WHEN a plan with `skip` rows is applied, the apply shall write the files that are not skipped and leave each skipped file byte for byte as it was.
- REQ-10 (REQ-9): IF a file was skipped, THEN the apply shall record no new baseline for it, so that the next plan without the pick reports `conflict` again.
- REQ-11: WHEN `claude-import` is picked and `CLAUDE.md` or `.claude/CLAUDE.md` exists without importing `AGENTS.md` while the profile hosts include `claude`, the plan shall report one `append` row of the import line for that file.
- REQ-12: WHEN the import row is applied, the file shall gain only the import line (`@AGENTS.md`, or `@../AGENTS.md` for `.claude/CLAUDE.md`) after its earlier bytes. · failure: none
- REQ-13 (REQ-11): IF the file already imports `AGENTS.md`, THEN the plan shall carry no import row.
- REQ-14: The plan shall report the path of a shadowing `CLAUDE.md` in an optional `claude_shadow` field that is absent when no file shadows `AGENTS.md`.
- REQ-15: A request without picks shall plan, digest and apply as it did before.
- REQ-16: The import row shall not record the file as a file MoAI owns.
- REQ-17: The `claude_shadow` field shall not change the plan digest.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-15): 선택 없는 요청은 계획·digest·적용이 이전과 같고, 모르는 id는 입력 오류로 거절돼요.
  verify: `cargo test --locked --offline --test agents_conflict_choices_cli picks`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): 기준선이 있는데 표식이 사라진 `AGENTS.md`가 `append-managed` 계획에서 `append`가 되고, 적용하면 앞 바이트가 그대로이고 표식 쌍이 하나이며, 선택 없이 만든 digest는 거절되고, 편집된 블록·깨진 표식은 `conflict`로 남고, 복구하면 원래 바이트로 돌아가요.
  verify: `cargo test --locked --offline --test agents_conflict_choices_cli append`
- AC-3 (REQ-8, REQ-9, REQ-10): `skip-conflicts` 계획이 충돌 파일만 `skip`으로 하고 나머지는 그대로이며, 적용이 나머지를 쓰고 건너뛴 파일은 바이트가 같고 기준선이 기록되지 않아요.
  verify: `cargo test --locked --offline --test agents_conflict_choices_cli skip`
- AC-4 (REQ-11, REQ-12, REQ-13, REQ-14, REQ-16, REQ-17): `AGENTS.md`를 가리는 `CLAUDE.md`가 보고되고, `claude-import` 계획이 import 줄 한 행을 내고, 적용이 줄만 덧붙이고, 이미 import하는 파일에는 행이 없어요.
  verify: `cargo test --locked --offline --test agents_conflict_choices_cli claude`
- AC-5: 이웃 검사.
  verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test review_section_cli`; `cargo test --locked --offline --test gitignore_block_cli`; `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test ac_command_lint`

## Scope

- write: `.moai/specs/SPEC-20261119-agents-conflict-choices/spec.md`, `src/agents.rs`, `src/generation.rs`(`AgentsRequest` 리터럴 두 곳에 `picks: Vec::new()`만), `src/claude.rs`(`imports_agents`를 `pub(crate)`로), `tests/agents_conflict_choices_cli.rs`
- read-only: `src/project_update.rs`, `src/contract.rs`, `src/apply.rs`
- forbidden: 실제 HOME 쓰기, `src/main.rs`, `src/help.rs`, `src/ops/**`, `skills/moai/**`, `src/project_update.rs` 편집

## Plan

1. 시험을 먼저 써요(RED): `picks` 필드가 없어서 요청이 거절되거나 충돌로 끝나야 해요. 기준선이 있고 표식이 없는 상태는 `product_cli.rs`의 충돌 시험처럼 만들어요(프로필 → `generation-apply` → 사람 글로 교체 → 언어 변경).
2. `agents.rs`: `picks`와 검증, `append-managed`(기준선 없이 다시 계획), `skip-conflicts` 후처리와 적용·기준선 기록 규칙, `claude_shadow`와 import 행.
3. `generation.rs`는 `AgentsRequest` 리터럴 두 곳에 `picks: Vec::new()`만 더해요(`generation` 요청으로 `picks`를 넘기는 일은 2b).
4. `claude.rs`는 `imports_agents`를 `pub(crate)`로 바꿔요.
5. 이웃 시험, 변이 시험, fmt·clippy·`spec check`.

## Risks

- `spec check` 가 "REQ 12개 초과" 경고를 내요(17개). 선택 셋이 한 계획·적용 경로를 같이 써서 세 SPEC 으로 쪼개면 같은 파일을 세 번 고치게 돼요 → 한 SPEC 으로 두고, 감사가 쪼개기를 요구하면 따라요.
- `skip` 행을 매니페스트 판정에서 `unchanged`처럼 다루지 않으면 매니페스트 행이 `merge`로 계속 남아요 → REQ-10 시험이 지켜요.
- 선택 없는 요청의 digest가 바뀌면 이미 만든 계획이 낡은 것으로 거절돼요 → `picks`는 비어 있으면 직렬화하지 않아요(`skip_serializing_if`), REQ-1 시험이 digest 같음을 봐요.
- `skip` 행의 `action`을 읽는 다른 곳(`update_tux`, `ui_summary` 등)이 모르는 값에 어떻게 반응하는지는 아직 읽지 않았어요 → 구현 중에 `rg 'action' src`로 읽고 결과를 Result에 적어요.
- `apply-recover`가 `append` 행을 되돌리는지는 조사에서 읽기만 했어요(왕복 시험 없음) → REQ-7 시험이 처음 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 선택은 CLI 플래그가 아니라 요청 JSON의 `picks`로 받는다 | `src/main.rs`·`src/help.rs`를 다른 카드가 쥐고 있고, 계획 digest가 요청을 묶어서 낡은 계획 거절이 공짜로 따라와요 | `picks`를 지우면 돼요 |
| D-2 | 선택 없이는 지금처럼 충돌로 멈춘다 | 기본값을 바꾸면 사용자의 글을 묻지 않고 건드릴 수 있어요 | 해당 없음 |
| D-3 | 단계 JSON의 선택지는 2b로 미룬다 | `src/project_update.rs`가 C-107 예약이에요 | 2b에서 이어요 |

## Result

(구현 뒤에 AC마다 한 줄씩 적어요.)

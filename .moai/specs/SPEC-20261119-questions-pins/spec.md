---
id: SPEC-20261119-questions-pins
title: "Pin the two question-repair sentences that no test reads"
status: draft
revision: 1
created: 2026-11-19
---

# Pin the two question-repair sentences that no test reads

## Intent

C-51 재감사(PR #3136, head `3dff58442d`, PASS)의 N8: 질문 규칙을 고친 문장 두 개에 시험이 없어서, 누가 지워도 `cargo test`가 초록이에요. 문장이 이미 글에 있으므로 이 SPEC은 글을 바꾸지 않고 **시험으로 고정**만 해요.

- `skills/moai/references/internal/questions.md`의 OMP 행: `the recommendation is named in the label text only, and a submission without a choice is not a pick` — OMP의 `ask`는 사용자가 제때 답하지 않으면 폼의 값을 제출하므로, 이 문장이 "무응답 제출은 선택이 아니다"를 지켜요. 같은 행의 앞부분(`nothing is preselected: no checked …`)만 지금 고정돼 있어요.
- `skills/moai/workflows/project.md`의 작업 원장 줄: `creating it writes under the real HOME and waits for its own pick after the read-only moai ledger-register-plan <project> preview` — 같은 뜻의 `setup.md` 문장만 고정돼 있고 `project.md` 쪽은 없어요.

같은 재감사의 N7은 tip에서 이미 참이에요(`moai-cli key-set --prompt`가 있고, `tests/update_route.rs`가 "the user types it into the product's own hidden prompt in their own terminal"을 고정해요). N9·N10은 글을 바꾸는 일이라 `skills/moai/references/source-manifest.json`이 필요한데 다른 카드가 그 파일을 예약 중이어서, 이 SPEC 뒤의 별도 SPEC으로 둬요.

## Non-goals

- N7: 글도 시험도 바꿀 것이 없어요(위 증거를 카드에 적어요).
- N9(route-end offer의 네 조항)와 N10(Codex Default 모드의 권한 요청 문장): 글 변경과 manifest 갱신이 필요해요 — 후속 SPEC.
- `skills/moai/**` 파일 편집, manifest 갱신: 이 SPEC은 글을 바꾸지 않아요.
- 문장의 내용 바꾸기: 지금 글 그대로 고정해요.

## Requirements

- REQ-1: The update-route test shall require the OMP row of `references/internal/questions.md` to say that a submission without a choice is not a pick.
- REQ-2: The update-route test shall require the OMP row to say that the recommendation is named in the label text only.
- REQ-3: The update-route test shall require `workflows/project.md` to say that creating the work ledger writes under the real HOME and waits for its own pick after the read-only `moai ledger-register-plan <project>` preview.
- REQ-4 (REQ-1): IF the sentence of REQ-1 is missing from the OMP row, THEN the update-route test shall fail and name the OMP row, even when the sentence appears elsewhere in the file.
- REQ-5 (REQ-3): IF the sentence of REQ-3 is missing from `workflows/project.md`, THEN the update-route test shall fail and name `project.md`.
- REQ-6: The change shall edit no file under `skills/moai/`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): OMP 행에서 두 구절 중 하나를 지우면 `every_question_is_a_selection_through_the_hosts_question_tool`이 실패하고, 지우지 않은 글에서는 통과해요.
  verify: `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool`
- AC-2 (REQ-3, REQ-5): `project.md`에서 원장 문장을 지우면 `the_ledger_is_local_cards_and_previews_run_without_asking`이 실패하고, 지우지 않은 글에서는 통과해요.
  verify: `cargo test --locked --offline --test update_route the_ledger_is_local_cards_and_previews_run_without_asking`
- AC-3 (REQ-6): 이웃 검사와 `skills/moai/` 무변경.
  verify: `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test ac_command_lint`

## Scope

- write: `.moai/specs/SPEC-20261119-questions-pins/spec.md`, `tests/update_route.rs`
- read-only: `skills/moai/references/internal/questions.md`, `skills/moai/workflows/project.md`, `skills/moai/workflows/setup.md`, `skills/moai/references/source-manifest.json`
- forbidden: `skills/moai/**` 편집, `src/**`, 실제 HOME

## Plan

1. 시험 줄을 `tests/update_route.rs`에 더해요. 시험이 지금 글을 고정하는 일이라 구현 전에 실패(RED)하는 줄은 없어요. 대신 글의 문장을 하나씩 지우는 변이 두 개를 만들어(`git checkout --`로 되돌려요) 시험이 실패하는 것을 읽어요.
2. 지우지 않은 글에서 통과(GREEN)를 읽어요.
3. 이웃 검사(`update_route` 전체, `ac_command_lint`, `home_isolation_scan`)를 돌려요.

## Risks

- OMP 문장이 같은 파일의 다른 곳에도 있으면 행 안에서 지워도 시험이 통과해요 → 시험은 OMP 행 한 줄만 읽어요(기존 `row("OMP")`와 같은 방식).
- 변이를 위해 `skills/moai/` 파일을 임시로 고쳐요 → 바로 `git checkout --`로 되돌리고 `git status`가 깨끗한지 읽어요; 커밋하지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | N7은 글도 시험도 바꾸지 않아요 | tip에서 `key-set --prompt`가 있고 그 문장이 이미 고정돼 있어요 | 없음 |
| D-2 | N9·N10은 이 SPEC 밖이에요 | 글 변경은 manifest 갱신이 같은 변경에 필요한데 그 파일이 다른 카드의 예약 중이에요 | 예약이 풀리면 후속 SPEC |
| D-3 | 기존 두 시험의 목록에 줄을 더해요(새 시험 함수 없음) | 같은 문장 묶음을 같은 곳에서 읽는 편이 한눈에 보여요 | 없음 |

## Result

- AC-1: met — `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool` ok(같은 파일 8 passed). OMP 행에 두 구절(`the recommendation is named in the label text only`, `a submission without a choice is not a pick`)을 더했어요. 시험이 현재 글을 고정하는 일이라 구현 전에 실패하는 줄은 없고, 대신 변이로 확인했어요: M1(OMP 행에서 두 번째 구절을 지움)·M2(첫 구절을 지움)·M4(두 번째 구절을 행 밖 파일 끝으로 옮김) 모두 KILLED이고 실패 글이 `OMP row lacks …`로 행을 이름 불러요(M4는 같은 문장이 파일 안에 있어도 행만 읽는다는 REQ-4의 증거예요).
- AC-2: met — `cargo test --locked --offline --test update_route the_ledger_is_local_cards_and_previews_run_without_asking` ok. `project.md`에서 원장 문장(`; creating it writes under the real HOME and waits for its own pick after the read-only … preview`)을 지운 변이 M3이 KILLED이고 실패 글이 `project.md lacks "Which ledger is used is what is not asked; creating it writes under the real HOME and waits for its own pick …"`예요.
- AC-3: met — `update_route` 전체 8 passed(통과 줄을 더한 head), `ac_command_lint` 5·`home_isolation_scan` 18·`spec_cli` 17·`product_cli skill_package` 18 passed, `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, `cargo fmt --check` 출력 없음, `spec check` errors 0·warnings 0. 변이 4개는 임시로 `skills/moai/` 두 파일을 고쳤다가 매번 `git checkout --`로 되돌렸고 각 회차 뒤 `git status`가 깨끗했으며, 이 가지가 바꾼 파일은 `tests/update_route.rs`와 이 SPEC뿐이에요(REQ-6).
- 남은 것: 같은 재감사의 N7은 tip에서 이미 참이고 고정돼 있어서(`moai-cli key-set --prompt`, `tests/update_route.rs`의 "the user types it into the product's own hidden prompt in their own terminal") 바꾼 것이 없어요. N9(route-end offer의 네 조항)와 N10(Codex Default 모드의 권한 요청 문장)은 글 변경이라 `skills/moai/references/source-manifest.json`이 같은 변경에 필요한데, 그 파일이 다른 카드의 예약 중이라 이 SPEC 밖이에요(후속 SPEC). SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.

---
id: SPEC-20261119-interview-guidance
title: "A new SPEC starts with an Interview section, and the edges of the interview check are written down and pinned"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-enforce-interview-record
---

# A new SPEC starts with an Interview section, and the edges of the interview check are written down and pinned

## Intent

SPEC-20261119-enforce-interview-record 가 승인된 SPEC 에 `## Interview` 절을 요구하게 했고(감사 PASS, head 308f356736), 그 감사가 남긴 비차단 항목을 이 조각이 닫아요. 지금은 `moai spec new` 가 만든 초안에 이 절이 없어서, 새 SPEC 을 승인하는 순간 `interview_missing` 이 나요(감사 N7: 기준일 이후의 초안 SPEC 13개가 같은 처지예요). 템플릿이 절을 미리 가지고 있으면 사람이 승인 전에 절을 채우게 돼요.

이 조각은 다섯 가지를 해요: (1) 템플릿 두 개에 `## Interview` 절을 넣어요(감사 N7). (2) 템플릿의 예시 항목이 모양이 맞는 항목으로 읽히는지 시험이 고정해요. (3) `rejected`·`archived` 가 `interview_missing` 을 받지 않고 `completed`·`superseded` 에서 `interview_entry` 가 오류인 것을 시험이 고정해요(감사 N1·N2). (4) 검사기의 모듈 주석과 운영 문서가 절과 두 코드를 말하게 해요(감사 N6). (5) 검사기가 못 닫는 가장자리(HTML 주석 안의 절, 코드 펜스 안의 항목, 손으로 쓴 Go 형식 id)를 Risks 에 이름으로 적고 지금 통과한다는 시험으로 고정해요(감사 N3·N4).

## Interview

- Q1: 후속 조각의 범위는 어디까지인가요? → A: 리더 지시(S2 감사 PASS 뒤): 템플릿과 plan.md 에 Interview 절 안내(N6·N7), N1·N2 고정 시험, N3·N4 는 Risks 에 적거나 막기; S4 재작업은 #3216 병합 뒤.
- Q2: plan.md 와 `source-manifest.json` 도 이 조각에서 고치나요? → A: 아직 아니에요. 매니페스트는 C-59 가 잡고 있어 쓰기 범위 충돌이 나서, 이 조각은 plan.md 를 빼고 리더에게 범위 요청을 했어요(Non-goals).

## Non-goals

- `skills/moai/workflows/plan.md` 에 인터뷰 절차를 적기: 같은 PR 에서 `skills/moai/references/source-manifest.json` 의 sha256 을 갱신해야 하는데 그 파일을 C-59 가 잡고 있어요. 범위가 풀리면 별도 조각으로 해요.
- 이미 만들어진 초안 SPEC 13개(기준일 이후)에 인터뷰 절을 채우기: 그 SPEC 들은 다른 카드의 것이고, 승인 전에 각 작성자가 채워야 해요(Risks).
- `## Interview` 절의 규칙(기준일, 항목 모양, 심각도) 바꾸기: 이 조각은 지금 동작을 이름으로 고정하기만 해요.
- HTML 주석과 코드 펜스를 읽는 파서로 고치기: 모든 절에 걸친 일이라 별도 조각이에요(Risks).
- Go 형식 id 의 새 SPEC 에 인터뷰를 요구하기: Go 세대 SPEC 의 오류는 경고로 낮춰 읽는 규칙(`LEGACY_KEEPS`)을 건드려야 해요(Risks).

## Requirements

- REQ-1: The built-in SPEC templates shall carry an `## Interview` section, in the full form and in the minimal form, whose example entry holds `<…>` placeholders.
- REQ-2: The SPEC checker shall read the template's example entry as a well-formed interview entry, so that a SPEC written by `moai spec new` has no `interview_entry` finding until its placeholders are replaced.
- REQ-3: The SPEC checker shall report `interview_missing` only for the statuses approved, completed and superseded, and `interview_entry` as an error for each of them.
- REQ-4: The module documentation of the SPEC checker and the SPEC operations document shall name the `## Interview` section and the codes `interview_missing` and `interview_entry`.
- REQ-5: The SPEC checker shall keep reading as satisfying the interview rule a section whose heading sits in an HTML comment, entries inside a code fence under a real heading, and a Go-form id, each named in this SPEC's Risks.

## Acceptance

- AC-1 (REQ-1, REQ-2): `render` 로 만든 전체·최소 템플릿 초안 둘 다 `## Interview` 절을 가지고, 예시 항목이 `interview_entry` 없이 읽히며, 승인 상태로 바꾸면 `placeholder` 오류가 나서 예시를 바꾸게 만들어요.
  verify: `cargo test --locked --offline --test spec_doc interview_template`
- AC-2 (REQ-3): `rejected`·`archived`(2026-11-19 이후)는 `interview_missing` 이 없고, `completed`·`superseded` 에서 틀린 항목은 오류 `interview_entry` 예요.
  verify: `cargo test --locked --offline --test spec_doc interview_statuses`
- AC-3 (REQ-4): 모듈 주석과 운영 문서가 절과 두 코드를 말해요.
  manual: `src/spec_doc.rs` 의 모듈 주석과 `docs/design-review/appendix/operations/spec.md` 에서 `## Interview`, `interview_missing`, `interview_entry` 를 찾아 읽어요.
- AC-4 (REQ-5): HTML 주석 안의 절, 코드 펜스 안의 항목, Go 형식 id 가 지금 통과한다는 시험.
  verify: `cargo test --locked --offline --test spec_doc interview_edges`
- AC-5 (REQ-1…REQ-5): 이웃 시험.
  verify: `cargo test --locked --offline --test spec_doc`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test card_spec_cli`; `cargo test --locked --offline --test spec_contract`; `cargo test --locked --offline --lib spec`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo fmt --check`

## Scope

- write: `src/spec_doc.rs`, `templates/spec/full.md`, `templates/spec/minimal.md`, `tests/spec_doc.rs`, `tests/spec_cli.rs`, `docs/design-review/appendix/operations/spec.md`, `.moai/specs/SPEC-20261119-interview-guidance/spec.md`
- read-only: `skills/moai/workflows/plan.md`, `skills/moai/references/source-manifest.json`, `src/spec.rs`, `src/spec_cli.rs`
- forbidden: the real HOME and the real host settings, `skills/**`, `src/hook_engine.rs`, `src/hook_guards.rs`, any change to the interview rule itself (the date, the entry forms, the severities)

## Plan

1. 시험 먼저(`tests/spec_doc.rs`): `interview_template`, `interview_statuses`, `interview_edges` 를 더해 템플릿 시험이 실패(RED)하는 것을 봐요. `interview_statuses` 와 `interview_edges` 는 지금 동작을 고정하므로 처음부터 통과해요.
2. 템플릿 두 개에 `## Interview` 절(`Q1: <…> → A: <…>` 예시 항목과, 물을 것이 없었을 때를 말하는 한 줄).
3. `src/spec_doc.rs` 모듈 주석과 운영 문서에 절·코드 이름.
4. 이웃 시험(템플릿 문구에 기대는 시험이 있는지 포함), mutation, SPEC `## Result`.

## Risks

- HTML 주석 안에 쓴 `## Interview` 제목도 절로 읽혀서 규칙을 채워요(`body` 가 `<!-- -->` 를 모르고, 다른 절의 제목도 같아요). 시험 `interview_edges` 가 이 사실을 고정하고, 고치려면 모든 절에 걸친 파서 변경이라 별도 조각이에요.
- 코드 펜스 안의 `- Q1: x → A: y` 줄도 제목 아래의 항목으로 읽혀요(`interview` 가 펜스를 몰라요). 같은 시험이 고정해요. 항목을 펜스에 넣어 "기록 있음"으로 위장할 수 있지만 이 검사는 모양만 보고(SPEC-20261119-enforce-interview-record Risks), 진짜 인터뷰였는지는 감사자가 대화와 비교해요.
- 손으로 쓴 Go 형식 id(`SPEC-AUTH-001`)의 SPEC 은 기준일 이후에 만들어도 인터뷰를 요구받지 않아요(새 형식 id 만 물어요). Go 세대의 오류를 경고로 낮춰 읽는 규칙 때문에 막으려면 그 규칙(`LEGACY_KEEPS`)을 바꿔야 해서 이 조각 밖이에요. 시험 `interview_edges` 가 고정해요.
- 이미 만들어진 기준일 이후의 초안 SPEC(감사가 센 13개)은 템플릿을 고쳐도 `## Interview` 가 없어요: 승인하는 순간 `interview_missing` 이 나고, 고치는 길은 `- Q<n>: … → A: …` 줄이나 `- none: <10자 이상의 이유>` 한 줄이에요. 이 SPEC 의 Result 에 그 목록과 확인한 방법을 적어요.
- 템플릿의 예시 항목은 `<…>` 자리표시자를 가져서, 채우지 않은 채 승인하면 `placeholder` 오류가 나요(승인 상태에서만 오류). 초안은 경고만이에요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 템플릿의 예시 항목은 `<…>` 자리표시자 | 기존 `placeholder` 검사가 승인 전에 바꾸게 강제해요(새 규칙 없이) | 자리표시자 없는 예시 문장(그대로 둔 채 승인돼도 안 걸려요) |
| D-2 | N3·N4 는 막지 않고 Risks 에 이름 + 통과 시험 | 막으려면 모든 절의 파서와 Go 세대 낮춤 규칙을 바꿔야 해요; 이 검사는 애초에 모양만 봐요 | HTML 주석·펜스를 아는 파서(별도 조각) |
| D-3 | plan.md 는 범위가 풀린 뒤 별도 조각 | 매니페스트 sha 를 같은 변경에서 갱신해야 하는데 C-59 가 잡고 있어요 | 매니페스트 없이 plan.md 만 수정(전체 시험 `source_manifest_covers_every_package_file` 이 깨져요) |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. macOS arm64, 이 worktree(`spec/20261119-interview-guidance`), 빌드 폴더 `~/.moai/build/moai-adk-86-s2`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 시험 개수는 `cargo test --test <이름>` 을 파일마다 돌려 그 파일 로그의 `test result` 줄에서 읽었어요(`target/c86s2/m-*.log`, 종료 코드는 `target/c86s2/nb2-exits.txt`).

RED 먼저: 이번 조각의 RED 는 따로 로그로 남기지 않았어요. 대신 변이 Y1·Y2(템플릿에서 절을 지움)·Y3(예시 항목에서 화살표를 지움)·Y4(자리표시자를 지움)가 같은 시험 `interview_template` 을 실패시키는 것을 `target/c86s2/mutants2.txt` 로 보였어요. 시험 `interview_statuses` 와 `interview_edges` 는 지금 동작을 이름으로 고정하는 것이라 처음부터 통과하는 게 맞아요(Plan 1). 규칙 코드는 이 조각에서 바꾸지 않았어요(모듈 주석만).

- AC-1: met — `cargo test --locked --offline --test spec_doc interview_template` 통과(전체·최소 템플릿 모두 `## Interview` 절을 가지고, 예시 항목이 `interview_entry` 없이 읽히고, 승인 상태로 바꾸면 예시 항목 줄에 `placeholder` 오류가 나고 `interview_*` 코드는 없어요). 템플릿 줄이 늘어서 `tests/spec_cli.rs` 의 기대 줄 `spec.md:23 ac_no_verify` 를 29 로 바꿨어요.
- AC-2: met — `interview_statuses` 통과(`rejected`·`archived`·`draft` 는 2026-11-19 에 `interview_*` 코드가 없고; `approved`·`completed`·`superseded` 는 기록이 없으면 `interview_missing` 오류 하나, 틀린 줄이 있으면 `interview_entry` 가 오류이고 경고가 아니에요).
- AC-3: met(`manual:`, 읽어서 확인) — `src/spec_doc.rs:24-30` 모듈 주석이 `## Interview`, `interview_missing`, `interview_entry` 를 말하고, `docs/design-review/appendix/operations/spec.md:21` 이 같은 셋을 말해요(`grep -n` 으로 줄을 찾았어요).
- AC-4: met — `interview_edges` 통과(HTML 주석 안의 `## Interview` 는 절로 읽혀 규칙을 채우고, 코드 펜스 안의 `- Q1: x → A: y` 도 항목으로 읽히고, 손으로 쓴 Go 형식 id(`SPEC-AUTH-001`)는 묻지 않아요). 셋 다 Risks 에 이름으로 적혀 있어요.
- AC-5: met — `--test spec_doc` 30, `spec_cli` 18, `card_spec_cli` 11, `spec_contract` 8, `ac_verify_required` 6, `audit_admission` 7, `boundary_repro` 16, `plan_admission` 10, `spec_close_cli` 8, `ac_command_lint` 5, `git_env_scrub` 3, `green_test_lock` 14(템플릿 경로를 읽는 시험), `--lib spec` 15 모두 통과, `cargo fmt --check` 종료 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 종료 0, `python3 -m unittest tests.ci_scope_test` 종료 0. `spec check --all` 은 오류 0·경고 5(모두 다른 SPEC 의 기존 `needs_clarification`·`size_requirements`, 이 SPEC 은 0).
- mutation: 11개를 돌려 11개 모두 KILLED(`target/c86s2/mutants2.txt`): Y1 전체 템플릿 절 없음, Y2 최소 템플릿 절 없음, Y3 예시 항목 화살표 없음, Y4 최소 템플릿 자리표시자 없음, Y5 `completed`·`superseded` 의 틀린 줄이 경고, Y6 `superseded` 를 안 물음, Y7 `rejected` 를 물음, Y8 id 형태를 안 봄, Y9 기준일을 하루 뒤로, Y10 절이 있는지 안 봄, Y11 `created` 를 안 봄. 돌린 뒤 `git status` 는 clean 이었어요.
- 기준일 이후 초안 SPEC 중 `## Interview` 가 없는 것(N7): 병합된 tip `f6dc362f6e` 에서 `git ls-tree` 로 `.moai/specs/` 폴더를 모두 읽고, 머리말의 `created:` 가 2026-11-19 이후인 것 13개를 골라 `^## Interview` 줄이 없는 것을 셌더니 13개 전부였어요(감사가 센 13개와 같아요): `audit-eval`, `card-reclaim`, `card-reclaim-followups`, `enforce-spec-hooks`(우리 C-86 S1 의 것), `flaky-policy`, `heavy-slot-pressure`, `hooks-followups`, `launcher-settings-receipt`, `macos-on-pr`, `nested-manifests`, `questions-pins`, `self-check-diff`, `verified-trailer`(모두 `SPEC-20261119-` 접두, 모두 `draft`). 병합 전 가지에만 있는 SPEC 은 세지 않았어요. 승인하기 전에 각 작성자가 `- Q<n>: … → A: …` 줄이나 `- none: <이유>` 한 줄을 채워야 해요.
- 이 조각 밖: `skills/moai/workflows/plan.md` 의 인터뷰 절차(N6 의 plan.md 절반)는 `source-manifest.json` 을 C-59 가 잡고 있어 쓰기 범위를 리더에게 요청해야 해요(Non-goals, D-3). 질문 도구를 실제로 썼는지는 이 검사가 알 수 없어요(Risks).

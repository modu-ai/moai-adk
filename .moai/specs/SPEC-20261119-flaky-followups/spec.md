---
id: SPEC-20261119-flaky-followups
title: "flaky policy follow-ups: a closable card, the repair rule, one record home, a list filter, the re-run signal (C-99 slice 3)"
status: draft
revision: 1
created: 2026-11-19
---

# flaky policy follow-ups (C-99 slice 3)

## Intent

카드 C-99 (p1): 간헐 실패 시험은 다시 돌려 통과시키지 않고 카드·원인 시도·삭제 기록으로 다뤄요. 조각 1(병합됨, PR #3225·#3229)이 규칙 문장(`skills/moai/workflows/mode-leader.md` 32행)을, 조각 2(`SPEC-20261119-flaky-budget-pins`)가 위치·예산 핀을 넣었어요. 2026-11-19 tip `747738af63` 에서 읽은 남은 일:

- **카드 기준 2 가 tip 에 없어요**: `skills/moai/workflows/mode-agent.md` 에 "수리가 sleep·재시도·건너뛰기를 더하지 않는다, 원인을 못 찾으면 이유를 적고 삭제한다"는 문장이 없어요(`sleep` 0건, `root cause` 0건; `retry` 6건은 원장 재시도 같은 다른 뜻).
- **카드 기준 3 이 tip 에 없어요**: `moai card list --help` 의 옵션은 `--state` 와 `--home` 뿐이라 간헐 카드를 걸러 볼 수 없어요(`src/card/command.rs:96` `list(project, home, state)`).
- **감사 N2(#3225)**: 규칙의 명령 `moai card add --kind task --title "flaky: <test>"` 에 `--ac` 가 없어요. `card done` 은 충족된 기준이 하나 이상 있어야 하고, `moai card criterion` 은 기준을 충족/철회로 표시만 하고 더하지 못해서(`moai card --help`, `card add --help` 로 확인) 그 카드는 닫을 수 없어요.
- **감사 N3**: `skills/moai/workflows/verify.md` 53행은 간헐 시험을 `[flake]` Issue 한 건에 기록하라고(`flake-record`) 하는데 `mode-leader.md` 32행은 카드에 기록하라고 해서 기록 장소가 둘이에요.
- **감사 N4**: "재실행의 통과는 검증으로 인용하지 않는다"에 게이트가 없어요. `src/gate.rs` 의 G5(`g5`, 1036행 근처)는 `CheckRun {name, conclusion, head_sha}` 만 읽어서 첫 실행 통과와 재실행 통과를 구분하지 못해요.

기준 1(규칙 연결)은 tip 에서 충족돼요(`mode-leader.md` 32행이 규칙을 이름 붙여 잇고 `workflow_size_budget` 위치 핀이 지켜요).

## Non-goals

- 규칙의 뜻 바꾸기: 재실행 통과도 병합을 막지 않는다는 C-99 AC-1 은 그대로예요. N4 는 그 통과를 *인용하지 않게* 하는 신호만 더하고 G5 의 판정은 바꾸지 않아요.
- `flake-classify`·`flake-record` 명령(`src/flake.rs`) 자체의 동작 변경.
- 기존 카드·원장 행 수편, 실제 HOME, 네트워크(N4 의 GitHub 필드 조사는 문서 읽기만).
- C-103 의 rule-prune 조각(별도 배차)과 겹치는 `mode-leader.md` 다른 문장.

## Requirements

- REQ-1: When the merge-train text tells the leader to open a flaky card, it shall give the command with one `--ac` criterion line that a person can mark met by fixing the test at its root cause or by deleting it with a written reason. · failure: none
- REQ-2: When the MODE-AGENT text describes the repair of a flaky test, it shall forbid adding a sleep, a retry or a skip to make the test pass and require a test whose root cause is not found to be deleted with the reason written in the card and the SPEC. · failure: none
- REQ-3: When the VERIFY text tells how a flaky test is recorded, it shall name the flaky card of the merge-train step as the record's home and keep the `[flake]` Issue and `flake-record` only for a project without a card ledger. · failure: none
- REQ-4: When `moai card list` is given `--title-prefix <text>`, the command shall list only the cards whose title starts with that text, case-sensitively, in the order and shape it uses without the option.
- REQ-5 (REQ-4): IF the prefix is empty or holds a control character, THEN `moai card list` shall refuse before it reads the ledger.
- REQ-6: When the gate renders the CI information line of a head, it shall add `attempt=<n>` for a `ci` run that succeeded as its attempt `n` of two or more. · failure: none
- REQ-7: When the gate reads a check run that carries no attempt number, it shall treat the run as attempt one and render the line exactly as before. · failure: none
- REQ-8: When any attempt number is present, the verdict of rule G5 shall be the one it gives without it. · failure: none
- REQ-9: When the added texts are in the workflow files, `tests/workflow_size_budget.rs` shall still pass. · failure: none
- REQ-10: When a skill file changes in a commit, the same commit shall carry the manifest line that `python3 -I tools/source_manifest.py write` produces for it. · failure: none

## Acceptance

- AC-1 (REQ-1): `mode-leader.md` 의 간헐 규칙 명령이 `--ac` 한 줄을 갖고, 그 문장 그대로 `moai card add` 를 임시 원장에서 돌려 만든 카드가 기준 하나(열림, "flake" 를 담음)를 갖고 `card show` 로 읽혀요. 닫는 일(`card done`)은 병합된 PR 연결을 요구해서 이 시험 밖이에요.
  verify: `cargo test --locked --offline --test flaky_followups the_flaky_card_command_carries_an_ac_and_the_card_can_be_read_with_it`
- AC-2 (REQ-2, REQ-3): `mode-agent.md` 에 sleep·재시도·건너뛰기 금지와 삭제 규칙 문장이, `verify.md` 에 기록 장소 문장이 있고 `[flake]` Issue 경로는 "원장이 없는 프로젝트"로 한정돼요(이름 붙인 파일만 읽는 문구 핀).
  verify: `cargo test --locked --offline --test flaky_followups the_repair_rule_is_in_the_mode_agent_workflow`
  verify: `cargo test --locked --offline --test flaky_followups the_record_home_is_in_the_verify_workflow`
- AC-3 (REQ-4, REQ-5): `--title-prefix "flaky: "` 가 접두어가 맞는 카드만, 빈 값과 제어 문자는 거부(원장을 읽기 전), 옵션이 없으면 전과 같은 출력이에요.
  verify: `cargo test --locked --offline --test card_list_title_prefix`
- AC-4 (REQ-6, REQ-7, REQ-8): 스냅샷에 시도 번호가 있는 성공 `ci` 가 `attempt=<n>` 을 정보 줄에 더하고 G5 판정은 같아요; 시도 번호 없는 스냅샷은 줄이 전과 같아요(기존 골든 불변).
  verify: `cargo test --locked --offline --test gate_check_cli ci_info_names`
  verify: `cargo test --locked --offline --test gate_collect a_`
- AC-5 (REQ-9, REQ-10): 크기 예산·위치 핀·manifest 검사가 통과해요.
  verify: `cargo test --locked --offline --test workflow_size_budget`
  verify: `python3 -m unittest tests.source_manifest_test`
- AC-6: 이웃 검사(대상 한정; 전량은 PR CI run 이 인용).
  verify: `cargo test --locked --offline --test product_cli skill_package`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-flaky-followups/**`; 조각 A(문구): `skills/moai/workflows/mode-leader.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `tests/flaky_followups.rs`; 조각 B(필터): `src/card/command.rs`, `src/main.rs`, `src/help.rs`, `docs/rust-product-path.md`, `tests/card_list_title_prefix.rs`; 조각 C(재실행 신호): `src/gate.rs`, `src/gate_collect.rs`, `src/flow.rs`(`run_id` 를 `pub(crate)` 로 한 줄), `tests/gate_check_cli.rs`, `tests/gate_collect.rs`
- read-only: `src/flake.rs`, `tests/workflow_size_budget.rs`, `src/ledger/**`(`card done` 의 기준 규칙)
- forbidden: 실제 HOME, 네트워크 쓰기, 원장 행 수편, G5 판정 규칙 변경, 다른 카드의 문장 편집

## Plan

한 조각에 한 가지만 묶고 조각마다 take 해요(예약이 겹치면 `LANE-BLOCKED`).

1. (지금) 이 SPEC — tip 에서 읽은 사실만, 코드·문구 변경 0.
2. 조각 A 문구: RED `tests/flaky_followups.rs`(이름 붙인 세 파일만 읽음 + 임시 원장에서 명령 실행) → `mode-leader.md` 32행의 명령에 `--ac`, `mode-agent.md` 에 REQ-2 문장, `verify.md` 53행 근처에 REQ-3 문장 → `source_manifest.py write` 같은 커밋. 막힘: `verify.md`·`mode-agent.md`·manifest 는 C-114(agent-17)가 쥐고 있어요(마지막 조각만 manifest 를 다시 받는 결정, 노트 3081) — C-114 가 비운 뒤 take. `mode-leader.md` 는 지금 비어 있어요.
3. 조각 B 필터: RED `tests/card_list_title_prefix.rs`(임시 원장) → `src/card/command.rs::list` 에 접두어 인자, `src/main.rs`·`src/help.rs`·docs 행. 막힘: `src/card/command.rs` 는 C-43 이 쥐고 있어요(병합 뒤 또는 그 head 위).
4. 조각 C 재실행 신호: 먼저 조사(문서 읽기만): GitHub 의 어느 필드가 시도 번호를 주는지(Actions 실행의 `run_attempt`, 확인은 공식 문서 인용) — 결과를 이 SPEC 의 Result 에 적고, 수집기가 그 필드를 읽는 변경이 이 SPEC 범위인지 리더가 정해요(`CheckRun` 의 선택 필드 + 정보 줄 + 골든).
5. 각 조각: fmt, clippy `--all-targets`, 관련 시험, 변이 ≤5, trailer-check, tip 과 merge-tree, push, 감사(조각 C 와 `src/gate.rs` 는 리더 감사).

## Risks

- N4 의 모양(조각 C)은 정책 선택이에요: (가) 정보 줄 `attempt=<n>` 만 더하고 G5 는 그대로(권장 — C-99 AC-1 과 부딪히지 않아요), (나) 문구만(게이트 없음), (다) G5 가 재실행 통과를 멈춰 세움(C-99 AC-1 "병합은 통과"와 충돌 — 비추천). 리더 결정이 필요해요.
- `verify.md`·`mode-agent.md`·manifest 는 다른 카드가 쥘 수 있어요. 예약이 겹치는 문제는 C-118 이 풀 예정이에요(그 연결이 병합되면 manifest 쌍은 겹쳐도 돼요).
- `card list` 의 접두어 필터는 제목 규칙(`flaky: `)에 기대요. 제목이 달라진 카드는 걸리지 않아요(규칙이 제목을 정해요).
- 문구가 늘면 크기 예산(`tests/workflow_size_budget.rs`)이 닫을 수 있어요. 문장은 한 줄씩만 더해요.

## Result

**조각 C(재실행 신호) — 2026-11-19~20, 가지 `spec/20261120-flaky-rerun-attempt`(로컬 `v2026.11.0` `e341e361cc` 에서 시작).** 조각 A 는 아래에, 조각 B 는 "남은 것"에 있어요.

조사(문서 읽기만, moai-research, 2026-11-19 15:43 UTC 읽음, WebFetch 요약을 거친 인용이라 원문과 글자 대조는 못 했어요):
- 체크 실행 객체(`GET /commits/{ref}/check-runs`)에는 시도 번호가 없어요. 문서화된 식별 필드는 `details_url`, `external_id`, `app`, `check_suite.id`, `head_sha` 뿐이고 페이지에 "attempt" 라는 단어가 없어요(https://docs.github.com/en/rest/checks/runs).
- 워크플로 실행 객체에는 `run_attempt`(integer)가 있어요(`GET /repos/{o}/{r}/actions/runs/{run_id}`, https://docs.github.com/en/rest/actions/workflow-runs). 재실행은 같은 `run_id` 를 두고 시도 번호만 올려요: "This number does not change if you re-run the workflow run." / "`github.run_attempt` … begins at 1 for the workflow run's first attempt, and increments with each re-run."(https://docs.github.com/en/actions/reference/workflows-and-actions/contexts).
- 문서에 없는 것: `details_url` 의 형식(`…/actions/runs/<id>/job/<job>` 는 `src/flow.rs::run_id` 가 이미 쓰는 가정이에요), check-runs `filter=all` 이 이전 시도를 어떻게 돌려주는지. 그래서 수집기는 `details_url` 에서 run id 를 얻지 못하거나 읽기가 실패하면 시도 없음(=1)으로 둬요.

정한 것(리더 결정 노트 3373, 요청은 노트 3120 — 되돌릴 수 있고 범주 none): N4 모양은 (가) 정보 줄 `attempt=<n>` 만, G5 는 그대로. N3 기록 장소는 카드가 기본, `[flake]` Issue·`flake-record` 는 원장 없는 프로젝트용. 감사 tier 는 A·B 가 1, C 는 2(리더 감사 — 판정은 안 바뀌어도 `src/gate.rs` 의 출력을 다른 도구가 읽을 수 있어서예요). 순서는 B → C → A.

바꾼 것: `gate::CheckRun.attempt: Option<u32>`(선택, 직렬화에서 빠짐, 기존 스냅샷·골든 불변), `gate_collect::run_attempt`(성공한 `ci`/`rc-ci` 집계마다 `GET /actions/runs/<id>` 한 번, 다른 실행·`details_url` 없는 실행·실패한 실행은 읽지 않음, 읽기 실패는 정보 없음), `gate::ci_info`(성공 상태에서 집계의 시도가 2 이상이면 `ci=success attempt=<n>`), `flow::run_id` `pub(crate)`.

AC-4 증거(RED→GREEN): RED 는 `tests/gate_collect.rs` 가 `no field attempt` 로 컴파일 실패, `tests/gate_check_cli.rs` 새 시험이 `attempt` 를 가진 스냅샷을 파싱하지 못해 실패. GREEN 뒤 새 시험 5개를 포함해 `gate_check_cli` 46, `gate_collect` 48, `gate_loop` 18, `gate_merge` 16, `gate_schedule_cli` 14, `gate_stacked_2859` 9, `flow_cli` 1 이 통과하고 `--lib -- gate flow` 45 가 통과해요(접두 시험 전부, 리더 안내 2026-11-19). 새 시험은 `ci_info_names_the_attempt_of_a_re_run_that_succeeded_and_the_verdict_does_not_move`, `a_ci_success_gets_its_attempt_from_the_workflow_run`, `a_run_without_a_details_url_has_no_attempt_and_no_extra_read`, `a_failed_workflow_run_read_leaves_the_attempt_empty_and_the_gate_unchanged`, `a_failed_ci_is_not_asked_for_its_attempt`). `tests.flow_metrics_test`(flow 접두의 파이썬 시험)는 PATH 의 `python3` 가 Xcode 3.9 라서 `datetime.fromisoformat('…Z')` 로 실패했고, pyenv 3.13.1 을 PATH 앞에 두면 7개가 통과해요(이 가지는 그 파일을 바꾸지 않았어요). 판정 불변(REQ-8)은 같은 스냅샷에서 `ci_info` 만 비운 응답 전체가 같음을 비교해요.
변이 5개(한 줄씩, 모두 시험이 잡음): M1 문턱 2→1, M2 성공이 아닌 상태에서도 표시, M3 번호 없음을 2 로 간주, M4 실패한 실행도 조회, M5 성공한 모든 실행 조회. 처음엔 M2·M4 가 살아남아 시험 둘(다른 잡 실패 사례, 실패한 `ci` 에 `details_url`)을 더한 뒤 잡혔어요.

**조각 A(문구) — 2026-11-20, 가지 `spec/20261119-flaky-followups-r2`(원격 이름; 조각 C 의 head `e428e9f5e2` 위에 쌓았어요).** REQ-1 은 글을 바꾸지 않았어요: C-108 조각 2 가 tip 에 `--ac` 를 이미 넣어서(커밋 040c8b49bd) 새 시험의 AC-1 은 RED 때부터 통과하고, 그 문장을 그대로 임시 원장에서 돌려 만든 카드가 기준 하나를 가진다는 것(`card show`)을 지켜요. 닫는 일(`card done`)은 병합된 PR 연결을 요구해서 이 시험 밖이에요.
- 바꾼 것: `skills/moai/workflows/mode-agent.md` 7번(heavy-checks) 끝에 한 문장 "Flaky repair (card C-99): …"(sleep·retry·skip 금지, 원인을 못 찾은 시험은 이유를 카드와 SPEC 에 적고 삭제), `skills/moai/workflows/verify.md` 의 간헐 시험 문단 앞에 한 문장 "Record home (card C-99): …"(카드 원장이 있는 프로젝트는 merge-train 단계의 flaky 카드가 기록 장소, `[flake]` Issue·두 명령은 카드 원장이 없는 프로젝트용), `source-manifest.json` 줄 둘(`source_manifest.py write` 가 쓴 값, 같은 커밋; manifest 예약은 약 1초 쥐고 풀었어요).
- RED→GREEN: `tests/flaky_followups.rs` 3개 중 둘이 RED(두 문장이 없어서), 문장 뒤 3/3. 변이 3개(`retry·skip` 삭제, 삭제 이유 삭제, 카드 원장 한정 삭제) 모두 시험이 잡았어요.
- 검사(head `f4cc4052c5`, 이후는 rustfmt 한 줄 모양 커밋과 C 병합·SPEC 문서): (1) `git diff --check` 0 · (2) `cargo fmt --check` 1 → `cargo fmt` 로 시험 파일 모양을 고친 뒤 0 · (3) `rustfmt --check --edition 2021 src/ops/*.rs` 0 · (4) clippy `--all-targets -D warnings` 0 · (5a) 5개 대상 0 · (5b) 파이썬 4개 0 · (5c) `--bins` 0 · (6) trailer-check 0 · 등재 검사 39개 전부 0. 두 문서를 읽는 시험 전부: `--test agent_defs_package --test claim_renew_doc_3042 --test claim_verify_cli --test claim_verify_flags_2838 --test close_step_2903 --test flaky_followups --test lane_next_rules --test orca_optional_3005 --test product_cli --test sdd_guidance --test test_isolation_rules` 0(164 통과), `tests.check_pr_self_audit_test tests.ci_scope_test tests.github_templates_test` 0(91 통과).

**조각 A 감사(head `dd0a677ed8`, 구현자가 아닌 새 읽기 전용 에이전트): FAIL, 막는 것 1.** B1 — SPEC 의 AC-1·AC-2 `verify:` 명령이 시험 이름과 달라 시험 0개를 골랐어요(AC-1 문구도 `card done` 까지라고 해서 시험과 달랐어요). 수리: `verify:` 이름을 실제 시험(`the_flaky_card_command_carries_an_ac_and_the_card_can_be_read_with_it`, `the_repair_rule_is_in_the_mode_agent_workflow`, `the_record_home_is_in_the_verify_workflow`)으로, AC-1 문구를 시험이 보이는 것(`card show` 로 읽힌 기준 하나; `card done` 은 병합된 PR 연결이 필요해 시험 밖)으로 고치고, 시험에 `SPEC-20261119-flaky-followups AC-<n>` 주석을 달았어요. N4 도 시험으로 닫았어요: 낱말 목록 대신 문장 전체를 핀해서 뜻이 뒤집힌 문장(`adds a sleep, a retry or a skip`, `is never deleted, without a reason`, `are for a project with a card ledger`)을 변이 3개로 넣었더니 모두 시험이 잡았어요.
비차단으로 남긴 것(문구 수리는 `verify.md`·`mode-agent.md` 와 manifest 를 다시 잡아야 해서 이번에 하지 않아요 — 필요하면 별도 한 조각): N1 `verify.md` 의 두 문장이 같은 말을 되풀이하고 "the two commands below" 가 분류만 하는 `flake-classify` 까지 카드 원장 없는 프로젝트용으로 묶어요. N2 "the merge-train step" 이 MODE-LEADER 의 단계(`mode-leader.md:32`)라는 말이 없어요. N3 `mode-agent.md` 7번 끝에 앞 문장과 이어지지 않고 덧붙은 느낌이고 "deleted"(여기)와 "retired"(`mode-leader.md:32` 의 기준)가 낱말이 달라요; 삭제 규칙은 flaky 시험에 한한 예외라는 말이 문장에 없어요(`core-rules` 4번과 약한 긴장). N5 `Moai-Session` 값이 감사를 맡긴 세션과 같다는 점은 형식상 사실이고, 감사자는 컨텍스트가 분리된 새 에이전트예요.

남은 것: 조각 B(`card list --title-prefix`: `card::command::list_filtered` 와 시험 4개는 로컬 가지 `keep/c99-3b-old` 에 있고, CLI 연결 32줄은 `src/ops/card.rs`(C-125 보유)가 풀리면 해요). 조각 A 는 위에 적은 대로 끝났어요.

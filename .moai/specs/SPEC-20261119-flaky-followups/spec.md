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

- AC-1 (REQ-1): `mode-leader.md` 의 간헐 규칙 명령이 `--ac` 한 줄을 갖고, 그 문장 그대로 `moai card add` 를 임시 원장에서 돌려 만든 카드가 기준 하나를 갖고 충족 표시 뒤 `card done` 이 닫아요.
  verify: `cargo test --locked --offline --test flaky_followups the_flaky_card_command_carries_an_ac_and_the_card_can_close`
- AC-2 (REQ-2, REQ-3): `mode-agent.md` 에 sleep·재시도·건너뛰기 금지와 삭제 규칙 문장이, `verify.md` 에 기록 장소 문장이 있고 `[flake]` Issue 경로는 "원장이 없는 프로젝트"로 한정돼요(이름 붙인 파일만 읽는 문구 핀).
  verify: `cargo test --locked --offline --test flaky_followups the_repair_rule_and_the_record_home_are_in_their_workflows`
- AC-3 (REQ-4, REQ-5): `--title-prefix "flaky: "` 가 접두어가 맞는 카드만, 빈 값과 제어 문자는 거부(원장을 읽기 전), 옵션이 없으면 전과 같은 출력이에요.
  verify: `cargo test --locked --offline --test card_list_title_prefix`
- AC-4 (REQ-6, REQ-7, REQ-8): 스냅샷에 시도 번호가 있는 성공 `ci` 가 `attempt=<n>` 을 정보 줄에 더하고 G5 판정은 같아요; 시도 번호 없는 스냅샷은 줄이 전과 같아요(기존 골든 불변).
  verify: `cargo test --locked --offline --lib gate::tests::ci_info`
  verify: `cargo test --locked --offline --test gate_cli`
- AC-5 (REQ-9, REQ-10): 크기 예산·위치 핀·manifest 검사가 통과해요.
  verify: `cargo test --locked --offline --test workflow_size_budget`
  verify: `python3 -m unittest tests.source_manifest_test`
- AC-6: 이웃 검사(대상 한정; 전량은 PR CI run 이 인용).
  verify: `cargo test --locked --offline --test product_cli skill_package`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-flaky-followups/**`; 조각 A(문구): `skills/moai/workflows/mode-leader.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `tests/flaky_followups.rs`; 조각 B(필터): `src/card/command.rs`, `src/main.rs`, `src/help.rs`, `docs/rust-product-path.md`, `tests/card_list_title_prefix.rs`; 조각 C(재실행 신호): `src/gate.rs`, `tests/gate_cli.rs`(+ 수집기가 읽는 파일은 조사 뒤에 확정)
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

(조각 1 은 계획과 SPEC 만이에요. 조각 A~C 뒤에 채워요.)

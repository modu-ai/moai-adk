---
id: SPEC-20261118-spec-close
title: "moai spec close — close a SPEC whose work is merged"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-sdd-foundation
---

# moai spec close — close a SPEC whose work is merged

## Intent

SDD 조각 SDD-8(승인된 SDD 계획 `sdd-lifecycle-plan-20261117` "Slices"): 일이 합쳐진 SPEC을 한 명령으로 닫아요. SPEC이 승인돼 있고,
그 일이 base에 합쳐졌고, 상태를 `completed`로 바꾼 SPEC이 `spec check --closure`를 통과할 때만 상태 한 줄을 바꿔요. 원하면 연결된 카드를
기존 `card done` 길로 닫아요. git 커밋·push는 하지 않고 커밋 명령을 미리 보여 줘요. 거절은 아무것도 바꾸지 않아요.

## Non-goals

- 승인 기록 쓰기(사용자의 일), SPEC 보관(archive), 릴리스 노트.
- git 커밋·push, PR 만들기·합치기.

## Requirements

- REQ-1: When the user closes a SPEC, `moai spec close` shall check that the SPEC exists, has the new id form, has no error and has status `approved`.
- REQ-2 (REQ-1): IF the SPEC is missing, has a Go-generation id, has an error or does not have status `approved` (a draft waits for the user's approval), THEN the command shall refuse and change nothing.
- REQ-3: When the SPEC is approved, the command shall check that its work is merged: the SPEC's local branch, when it exists, is reachable from the base, and a commit reachable from the base carries its `Moai-Spec` trailer.
- REQ-4 (REQ-3): IF the base is not a commit, the branch is not reachable from the base, or no commit reachable from the base carries the trailer, THEN the command shall refuse and change nothing.
- REQ-5: When the work is merged, the command shall change the SPEC's status line to `completed` and nothing else in the file.
- REQ-6 (REQ-5): IF the SPEC with status `completed` would not pass `spec check --closure`, THEN the command shall refuse and change nothing.
- REQ-7: Where `--card` names a card, the command shall close that card through `card done` before it writes the SPEC.
- REQ-8 (REQ-7): IF the card names no SPEC or another SPEC, or `card done` refuses, THEN the command shall refuse with that reason and change nothing.
- REQ-9: The command shall print the commit command for the SPEC file as a preview without running it.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5, REQ-9): 합쳐진 SPEC의 상태 줄만 `completed`로 바뀌고, 결과 파일이 `spec check --closure`를 통과하며, 커밋 명령이 미리 보여요.
  verify: `cargo test --locked --offline --test spec_close_cli close_completes_a_merged_spec_and_prints_the_commit`
- AC-2 (REQ-2, REQ-4, REQ-6): 없는 SPEC, 옛 id, 오류 있는 SPEC, draft, 이미 completed, 커밋이 아닌 base, 합쳐지지 않은 가지, trailer 커밋 없음,
  Result 없음·열린 AC는 거절되고 임시 HOME 아래 모든 바이트가 그대로예요.
  verify: `cargo test --locked --offline --test spec_close_cli close_refuses_and_changes_nothing`
- AC-3 (REQ-7, REQ-8): `--card`는 카드를 `done`으로 닫고 SPEC을 써요; 다른 SPEC의 카드, SPEC 없는 카드, `card done`의 거절(합쳐진 PR 없음)은
  아무것도 바꾸지 않아요.
  verify: `cargo test --locked --offline --test spec_close_cli close_with_a_card_marks_it_done`; `cargo test --locked --offline --test spec_close_cli a_card_refusal_changes_nothing`
- AC-4 (REQ-1…REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test spec_take_cli`; `cargo test --locked --offline --test card_spec_cli`;
  `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test sdd_guidance`;
  `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/spec_close.rs`, `src/spec_take.rs`, `src/lib.rs`, `src/ops/spec.rs`, `tests/spec_close_cli.rs`, `tests/sdd_guidance.rs`,
  `tests/fixtures/help/*.txt`, `docs/design-review/appendix/operations/spec.md`, `skills/moai/references/internal/sdd.md`,
  `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261118-spec-close/spec.md`
- read-only: `src/spec_doc.rs`, `src/card/work.rs`, `src/card/spec_link.rs`, `src/card/dispatch.rs`, `src/apply.rs`, `src/trailers.rs`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소

## Plan

1. 시험 먼저: 임시 저장소(SPEC 가지를 main에 merge), 임시 HOME 원장, 가짜 `gh`로 카드를 `review`까지 옮겨 CLI 시험.
2. `src/spec_close.rs`: 검사(읽기만) → 카드 → SPEC 파일 원자적 교체. git은 `spec_take`와 같은 도우미.
3. 도움말 골든, op 문서, `internal/sdd.md` 한 줄, `sdd_guidance` 목록, skill 목록.

## Risks

- squash merge는 가지를 base의 조상으로 만들지 않아요 → 로컬 가지가 남아 있으면 거절되고, 메시지가 가지를 지우거나 `--base`를 보라고 말해요.
  squash 메시지가 `Moai-Spec` trailer를 버리면 trailer 검사도 실패해요.
- `--card` 뒤에 SPEC 쓰기가 실패하면(파일이 그 사이 바뀜) 카드는 이미 `done`이에요 → 메시지가 그 사실을 말해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 닫힌 상태는 새 낱말 없이 `spec_doc::STATUSES`의 `completed` | 코디네이터: 있는 어휘를 쓰고 만들지 않음; `--closure`가 `completed`를 요구 | 상태 낱말 추가 |
| D-2 | base 기본값은 `spec take`가 기록한 `branch.<가지>.moai-base`, 없으면 지금 가지; `--base`로 바꿈 | `spec take`와 같은 base | 늘 `--base` 요구 |
| D-3 | 합쳐짐 = 로컬 가지가 있으면 base의 조상 + base에서 닿는 커밋 하나가 `Moai-Spec: <id>` trailer를 가짐 | 가지를 지운 뒤에도 판단 가능; 빈 가지로 속지 않음 | 가지 조상만 보기 |
| D-4 | `--card`는 SPEC을 쓰기 전에 `card done`(그 거절 그대로)으로 닫음; 카드가 이 SPEC을 가리켜야 함 | 거절은 아무것도 바꾸지 않음; `card done`은 한 거래 | SPEC 먼저 쓰기 |
| D-5 | 이미 `completed`인 SPEC은 거절(아무 일 없는 성공이 아님) | 두 번 닫기는 실수일 가능성이 큼 | 변화 없는 성공 |
| D-6 | SPEC 파일은 읽은 바이트 그대로일 때만 원자적으로 바꿈(`apply::publish_replace`) | 그 사이 편집을 덮어쓰지 않음 | 그냥 쓰기 |

## Result

- AC-1: met — `spec_close_cli` (RED 먼저: 4 failed, `close` 없음 → usage 오류); 바뀐 바이트는 상태 줄 하나, 결과는 `--closure` 통과, HEAD 그대로.
- AC-2: met — 같은 시험, 거절 10가지(`-`로 시작하는 기록된 base 포함)에서 임시 HOME 아래 모든 바이트가 그대로. 변이 6개 중 5개 KILLED
  (C2는 시험을 고친 뒤 KILLED); C6(base의 `-` 검사 끄기)은 `rev-parse --end-of-options`가 같은 입력을 거절하는 동등 변이.
- AC-3: met — 같은 시험 두 개: 가짜 `gh`로 PR을 합친 카드가 `done`; 열린 PR(`card done`의 거절), SPEC 없는 카드, 다른 SPEC의 카드, 세션 없음은 아무것도 바꾸지 않음.
- AC-4: met — spec_take_cli 9, spec_cli 17, card_spec_cli 11, help_layout 10, product_cli 76, sdd_guidance 6(목록에 `close`), home_isolation_scan 16,
  git_env_scrub 3, ac_command_lint 5, `--lib` 1418, `--bin moai` 27, clippy·fmt 0, Python 39·19·25·5 OK, manifest check 0.

---
id: SPEC-20261118-spec-take
title: "moai spec take — start work on a SPEC"
status: draft
revision: 1
created: 2026-11-18
parent: SPEC-20261117-sdd-foundation
---

# moai spec take — start work on a SPEC

## Intent

SDD 조각 SDD-7(승인된 SDD 계획 `sdd-lifecycle-plan-20261117` "Slices"): 한 명령으로 SPEC 작업을 시작해요. SPEC이 있고 `spec check`를
통과하는지 보고, 그 SPEC의 가지 `spec/<yyyymmdd>-<slug>`를 base에서 만들거나 그 가지로 옮기고, 원하면 카드를 SPEC에 잇고(SDD-4),
다음 명령(trailer, push, draft PR)을 미리 보여 줘요. push와 PR은 바깥 쓰기라 기본은 보여 주기만 하고, `--push`·`--draft-pr`을 줄 때만
실행해요. 팀에서 누가 SPEC을 맡았는지는 그 가지의 이른 draft PR로 보여요(결정 2026-11-17).

## Non-goals

- SPEC 승인 기록 쓰기, `spec close`(SDD-8), worktree 만들기.
- PR 본문의 AC 표 만들기(SDD-6 이후 조각).

## Requirements

- REQ-1: When the user takes a SPEC, `moai spec take` shall check that the SPEC exists in the project, has the new id form and has no error.
- REQ-2 (REQ-1): IF the SPEC is missing, has a Go-generation id or has an error, THEN the command shall refuse and change nothing.
- REQ-3: When the SPEC passes, the command shall create the branch `spec/<yyyymmdd>-<slug>` from the base (`--base`, else the current HEAD) and switch to it, or switch to it when it exists. · failure: none — 실패 경로는 REQ-4·REQ-5예요.
- REQ-4 (REQ-3): IF the work tree has a change outside the SPEC's own folder, THEN the command shall refuse and change nothing.
- REQ-5 (REQ-3): IF the branch exists and a commit on it carries a `Moai-Spec` trailer for another SPEC, THEN the command shall refuse and change nothing.
- REQ-6: Where `--card` names a card, the command shall link that card to the SPEC through the local ledger as `moai card spec` does.
- REQ-7: The command shall print the next commands — `moai trailers --spec`, the push and the draft pull request — without running them.
- REQ-8: Where `--push` is given, the command shall push the branch to the remote.
- REQ-9 (REQ-8): IF `--draft-pr` is given without `--push`, THEN the command shall refuse before any write.
- REQ-10: Where `--draft-pr` is given with `--push`, the command shall open a draft pull request with `gh` after the push.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-7): 임시 저장소에서 가지를 base에서 만들고 옮기며, 다시 실행하면 그대로 옮기기만 하고, 다음 명령을 보여 줘요.
  verify: `cargo test --locked --offline --test spec_take_cli take_creates_the_branch_and_prints_the_next_commands`
- AC-2 (REQ-2, REQ-4, REQ-5, REQ-9): 없는 SPEC, 옛 id, 오류 있는 SPEC, 더러운 트리, 다른 SPEC의 가지, push 없는 draft PR은 거절되고 아무것도 바뀌지 않아요.
  verify: `cargo test --locked --offline --test spec_take_cli take_refuses_and_changes_nothing`
- AC-3 (REQ-6): `--card`가 임시 HOME 원장에서 카드를 SPEC에 이어요.
  verify: `cargo test --locked --offline --test spec_take_cli take_links_the_card_in_the_local_ledger`
- AC-4 (REQ-8, REQ-10): `--push`는 임시 bare 저장소로 push하고, `--draft-pr`은 가짜 `gh`를 정해진 인자로 불러요(네트워크 없음).
  verify: `cargo test --locked --offline --test spec_take_cli push_and_draft_pr_run_only_when_asked`
- AC-5 (REQ-1…REQ-10): 이웃 검사.
  verify: `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test help_layout`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test git_env_scrub`;
  `cargo test --locked --offline --bin moai`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/spec_take.rs`, `src/lib.rs`, `src/ops/spec.rs`, `tests/spec_take_cli.rs`, `tests/fixtures/help/*.txt`,
  `docs/design-review/appendix/operations/spec.md`, `skills/moai/references/internal/sdd.md`,
  `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261118-spec-take/spec.md`
- read-only: `src/spec_doc.rs`, `src/card/spec_link.rs`, `src/identity.rs`, `src/forge/transport.rs`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소

## Plan

1. 시험 먼저: 임시 저장소·bare 원격·가짜 `gh`·임시 HOME으로 CLI 시험.
2. `src/spec_take.rs`(검사, git은 `scrub_git_env`로 호출자의 `GIT_*`를 지우고), `spec take` 동작.
3. 도움말 골든, op 문서, `internal/sdd.md` 한 줄, skill 목록.

## Risks

- 새로 만든 SPEC 파일은 아직 커밋되지 않아 트리가 더러워요 → SPEC 자기 폴더의 변화는 허용하고 가지로 함께 옮겨요(D3).
- `gh pr create`의 인자는 gh CLI 기준이에요 → 시험은 가짜 `gh`가 받은 인자를 확인할 뿐, 실제 GitHub 동작은 재지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | base 기본값은 지금 HEAD(통합 가지에서 실행한다는 m-flow), `--base`로 바꿈 | 프로젝트마다 통합 가지 이름이 다름 | 프로필의 `work_flow.integration_branch` 읽기 |
| D-2 | 가지가 있으면 옮기기만 해요(다시 실행해도 안전); 그 가지의 `Moai-Spec`이 다른 SPEC을 가리키면 거절 | 한 SPEC = 한 가지 | 있으면 늘 거절 |
| D-3 | 더러운 트리는 거절하되 SPEC 자기 폴더의 변화는 허용 | `spec new` 직후 `spec take` 흐름 | 모든 변화 거절 |
| D-4 | push·draft PR은 `--push`·`--draft-pr`을 줄 때만; `--draft-pr`은 `--push`가 필요 | 바깥 쓰기는 사용자 선택(case 2) | 기본 실행 |
| D-5 | PR 제목은 SPEC 제목, 본문 첫 줄은 `SPEC: <id>`, 카드 번호는 넣지 않음 | 카드 번호는 공유 장소에 쓰지 않음 | 본문 바꾸기 |
| D-6 | git은 `audit_delta::command`(전역 설정 무시, 모든 프로토콜 금지)가 아니라 `scrub_git_env` + 사용자 자신의 설정 + `GIT_TERMINAL_PROMPT=0` | 실제 push에는 사용자의 자격 증명 도우미·URL 바꿈이, switch에는 줄끝 설정이 필요; 첫 구현은 `transport 'file' not allowed`로 push가 실패 | `audit_delta::command`에 프로토콜만 열기 |

## Result

- AC-1: met — `spec_take_cli` 4 passed (RED 먼저: 4 failed, `take` 없음 → usage 오류).
- AC-2: met — 같은 시험; 첫 GREEN에서 상태 첫 줄의 앞 공백이 잘려 경로가 `.txt`로 읽히는 버그를 찾아 `-z` 읽기로 고침.
- AC-3: met — 같은 시험, `spec show`의 `local_cards.count == 1`.
- AC-4: met — 같은 시험, 임시 bare 원격과 가짜 `gh`.
- AC-5: met — spec_cli 17, help_layout 10, product_cli 76, git_env_scrub 3, card_spec_cli 11, `--lib` 1417, `--bin moai` 27, clippy·fmt 0.

---
id: SPEC-20261119-verified-trailer
title: "Every commit says how it was verified (Moai-Verified trailer)"
status: draft
revision: 1
created: 2026-11-19
---

# Every commit says how it was verified (Moai-Verified trailer)

## Intent

레인 커밋이 어떻게 검증됐는지 커밋 자체에 새겨요(리더 ORIGIN note 2337; Addy Osmani, "How the Claude Code team thinks about verification", claude.dev 2026-11-16, 표 2의 C — "Each commit is stamped with how it was verified, including none"). 식별 트레일러(`Moai-Spec`, `Moai-Session`, 2026-11-17) 옆에 `Moai-Verified: <tests|mutants|e2e|smoke|none>[, …]` 한 줄이 더 실려요: `moai trailers` 가 그 줄을 인쇄하고, `moai trailer-check` 가 값을 검사하고, `moai trailers --list` 가 커밋을 종류별로 나열해 `none` 커밋과 되돌려진(revert) 커밋을 함께 셀 수 있게 해요. `none` 도 정직한 값이라서 아무 커밋이나 한 줄로 검증 상태를 밝히게 돼요.

## Non-goals

- 검증 증명(어떤 시험·몇 개)의 기록 — 종류 이름만; 근거는 PR 본문·SPEC Result 에 남아요.
- 원격 저장소·CI 결과 읽기, 트레일러 강제 커밋 훅, 과거 커밋 다시쓰기.
- `Moai-Verified` 를 사람 커밋에 요구하기 — 규칙은 시작 커밋 이후의 커밋에만 묻고, 그 전은 `none` 으로 읽어요.

## Requirements

- REQ-1: When a lane prints its commit trailers with `moai trailers --verified <kinds>`, the command shall append one `Moai-Verified: <kinds>` line that lists only the accepted kinds (`tests`, `mutants`, `e2e`, `smoke`, `none`), comma-separated in the given order.
- REQ-2 (REQ-1): IF a given kind is outside the accepted set, THEN `moai trailers` shall refuse with an error and print no trailer text.
- REQ-3: When `moai trailer-check` reads a commit that carries a `Moai-Verified` line, it shall read the commit's verification as the listed kinds.
- REQ-4 (REQ-3): IF a listed kind is not accepted, THEN `moai trailer-check` shall fail the commit.
- REQ-5: When `moai trailer-check` reads a commit after the rule's start commit that carries no `Moai-Verified` line, it shall report `missing_verified` and fail the commit. · failure: none — 조건 자체가 실패 보고라 덮을 미원 동작이 없어요.
- REQ-6: When `moai trailer-check` reads a commit before the rule's start commit that carries no `Moai-Verified` line, it shall read the commit as `none` and report no finding. · failure: none — 면제 경로에는 거부가 없어요.
- REQ-7: Where the checked range names no start (no commit of the range carries the line and `--verified-since` is absent), the command shall treat every commit of the range as before the start.
- REQ-8: Where `--verified-since` names a revision, the command shall take that revision's commit as the rule's start in place of the range's first carrying commit.
- REQ-9: When `moai trailers --list` reads a range, it shall list each commit with its verification kinds, group the commits by kind (a commit with several kinds appears in each), and mark a commit that a later commit's standard revert message names. · failure: none — 나열은 읽기 전용이라 거부 경로가 없어요.
- REQ-10: While a `Moai-Verified` value holds the log's field separator or a control character, `moai trailer-check` shall treat the commit as malformed and fail it.

## Acceptance

- AC-1 (REQ-1, REQ-2): `moai trailers --spec <id> --verified tests mutants` 처럼 여러 값을 인쇄에 실고, 알 수 없는 값(`tset`)은 거부하며 아무 텍스트도 내지 않아요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailers_prints_the_verified_line`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): trailer-check 가 인정 값을 받고 오타를 실패시켜요; 시작 커밋(범위 안 첫 `Moai-Verified` 커밋) 이전의 빠진 줄은 `none` 으로 통과, 이후의 빠진 줄은 `missing_verified` 실패; 범위에 한 번도 없으면 모두 통과; `--verified-since` 가 시작을 정해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailer_check_cuts_at_the_rule_start`
- AC-3 (REQ-9): `--list` 가 커밋을 종류별로 묶어 나열하고 `This reverts commit <sha>` 메시지가 되돌린 커밋을 표시해요.
  verify: `cargo test --locked --offline --test trailer_verified_cli trailers_list_groups_and_marks_reverts`
- AC-4 (REQ-1…REQ-10): 이웃 검사 — 단위 시험(`--lib trailers::`), `cargo fmt --check`, fmt-ops(`rustfmt --check --edition 2021 src/ops/*.rs`: build.rs 가 src/ops 모듈 목록을 만들어 cargo fmt 는 그곳에 닿지 않음), `cargo clippy --offline --locked --all-targets -- -D warnings`, 기존 trailer 시험 회귀(`--test spec_close_cli` 는 trailer 커밋을 만드는 대표 이웃).
  verify: `cargo test --locked --offline --lib trailers::`
  verify: `rustfmt --check --edition 2021 src/ops/*.rs`

## Scope

- write: `src/trailers.rs`, `src/ops/trailers.rs`, `src/ops/trailer_check.rs`, `tests/trailer_verified_cli.rs`, `.moai/specs/SPEC-20261119-verified-trailer/spec.md`
- read-only: `src/main.rs`(배선은 build.rs 가 `src/ops/*` 에서 생성 — 수정 불필요), `docs/design-review/appendix/operations/trailers.md`·`trailer_check.md`(도움말 갱신은 카드 범위 밖 — 리더 후속), `src/audit_delta.rs`(git 호출 스크럽)
- forbidden: 실제 HOME 과 원장, 네트워크, 실제 원격 저장소, 키

## Plan

1. RED 먼저: `tests/trailer_verified_cli.rs` — 임시 저장소·임시 HOME; (1) `--verified` 인쇄·거부, (2) trailer-check 컷오프(이전 통과/이후 실패/오타 실패/`--verified-since`), (3) `--list` 묶음·revert 표시. 명령·플래그가 없어 먼저 실패해요.
2. `src/trailers.rs`: 인정 값 상수·`verified` 필드 파싱·`judge` 의 세 finding(`missing_verified`·`malformed_verified`·`malformed_trailers`)·시작 커밋 규칙(범위 안 첫 실은 커밋 또는 `--verified-since`)·`--list` 나열.
3. `src/ops/trailers.rs`: `--verified <kinds>…`(값 검증)와 `--list [--range] [--dir]`(읽기 전용); `src/ops/trailer_check.rs`: `--verified-since <rev>`.
4. 이 가지의 커밋부터 `Moai-Verified: tests` 등을 실어 규칙을 스스로 따름.

## Risks

- 시작 커밋의 자기표시 규칙(범위 안 첫 `Moai-Verified` 커밋)은 범위 안에서 일부러 먼 옛 커밋에 줄을 실어 이후 커밋을 면제로 만들 수 있어요 — 감사가 범위 전체를 읽으므로 눈에 띄고, `--verified-since` 로 리더가 고정할 수 있어요.
- 되돌림 표시는 표준 git revert 메시지(`This reverts commit <sha>`)만 읽어요 — 손으로 쓴 되돌림은 못 봐요(문서에 밝힘).
- 도움말 문서 조각(`docs/design-review/appendix/operations/`) 갱신이 카드 범위 밖이에요 — 리더 후속으로 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙 시작 커밋 = 검사 범위 안에서 `Moai-Verified` 를 처음 실은 커밋; `--verified-since <rev>` 가 대신 정함 | 어떤 저장소에서든 성립(설정·상수 불필요), 채택 순간이 스스로 표시됨 | 설정 파일·상수로 시작 지점 기록 |
| D-2 | `--list` 의 되돌림 표시는 표준 revert 메시지의 sha 대조(짧은 접두 일치) | git revert 가 남기는 표준 문구; 추가 데이터 없음 | 되돌림 감지 제거 |
| D-3 | 값 어휘는 카드가 정한 다섯 종류뿐, 쉼표 나열 허용, 순서 보존 | 카드 기준 1; `none` 도 정직한 값 | 어휘 확장 |

## Result

- AC-1: met — 2026-11-19 이 worktree에서 `cargo test --locked --offline --test trailer_verified_cli` 3 passed 0 failed(`trailers_prints_the_verified_line` 포함: 인쇄·`none`·미지정 시 부재·오타 거부), `--lib trailers::` 8 passed 0 failed(`format_appends_the_verified_line_or_refuses_the_unknown_kind` 포함).
- AC-2: met — 같은 실행의 `trailer_check_cuts_at_the_rule_start` 통과: 시작 경계(범위 첫 실은 커밋 c2 전 c1 면제·c2 이후 c3 실패, `c2..` 범위의 자체 시작 c4, `--verified-since` 의 시작 자신 면제·범위 밖 시작 0), 오타 `tset` 실패, 인정 값·쉼표 나열 통과.
- AC-3: met — `trailers_list_groups_and_marks_reverts` 통과: 종류별 묶음(다중 종류 각 묶음 등장), 빠진 줄 `none` 묶임, `This reverts commit` 표시·개수.
- AC-4: met — `cargo fmt --check` exit 0, `rustfmt --check --edition 2021 src/ops/*.rs` exit 0(수리 감사 B1: 110컬럼 호출을 인자마다 줄로 감쌈), `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, 이웃 `--test spec_close_cli` 8 passed(format 호출부 무변경 — `format` 은 3인자 그대로, 새 `format_verified` 위임); 변이 2 KILLED(시작 면제 제거 at+1→at → 컷오프 시험 exit 101; 쉼표 분할 제거 → 2 실패 exit 101), 복원 후 재검증 모두 exit 0.

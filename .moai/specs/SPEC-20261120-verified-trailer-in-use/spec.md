---
id: SPEC-20261120-verified-trailer-in-use
title: "The Moai-Verified trailer is written in the guidance, checked by the local check set and summarised in the bundle body"
status: draft
revision: 1
created: 2026-11-20
---

# The Moai-Verified trailer is written in the guidance, checked by the local check set and summarised in the bundle body

## Intent

`Moai-Verified: <tests|mutants|e2e|smoke|none>[, …]` 트레일러(SPEC-20261119-verified-trailer)는 코드로는 있어요: `moai trailers --verified` 가 줄을 찍고, `moai trailer-check` 가 값과 빠진 줄을 검사하고, `moai trailers --list` 가 종류별로 모아 줘요. 그런데 아무도 쓰지 않아요. 2026-11-20 이 기준 가지(`1c9f1a1f95`)에서 직접 센 결과, 지침 파일 `skills/moai/references/internal/commit-trailers.md`, `agents/impl.md`, `AGENTS.md`, `REVIEW.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/sync.md` 어디에도 이름이 없고(`grep -c Moai-Verified`), `moai trailer-check` 를 부르는 곳은 `commit-trailers.md` 의 안내 문장뿐이에요. 줄이 없는 커밋이 대부분이어서 "검증 안 된 커밋을 조회"할 수 없어요.

이 SPEC 은 세 곳에서 트레일러가 실제로 쓰이게 해요: (1) 지침이 트레일러와 다섯 값을 말해서 에이전트가 쓰게 하고, (2) 로컬 검사 모음(`tools/local_check.sh`)이 `trailer-check` 로 규칙 시작 커밋 이후의 빠진 줄을 실패시키고, (3) 묶음 PR 본문이 `trailers --list` 의 한 줄 요약(종류별 수, `none` 수, 되돌린 수)을 실어요. 근거: 사용자 결정 2026-11-20(리더 대화, 검증 2차 묶음 W2-3). 원 글은 "모든 커밋에 어떻게 검증했는지 표시한다, `none` 도 값이다"(Addy Osmani, "How the Claude Code team thinks about verification", 2026-11-16).

규칙 시작 커밋의 뜻은 기존 SPEC 그대로예요: 검사 범위 안에서 이 줄을 처음 실은 커밋이고, 그 앞 커밋은 `none` 으로 읽혀 면제돼요. 그래서 (2)는 줄을 한 번도 쓰지 않은 가지에는 아무것도 요구하지 않고, 쓰기 시작한 뒤에만 빠뜨림을 잡아요(이 단계가 그 규칙을 바꾸지 않아요).

## Non-goals

- `trailers`·`trailer-check` 의 동작과 값 어휘 변경: 이미 SPEC-20261119-verified-trailer 가 정했어요.
- 검증 증명(어떤 시험, 몇 개)의 기록: 종류 이름만이고 근거는 PR 본문과 SPEC Result 에 남아요.
- 과거 커밋 다시쓰기, 커밋 훅으로 강제하기, 원격 저장소 읽기.
- `AGENTS.md`, `REVIEW.md`, `mode-agent.md`, `sync.md` 의 문장 추가: 카드 범위는 `commit-trailers.md` 와 `agents/impl.md` 두 지침이에요(나머지는 그 둘을 읽는 곳이라 따로 적지 않아요).
- 새 크레이트와 모델 호출: 이 검사는 결정적이고 네트워크를 쓰지 않아요.

## Requirements

- REQ-1: The reference `internal/commit-trailers` shall name the `Moai-Verified` trailer with its five values (`tests`, `mutants`, `e2e`, `smoke`, `none`), the command `moai trailers --spec <SPEC-id> --verified <kinds>` that prints the line and the command `moai trailers --list` that groups the commits by kind.
- REQ-2: The agent definition `agents/impl.md` shall tell the agent to end each commit with the `Moai-Verified` line that `moai trailers --verified` prints and to name only checks it ran.
- REQ-3: The local check set `tools/local_check.sh` shall run `trailer-check` over the commits of the head beyond the base as one light step, listed after `diff-check`, in both modes and in every leg that runs the rest of the set.
- REQ-4 (REQ-3): IF `trailer-check` reports a failing commit, THEN the step shall fail with the exit code of the command, which makes the result line say fail.
- REQ-5: When `moai bundle body` writes a body, it shall add one line that gives the commit count per `Moai-Verified` kind (a commit without the line counts under `none`, as `moai trailers --list` reads it) and the count of reverted commits, over the commits of the merge base of the tip and the branch up to the pinned head.
- REQ-6 (REQ-5): IF the by-kind listing of that range cannot be read or lacks a count, THEN `moai bundle body` shall refuse with a message that names the failure and write no body.
- REQ-7 (REQ-5): The summary line shall hold only counts and the fixed words, never a card number, a session id or a commit message.

## Acceptance

- AC-1 (REQ-1): the reference holds the trailer name, the five values, both commands and the sentence that `none` is a value · verify: `cargo test --locked --offline --test verified_trailer_in_use_cli commit_trailers_reference`
- AC-2 (REQ-2): the agent definition holds the trailer name, the command and the "only checks it ran" sentence · verify: `cargo test --locked --offline --test verified_trailer_in_use_cli impl_agent`
- AC-3 (REQ-3, REQ-4): a dry run of `tools/local_check.sh` lists the step `trailer-check` right after `diff-check` with the range of the head beyond the base, and a fake binary that exits 1 fails the step and the result line · verify: `cargo test --locked --offline --test verified_trailer_in_use_cli local_check`
- AC-4 (REQ-5, REQ-7): the body of a bundle whose branch holds commits of the kinds `tests`, `mutants` with `e2e`, and none, with one commit reverted, has exactly one summary line with those counts, and the line holds no card number or session id · verify: `cargo test --locked --offline --test bundle_body_cli verified_summary_counts`
- AC-5 (REQ-6): a listing that lacks a count, or an error from the listing, gives a refusal that names the failure, in the function that builds the line (the real command cannot make the range unreadable, so the check is a unit test) · verify: `cargo test --locked --offline --lib bundle::tests::verified_summary`
- AC-6 (REQ-1…REQ-7): the lines of the body that exist today, the manifest digest of the skill package and the neighbour tests stay as they were · verify: `cargo test --locked --offline --test bundle_body_cli`

## Scope

- write: `src/bundle.rs`, `tests/bundle_body_cli.rs`, `tests/verified_trailer_in_use_cli.rs`, `.moai/specs/SPEC-20261120-verified-trailer-in-use/**`; after the holding cards return them: `skills/moai/references/internal/commit-trailers.md`, `agents/impl.md`, `tools/local_check.sh`, `tests/local_check_test.py`, `skills/moai/references/source-manifest.json`
- read-only: `src/trailers.rs` (`list`), `src/ops/trailers.rs`, `src/ops/trailer_check.rs`, `.github/workflows/ci.yml`
- forbidden: `skills/**` other than the two named files, other projects' hook files, the real HOME and ledger, the network

## Plan

Tests first (RED) for every part, but the three parts depend on files that other cards hold today (`commit-trailers.md` by one card, `agents/impl.md` by another, `tools/local_check.sh` and its step-list test by a third, the manifest by two more). The part with free files comes first and the others follow as the files return, each by `card scope` (it only widens), never by writing under another card's reservation.

1. Bundle summary (REQ-5…REQ-7): RED in `tests/bundle_body_cli.rs`: a temporary repository with commits `tests`, `mutants, e2e`, none and a standard revert; the body must hold one line with the counts. GREEN in `src/bundle.rs`: call `trailers::list(dir, "<base>..<pinned>")` in `body`, add a missing fact when it fails, print the line after the integration paragraph. Mutants (at most 5): a count off by one, the none count, the reverted count, the range start, the refusal.
2. Local step (REQ-3, REQ-4), once `tools/local_check.sh` and `tests/local_check_test.py` are free: a `light` step `trailer-check` after `diff-check`, run with the binary the set already built (`cargo run --offline --locked --quiet --bin moai-cli --`) over `<base>..<head>` with `--branch <the working folder's branch>` in candidate mode, where the checkout is detached; `tests/local_check_test.py` pins the step list, so its pin moves with the step. The step goes in the leg that already builds the binary.
3. Guidance (REQ-1, REQ-2), once the two guidance files are free: one paragraph in each, the pin test reads both files from the repository and checks the fixed sentences; the manifest digest moves with `skills/` edits (`source-manifest.json`).
4. Neighbours: `--test bundle_body_cli`, `--test trailer_verified_cli`, `--test agent_defs_package`, `--test local_check_test` through python unittest, fmt, clippy, `trailer-check`.

Research notes (read 2026-11-20, in `1c9f1a1f95`): `src/trailers.rs` (`list`, `VERIFIED_KINDS`), `src/ops/trailer_check.rs` (flags `--branch`, `--verified-since`; exit 1 on a failing commit), `src/bundle.rs` (`body`, `judge_all` already runs a `trailer_check` check per head), `tools/local_check.sh` (`step`, `diff_range`, legs `nextest` and `rest`), `tests/bundle_body_cli.rs` (the fixture model).

## Risks

- Another card may rewrite `tools/local_check.sh` (parallel python tests) before this lane gets it: the step is one `step` call and one pin, so a merge is small; the lane re-reads the file at the moment it takes it.
- A CI checkout that is too shallow cannot resolve the base: the range must come from the same base the other steps use, and a fetch failure already fails the set before any step.
- `trailer-check` in the `rest` leg needs a built binary: the step goes in the leg that builds it, and the dry-run test pins the leg.
- A body line that counts commits could be read as a quality claim: the line says counts of declared kinds only, and the body keeps its own audit and result sections.

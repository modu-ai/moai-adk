---
id: SPEC-20261120-verified-trailer-in-use
title: "The Moai-Verified trailer is written in the implementation guidance and summarised in the bundle body"
status: draft
revision: 2
created: 2026-11-20
---

# The Moai-Verified trailer is written in the implementation guidance and summarised in the bundle body

## Intent

`Moai-Verified: <tests|mutants|e2e|smoke|none>[, …]` 트레일러(SPEC-20261119-verified-trailer)는 코드로는 있어요: `moai trailers --verified` 가 줄을 찍고, `moai trailer-check` 가 값과 빠진 줄을 검사하고, `moai trailers --list` 가 종류별로 모아 줘요. 그런데 아무도 쓰지 않아요. 2026-11-20 이 기준 가지(`1c9f1a1f95`)에서 직접 센 결과, 지침 파일 `skills/moai/references/internal/commit-trailers.md`, `agents/impl.md`, `AGENTS.md`, `REVIEW.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/sync.md` 어디에도 이름이 없고(`grep -c Moai-Verified`), `moai trailer-check` 를 부르는 곳은 `commit-trailers.md` 의 안내 문장뿐이에요. 줄이 없는 커밋이 대부분이어서 "검증 안 된 커밋을 조회"할 수 없어요.

원래 SPEC(개정 1)은 세 곳에서 트레일러가 쓰이게 하려고 했어요: (1) 지침이 트레일러와 다섯 값을 말하고, (2) 로컬 검사 모음(`tools/local_check.sh`)이 `trailer-check` 로 빠진 줄을 실패시키고, (3) 묶음 PR 본문이 `trailers --list` 의 한 줄 요약을 실어요. 근거: 사용자 결정 2026-11-20(리더 대화, 검증 2차 묶음 W2-3). 원 글은 "모든 커밋에 어떻게 검증했는지 표시한다, `none` 도 값이다"(Addy Osmani, "How the Claude Code team thinks about verification", 2026-11-16).

개정 2 는 지금 끝난 두 조각만 이 SPEC 에 남겨요: 구현 에이전트 지침의 문장(`agents/impl.md`)과 묶음 본문의 한 줄 요약(`src/bundle.rs`). 나머지 두 조각 — `commit-trailers.md` 의 문단과 `tools/local_check.sh` 의 `trailer-check` 단계 — 는 그 파일을 다른 카드가 예약하고 있어 쓸 수 없어서(개정 1 의 Plan 이 이미 그렇게 예상했어요), 파일이 풀린 뒤 이어서 하는 후속 조각으로 아래 "Follow-up"에 그대로 적어 둬요. 규칙 시작 커밋의 뜻은 기존 SPEC 그대로예요: 검사 범위 안에서 이 줄을 처음 실은 커밋이고, 그 앞 커밋은 `none` 으로 읽혀 면제돼요.

## Non-goals

- `trailers`·`trailer-check` 의 동작과 값 어휘 변경: 이미 SPEC-20261119-verified-trailer 가 정했어요.
- 검증 증명(어떤 시험, 몇 개)의 기록: 종류 이름만이고 근거는 PR 본문과 SPEC Result 에 남아요.
- 과거 커밋 다시쓰기, 커밋 훅으로 강제하기, 원격 저장소 읽기.
- `AGENTS.md`, `REVIEW.md`, `mode-agent.md`, `sync.md` 의 문장 추가: 지침 범위는 `commit-trailers.md` 와 `agents/impl.md` 두 파일이에요.
- 새 크레이트와 모델 호출: 이 검사는 결정적이고 네트워크를 쓰지 않아요.

## Requirements

- REQ-1: The agent definition `agents/impl.md` shall tell the agent to end each commit with the trailers that `moai trailers --spec <SPEC id> --verified <kinds>` prints and to name in `Moai-Verified` only the kinds of check it ran, `none` when no check ran.
- REQ-2: When `moai bundle body` writes a body, it shall add one line that gives the commit count per `Moai-Verified` kind (a commit without the line counts under `none`, as `moai trailers --list` reads it) and the count of reverted commits, over the commits of the merge base of the tip and the branch up to the pinned head.
- REQ-3 (REQ-2): IF the by-kind listing of that range cannot be read or lacks a count, THEN `moai bundle body` shall refuse with a message that names the failure and write no body.
- REQ-4 (REQ-2): The summary line shall hold only counts and the fixed words, never a card number, a session id or a commit message.

## Acceptance

- AC-1 (REQ-1): the agent definition holds one line with the pinned sentence, built from the trailer module's own constants (the key and the five kinds), so a kind added to or dropped from the module fails the pin · verify: `cargo test --locked --offline --test verified_trailer_in_use_cli implementation_agent`
- AC-2 (REQ-2, REQ-4): the body of a bundle whose branch holds commits of the kinds `tests`, `mutants` with `e2e`, and none, with one commit reverted, has exactly one summary line with those counts, a commit without the line counts under `none`, and the line holds no card number or session id · verify: `cargo test --locked --offline --test bundle_body_cli verified_summary`
- AC-3 (REQ-3): a listing that lacks a count, or an error from the listing, gives a refusal that names the failure, in the function that builds the line (the real command cannot make the range unreadable, so the check is a unit test) · verify: `cargo test --locked --offline --lib bundle::tests::verified_summary`
- AC-4 (REQ-1…REQ-4): the lines of the body that exist today, the manifest digest of the skill package and the neighbour tests stay as they were · verify: `cargo test --locked --offline --test bundle_body_cli`

## Scope

- write: `src/bundle.rs`, `tests/bundle_body_cli.rs`, `tests/verified_trailer_in_use_cli.rs`, `agents/impl.md`, `.moai/specs/SPEC-20261120-verified-trailer-in-use/**`
- read-only: `src/trailers.rs` (`list`, `VERIFIED_KEY`, `VERIFIED_KINDS`), `src/ops/trailers.rs`, `src/ops/trailer_check.rs`, `.github/workflows/ci.yml`
- forbidden: `skills/**`, `tools/local_check.sh`, other projects' hook files, the real HOME and ledger, the network

## Plan

Tests first (RED) for every part.

1. Bundle summary (REQ-2…REQ-4): RED in `tests/bundle_body_cli.rs`: a temporary repository with commits `tests`, `mutants, e2e`, none and a standard revert; the body must hold one line with the counts. GREEN in `src/bundle.rs`: call `trailers::list(dir, "<base>..<pinned>")` in `body`, refuse when it fails or lacks a count, print the line after the integration paragraph. Mutants (at most 5): the reverted count, the range, a missing count read as zero, the singular rule, the line dropped from the body.
2. Guidance sentence (REQ-1): RED: `tests/verified_trailer_in_use_cli.rs` builds the sentence from `VERIFIED_KEY` and `VERIFIED_KINDS` and requires a line of `agents/impl.md` that holds it. GREEN: the fifth step of `agents/impl.md`.
3. Neighbours: `check affected` over the range (the library and binary tests, the tests of the area, the listed scans, the Python tests, fmt, clippy, `git diff --check`), `spec check --closure`, `trailer-check`.

Research notes (read 2026-11-20, in `1c9f1a1f95`): `src/trailers.rs` (`list`, `VERIFIED_KINDS`), `src/ops/trailer_check.rs` (flags `--branch`, `--verified-since`; exit 1 on a failing commit), `src/bundle.rs` (`body`, `judge_all` already runs a `trailer_check` check per head), `tests/bundle_body_cli.rs` (the fixture model).

### Follow-up

Two pieces of revision 1 wait for files that other cards hold; each is its own small change when the file is free, by `card scope` (it only widens), never by writing under another card's reservation:

- `skills/moai/references/internal/commit-trailers.md` names the `Moai-Verified` trailer with its five values, the command that prints the line and the command `moai trailers --list`, with a pinned sentence in `tests/verified_trailer_in_use_cli.rs` and the manifest digest moved (`skills/moai/references/source-manifest.json`). Held by another card at this revision.
- `tools/local_check.sh` runs `trailer-check` over the commits of the head beyond the base as one light step listed after `diff-check`, in both modes and in every leg that runs the rest of the set, with the exit code of the command, and `tests/local_check_test.py` pins the step list. Held by another card at this revision.

## Risks

- A body line that counts commits could be read as a quality claim: the line says counts of declared kinds only, and the body keeps its own audit and result sections.
- The guidance sentence is only read by an agent that opens `agents/impl.md`: the pin proves the sentence exists, not that every agent follows it; the bundle line shows the result in each body.
- A rule start in a long branch exempts the commits before it: this is the rule of the first SPEC, not changed here.

## Result

All four criteria are met for the two pieces this revision keeps. Commits on `spec/verified-trailer-in-use-20261120` (base `1c9f1a1f95`): SPEC `819fc7b271`, RED of the bundle line `f4266de2aa` (`moai tdd red` recorded the failure as `test-failure`), GREEN `533e2fb7c3`, RED of the guidance pin `3bc2201714` (`moai tdd red` recorded it), GREEN `3c164381ae`, this revision `2c8d4e5633`.

- AC-1: met — `verified_trailer_in_use_cli::the_implementation_agent_names_the_trailer_and_its_five_values` builds the sentence from `VERIFIED_KEY` and `VERIFIED_KINDS` and needs a line of `agents/impl.md` that holds it. RED first: at `3bc2201714` the test failed because the file had no such line; with the fifth step added (`3c164381ae`) it passes (1 passed, 0.00 s), and `moai tdd green` recorded the check as passed. The filter `implementation_agent` of the verify command matches this test's name.
- AC-2: met — `bundle_body_cli::verified_summary_counts_the_kinds_and_the_reverted_commit` expects exactly the line `Verification kinds (`Moai-Verified`, 6 commits): tests 1, mutants 1, e2e 1, smoke 0, none 4 (a commit without the line counts as none); reverted 1.` between the integration paragraph and `## Self-Audit`, and `verified_summary_reads_a_missing_line_as_none` pins the none reading; the line holds counts and fixed words only. Mutants (python3 of a script that edits `src/bundle.rs`, runs `--lib bundle::` and then `--test bundle_body_cli`, and restores the file; 5 of 5 killed, `restored: True`): the reverted count flipped (killed by the integration test), the range emptied to base..base (integration), a missing count read as zero (unit), the singular rule off (unit), the summary dropped from the body (integration).
- AC-3: met — `bundle::tests::verified_summary_refuses_an_unreadable_or_incomplete_listing` (an error from the listing and a listing without a count each give a refusal that names the failure) and `verified_summary_names_the_counts_in_the_pinned_order`; `--lib bundle::` 20 passed. The count-as-zero mutant above is killed by this unit test.
- AC-4: met — `--test bundle_body_cli` 13 passed (the 11 older tests unchanged, 2 new), and the tests of the area ran in `check affected` below, exit 0 each.
- Neighbours and the wide run, on head `3c164381ae` (the later commit changes this SPEC only): `check-affected base=1c9f1a1f95c5 head=3c164381ae6a targets=45 py=8 scans=listed:39 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS` (run in this worktree with the `moai-cli` binary this lane built in another worktree, which holds the same `check affected` code). On the branch: `git diff --check 1c9f1a1f95..HEAD` 0, `trailer-check 1c9f1a1f95..HEAD` ok (6 commits, 0 failing), `spec check` 0 errors 0 warnings.
- Not met in this revision, by design: the `commit-trailers.md` paragraph and the `local_check.sh` step of revision 1 (REQ-1, REQ-3 and REQ-4 of that revision, AC-1 and AC-3 of it) wait for files that other cards hold; the Follow-up section names both. The card stays open for them.
- Not run: a Linux or Windows build (macOS only here), the CI of a real pull request (no pull request by the M-Flow rule), a bundle built from real merged heads (the body is built in temporary repositories), and an independent read-only audit of this branch (the leader starts it).
- Slips: the first mutant run did not apply the singular-rule mutant (its pattern matched two places) and was repaired and rerun; a test count was off by one in the first draft of the integration test and was fixed before its commit.

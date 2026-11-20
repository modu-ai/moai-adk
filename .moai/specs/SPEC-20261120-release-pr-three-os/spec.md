---
id: SPEC-20261120-release-pr-three-os
title: "Release pull requests run the full three-OS test set before merge"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

The release pull request (head branch name starting with `release/v`, e.g. the next
`release/v4.2647.10`) is the release unit from the local integration branch to the remote
`v2026.11.0`, so its CI runs the full check set on linux-x86_64, macos-arm64 and
windows-x86_64 before merge (user decision 2026-11-20 on the card). Today a pull_request run
keeps linux-x86_64 only unless the platform-sensitive classifier answers otherwise (C-76),
the windows-build job is skipped on pull_request (`if: github.event_name != 'pull_request'`),
and Windows never runs the test set anywhere — it compiles all targets after a merge
(windows-build) and the supported-platforms decision notes its test legs as later work.

Survey (read-only): the `targets` job's decide step already builds the matrix from the two
legged literals `full` and `linux` and narrows on pull_request through the classifier
(SPEC-20261119-ci-check-set-legs made every target a nextest leg and a rest leg). Three known
Windows frictions are confirmed before implementation: the runner ships `python` and no
`python3` (actions/setup-python issue #1079, maintainer-confirmed; the Windows 2025 image
readme lists Python 3.12 with no python3.exe) while the unittest step pins `python3`;
ripgrep is absent from the Windows 2025 image (its toolset json lists no ripgrep package and
the add-ripgrep request closed as not_planned) while the install step's fallback branch exits
for an unconfigured platform; and the pinned cargo-nextest release ships an
x86_64-pc-windows-msvc asset whose sha256 the download step must verify like the others.

## Non-goals

- Changing the ordinary pull_request matrix, the push and workflow_dispatch matrix, or the
  classifier and its fail-safe: an ordinary PR keeps today's rule exactly (REQ-3).
- The required check names (`ci`, `moai/audit`), the integration ruleset, the release
  workflow, and the name-switch shim: untouched.
- Fixing every Windows-only test failure up front: the first release-PR run surfaces them and
  this card repairs what it shows (REQ-8); compile-level Windows health stays with the
  windows-build job, which already runs it after every merge.
- The candidate verification workflows and the smoke harness: untouched.

## Requirements

- REQ-1: While a pull request's head branch name starts with release slash v, the decide step shall emit a check matrix of six entries — linux-x86_64 and macos-arm64 with the nextest and rest legs as today, plus windows-x86_64 on windows-2025 runners with the nextest and rest legs. · failure: none — pinned by AC-1.
- REQ-2: While a release pull request runs, the windows-build job shall run as it does after a merge, keeping its build and self-replace proof in the release path too. · failure: none — pinned by AC-4.
- REQ-3: While a pull request's head branch name does not start with release slash v, the matrix decision shall keep today's rule unchanged — the classifier narrows to linux-x86_64 when it answers not needed, the full matrix otherwise, and any listing, classifier or parse failure keeps the full matrix. · failure: none — pinned by AC-3.
- REQ-4: While a release pull request's check matrix carries the windows-x86_64 legs, the aggregate ci check shall treat them like every other matrix entry, so a Windows test failure blocks the release merge. · failure: none — pinned by AC-2.
- REQ-5: While the cargo-nextest download step runs on the windows-x86_64 nextest leg, it shall verify the x86_64-pc-windows-msvc asset against its recorded sha256 the way the linux and macos assets are verified. · failure: none — pinned by AC-5.
- REQ-6: While the ripgrep install step runs on a Windows runner, it shall install ripgrep there instead of failing as an unconfigured platform. · failure: none — pinned by AC-5.
- REQ-7: While the check set runs on a Windows runner, the workflow shall provide the python3 command name to the steps it starts, so the check script's unittest step runs the same command it runs on every runner instead of failing on the interpreter name. · failure: none — pinned by AC-6.
- REQ-8: While a release pull request run exposes a test failing only on Windows, the lane shall gate that test behind its platform cfg with the reason written beside the gate. · failure: none — pinned by AC-7.

## Acceptance

- AC-1 (REQ-1): the decide script derives the release matrix by joining the two windows entries onto today's `full` list (so push and the ordinary PR lists stay byte-identical), dispatches on the head branch name that reaches it through env only (no inline `github.head_ref` expression in the script — it is an injection surface), and the workflow test pins the six-entry result, the windows-2025 runner and the branch guard.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-2 (REQ-4): the aggregate keeps `ci` as the matrix result's judge — the check job consumes `needs.targets.outputs.matrix`, so the windows legs are required by the same gate as every other entry; the test re-pins that flow and the windows entries' presence in the release list.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-3 (REQ-3): the ordinary-PR regression — the matrix helper's on-pr set stays exactly the two linux legs, the classifier branch and its fail-safe survive byte-for-byte, and push and dispatch keep the full list.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-4 (REQ-2): the windows-build condition reads push and dispatch as today and adds a release head branch name on pull_request; the test pins the exact expression.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-5 (REQ-5, REQ-6): the nextest case gains the windows-x86_64 asset with its recorded sha256 (sum kept next to the linux and macos sums), and the ripgrep step gains the Windows install branch with the image fact recorded beside it.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-6 (REQ-7): the workflow gains one small machine-setup step that links the runner's `python` under the name `python3` on a directory prepended to `GITHUB_PATH` — only when the runner is Windows — with the interpreter fact recorded beside it; the test pins the step's condition, command and effect, and tools/local_check.sh stays untouched (its unittest step, and C-198's parallel runner replacement of it, then keep working unchanged).
  verify: `python3 -m unittest tests.github_assets_test`
- AC-7 (REQ-8): observation — the first release pull request run (the next release branch, `release/v4.2647.10`) shows the six-leg matrix green with any Windows-only exclusion cfg-gated and reasoned; the run id and the exclusion list are recorded on the card.
  verify: 카드에 기록한 run id 로 `gh run view <id> --json jobs` 재읽기
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8): neighbours — the whole workflow test module, the local-check step, leg and slot tests, the superseded verdict and the scope tests keep today's behaviour.
  verify: `python3 -m unittest tests.github_assets_test tests.local_check_test tests.ci_superseded_test tests.ci_scope_test`; `moai spec check SPEC-20261120-release-pr-three-os`

## Scope

- write: `.github/workflows/ci.yml`, `tests/github_assets_test.py`, this SPEC folder
- read-only: `tools/ci_scope.py`, `tools/ci_superseded.py`, `tools/local_check.sh` (C-198's reservation — read only), `.github/workflows/candidate-target.yml` (the Windows runner precedent)
- reserved by other cards: ci.yml and tests/github_assets_test.py are C-162's reservation until its integration lands — this card widens onto them only after that tip reaches the local integration branch (this decision's diff is written on top of that copy); C-198 holds tools/local_check.sh and rewrites the unittest step to its parallel runner, which this card's interpreter shim is designed to keep working unchanged
- forbidden: changing the required check names, the ruleset, the ordinary PR's matrix, tools/local_check.sh (C-198's reservation), the real HOME, another card's branch

## Plan

1. Commit this SPEC (its folder only — ci.yml and the two test files are C-162's reservation while it waits for integration).
2. Watch the local integration branch: once C-162's tip is merged in, merge the moved tip into this card's branch (M-Flow: merge, never rebase), widen the reservation to the four write paths, and implement on that copy.
3. ci.yml, smallest lines: the `windows` two-entry literal and the `release` join under the decide literals; the head-branch `case` that sets `matrix="$release"` before the classifier block (guarding the classifier behind `matrix != release` so its lines do not move); `HEAD_REF` in the decide step's env block; the windows-x86_64 case in the pinned nextest download; the Windows branch of the ripgrep install (choco, with the image fact as a comment); the windows-build `if` extended with the release branch-name test; one Windows-only python3 shim step beside the ripgrep install.
4. Tests: github_assets_test.py — the six-entry release list (derived from the same join the script does), the env-only head-branch guard, the windows-build condition, the nextest windows sum, the ripgrep Windows branch, the python3 shim step, and the ordinary-PR regression.
5. Neighbours and the full pre-push check set, then the audit request note (tier 2 — leader audit; LANE-READY-INTEGRATION only after its verdict).

## Risks

- The first Windows test run can surface more than the three known frictions: platform-assumption tests (path shapes, permission bits, lock flavours) may fail only there. The card's contract is explicit — gate them behind their cfg with reasons, fix or record, never silently skip; AC-7 records the outcome. New fix commits beyond those gates stay on this branch with an audit delta note.
- The windows leg timeouts start at 45 minutes (the macOS cap; it must exceed windows-build's 30-minute all-targets compile because the nextest leg compiles and runs). The first release run measures them; the leader adjusts from observation (D4).
- ci.yml is a shared file (C-162 today, C-198 for one test line): this card merges onto the tip that contains their changes and keeps its ci.yml diff to the decide literals, the case line, one env line, one nextest case, one ripgrep branch, one `if` and one shim step — no reformatting.
- The windows asset sha256 is read once from the release's own .sha256 file (2026-11-20) and pinned like the others; a checksum mismatch fails the leg loudly instead of falling back.
- The python3 shim is a CI-machine step, not a change of the one check set (#2906 S5a): a local Windows checkout would still need python3 under its own name. Recording that fact beside the step keeps the boundary explicit; moving the fallback into the check script is the C-198 follow-up (D3 undo).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The release branch is judged by its head branch name, passed through the decide step's env (never an inline expression) | `github.head_ref` is PR-author-controlled; the script's existing env block is the injection-safe path (C-162 pattern) | Inline with parameter cleansing |
| D2 | The release list joins the two windows entries onto the existing `full` literal instead of a third full copy | Ordinary-PR and push lists stay byte-identical; the join is one line | A third maintained literal |
| D3 | The Windows interpreter gap is closed in the workflow: one step links the runner's `python` as `python3` on a `GITHUB_PATH` dir, only on Windows | Windows runners ship `python` and no `python3` (setup-python #1079, maintainer-confirmed); tools/local_check.sh is C-198's reservation this round, and the shim keeps its unittest step — including C-198's parallel-runner rewrite of it — working unchanged | An interpreter fallback inside the check script (after C-198's integration, its own card) |
| D4 | Windows leg timeout 45 minutes, to be re-measured on the first release run | Above windows-build's 30-minute compile-only cap; macOS parity | Change the value after observation |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.

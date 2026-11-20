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
(SPEC-20261119-ci-check-set-legs made every target a nextest leg and a rest leg). Windows
frictions are confirmed before implementation, two from the tier-2 audit: the runner's Git
Bash answers `python3` on the real runner (run 44634311021 — the earlier survey's #1079
reading was wrong for this image, so there is no interpreter shim), ripgrep is absent from
the Windows 2025 image (its toolset json lists no ripgrep package and the add-ripgrep
request closed as not_planned) while the install step's fallback branch exits for an
unconfigured platform, and the pinned cargo-nextest release's Windows archive carries the
`.exe` member under a Windows-shaped temp path — the download step needs the
candidate-target precedent (member name and cygpath) to verify and unpack it.

## Non-goals

- Changing the ordinary pull_request matrix, the push and workflow_dispatch matrix, or the
  classifier and its fail-safe: an ordinary PR keeps today's rule exactly (REQ-3).
- The required check names (`ci`, `moai/audit`), the integration ruleset, the release
  workflow, and the name-switch shim: untouched.
- Fixing every Windows-only test failure up front: the first release-PR run surfaces them and
  this card repairs what it shows (REQ-7); compile-level Windows health stays with the
  windows-build job, which already runs it after every merge.
- The candidate verification workflows and the smoke harness: untouched.

## Requirements

- REQ-1: While a pull request's head branch name starts with release slash v, the decide step shall emit a check matrix of six entries — linux-x86_64 and macos-arm64 with the nextest and rest legs as today, plus windows-x86_64 on windows-2025 runners with the nextest and rest legs. · failure: none — pinned by AC-1.
- REQ-2: While a release pull request runs, the windows-build job shall run as it does after a merge, keeping its build and self-replace proof in the release path too. · failure: none — pinned by AC-4.
- REQ-3: While a pull request's head branch name does not start with release slash v, the matrix decision shall keep today's rule unchanged — the classifier narrows to linux-x86_64 when it answers not needed, the full matrix otherwise, and any listing, classifier or parse failure keeps the full matrix. · failure: none — pinned by AC-3.
- REQ-4: While a release pull request's check matrix carries the windows-x86_64 legs, the aggregate ci check shall treat them like every other matrix entry, so a Windows test failure blocks the release merge. · failure: none — pinned by AC-2.
- REQ-5: While the cargo-nextest download step runs on the windows-x86_64 nextest leg, it shall verify the x86_64-pc-windows-msvc asset against its recorded sha256 the way the linux and macos assets are verified, and unpack the tar member the Windows archive actually carries (the .exe name) after the Windows temp path is converted for the shell, following the candidate-target precedent. · failure: none — pinned by AC-5.
- REQ-6: While the ripgrep install step runs on a Windows runner, it shall install ripgrep there from the pinned official release — version and sha256 fixed and checked before unpack, placed on PATH — instead of failing as an unconfigured platform or drawing from a package manager. · failure: none — pinned by AC-5.
- REQ-7: While a release pull request run exposes a test failing only on Windows, the lane shall gate that test behind its platform cfg with the reason written beside the gate. · failure: none — pinned by AC-6.

## Acceptance

- AC-1 (REQ-1): the decide script derives the release matrix by joining the two windows entries onto today's `full` list (so push and the ordinary PR lists stay byte-identical), dispatches on the head branch name that reaches it through env only (no inline `github.head_ref` expression in the script — it is an injection surface), and the workflow test pins the six-entry result, the windows-2025 runner and the branch guard.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-2 (REQ-4): the aggregate keeps `ci` as the matrix result's judge — the check job consumes `needs.targets.outputs.matrix`, so the windows legs are required by the same gate as every other entry; the test re-pins that flow and the windows entries' presence in the release list.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-3 (REQ-3): the ordinary-PR regression — the matrix helper's on-pr set stays exactly the two linux legs, the classifier branch and its fail-safe survive byte-for-byte, and push and dispatch keep the full list.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-4 (REQ-2): the windows-build condition reads push and dispatch as today and adds a release head branch name on pull_request; the test pins the exact expression.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-5 (REQ-5, REQ-6): the nextest step carries the windows asset with its recorded sha256 and the candidate-target precedent's Windows handling (the .exe member, the temp path converted for the shell), and the ripgrep step installs the pinned official release (14.1.1, sha256 checked before unpack, on PATH) — the tests pin those shapes and that no package manager is used.
  verify: `python3 -m unittest tests.github_assets_test`
- AC-6 (REQ-7): observation — the first release pull request run (the next release branch, `release/v4.2647.10`) shows the six-leg matrix green with any Windows-only exclusion cfg-gated and reasoned; the run id and the exclusion list are recorded on the card.
  verify: 카드에 기록한 run id 로 `gh run view <id> --json jobs` 재읽기
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): neighbours — the whole workflow test module, the local-check step, leg and slot tests, the superseded verdict and the scope tests keep today's behaviour; the guard's near misses (a wider family, a different case, a different separator) are pinned by running the decide script's own `case` pattern.
  verify: `python3 -m unittest tests.github_assets_test tests.local_check_test tests.ci_superseded_test tests.ci_scope_test`; `moai spec check SPEC-20261120-release-pr-three-os`

## Scope

- write: `.github/workflows/ci.yml`, `tests/github_assets_test.py`, this SPEC folder
- read-only: `tools/ci_scope.py`, `tools/ci_superseded.py`, `tools/local_check.sh` (C-198's reservation — read only), `.github/workflows/candidate-target.yml` (the Windows runner precedent)
- reserved by other cards: ci.yml and tests/github_assets_test.py are C-162's reservation until its integration lands — this card widens onto them only after that tip reaches the local integration branch (this decision's diff is written on top of that copy); C-198 holds tools/local_check.sh and rewrites the unittest step to its parallel runner, which this card's interpreter shim is designed to keep working unchanged
- forbidden: changing the required check names, the ruleset, the ordinary PR's matrix, tools/local_check.sh (C-198's reservation), the real HOME, another card's branch

## Plan

1. Commit this SPEC (its folder only — ci.yml and the two test files are C-162's reservation while it waits for integration).
2. Watch the local integration branch: once C-162's tip is merged in, merge the moved tip into this card's branch (M-Flow: merge, never rebase), widen the reservation to the four write paths, and implement on that copy.
3. ci.yml, smallest lines: the `windows` two-entry literal and the `release` join under the decide literals; the head-branch `case` that sets `matrix="$release"` before the classifier block (guarding the classifier behind `matrix != release` so its lines do not move); `HEAD_REF` in the decide step's env block; the windows-x86_64 case in the pinned nextest download with the candidate-target precedent (the .exe member, the temp path converted for the shell); the Windows branch of the ripgrep install from the pinned official release (version and sha256 fixed, no package manager); the windows-build `if` extended with the release branch-name test.
4. Tests: github_assets_test.py — the six-entry release list (derived from the same join the script does), the env-only head-branch guard, the windows-build condition, the nextest windows asset with the Windows handling, the pinned ripgrep branch, the guard's near misses run through its own `case` pattern, the release.yml naming rule pinned against the ci guard, and the ordinary-PR regression.
5. Neighbours and the full pre-push check set, then the audit request note (tier 2 — leader audit; LANE-READY-INTEGRATION only after its verdict).

## Risks

- The first Windows test run can surface more than the confirmed frictions: platform-assumption tests (path shapes, permission bits, lock flavours) may fail only there. The card's contract is explicit — gate them behind their cfg with reasons, fix or record, never silently skip; AC-6 records the outcome. New fix commits beyond those gates stay on this branch with an audit delta note. The audit already split two pre-existing Windows defects (the smoke harness's pwd·pty top-level imports, and 27 clippy findings against the Windows target) into their own card.
- The windows leg timeouts start at 45 minutes (the macOS cap; it must exceed windows-build's 30-minute all-targets compile because the nextest leg compiles and runs). The first release run measures them; the leader adjusts from observation (D4).
- ci.yml is a shared file: this card merges onto the integration tip that contains the other cards' changes and keeps its ci.yml diff to the decide literals, the case line, one env line, the nextest Windows handling, the pinned ripgrep branch and one `if` — no reformatting.
- The windows asset sha256 is read once from the release's own .sha256 file (2026-11-20) and pinned like the others; a checksum mismatch fails the leg loudly instead of falling back. The ripgrep sha256 is the candidate-target precedent's reviewed literal.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The release branch is judged by its head branch name, passed through the decide step's env (never an inline expression) | `github.head_ref` is PR-author-controlled; the script's existing env block is the injection-safe path (the base-scope card's pattern) | Inline with parameter cleansing |
| D2 | The release list joins the two windows entries onto the existing `full` literal instead of a third full copy | Ordinary-PR and push lists stay byte-identical; the join is one line | A third maintained literal |
| D3 | No interpreter shim: the runner's Git Bash answers `python3` on the real Windows runner | The tier-2 audit's measurement (run 44634311021) disproved the earlier #1079 reading for this image; the shim step is removed | Re-add a shim if an image change removes python3 again, measured on a real run |
| D4 | Windows leg timeout 45 minutes, to be re-measured on the first release run | Above windows-build's 30-minute compile-only cap; macOS parity | Change the value after observation |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.

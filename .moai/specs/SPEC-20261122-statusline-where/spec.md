---
id: SPEC-20261122-statusline-where
title: "The first line of the status line says who the session is and where it stands, and git runs only when something changed"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-statusline-flow
---

# 상태줄 1번째 줄: 누구이고 어디에 있나, git 은 바뀔 때만

## Intent

부모 SPEC `SPEC-20261122-statusline-flow` 의 줄 역할 중 **1번째 줄**이에요. 사용자가 고른 모양(리더 결정 기록 5935): 모델, 저장소, 작업트리, 브랜치, 통합 브랜치와의 거리, PR. 지금 1번째 줄은 태그·모델·노력·Claude Code 버전·MoAI 버전이고, 저장소·브랜치·PR·작업트리 이름표는 2번째 줄에 있어요. 줄 바꿈의 이유와 조사 표는 부모 SPEC 에 있어요(조사 보고서 `statusline-redesign-2026-11-22.md`).

이 SPEC 이 더 정하는 것은 **비용**이에요. 지금은 그릴 때마다 `git status` 를 돌려서 이 저장소에서 중앙값 70 ms(p90 105 ms), 분리된 HEAD 에서는 git 이 세 번 돌아 중앙값 89 ms(p90 123 ms)예요. 호스트는 갱신을 300 ms 모아서 한 번 부르고 실행 중인 스크립트를 취소하니(공식 문서, 2026-11-21T23:52Z 받음) 한 번의 비용이 곧 화면이 느려지는 정도예요. 사용자 선택에 따라 그릴 때는 읽기만 한다는 옛 규칙을 바꿔서 `$TMPDIR` 에 세션별 캐시를 쓰고, git 은 바뀐 것이 보일 때만 돌려요. 공식 문서도 같은 방식을 권해요("세션 id 로 임시 파일에 캐시").

**고르지 않아서 그대로 두는 것:** 이슈·PR 수 `⬇️ n 🔀 n` 과 그 `gh` 갱신(10분마다, 네트워크). 사용자가 고르지 않았어요(결정 기록 5935). 대신 제품 문서의 "네트워크를 쓰지 않는다"는 문장이 틀리다는 점은 부모 SPEC REQ-12 가 고쳐요.

## Non-goals

- 이슈·PR 수와 `gh` 갱신을 없애거나 바꾸는 것(고르지 않았어요).
- 줄 2·3 의 내용. 그리고 작업트리 이름표 `WT-c<카드>-<이름>`: 카드는 부모 SPEC 의 🎯 로 옮겨 가요.
- 캐시를 다른 세션이나 다른 컴퓨터와 나누는 것.

## Requirements

- REQ-1: The status line shall write its first line as the items model with effort, repository with the issue and pull request counts, worktree or branch with its ahead and dirty marks, distance to the integration branch and pull request in this order · failure: none — an item without its input is left out with its separator and the first line is left out when all are
- REQ-2: The status line shall read the branch name or the short commit of a detached HEAD from the `HEAD` file of the git directory without starting a process · failure: none — an unreadable `HEAD` leaves the branch items out
- REQ-3: When the session is in a linked worktree and its branch name equals the worktree name, the status line shall show the worktree item alone and else the worktree item and a `🅱️  <branch>` item · failure: none — a session in the main checkout shows the branch item alone
- REQ-4: When the branch is not the integration branch, the status line shall show `🧭 <name> ↑<n> ↓<n>` as the commits it is ahead of and behind the integration branch · failure: none — on the integration branch itself the branch item shows `↑<n>` as the landings not yet pushed and an unknown integration branch leaves the item out
- REQ-5: The status line shall keep the git facts it shows in a per-session cache file under `$TMPDIR` named by the stdin session id and owned by the user alone · failure: none — an unwritable cache folder means every render asks git and the line is still the same
- REQ-6: While the cache of a session holds the same `HEAD` content, the same integration branch commit and an age below ten seconds, the status line shall start no process for git · failure: none — any difference or an older cache asks git once and rewrites the cache
- REQ-7: The status line shall start at most one git process in one render and none for the short commit of a detached HEAD · failure: none — a second fact is read from the same output
- REQ-8: The status line shall show no tag on its first line · failure: none — the lane registry and the stdin session name are not read and a session name changes nothing on the line
- REQ-9: The status line shall write the pull request item as `🔗 #<n> <state>` and as `🔗 !<n> <state>` when the host says the request kind is a merge request · failure: none — a missing pull request leaves the item out
- REQ-10: The status line shall read the model name only from the stdin model and the effort only from the stdin effort and thinking fields · failure: none — a missing model shows the shortened model id and then nothing
- REQ-11: The status line shall write the first line in the widest of the forms of the Layouts section whose width fits the budget of the parent SPEC · failure: none — the narrowest form is written when none fits
- REQ-12: The statusline documentation shall say that the status line starts git and a detached `gh` refresh and writes a cache under `$TMPDIR` · failure: none — the stale sentence "reads only" is found by the doc test

## Acceptance

- AC-1 (REQ-1, REQ-9, REQ-10): the first line of every case — worktree session with a pull request, main checkout, generic project, merge request, missing model — is exactly the string of its golden, plain and with color.
  verify: `cargo test --offline --locked --test statusline_cards line_one_items_`
- AC-2 (REQ-2, REQ-7): with a fake `git` first on `PATH` that counts its starts, a branch, a detached HEAD and a missing `HEAD` start at most one process and none for the short commit.
  verify: `cargo test --offline --locked --test statusline_git_env head_file_and_process_count`
- AC-3 (REQ-3, REQ-4): the worktree-name rule and the integration distance are pinned for equal names, unequal names, the main checkout, a branch that is the integration branch, an unknown integration branch and a detached HEAD.
  verify: `cargo test --offline --locked --lib statusline::state::where_tests`
- AC-4 (REQ-5, REQ-6): a second render within ten seconds with the same `HEAD` starts no `git`, a changed `HEAD` or a changed integration commit or an eleven-second-old cache starts one, an unwritable folder still renders, the file is mode 0600 in a 0700 folder and named only by the session id.
  verify: `cargo test --offline --locked --test statusline_git_env session_cache_`
- AC-5 (REQ-8): a session name of `leader`, `agent-3` or anything else changes no byte of line 1 and no lane registry file is opened.
  verify: `cargo test --offline --locked --test statusline_tag_order the_session_name_changes_nothing`
- AC-6 (REQ-11): each form of line 1 is exactly the string written in the Layouts section and changes at its boundary.
  verify: `cargo test --offline --locked --lib statusline::render::line_one_form_tests`
- AC-7 (REQ-12): the document test finds the sentences about git, `gh` and the cache and not the stale ones.
  verify: `cargo test --offline --locked --test statusline_numbers_only the_statusline_doc_describes_the_cache`

### Layouts

Widths as `unicode-width` counts them (probe 2026-11-22). `next` is a fixture integration branch name; a name over twelve characters is cut with `…`.

| form | text | cols |
|---|---|---|
| W0 widest | `🧠 Opus 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 \| 🌳 statusline-flow ↑3 +2 \| 🧭 next ↑3 ↓0 \| 🔗 #12 pending` | 112 |
| W1 no counts | `🧠 Opus 5.5 high·t \| 📡 modu-ai/moai-adk \| 🌳 statusline-flow ↑3 +2 \| 🧭 next ↑3 ↓0 \| 🔗 #12 pending` | 100 |
| W2 short | `🧠 Opus high \| 📡 moai-adk \| 🌳 statusline-flow ↑3 +2 \| 🧭 next ↑3 ↓0 \| 🔗 #12` | 78 |
| W3 | `🧠 Opus \| 🌳 statusline-flow ↑3 +2 \| 🧭 next ↓0 \| 🔗 #12` | 56 |
| W4 narrowest | `🧠 Opus \| 🌳 statusline-flow ↑3 +2` | 34 |
| main checkout | `🧠 Opus 5.5 high·t \| 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 \| 🅱️  next ↑2 +1 \| 🔗 #12 pending` | 86 |
| generic project | `🧠 Sonnet 5.5 \| 📡 acme/webapp \| 🅱️  feature/login ↑2 ↓1 +4 \| 🔗 #88 approved` | 77 |

In the main checkout on the integration branch `↑2` is the number of landings not yet pushed. The forms drop in the order counts, repository owner and effort, the distance's ahead count, the repository, the pull request state.

## Scope

- write: `.moai/specs/SPEC-20261122-statusline-where/**` (this slice); the code slice writes `src/statusline/{state,render,counts}.rs`, `tests/statusline_*.rs`, the goldens and the statusline section of the product document.
- forbidden: any `git` command that writes, a cache outside `$TMPDIR`, the lane registry.

## Plan

One landing after the flow line of the parent SPEC: (1) red tests for the file read of `HEAD` and the process counts; (2) the cache with its key and age rule behind a small struct so the key and the folder are one place; (3) the new items and forms; (4) goldens and the document sentences. The detached `gh` refresh is not touched.

## Risks

1. **The cache key misses a dirty edit.** A change to a tracked file that does not touch `HEAD` or the integration commit is not seen until the cache is ten seconds old; the host's `refreshInterval` of 30 seconds and every assistant message re-run the command, so the dirty mark is at most ten seconds late. That is the price of "git only when something changed"; the age is one constant.
2. **`$TMPDIR` may be empty or shared.** On macOS it is per user; where it is unset the cache goes under `/tmp` in a folder made `0700` and the render refuses a folder that another user owns. A refused or unwritable folder falls back to asking git, never to failing.
3. **The integration branch name.** It is the one the project profile names and else the remote's default branch; a repository with neither shows no `🧭`. How the profile names it is read in the code slice and noted in its Result.
4. **Sessions multiply files.** One small file per session id in `$TMPDIR`; files older than a day are removed by the render that writes a new one, at most ten per render.

## Decisions

| Id | Decision | Basis | Undo |
|---|---|---|---|
| D-1 | cache age ten seconds, keyed by the `HEAD` file content and the integration branch commit | the host re-runs the line on each assistant message and every 30 seconds idle; `git status` costs 70–145 ms | one constant |
| D-2 | `HEAD` is read from the file, not from `git symbolic-ref`; the short commit is the first seven characters | a detached HEAD cost three git processes today (89 ms median) | go back to asking git |
| D-3 | the issue and pull request counts and their `gh` refresh stay on the repository item | the user did not pick their removal (note 5935) | a later pick |
| D-4 | the `🌳` item shows the host's worktree name only; the card number moves to `🎯` of the parent SPEC | the parent SPEC's D-2 of revision 1 (the label repeated the card) | restore the label |
| D-5 | no tag; no lane registry; no session-name shapes | the redesign removes the resident leader and agent names (note 5934) | none planned |

## Result

SPEC 만이에요. 돌린 것: 줄 1 모양 너비 측정(`unicode-width` 프로브, 2026-11-22). 돌리지 않은 것: 실제 git 저장소에서의 캐시 적중 측정, 시험들.

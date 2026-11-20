---
id: SPEC-20261120-shipped-workflows-commands
title: "Shipped workflows name moai commands only — the repo-only tools scripts leave the shipped text"
status: draft
revision: 2
created: 2026-11-21
---

## Intent

Every user installs moai and gets `skills/moai` and `agents`; a shipped workflow that says
"in the moai-adk repository run `tools/X`" gives other users nothing (user instruction
2026-11-20T20:1xZ: "로컬 개발 도구와 오픈소스 기능을 혼동하지 마라"). This card slices the fix
into three moves per row of the survey below: (a) a step that any user's project needs
becomes a moai command (Rust with tests) and the shipped text names only the command;
(b) a moai-adk-only tool leaves the shipped skills and agents and moves to this
repository's own places (`AGENTS.local.md`, `REVIEW.md`, `docs/`); (c) a step that is
already a moai command keeps the shipped text naming only the command. 애매한 행은 (b)로 두고
리더에게 노트한다(leader decision, card note 4548).

Survey (re-run at the integration tip `bcfb1bc5b7`, `git grep -nE 'tools/[^ ]+\.(sh|py)' --
skills agents` = 21 lines in 6 shipped files; one row per line, with its move, its target
and its slice):

| Shipped file | `tools/` reference | move | target | slice |
| --- | --- | --- | --- | --- |
| `verify.md` | `tools/smoke/run.sh`:20 | (b) | `docs/moai-adk-tools.md` | later slice (verify.md returns) |
| `verify.md` | `tools/source_manifest.py`:35·80 | (b) | `docs/moai-adk-tools.md` | later slice (verify.md returns) |
| `verify.md` | `tools/ci_scope.py`:57·78 | (b) | `docs/moai-adk-tools.md` | later slice (verify.md returns) |
| `verify.md` | `tools/self_check.sh`:17·49 | (b) | `docs/moai-adk-tools.md` | later slice (verify.md returns) |
| `source-manifest.json` | `tools/legacy_verdicts.py`:33 | kept as recorded history | — (Non-goals) | none |
| `source-manifest.json` | `tools/build_folder.py`:125 | kept as recorded history | — (Non-goals) | none |
| `source-manifest.json` | `tools/self_check.sh`:269 | kept as recorded history | — (Non-goals) | none |
| `source-manifest.json` | `tools/gen_goal_block.py`:425·434 | kept as recorded history | — (Non-goals) | none |
| `source-manifest.json` | `tools/repo_scans.py`:434 | kept as recorded history | — (Non-goals) | none |
| `build-folder.md` | `tools/build_folder.py` env:10·15 | (c) | `moai lane-env` | slice 1 (done) |
| `build-folder.md` | `tools/build_folder.py` clean:13 | (b) | the generic rule (the repo's own tool is named only in the repository's own places) | slice 1 (done) |
| `build-folder.md` | `tools/build_folder.py` remove:14 | (c) | `moai optimize-plan` + the fresh-preview `trim` rule (no apply command exists yet — none is promised) | slice 1 (done) |
| `mode-agent.md` | `tools/self_check.sh`:63 | (b) | `AGENTS.local.md` | later slice (mode-agent.md returns) |
| `mode-agent.md` | `tools/local_check.sh`:64 | (b) | `AGENTS.local.md` | later slice (mode-agent.md returns) |
| `self-check.md` | `tools/self_check.sh`:5 | (b) | the generic rules + the repository's own agent instructions name the repo tool | slice 1 (done) |
| `self-check.md` | `tools/source_manifest.py`:23 | (b) | the generic rules + the repository's own agent instructions | slice 1 (done) |
| `rule-history.md` | `tools/local_check.sh`:118 | (b) | `AGENTS.local.md` | later slice (rule-history.md returns) |

Slice 1 re-verified this survey, wrote the per-row plan into this SPEC, and cleaned the
shipped text of the files this card holds; the rows whose files are another card's
reservation are done in later slices right after those cards return (the reservations are
listed under Scope). Slice 2 starts the first command of move (a) families (REQ-9).

## Non-goals

- Writing the new Rust commands of move (a): each command is its own later slice behind
  this SPEC's per-row plan — slice 1 only rewrites the shipped text of the rows that need
  no new command and records the plan for the rest. The first such command starts in
  slice 2 (REQ-9) and its own tests; the other rows of move (a) stay later slices.
- Changing `tools/*.sh|*.py` themselves: the repo-only scripts stay in `tools/` for this
  repository's own development; only the shipped text stops pointing users at them.
- The scan test that pins criterion 1 (no shipped file names a `tools/` script): it lands
  in the last slice, after every shipped file has been cleaned — before that, the survey
  table in this SPEC is the working list.
- `tools/gen_goal_block.py` and `tools/repo_scans.py` mentions inside the manifest's
  historical `reason` prose: those sentences record what happened (repo_scans is already
  `moai-cli repo-scans` since #2708), and rewriting history lines is not this card's work.

## Requirements

- REQ-1: While slice 1 finishes, the shipped text of `build-folder.md` and `self-check.md` shall name no `tools/` script — the env and cleanup steps name `moai lane-env` (and the planned end-of-task cleanup option), the removal step names `moai optimize-plan` and `moai optimize-apply`, the manifest step moves to this repository's own docs, and the generic rule stays for every user. · failure: none — pinned by AC-1.
- REQ-2: While slice 1 finishes, this repository's own places shall carry what the shipped text dropped — `docs/moai-adk-tools.md` holds the moai-adk-only tool procedures (the repository's self-check runner, the skill-package manifest commands, the smoke harness, the ci_scope classifier, the local check set and the tool notes), so every moved row keeps its procedure in this repository. (`AGENTS.local.md` and the reserved workflow files follow in later slices.) · failure: none — pinned by AC-2.
- REQ-3: While the survey is written into this SPEC, every one of the 21 lines shall carry its move (a, b or c), its target command or destination, and its slice — with the rows whose files are another card's reservation marked for later slices, so no row is lost. · failure: none — pinned by AC-3.
- REQ-9: While `moai check-ready` runs in a project, it shall check only what that project declares — the SPEC documents under the project's `.moai/specs/` folder (a project without them reports the folder as absent and checks nothing) and every path its profile declares under `test_paths` (a declared path that does not exist in the project is a problem) — and answer `result=PASS` with a one-line summary when nothing is wrong, naming each problem when something is, while it never reads a git range nor spawns a test, so it stays separate from `moai check affected`, which selects its checks from a change range. · failure: a project with no profile and no SPEC folder passes with `nothing declared` and still exits 0.

## Acceptance

- AC-1 (REQ-1): `grep -E 'tools/[^ ]+\.(sh|py)'` over `build-folder.md` and `self-check.md` finds nothing, and the rewritten steps name `moai lane-env`, `moai optimize-plan` and `moai optimize-apply` where the tool was named before.
  verify: `grep -nE 'tools/[^ ]+\.(sh|py)' skills/moai/references/internal/build-folder.md skills/moai/references/internal/self-check.md` (no output)
- AC-2 (REQ-2): `docs/moai-adk-tools.md` exists with the moved procedures — the self-check runner, the manifest write/check commands, the smoke areas, the ci_scope classifier notes and the local check set — each named with the shipped file it moved from.
  verify: `grep -n 'self_check.sh\|source_manifest\|smoke/run.sh\|ci_scope.py\|local_check.sh' docs/moai-adk-tools.md` (the moved tools are named in this repository's own docs, not in the shipped text)
- AC-3 (REQ-3): this SPEC's survey table lists all 21 lines with move (a/b/c) and the later-slice marks; the file set matches `git grep` at the slice's tip.
  verify: `moai spec check SPEC-20261120-shipped-workflows-commands`
- AC-4 (REQ-9): in a temporary project the CLI test drives `moai check-ready` through: a project with a valid SPEC and an existing declared test path passes with `result=PASS`; a declared path that does not exist fails naming the path; a broken SPEC fails; a project with neither a profile nor a SPEC folder passes with `nothing declared`; nothing on the project tree is written by any of the runs.
  verify: `cargo test --offline --locked --test check_ready_cli`

## Scope

- write: `skills/moai/references/internal/build-folder.md`, `skills/moai/references/internal/self-check.md`, `docs/moai-adk-tools.md`, this SPEC folder (`REVIEW.md` returned to scope-less state: the grep found no `tools/` reference in it to move); slice 2 adds `src/ops/check_ready.rs`, `tests/check_ready_cli.rs` and the help registrations in `src/main.rs` and `src/help.rs`
- reserved by other cards (later slices take them when returned): `skills/moai/workflows/verify.md` and `skills/moai/references/source-manifest.json` (C-151, place 10), `skills/moai/workflows/mode-agent.md` and `AGENTS.local.md` (C-204), `skills/moai/references/rule-history.md` (C-214)
- read-only: `tools/*.sh`, `tools/*.py`, `src/ops/` (the commands the rows name or will name)
- forbidden: deleting or editing anything under `tools/`, changing the shipped text's Korean voice, a shipped file naming a `tools/` script after its slice lands

## Plan

1. Commit this SPEC with the survey table (REQ-3).
2. Slice 1 text cleanup, this card's own files: `build-folder.md` (env → `moai lane-env`; clean → the lane-env cleanup option named in the later slice, with the relative-path rule kept; remove → `moai optimize-plan`/`moai optimize-apply`) and `self-check.md` (the manifest step moves to `docs/moai-adk-tools.md`; the generic rule stays).
3. `docs/moai-adk-tools.md` collects the moved moai-adk-only procedures; `REVIEW.md` names it where the review flow used a tool script.
4. Slice 2 (REQ-9): `moai check-ready` — the declared checks of the project the command runs in: the SPEC documents and the profile's `test_paths`, never a git range (that is `check affected`'s input, not this command's); tests first in `tests/check_ready_cli.rs`, the logic in `src/ops/check_ready.rs`, the help registrations in `src/main.rs` and `src/help.rs`.
5. Later slices, in file order after the reservations return: `verify.md` (b+c rows), `mode-agent.md` (c + the moved rows), `source-manifest.json` (b rows — only where the manifest itself is edited by that slice's work), `rule-history.md` (b), `AGENTS.local.md` (b) — then the scan test of criterion 1, then the (a) commands one slice each.
6. Neighbours: the skills-related test modules and the manifest test keep passing (the manifest's sha256 entries change only when a shipped file's bytes change — slice 1 changes shipped bytes, so the manifest entries for the two edited files are regenerated in the same commit).

## Risks

- `source-manifest.json` is C-151's reservation while place 10 works: slice 1 changes the bytes of `build-folder.md` and `self-check.md`, so their manifest sha256 entries must be regenerated — that is a manifest edit inside C-151's reservation. The slice therefore lands its text cleanup and the manifest regeneration in one commit only after C-151 returns, or hands the regenerated entries to C-151 as a note (leader decides; this SPEC records the dependency here).
- The verify.md rows (7 of the 21) are the largest block and sit behind the same reservation.
- A shipped sentence that names both the generic rule and the repo-only tool must keep the generic half — the moved text must not take the rule away from users.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | Slice 1 cleans only the files this card holds today (build-folder.md, self-check.md, REVIEW.md, the new docs file) and records the rest as later slices | Five of the six shipped files are other cards' reservations this round (C-151, C-204, C-214) | Take the whole set at once when all return |
| D2 | `build-folder.md`'s steps name existing commands (`moai lane-env`, `moai optimize-plan`, `moai optimize-apply`) before any new command exists | The three steps are already covered by commands; no new Rust work is needed for this row | New commands where the mapping is missing |
| D3 | The manifest sha256 regeneration rides the same commit as the shipped-text change | The manifest test fails otherwise; C-151's reservation decides the timing, not the shape | — |
| D4 | `moai check-ready` checks the project's declarations (the SPEC documents, the profile's `test_paths`), never a change range | The shipped text names it where a user's project needs a pre-ready run; `moai check affected` already owns the change-range selection of this repository's lanes, so one command reads a range and the other reads declarations — no overlap | Widen `check-ready` with flags if a user needs range selection too |

## Result

- AC-1: met (REQ-1) — 배포 두 파일에 `tools/` 이름이 남지 않았어요. `grep -nE 'tools/[^ ]+\.(sh|py)' skills/moai/references/internal/build-folder.md skills/moai/references/internal/self-check.md` — no output, exit 1. 고친 단계는 이름을 실제 있는 명령(`moai lane-env`, `moai optimize-plan`)과 일반 규칙으로 바꾸고, 아직 없는 적용 명령은 약속하지 않았어요.
- AC-2: met (REQ-2) — `docs/moai-adk-tools.md` 가 옮겨진 절차 전부를 담아요. `grep -c 'self_check.sh\|source_manifest\|smoke/run.sh\|ci_scope.py\|local_check.sh' docs/moai-adk-tools.md` = 9. `build_folder.py` 항목은 사실로 고쳤어요(삭제는 이 저장소의 도구, moai 명령은 폴더를 지우지 않음 — 감사 B3). 매니페스트 sha256 재생성도 이 AC 의 일부: `python3 -I tools/source_manifest.py write` 가 `references/internal/build-folder.md`·`references/internal/self-check.md` 두 항목을 갱신하고 `python3 -I tools/source_manifest.py check` 가 통과(exit 0), `cargo test --offline --locked --test product_cli skill_package::` 18 passed(리더 지시: 생성 파일이라 이 카드가 재생성 — 충돌은 머지 때 다시 생성).
- AC-3: met (REQ-3) — 이 SPEC 의 Survey 표가 21 줄 전부를 행별로 담아요(이동·대상·조각 — 위 표). `moai spec check SPEC-20261120-shipped-workflows-commands` — ok true, 0 errors · 0 warnings.
- AC-4: met (REQ-9) — `moai check-ready` 구현·시험 — `cargo test --offline --locked --test check_ready_cli` — 6 passed; 0 failed(선언 통과·없는 선언 경로 실패·깨진 SPEC 실패·선언 없음 성공·쓰기 0·잘못된 프로필 실패). `cargo test --bins` — 28 passed(help 표·예시·OPERATIONS 조각 포함).
- Not run: verify.md·mode-agent.md·rule-history.md·AGENTS.local.md 의 행(above later slices)은 그 파일들이 돌아온 조각에서, 스캔 시험은 마지막 조각에서. closure_status: draft — 승인은 병합 시 사용자 선택 뒤예요.

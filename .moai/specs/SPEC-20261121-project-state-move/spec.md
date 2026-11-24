---
id: SPEC-20261121-project-state-move
title: "One state folder: the update flow moves .moai-cli to .moai with a preview, a token, a receipt and a revert"
status: draft
revision: 2
created: 2026-11-20
parent: null
---

# One state folder: the update flow moves .moai-cli to .moai with a preview, a token, a receipt and a revert

## Intent

판 2 는 NeXT 팁 `02746488f9` 를 다시 읽고 쓴 판이에요. 판 1(`11c3ea4416`)은 같은 가지 조상에 합쳐 있고, 이동 계약(미리보기 → 토큰 → 영수증 → 되돌리기, 마커 마지막)은 그대로 이어요. 달라진 것은 읽은 사실이에요(아래 번호는 모두 이 트리에서 읽은 값이에요).

1. 막다른 길이에요. 이 저장소의 `.moai/` 는 `project.yaml`·`evals/`·`specs/` 셋이고 마커가 없어서 `moai project-update-plan .`(읽기 전용, 2026-11-24)의 `state_root` 판정이 `unknown` 이에요: "`.moai/` holds 3 entries but no generation marker; nothing moves until the user decides". 같은 미리보기의 go-state 단계도 "nothing moves, the user decides outside this route" 로 막히는데 그 "길 밖"이 없어요. 까닭은 도착지로 받는 `.moai/` 가 `specs/` 하나뿐인 모양(`state_root::specs_only`, src/state_root.rs:391-414)인데, 팀이 git 에 두는 종류(`KEPT`, src/state_root.rs:752)는 열한 가지이고 이 저장소는 `project.yaml`·`evals/` 도 추적하기 때문이에요. 이 SPEC 의 첫 일은 이 모양을 길 위로 올리는 거예요.
2. 한 트리에 상태 폴더가 둘로 갈라져 써져요. 판정이 `unknown` 이고 `.moai-cli` 가 있으면 `state_root::dir` 은 `.moai-cli` 를 답해요(`resolves_to_moai`, src/state_root.rs:619-634). 그래서 상태줄 캐시는 `.moai-cli/state/github/counts.json` 을 만들고(`src/statusline/counts.rs:54-56`, 카드 노트 8122) 읽기만 `.moai/project.yaml` 로 가요: `profile::read_dir`·`mixed_layout_ids`(src/profile.rs:546-654)가 두 `project.yaml` 중 새 것을 고르는 임시 처리예요. 이 임시 처리는 이동이 끝난 트리에서만 필요 없어져요.
3. git 이 추적하는 옛 폴더예요. `git ls-files .moai-cli` 는 157개: `manifests/` 19, `project.yaml` 1, `reports/` 137. `.moai/` 는 `specs/` 413, `evals/` 4, `project.yaml` 1. 두 `project.yaml` 은 내용이 달라요(옛 것은 호스트 목록에 codex 가 있고 새 것은 없어요). 이동이 이름 바꾸기이면 git 은 지움 + 새 파일로 봐요.
4. 옮기는 곳이 git 에 가려져요. `git check-ignore -v` 가 `.moai/moai.json` 과 `.moai/manifests/generation.json` 을 `.gitignore:361 /.moai/*` 로 가린다고 답해요(허용 줄은 `specs/`·`evals/`·`project.yaml` 뿐). 그대로 옮기면 추적하던 manifests 가 조용히 사라진 것처럼 보이고, 마커도 새 worktree 가 못 받아서 `.moai-cli` 로 되돌아가요(src/profile.rs:541-545 주석이 같은 증상을 말해요). 그래서 미리보기가 이걸 먼저 알려야 해요.
5. 직접 쓴 이름은 많이 줄어 있어요. 제품 코드(`mod tests` 앞까지, 주석 제외)에서 옛 폴더 이름·상수를 쓰는 곳은 28개 파일, 약 60줄이에요. 정의(src/lib.rs, state_root.rs, home_root.rs, clone_root.rs), 옛 폴더를 읽거나 거르는 곳(src/scope.rs:27, src/check_tree.rs:72, src/project_detect.rs:180, src/apply.rs:456, src/go_state.rs:155), 설명 문구(src/help.rs, src/hooks.rs:605, src/project_update.rs:53·473, src/agents.rs:610·632·652)로 나뉘어요. src/migrate.rs:504·672 와 src/forge/report.rs:367 은 쓰는 곳인지 아직 따라가지 않았어요(미확인). HOME 쪽(`~/.moai-cli`)은 SPEC-20261123-home-state-retire, 클론 공유 상태는 뒤 조각이에요.
6. 시험은 옛 이름을 많이 써요: `tests/fixtures/state_folder_literals/allow.tsv` 는 266개 파일에 옛 이름 1546·새 이름 2013(2026-11-24). 이 SPEC 은 그 시험 파일들을 고치지 않아요.

이 SPEC 이 하는 일: (a) 팀이 추적하는 이름만 든 `.moai/` 를 `unknown` 이 아니라 이 세대의 것으로 읽어요(REQ-1). (b) `/moai update` 미리보기에 이동 단계를 얹어요: `.moai-cli` 의 항목마다 옮김/둠/치움을 사유와 함께 보이고, git 이 가리는 도착지는 막아요(REQ-2~5). (c) 적용은 토큰·이름 바꾸기·마커 마지막·영수증이고 되돌릴 수 있어요(REQ-6~8). (d) 미리보기 단계로 이어 붙여요(REQ-9). (e) 이동이 끝난 트리에서 상태줄 캐시가 `.moai/state/github/` 로 가는 것을 시험으로 못 박아요(REQ-10, 카드 기준 2). (f) 제품 코드가 옛 폴더 이름을 새로 쓰지 못하게 기준선 시험을 둬요(REQ-11). 이 저장소 자신에 이동을 적용하는 일은 이 SPEC 이 아니라 착지 뒤 별도 한 걸음이에요(Plan 7).

## Non-goals

- Go 세대 `.moai/` 내용의 이전(go-state 가 이미 하고), 그 되돌리기의 변경. `go`·`mixed`·`interrupted`·`refuse` 판정과 이름이 낯선 `.moai/` 는 이 이동이 거절하고 go-state 를 가리켜요.
- `~/.moai-cli`(SPEC-20261123-home-state-retire)와 `<git-common>` 아래 클론 공유 상태: 뒤 조각이에요.
- `reports/` 와 `worktrees/` 의 이동: 각각 reports-move 단계와 worktree-move 단계의 몫이고, 이 이동은 `keep` 과 그 주인을 보여요. 추적된 보고서 137개를 git 에서 빼는 일도 여기 없어요(지우기라서 사용자 선택).
- `.gitignore` 편집: 이 도구는 가리는 줄을 알려 주고 막기만 해요. 줄을 고치는 일은 관리 블록 단계(레이아웃 SPEC)나 사람이 해요.
- `git add`·`git commit`: 이동은 작업 폴더만 바꿔요. 커밋은 사람이 해요.
- `profile::read_dir`·`mixed_layout_ids` 같은 두 폴더 임시 처리를 지우는 일: 이동을 아직 안 한 트리가 있어서 남겨요. 지우기는 뒤 카드예요.
- 항목 삭제: 이 이동은 옮기거나 치우거나 두기만 해요.

## Requirements

- REQ-1: The state-root classification shall read a `.moai/` that holds no marker and no Go marker and whose first-level names are all names the team keeps in git as a tree this generation owns instead of `unknown`. · failure: none — a name outside the kept list or a Go marker keeps the verdict `unknown` or `go`.
- REQ-2: The state-move plan shall list every top-level entry of the earlier `.moai-cli/` with the action `move`, `set-aside` or `keep` and one reason each while it reads only. · failure: none — a plan over the entry limit keeps the whole list and says so.
- REQ-3: The plan shall give an entry whose name already exists under `.moai/` the action `keep` with the collision named unless the user picked that entry for `set-aside`. · failure: none — a picked entry moves into the set-aside folder and the `.moai/` entry is never overwritten.
- REQ-4: The plan shall give `reports` and `worktrees` and every name outside the product-state list and every link or special file the action `keep` with the owner of that entry named. · failure: none — the plan does not follow a link or read a file beyond its name and type.
- REQ-5 (REQ-2): IF the target of a `move` entry or of the marker is ignored by a `.gitignore` rule while its source is tracked THEN the plan shall be `blocked` and name the entry and the rule. · failure: none — without a git repository the plan says the tracked count is unknown and skips this check.
- REQ-6: The state-move apply shall run only with the plan digest recomputed at apply time and a permission reference and move each `move` entry by rename without overwrite and write the marker last. · failure: none — a changed folder refuses the apply and asks for a fresh preview.
- REQ-7: The apply shall record a manifest of original and new paths and a receipt whose status starts `pending` and ends `committed` after the marker so that `state_root::dir` answers `.moai` for every later read. · failure: none — an apply that stops before the marker leaves the receipt `pending` and the revert reads it.
- REQ-8: The state-move revert shall take a stamp and read that receipt and remove the marker first and move the recorded entries back and mark the receipt `reverted`. · failure: none — the revert refuses a receipt that is not `committed` or `pending`.
- REQ-9: The project update preview shall show the state-move step with the apply command of its plan while the steps that read the state folder wait for it with their own wait reason. · failure: none — the preview writes nothing and the go-state step no longer says the user decides outside this route for a tree this generation owns.
- REQ-10: The status line counts cache shall be written under the state folder the resolver names. · failure: none — a tree with a marker and a tracked `.moai-cli/` gets `.moai/state/github/counts.json` and no `.moai-cli/state`.
- REQ-11: A scan test shall fail when a product source file outside the listed files spells the earlier state folder name or its constant. · failure: none — a listed file whose count falls still passes.

## Acceptance

- AC-1 (REQ-1): a `.moai/` with `project.yaml` and `specs/` and `evals/` and no marker reads as owned by this generation, while a `.moai/` that also holds a Go marker or an unlisted name keeps its verdict, and `adopt_new` and the project-detect root check follow the same predicate.
  verify: `cargo nextest run --locked --offline --test state_root_cli --test path_root_default`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5): in a temporary git project shaped like this repository (tracked `.moai-cli/manifests` and `reports`, a `project.yaml` on both sides with different bytes, a link, a `.gitignore` with `/.moai/*`) the preview lists every entry with its action and reason, names the collision, names `reports` as owned by the reports step, is `blocked` on the ignored targets, and writes nothing.
  verify: `cargo nextest run --locked --offline --test state_move_cli`
- AC-3 (REQ-6, REQ-7): the apply with the plan digest moves the entries, writes the marker last, and the receipt reads `committed` with the manifest, `state_root::dir` answers `.moai` afterwards even while a kept entry leaves `.moai-cli/` in place, and a changed folder refuses the apply.
  verify: `cargo nextest run --locked --offline --test state_move_cli`
- AC-4 (REQ-8): the revert restores every byte and path, removes the marker first, marks the receipt `reverted`, and a second revert refuses; the same scene runs on a receipt left `pending`.
  verify: `cargo nextest run --locked --offline --test state_move_cli`
- AC-5 (REQ-9): the project update preview of a tree with `.moai-cli/` and a kept-names-only `.moai/` shows the state-move step with its apply command, the reading steps wait with their own reason, and no step says to decide outside this route.
  verify: `cargo nextest run --locked --offline --test project_update_plan_cli`
- AC-6 (REQ-10): with a marker and a git-tracked `.moai-cli/` in a temporary project the counts cache refresh writes `.moai/state/github/counts.json` and creates nothing under `.moai-cli/`.
  verify: `cargo nextest run --locked --offline --test statusline_counts_refresh`
- AC-7 (REQ-11): the scan passes the repository against its baseline list and fails a fixture source that adds a spelling outside the list.
  verify: `cargo nextest run --locked --offline --test path_root_direct_sites`
- AC-8: the wider set around the change passes.
  verify: `cargo nextest run --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-project-state-move/**`, `src/state_move.rs` (new), `src/ops/state_move_plan.rs`, `src/ops/state_move_apply.rs`, `src/ops/state_move_revert_plan.rs`, `src/ops/state_move_revert.rs` (new), `tests/state_move_cli.rs` (new), `src/state_root.rs` (the REQ-1 predicate), `src/project_update.rs` (the step and its wait reasons), `src/lib.rs` and `src/help.rs` (one line each, by line lease), `tests/path_root_direct_sites.rs`, `tests/statusline_counts_refresh.rs`, `tests/state_root_cli.rs`, `tests/project_update_plan_cli.rs`
- read-only: `src/update_cleanup.rs`, `src/go_state.rs`, `src/reports_move.rs`, `src/profile.rs`, `src/statusline/counts.rs`
- forbidden: deleting any entry, following a link or special file, overwriting a `.moai/` entry, `git add` and `git commit` by the tool, editing the user's `.gitignore`, any real HOME state folder, card numbers in outputs

## Plan

1. RED: `tests/state_move_cli.rs` with the repository-shaped scene of AC-2, then the apply and revert scenes. The dangerous path runs in a throwaway project: a link entry, a collision, an ignored target, a changed folder between preview and apply, a stop before the marker.
2. REQ-1: one predicate in `state_root` for "names the team keeps" beside `specs_only`; its five callers (go_state.rs:2005 and 4244, state_root.rs:442 and 715, project_detect.rs:977) are read one by one and each gets a case.
3. Engine `src/state_move.rs` on the plan → digest → receipt → revert contract of `go_state` (`plan_state_moves`, `move_entry`, the marker-last step) and `update_cleanup`; the ops files are thin.
4. Step wiring in the project update preview, with the reading steps waiting (REQ-9).
5. Pin tests for the counts cache (REQ-10) and the direct-site scan (REQ-11), whose baseline lists the sites of Intent 5.
6. `moai check affected` once at READY; guard mutants on digest recompute, marker-last, no-overwrite, the ignore check and the revert order.
7. After landing (not in this SPEC): the leader runs the plan on this repository and records the `.gitignore` allow lines it needs, the choice for the tracked `reports/`, and the commit the user makes.

## Risks

- REQ-1 widens a predicate that `adopt_new` and go-state share; a caller that expected only `specs/` could treat a `project.yaml` as free to overwrite. Plan 2 reads each of the five callers and pins each with a case.
- The move renames git-tracked files, so git shows deletions and additions until the person commits; other worktrees keep the old paths until they merge. The plan counts tracked files per entry so the person sees the size before the pick.
- A running session can write `.moai-cli/state` during the apply. The preview carries the same warning go-state gives; the tool cannot detect a running session.
- The set-aside folder and the receipt live under `.moai/local/`, which the layout rule ignores; where a project's `.gitignore` lacks that rule the receipt would show as untracked, which is harmless.
- Not run in this revision: any build or test (this is a SPEC revision), Windows and Linux rename behavior, and any real HOME.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이동 계약은 update_cleanup 의 plan→토큰→영수증→revert 계약과 go_state 의 신중한 이동(`move_entry`, 마커 마지막)을 따른다 — 새 계약을 만들지 않는다. | src/update_cleanup.rs:1-24, src/go_state.rs:3825 and 4175 | 카드 열림 |
| D-2 | 팀이 추적하는 이름만 든 `.moai/` 는 `unknown` 이 아니라 이 세대의 것으로 읽는다(REQ-1). 아니면 이 저장소 같은 트리는 이동 길이 없다. | `moai project-update-plan .` 의 `unknown` 판정, src/state_root.rs:391 | 판 2 의 값으로 되돌림 |
| D-3 | 같은 이름이 양쪽에 있으면 기본은 `keep`, 사용자가 항목마다 고르면 `set-aside`(`.moai/local/backups/<UTC>/state-move/` 로 치움, 덮어쓰기 없음). `keep` 만이면 이 저장소의 `project.yaml` 충돌이 `.moai-cli` 를 영원히 남겨요. | 설계 보고서 "지우는 일은 미리보기 → 항목마다 선택 → 적용 → 되돌림 기록" | 영수증으로 되돌림 |
| D-4 | 도착지나 마커를 git 이 가리면 계획 전체를 `blocked` 로 한다(조용히 옮기지 않는다). | `git check-ignore -v`: `.gitignore:361 /.moai/*` | 규칙을 고친 뒤 다시 미리봐요 |
| D-5 | 이동은 `.moai-cli` 에 항목이 남아도 마커로 `.moai` 가 앞서게 한다(`keep` 이 남는 동안도 한 폴더로 풀려요). 남은 `reports/` 는 reports 단계와 사용자 선택의 몫이다. | src/state_root.rs:626-627 (`rust` → `.moai`) | 되돌림 |
| D-6 | 카드 기준 2 의 경로 `.moai/state/github/` 를 그대로 쓴다. 레이아웃 규칙의 `.moai/local/state/` 로의 이동은 그 규칙의 `state` 조각이 하고 이 SPEC 이 앞질러 하지 않는다. | src/statusline/counts.rs:54, SPEC-20261120-moai-layout Non-goals | 카드 열림 |

## Result

(status draft — written when the slice lands)

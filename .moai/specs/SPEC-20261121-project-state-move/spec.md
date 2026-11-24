---
id: SPEC-20261121-project-state-move
title: "One state folder: the go-state move takes .moai-cli into a .moai that holds the team's kept names, with a preview, a digest, a receipt and a revert"
status: draft
revision: 3
created: 2026-11-20
parent: null
---

# One state folder: the go-state move takes .moai-cli into a .moai that holds the team's kept names, with a preview, a digest, a receipt and a revert

## Intent

판 3 은 판 2(`65071038b8`, 승인 노트 8307)의 설계를 바꿔요. 코드 조각을 읽다가 `go_state` 에 "옛 `.moai-cli/` 상태만 새 `.moai/` 로" 가는 이동이 끝까지 있다는 것을 알았어요(카드 노트 8309): 계획 `plan_state_moves`(src/go_state.rs:1977), 적용 `apply_legacy_move`(:4224, 대기 영수증 → 항목 이동 → 마커 마지막 → `committed`), 되돌리기 `go-state-revert-plan`·`go-state-revert`, 계획의 `layout` 값 `legacy_into_new`. 임시 폴더에 `.moai-cli/` 만 둔 프로젝트에서 `moai go-state-plan` 이 이 배치와 `state_moves` 를 내는 것을 읽었어요(읽기 전용, 2026-11-24). 리더 결정(노트 8310): 같은 `.moai-cli` 항목을 두 엔진·두 미리보기 단계가 주장하지 않게 이 엔진을 넓혀요. 판 2 의 새 엔진·새 명령·새 단계는 없어져요. 판 1·2 의 이동 계약(미리보기 → 토큰 → 영수증 → 되돌리기, 마커 마지막)은 그대로 이 엔진의 것이에요.

읽은 사실(NeXT 팁 `02746488f9`, 이 트리):

1. 막다른 길이에요. 이 저장소의 `.moai/` 는 `project.yaml`·`evals/`·`specs/` 셋이고 마커가 없어서 `moai project-update-plan .` 의 `state_root` 판정이 `unknown` 이에요. go-state 계획은 판정이 `none` 일 때만 옛 상태 이동을 내요(src/go_state.rs:2816). 그래서 같은 미리보기의 go-state 단계가 "nothing moves, the user decides outside this route" 로 막히고 그 길 밖이 없어요. 까닭은 `.moai/` 가 "`specs/` 하나뿐"일 때만 이 세대의 것으로 읽는 `state_root::specs_only`(src/state_root.rs:391-414)인데 팀이 git 에 두는 종류(`KEPT`, src/state_root.rs:752)는 열한 가지이고 이 저장소는 `project.yaml`·`evals/` 도 추적하기 때문이에요.
2. 도착지가 `specs/` 하나뿐인 `.moai/` 로 한정돼 있어요. 계획은 `specs_only` 일 때만 `specs/` 를 SPEC 단위로 합치고(`plan_spec_merge`, src/go_state.rs:2005·2071), 적용은 `specs_only` 가 아니면 `.moai/` 를 새로 만들어요(:4244-4250, 이미 있으면 "folder create failed"). 판정만 `none` 으로 바꾸면 계획은 `ready` 인데 적용이 실패해요(코드를 읽은 추정, 미실행). 그래서 판정과 엔진은 한 번에 착지해야 해요.
3. 같은 이름 충돌은 SPEC 안에서만 다뤄요. `plan_spec_merge` 는 같은 SPEC id 가 양쪽에 있으면 `conflict` 로 계획 전체를 막고, 다른 이름(`project.yaml`·`evals`·`manifests`)은 충돌을 보지 않아요. 이 저장소는 `project.yaml` 이 두 군데이고 내용이 달라요(옛 것은 호스트 목록에 codex).
4. git 이 가려요. `git check-ignore -v` 가 `.moai/moai.json` 과 `.moai/manifests/generation.json` 을 `.gitignore:361 /.moai/*` 로 가린다고 답해요(허용 줄은 `specs/`·`evals/`·`project.yaml` 뿐). 그대로 옮기면 추적하던 manifests 가 사라진 것처럼 보이고 마커는 새 worktree 가 못 받아서 `.moai-cli` 로 되돌아가요(src/profile.rs:541-545 주석이 같은 증상을 말해요). 계획은 이 도착지를 몰라요.
5. 되돌리기가 `.moai/` 통째로 치워요. 되돌리기 계획은 `legacy_move` 영수증에 이 세대 `.moai/` 를 `.moai.rust-<UTC>` 로 통째로 치우고 `specs/` 만 되돌려 놓아요(src/go_state.rs:4735-4761). 이 저장소처럼 `project.yaml`·`evals/` 가 먼저 있던 `.moai/` 면 팀의 추적 파일이 치워져요.
6. 한 트리에 상태가 둘로 갈라져 써져요. 판정이 `unknown` 이고 `.moai-cli` 가 있으면 `state_root::dir` 은 `.moai-cli`(`resolves_to_moai`, src/state_root.rs:619-634)라서 상태줄 캐시가 `.moai-cli/state/github/counts.json` 을 만들고(`src/statusline/counts.rs:54-56`, 노트 8122), 읽기만 `profile::read_dir`·`mixed_layout_ids`(src/profile.rs:546-654)가 새 `project.yaml` 을 골라요.
7. 추적된 옛 폴더: `git ls-files .moai-cli` 157개(manifests 19, project.yaml 1, reports 137). 계획의 `state_moves` 는 git 이 추적하는 파일 수를 몰라요.
8. 직접 쓴 옛 이름은 제품 코드(`mod tests` 앞, 주석 제외) 28개 파일 약 60줄이고, 정의·옛 폴더 읽기/거르기·설명 문구로 나뉘어요. src/migrate.rs:504·672 와 src/forge/report.rs:367 은 쓰는 곳인지 아직 따라가지 않았어요(미확인). HOME 쪽은 SPEC-20261123-home-state-retire, 클론 공유 상태는 뒤 조각이에요. 시험의 옛 이름은 `tests/fixtures/state_folder_literals/allow.tsv` 가 266개 파일에 옛 이름 1546·새 이름 2013(2026-11-24)이고 이 SPEC 은 그 파일들을 고치지 않아요.

이 SPEC 이 하는 일: (a) 팀이 추적하는 이름만 든 `.moai/` 를 `unknown` 이 아니라 이 세대의 것으로 읽어요(REQ-1). (b) go-state 계획이 그런 `.moai/` 로 옛 상태를 옮기는 `legacy_into_new` 를 내요. 항목마다 옮김/둠/치움을 사유와 함께 보이고 같은 이름 충돌은 기본 `leave`, 사용자가 고르면 `set-aside`, git 이 가리는 도착지는 막고, 추적 파일 수를 세요(REQ-2~5). (c) 적용은 `.moai/` 를 그대로 두고 그 옆에 옮기며 마커를 마지막에 써요(REQ-6). (d) 되돌리기는 적용 전에 `.moai/` 에 있던 이름을 건드리지 않아요(REQ-7). (e) 업데이트 미리보기의 기존 go-state 단계가 그대로 준비 상태로 보이고 계획 호출 모양이 같아서 `src/project_update.rs` 는 바꾸지 않아요(REQ-8). (f) 이동이 끝난 트리에서 상태줄 캐시가 `.moai/state/github/` 로 가는 것을 시험으로 못 박아요(REQ-9, 카드 기준 2). (g) 제품 코드가 옛 폴더 이름을 새로 쓰지 못하게 기준선 시험을 둬요(REQ-10). 이 저장소 자신에 적용하는 일은 이 SPEC 이 아니라 착지 뒤 별도 한 걸음이에요(Plan 8).

## Interview

- none: 사용자에게 묻지 않았어요. 설계는 카드 본문의 사용자 결정(2026-11-20)을 따르고, 판 2 의 선택 D-2~D-6 은 노트 8307, 판 3 의 엔진 선택은 노트 8310 의 리더 결정(사용자 위임 2026-11-23, 리더 판단, 사용자 직접 선택 아님)이에요.

## Non-goals

- Go 세대 `.moai/` 내용의 이전과 그 되돌리기의 변경(`folder_moves`, `folder_stays`). `go`·`mixed`·`interrupted`·`refuse` 판정과 이름이 낯선 `.moai/` 는 이 이동이 거절하고 지금처럼 go-state 의 다른 길이 맡아요.
- 새 명령, 새 모듈, 새 미리보기 단계: 판 2 의 `src/state_move.rs`·`state-move-*` 명령·update 단계는 만들지 않아요.
- `~/.moai-cli`(SPEC-20261123-home-state-retire)와 `<git-common>` 아래 클론 공유 상태: 뒤 조각이에요.
- `reports/` 와 `worktrees/` 의 이동: 각각 reports-move 단계와 worktree-move 단계의 몫이고, 계획은 `leave` 와 그 주인을 보여요. 추적된 보고서 137개를 git 에서 빼는 일도 여기 없어요(지우기라서 사용자 선택).
- `.gitignore` 편집: 이 도구는 가리는 줄을 알려 주고 막기만 해요. 줄을 고치는 일은 관리 블록 단계(레이아웃 SPEC)나 사람이 해요.
- `git add`·`git commit`: 이동은 작업 폴더만 바꿔요. 커밋은 사람이 해요.
- `profile::read_dir`·`mixed_layout_ids` 같은 두 폴더 임시 처리를 지우는 일: 이동을 아직 안 한 트리가 있어서 남겨요. 지우기는 뒤 카드예요.
- `src/project_update.rs` 와 `src/project_detect.rs` 의 변경: 다른 카드가 쓰는 중이라 이 SPEC 의 코드는 건드리지 않아요. `project_detect` 가 같은 술어를 따르는 일은 그 카드 착지 뒤 한 줄이에요.
- 항목 삭제: 이 이동은 옮기거나 치우거나 두기만 해요.

## Requirements

- REQ-1: The state-root classification shall read a `.moai/` that holds no marker and no Go marker and whose first-level names are all names the team keeps in git as a tree this generation owns instead of `unknown`. · failure: none — a name outside the kept list or a Go marker keeps the verdict `unknown` or `go`.
- REQ-2: The go-state plan of a project whose verdict is `none` and whose `.moai/` holds kept names shall list every top-level entry of the earlier `.moai-cli/` with the action `move`, `set-aside` or `leave` and one reason each while it reads only. · failure: none — an earlier root that is a link or not a folder gives no move and the plan says why.
- REQ-3: The plan shall give an entry whose name already exists under `.moai/` the action `leave` with the collision named unless the user picked that entry for `set-aside`. · failure: none — a picked entry moves into the backup area of the apply and the `.moai/` entry is never overwritten.
- REQ-4: The plan shall give `reports` and `worktrees` and every name outside the product-state list and every link or special file the action `leave` with the step that owns that entry named. · failure: none — the plan does not follow a link or read a file beyond its name and type.
- REQ-5 (REQ-2): IF the target of a `move` or `set-aside` entry or of the marker is ignored by a `.gitignore` rule while its source is tracked THEN the plan shall move nothing and name the entry and the rule. · failure: none — without a git repository the plan says the tracked count is unknown and skips this check.
- REQ-6: The go-state apply of a `legacy_into_new` plan shall keep the `.moai/` that exists and move each entry beside its kept names without overwrite and write the marker last. · failure: none — a changed folder or a different set of picks refuses the apply and asks for a fresh preview.
- REQ-7: The go-state revert of such an apply shall move every recorded entry back and remove the marker first and leave every name the `.moai/` held before the apply where it is. · failure: none — the revert refuses a receipt that is not `committed` or `pending` and never sets aside a `.moai/` the apply did not create.
- REQ-8: The project update preview shall show the existing go-state step as ready with its apply command for a tree whose `.moai/` holds only kept names. · failure: none — the preview writes nothing and no step says the user decides outside this route for such a tree.
- REQ-9: The status line counts cache shall be written under the state folder the resolver names. · failure: none — a tree with a marker and a tracked `.moai-cli/` gets `.moai/state/github/counts.json` and no `.moai-cli/state`.
- REQ-10: A scan test shall fail when a product source file outside the listed files spells the earlier state folder name or its constant. · failure: none — a listed file whose count falls still passes.

## Acceptance

- AC-1 (REQ-1): a `.moai/` with `project.yaml` and `specs/` and `evals/` and no marker reads as owned by this generation, while a `.moai/` that also holds a Go marker or an unlisted name keeps its verdict, and `adopt_new` follows the same predicate.
  verify: `cargo nextest run --locked --offline --test state_root_cli --test path_root_default`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5): in a temporary git project shaped like this repository (tracked `.moai-cli/manifests` and `reports`, a `project.yaml` on both sides with different bytes, a link, a `.gitignore` with `/.moai/*`) the plan lists every entry with its action and reason, names the collision, names the owner of `reports`, moves nothing while the targets are ignored, offers `set-aside` for the picked entry, and writes nothing.
  verify: `cargo nextest run --locked --offline --test go_state_cli`
- AC-3 (REQ-6): the apply with the plan digest and the same picks moves the entries beside the kept names, writes the marker last, the receipt reads `committed`, `state_root::dir` answers `.moai` afterwards even while a left entry keeps `.moai-cli/` in place, and a changed folder or different picks refuse the apply.
  verify: `cargo nextest run --locked --offline --test go_state_cli`
- AC-4 (REQ-7): the revert restores every byte and path, removes the marker first, leaves `project.yaml` and `evals/` of the earlier `.moai/` where they were, marks the receipt `reverted`, and the same scene runs on a receipt left `pending`.
  verify: `cargo nextest run --locked --offline --test go_state_cli`
- AC-5 (REQ-8): the project update preview of such a tree shows the go-state step ready with its apply command and no step names a decision outside this route.
  verify: `cargo nextest run --locked --offline --test project_update_plan_cli`
- AC-6 (REQ-9): with a marker and a git-tracked `.moai-cli/` in a temporary project the counts cache refresh writes `.moai/state/github/counts.json` and creates nothing under `.moai-cli/`.
  verify: `cargo nextest run --locked --offline --test statusline_counts_refresh`
- AC-7 (REQ-10): the scan passes the repository against its baseline list and fails a fixture source that adds a spelling outside the list.
  verify: `cargo nextest run --locked --offline --test path_root_direct_sites`
- AC-8: the wider set around the change passes.
  verify: `cargo nextest run --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-project-state-move/**`, `src/go_state.rs`, `src/state_root.rs`, `src/main.rs` and `src/help.rs` (the `--set-aside` argument and its help line, by line lease), `tests/go_state_cli.rs`, `tests/state_root_cli.rs`, `tests/path_root_default.rs`, `tests/path_root_direct_sites.rs`, `tests/statusline_counts_refresh.rs`, `tests/project_update_plan_cli.rs`
- read-only: `src/update_cleanup.rs`, `src/reports_move.rs`, `src/profile.rs`, `src/statusline/counts.rs`, `src/project_update.rs`, `src/project_detect.rs`
- forbidden: deleting any entry, following a link or special file, overwriting a `.moai/` entry, `git add` and `git commit` by the tool, editing the user's `.gitignore`, any real HOME state folder, card numbers in outputs

## Plan

1. RED: scenes in `tests/go_state_cli.rs` for AC-2 to AC-4 on a repository-shaped temporary project. The dangerous path runs in a throwaway: a link entry, a collision with different bytes, an ignored target, a changed folder or different picks between plan and apply, a stop before the marker, a revert of a `pending` receipt.
2. REQ-1: one predicate in `state_root` beside `specs_only` ("only kept names, and no Go marker"); `classify` and `adopt_new` use it, `go_state` (:2005, :4244) uses it for the destination, and each of the four callers in `state_root.rs`·`go_state.rs` gets a case. The fifth caller, `project_detect.rs:977`, follows after C-256 lands.
3. Plan: `plan_state_moves` takes the picks and adds the collision rule (REQ-3), the owner words (REQ-4), the ignore check through a read-only `git check-ignore` (REQ-5) and a tracked count per entry; `plan(project)` stays as the picks-free wrapper so `project_update.rs` and its callers compile and behave as before.
4. Apply and receipt: `apply_legacy_move` keeps a `.moai/` that holds kept names and moves `set-aside` entries into `.moai/backups/<UTC>/`; the receipt records the names the `.moai/` held before, with a default so earlier receipts read as they did.
5. Revert plan: the whole-`.moai/` set-aside happens only when the receipt records no earlier names; otherwise the revert removes the marker first, moves entries back and leaves earlier names.
6. Arguments: `go-state-plan` and `go-state-apply` take the same repeatable `--set-aside <entry>`; the picks are part of the plan digest.
7. Pin tests for the counts cache (REQ-9) and the direct-site scan (REQ-10); `moai check affected` once at READY; guard mutants on digest recompute, marker-last, no-overwrite, the ignore check and the revert order.
8. After landing (not in this SPEC): the leader runs the plan on this repository and records the `.gitignore` allow lines it needs, the choice for the tracked `reports/`, and the commit the person makes.

## Risks

- REQ-1 widens a predicate that `adopt_new`, the classification and go-state share; a caller that expected only `specs/` could treat a `project.yaml` as free to overwrite. Plan 2 reads each caller and pins each with a case.
- `src/go_state.rs` is a 235 KB file other work may hold; the scope request names the conflict as a card note if there is one.
- The receipt gains a field and the revert plan a branch; an earlier `legacy_move` receipt without the field must read exactly as before, which a case pins.
- The move renames git-tracked files, so git shows deletions and additions until the person commits; other worktrees keep the old paths until they merge. The plan counts tracked files per entry so the person sees the size before the pick.
- A running session can write `.moai-cli/state` during the apply. The plan carries the same warning go-state gives; the tool cannot detect a running session.
- Until `project_detect.rs` follows the predicate (after C-256), project detection reads a kept-names-only `.moai/` as a folder, not as absent; the effect is not measured.
- Not run in this revision: any build or test of the engine change (this is a SPEC revision), Windows and Linux rename behavior, and any real HOME.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이동은 go_state 의 `legacy_into_new` 엔진을 넓혀서 한다: 새 엔진·새 명령·새 단계를 만들지 않아요. 같은 `.moai-cli` 항목을 두 엔진이 주장하지 않아요. | 노트 8309·8310, src/go_state.rs:1977·4224 | 판 2 로 되돌림(판 2 의 digest 는 50d64383…) |
| D-2 | 팀이 추적하는 이름만 든 `.moai/` 는 `unknown` 이 아니라 이 세대의 것으로 읽는다(REQ-1). | 노트 8307 D-2 (a) | 판 2·3 의 값으로 되돌림 |
| D-3 | 같은 이름이 양쪽에 있으면 기본 `leave`, 사용자가 항목마다 고르면 `set-aside`(`.moai/backups/<UTC>/` 로 치움, 덮어쓰기 없음). | 노트 8307 D-3 (a) | 영수증으로 되돌림 |
| D-4 | 도착지나 마커를 git 이 가리면 아무것도 옮기지 않는다. | 노트 8307 D-4 (a), `git check-ignore -v` | 규칙을 고친 뒤 다시 미리봐요 |
| D-5 | 이동은 `.moai-cli` 에 항목이 남아도 마커로 `.moai` 가 앞서게 한다. 남은 `reports/` 는 reports 단계와 사용자 선택의 몫이다. | 노트 8307 D-5 (a), src/state_root.rs:626-627 | 되돌림 |
| D-6 | 카드 기준 2 의 경로 `.moai/state/github/` 를 그대로 쓴다. 레이아웃 규칙의 `.moai/local/state/` 로의 이동은 그 규칙의 `state` 조각이 하고 이 SPEC 이 앞질러 하지 않는다. | 노트 8307 D-6 (a) | 카드 열림 |
| D-7 | 되돌리기는 적용 전에 `.moai/` 에 있던 이름을 건드리지 않는다. 영수증이 그 이름을 기록하고 기록이 없는 옛 영수증은 지금처럼 읽는다. | src/go_state.rs:4735-4761 | 카드 열림 |

## Result

(status draft — written when the slice lands)

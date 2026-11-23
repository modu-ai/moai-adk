---
id: SPEC-20261124-statusline-cache-untracked
title: "The repository .gitignore ignores the statusline counts cache so no session or worktree shows an untracked .moai-cli/state/"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7689 (2026-11-23T16:32:27Z, 사용자 위임 2026-11-23T13:57Z, 노트 7688 에 대한 답, 옵션 A 승인 + /.moai/state/ 한 줄 추가 + 카드 쓰기 범위를 .gitignore 와 tests/gitignore_block_cli.rs 로 넓힘); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요. 사용자 직접 선택이 아니라 리더 판단이에요"
  digest: "sha256:a7aee7feb4bf699c4b6fbddd22949624b36c3efb2e8d9cbbdf822b940efb8b32"
---

# The repository .gitignore ignores the statusline counts cache so no session or worktree shows an untracked .moai-cli/state/

## Intent

카드가 말한 현상: 워크트리에 추적되지 않는 `.moai-cli/state/github/counts.json` 이 생겨요. 카드는 "어떤 테스트가 임시 폴더 대신 작업 폴더에 쓴다"고 봤어요. 코드와 측정으로 읽어 보니 쓴 쪽은 테스트가 아니라 **세션의 진짜 상태줄**이에요. 그래서 이 SPEC 은 카드의 전제를 고쳐서 적고, 고칠 곳을 좁혀요.

### 찾은 것 (2026-11-24 에 NeXT 끝 `05fec214a8` 에서 읽고 잰 것)

1. 쓰는 코드: `src/statusline/counts.rs` 의 `refresh` → `write` 가 `state_root::path(anchor, "state/github")/counts.json` 을 써요(`cache_path`). 자식은 `maybe_refresh`(`src/statusline/counts.rs`)가 `statusline --refresh-counts --board-root <anchor>` 로 띄워요. 시작 조건은 셋: 실행 파일이 제품 이름(`moai-cli`/`moai`), `origin` 이 github.com, 캐시가 10분(TTL)보다 오래됨. `anchor` 는 상태줄 입력의 `workspace.project_dir` 에서 와요(`resolve_anchor`, `src/statusline/state.rs`).
2. 이 프로젝트의 `.claude/settings.json` 과 `.claude/settings.local.json` 의 `statusLine.command` 가 `"$HOME/.moai-cli/path/moai-cli" statusline` 이에요(local 은 `refreshInterval: 60`). 즉 레인·리더 세션이 열려 있는 워크트리마다 상태줄이 그 워크트리를 anchor 로 삼아 실제 `gh` 로 숫자를 받아 써요.
3. 측정 A (테스트 없이 생김): 이 카드의 새 워크트리(`c416-statusline-cache-state`)를 만든 직후, 테스트를 한 번도 돌리지 않았는데 `.moai-cli/state/github/counts.json`(80바이트, `{"open_issues":1,"open_prs":0,…}`)이 생겼고 `git status --short` 에 `?? .moai-cli/state/` 가 보였어요(`date -u` 16:29Z, 파일 시각 01:29:18 KST).
4. 측정 B (다른 워크트리): C-409 워크트리 `c409-identity-home-arg` 는 00:05:25 에 만들어졌고 `.moai-cli/state/` 는 00:05:33(8초 뒤, 세션의 첫 상태줄)에 생겼어요. 카드가 본 00:56 은 마지막 10분 갱신 시각이에요(파일은 갱신마다 rename 으로 바뀌어 생성 시각도 따라가요). `.claude/worktrees/` 의 다른 워크트리 여럿에도 같은 파일이 있고 값은 실제 저장소 `modu-ai/moai-adk` 의 열린 Issue 1·PR 0 이에요.
5. 상태줄을 실제 바이너리로 돌리는 테스트들(`tests/statusline_*.rs`, `tests/status_line_cli.rs`, `tests/usage_limit_cli.rs`, `tests/statusline_integration_distance.rs`, `tests/diff_go.rs`)은 입력의 `project_dir`/`cwd` 가 임시 프로젝트이거나 아예 없어요. `tests/statusline_counts_refresh.rs` 는 임시 저장소 + `PATH` 에 가짜 `gh` 만 둬요. 이 읽기에서 작업 폴더를 anchor 로 삼는 테스트는 찾지 못했어요(정적 읽기; 아래 Risks R-1).
6. 사용자 프로젝트는 이미 덮여 있어요: 관리 블록(`moai-local`, `src/agents.rs` `IGNORE_SECTION`)은 `.moai-cli/state/…` 를 기계 출력으로 무시하고 `tests/gitignore_block_cli.rs` 의 `MACHINE_FILES` 가 `.moai-cli/state/lane.json` 으로 그걸 고정해요.
7. 빈틈은 이 저장소 하나예요: 손으로 쓴 `.gitignore` 가 `/.moai-cli/worktrees/`(381행)와 `/.moai-cli/receipts/`(384행)만 무시하고 `/.moai-cli/state/` 는 무시하지 않아요(`.moai/state/` 는 242행에서 무시). 그래서 `.moai-cli/` 세대 상태 폴더를 쓰는 이 저장소의 워크트리마다 `?? .moai-cli/state/` 가 보여요. 이 저장소의 `.gitignore` 를 고정하는 테스트는 이미 있어요: `the_repository_gitignore_tracks_specs_and_ignores_the_rest`(`tests/gitignore_block_cli.rs`).

### 결정이 필요한 것

카드의 지시("테스트가 임시 프로젝트를 쓰게 고치고 작업 폴더에 새 파일이 없음을 지키기")는 위 3~5번과 맞지 않아요. 고칠 대상이 테스트가 아니므로 범위를 바꾸는 선택이라서 승인을 받아요(CLAUDE.md "코드와 출력이 지시와 다르면 코드가 이겨요").

- 옵션 A (기본, 권장): 이 저장소의 `.gitignore` 에 `/.moai-cli/state/` 한 줄을 더하고, 기존 저장소 `.gitignore` 테스트 옆에 그 줄을 고정하는 테스트를 더해요. 제품 코드 변경 없음. 한 줄을 지우면 되돌려져요.
- 옵션 B: 상태줄 캐시와 잠금 파일을 프로젝트 폴더 밖(상태줄이 이미 쓰는 `$TMPDIR/moai-statusline/` 캐시 폴더, `owner/name` 키)으로 옮겨요. 어느 프로젝트에서도 작업 폴더에 파일이 안 생기고 워크트리마다 같은 숫자를 따로 받아 오는 일도 줄어요. 제품 코드(`counts.rs`, `state.rs` 읽기, `card/early_reclaim.rs` 의 재생성 가능 목록)와 테스트 대여섯 파일을 건드려 따로 카드가 필요하고, 사용자 프로젝트는 옵션 A 의 관리 블록으로 이미 덮여서 얻는 것이 적어요.
- 옵션 C: 낯선 파일만 지우기. 거절: 세션마다 다시 생기고(측정 A·B), 이 레인은 `~/.moai-cli` 와 루트 체크아웃의 파일을 지우지 않아요.
- 기본값: 옵션 A. 되돌릴 수 있고 이 저장소 안이며 HOME·삭제·푸시·릴리스가 아니에요. 옵션 B 는 이 SPEC 에서 하지 않고 필요하면 따로 카드를 열어요.

### 승인 결과 (리더 결정 노트 7689)

옵션 A 가 승인됐어요. 리더가 덧붙인 것: `.gitignore` 에 `/.moai-cli/state/` 와 함께 `/.moai/state/` 도 한 줄 더해요(상태 폴더가 `.moai/` 로 옮겨 간 뒤 같은 일이 다시 생기지 않게. 리더가 NeXT 를 읽어 `src/lib.rs:257` 의 `PROJECT_STATE_DIR` 가 아직 `.moai-cli` 임을 확인했어요). `.moai/` 의 `specs`·`project.yaml`·`evals` 는 계속 추적되고 테스트가 그걸 함께 지켜요. 옵션 B 는 이 카드에서 하지 않고 옵션 C 는 거절했어요. 카드 쓰기 범위를 `.gitignore` 와 `tests/gitignore_block_cli.rs` 로 넓히는 것도 승인됐어요(`moai card` 범위 명령으로 하나씩).

이 레인이 읽은 것 하나: 이 저장소의 `.gitignore` 는 이미 `.moai/state/`(242행)와 `/.moai/*`(361행, `specs`·`project.yaml`·`evals` 만 예외)로 `.moai/state/` 를 무시해요. 그래서 `/.moai/state/` 줄은 지금 효과가 겹쳐요. 승인된 대로 더하고, 겹침은 보고에 적어요.

## Interview

- Q1: 카드는 "테스트가 작업 폴더에 쓴다"고 했는데 고칠 대상이 테스트인가요? → A: 아니에요. 쓰는 쪽은 세션의 진짜 상태줄이에요(측정 A·B). 그래서 테스트를 임시 프로젝트로 고치지 않고 이 저장소의 `.gitignore` 빈틈을 닫아요(리더 노트 7689 가 옵션 A 로 승인).
- Q2: 사용자 프로젝트도 고치나요? → A: 안 고쳐요. 관리 블록(`moai-local`)이 이미 `.moai-cli/state/` 를 무시해요(`tests/gitignore_block_cli.rs` 의 `MACHINE_FILES`).
- Q3: `/.moai/state/` 줄도 더하나요? → A: 더해요(노트 7689 의 덧붙임). 지금은 242행과 361행 규칙으로 이미 무시돼서 효과가 겹쳐요. 상태 폴더가 `.moai/` 로 옮겨 간 뒤에도 규칙이 이름으로 남게 하는 보험이고, 테스트가 `.moai/state/github/counts.json` 무시와 `specs`·`project.yaml`·`evals` 추적을 함께 지켜요.
- Q4: 카드 범위는요? → A: SPEC 폴더로 잡았고, `.gitignore` 와 `tests/gitignore_block_cli.rs` 를 `moai card scope` 로 하나씩 넓혀요(노트 7689).

## Non-goals

- 상태줄 코드(`src/statusline/**`)의 변경, 캐시 위치 이동(옵션 B), `refresh` 자식의 시작 조건 변경.
- `~/.moai-cli`, 루트 체크아웃, 다른 워크트리에 이미 있는 `.moai-cli/state/` 파일의 삭제나 수정.
- 사용자 프로젝트의 관리 블록(`src/agents.rs`)의 변경(이미 `state/` 를 덮어요).
- "테스트가 작업 폴더에 파일을 안 남긴다"는 일반 감시 장치: 이 읽기에서 그런 테스트를 찾지 못했어요. 비슷한 재발 방지는 C-406(실제 HOME 을 읽는 테스트 잡기)이 맡아요.

## Requirements

- REQ-1: The repository `.gitignore` shall ignore `.moai-cli/state/` and `.moai/state/` at the repository root with rules that carry a comment naming the statusline counts cache.
- REQ-2: A test next to `the_repository_gitignore_tracks_specs_and_ignores_the_rest` in `tests/gitignore_block_cli.rs` shall fail when git, reading the repository `.gitignore`, does not ignore `.moai-cli/state/github/counts.json`, `.moai-cli/state/github/refresh.lock`, a temporary `.moai-cli/state/github/counts.json.<pid>.tmp` or `.moai/state/github/counts.json`.
- REQ-3: The same test shall fail when git ignores a team file that the repository tracks: `.moai-cli/project.yaml`, `.moai-cli/manifests/profile-baseline.json`, `.moai/project.yaml`, a file under `.moai/specs/` or a file under `.moai/evals/`.
- REQ-4: The change shall touch no file other than the SPEC folder, `.gitignore` and `tests/gitignore_block_cli.rs`.
- REQ-5: The lane shall delete or modify no existing `.moai-cli/state/` file in this worktree, in the primary checkout or in the HOME state folder.

## Acceptance

- AC-1 (REQ-1, REQ-2): git ignores the three cache paths with the repository `.gitignore` as it stands at the head · verify: `cargo test --locked --offline --test gitignore_block_cli the_repository_gitignore_ignores_the_statusline_counts_cache`
- AC-2 (REQ-3): the same test keeps the five tracked team paths of REQ-3 unignored, and `the_repository_gitignore_tracks_specs_and_ignores_the_rest` stays green · verify: `cargo test --locked --offline --test gitignore_block_cli`
- AC-3 (REQ-4, REQ-5): the range changes exactly the three paths and no `.moai-cli/state/` file · verify: `git diff --name-only <base>..<head>`
- AC-4 (REQ-1): in a worktree of the head, a statusline refresh leaves `git status --short` without `.moai-cli/state` · verify: the READY note records `git status --short` at the head after the lane's own session has rendered its status line (observed, not asserted from a test).

## Scope

- write: `.moai/specs/SPEC-20261124-statusline-cache-untracked/**`, `.gitignore`, `tests/gitignore_block_cli.rs`
- read-only: `src/statusline/counts.rs`, `src/statusline/state.rs`, `src/agents.rs`, `.claude/settings.json`
- forbidden: `src/**` (every change), the HOME state folders, the primary checkout's files, any deletion of a file named in REQ-5

The card's own scope is empty and the lane took it with the SPEC folder only; `.gitignore` and `tests/gitignore_block_cli.rs` are added one at a time with `moai card` (scope widen) on the card's reservation, as the leader's note 7689 approved.

## Plan

1. This SPEC, with the approval request as `LANE-BLOCKED kind=decision` on the card. The lane writes no approval; the leader's note 7689 is quoted in the front matter.
2. Widen the card's write scope to the two paths, one at a time.
3. RED test `the_repository_gitignore_ignores_the_statusline_counts_cache`, run once to show it fails on the unchanged `.gitignore`.
4. The one-line `.gitignore` rule; the test turns green.
5. Guard mutant: remove the rule, the test fails with a non-zero exit; restore.
6. `cargo test --locked --offline --test gitignore_block_cli` under `moai heavy-slot-acquire --max 4`, then `moai spec check`, `moai trailer-check`, one bare `moai check affected <base>..<head>` at the READY head, then the READY note and `LANE-DONE`.

## Risks

- R-1: "No test writes it" is a static reading plus two measurements, not a full run: a `check affected` over a checkout where no Claude Code session renders a status line was not run. If a test did write the file, AC-1's rule would still hide it from `git status` but would not stop the write; the leader can then open a card for that test with its name.
- R-2: A tier-2 floor path may be touched (`moai bundle check` names the audit tier at READY); the lane does not pick the tier.
- R-3: The rule hides every future file under `.moai-cli/state/` from `git status`, including a file someone meant to track there. `git ls-files .moai-cli/state` lists none today (checked at the base).

## Result

Status: SPEC approved by the leader (revision 1, note 7689, a leader judgment and not the user's own pick); the slice is built. No product code changed.

As built:
- `.gitignore`: `/.moai-cli/state/` and `/.moai/state/` after the receipts rule, with a comment naming the status line cache and this SPEC.
- `tests/gitignore_block_cli.rs`: the test `the_repository_gitignore_ignores_the_statusline_counts_cache`. It reads the repository `.gitignore` in a temporary git repository without global rules, asserts that the cache, the lock, the temporary file and `.moai/state/github/counts.json` are ignored, that the deciding rule of the two cache paths is the added line (`git check-ignore -v`), and that five tracked team paths stay unignored.
- The test was committed alone first and failed on the unchanged `.gitignore` (`.moai-cli/state/github/counts.json is not ignored`, exit 101), then the rule turned it green.

Results (each from a command run on this branch):
- AC-1 met: `cargo test --locked --offline --test gitignore_block_cli the_repository_gitignore_ignores_the_statusline_counts_cache` passes; the whole file passes, 33 tests, exit 0.
- AC-2 met: the same test keeps `.moai-cli/project.yaml`, `.moai-cli/manifests/profile-baseline.json`, `.moai/project.yaml`, `.moai/specs/SPEC-x/spec.md` and `.moai/evals/audit/cases-tuning.json` unignored, and `the_repository_gitignore_tracks_specs_and_ignores_the_rest` stays green in the 33.
- AC-3 met: `git diff --name-only 05fec214a8..HEAD` lists the SPEC file, `.gitignore` and `tests/gitignore_block_cli.rs` only; no `.moai-cli/state/` file was deleted or modified.
- AC-4 met: in this worktree the real cache file the status line wrote at 01:29:18 KST is listed by `git status --short --ignored` as `!! .moai-cli/state/`, `git check-ignore -v` names `.gitignore:389:/.moai-cli/state/`, and `git status --short` prints no line.

Guard mutants (each applied by an edit, confirmed with `git diff --stat`, run on the test above, then restored with `git checkout -- .gitignore`): removing `/.moai-cli/state/` fails (exit 101); widening it to `/.moai-cli/` fails (exit 101, decided by another rule); adding `/.moai-cli/project.yaml` fails (exit 101, `.moai-cli/project.yaml is ignored`); removing `/.moai/state/` first survived (exit 0) because the older `.moai/state/` and `/.moai/*` rules ignore the same path, so the test was changed to assert the deciding rule by name and the same mutant then fails (exit 101).

Gaps: a full `check affected` over a checkout where no status line renders was not run, so "no test writes the cache" stays a static reading plus two measurements (R-1). The `/.moai/state/` line has no effect today (older rules already ignore the path); it is the approved insurance for the state move.

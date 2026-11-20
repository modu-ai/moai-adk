---
id: SPEC-20261120-moai-layout
title: "One .moai/ layout rule for every project: tracked team files, ignored .moai/local/, project tools in .moai/tools/"
status: draft
revision: 1
created: 2026-11-20
---

# One .moai/ layout rule for every project

## Intent

사용자 결정 2026-11-20(리더 대화, 선택): 모든 MoAI 프로젝트가 같은 `.moai/` 구조를 써요. 설계 보고서는 `~/.moai/reports/moai-adk/moai-layout-and-worktree-disk-design-2026-11-20.md` 4.1 이에요.

```
.moai/
  moai.json  project.yaml                 표지·프로필 (추적)
  config/ decisions/ docs/ specs/ manifests/ workflows/   (추적)
  tools/                                  프로젝트 전용 개발 지원 스크립트 (추적, 새로)
  evals/                                  평가 세트 (추적, 필요할 때)
  local/                                  이 기계의 것 (무시): receipts/ backups/ index/ logs/ state/ cache/ …
```

까닭: 무시되는 것은 worktree 에 복제되지 않아요(추적 파일은 worktree 를 만들 때마다 체크아웃돼요). 그리고 MoAI 와 에이전트가 만드는 개발 지원 파일이 프로젝트 루트(`scripts/` 등)에 프로젝트 코드와 섞이지 않아요. 호스트가 자리를 정한 파일(`AGENTS.md`, `.claude/`, `.codex/`, `.github/ISSUE_TEMPLATE/moai-*`)은 그 자리에 둬요.

이미 있는 프로젝트는 계속 돌아가요: 읽기는 옛 경로까지, 쓰기는 새 경로로.

이 SPEC 은 규칙 하나를 여러 조각으로 나눠 들여요. 조각 1 은 규칙의 정의(코드 한 곳)와 관리되는 `.gitignore` 블록이에요. 쓰기 쪽을 `local/` 로 옮기는 일은 조각 2 이후이고, 안내 문구는 조각 3 이에요.

## Non-goals

- `.moai-cli` → `.moai` 이전 마무리(미리보기·토큰·되돌리기, 두 폴더 공존 해소)는 하지 않아요: 따로 다뤄요.
- 보고서·증거 위치는 정하지 않아요: 프로젝트 밖 `~/.moai/reports/<project>/` 이고 다른 SPEC 이 다뤄요.
- moai-adk 자체의 `tools/`(제품 개발 코드)는 루트에 그대로 둬요.
- 호스트 소유 파일의 자리는 바꾸지 않아요.
- worktree 개수 상한·빌드 폴더 정리는 하지 않아요.
- 옛 경로의 파일을 지우거나 옮기지 않아요(읽기 호환만).

## Requirements

- REQ-1: The state layout shall name the kept kinds `project.yaml`, `moai.json`, `moai-cli.json`, `config`, `decisions`, `docs`, `specs`, `manifests`, `workflows`, `tools` and `evals`, and every other kind is local to the machine and lives under `local/`.
- REQ-2: The state layout shall give one function that returns the folder of a local kind under the state folder's `local/` and one that returns the folder where an earlier project kept it, so no writer spells either path itself.
- REQ-3: The managed `.gitignore` block shall ignore the state folder's `local/` explicitly and allow `tools/` and `evals/` beside the earlier kept list.
- REQ-4: When a project's managed `.gitignore` block lacks the new lines, the project update plan shall show the block as changed.
- REQ-5: When a reader looks for a local kind, the reader shall read `local/<kind>` first and the earlier `<state>/<kind>` when the new one is absent.
- REQ-6: A writer that creates a local kind shall create it under `local/<kind>`.
- REQ-7: The generated `AGENTS.md` text and the skill guidance shall name `.moai/tools/` as the place for project development scripts.
- REQ-8: The layout shall leave the host-owned files `AGENTS.md`, `.claude/`, `.codex/` and `.github/ISSUE_TEMPLATE/moai-*` where the host reads them.
- REQ-9 (REQ-4): IF the user has not applied the project update plan, THEN the project update shall leave the managed `.gitignore` block unchanged.
- REQ-10 (REQ-5): IF neither the new nor the earlier path exists, THEN the reader shall report the kind as absent.
- REQ-11: A writer that creates a local kind shall leave an earlier `<state>/<kind>` untouched.
- REQ-12: The generated `AGENTS.md` text and the skill guidance shall not name the project root or `scripts/` as the place for project development scripts.

## Acceptance

- AC-1 (REQ-1, REQ-2): 레이아웃 정의가 추적 종류 열한 개를 이름으로 대고 나머지는 `local/` 아래라고 하며, `local` 종류의 새 경로와 옛 경로를 한 함수씩 돌려줘요.
  verify: `cargo test --locked --offline --lib state_root::tests::the_layout_names_kept_kinds_and_local_paths`
- AC-2 (REQ-3): 새 프로젝트의 관리 블록은 `<state>/local/` 을 무시하고 `tools/`·`evals/` 를 허용하고, git 이 그 두 폴더의 새 파일은 추적 후보로·`local/` 안은 무시로 본다고 답해요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_local_and_keeps_tools_and_evals`
- AC-3 (REQ-4, REQ-9): 옛 블록이 있는 프로젝트의 업데이트 계획은 블록이 바뀐다고 보여 주고, 적용 전에는 파일이 그대로예요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_earlier_block_is_shown_as_changed_and_left_until_applied`
- AC-4 (REQ-5, REQ-6, REQ-10, REQ-11): 읽는 쪽은 `local/<kind>` 를 먼저 읽고 없으면 옛 경로를, 쓰는 쪽은 새 경로에만 써요(조각 2 에서 종류마다 시험이 더해져요).
  verify: `cargo test --locked --offline --lib state_root::tests::a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`
- AC-5 (REQ-7, REQ-12): 만들어지는 `AGENTS.md` 와 스킬 지침이 `.moai/tools/` 를 개발 스크립트 자리로 말하고 프로젝트 루트나 `scripts/` 를 그 자리로 말하지 않아요(조각 3).
  verify: `cargo test --locked --offline --test repo_agents_md the_guidance_names_moai_tools_for_development_scripts`
- AC-6 (REQ-8): 호스트 소유 파일 네 곳은 레이아웃 정의의 이동 대상에 없고, 만드는 쪽도 그 자리에 둬요.
  verify: `cargo test --locked --offline --lib state_root::tests::host_owned_files_are_not_in_the_layout`

## Scope

- write (조각 1): `.moai/specs/SPEC-20261120-moai-layout/spec.md`, `src/state_root.rs`, `src/agents.rs`(관리 블록 `STATE_KEPT` 와 `render_ignore`), `tests/gitignore_block_cli.rs`, 골든 `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`·`17-agents-apply-ok.json`(블록 줄이 바뀌어서), 규칙 설명 한 조각(`docs/design-review/appendix/` 아래)
- write (조각 2 이후): 쓰는 쪽 모듈 — `src/apply.rs`, `src/index.rs`, `src/go_state.rs`, `src/update_hooks.rs`, `src/optimize.rs`, `src/forge/*`, `src/lane_loop*` 등 — 은 조각마다 따로 예약해요
- write (조각 3): `agents/*.md`, `skills/moai/references/**`, `src/agents.rs` 의 AGENTS.md 생성 문구, 매니페스트
- read-only: `src/layout.rs`(P09 홈 배치와는 별개예요), `src/go_state.rs`(옛 이름 목록)
- forbidden: 옛 경로 파일 이동·삭제, `.moai-cli` 이전 변경, 보고서 위치 변경

## Plan

조각 1(이번 설계의 첫 구현):

1. SPEC 을 커밋해요(이 문서). `spec check` 와 `spec check --closure`(Result 는 구현 때 채워요)를 돌려요.
2. RED: `state_root` 단위 시험(`the_layout_names_kept_kinds_and_local_paths`, `a_reader_falls_back_to_the_earlier_path_and_a_writer_uses_the_new_one`, `host_owned_files_are_not_in_the_layout`)과 `gitignore_block_cli` 시험 둘을 먼저 써서 실패를 봐요.
3. GREEN: `src/state_root.rs` 에 `KEPT`(열한 종류)와 `local_path`·`earlier_path`·`read_path`(새 경로 우선, 없으면 옛 경로)를 두고, `src/agents.rs` 의 `STATE_KEPT` 를 거기서 가져오게 해요. 블록에 `<state>/local/` 줄을 더하고 `tools/`·`evals/` 허용 줄을 `specs/` 뒤에 붙여요(앞의 아홉 줄 순서는 그대로).
4. 확인: 두 골든을 `moai` 출력에서 다시 만들고 `git diff` 로 바뀐 줄이 블록 줄뿐인지 봐요. `agents` 계열·`project_update`·`gitignore_block_cli`·골든 시험 전체, fmt, 운영 파일 rustfmt, clippy `--all-targets`, `moai check affected`.

조각 2: 쓰는 쪽을 종류별로 `local/` 로 옮겨요(`receipts` → `index` → `state`·`forge/outbox` → `backups`·`lane-stop-hook`·`project-skill-rollback` → 나머지). 한 조각에 한 종류 묶음, 읽기 호환 시험과 함께예요.

조각 3: 안내 문구(`AGENTS.md` 생성 문구와 스킬 지침)가 `.moai/tools/` 를 말해요.

## Risks

- 관리 블록의 줄이 바뀌어서 이미 블록이 있는 프로젝트는 다음 업데이트 계획에서 "블록이 바뀐다"를 보게 돼요: 적용은 사용자가 계획을 확인한 뒤에만이에요. 골든 두 개(`16-agents-plan-create`, `17-agents-apply-ok`)와 `gitignore_block_cli` 의 줄 수 기대가 함께 바뀌어요.
- 모든 상태가 이미 `<state>/*` 규칙으로 무시되고 있어서(허용 목록 밖은 무시) `local/` 줄은 읽는 사람에게 규칙을 말해 주는 줄이에요. 무시 결과는 달라지지 않지만 사용자 규칙이 `local/` 을 허용으로 뒤집지 못하게 하는 효과는 확인해야 해요.
- 영수증·결과물이 옛 경로 문자열(`.moai-cli/receipts/…`)을 기록해요: 쓰는 쪽을 옮기면 기록된 경로가 새 경로가 되고, 옛 기록은 읽기 호환(REQ-5)으로만 읽혀요. 되돌림 도구(`go_state` 백업·`update_cleanup`)가 기록된 경로를 따라가므로 조각 2 에서 종류마다 되돌림 시험을 둬요.
- 쓰는 쪽이 서른 곳 넘게 흩어져 있어요(`state_root::dir`·`path`·`rel` 사용처 31곳, 직접 문자열 경로 여러 곳): 한 번에 바꾸지 않고 종류별로 나눠요. 그 사이에 같은 종류를 두 길로 쓰는 코드가 남지 않게 종류 하나는 한 조각에서 끝내요.
- `go_state` 의 옛 이름 목록(`archive`, `evidence`, `settings-backups` 등)과 `local` 종류 목록이 어긋나면 이전 도구가 새 폴더를 모르는 폴더로 봐요: 조각 1 이 두 목록의 관계를 시험으로 못 박아요.
- `.moai-cli` 와 `.moai` 가 함께 있는 프로젝트(이 저장소 포함)에서 `local/` 이 어느 쪽 아래인지: `state_root::dir` 가 정하는 쪽 아래예요. 이전 마무리는 이 SPEC 밖이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 추적 종류는 열한 개 이름을 코드 한 곳에 두고, 그 밖의 모든 종류는 `local/` 아래로 본다 | 사용자 결정: "추적은 팀 파일, 나머지는 이 기계의 것" — 목록이 닫혀 있어야 새 종류가 실수로 추적되지 않아요 | 목록에 이름 더하기 |
| D-2 | 새 경로 함수와 옛 경로 함수를 한 모듈에서 주고 읽기 도우미는 새 경로 우선·없으면 옛 경로 | 사용자 결정: "옛 경로를 읽고 새 경로에 쓴다" | 도우미가 옛 경로만 쓰게 하기 |
| D-3 | 관리 블록의 `tools/`·`evals/` 허용 줄은 `specs/` 뒤에 붙인다 | 앞의 아홉 줄 순서가 그대로여야 이미 쓰인 블록과 줄 단위로 비교돼요 | 두 줄 지우기 |
| D-4 | 쓰는 쪽 이동은 종류 묶음별 조각으로 한다 | 쓰는 곳이 서른 곳 넘고 각각 되돌림·영수증이 걸려 있어요 | 조각 단위로 되돌리기 |
| D-5 | 이 저장소(moai-adk)의 루트 `tools/` 는 옮기지 않는다 | 사용자 결정(설계 D2): 제품 개발 코드이고 CI·검사 보호 경로가 가리켜요 | 해당 없음 |

## Result

설계 단계(2026-11-20): 조각 1 시작 전. 구현이 끝나면 AC 마다 `- AC-n: met — <근거>` 줄을 이 자리에 적어요.

---
id: SPEC-20261120-worktree-project-folder
title: "Worktrees live at ~/.moai/worktrees/<project>/<name>/; older forms stay recognised"
status: draft
revision: 1
created: 2026-11-20
---

# 워크트리 위치 — `~/.moai/worktrees/<project>/<name>/` 로, 옛 모양은 계속 알아본다

## Intent

사용자 결정(2026-11-20, 리더 대화): 워크트리를 `~/.moai/worktrees/<project>-<id>/` 평평한 모양에서
`~/.moai/worktrees/<project>/<name>/`(프로젝트마다 폴더 하나)로 바꿔요. 옛 평평한 모양·Go 세대의
`worktrees/projects/<project>/<name>`·저장소 안 옛 위치(`.moai-cli/worktrees/`, `.claude/worktrees/`)는
계속 알아봐요. 기존 워크트리는 미리보기 + 사용자 선택으로만 옮기고, 사용 중인 것은 카드가 끝난 뒤에,
아무것도 지우지 않아요. 경로 문장 4곳(AGENTS.md:30, AGENTS.local.md:34-35·125,
skills/moai/references/internal/git-safety.md:7)도 이 카드 몫인데 그 파일들은 다른 카드 예약 중이라
반환되는 대로 이어가요(아래 Plan 슬라이스 2).

## Non-goals

- 기존 워크트리의 자동 이동·삭제 — 이동은 카드 기준 2 의 마이그레이션(미리보기 + 선택)만.
- `worktrees/projects`(Go 세대, reserved)와 저장소 안 옛 위치의 쓰기 — 읽기(분류)만.
- 문장 4곳 편집의 지금 실행 — AGENTS.local.md(C-193), git-safety.md(C-200),
  tests/repo_agents_md.rs(C-169) 예약 중; 반환 뒤 슬라이스 2.
- 마이그레이션 미리보기/적용 구현 — 슬라이스 3(같은 브랜치, 별도 커밋).

## Requirements

- REQ-1: When a new worktree is planned, created or named, its folder shall be `<HOME>/.moai/worktrees/<project>/<id>` — the project component is the safe form of the repository's top-level folder name and the id keeps the existing safe spelling — through the one naming function every caller shares · failure: none — a reserved project component (`worktrees`, `projects`, case-insensitive) is refused, naming the collision, and every row is as it was.
- REQ-2: When the gate classes where a worktree path sits, the new `<project>/<name>` form shall be recognised beside every older form — the flat `<project>-<id>`, the Go `worktrees/projects/<project>/<name>`, `<primary>/.moai-cli/worktrees/<name>` and `<primary>/.claude/worktrees/<name>` — with the same five location words · failure: none — the reserved `worktrees/projects` folder itself and anything deeper or shallower than one worktree stays `other`.
- REQ-3: When a tool names a move destination or a bundle folder, it shall reuse the same naming function, so no caller spells the layout itself · failure: none — when naming fails, the tool keeps its existing fallback behaviour.

## Acceptance

- AC-1 (REQ-1): 단위 — 이름 함수가 두 컴포넌트 `<project>/<id>` 를 내고 reserved 프로젝트는 사유와 함께 거절. verify: `cargo test --locked --offline --lib home_workdir`
- AC-2 (REQ-2): 단위 — location_of 가 새 모양·평평한 옛 모양·go·project·claude 를 다르게 알아보고, `worktrees/projects` 자체와 더 깊은 경로는 `other`. verify: `cargo test --locked --offline --lib home_workdir`
- AC-3 (REQ-1, REQ-2, REQ-3): 명령 — 임시 HOME 에서 만든 워크트리가 새 모양으로 생기고(AC-2 고정), 이동·gc 도구가 옛 모양 fixture 를 여전히 같은 위치 단어로 분류. verify: `cargo test --locked --offline --test worktree_cli`; `cargo test --locked --offline --test lane_safe_to_reset`; `cargo test --locked --offline --test worktree_gc`
- AC-4 (이웃): 이웃 검사. verify: `cargo test --locked --offline --lib`; `cargo test --locked --offline --bins`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/home_workdir.rs`(이름 함수·location_in), `src/worktree.rs`(plan 기본값·문서), `src/worktree_move.rs`(목적지 이름), `src/bundle/build.rs`(호출 유지 확인), `tests/worktree_cli.rs`, `tests/lane_safe_to_reset.rs`, `.moai/specs/SPEC-20261120-worktree-project-folder/**`
- read-only: `src/worktree_gc.rs`(분류 소비), `src/statusline.rs`(라벨), `tests/worktree_gc*.rs`, `tests/worktree_move.rs`
- blocked(다른 카드 예약 — 슬라이스 2 때 widen): `AGENTS.local.md`(C-193), `skills/moai/references/internal/git-safety.md`(C-200), `tests/repo_agents_md.rs`(C-169), `skills/moai/references/source-manifest.json`
- forbidden: 실제 워크트리 이동·삭제, 실제 HOME 사용, 카드 번호를 SPEC id·가지·커밋에 사용

## Plan

1. 슬라이스 1(이번): REQ-1·REQ-2 코드 — `worktree_rel`(두 컴포넌트)로 이름 함수 교체, location_in 이 깊이 2(첫 컴포넌트 non-reserved)를 `home`으로, 호출점(plan·move·bundle) 유지. RED: worktree_cli·lane_safe_to_reset 의 경로 단정을 새 모양으로 바꿔 실패 확인 → GREEN.
2. 슬라이스 2: 문장 4곳 + 핀 시험(tests/repo_agents_md.rs)+ source-manifest — 예약 풀리면(기준 3의 문장 절반). 프로젝트별 상한 카운트 지점은 이때 확인.
3. 슬라이스 3: 마이그레이션 — optimize(또는 update)의 미리보기(대상 경로·사용 중 사실) + 선택 적용(`git worktree move` + 기록), 아무것도 지우지 않음(기준 2).

## Result

base `1c9f1a1f95`(로컬 v2026.11.0), worktree `~/.moai/worktrees/moai-adk/c184`(새 모양), heavy-slot `--max 4`. 커밋 세 개(같은 가지 spec/worktree-project-folder-20261120).

- 슬라이스 1 `eabf2dcefe` — REQ-1·REQ-2: `worktree_rel`(두 컴포넌트, reserved 프로젝트 거절), `location_in` 깊이 2 = home(다섯 위치 단어 불변), plan 기본값·이동 목적지·bundle 호출 유지. RED: E0425 `worktree_rel` 부재 → GREEN: lib 1619 · lane_safe_to_reset 22 · worktree_cli 23 · worktree_gc 14 · worktree_move 12, 전부 0 failed. 제품 구멍 1건: `git worktree move` 는 중간 폴더를 만들지 않아 apply 가 목적지의 프로젝트 폴더를 생성(이동 시험이 잡음).
- 슬라이스 3a `d0439c01bb` — REQ-3·기준 2 적용 절반: `worktree-move-plan/apply --relayout` — root 바로 아래 평평한 워크트리가 `<root>/<project>/<id>` 이동 후보(flat_to_rel: `<project>-` 접두 탈락, 나머지가 비면 전체 이름 유지), 작업 남은 것은 사유와 함께 blocked, 이미 두 컴포넌트인 것은 keep, 플래그 없는 기본 동작 불변. RED: `--relayout` 미지원 거절 → GREEN: lib 1619 · worktree_move 15(신규 3 포함) · worktree_cli 23 · worktree_gc 14 · lane_safe_to_reset 22.
- 슬라이스 3b `a2cb1b3679` — REQ-3·기준 2 미리보기 절반: `optimize.plan` 의 `worktree_moves` — 평평한 워크트리의 대상 경로(이름 못 만들면 null+사유)·process_inside·busy, 미리보기는 아무것도 옮기지 않음. RED: 키 부재 → GREEN: lib 1619 · optimize_plan_cli 20(신규 1 포함) · worktree_move 15.
- 리더 결정(노트 4057): (가) — optimize.plan 미리보기 + worktree-move-plan/apply --relayout 적용. 삭제 없음, 사용 중 제외.
- 대기: 슬라이스 2(문장 4곳 — AGENTS.local.md C-193·git-safety.md C-200·tests/repo_agents_md.rs C-169 예약 중). READY 는 슬라이스 2 착지 뒤 폭 검사와 함께.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `worktree_name`(한 컴포넌트 `<project>-<id>`)를 `worktree_rel`(두 컴포넌트)로 교체 — 호출점 3곳이 같은 함수 | 레이아웃을 아는 곳은 하나(카드 REQ-3) | 호출점마다 모양 문자열 |
| D-2 | 프로젝트 컴포넌트가 reserved(`worktrees`·`projects`)면 생성 거절 | `worktrees/projects` 는 Go 세대 전용(D8b) — 새 모양 깊이 2와 뒤섞임 방지 | 이름을 `<project>-<id>` 로 에스케이프 |
| D-3 | location 단어 집합(home/go/project/claude/other) 불변, 새 모양만 추가 | 소비자(gc·move·statusline)가 단어에 의존 | 단어 재정의(소비자 전부 수정) |
| D-4 | 문서 문장은 이 카드 몫이나 파일 예약 때문에 슬라이스 2로 미룸 | 리더 범위 결정 노트 | 다른 카드에 문장 이관 |
| D-5 | 깊이 2는 새 모양의 서명 — 판정은 순수(경로 모양만)로 유지하고, 옛 평평한 워크트리의 하위 폴더와 모양이 같음은 분류 입력이 등록된 워크트리 루트뿐임으로 문서화 | location_in 은 fs 를 읽지 않는 계약; 소비자(gc·move)는 `git worktree list` 루트만 넘김 | fs 를 읽어 부모 폴더 상태로 구분(판정이 상태 의존이 됨) |

## Risks

- statusline·gc 등이 위치 단어("home")에 의존해요 — 단어 집합을 불변으로 두고 모양만 넓혀요(D-3).
- 옛 평평한 모양을 아는 외부 스크립트·사용자 습관은 코드 밖이에요 — 문장 슬라이스(2)와 마이그레이션(3)이 그 절을 닫아요.

---
id: SPEC-20261117-sdd-foundation
title: SDD 기반 — 2026-11-17 결정 기록과 SPEC 파일 git 추적
status: approved
revision: 1
created: 2026-11-17
parent: null
approval: {by: user, at: 2026-11-17, basis: "in-session selection: 제안대로 모두 진행", digest: null}
---

## Intent

MoAI는 항상 SDD로 일해요. 저장소를 바꾸는 모든 작업은 SPEC을 쓰고, SPEC은 git에 남아 팀이 함께 봐요.
그런데 지금 이 저장소의 `.gitignore`는 `.moai/` 폴더 전체를 무시하고(`/.moai/`, `.moai/`), 생성기
`src/agents.rs`는 GitHub Issues를 쓰지 않는 프로젝트에서만 `specs/`를 git에 남겨요. 이 SPEC은 SDD 실행
순서의 1·2단계예요.

1. 2026-11-17 사용자 결정을 결정 기록(`docs/decisions-2026-11-08.md`)에 행마다 남기고, 각 행이 대체하는
   기록을 적어요.
2. SPEC 파일이 git에 남도록 해요: 이 저장소의 `.gitignore`와, 생성기가 사용자 프로젝트에 쓰는 관리 구간.

이 SPEC 자신이 첫 SPEC 파일이에요(직접 작성, `moai spec new`는 후속 SDD-2).

## Non-goals

- `.moai-cli/project.yaml`·`.moai-cli/manifests/**` 등 이미 추적 중인 파일을 `.moai/`로 옮기는 일(시험이
  요구하지 않아 이번에는 하지 않아요).
- SPEC 해석기·`moai spec` 명령·관문 트레일러 검사·워크플로 문서(SDD-1, SDD-2, SDD-5, SDD-6).
- 결정 기록 머리말의 옛 Issue 원장 문장 정리.

## Requirements

- REQ-1: The decisions log shall record each 2026-11-17 user decision of the SDD rollout as one row in the
  existing table format, and each row whose decision replaces an earlier record shall name that record.
- REQ-2: When the generator writes the managed `.gitignore` block of a project, the block shall re-include
  `<state>/specs/` for every project, whatever the profile's `github_issues` value.
- REQ-3: IF a user's own ignore rule already ignores a path under `<state>/specs/`, THEN the generator shall
  write no `!` line that undoes the rule and the plan `detail` shall name the rule. (unwanted behaviour; the
  existing user-rule guard, unchanged)
- REQ-4: The development repository's `.gitignore` shall leave every file under `.moai/specs/` not ignored and
  keep every other path under `.moai/` ignored, and keep `.moai/` folders below the root ignored.
- REQ-5: IF a file under `.moai/specs/` matches a credential pattern of the repository's `.gitignore`, THEN git
  shall still ignore it.
- REQ-6: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`, tracked by
  git in the same PR.

REQ-2 failure path: REQ-3.

## Acceptance

- AC-1 (REQ-1): 결정 기록에 2026-11-17 행이 11개 있고, 대체 대상이 있는 행마다 `**대체**`가 있어요.
  verify: `grep -c '^| xx:xxZ(11-17) |' docs/decisions-2026-11-08.md` → `11`;
  `grep '^| xx:xxZ(11-17) |' docs/decisions-2026-11-08.md | grep -c '\*\*대체\*\*'` → `8` (U-C, 강제, A–G 승인
  행은 대체 대상이 없어요)
- AC-2 (REQ-2): `github_issues`가 없음·false·true인 세 프로젝트 모두 `.moai-cli/specs/SPEC-1/spec.md`가 무시되지
  않고 구간에 `!.moai-cli/specs/`가 있어요. 폴더 이동 뒤에는 `specs/`에 "kept because" 메모가 필요 없어요.
  verify: `cargo test --locked --offline --test gitignore_block_cli` (전부 통과, 그중
  `specs_stay_in_git_for_every_project`, `what_git_tracks_in_the_earlier_folder_keeps_its_place_in_the_new_one`)
- AC-3 (REQ-3): 사용자 규칙 `.moai-cli/*`·`.moai-cli/**`가 무시하던 `.moai-cli/specs/SPEC-1/spec.md`는 적용 뒤에도
  무시돼요. verify: `cargo test --locked --offline --test gitignore_block_cli a_user_rule_keeps_its_effect_and_the_preview_names_it`
- AC-4 (REQ-4): verify: `git check-ignore -v --no-index .moai/specs/SPEC-x/spec.md` → 출력 없음, exit 1;
  `git check-ignore -v --no-index .moai/backups/x .moai/logs/run.log .moai/cards-pending/p.md .moai/issue-archive/x.json .moai/docs/a.md .moai/manifests/m.json .moai/reports/plan-audit/a.md .moai/project.yaml .moai/state/s.json sub/.moai/x a/b/.moai/specs/y.md`
  → 경로마다 한 줄, exit 0
- AC-5 (REQ-5): verify: `git check-ignore -v --no-index .moai/specs/SPEC-x/credentials.yaml` →
  `.moai/**/credentials.yaml` 규칙 한 줄
- AC-6 (REQ-6): verify: `git ls-files .moai/specs` → `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- AC-7 (all): 저장소 기본 검사가 통과해요. verify: `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test repo_agents_md`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `docs/decisions-2026-11-08.md`, `docs/rust-product-path.md` (the paragraph that describes the
  allow-list), `.gitignore`, `src/agents.rs`, `tests/gitignore_block_cli.rs`,
  `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- read-only: `src/state_root.rs`, `.moai-cli/**`, `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`,
  `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json`, `skills/moai/**`
- forbidden: the real HOME, moving tracked `.moai-cli/` files, network writes (push, PR, Issue)

## Plan

1. RED: `tests/gitignore_block_cli.rs` — `specs_stay_in_git_for_every_project` (replaces the Issue-dependent
   test) and the folder-move test without the `specs/` note; run, expect both to fail at their new assertions.
2. GREEN: `src/agents.rs` — `specs/` joins `STATE_KEPT` as its last entry (blocks written before keep their
   bytes), the `github_issues` branch and parameter of `ignore_block` go away.
3. `.gitignore`: `/.moai/` → `/.moai/*` + `!/.moai/specs/`; the any-depth `.moai/` line → `*/**/.moai/` so it no
   longer matches the root folder.
4. Decision rows (Korean) and the `docs/rust-product-path.md` sentence.
5. Checks of AC-1…AC-7; one mutation of `STATE_KEPT` after the first commit.

## Result

Base `41b6d67991e536d4ab312065140dde0fa350d08d` (origin/v2026.11.0), change commit
`1ae3cdfc6f0f55cc86db34873f91ce4c27ac4d13`. Filled by the implementing agent, 2026-11-17; SYNC and AUDIT
re-check it. Build folder: the worktree's own `target/`, `CARGO_INCREMENTAL=0`, macOS arm64.

| AC | Result | Evidence |
|---|---|---|
| AC-1 | met | `grep -c` → `11`; rows with `**대체**` → `8` |
| AC-2 | met | RED before the code change: `--test gitignore_block_cli` exit 101, 28 passed / 2 failed — `specs_stay_in_git_for_every_project` at tests/gitignore_block_cli.rs:279 (`github_issues Some(true)`), the folder-move test at :767 (the `specs/` note). GREEN: exit 0, 30 passed. Mutation M1 (`specs/` removed from `STATE_KEPT`, after the commit): exit 101, 5 failed incl. `specs_stay_in_git_for_every_project`; restored, `git status` clean |
| AC-3 | met | `a_user_rule_keeps_its_effect_and_the_preview_names_it`: exit 0, 1 passed (29 filtered) |
| AC-4 | met | `.moai/specs/SPEC-x/spec.md`: no output, exit 1. The 11 other paths: one line each (`.gitignore:359:/.moai/*` ×9, `.gitignore:372:*/**/.moai/` ×2). Probe: with the old any-depth `.moai/` line put back, the SPEC path printed `.gitignore:372:.moai/`; restored |
| AC-5 | met | `.gitignore:164:.moai/**/credentials.yaml` for `.moai/specs/SPEC-x/credentials.yaml` |
| AC-6 | met | `git ls-files .moai/specs` → `.moai/specs/SPEC-20261117-sdd-foundation/spec.md` |
| AC-7 | met | `cargo fmt --check` 0 (first run 1: one signature reflowed, fixed before the commit); clippy 0; `--bin moai` 0 (27 passed); `--lib agents::` 0 (8 passed); `--test home_isolation_scan` 0 (16); `--test repo_agents_md` 0 (3); also `project_update_plan_cli` 0 (57), `github_profile_cli` 0 (9), `github_cli` 0 (11), `agents_project_cli` 0 (5), `product_cli` 0 (75); `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py` 0 (44 tests; the first run had 6 errors from a relative `CARGO_TARGET_DIR`, rerun with an absolute one) |

Not run: the full `cargo test` (targeted tests only, machine under memory pressure); Linux and Windows.

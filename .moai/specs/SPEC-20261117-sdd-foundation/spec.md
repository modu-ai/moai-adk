---
id: SPEC-20261117-sdd-foundation
title: SDD 기반 — 2026-11-17 결정 기록과 SPEC 파일 git 추적
status: approved
revision: 1
created: 2026-11-17
parent: null
approval: {by: user, at: 2026-11-17, basis: "in-session selection: 제안대로 모두 진행 (SDD rollout steps 1-2)", digest: null}
---

## Intent

MoAI는 항상 SDD로 일해요. 저장소를 바꾸는 모든 작업은 SPEC을 쓰고, SPEC은 git에 남아 팀이 함께 봐요.
그런데 이 변경 전에는 저장소의 `.gitignore`가 `.moai/` 폴더 전체를 무시했고(`/.moai/`, `.moai/`), 생성기
`src/agents.rs`는 GitHub Issues를 쓰지 않는 프로젝트에서만 `specs/`를 git에 남겼어요. 이 SPEC은 SDD 실행
순서의 1·2단계예요.

1. 2026-11-17 사용자 결정을 결정 기록(`docs/decisions-2026-11-08.md`)에 행마다 남기고, 행마다 대체하는
   기록을 적거나 "대체 없음"이라고 적어요.
2. SPEC 파일이 git에 남게 해요: 이 저장소의 `.gitignore`와, 생성기가 사용자 프로젝트에 쓰는 관리 구간.

이 SPEC 자신이 첫 SPEC 파일이에요(손으로 작성, `moai spec new`는 후속 SDD-2).

## Non-goals

- 이미 추적 중인 `.moai-cli/project.yaml`·`.moai-cli/manifests/**`를 `.moai/`로 옮기는 일(시험이 요구하지
  않아 이번에는 하지 않아요).
- SPEC 해석기·`moai spec` 명령·관문 트레일러 검사·워크플로 문서(SDD-1, SDD-2, SDD-5, SDD-6).
- 생성기가 사용자 프로젝트 구간에 비밀값 모양 파일 규칙을 넣는 일(후속, 아래 Decisions).
- 결정 기록 머리말의 옛 Issue 원장 문장 정리.

## Requirements

요구 문장은 승인된 템플릿 예시처럼 영어 GEARS로 써요(`src/spec.rs:194` `parse_gears`가 읽는 형식). `IF …,
THEN …`은 원치 않는 동작이에요.

- REQ-1: The decisions log shall hold one row in the existing table format for each 2026-11-17 user decision of the SDD rollout.
- REQ-2: Each such row shall name the record it replaces or state that it replaces none.
- REQ-3: When the generator writes the managed `.gitignore` block of a project, the block shall re-include `<state>/specs/` whatever the profile's `github_issues` value.
- REQ-4: When a project's existing block lacks the `specs/` line, the next plan shall offer a `merge` that adds only that line.
- REQ-5: IF a user's own ignore rule already ignores a path under `<state>/specs/`, THEN the generator shall write no `!` line that undoes the rule.
- REQ-6: IF a user's own rule leaves such a path ignored, THEN the plan `detail` shall name the rule.
- REQ-7: The repository `.gitignore` shall leave each file under `.moai/specs/` that is not secret-shaped not ignored.
- REQ-8: The repository `.gitignore` shall keep every path under the root `.moai/` outside `specs/` and every `.moai/` folder below the root ignored.
- REQ-9: IF a file under `.moai/specs/` is credential- or secret-shaped (`credentials.yaml`, `*.credentials`, `.env*`, `*.key`, `id_rsa*`, `secrets.*`), THEN git shall ignore it.
- REQ-10: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`, tracked by git in the same PR.

실패 경로: REQ-3 → REQ-5, REQ-6. REQ-4 → 없음(사용자가 고친 구간이 `conflict`가 되는 기존 동작은 그대로).
REQ-7의 "secret-shaped"는 REQ-9의 목록이에요.

## Acceptance

- AC-1 (REQ-1): 2026-11-17 행이 16개예요.
  verify: `grep -c '^| xx:xxZ(11-17) |' docs/decisions-2026-11-08.md` → `16`
- AC-2 (REQ-2): 모든 행에 `**대체**` 또는 `대체 없음`이 있고, 행에 개인 카드 번호가 없어요.
  verify: `grep '^| xx:xxZ(11-17) |' docs/decisions-2026-11-08.md | grep -v -c -e '\*\*대체\*\*' -e '대체 없음'` → `0`;
  `grep '^| xx:xxZ(11-17) |' docs/decisions-2026-11-08.md | grep -c -E '(^|[^A-Za-z])C-[0-9]'` → `0`
- AC-3 (REQ-3): `github_issues` 없음·false·true 세 프로젝트 모두 `.moai-cli/specs/SPEC-1/spec.md`가 무시되지 않아요.
  verify: `cargo test --locked --offline --test gitignore_block_cli specs_stay_in_git_for_every_project`
- AC-4 (REQ-4): `github_issues: true` 프로젝트의 옛 구간(`specs/` 줄 없음)에 계획이 `merge`(1줄 추가)를 내고, 적용 뒤
  다음 계획은 `unchanged`예요.
  verify: `cargo test --locked --offline --test gitignore_block_cli an_issue_project_block_without_specs_gets_the_line_as_a_merge`
- AC-5 (REQ-5, REQ-6): 사용자 규칙 `.moai-cli/*`·`.moai-cli/**`가 무시하던 SPEC 경로는 적용 뒤에도 무시되고, 계획이
  규칙을 적어요.
  verify: `cargo test --locked --offline --test gitignore_block_cli a_user_rule_keeps_its_effect_and_the_preview_names_it`
- AC-6 (REQ-7, REQ-8, REQ-9): 전역 규칙 없이(`-c core.excludesFile=/dev/null`) 이 저장소의 `.gitignore`를 읽어요.
  verify: `cargo test --locked --offline --test gitignore_block_cli the_repository_gitignore_tracks_specs_and_ignores_the_rest`;
  `git check-ignore -v --no-index .moai/specs/SPEC-x/spec.md` → 출력 없음, exit 1
- AC-7 (REQ-10): verify: `git ls-files .moai/specs` → `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- AC-8 (REQ-1…REQ-10): `IF`로 시작하지 않는 요구는 모두 `parse_gears`를 통과해요. verify: 커밋하지 않는 임시 시험
  파일로 이 파일의 REQ 줄을 `moai_cli::spec::parse_gears`에 넣어 봐요(결과는 Result).
- AC-9 (전체): verify: `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --lib agents::`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test repo_agents_md`;
  `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test project_update_plan_cli`;
  `cargo test --locked --offline --test github_profile_cli`; `cargo test --locked --offline --test github_cli`;
  `cargo test --locked --offline --test agents_project_cli`; `cargo test --locked --offline --test product_cli`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py`

## Scope

- write: `docs/decisions-2026-11-08.md`, `docs/rust-product-path.md`(허용 목록 문단), `.gitignore`,
  `src/agents.rs`, `tests/gitignore_block_cli.rs`, `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- read-only: `src/state_root.rs`, `src/spec.rs`, `.moai-cli/**`,
  `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`,
  `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json`, `skills/moai/**`
- forbidden: 실제 HOME, 추적 중인 `.moai-cli/` 파일 이동, 네트워크 쓰기(push, PR, Issue)

## Plan

1. RED: `tests/gitignore_block_cli.rs` — 모든 프로젝트의 `specs/` 시험, 폴더 이동 시험의 `specs/` 메모, 저장소
   `.gitignore` 시험(비밀값 모양 파일 포함). 새 단언에서 실패하는지 확인.
2. GREEN: `src/agents.rs` — `STATE_KEPT` 끝에 `specs/`(전에 쓴 구간의 바이트 유지), `ignore_block`의
   `github_issues` 분기와 인자 제거.
3. `.gitignore`: `/.moai/` → `/.moai/*` + `!/.moai/specs/` + 비밀값 모양 규칙 4줄; 깊이 무관 `.moai/` 줄 →
   `*/**/.moai/`(루트 폴더와 맞지 않게).
4. 결정 기록 행(한국어)과 `docs/rust-product-path.md` 문장.
5. AC-1…AC-9 검사, 커밋 뒤 mutation, GEARS 임시 시험.

## Decisions

- D1 요구 문장은 영어 GEARS, 나머지 본문은 한국어, 제목은 영어(리더 기본값 2026-11-17, 승인된 템플릿 예시).
- D2 `approval.basis`는 사용자 선택 "제안대로 모두 진행"이에요 — SDD 실행 순서 1·2단계(이 SPEC)를 덮어요.
  `approval.digest`는 `null`: `spec.rs` `content_digest`를 쓰는 도구(SDD-1/SDD-2)가 생길 때까지 비워 둬요.
- D3 `specs/`는 `STATE_KEPT`의 마지막 항목이에요. Issues를 쓰지 않던 프로젝트의 기존 구간은 바이트가 그대로라
  새 계획이 아무것도 바꾸지 않아요(프로토콜 fixture 16·17도 그대로).
- D4 깊이 무관 `.moai/` 규칙은 `*/**/.moai/`로 바꿨어요. 원래 줄은 루트 `.moai/`와도 맞아 `!/.moai/specs/`를
  무력하게 만들어요(git은 제외된 폴더의 자식을 다시 넣지 못해요).
- D5 비밀값 모양 규칙은 이 저장소 `.gitignore`에만 넣었어요. 생성기 구간에 넣는 일은 프로토콜 fixture와 구간 줄
  수가 바뀌는 별도 변경이라 후속으로 둬요.
- D6 결정 행의 시각은 `xx:xxZ`예요: 근거 문서(SDD 계획·CI 재설계 메모)에 결정 시각이 없어요.
- D7 `Moai-Session` 트레일러는 관문 조각(SDD-6)이 정할 때까지 커밋에 넣지 않아요.

## Result

(채우는 중)

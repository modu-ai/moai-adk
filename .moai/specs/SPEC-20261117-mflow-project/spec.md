---
id: SPEC-20261117-mflow-project
title: /moai project가 git 작업 흐름을 묻고 m-flow를 권장 — 카드 C-70
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

사용자 결정(2026-11-17): "project 설정시 m-flow를 기본으로 추천하자" → 선택 "Git 작업 흐름 질문 추가 (권장)", "푸시 + PR 열기
(권장)". Claude Code의 worktree 시작점은 `worktree.baseRef: head`(Codex·OMP는 이미 HEAD). 이 SPEC은 인터뷰 질문 한 줄, 프로필 키,
흐름 설명, m-flow를 고르면 프로젝트 `.claude/settings.json`에 그 값을 미리 보기 뒤 쓰는 길을 더해요.

## Non-goals

- 실제 HOME(`~/.claude/settings.json`) 쓰기.
- AGENTS.md 관리 구간에 흐름 줄 넣기(`src/agents.rs`는 범위 밖).
- 흐름별로 가지·PR을 만들어 주는 명령.

## Requirements

- REQ-1: Where the profile has no `work_flow`, the PROJECT interview shall ask the `work_flow` row with m-flow first and recommended, then GitHub flow, trunk-based and git-flow.
- REQ-2: The profile shall store `work_flow` as one of the four flows with an optional integration branch.
- REQ-3 (REQ-2): IF `work_flow` names an unknown flow, an unknown field or a branch name outside the portable subset, THEN reading the profile shall fail without logging the value.
- REQ-4: When the user picks m-flow and `claude` is a host, the workflow shall preview `worktree.baseRef: head` for the project's `.claude/settings.json` and write it only after the user picks it. · failure: none — 막는 경우는 REQ-5예요.
- REQ-5 (REQ-4): IF the `worktree` value is anything but `{"baseRef": "head"}` or `{"baseRef": "fresh"}`, THEN the settings plan shall refuse it and write nothing.
- REQ-6: Where the request carries `--keep-owned`, the settings plan shall keep every key MoAI already owns that the request does not name.

## Acceptance

- AC-1 (REQ-1): 질문 은행에 `work_flow` 줄(버전 줄 다음, m-flow 먼저), 흐름 절, PROJECT 단계가 있어요.
  verify: `cargo test --locked --offline --test agent_defs_package work_flow_row_offers_m_flow_first_and_the_bank_states_each_flow`
- AC-2 (REQ-2, REQ-3): 키가 더해지고(additive), 나쁜 값은 거절되고, 옛 프로필이 읽혀요.
  verify: `cargo test --locked --offline --lib profile::`
- AC-3 (REQ-4…REQ-6): 미리 보기는 쓰지 않고, digest로만 쓰고, 다른 키 바이트가 그대로이고, 되돌리기가 파일을 바이트까지 되돌려요.
  verify: `cargo test --locked --offline --test claude_settings_cli the_worktree_base_ref_is_previewed_applied_and_reverted`
- AC-4 (REQ-1…REQ-6): 이웃 검사.
  verify: `cargo test --locked --offline --test claude_settings_cli`; `cargo test --locked --offline --bin moai`;
  `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py`

## Scope

- write: `src/profile.rs`, `src/claude.rs`, `src/main.rs`, `src/help.rs`, `tests/claude_settings_cli.rs`, `tests/agent_defs_package.rs`,
  `skills/moai/references/internal/interview.md`, `skills/moai/workflows/project.md`,
  `skills/moai/references/source-manifest.json`, `tests/fixtures/help/*.txt`, `.moai/specs/SPEC-20261117-mflow-project/spec.md`
- read-only: `docs/worktree-strategy-2026-11-10.md`, `src/agents.rs`
- forbidden: 실제 HOME, `project.md`의 DDD 문구 줄(SDD-5가 고침)

## Plan

1. 시험 먼저: 프로필 키, settings 미리 보기·적용·되돌리기, 질문 은행 줄.
2. `src/profile.rs`의 `work_flow`, `src/claude.rs`의 `worktree` 관리 키와 `--keep-owned`.
3. `internal/interview`의 줄과 흐름 절, `workflows/project.md` 3단계 한 줄, skill 목록.

## Risks

- `worktree`를 통째로 소유해요 → 사용자가 이미 `worktree`를 정했으면 plan이 `conflict`로 보여 주고 쓰지 않아요(기존 규칙).
- `--keep-owned` 없이 한 키만 보내면 예전처럼 다른 소유 키가 지워져요 → PROJECT 단계는 늘 `--keep-owned`로 보내요.
- Claude Code 문서의 두 값은 2026-11-10 읽기 기준이에요(재확인은 하지 않았어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 흐름 설명은 새 참조 파일이 아니라 `internal/interview`의 절 "Git work flow" | 카드("또는 기존 참조의 절"), 질문 은행 옆 | 새 `internal/work-flow` 파일로 옮기기 |
| D2 | 프로젝트 settings 쓰기는 기존 `claude-settings-plan`/`-apply`(미리 보기·digest·영수증)에 `worktree` 관리 키를 더해서 | 조정자 지시(기존 plan/apply) | 관리 키에서 빼기 |
| D3 | `--keep-owned` 플래그: 요청에 없는 소유 키를 유지(없으면 예전처럼 빠진 키는 지움) | 한 키만 바꾸는 요청이 다른 소유 키를 지우지 않게 | 플래그 빼기 |
| D4 | `worktree` 값은 `{"baseRef": "head"}`·`{"baseRef": "fresh"}`만(문서 두 값), 다른 worktree 키는 사용자 몫 | `docs/worktree-strategy-2026-11-10.md` H1 | 허용 값 넓히기 |
| D5 | 통합 가지 이름은 이식 가능한 부분집합(1–100바이트, 영숫자 `.-_/`, `..`·`//`·`.lock` 등 금지) | git refname 규칙의 안전한 부분 | 규칙 넓히기 |
| D6 | "푸시 + PR 열기 (권장)"은 m-flow 설명의 끝맺음 문장으로 | 사용자 선택 | 문장 빼기 |

## Result

아직 비어 있어요(구현 뒤 채워요).

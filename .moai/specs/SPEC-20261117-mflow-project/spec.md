---
id: SPEC-20261117-mflow-project
title: /moai project가 git 작업 흐름을 묻고 m-flow를 권장 — 카드 C-70
status: draft
revision: 2
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
- Claude Code 문서의 두 값: 감사자가 2026-11-17T09:59:58Z에 https://code.claude.com/docs/en/worktrees.md를 다시 읽었고 `baseRef`는
  여전히 `"fresh"`·`"head"`뿐이에요(감사 N2). 문서가 값을 더하면 D4를 다시 봐요.
- 사용자가 이미 `worktree` 객체(예: `sparsePaths`)를 가지면 plan이 `conflict`예요 → 객체를 바꾸지 않고, 사용자가 자기 객체에
  `"baseRef": "head"`를 더하도록 PROJECT 단계가 말해요(감사 N1).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 흐름 설명은 새 참조 파일이 아니라 `internal/interview`의 절 "Git work flow" | 카드("또는 기존 참조의 절"), 질문 은행 옆 | 새 `internal/work-flow` 파일로 옮기기 |
| D2 | 프로젝트 settings 쓰기는 기존 `claude-settings-plan`/`-apply`(미리 보기·digest·영수증)에 `worktree` 관리 키를 더해서 | 조정자 지시(기존 plan/apply) | 관리 키에서 빼기 |
| D3 | `--keep-owned` 플래그: 요청에 없는 소유 키를 유지(없으면 예전처럼 빠진 키는 지움) | 한 키만 바꾸는 요청이 다른 소유 키를 지우지 않게 | 플래그 빼기 |
| D4 | `worktree` 값은 `{"baseRef": "head"}`·`{"baseRef": "fresh"}`만(문서 두 값), 다른 worktree 키는 사용자 몫 | `docs/worktree-strategy-2026-11-10.md` H1 | 허용 값 넓히기 |
| D5 | 통합 가지 이름은 이식 가능한 부분집합(1–100바이트, 영숫자 `.-_/`, `..`·`//`·`/.` 금지, `/` 조각마다 `.lock`으로 끝나지 않음, `HEAD` 아님; 감사 B2 뒤 `git check-ref-format --branch`와 시험 이름마다 맞춰 봄) | git refname 규칙의 안전한 부분 | 규칙 넓히기 |
| D7 | m-flow 설명은 한 SPEC = 한 가지 `spec/<yyyymmdd>-<slug>` = 한 PR, 카드 번호는 가지·커밋·PR에 쓰지 않음(감사 B1) | 결정 2026-11-17(`docs/decisions-2026-11-08.md` SPEC id·담당 표시 행) | 설명 문장 되돌리기 |
| D6 | "푸시 + PR 열기 (권장)"은 m-flow 설명의 끝맺음 문장으로 | 사용자 선택 | 문장 빼기 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 기준 `25881165be`, 커밋 `9671077fd2`, 이 worktree의 `target/`,
`CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나.

RED(09:44:10Z): `--lib profile::`은 `WorkFlowSettings`·`work_flow` 필드가 없어 컴파일 실패(새 타입이라 예상한 RED), settings 시험은
`--keep-owned`가 없어 usage 오류(exit 2), 질문 은행 시험은 `work_flow` 줄 없음. 통과 09:46:55Z.

- AC-1: met — `agent_defs_package` 19 passed(새 시험과, 늘린 고정 목록 두 개: 줄 id·목적지)
- AC-2: met — `--lib profile:: claude:: status_line:: agents::` 42 passed
- AC-3: met — `claude_settings_cli` 18(미리 보기 무쓰기, digest가 `--keep-owned` 묶음, 사용자·소유 키 유지, 되돌리기 바이트 일치, 문서 밖 값 거절)
- AC-4: met — `--bin moai` 27, `product_cli` 75, `home_isolation_scan` 16, `help_layout` 10, `status_line_cli` 16, `apply_receipt` 8, `profile_apply_engine` 4, `project_update_plan_cli` 57, `project_detect_cli` 18, `go_settings_cli` 28, `e2e_host_fixture` 6 등 이웃 시험과 저장소 scan 대상 모두 exit 0; fmt 0, `rustfmt --check src/ops/*.rs` 0, clippy 0; Python 85 OK; `source_manifest.py check` ok; `package-inspect skills/moai` ok; `moai spec check --all` ok

mutation 3개 모두 KILLED: 아무 `baseRef` 받기, `--keep-owned`가 아무것도 유지하지 않기, 가지 이름의 `..` 받기.

실행 안 함: 실제 Claude Code 세션에서 `worktree.baseRef: head`가 worktree 시작점을 바꾸는지(host 확인), 전체 `cargo test`, Linux·Windows.

---
id: SPEC-20261117-hold-release-case2
title: HOLD 해제와 릴리스 태그·게시는 모든 자율 등급에서 경우 (2)
status: draft
revision: 1
created: 2026-11-17
parent: null
---

## Intent

`SKILL.md`의 질문 정렬 규칙은 `semi`에서 HOLD 해제와 릴리스를 경우 (4)에 두었어요. 경우 (4)는 되돌릴 수 있는
선택이라 기한이 지나면 기본값이 적용돼요. 그런데 경우 (2), `internal/decide` 1번, `internal/versioning`(태그는 옮기지
않음)은 되돌릴 수 없는 일은 사용자가 고르고 기다린다고 말해요. 카드 C-81(전체 조사 2026-11-17 S3)은 HOLD 해제와
릴리스 태그·게시를 모든 자율 등급, `full` 포함, 경우 (2)로 두라고 해요.

## Non-goals

- RELEASE 경로나 버전을 계산하는 명령 만들기(조사 S4, 다른 카드).
- keep-set 3종, 새 Issue, 병합 정책 변경의 등급 바꾸기.
- 지난 결정 기록 문서(`docs/decisions-2026-11-08.md`)와 조사 보고서 고쳐 쓰기.

## Requirements

- REQ-1: The `moai` skill's question-sorting rule shall put lifting a HOLD and pushing a tag or publishing a release in case (2) at every autonomy level. · failure: none — text; a wrong copy fails REQ-3's tests.
- REQ-2: Every text that restates the decision order shall say the same and keep only new Issues and merge-policy changes as the extra `semi` choices. · failure: none — text in the leader workflow, the autonomy SPEC table and the internal references `decide` and `versioning`, pinned by AC-2.
- REQ-3 (REQ-1, REQ-2): IF any of these texts again lists lifting a HOLD or a release as a case-(4) choice or only at `semi`, THEN a test shall fail.

## Acceptance

- AC-1 (REQ-1, REQ-3): `SKILL.md`의 새 문장과 바뀐 경우 (4) 목록이 고정되고, 옛 목록 문구가 다시 나오면 실패해요.
  verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test core_rules_2876`
- AC-2 (REQ-2, REQ-3): mode-leader 두 곳, `decide`, `versioning` 문장이 고정돼요.
  verify: `cargo test --locked --offline --test mode_workflow_steps`; `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test core_rules_2876`
- AC-3 (REQ-1…REQ-3): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --test mode_cards_53`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `python3 -m unittest tests.skill_references_test tests.source_manifest_test tests.repo_scans_test tests.ci_scope_test`;
  `python3 -I tools/source_manifest.py check`

## Scope

- write: `skills/moai/SKILL.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/internal/decide.md`,
  `skills/moai/references/internal/versioning.md`, `skills/moai/references/source-manifest.json`,
  `docs/design-review/autonomy-modes-spec.md`, `tests/product_cli.rs`, `tests/mode_workflow_steps.rs`,
  `tests/agent_defs_package.rs`, `tests/core_rules_2876.rs`, `.moai/specs/SPEC-20261117-hold-release-case2/spec.md`
- read-only: `docs/decisions-2026-11-08.md`
- forbidden: `skills/moai/references/internal/{sdd,code-atlas,quality,build-folder}.md`,
  `skills/moai/workflows/{run,sync,audit,plan,verify,read}.md`, `agents/*.md`, `AGENTS.md`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 고칠 곳 찾기: `lifting a HOLD`, `releases at`, `cases (2) and (3)`를 skill·docs·tests에서 grep.
2. 문장 고치기: `SKILL.md` 경우 (2)에 한 문장, 경우 (4) 목록에서 두 항목 빼기; mode-leader 두 곳; `decide` 1번; `versioning` 릴리스 문장; 자율 등급 SPEC 표.
3. 고정 시험 옮기기와 옛 문구 금지 단언; manifest write·check.
4. 시험·형식·Python 검사, mutation 3개(옛 문장 되살리기 = RED).

## Risks

- `full` 등급의 리더가 이제 HOLD 해제와 릴리스에서 사용자를 기다려요 → 의도된 동작이에요(태그는 되돌릴 수 없음).
- 다른 문서에 옛 문구가 남을 수 있어요 → grep으로 찾은 것은 지난 결정 기록과 조사 보고서뿐이에요(Non-goals).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 카드 범위(`SKILL.md`, `internal/**`, `tests/**`) 밖 `workflows/mode-leader.md`도 고침 | 리더 지시(자율 등급 목록) | mode-leader 두 곳 되돌리기 |
| D2 | 자율 등급 SPEC 표(`docs/design-review/autonomy-modes-spec.md` §6)도 고침 | mode-leader가 그 표를 SPEC §6으로 가리킴 | 표 두 줄 되돌리기 |
| D3 | `semi`의 추가 목록은 새 Issue와 병합 정책 변경만 | 카드 C-81 기준 1 | 목록에 두 항목 되살리기 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).
기준 `25881165be`. 빌드 폴더는 이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, 무거운 작업
슬롯 안에서 cargo 한 번에 하나. RED는 mutation으로 보였어요: 옛 문장 세 개를 하나씩 되살리면 고정 시험이 실패해요.

- AC-1: met — `product_cli` 75 passed, `core_rules_2876` 5 passed. mutation M1(`SKILL.md` 경우 (4) 옛 목록) KILLED
- AC-2: met — `mode_workflow_steps` 42, `agent_defs_package` 18. mutation M2(mode-leader 옛 문장) KILLED, M3(`versioning` 옛 문장) KILLED
- AC-3: met — `mode_cards_53` 9, `spec_cli` 17, `--bin moai` 27, `cargo fmt --check` 0; Python 4개 모듈 85개 OK; manifest check 0;
  `moai spec check --all` 0; `moai package-inspect skills/moai` 0

실행 안 함: 전체 `cargo test`(바뀐 곳 시험만), clippy(Rust 코드는 시험 파일 문자열만 바뀜), Linux·Windows.
크기: `SKILL.md` 6,536 → 6,666바이트, `mode-leader.md` 44,432 → 44,618바이트, `decide.md` 1,256바이트,
`versioning.md` 15,475바이트(모두 65,536 아래).

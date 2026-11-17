---
id: SPEC-20261117-hold-release-case2
title: 릴리스 태그·게시는 모든 자율 등급에서 경우 (2)
status: draft
revision: 2
created: 2026-11-17
parent: null
---

## Intent

`SKILL.md`의 질문 정렬 규칙은 `semi`에서 릴리스를 경우 (4)에 두었어요. 경우 (4)는 되돌릴 수 있는 선택이라 기한이
지나면 기본값이 적용돼요. 그런데 경우 (2), `internal/decide` 1번, `internal/versioning`(태그는 옮기지 않음)은
되돌릴 수 없는 일은 사용자가 고르고 기다린다고 말해요. 카드 C-81(전체 조사 2026-11-17 S3)에 따라 릴리스 태그·게시를
모든 자율 등급, `full` 포함, 경우 (2)로 둬요. 이는 기존 사용자 결정(`docs/decisions-2026-11-08.md` 22:0x "제약 해제
범위": release·deploy 금지 유지)과 같은 방향이에요.

revision 2(리더 결정, 감사 B1 뒤): 첫 판은 HOLD 해제도 경우 (2)로 옮겼어요. 그런데 같은 22:0x 사용자 결정은 "사용자 결정
대기 HOLD는 리더가 현재 구현값으로 확정"할 수 있다고 정해 두었고, 카드의 근거는 사용자 선택이 아니라 조사(S3)예요.
그래서 HOLD 해제 문구는 tip 그대로 두고(`semi`에서 경우 (4)), 그 부분은 나중에 사용자에게 물어요.

## Non-goals

- HOLD 해제의 등급 바꾸기: 사용자 결정이 먼저예요(D4).
- RELEASE 경로나 버전을 계산하는 명령 만들기(조사 S4, 다른 카드).
- keep-set 3종, 새 Issue, 병합 정책 변경의 등급 바꾸기.
- 지난 결정 기록 문서(`docs/decisions-2026-11-08.md`)와 조사 보고서 고쳐 쓰기.

## Requirements

- REQ-1: The `moai` skill's question-sorting rule shall put pushing a tag or publishing a release in case (2) at every autonomy level. · failure: none — text; a wrong copy fails REQ-3's tests.
- REQ-2: Every text that restates the decision order shall say the same and keep new Issues, lifting a HOLD and merge-policy changes as the extra `semi` choices. · failure: none — text in the leader workflow, the autonomy SPEC table and the internal references `decide` and `versioning`, pinned by AC-2.
- REQ-3 (REQ-1, REQ-2): IF any of these texts again lists a release as a case-(4) choice or only at `semi`, or moves lifting a HOLD into case (2), THEN a test shall fail.

## Acceptance

- AC-1 (REQ-1, REQ-3): `SKILL.md`의 새 문장과 경우 (4) 목록(새 Issue, HOLD 해제, 병합 정책 변경)이 고정되고, 옛 목록
  문구나 HOLD 해제를 경우 (2)로 옮긴 문구가 나오면 실패해요.
  verify: `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test core_rules_2876`
- AC-2 (REQ-2, REQ-3): mode-leader 두 곳, `decide`, `versioning` 문장과 자율 등급 SPEC §6의 `full`·`semi` 행이 고정돼요
  (감사 K1: `full` 행의 release 항목).
  verify: `cargo test --locked --offline --test mode_workflow_steps`; `cargo test --locked --offline --test agent_defs_package`; `cargo test --locked --offline --test core_rules_2876`; `cargo test --locked --offline --test product_cli`
- AC-3 (REQ-1…REQ-3): 이웃 검사가 통과해요.
  verify: `cargo test --locked --offline --test mode_cards_53`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --bin moai`; `cargo fmt --check`;
  `python3 -m unittest tests.skill_references_test tests.source_manifest_test tests.repo_scans_test tests.ci_scope_test`;
  `python3 -I tools/source_manifest.py check`

## Scope

- write: `skills/moai/SKILL.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/internal/decide.md`,
  `skills/moai/references/internal/versioning.md`, `skills/moai/references/source-manifest.json`,
  `docs/design-review/autonomy-modes-spec.md`, `tests/product_cli.rs`, `tests/mode_workflow_steps.rs`,
  `tests/agent_defs_package.rs`, `tests/core_rules_2876.rs`, `.moai/specs/SPEC-20261117-hold-release-case2/spec.md`
- read-only: `docs/decisions-2026-11-08.md`
- forbidden: HOLD 해제 문구 바꾸기(D4), `skills/moai/references/internal/{sdd,code-atlas,quality,build-folder}.md`,
  `skills/moai/workflows/{run,sync,audit,plan,verify,read}.md`, `agents/*.md`, `AGENTS.md`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 고칠 곳 찾기: `lifting a HOLD`, `releases at`, `cases (2) and (3)`를 skill·docs·tests에서 grep.
2. 문장 고치기: `SKILL.md` 경우 (2)에 릴리스 한 문장, 경우 (4) 목록에서 릴리스만 빼기; mode-leader 두 곳; `decide` 1번;
   `versioning` 릴리스 문장; 자율 등급 SPEC 표.
3. 고정 시험 옮기기와 옛 문구 금지 단언, §6 두 행 고정; manifest write·check.
4. 시험·형식·Python 검사, mutation 3개(옛 문장 되살리기 = RED).
5. revision 2: HOLD 해제 문구를 tip으로 되돌리고 시험·SPEC을 맞추고, tip을 병합해요.

## Risks

- `full` 등급의 리더가 이제 릴리스 태그·게시에서 사용자를 기다려요 → 의도된 동작이에요(태그는 되돌릴 수 없음).
- 다른 문서에 옛 문구가 남을 수 있어요 → grep으로 찾은 것은 지난 결정 기록과 조사 보고서뿐이에요(Non-goals).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 카드 범위(`SKILL.md`, `internal/**`, `tests/**`) 밖 `workflows/mode-leader.md`도 고침 | 리더 지시(자율 등급 목록) | mode-leader 두 곳 되돌리기 |
| D2 | 자율 등급 SPEC 표(`docs/design-review/autonomy-modes-spec.md` §6)도 고침 | mode-leader가 그 표를 SPEC §6으로 가리킴 | 표 두 줄 되돌리기 |
| D3 | `semi`의 추가 목록은 새 Issue, HOLD 해제, 병합 정책 변경 | 카드 C-81 기준 1(릴리스만 빠짐), D4 | 목록에 릴리스 되살리기 |
| D4 | HOLD 해제는 바꾸지 않고 나중에 사용자에게 물음 | 리더 결정(감사 B1): 22:0x 사용자 결정이 리더의 HOLD 확정을 허용하고, 카드 근거는 조사 S3 | 사용자 답에 따라 HOLD 해제도 경우 (2)로 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 상태는 `draft` 그대로예요(승인은 사용자 몫).

revision 1: 기준 `25881165be` 위 `df4f62c075`. 감사 FAIL(B1: HOLD 해제를 사용자 결정 기록 없이 경우 (2)로 옮김, K1: §6
`full` 행의 release 항목이 시험에 없어 mutation이 살아남음).

revision 2: `def8349a40`에서 HOLD 해제 문구를 tip으로 되돌리고 릴리스 규칙만 남겼어요. 그 뒤 `cbf1c7d959`를
병합했어요(`092c882ef6`; manifest는 병합 드라이버가 스스로 맞췄고 check 0). 병합된 head에서(무거운 작업 슬롯 안,
2026-11-17 10:09–10:12Z):

- AC-1: met — `product_cli` 75, `core_rules_2876` 6. mutation N1(`SKILL.md`가 HOLD 해제를 다시 경우 (2)로) KILLED
- AC-2: met — `mode_workflow_steps` 42, `agent_defs_package` 18. mutation K1(§6 `full` 행의 release 항목 빼기, 감사에서
  살아남았던 것) KILLED, N2(`decide`의 릴리스 규칙 빼기) KILLED
- AC-3: met — `mode_cards_53` 9, `spec_cli` 17, `--bin moai` 27, `cargo fmt --check` 0; Python 4개 모듈 88개 OK;
  manifest check 0; `moai spec check --all` 0; `moai package-inspect skills/moai` 0

실행 안 함: 전체 `cargo test`, clippy(Rust 코드는 시험 파일 문자열만 바뀜), Linux·Windows.
크기: `SKILL.md` 6,662바이트, `mode-leader.md` 병합 뒤 44,580바이트(둘 다 65,536 아래).

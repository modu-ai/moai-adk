---
id: SPEC-20261120-destructive-ops-checklists
title: "Checklists for code that deletes or moves data, for units and limits, and for external-tool versions"
status: draft
revision: 1
created: 2026-11-20
---

# Checklists for code that deletes or moves data, for units and limits, and for external-tool versions

## Intent

사용자 지시(2026-11-20, "skills, agents 개선 필요한 것 전수조사해서 소스 코드 업데이트")로 만든 조사 보고서(`~/.moai/reports/moai-adk/skills-agents-gap-2026-11-20.md`)의 G3·G4·G5 를 배포되는 지침에 넣어요. 모든 사용자에게 같은 말이고, 이 저장소만의 도구 이름은 쓰지 않아요. 기본 에이전트 다섯은 그대로예요.

- G3 지우기·옮기기 점검표: 경로의 모든 구성 요소가 링크가 아닌지, `-z` 로 파싱해 유일하게 맞는 행만, 닿지 않게 되는 커밋 보존, 지우기 전 기록 저장, 막 만든 빈 대상 보호, 남의 저장소 내용물 존중, 읽을 수 없으면 아무것도 안 함.
- G4 단위·한도: 두 쪽이 같은 단위로 세고, 쓰는 쪽 한도와 읽는 쪽 한도가 같거나 쓰는 쪽이 먼저 거절하고, 한도 시험은 N·N+1·빈 경우.
- G5 외부 도구의 새 옵션: 지원 플랫폼의 가장 오래된 버전으로 확인하고, 모르면 대체 분기와 그 분기를 도는 시험.

근거는 오늘 감사에서 나온 사고들이에요: 링크된 부모를 따라가 다른 저장소의 추적 파일을 옮김, 하위 폴더의 중첩 저장소가 통째로 옮겨짐, 줄바꿈 든 경로가 가짜 행을 만듦, 병합 안 된 커밋이 든 작업 트리를 지움, 쓰기 한도와 읽기 한도가 달라 크기 거절이 다른 오류로 둔갑, 변경된 파일 줄 수를 두 줄로 세어 잘림을 놓침, 지원 플랫폼의 옛 git 이 모르는 옵션 사용.

## Non-goals

- 점검표의 항목을 코드로 강제하는 도구는 만들지 않아요(지침과 그 지침을 못 박는 시험만).
- 기존 코드의 지우기·옮기기 경로를 이 SPEC 에서 고치지 않아요.
- 에이전트 정의의 다른 문장(구현·감사의 단계 문구)은 바꾸지 않아요: `reads:` 줄만 바뀌어요.
- 새 기본 에이전트나 새 워크플로를 만들지 않아요.

## Requirements

- REQ-1: The shipped references shall include `internal/destructive-ops`, which lists the checks for code that removes, moves or overwrites data the user owns.
- REQ-2: The `destructive-ops` reference shall name the link check on every path component, the NUL-separated reading of machine output with exactly one matching record, keeping every commit reachable, saving the way back before the removal, leaving a just-created empty target alone, leaving a nested repository's contents alone and failing closed.
- REQ-3: The `boundary-checks` reference shall state that both sides of a comparison count the same unit and that the writer's limit equals the reader's limit.
- REQ-4: The `boundary-checks` reference shall state that a change using a new option of an external tool names the oldest supported version, checks support at run time and tests the fallback branch.
- REQ-5: The `moai-audit` and `moai-impl` agent definitions shall list `internal/destructive-ops` among the references they read.
- REQ-6: The guidance test shall fail when a reference lacks one of these items or an agent definition does not read the reference.

## Acceptance

- AC-1 (REQ-1, REQ-2): `destructive-ops.md` 가 있고 첫 줄 "When to read:" 아래 일곱 가지 점검(경로 구성 요소 링크, `-z` 유일 행, 닿지 않게 되는 커밋, 기록 먼저, 막 만든 대상, 남의 저장소, 닫힌 실패)과 "가드를 일부러 깨 본다"를 말해요.
  verify: `cargo test --locked --offline --test destructive_ops_guidance destructive_ops_names_every_check`
- AC-2 (REQ-3, REQ-4): `boundary-checks.md` 의 8번 항목이 단위·한도를, 9번 항목이 외부 도구의 가장 오래된 버전·실행 중 확인·대체 분기 시험을 말해요.
  verify: `cargo test --locked --offline --test destructive_ops_guidance boundary_checks_state_units_limits_and_tool_versions`
- AC-3 (REQ-5): `agents/audit.md` 와 `agents/impl.md` 의 `reads:` 에 `internal/destructive-ops` 가 있어요.
  verify: `cargo test --locked --offline --test destructive_ops_guidance both_agents_read_destructive_ops`
- AC-4 (REQ-6): 위 문장 하나를 지우면 해당 시험이 실패해요(변이 확인).
  verify: `cargo test --locked --offline --test destructive_ops_guidance the_guidance_test_fails_when_an_item_is_missing`

## Scope

- write: `.moai/specs/SPEC-20261120-destructive-ops-checklists/spec.md`, `skills/moai/references/internal/destructive-ops.md`(새), `skills/moai/references/internal/boundary-checks.md`, `skills/moai/references/source-manifest.json`, `agents/audit.md`(`reads:` 줄), `agents/impl.md`(`reads:` 줄), `tests/agent_defs_package.rs`(읽기 목록·참조 이름 목록), `tests/destructive_ops_guidance.rs`(새)
- read-only: 조사 보고서의 G3·G4·G5, `skills/moai/references/internal/security-checklist.md`, `skills/moai/references/internal/self-check.md`
- forbidden: 에이전트 정의의 `reads:` 밖 문장, 새 기본 에이전트, 코드의 지우기 경로

## Plan

1. SPEC 을 먼저 커밋해요(이 문서).
2. RED: `tests/destructive_ops_guidance.rs` 에 시험 넷을 먼저 써서 문장이 없을 때 실패하는 것을 봐요.
3. GREEN: `destructive-ops.md` 를 쓰고 `boundary-checks.md` 에 8·9번 항목을 더하고 두 에이전트의 `reads:` 줄에 `internal/destructive-ops` 를 더해요. `tests/agent_defs_package.rs` 의 두 목록(에이전트별 읽기 목록, 참조 이름 목록)에도 이름을 더해요. 스킬 파일을 고쳤으니 `python3 -I tools/source_manifest.py write` 로 매니페스트를 갱신해요.
4. 확인: `destructive_ops_guidance`·`agent_defs_package`·`product_cli skill_package`·`update_tux_golden`·`tests.source_manifest_test`·`tests.skill_references_test`·`tests.repo_scans_test`, fmt, clippy `--all-targets`, `moai check affected`, 변이(문장·읽기 항목 지우기).

## Risks

- `agents/audit.md`·`agents/impl.md`·`tests/agent_defs_package.rs` 는 다른 변경이 쥐고 있어서 이 SPEC 의 AC-3 은 그 예약이 풀린 뒤에야 닫혀요: 풀리기 전에는 예약을 요청하는 노트를 남기고, 풀리지 않으면 AC-3 을 뒤 SPEC 으로 넘기고 이 SPEC 의 AC-3 을 철회(`withdrawn`)로 적어요.
- 새 참조 파일이 늘면 업데이트 화면의 고정 파일 목록(`tests/fixtures/update_tux/*`)과 매니페스트가 어긋날 수 있어요: 매니페스트는 도구로 다시 쓰고, 고정 파일은 시험이 실제로 읽는지 확인해요.
- 점검표가 길어지면 에이전트가 읽지 않아요: 일곱 항목과 한 줄 확인으로 짧게 두고, 상황 설명은 항목마다 한 문장으로 해요.
- G4·G5 를 `boundary-checks` 에 더하면 경계 점검의 범위가 넓어져요: 항목은 번호를 이어 8·9 로 붙이고 앞의 일곱 항목은 그대로 둬요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | G3 는 새 참조 `internal/destructive-ops` 로 둔다(`security-checklist` 안의 한 절로 두지 않는다) | 보안 점검표는 인증·입력·비밀을 다뤄 읽는 때가 달라요: 지우기·옮기기를 쓰는 변경만 읽도록 따로 둬요 | 파일 지우고 한 절로 합치기 |
| D-2 | G4·G5 는 `boundary-checks` 에 8·9번 항목으로 더한다 | 단위·한도·외부 도구는 둘이 만나는 경계의 결함이에요 | 두 항목 지우기 |
| D-3 | 읽기 표시는 `moai-audit` 과 `moai-impl` 의 `reads:` 줄만 바꾼다 | 두 에이전트가 이 변경을 쓰고 검사해요 | 줄에서 이름 지우기 |
| D-4 | 한 줄 확인 "가드를 일부러 깨 본다"를 점검표에 넣는다 | 오늘 감사의 변이가 가드의 시험 없음을 여러 번 드러냈어요 | 항목 지우기 |

## Result

구현 전(2026-11-20, 가지 `spec/20261120-destructive-ops-checklists`). 구현이 끝나면 AC 마다 `- AC-n: met — <근거>` 줄을 이 자리에 적어요.

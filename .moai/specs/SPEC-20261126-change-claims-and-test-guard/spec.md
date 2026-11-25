---
id: SPEC-20261126-change-claims-and-test-guard
title: "Audit matches every claim to the diff; implementation reports only facts and keeps others' tests"
status: draft
revision: 1
created: 2026-11-26
---

# Audit matches every claim to the diff; implementation reports only facts and keeps others' tests

## Intent

감사가 놓친 약속이 반복돼요. 커밋 메시지는 "도움말이 다섯 개를 말한다"고 썼는데 그 파일은 두 번 다 바뀌지 않았고, READY 노트와 SPEC Result 의 줄이 diff 와 어긋난 일도 여러 번이었어요. 지금 `moai-audit` 의 단계에는 주장(claim)을 diff 와 하나씩 맞대는 줄이 없고, `moai-impl` 의 출력에는 주장의 근거를 적는 곳이 없어요. 시험을 지우거나 약하게 만드는 일은 `internal/quality` 에 한 줄 있지만 두 에이전트 모두 그 문서를 읽지 않아요. 이 SPEC 은 (1) 감사가 커밋 메시지·READY 노트·SPEC Result·`verify:` 줄·자체 점검 줄을 diff 와 명령 출력에 하나씩 맞대고, 맞는 변경이 없는 주장을 지적 사항으로 쓰게 하고, (2) 구현 에이전트가 사실만 보고하되(Claims 섹션) 범위 밖 줄은 "고쳤다"가 아니라 "누구에게 넘김"으로 쓰게 하고, (3) 구현은 범위 밖 시험과 SPEC 을 지우거나 약하게 만들지 않고 감사는 지워지거나 바뀐 시험 목록을 보고 이유를 묻게 하고, (4) 수리 READY 노트가 지적마다 수리 전·후 재현 결과를 싣게 해요. 모든 사용자 프로젝트에 같은 말이 통해야 해서 moai-adk 전용 도구 이름은 쓰지 않아요. 기본 에이전트 다섯 개는 그대로예요.

## Non-goals

- 새 에이전트: 만들지 않아요. 기본 에이전트 다섯 개를 유지해요.
- 주장을 자동으로 맞대는 도구(코드): 이번에는 에이전트 글과 참조 문서만 고쳐요. 도구는 이 규칙이 쓰인 뒤에 필요하면 따로 해요.
- 시험 약화를 diff 에서 기계로 찾기: 감사가 목록을 보고 이유를 묻는 규칙까지예요.
- 코드 쪽 두 조각(워크플로 파일을 바꾼 범위의 린트 단계, 도움말 변경이 도움말 시험을 고르는지의 고정)은 `src/check_affected.rs` 를 가진 카드의 몫이에요. 이 SPEC 은 글만 고쳐요.

## Requirements

- REQ-1: When `moai-audit` checks a change, its steps shall tell it to match each claim of the change (commit messages, the hand-off or ready note, the SPEC Result, `verify:` lines and self-check lines) one by one against the diff and the command output. · failure: none — the step is text the agent reads, not a call that can fail
- REQ-2 (REQ-1): IF a claim has no matching change or output, THEN `moai-audit` shall report it as a finding of the audit contract's own classes, a `behaviour` finding when the claim says a criterion is met and `non-blocking` otherwise.
- REQ-3: When `moai-audit` answers, its output shall hold a `Claims` section with one row per claim: where it was written, the change or command that matches it, and its label. · failure: none — the section is text the agent writes
- REQ-4: When a change deletes or alters a test file, `moai-audit` shall list those files with `git diff --stat --diff-filter=DM <base>..<head>` limited to the project's test folders and ask why each was changed. · failure: none — the step is a read-only git command
- REQ-5: When `moai-impl` answers, its output shall hold a `Claims` section that names the `file:line` or the command and exit code behind each claim. · failure: none — the section is text the agent writes
- REQ-6 (REQ-5): IF a line a claim names lies outside the write scope, THEN `moai-impl` shall write it as handed to its holder and never as fixed.
- REQ-7: While `moai-impl` works inside a write scope, it shall not delete or weaken a test or a SPEC that the task does not name.
- REQ-8: Both `moai-audit` and `moai-impl` shall list `internal/change-claims` and `internal/quality` in `reads:`.
- REQ-9: The reference `internal/change-claims` shall say how to list the claims of a change, the label each gets, and how to check the test files a change deleted or altered, without naming a tool that only this repository has.
- REQ-10: When a change alters what a command prints or how a file is laid out, the agent texts shall tell `moai-impl` to run every test that reads that output and `moai-audit` to list those tests, check that the lane's receipt covered them, report a test the receipt missed as a finding and a test not run as `not_run`. · failure: none — both are steps of agent text
- REQ-11: The package shall list the reference `internal/change-claims` in the source manifest with its current sha256.
- REQ-12: When a lane answers a FAIL with a repair, the work route shall require its ready note to list each finding it answers with the probe or test that reproduced it, its result on the audited head and on the repair head, and any change it could not make because the file lies outside its scope under `Not done` with the file and its holder, never in a commit message. · failure: none — the rule is text the lane reads

## Acceptance

- AC-1 (REQ-1, REQ-2): `agents/audit.md` 의 Steps 에 다섯 종류의 주장을 diff 와 명령 출력에 하나씩 맞댄다는 줄이 있고, 맞는 변경이 없는 주장은 감사 계약의 분류(`behaviour`, `non-blocking`)로 지적한다는 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-2 (REQ-3, REQ-5): 두 에이전트의 Output 에 `Claims` 섹션이 있고, 감사의 행은 출처·맞는 변경 또는 명령·라벨이고 구현의 행은 `file:line` 또는 명령과 종료 코드예요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-3 (REQ-4, REQ-7): 감사 에이전트는 지워지거나 바뀐 시험 파일 목록을 보고 이유를 묻고, 구현 에이전트는 범위 밖 시험과 SPEC 을 지우거나 약하게 만들지 않는다는 줄이 에이전트 글에 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-4 (REQ-6): 구현 에이전트가 범위 밖 줄을 "고쳤다"가 아니라 "넘김"(`handed to`)으로 쓰라는 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-5 (REQ-8, REQ-9): 두 에이전트의 `reads:` 에 `internal/change-claims` 와 `internal/quality` 가 들어 있고, 참조 문서가 패키지에 있으며 moai-adk 전용 도구 이름이 없어요 · verify: `cargo test --locked --offline --test agent_defs_package reads_lists_match_the_design_and_render_on_every_host`
- AC-6 (REQ-10): 감사 에이전트 글과 참조 문서에 "출력을 바꾸면 그 출력을 읽는 시험을 찾아 목록으로 적고, 레인의 영수증이 덮었는지 확인하고, 못 덮은 시험은 지적이고 안 돌린 시험은 `not_run`" 줄이 있고, 구현 에이전트 글에 그 시험을 모두 돌린다는 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-7 (REQ-11): 소스 목록이 파일과 맞아요 · verify: `cargo test --locked --offline --test source_manifest_facts this_repository_manifest_is_what_the_tool_writes`
- AC-8 (REQ-12): `workflows/work.md` 에 수리 READY 노트의 지적별 수리 전·후 재현 줄과 `Not done` 규칙이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package the_repair_ready_note_names_each_finding_with_its_probe`

## Scope

- write: `.moai/specs/SPEC-20261126-change-claims-and-test-guard/`, `skills/moai/references/internal/change-claims.md`(새), `skills/moai/references/source-manifest.json`(한 줄 자리 예약으로만), `agents/audit.md`, `agents/impl.md`, `skills/moai/workflows/work.md`, `tests/agent_defs_package.rs`
- read-only: `skills/moai/references/internal/quality.md`, `skills/moai/references/internal/doc-claims.md`, `skills/moai/references/audit.md`
- forbidden: 새 에이전트, moai-adk 전용 도구 이름(일반 사용자 프로젝트에도 통하는 말만), 카드 번호를 SPEC id·가지·커밋에 쓰는 것, `src/` 코드

## Plan

1. SPEC 과 승인 요청(`agents/audit.md` 가 tier 2 바닥 목록에 있어서 승인된 SPEC 이 필요해요).
2. 참조 문서 `internal/change-claims.md`, `reads:` 표 시험과 새 규칙 줄 시험(RED).
3. `agents/audit.md`, `agents/impl.md`, `workflows/work.md` 의 줄을 더해 GREEN, 소스 목록 한 줄 자리 예약으로 기록.
4. 결과 기록(SPEC Result), 가드 변이 한 줄씩, READY.

## Risks

- **`agents/audit.md` 는 tier 2 바닥 목록에 있어요**(`src/gate.rs` 의 `TIER2_PREFIXES`, 감사 계약의 "Audit tiers" 목록, `tools/protected_checks.json`). 그래서 이 변경은 승인된 SPEC, 가드 변이 줄, 전체 감사를 받아요. 줄을 더할 뿐 약하게 하지 않지만, 풀 요청 감사는 기준 가지의 글을 읽으므로 이 변경이 합쳐지기 전 감사에는 새 줄이 적용되지 않아요.
- 에이전트 글이 길어져서 패키지 파일 한도(`profile::LIMIT`)를 넘을 수 있어요. 규칙은 새 참조 문서에 두고 에이전트 글에는 한 줄씩만 더해요.
- 렌더된 에이전트가 "다른 internal 참조는 읽지 않는다"고 말하므로 `reads:` 에 없는 문서를 가리키면 안 돼요. `reads:` 표를 고정한 시험(`reads_lists_match_the_design_and_render_on_every_host`)을 함께 고쳐요.
- 감사 단계표는 가벼운 감사가 아무것도 돌리지 않고 tier 2 감사가 시험 모듈을 둘까지만 돌리게 해요. 그래서 "출력을 읽는 시험을 모두 돌린다"는 구현 몫이고, 감사 몫은 목록을 적고 영수증이 덮었는지 확인하는 데까지예요(D-5).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙 본문은 새 참조 문서 `internal/change-claims.md` 에 두고 에이전트 글에는 단계 한 줄과 출력 섹션만 더해요 | 에이전트 글을 짧게 두는 기존 방식과 패키지 파일 한도 | 줄을 에이전트 글로 옮기기 |
| D-2 | 주장 라벨은 `matched`, `no-change`, `not-run`, `handed-over` 네 개예요 | 문서 주장 검사(`doc-claims`)의 라벨 방식을 따르되 변경 검사에 맞게 줄였어요 | 라벨 추가 |
| D-3 | 시험 변경 확인은 diff 의 시험 폴더에서 지워지거나(D) 바뀐(M) 파일을 `git diff --stat --diff-filter=DM <base>..<head>` 로 모으는 한 줄이에요 | 일반 사용자 프로젝트에도 통하는 읽기 전용 git 명령 하나 | 폴더 목록을 설정으로 |
| D-4 | 거짓 주장의 무게는 감사 계약의 네 종류(`behaviour`, `data-loss`, `security`, `gate`)와 `non-blocking` 에 맞춰요: 기준을 충족했다는 거짓 주장은 `behaviour`, 그 밖은 `non-blocking` | `references/audit.md` 의 FAIL 은 네 가지뿐이라는 규칙 | 분류 줄 고치기 |
| D-5 | 출력을 바꾸는 변경의 시험 전수 실행은 구현 몫이고, 감사는 목록과 영수증 확인까지예요 | 감사 단계표(가벼운 감사는 실행 없음, tier 2 는 모듈 둘까지) | 감사 몫을 늘리려면 단계표를 먼저 바꾸기 |

## Result

(아직 구현 전이에요.)

---
id: SPEC-20261120-change-claims-and-test-guard
title: "Audit matches every claim to the diff; implementation reports only facts and keeps others' tests"
status: draft
revision: 1
created: 2026-11-20
---

# Audit matches every claim to the diff; implementation reports only facts and keeps others' tests

## Intent

감사가 놓친 약속이 반복돼요. 커밋 메시지는 "도움말이 다섯 개를 말한다"고 썼는데 그 파일은 두 번 다 바뀌지 않았고, READY 노트와 SPEC Result 의 줄이 diff 와 어긋난 일도 여러 번이었어요. 지금 `moai-audit` 의 단계에는 주장(claim)을 diff 와 하나씩 맞대는 줄이 없고, `moai-impl` 의 출력에는 주장 근거를 적는 곳이 없어요. 또 시험을 지우거나 약하게 만드는 일은 `internal/quality` 에 한 줄 있지만 두 에이전트 모두 그 문서를 읽지 않아요. 이 SPEC 은 (1) 감사가 커밋 메시지·READY 노트·SPEC Result·`verify:` 줄·자체 점검 줄을 diff 와 명령 출력에 하나씩 맞대고 맞는 변경이 없는 주장을 지적 사항으로 쓰게 하고, (2) 구현 에이전트가 사실만 보고하되(Claims 섹션) 범위 밖 줄은 "고쳤다"가 아니라 "누구에게 넘김"으로 쓰게 하고, (3) 구현은 범위 밖 시험과 SPEC 을 지우거나 약하게 만들지 않고 감사는 지워지거나 바뀐 시험 목록을 보고 이유를 묻게 해요. 모든 사용자 프로젝트에 같은 말이 통해야 해서 moai-adk 전용 도구 이름은 쓰지 않아요. 기본 에이전트 다섯 개는 그대로예요.

## Non-goals

- 새 에이전트: 만들지 않아요. 기본 에이전트 다섯 개를 유지해요.
- 주장을 자동으로 맞대는 도구(코드): 이번에는 에이전트 글과 참조 문서만 고쳐요. 도구는 이 규칙이 쓰인 뒤에 필요하면 따로 해요.
- 시험 약화를 diff 에서 기계로 찾기: 별도 조각(테스트 약화 감지)이 해요. 여기서는 감사가 목록을 보고 이유를 묻는 규칙까지예요.

## Requirements

- REQ-1: When `moai-audit` checks a change, its steps shall tell it to match each claim of the change one by one against the diff and the command output. · failure: none
- REQ-2 (REQ-1): IF a claim has no matching change or output, THEN `moai-audit` shall report it as a finding that blocks when the claim says a requirement is met.
- REQ-3: When `moai-audit` answers, its output shall hold a `Claims` section with one row per claim: the claim, where it was written, the change or command that matches it, and the result. · failure: none
- REQ-4: When a change deletes or alters a test file, `moai-audit` shall list those files from the test folders of the diff and ask the author why each was changed. · failure: none
- REQ-5: When `moai-impl` answers, its output shall hold a `Claims` section that names the evidence of each claim. · failure: none
- REQ-6 (REQ-5): IF a line a claim names lies outside the write scope, THEN `moai-impl` shall write it as handed to its holder, never as fixed.
- REQ-7: The `moai-impl` rules shall forbid deleting or weakening a test or a SPEC outside the task's scope. · failure: none
- REQ-8: Both `moai-audit` and `moai-impl` shall list `internal/change-claims` and `internal/quality` in `reads:`. · failure: none
- REQ-9: The reference `internal/change-claims` shall say how to list the claims of a change, the label each gets, and how to check the test files a change deleted or altered, without naming a tool that only this repository has. · failure: none
- REQ-10: The package shall ship the reference, listed in the source manifest. · failure: none
- REQ-11: When a change alters what a command prints or how a file is laid out, `moai-audit` and `moai-impl` shall find every test that reads that output and run all of them, and report a test not run as `not_run`. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): `agents/audit.md` 의 Steps 에 다섯 종류의 주장을 diff 와 명령 출력에 하나씩 맞댄다는 줄과, 맞는 변경이 없으면 지적 사항이고 요구사항을 충족했다는 주장이면 막는 사항이라는 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-2 (REQ-3, REQ-5): 두 에이전트의 Output 에 `Claims` 섹션이 있고, 감사의 행은 주장·출처·맞는 변경 또는 명령·결과이고 구현의 행은 `file:line` 또는 명령과 종료 코드예요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-3 (REQ-4, REQ-7): 감사 에이전트는 지워지거나 바뀐 시험 파일 목록을 보고 이유를 묻고, 구현 에이전트는 범위 밖 시험과 SPEC 을 지우거나 약하게 만들지 않는다는 줄이 에이전트 글에 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-4 (REQ-6): 구현 에이전트가 범위 밖 줄을 "고쳤다"가 아니라 "넘김"(handed to)으로 쓰라는 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_agents_and_reference`
- AC-5 (REQ-8, REQ-9, REQ-10): 두 에이전트의 `reads:` 에 `internal/change-claims` 와 `internal/quality` 가 들어 있고 참조 문서가 패키지에 있으며 moai-adk 전용 도구 이름이 없고 소스 목록에 올라 있어요 · verify: `cargo test --locked --offline --test agent_defs_package reads_lists_match_the_design_and_render_on_every_host`
- AC-6 (REQ-10): 소스 목록이 파일과 맞아요 · verify: `python3 -I tools/source_manifest.py check`
- AC-7 (REQ-11): 감사 에이전트 글과 참조 문서에 "출력을 바꾸면 그 출력을 읽는 시험을 모두 찾아 모두 돌린다, 변경이 이름 붙인 시험만이 아니다, 안 돌린 시험은 not_run" 줄이 있어요 · verify: `cargo test --locked --offline --test agent_defs_package change_claims_rules_are_in_the_shipped_audit_agent_and_reference`

## Scope

- write: 이 SPEC, `skills/moai/references/internal/change-claims.md`(새), `tests/agent_defs_package.rs`, `agents/audit.md`, `agents/impl.md`, `skills/moai/references/source-manifest.json`
- 범위 보강(리더 노트 2026-11-20): "출력 모양을 바꾸면 그 출력을 읽는 시험을 모두 돌린다"를 구현자뿐 아니라 감사자 점검 목록에도 넣어요(REQ-11, 오늘 도움말 골든 실패를 감사가 놓친 일)
- 코드 쪽 두 조각(워크플로 린트 단계, 도움말 변경이 도움말 시험을 고르는지)은 `src/check_affected.rs` 를 쥔 카드가 풀어야 해서 따로 묶었어요: `SPEC-20261120-affected-check-workflow-lint`(`parent` 로 이 SPEC 을 가리켜요). 요구사항이 12개를 넘지 않게 하려는 분리예요
- 다른 카드가 쥔 파일(2026-11-20 예약 읽음): `agents/audit.md` 와 `skills/moai/references/source-manifest.json` 은 다른 카드가, `agents/impl.md` 는 또 다른 카드가 쥐고 있어요. 그 카드가 풀거나 리더가 순서를 정한 뒤에 써요. `internal/quality.md` 는 쥔 카드가 있어서 고치지 않고 읽기 목록에만 더해요
- read-only: `skills/moai/references/internal/quality.md`, `skills/moai/references/internal/doc-claims.md`, `skills/moai/workflows/audit.md`
- forbidden: 새 에이전트, moai-adk 전용 도구 이름(일반 사용자 프로젝트에도 통하는 말만), 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): SPEC 과 조사 보고, 막힌 파일의 순서 요청.
2. 조각 2: 참조 문서 `internal/change-claims.md` 와 RED 시험(AC-1~5), `agents/audit.md`·`agents/impl.md`·소스 목록을 쥔 카드가 풀린 뒤 에이전트 글을 고쳐 GREEN.
3. 조각 3: 결과 기록(SPEC Result), 소스 목록 확인, READY.

## Risks

- **`agents/audit.md` 는 보호 경로예요**(`src/gate.rs` 의 TIER2 목록과 `tools/protected_checks.json`). 그래서 이 변경은 tier 2 감사를 받아요. 줄을 더할 뿐 약하게 하지 않지만, 풀 요청 감사는 기준 가지의 글을 읽으므로(`git show <base>:agents/audit.md`) 이 변경이 합쳐지기 전 감사에는 새 줄이 적용되지 않아요.
- 에이전트 글이 길어져서 패키지 파일 한도(`profile::LIMIT`)를 넘을 수 있어요. 규칙은 새 참조 문서에 두고 에이전트 글에는 한 줄씩만 더해요.
- 렌더된 에이전트가 "다른 internal 참조는 읽지 않는다"고 말하므로 `reads:` 에 없는 문서를 가리키면 안 돼요. `reads:` 표를 고정한 시험(`reads_lists_match_the_design_and_render_on_every_host`)을 함께 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙 본문은 새 참조 문서 `internal/change-claims.md` 에 두고 에이전트 글에는 단계 한 줄과 출력 섹션만 더해요 | 에이전트 글을 짧게 두는 기존 방식(자체 점검 목록도 참조 문서)과 패키지 파일 한도 | 줄을 에이전트 글로 옮기기 |
| D-2 | 주장 라벨은 `matched`(맞는 변경 또는 출력 있음), `no-change`(없음), `not-run`(명령을 안 돌림), `handed-over`(범위 밖이라 넘김) 네 개예요 | 문서 주장 검사(`doc-claims`)의 라벨 방식을 따르되 변경 검사에 맞게 줄였어요 | 라벨 추가 |
| D-3 | 시험 변경 확인은 diff 의 시험 폴더에서 지워지거나(D) 바뀐(M) 파일을 `git diff --stat --diff-filter=DM <base>..<head>` 로 모으는 한 줄이에요(폴더 이름은 프로젝트의 시험 폴더) | 일반 사용자 프로젝트에도 통하는 읽기 전용 git 명령 하나 | 폴더 목록을 설정으로 |

## Result

(아직 구현 전이에요.)

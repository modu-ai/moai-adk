---
id: SPEC-20261125-audit-text-facts
title: "Make the two audit texts match the real commands and read as one ordered procedure"
status: draft
revision: 1
created: 2026-11-25
---

# Make the two audit texts match the real commands and read as one ordered procedure

## Interview

- Q1: 스킬 지침을 어떻게 고칠까요? → A: 사용자 지시 그대로(스킬 전수 조사, 2026-11-25): "moai 스킬 지침 전체 전수 조사해서 검색 말고 전체 개별로 읽고 개선점을 찾아서 모두 업데이트 개선 보고 후 진행". 이 SPEC 은 그 조각 중 감사 글 둘(`skills/moai/workflows/audit.md`, `skills/moai/references/audit.md`)이에요.
- Q2: 감사 관문(FAIL 네 가지, 작성자는 감사하지 않음, 판정은 감사자가 기록)을 바꾸나요? → A: 아니요. 관문은 약해지지 않아요. 글이 실제 명령과 같아지고 순서가 하나가 될 뿐이에요.
- Q3: tier 0(문서만 바뀐 변경)도 판정이 필요한가요? → A: 네, 실제 명령이 모든 모드에서 독립 PASS 를 요구해요(읽은 날 2026-11-25). 글을 명령에 맞추는 선택은 결정 D-1 이에요.

## Intent

두 감사 글에는 실제 명령과 다른 문장이 있고, 절차가 한 줄기가 아니에요. 읽은 사실(2026-11-25, `moai card verdict --help`, `moai card land --help`, `src/card/land/check.rs`, `moai bundle check --help`, `moai worktree-gc-apply --help`):

- `workflows/audit.md` 5행은 "판정 저장은 구현되지 않았다"고, 28행은 "착지는 판정을 읽지 않는다"고 말해요. 실제로는 `moai card verdict` 가 판정을 원장 이벤트로 저장하고, `moai card land` 는 그 이벤트의 독립 PASS 가 없으면 모든 모드에서 `no_pass` 로 거절해요.
- `references/audit.md` 36행은 docs-only(tier 0)를 "lander 가 읽고 착지한다"고 해요. 실제로는 tier 0 도 독립 PASS 가 필요하고 영수증만 면제돼요(`--receipt-head`·`--receipt-step` 은 PASS 가 tier 0 보다 위일 때).
- `workflows/audit.md` 의 감사 worktree 정리는 `moai worktree-gc-apply --worktree <path>` 라고 하지만, 그 명령의 옵션은 `--integration-ref`, `--plan-digest`, `--permission-reference` 예요(`worktree-gc-plan` 이 먼저 와요).
- 판정의 base: 레인 head 는 레인 READY 노트의 base 로, 머지 head 는 착지할 tip 으로 판정해야 해요. 범위에 감사 세션의 커밋이 들어가면 `self_verdict` 로 거절돼요(`src/card/land/check.rs` 규칙 4).
- 절차가 번호 1-9, "Review steps" 1-4, 그 뒤 "Reserve before starting" 순서라서 예약이 맨 끝에 있어요. tier 줄·FAIL 네 가지·판정 줄이 한 글 안에서 여러 번 나와요. Issue 번호 여섯, moai-adk 내부 Python 도구, Kiro·Cursor 줄이 남아 있어요.

이 SPEC 이 하는 일: 감사 절차를 예약으로 시작하는 번호 목록 하나로 만들고, tier 줄·FAIL 네 가지·판정 줄 모양을 `references/audit.md` 한 곳에 두고, 명령에 관한 문장을 실제 `--help` 와 같게 고쳐요.

### 문장 대조표 (무엇이 옮겨가고 무엇이 그대로인가)

| 지금 글 | 고친 뒤 |
| --- | --- |
| workflows `## Input` (5행, "판정 저장은 구현되지 않았다", `audit-reduce`) | 입력은 카드·SPEC digest·정확한 head 와 base·레인 READY 노트의 영수증으로 고쳐요. `audit-reduce` 와 틀린 문장은 지워요(REQ-3) |
| workflows Procedure 1-5 (패키지, 교차 확인, 지적 기록, 결정적 admission, 재검증) | 새 번호 4-7 단계(패키지, 찾기, 확인, 순위와 게시)로 합쳐요. 명령과 규칙은 그대로예요 |
| workflows Procedure 6 (검사 실행·CI 인용) | 새 8단계 한 줄이에요: tier 가 정한 검사만 돌려요. CI 인용은 RUN 글의 "Check-run rules" 를 가리켜요 |
| workflows Procedure 7 (`gate_unavailable`, Issue 번호) | "Release pull request" 절의 한 문단으로 옮겨요. Issue 번호는 지워요 |
| workflows Procedure 8 (판정 노트, 첫 줄 모양, 예시) | 새 9단계는 `moai card verdict` 호출과 노트 쓰기예요. 첫 줄 모양과 예시 줄은 `references/audit.md` 로 옮겨요(REQ-2) |
| workflows Procedure 9 (tier 문단, "착지는 판정을 읽지 않는다") | tier 는 `references/audit.md` 에만 둬요. 틀린 문장은 `no_pass` 사실로 바꿔요(REQ-4) |
| workflows 다리 문장 ("The review package is described in…") | 목록이 하나가 되어 지워요 |
| workflows `## Review steps` 1-4 | 새 번호 4-7 단계예요. `review-package`·`review-check`·`Follow-up order:` 줄은 그대로예요 |
| workflows "Reserve before starting" | 새 1단계로 올려요. `AUDIT-CLAIM` 노트를 읽는 명령은 없고 노트 약속이라는 말을 붙여요 |
| workflows "Record interim results", "Judge behaviour…", "Make every mutation real", "Read revisions safely" | 규칙은 그대로 두고 Issue 번호를 지워요. `audit-delta` 설명은 `--help` 가 말하는 만큼으로 줄여요 |
| workflows `## Re-audit PR body check` (필드별 설명, #2388 표) | "Release pull request" 절로 옮기고 `moai pr-self-audit-check --reaudit --file BODY.md` 한 줄과 `--help` 를 가리켜요. #2388 표는 지워요(REQ-7) |
| workflows `## Base agents` (tier 네 줄, Kiro·Cursor 줄, Issue #3054) | 에이전트 블록과 `RESULT` 줄 규칙은 그대로예요. tier 네 줄은 "`references/audit.md` 의 Audit tiers 를 말 그대로 브리프에 넣는다"로 바꿔요. Kiro·Cursor 줄과 Issue 번호는 지워요. worktree 정리는 리더의 `worktree-gc-plan`·`worktree-gc-apply` 로 바꿔요(REQ-8) |
| workflows `## Web pages` | 그대로예요 |
| workflows `## Resume`, `## Exceptions and Permissions`, `## Completion Check` | Resume 은 2단계(READY 노트의 head 와 다르면 blocked)에 들어가요. 다른 둘은 그대로이고, 구성원 댓글 규칙은 Issue 번호 없이 "Release pull request" 절로 가요 |
| references `What the auditor receives/returns`, `What blocks`, `Rules for every audit`, `Side rules`, 스크래치 규칙 | 그대로예요(FAIL 네 가지, 작성자는 감사하지 않음, 파일과 줄 규칙 포함) |
| references `Who records the verdict` | 그대로 두고 `card verdict` 의 사실(영수증은 tier 0 위, base 는 head 의 조상, 판정 노트 첫 줄 모양과 예시)과 base 고르는 규칙을 더해요 |
| references `Audit tiers` 네 줄과 floor 목록 | floor 목록은 코드가 읽는 목록이라 그대로예요. docs-only 줄의 "lander 가 읽고 착지한다"만 고쳐요(D-1) |

## Non-goals

- 관문을 느슨하게 하거나 바꾸지 않아요: FAIL 네 가지, 작성자는 감사하지 않음, 판정은 감사자가 기록, floor 목록.
- `agents/audit.md`, `skills/moai/workflows/run.md`, 그 밖의 스킬 글과 Rust 코드는 건드리지 않아요. 같은 tier 줄을 가진 다른 글의 정리는 그 글의 SPEC 몫이에요.
- `## Web pages` 절은 그대로예요.
- `moai card verdict`, `moai card land` 와 그 밖의 명령의 동작은 바꾸지 않아요.

## Requirements

<!-- This SPEC edits two audit texts and keeps every blocking rule they state; no check that blocks is added or loosened. -->
- REQ-1: The `## Procedure` of `workflows/audit.md` shall be one numbered list in work order: reserve the head, read the lane's ready note, pick the tier, build the package, find, verify, rank and post, run the checks the tier names, record the verdict, leave the audit worktree. · failure: none — a reader who follows the numbers meets each step once and finds no second list.
- REQ-2: `references/audit.md` shall hold the four tier lines, the four FAIL kinds and the form of the verdict line as their only statement in the two audit texts. · failure: none — `workflows/audit.md` points to them by section name.
- REQ-3: When a sentence of the two audit texts says what `moai card verdict` or `moai card land` does, the sentence shall equal what `moai card verdict --help` and `moai card land --help` say. · failure: none — a mismatch is a finding of the audit of this change.
- REQ-4: The two audit texts shall say that `moai card land` refuses a head with no independent PASS from `moai card verdict` in every mode and that a PASS above tier 0 cites a receipt of the head. · failure: none — the sentences state what the command does today.
- REQ-5: `references/audit.md` shall name the base of a verdict as the base of the lane's ready note for a lane head and as the tip the head will land on for a merge head. · failure: none — `card verdict` still refuses a base that is no ancestor of the head.
- REQ-6: The two audit texts shall carry no Issue number, no moai-adk internal tool and no retired host. · failure: none — a leftover is a non-blocking finding.
- REQ-7: The release pull request part of `workflows/audit.md` shall keep `moai pr-self-audit-check --reaudit --file BODY.md`, the rule for citing a CI run, the line `ci: gate_unavailable(<cause> since <since>)` and the member-comment rule, each in one short paragraph. · failure: none — every command it names answers `--help`.
- REQ-8: `workflows/audit.md` shall leave the removal of an audit worktree to the leader's `moai worktree-gc-plan` and `moai worktree-gc-apply`. · failure: none — the auditor deletes no worktree.
- REQ-9: The two audit texts shall keep the four FAIL kinds, the rule that the author never audits and the rule that the auditing session records its own verdict with the same meaning. · failure: none — a missing rule is a FAIL of the weakened-gate kind.
- REQ-10: A test that pins a sentence these texts lose shall pin the new sentence or be deleted with its reason in the commit message. · failure: none — an unexplained deletion is a finding.

## Acceptance

- AC-1 (REQ-1): `## Procedure` holds one numbered list whose items name, in this order, `AUDIT-CLAIM`, the ready note, the tier, `moai review-package`, `moai review-check`, `moai card verdict` and the leader's worktree removal, and the file has no second numbered procedure list · verify: `cargo test --offline --locked --test audit_text_facts ac1_`
- AC-2 (REQ-2): each of the four tier lines, the FAIL four-kinds sentence and the verdict-line form appears once across the two files, in `references/audit.md`, and `workflows/audit.md` names its section · verify: `cargo test --offline --locked --test audit_text_facts ac2_`
- AC-3 (REQ-3, REQ-4): the texts contain none of "not implemented", "reads no verdict", "the lander reads the change and lands it", state the `no_pass` refusal in every mode, and every `moai` command either text names answers `--help` with exit 0 · verify: `cargo test --offline --locked --test audit_text_facts ac3_`
- AC-4 (REQ-5): `references/audit.md` states the base rule for a lane head and for a merge head · verify: `cargo test --offline --locked --test audit_text_facts ac4_`
- AC-5 (REQ-6): neither text contains `Issue #`, a `#` followed by four digits, `tools/`, `Kiro`, `Cursor`, `Codex` or `OMP` · verify: `cargo test --offline --locked --test audit_text_facts ac5_`
- AC-6 (REQ-7, REQ-8): the release pull request part holds the four named items and the text names `moai worktree-gc-plan`, `--plan-digest` and no `worktree-gc-apply --worktree` · verify: `cargo test --offline --locked --test audit_text_facts ac6_`
- AC-7 (REQ-9): the four FAIL kinds, "never audits" and "records the verdict itself" stay in `references/audit.md`; removing any of them makes the test fail (guard mutant) · verify: `cargo test --offline --locked --test audit_text_facts ac7_`
- AC-8 (REQ-10): every pinned sentence that changed is pinned again or its test is gone with a reason · verify: `moai check affected --card <base>..<head>`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천(A)이 맨 앞이고 사용자가 고르지 않으면 기본으로 A 를 써요.

- D-1 (tier 0 문서만 바뀐 변경): (A, 기본·권장) 실제 명령대로 tier 0 도 독립 PASS 를 기록해요. docs-only 줄은 "감사자가 영수증 없이 PASS 를 기록한다. 아무것도 빌드하지 않는다"로 고쳐요. (B) 글은 그대로 두고 명령을 바꿔요 — 이 SPEC 의 범위 밖이에요.
- D-2 (판정 base): (A, 기본·권장) 레인 head 는 레인 base, 머지 head 는 착지할 tip 을 base 로 써요. (B) 항상 카드 base 를 써요 — 머지 head 에서 `self_verdict` 거절을 겪어요.
- D-3 (감사 worktree 정리): (A, 기본·권장) 감사자는 지우지 않고, 리더가 `worktree-gc-plan` 을 읽은 뒤 `worktree-gc-apply` 로 지워요(CLAUDE.md 와 같아요). (B) 감사자가 지워요 — 명령 옵션이 `--worktree` 를 받지 않아요.
- D-4 (`AUDIT-CLAIM` 예약 노트): (A, 기본·권장) 절차 1단계에 두고 "읽는 명령은 없고 노트 약속"이라 적어요. (B) 지워요 — 같은 head 를 두 세션이 감사하는 일이 생겨요.
- D-5 (릴리스 PR 검사): (A, 기본·권장) 짧은 절 하나로 줄여 명령 이름과 규칙만 남겨요(시험 하나가 `pr-self-audit-check --reaudit --file BODY.md` 를 고정해요). (B) 절 전체를 지워요.

## Scope

- write: `.moai/specs/SPEC-20261125-audit-text-facts/**`, `skills/moai/workflows/audit.md`, `skills/moai/references/audit.md`, `tests/audit_text_facts.rs`
- write (고정 문장이 바뀔 때만): `tests/next_flow_skill_scan.rs`, `tests/review_check_cli.rs`, `tests/self_audit_cli.rs`, `tests/gate_collect.rs`, `tests/skill_structure.rs`, `tests/audit_scratch_rule.rs`, `tests/agent_defs_package.rs`, `tests/core_rules_2876.rs`, `tests/sdd_guidance.rs`
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`, `tests/fixtures/update_tux/*.json`, `tools/repo_scans.json`
- read-only: `src/card/land/check.rs`, `src/gate.rs`, `agents/audit.md`, `skills/moai/workflows/run.md`
- forbidden: `agents/audit.md` 와 `workflows/run.md` 의 수정, 관문을 느슨하게 하는 변경, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. SPEC 승인을 기다려요(tier 2: 두 글 모두 floor 목록의 경로예요). 승인 전에는 코드·글을 쓰지 않아요.
2. 두 글이 다른 카드의 범위에 들어 있는 동안은 `card check` 가 쓰기를 거절해요. 풀리면 `card scope` 로 넓혀요.
3. RED: `tests/audit_text_facts.rs`(AC-1..AC-7)가 지금 글에서 실패해요.
4. 두 글을 고쳐요(대조표대로). GREEN, 그다음 고정 시험을 고치거나 이유를 적어 지워요.
5. 가드 뮤턴트(FAIL 네 가지, "never audits", "records the verdict itself" 를 글에서 지우면 AC-7 이 죽어요), manifest 해시, `cargo fmt --all`, `moai check affected --card`, READY.

## Risks

- R-1: 같은 tier 줄을 가진 다른 글(`agents/audit.md`, `workflows/run.md`)은 이 SPEC 이 고치지 않아서 docs-only 줄 문장이 잠시 달라져요. READY 에 그 차이를 적고, 그 글의 SPEC 이 맞춰요.
- R-2: 고정 시험이 많아요(위 목록). 문장 하나가 빠지면 그 시험이 실패하므로 하나씩 새 문장에 다시 고정해요.
- R-3: 글을 줄이다가 관문 규칙이 빠질 수 있어요. AC-7 의 가드 뮤턴트가 FAIL 네 가지·작성자 규칙·판정 기록 규칙이 빠지면 실패하게 해요.
- R-4: 스킬 글이 바뀌면 `source-manifest.json` 의 sha256 과 update 픽스처가 같이 바뀌어요. line-lease 로 쓰고 READY 에 적어요.

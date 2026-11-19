---
id: SPEC-20261119-leader-auditor-verdict
title: "The audit gate and the merge path read a leader-assigned auditor's verdict"
status: draft
revision: 1
created: 2026-11-19
---

# The audit gate and the merge path read a leader-assigned auditor's verdict

## Intent

감사 중앙화(사용자 결정 2026-11-16, 리더가 `moai-audit` 에이전트를 배정) 뒤에도 `src/gate.rs` 의 `auditor` 파서는 `agent-<n>`·`pr-<n>` 레인 이름만 읽어요. 리더가 배정한 감사 에이전트의 판정 줄은 감사자 이름을 레인 이름으로 쓰지 않으니 tier 1/2 의 "비작성자 PASS" 로 인식되지 않고, 병합 관문이 영원히 기다려요. 이 SPEC 은 판정 본문의 고정 줄 `auditor: leader-audit <token>` 을 감사 관문이 감사자로 읽게 하고, 판정이 관문을 통과하면 CI 작업이 `moai/audit` 커밋 상태를 직접 게시해 사람이 손으로 상태를 게시할 때까지 PR 이 BLOCKED 로 대기하지 않게 해요(카드 C-95, 워크플로우 W2).

## Non-goals

- `src/gate.rs` 의 파서 변경(병합 관문·감사 판정 줄 생성기가 공유 — 이 카드 범위 밖; 감사 관문만 로컬 인식을 더해요).
- ruleset 의 상태 요구를 GitHub Actions 앱으로 고정하는 규칙집합 변경(사용자의 클릭 — Risks 에 우회 경로와 함께 기록).
- PASS 가 CI 실행을 인용하는지의 검사(원래 이후 조각, 이 카드도 아님).

## Requirements

- REQ-1: When the audit gate reads a verdict whose body carries, below its first line, a line `auditor: leader-audit <token>`, it shall read the auditor as `leader-audit <token>` when no `agent-`/`pr-` lane name is found.
- REQ-2 (REQ-1): IF the token is empty or holds a character outside `A-Z a-z 0-9 . _ -`, THEN the gate shall read no auditor from that line.
- REQ-3: When the gate's decision is `pass` or `fail`, the result shall name the id of the comment that decided it. · failure: none — 그 두 상태는 언제나 결정한 댓글을 가지므로 빠짐이 없어요.
- REQ-4: When the CI job's decision is `pass`, the workflow shall post the `moai/audit` commit status `success` on the head itself, and when it is not `pass`, the state `failure` and a non-zero exit, so no PR waits for a hand-posted status. · failure: none — 상태 게시 자체가 거부 경로가 아니라 게시 실패는 CI 작업 실패로 보여요.
- REQ-5: While a comment is not from a repository member, the gate shall read no leader-assigned auditor from it, as with every verdict.
- REQ-6: While a leader-assigned auditor's token equals a `Moai-Session:` author of the pull request, the gate shall read the PASS as an author's and not count it.

## Acceptance

- AC-1 (REQ-1, REQ-2): `auditor: leader-audit audit-3218-1` 줄이 감사자로 읽혀 tier 1 통과를 만들고, 빈 토큰·허용 밖 문자는 감사자 없음으로 읽혀요.
  verify: `cargo test --locked --offline --lib audit_gate::tests::the_leader_assigned_auditor_line_is_a_verdict`
- AC-2 (REQ-3, REQ-4): pass/fail 결과가 결정한 댓글 id 를 내고, `.github/workflows/audit-gate.yml` 이 `statuses: write` 로 상태를 게시해요(성공·실패 모두; 실패는 비영 종료).
  verify: `cargo test --locked --offline --test audit_gate_cli the_result_names_the_comment_and_the_workflow_posts_the_status`
- AC-3 (REQ-5, REQ-6): 비회원의 위조 `leader-audit` 줄은 버려지고, 토큰이 PR 작성자 세션과 같으면 작성자 PASS 로 읽혀요.
  verify: `cargo test --locked --offline --lib audit_gate::tests::the_leader_assigned_auditor_line_is_a_verdict`
  verify: `cargo test --locked --offline --lib audit_gate::tests::the_decision_names_the_comment_that_decided_it`

## Scope

- write: `src/audit_gate.rs`, `tests/audit_gate_cli.rs`, `.github/workflows/audit-gate.yml`, `.moai/specs/SPEC-20261119-leader-auditor-verdict/spec.md`
- read-only: `src/gate.rs`(공유 파서 — 불변), `src/comment_authority.rs`, `.github/workflows/ci.yml`(검사 집합 대조)
- forbidden: 실제 HOME 과 원장, 네트워크(기록 전송 제외 — 시험은 RecordedTransport), 실제 원격 저장소, 키

## Plan

1. RED 먼저: `tests/audit_gate_cli.rs` 에 leader-audit 인식·위조·작성자·comment 필드 시험, `src/audit_gate.rs` 단위 시험에 같은 경계; 워크플로우 YAML 고정 시험. 명령·필드가 없어 먼저 실패해요.
2. `src/audit_gate.rs`: `verdict_auditor`(gate::auditor 후 leader-audit 줄), `decide()` 의 `comment` 필드.
3. `.github/workflows/audit-gate.yml`: 상태 게시 단계(`gh api repos/…/statuses/<sha>`), `statuses: write` 권한, state≠pass 시 failure 게시 후 비영 종료.
4. 전체 `--bin moai`(무필터, 자기점검 항목 3) + fmt + fmt-ops + clippy → M-Flow push → tier 1 자기 감사 → LANE-READY-INTEGRATION.

## Risks

- 상태 게시 단계는 워크플로우 파일(merge ref 버전)에 있어 PR 이 고칠 수 있어요 — 게이트 본체는 base 커밋에서 빌드한 바이너리라 판단 자체는 못 바꾸지만, 상태 게시를 무조건 성공으로 고치는 우회 경로는 남아요. 완전한 닫힘은 ruleset 의 상태 생성자를 GitHub Actions 앱으로 고정하는 사용자 클릭(병합 절차 문서의 D3)이에요; 그때까지 리더는 판정 댓글을 읽고 병합해요(수동 검토가 우회를 잡아요).
- `comment` 필드를 새로 읽는 도구는 없음 — 출력 추가라 하위 호환.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 감사자 인식은 audit_gate.rs 의 로컬 함수(gate::auditor 우선, 그 후 leader-audit 줄); gate.rs 불변 | 병합 관문과 파서 공유 — 이 카드 범위 밖; 레인 이름 우선순위 유지 | gate.rs 로 올리기(후속 카드) |
| D-2 | 상태 게시는 성공·실패 모두, 실패는 비영 종료 | BLOCKED 대기 제거 + 빨간 표시 유지 | 성공만 게시 |
| D-3 | `comment` 필드는 decide() 출력에 숫자로 | 워크플로우가 상태의 target_url 앵커를 만들게; reason 텍스트 파싱 금지 | reason 파싱 |

## Result

- AC-1: met — 2026-11-19 이 worktree(팁 a4cd5c72)에서 `cargo test --locked --offline --lib audit_gate::` 9 passed 0 failed(`the_leader_assigned_auditor_line_is_a_verdict`: `audit-3218-1` 토큰 인식 통과, 빈 토큰·`au@dit`·공백 토큰은 감사자 없음).
- AC-2: met — `cargo test --locked --offline --test audit_gate_cli` 4 passed 0 failed(`the_result_names_the_comment_and_the_workflow_posts_the_status`: pass=댓글 5·open-fail=댓글 6·tier 0 pass=null, YAML 고정 `statuses: write`·`statuses/$PR_HEAD`·`moai/audit`·`-f state=`); `.github/workflows/audit-gate.yml` 이 상태를 성공·실패 모두 게시하고 실패 시 비영 종료.
- AC-3: met — 같은 lib 시험: 비회원 위조 누락(NONE), `id:key` 토큰은 문자집합(`:` 불가)으로 거부, 베어 세션 id 토큰은 `leader_token` 작성자 검사로 거부. 변이 2 KILLED — M1(`leader_token` 작성자 검사 제거)은 첫 실행에서 생존(`id:key` 토큰이 문자집합에서 먼저 걸려 검사 미도달 — 시험 공백), 베어 id 사례 추가 후 exit 101; M2(문자집합 무효화) exit 101. 복원 후 `diff -q` 클린.
- 이웃: `cargo test --locked --offline --bin moai` 28 passed·`--bin moai-cli` 28 passed(무필터), `cargo fmt --check` 0, `rustfmt --check --edition 2021 src/ops/*.rs` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0. RED 먼저: lib 신규 2 실패(101) + cli 1 실패(101).

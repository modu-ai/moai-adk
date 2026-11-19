---
id: SPEC-20261119-agents-local-rules
title: "AGENTS.local.md states five lane and leader rules from the 2026-11-19 bundle 5 delay"
status: draft
revision: 1
created: 2026-11-19
---

# AGENTS.local.md states five lane and leader rules from the 2026-11-19 bundle 5 delay

## Intent

사용자가 2026-11-19 에 PR 3229(묶음 5)가 한 시간쯤 기다린 일을 두고, 같은 일이 되풀이되지 않도록 다섯 규칙을 `AGENTS.local.md` 에 적으라고 했어요(카드 C-123 본문). 이 조각은 문서만 고쳐요: 코드와 시험은 한 줄도 바꾸지 않아요. 다섯 규칙은 (1) 묶음 PR 의 판정은 커밋을 쓰지 않은 레인이 자기 이름과 `Moai-Session` 을 적어 쓰고, 판정 뒤 `audit-gate` 를 다시 돌려 그 job 이 같은 head 에서 성공했을 때만 리더가 `moai/audit` 를 올린다, (2) 감사 배차는 한도를 이름한다(전체 `cargo test` 금지, `python3 -m unittest discover` 금지, 이름 붙인 모듈만), (3) 묶음을 조립한 직후와 레인이 push 하기 전에 `python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 를 먼저 돌린다, (4) 세션 종료 알림은 "지금 쥔 카드를 끝내고, 새 카드는 받지 않고, `SESSION-END-READY` 를 보고한다" 는 뜻이지 하던 일을 멈추라는 뜻이 아니다, (5) 묶음은 작게 둔다(5~7개 head), 검증에 실패한 head 는 빼고 나머지를 보낸다.

## Non-goals

- 규칙을 코드나 훅으로 강제하기: 이 조각은 글만 적어요(`audit-gate` 의 `non-author-pass` 규칙은 이미 `src/audit_gate.rs` 에 있어요).
- `skills/moai/workflows/**` 나 `AGENTS.md` 고치기: 두 곳은 이 카드 범위 밖이에요.
- 기존 규칙 문장 바꾸기·지우기: `Gate` 절의 오래된 줄(예: auditor 는 `agent-2`)도 이 조각에서는 그대로 둬요.
- 카드 본문에 없는 사건이나 원인 지어내기: 원인은 카드 본문과 PR 3229 본문에서 읽은 사실만 적어요.

## Requirements

- REQ-1: The file `AGENTS.local.md` shall state that a bundle PR's verdict is written by a lane that wrote none of the PR's commits and that the comment names the lane (`agent-N`) and its `Moai-Session` value, with the date 2026-11-19 and the cause.
- REQ-2: The file `AGENTS.local.md` shall state that a verdict written by the leader's own subagent is not read by `audit-gate` (rule `non-author-pass`).
- REQ-3: The file `AGENTS.local.md` shall state that the `audit-gate` run is re-run after the verdict and that the leader posts `moai/audit` only when that job succeeded on the same head, never on a green `ci` and a PASS comment alone.
- REQ-4: The file `AGENTS.local.md` shall state that an audit dispatch names its caps: no whole `cargo test`, no `python3 -m unittest discover`, named modules only, with the date and the cause.
- REQ-5: The file `AGENTS.local.md` shall state that `python3 -m unittest tests.repo_scans_test tests.source_manifest_test` runs first, right after a bundle is assembled and before every lane push, with the date and the cause.
- REQ-6: The file `AGENTS.local.md` shall state that a session-end notice means finish the card in hand, take no new card and report `SESSION-END-READY`, that it never means stop the work in hand, and that a card in `review` cannot be released so its session stays until the merge.
- REQ-7: The file `AGENTS.local.md` shall state that bundles stay small at five to seven heads and that a head failing verification is dropped while the rest ships, with the date and the cause.
- REQ-8: The file `AGENTS.local.md` shall carry at least one row per rule in its `Rule provenance` table that names the user instruction of 2026-11-19 and card C-123.
- REQ-9: The change shall leave the repository's doc-reference, repo-scan and source-manifest checks passing.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): `Gate` 절이 규칙 (1)을 날짜·원인(PR #3229, 리더 하위 에이전트의 PASS 댓글 7004447253 이 감사자 레인 이름 없음으로 거절됨, 한 시간쯤 기다림)과 함께 적어요.
  manual: `AGENTS.local.md` 를 읽어 `## Gate` 절에서 `non-author-pass`·`gh run rerun`·`moai/audit` 를 찾고 결과를 Result 에 한 줄로 적어요.
- AC-2 (REQ-4): `Gate` 절이 감사 배차 한도(규칙 2)를 날짜·원인(PR #3229 감사가 41분 걸림)과 함께 적어요.
  manual: `AGENTS.local.md` 에서 `no whole` 과 `discover` 를 찾아 Result 에 적어요.
- AC-3 (REQ-5): `Agent loop` 절이 규칙 (3)을 날짜·원인(2026-11-19 에 다섯 head 가 목록에 없는 훑기 시험이나 낡은 manifest 로 묶음 검증에 실패)과 함께 적어요.
  manual: `AGENTS.local.md` 에서 `repo_scans_test` 를 찾아 Result 에 적어요.
- AC-4 (REQ-6): `Agent loop` 절이 규칙 (4)를 `SESSION-END-READY` 와 함께 적어요.
  manual: `AGENTS.local.md` 에서 `SESSION-END-READY` 를 찾아 Result 에 적어요.
- AC-5 (REQ-7): `Leader supervision` 절이 규칙 (5)를 날짜·원인(묶음 5 가 13개 head 였고 셋을 빼야 했으며 감사가 41분 걸림)과 함께 적어요.
  manual: `AGENTS.local.md` 에서 `5 to 7` 을 찾아 Result 에 적어요.
- AC-6 (REQ-8): `Rule provenance` 표에 여섯 행이 늘어요(규칙 (1)은 두 줄이라 두 행).
  manual: `AGENTS.local.md` 에서 `card C-123` 이 든 표 행을 세어 Result 에 적어요.
- AC-7 (REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline --test doc_references`
  verify: `cargo test --locked --offline --test doc_reference_boundaries`

## Scope

- write: `AGENTS.local.md`, `.moai/specs/SPEC-20261119-agents-local-rules/spec.md`
- read-only: `skills/moai/workflows/mode-agent.md`, `skills/moai/references/internal/commit-trailers.md`, `src/audit_gate.rs`
- forbidden: 실제 HOME, 네트워크, `AGENTS.md`, `skills/moai/workflows/**`

## Plan

1. SPEC 을 쓰고 `spec check` 로 확인해요(빌드한 `moai-cli`).
2. `AGENTS.local.md` 를 고쳐요: `Gate` 절에 규칙 (1)·(2)를, `Agent loop` 절에 규칙 (3)·(4)를, `Leader supervision` 절에 규칙 (5)를, `Rule provenance` 표에 다섯 행을 더해요. 문장은 영어, 날짜와 원인은 카드 본문과 PR 3229 에서 읽은 사실만 써요.
3. `python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 와 `cargo test --test doc_references` 를 돌려요.

## Risks

- 이 파일은 `AGENTS.md` 와 합쳐지는 입력이라(`tests/agents_local_merge_cli.rs`) 줄 모양을 깨면 안 돼요: 기존 절과 표 머리를 그대로 두고 줄만 더해요.
- 원인을 카드 본문이 주지 않은 규칙(4)은 ledger 의 규칙(`card release` 는 `review` 카드를 거절)을 원인으로 적어요. 사건 원인을 지어내지 않아요.
- 훑기 시험(`repo_scans_test`)이 새 문구에 걸릴 수 있어요: 시험을 돌려 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙 (1)(2)는 `Gate` 절, (3)(4)는 `Agent loop` 절, (5)는 `Leader supervision` 절에 둔다 | 각 규칙이 가장 가까이 속한 절; 새 절을 만들지 않아 파일 모양 유지 | 줄을 다른 절로 옮기기 |
| D-2 | 날짜는 인라인과 `Rule provenance` 표 둘 다 적는다 | 카드 기준 1("각 규칙에 날짜와 원인")과 카드 본문("출처는 Rule provenance")을 함께 만족 | 표 행 빼기 |
| D-3 | 규칙 (4)의 원인은 카드가 사건을 주지 않으므로 ledger 규칙으로 적는다 | 사건을 지어내지 않기(코어 규칙 10) | 리더가 사건을 주면 한 줄 고치기 |

## Result

- AC-1: met — `AGENTS.local.md:80-81` (`## Gate`): 규칙 (1) 두 줄. `non-author-pass` 1회, `gh run rerun` 1회, `moai/audit` 가 든 줄 둘(`only when that job succeeded on the same head`), 날짜 2026-11-19, 원인 PR #3229 댓글 7004447253(감사자 레인 이름 없음, 한 시간쯤 기다림). 읽은 방법: `grep -c` 로 센 값과 `grep -n`.
- AC-2: met — `AGENTS.local.md:82` (`## Gate`): `no whole` 이 든 줄 둘(82 본문, 133 provenance), `unittest discover` 가 든 줄 셋(50 은 이미 있던 `Check set` 코드 블록, 82 본문, 133 provenance), 원인 PR #3229 감사가 41분 걸림. 명령 `grep -n "no whole" AGENTS.local.md` 와 `grep -n "unittest discover" AGENTS.local.md`. 50줄의 `Check set` 은 작성자가 PR 전에 돌리는 목록이고 새 규칙은 감사 배차의 한도라서 그대로 뒀어요(둘의 관계는 후속으로 리더가 정해요).
- AC-3: met — `AGENTS.local.md:41` (`## Agent loop`): `tests.repo_scans_test tests.source_manifest_test` 1회, 날짜 2026-11-19, 원인 다섯 head 가 목록에 없는 훑기 시험이나 낡은 manifest 로 실패.
- AC-4: met — `AGENTS.local.md:42` (`## Agent loop`): `SESSION-END-READY` 가 파일에 2회(본문 1줄, provenance 1줄), `review` 카드는 풀 수 없다는 문장 포함.
- AC-5: met — `AGENTS.local.md:73` (`## Leader supervision`): `5 to 7` 이 파일에 2회(본문, provenance), 원인 묶음 5 가 13개 head 였고 C-103·C-59·C-110 셋을 뺐고 감사가 41분 걸림.
- AC-6: met — `grep -c '^| `AGENTS.local.md` ' AGENTS.local.md` = 6 (`AGENTS.local.md:131-136`), 모두 `card C-123` 을 이름해요.
- AC-7: met — `cargo test --offline --locked --test doc_references --test doc_reference_boundaries --test ac_command_lint --test spec_cli`: doc_references 7, doc_reference_boundaries 15, ac_command_lint 5, spec_cli 18 통과(exit 0); `python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 68 OK; `python3 -m unittest tests.ci_scope_test` OK; 빌드한 `moai-cli spec check` errors 0. 전체 `cargo test`·`python3 -m unittest discover`·CI 는 돌리지 않았어요(`not_run`).

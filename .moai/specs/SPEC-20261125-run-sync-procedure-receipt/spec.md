---
id: SPEC-20261125-run-sync-procedure-receipt
title: "Cut RUN and SYNC to numbered steps and a failure table, and state the receipt rule once"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9183): \"승인, D-1~D-6 기본값 A (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:8cafbb9d36a702bce1cb95663d7d6301fccd1f86daeeb612022530f63d1e4d59"
---

# Cut RUN and SYNC to numbered steps and a failure table, and state the receipt rule once

## Interview

- Q1: 스킬 지침을 어떻게 고칠까요? → A: 사용자 지시 그대로(스킬 전수 조사, 2026-11-25): "moai 스킬 지침 전체 전수 조사해서 검색 말고 전체 개별로 읽고 개선점을 찾아서 모두 업데이트 개선 보고 후 진행". 이 SPEC 은 그 조각 중 구현·문서 맞춤 글(`workflows/run.md`, `workflows/sync.md`, `references/internal/self-check.md`, `test-design.md`)이에요(조사 보고서 5번).
- Q2: 관문(작성자는 감사하지 않음, 판정은 감사자가 기록, 영수증 없이는 PASS 아님, HOLD, Python 금지)을 바꾸나요? → A: 아니요. 글이 짧아지고 한 곳에 모일 뿐 관문은 약해지지 않아요(REQ-9).
- Q3: 영수증 규칙은 무엇인가요? → A: 사용자 결정 그대로(CLAUDE.md "Verify", 2026-11-24·25): `check affected --card` 를 READY 에 한 번, 통과한 같은 트리에서는 다시 돌리지 않고, 넓은 세트(맨 `check affected`)는 lander 가 묶음마다 한 번이에요.
- Q4: 시험과 뮤턴트 규칙은요? → A: 사용자 결정 그대로: 뮤턴트는 tier 2 에서만(CLAUDE.md "Guard mutants"), 이 저장소의 시험은 Rust 만이고 Python 은 쓰지도 실행하지도 않아요(CLAUDE.local.md, 2026-11-25).

## Intent

두 글에는 오늘의 결정과 어긋난 문장과 읽는 사람이 따라갈 수 없는 긴 규칙이 섞여 있어요. 읽은 사실(2026-11-25, `wc -w` 와 파일을 처음부터 끝까지 읽음, 조사 보고서 `~/.moai/reports/moai-adk/skill-survey-20261125.md`):

- `workflows/run.md` 는 6,065 단어예요. 번호 절차는 8단계뿐이고 그 뒤에 규칙 문단이 4,600 단어쯤 이어져요(슬롯·정리·캐시가 66행 문단 하나에 약 900 단어). 절차의 어느 단계도 `moai check affected` 를 부르지 않아요. 내부 번호(Issue #, 카드 C-) 25개 이상, moai-adk 내부 Python 도구(`tools/ci_scope.py`, `tools/source_manifest.py`, `tools/self_check.sh`, `tools/build_folder.py`), 퇴역 호스트(Kiro, Cursor)가 남아 있어요.
- 결정과 어긋난 문장: `run.md` 36·68행과 `sync.md` 23행(영수증 때 전체 세트를 돌림), `self-check.md` 11행("never narrowed"), `self-check.md` 10행과 `test-design.md` 8행(뮤턴트를 모든 변경에서 요구; 결정은 tier 2 만), `self-check.md` 14행·`sync.md` 23행(Python 시험), `test-design.md` 12행("Rust 로 쓴다" 는 Rust 가 아닌 사용자 프로젝트에서 틀림).
- `sync.md` 는 레인이 하는 일(SPEC Result·문서·`@cc:` 주석)과 lander·릴리스가 하는 일(착지, 릴리스 pull request, CI 증거, 효과 기록)이 한 번호 목록에 섞여 있어요. 레인은 앞쪽만 읽으면 되는데 둘 다 읽어야 해요.
- 어느 경로 글도 이름으로 부르지 않는 참조 다섯(`tdd-results`, `check-completeness`, `doc-claims`, `boundary-checks`, `security-checklist`)은 기본 에이전트의 `reads:` 줄로만 읽혀요. 위임하지 않는 세션은 영영 읽지 않아요.

이 SPEC 이 하는 일: `run.md` 를 번호 절차와 실패표로 줄이고(1,200 단어 이하), 긴 규칙은 그 단계가 부르는 참조 넷으로 옮겨요. 영수증 규칙은 한 문장으로 한 곳(`run.md` 의 영수증 단계)에 두고 다른 글은 거기를 가리켜요. `sync.md` 는 레인 몫만 남기고 lander·릴리스 몫은 참조 하나로 옮겨요. `self-check.md` 와 `test-design.md` 는 결정과 어긋난 줄을 고치고 moai-adk 내부 도구를 지워요.

### 문장 대조표 (무엇이 어디로 가는가)

| 지금 글 | 고친 뒤 |
| --- | --- |
| `run.md` Input·Output·Procedure 1-8 | 새 Procedure: 한 행동 한 단계, 명령이나 참조를 부름. 영수증 단계가 `moai check affected --card <base>..<head>` 와 한 문장 규칙을 가져요(REQ-1, REQ-3) |
| `run.md` "Check-run rules" 표·CI 인용 조건·릴리스 CI 범위·docs-only 가벼운 CI | `references/internal/check-rules.md` 로 옮겨요. 내부 번호와 `tools/ci_scope.py` 는 지워요(REQ-5, REQ-6) |
| `run.md` "Check-run rules" 의 tier 네 줄(docs-only, test-only, code, high-risk)과 floor 목록 가리킴 | 옮기지 않고 지워요. tier 줄의 단 하나의 자리는 감사 계약 `references/audit.md` 의 "Audit tiers" 절이고(감사 글 정리가 정한 대로), `check-rules.md` 는 그 절을 가리키기만 해요. 옛 docs-only 줄("lander 가 읽고 착지한다")은 이렇게 사라져요(REQ-5) |
| `run.md` 시험 규칙(시계 금지, 흔들리는 시험 규칙, 병렬 임시 경로, 자식 프로세스 예산) | 같은 `check-rules.md` 의 "Tests" 절. `moai flake-classify`·`flake-record` 는 그대로 이름을 불러요 |
| `run.md` "Heavy checks are serialised", "Lane worktrees and build cache" | `references/internal/build-slots.md`. 슬롯 명령, 빌드 폴더, 정리, 디스크·메모리 한도, 빌드 캐시 사실. Python 슬롯 문장과 `CARGO_HOME` 관찰 한 줄은 지워요 |
| `run.md` "Change-specific rules", "Shell and helper rules", "Host worktree guard" 표 | `references/internal/shell-guards.md`. 규칙은 그대로이고 Issue 번호와 moai-adk 내부 스크립트 이름만 지워요 |
| `run.md` "Web pages" | `references/internal/web-pages.md` 로 그대로 옮겨요(브라우저 안내 자체는 따로 맡은 카드 몫) |
| `run.md` "Base agents" | 그대로 두되 `Kiro and Cursor` 줄을 지우고 impl 이 `base` 를 확인하는 줄은 남겨요. 위임 글은 이미 착지한 위임 글 정리의 모양을 따라요 |
| `run.md` Resume, Exceptions and Permissions, Completion Check | 짧게 남겨요. 실패표가 맡는 행동은 표로 가요 |
| `sync.md` Input·Output·States·Before the audit·Procedure 3 | `sync.md` 에 남아요(레인 몫): 기준 기록, 차이 분류(표 한 장), SPEC Result·문서·`@cc:`·`moai spec check`·커밋, 위임 `moai-docs`(`doc-claims` 를 이름으로 부름) |
| `sync.md` Procedure 1(릴리스 가지), 2(전체 검사 세트), 4-9(효과, 착지, 릴리스 PR, CI 증거, 독립 검토, 릴리스 병합, 대조) 와 영수증 필드 | `flows/landing-release`(파일 `references/flows/landing-release.md`)로 옮겨요. 2는 지워요(SYNC 는 검사 세트를 돌리지 않고 레인의 영수증을 인용해요) |
| `self-check.md` 5행 `tools/self_check.sh` 문단, 6번 Python, 13번 `tools/source_manifest.py` | 지워요. 일반 항목(`moai check affected --card`, `git diff --check`, 프로젝트 자기 검사)만 남아요 |
| `self-check.md` 2번 뮤턴트 상한 "tier 1 5개, tier 2 12개", 3번 "never narrowed" | 2번: tier 2 에서만 `guard-mutant:` 줄, 아래 tier 는 요구 없음. 3번: 영수증 단계(`run.md`)를 가리켜요 |
| `test-design.md` 4번(모든 시험에 뮤턴트), 8번(Rust·Python), 9번(작업량 세기 200 단어) | 4번은 tier 2 에서만, 8번은 "프로젝트의 시험 언어로 쓰고 시험을 위해 두 번째 언어를 들이지 않는다", 9번은 한 문단으로 줄여요 |

## Non-goals

- 관문을 느슨하게 하거나 바꾸지 않아요: 작성자는 감사하지 않음, 판정은 감사자가 기록, 영수증 없이는 PASS 아님, HOLD, Python 금지, FAIL 네 가지.
- `workflows/audit.md`, `references/audit.md`, `workflows/work.md`, `workflows/leader.md`, `flows/next-flow.md`, `SKILL.md` 의 문장은 고치지 않아요(각 카드 몫). 이 글들이 `run.md` 의 "Check-run rules" 를 가리키는 줄만 새 참조 이름으로 바꿔요.
- `moai` 명령의 동작과 Rust 코드는 바꾸지 않아요. `moai check affected` 의 다시 돌리기 동작은 그 명령을 고치는 다른 카드 몫이에요.
- `plan.md`, `sdd.md`(SPEC 없는 카드 갈래), `versioning.md`, `project.md`, `update.md` 는 건드리지 않아요.

## Requirements

<!-- This SPEC edits skill texts and keeps every blocking rule they state; no check that blocks is added or loosened. -->
- REQ-1: The `## Procedure` of `workflows/run.md` shall be one numbered list of one-action steps in work order (preflight, read, test first, change, run the SPEC's `verify:` commands, self-check, receipt, hand to SYNC), each naming the command or the reference that does it. · failure: none — a reader who follows the numbers meets each action once.
- REQ-2: `workflows/run.md` shall hold at most 1,200 words by `wc -w`, in exactly these second-level sections: `## Input`, `## Output`, `## Procedure`, `## When a step stops` (the failure table), `## Base agents`, `## Resume`, `## Exceptions and Permissions` and `## Completion Check`. · failure: none — a longer or wider file is a finding of the audit of this change.
- REQ-3: The failure table of `workflows/run.md` shall have one row, each naming its next action, for a red check, a `0 passed` run, a missing guard, a refused `moai card check`, a slot that exits 75, a failing receipt, a moved base, a `moai-impl` that ends `blocked` and a stalled or interrupted attempt. · failure: none.
- REQ-4: The receipt rule shall appear in the edited files as one sentence, in the receipt step of `workflows/run.md`: run `moai check affected --card <base>..<head>` once, at READY, on the checkout at the head, never again on a tree that already passed it, and leave the wide set (a bare `moai check affected`) to the lander's `moai card land-batch`; every other edited file that touches the receipt points to that step. · failure: none — a second statement is a finding.
- REQ-5: The long rules of `workflows/run.md` shall move, without losing a rule that still holds, into `references/internal/check-rules.md` (who runs which check, citing a remote run, the test rules, the flaky rule), `references/internal/build-slots.md` (slots, build folders, cleanup, limits, build cache), `references/internal/shell-guards.md` (change-specific rules, shell and helper rules, the host worktree guard table) and `references/internal/web-pages.md`, each named by the step or table row of `run.md` that needs it; the four tier lines and the floor list leave `run.md` and appear in no edited file, which point to the section "Audit tiers" of the audit contract instead. · failure: none — a rule with no home is a finding.
- REQ-6: The edited and the new texts shall carry no Issue number, no card number, no moai-adk internal tool (`tools/…`), no Python step and no retired host (Kiro, Cursor, Codex, OMP, the Go generation), and name only `moai` commands that answer `--help` with exit 0. · failure: none — a leftover is a non-blocking finding.
- REQ-7: The steps of `workflows/run.md` and the `moai-docs` step of `workflows/sync.md` shall name the five references that no route text named before: `internal/tdd-results` (the test-first step), `internal/check-completeness` (when a check, gate or criterion is added or judged), `internal/boundary-checks` (when a change crosses a boundary), `internal/security-checklist` (when it handles authentication, input, secrets, outgoing requests or an interface others call) and `internal/doc-claims`. · failure: none — an unnamed reference is a finding.
- REQ-8: `workflows/sync.md` shall keep the lane part (record the base, classify differences, write `## Result`, run `moai spec check`, commit with the `Moai-Spec` trailer, hand the writing to `moai-docs`), cite the lane's receipt instead of running a check set, and name `flows/landing-release` once for the rest. · failure: none.
- REQ-9: `flows/landing-release` (the file `references/flows/landing-release.md`) shall hold the lander and release part of today's `sync.md` whole: effects, landing, the release pull request, CI evidence, the independent review of a release, the release merge, reconcile and the integration receipt. · failure: none — a part with no home is a finding.
- REQ-10: The edited texts shall keep, with the same meaning, that the author never audits, that the auditing session records its own verdict, that a check that was not run is `not_run` and never a pass, that a HOLD or a `guarded` choice is never decided by a lane, and that tests are never Python. · failure: a missing rule is a FAIL of the weakened-gate kind.
- REQ-11: `references/internal/self-check.md` and `references/internal/test-design.md` shall state the decisions of 2026-11-24 and 25: `guard-mutant:` lines and a mutation proof at tier 2 only, the set before READY by a pointer to the receipt step of `workflows/run.md`, and the project's own test language instead of Rust or Python. · failure: none.
- REQ-12: Each test and each fixture and each manifest entry and each pointer that names a sentence or section these texts lose or move shall name its new place in the same commit, and a pointer to the build rule names step 1 of `workflows/run.md` so that no agent's `reads:` list changes; a test that cannot is deleted with its reason in the commit message. · failure: none — an unexplained deletion or a dangling pointer is a finding.

## Acceptance

- AC-1 (REQ-1, REQ-7): `## Procedure` of `run.md` is one numbered list whose items name, in this order, `moai card check`, `internal/test-design`, `internal/tdd-results`, `moai-impl`, the SPEC's `verify:` commands, `internal/self-check`, `moai check affected --card` and SYNC; the other reference names of REQ-7 appear in the step they belong to, and `sync.md` names `internal/doc-claims` · verify: `cargo test --offline --locked --test run_sync_text ac1_`
- AC-2 (REQ-2, REQ-3): `run.md` has at most 1,200 words, its second-level headings are exactly the eight of REQ-2, and the table under `## When a step stops` has the nine rows of REQ-3 · verify: `cargo test --offline --locked --test run_sync_text ac2_`
- AC-3 (REQ-4): the receipt sentence appears once across the edited files, in `run.md`; `self-check.md`, `sync.md` and `check-rules.md` point to the step; none of "never narrowed", "full check set" or "wide set" outside that sentence and `check-rules.md`'s pointer to the lander remains · verify: `cargo test --offline --locked --test run_sync_text ac3_`
- AC-4 (REQ-5): the four new references exist, are listed in `source-manifest.json` with an `origin`, hold the rule texts named in REQ-5 (a probe per rule: `limits.json`, `moai heavy-slot-acquire`, `moai flake-classify`, the host worktree guard table, a browser rule), `run.md` names each, no file of `skills/` or `agents/` names "Check-run rules", and the four tier lines appear only in the audit contract · verify: `cargo test --offline --locked --test run_sync_text ac4_`
- AC-5 (REQ-6): none of the edited or new texts contains `Issue #`, a `#` followed by four digits, a card id `C-<digits>`, `tools/`, `python3`, `.py`, `unittest`, `Kiro`, `Cursor`, `Codex`, `OMP` or `Go generation`, and every `moai` subcommand they name answers `--help` with exit 0 · verify: `cargo test --offline --locked --test run_sync_text ac5_`
- AC-6 (REQ-8, REQ-9): `sync.md` has no step that runs a check set, names `flows/landing-release` once and keeps the lane part; that file holds the effects, landing, release pull request, CI evidence, independent review, release merge, reconcile and the receipt fields · verify: `cargo test --offline --locked --test run_sync_text ac6_`
- AC-7 (REQ-10): the five rules of REQ-10 are present in the edited texts; deleting any one sentence makes the test fail (guard mutant) · verify: `cargo test --offline --locked --test run_sync_text ac7_`
- AC-8 (REQ-11): `self-check.md` has no `tools/`, no Python item and mentions `guard-mutant` only with "tier 2"; `test-design.md` names the project's test language and a mutation proof only at tier 2 · verify: `cargo test --offline --locked --test run_sync_text ac8_`
- AC-9 (REQ-12): each pinned test that read a moved sentence now reads its new place or is gone with the reason in the commit message, the manifest entries match `shasum -a 256` of their files, and no file of `skills/` or `agents/` points to a section that no longer exists · verify: `cargo test --offline --locked --test skill_structure` and `cargo test --offline --locked --test run_sync_text ac9_`
- AC-10 (REQ-12): `moai check affected --card <base>..<head>` ends `result=PASS` at READY · verify: `moai check affected --card <base>..<head>`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천(A)이 맨 앞이고 사용자가 고르지 않으면 기본으로 A 를 써요.

- D-1 (긴 규칙을 어디로): (A, 기본·권장) 규칙은 지키되 목적별 참조 넷(`check-rules`, `build-slots`, `shell-guards`, `web-pages`)으로 옮기고 단계가 이름으로 불러요. (B) 명령이 이미 스스로 지키는 규칙(슬롯 대기, 한도 파일)은 글에서 지워요 — 더 짧지만 명령이 하는 일을 사람이 못 읽어요.
- D-2 (`sync.md` 나누기): (A, 기본·권장) 레인 몫은 `sync.md`, lander·릴리스 몫은 `flows/landing-release` 로 파일을 나눠요. (B) 한 파일 안에서 두 절로만 나눠요 — 레인이 여전히 둘 다 읽어요.
- D-3 (낡은 글): (A, 기본·권장) Python 슬롯, `tools/ci_scope.py`, 스모크 하네스, Issue 원장 시절 흔들리는 시험 절차, Kiro·Cursor 줄, 내부 번호를 지워요. 지금도 맞는 CI 인용 조건과 릴리스 CI 범위는 `check-rules.md` 에 남겨요. (B) 낡은 글도 옮겨 보존해요 — 글이 줄지 않아요.
- D-4 (다섯 참조를 부르는 자리): (A, 기본·권장) REQ-7 의 조건을 단계에 한 줄로 달아요. (B) 에이전트의 `reads:` 줄에만 두어요 — 위임하지 않는 세션은 읽지 않아요.
- D-5 (Rust 고정 시험): (A, 기본·권장) 옮겨진 규칙을 고정한 시험은 새 자리를 읽게 고치고, 결정이 뒤집은 규칙(영수증 때 전체 세트, 모든 변경에 뮤턴트, Python 시험)을 고정한 시험은 이유를 커밋 메시지에 적어 지워요. 낱말을 약하게 하지 않아요. (B) 모든 고정 시험을 지워요 — 지켜야 할 규칙이 안 지켜져도 모르게 돼요.
- D-6 (이 글을 고정한 Python 시험): 읽은 사실(2026-11-25): Python 시험 여섯 파일이 이 글들을 읽어요 — `tests/check_pr_self_audit_test.py`(`run.md` 의 Check-run rules 표와 CI 인용 문장, `sync.md`), `tests/self_check_test.py`(`self-check.md` 의 `tools/self_check.sh`·`--full-python`·`--pre-ready` 문장과 `run.md` 슬롯 문단), `tests/smoke_harness_test.py`(`run.md` 의 스모크 하네스 줄), `tests/skill_web_pages_test.py`(`run.md` 의 Web pages), `tests/sync_drift_fixture_test.py`(`sync.md` 3단계), `tests/skill_references_test.py`(스킬 글의 경로가 실제로 있는지). `moai check affected` 는 Python 시험 파일을 고르지 않아요(`src/check_affected.rs`: "the Python tests that remain run in CI only"). 그래서 영수증은 영향이 없고, 원격 CI 의 Python 단계만 깨져요(원격 검사는 착지를 막지 않아요). Python 시험을 Rust 로 옮기고 지우는 일은 이미 다른 카드(`Tests and tools are Rust only`)가 맡고 있어요. (A, 기본·권장) Python 시험 파일은 건드리지 않아요(새 Python 도, Python 파일 편집도 없어요). 이 글의 규칙은 `tests/run_sync_text.rs` 가 Rust 로 계속 고정하고, 깨질 Python 시험 이름과 원인을 `docs/later.md` 한 줄과 그 Rust 전환 카드의 노트로 남겨요(`sync.md` 3단계는 낱말을 그대로 둬요). (B) Python 시험의 해당 함수만 지워요 — 지우기만 하지만 `.py` 파일 편집이라 사용자 지시("Python 을 쓰지도 돌리지도 않는다")에 닿으니 사용자가 정해야 해요. (C) 글 문장을 그대로 두어 Python 시험이 계속 통과하게 해요 — 글이 줄지 않아요.

## Scope

- write: `.moai/specs/SPEC-20261125-run-sync-procedure-receipt/**`, `skills/moai/workflows/run.md`, `skills/moai/workflows/sync.md`, `skills/moai/references/internal/self-check.md`, `skills/moai/references/internal/test-design.md`, `skills/moai/references/internal/tdd-results.md`, `skills/moai/references/internal/check-completeness.md`, `skills/moai/references/internal/doc-claims.md`, `skills/moai/references/internal/boundary-checks.md`, `skills/moai/references/internal/security-checklist.md`, `tests/run_sync_text.rs`
- write (새 파일): `skills/moai/references/internal/check-rules.md`, `skills/moai/references/internal/build-slots.md`, `skills/moai/references/internal/shell-guards.md`, `skills/moai/references/internal/web-pages.md`, `skills/moai/references/flows/landing-release.md`
- write (가리키는 줄만): `skills/moai/references/audit.md`, `skills/moai/workflows/audit.md`, `skills/moai/references/internal/git-safety.md`, `skills/moai/references/chains/index.md`, `agents/impl.md`, `agents/audit.md`
- write (고정 문장이 바뀔 때만): `tests/agent_delegation_text.rs`, `tests/agent_defs_package.rs`, `tests/next_flow_skill_scan.rs`, `tests/sdd_guidance.rs`, `tests/core_rules_2876.rs`, `tests/test_isolation_rules.rs`, `tests/flaky_followups.rs`, `tests/review_checklist_cli.rs`, `tests/skill_structure.rs`, `tests/workflow_size_budget.rs`, `tests/lane_rules_2879.rs`, `tests/lane_rules_2836.rs`, `tests/lane_rules_2401.rs`, `tests/sync_marker_form.rs`, `tests/plan_output_sections.rs`, `tests/product_cli.rs`, `tests/card/card_ready_leftovers_cli.rs`, `src/heavy_slot.rs`(`include_str!` 시험의 읽는 자리만), `tests/home_style_ownership_cli.rs`
- write (한 줄): `docs/later.md`(D-6 의 A: 깨질 Python 시험과 원인), `src/lane_env.rs`(머리 주석이 옛 절 이름을 부르는 줄만)
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`, `tests/fixtures/update_tux/*.json`, `tools/repo_scans.json`(흔들리는 시험 항목의 범위 경로)
- read-only: `src/card/**`, `skills/moai/workflows/work.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/next-flow.md`, `tools/review_checklist.json`(`review_checklist_cli` 시험이 옮긴 규칙과 글자 그대로 맞춰 읽어요), 모든 `tests/*.py`(D-6)
- forbidden: `workflows/work.md`·`leader.md`·`flows/next-flow.md`·`SKILL.md` 의 문장 수정, 관문을 느슨하게 하는 변경, 새 Python 코드, 전역 설치본과 실제 HOME, 네트워크

## Plan

1. SPEC 승인을 기다려요(tier 2: `skills/` 글과 관문 글을 옮겨요). 승인 전에는 글을 쓰지 않아요.
2. 승인 뒤 `card scope` 로 쓰기 범위를 위 Scope 대로 넓히고(다른 카드가 쥔 경로는 반납 뒤에: `tests/core_rules_2876.rs` 는 스킬 입구 글을 고치는 카드가 감사 중이라 그 카드가 착지한 뒤에만 고쳐요), 고정 시험 조사 결과의 표(어느 시험이 어느 문장을 고정하는가)를 카드 노트에 남겨요.
3. RED: `tests/run_sync_text.rs`(AC-1..AC-9)가 지금 글에서 실패해요.
4. 새 참조 다섯을 쓰고(규칙은 옮기고 낡은 글은 지워요), `run.md`·`sync.md`·`self-check.md`·`test-design.md` 를 고치고, 가리키는 줄을 바꿔요. 참조는 지금 글의 꼴대로 `internal/<이름>` 과 `flows/<이름>` 으로 가리켜요. 빌드 규칙을 가리키는 줄(`self-check.md`, `git-safety.md`, 두 에이전트 파일, 두 감사 글)은 새 참조 이름이 아니라 `run.md` 1단계를 가리켜서 에이전트의 `reads:` 줄이 바뀌지 않아요. 새 글의 모든 명령은 `--help` 로 확인해요. GREEN.
5. 고정 시험을 새 자리로 고치거나 이유를 적어 지워요(D-5). `agent_defs_package.rs` 의 `references/internal/` 파일 목록(고정돼 있어요)에 새 파일 넷을 더하고, 새 참조마다 `chains/index.md` 가 불러요. Python 시험은 D-6 대로 건드리지 않고 `docs/later.md` 한 줄만 더해요. `source-manifest.json`(새 파일의 `origin` 포함)과 `update_tux` 픽스처를 `shasum -a 256` 으로 갱신해요(line-lease).
6. 가드 뮤턴트(관문 문장을 글에서 지우면 AC-7 이 죽어요), `cargo fmt --all`, `moai check affected --card`, READY.

## Risks

- R-1: 규칙을 옮기다 빠뜨릴 수 있어요. 대조표가 옮길 곳을 정하고 AC-4 가 규칙마다 탐침 한 줄로 확인해요.
- R-2: 고정 시험이 많고(조사: Rust 시험 20여 개와 Python 시험 여섯 파일) 문장 하나가 빠지면 실패해요. 하나씩 새 자리에 다시 고정하고, 결정이 뒤집은 규칙을 고정한 시험만 이유를 적어 지워요.
- R-3: 글을 줄이다가 관문 규칙이 빠질 수 있어요. AC-7 의 가드 뮤턴트가 REQ-10 의 규칙이 빠지면 실패하게 해요.
- R-4: 같은 줄을 고치는 다른 카드(영수증을 다시 돌리는 동작을 고치는 카드, 스킬 입구와 `core_rules_2876.rs` 를 고치는 카드)와 병합 충돌이 날 수 있어요. 영수증 한 문장은 `check affected` 의 동작을 말하지 않고 "한 번, 같은 트리에서 다시 안 돌림"만 말해서 앞의 카드와 부딪히지 않아요.
- R-5: 스킬 글이 바뀌면 `source-manifest.json` 의 sha256 과 update 픽스처가 같이 바뀌어요. line-lease 로 쓰고 READY 에 적어요.
- R-6: `agents/impl.md`·`agents/audit.md` 는 기본 에이전트 정의라서 줄 하나만 바꿔도 `agent_defs_package` 시험이 읽어요. 가리키는 줄만 바꾸고 `reads:` 줄은 그대로 둬요.
- R-7: 이 글을 고정한 Python 시험은 원격 CI 에서만 돌아서 깨져도 영수증은 PASS 예요. 그래서 깨짐이 눈에 안 띄지 않게 D-6 의 A 는 `docs/later.md` 한 줄과 Rust 전환 카드의 노트로 남겨요. 사용자가 D-6 의 B 를 고르면 그 함수 삭제는 범위에 `.py` 파일을 넣고 다시 시작해요.

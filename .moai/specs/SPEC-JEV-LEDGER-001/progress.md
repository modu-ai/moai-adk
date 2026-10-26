# progress — SPEC-JEV-LEDGER-001

status: draft

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-26T10:21:34Z
tier: M
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
tree_measured: 954adab25e7d6d83b63f8fb1c8f1aafa859f50c1
card: t1762 (split "a")
note: scope reduced at 0.2.0 by factory-leader ruling (operator-delegated); 11 files, ~770 lines est.; REQ 12/16, AC 11/16 (plan.md §E)

iter2_repair:
  trigger: plan-audit FAIL 0.6875 at audited_sha d1aa4ed656af948c95c2567718c4f9a7be54e191 (.moai/reports/t1762/jev-ledger-plan-audit.md)
  repaired_at: 2026-10-26T10:41:41Z
  defects: D1-D8 (see git history of this file at 0c2d2aa4c2)

iter3_repair:
  trigger: plan-audit FAIL 0.81 at audited_sha 82aff00a6d3b9d1b3dd6f9c2a6fc68ebf5b8e0d0 (.moai/reports/t1762/plan-audit-verdict.md; that path now holds the final verdict at 0c2d2aa4c2)
  version: 0.1.2
  defects: N1-N6 (see git history of this file at 0c2d2aa4c2)

scope_reduction_0_2_0:
  trigger: Tier M audit ceiling reached with FAIL 0.88 at audited_sha 0c2d2aa4c230d2a6671037cefea260ede18e295d (.moai/reports/t1762/plan-audit-verdict.md R1-R4; .moai/reports/t1762/iter-3-adversarial-verdict.md); hold record .moai/reports/t1762/hold-and-split.md
  ruling: factory leader, operator-delegated — split; t1762 keeps M1 (config map, budget constants, template, make build) and M2 (ledger recorder with the regular-file-checked write open)
  migrated: hook adapter, call site, shadow byte identity, deadline anchor and T, M3 measurement (R3), credential wiring (R4) -> SPEC-JEV-LEDGER-002 (card t1970); moai jev cost and its unit label -> SPEC-JEV-LEDGER-003 (card t1971); full id table in spec.md §D.1
  repaired_in_scope: R1 (FIFO fixture and Unix-only flags in a build-tagged test file with a complement-tagged counterpart; DoD adds GOOS=windows go vet); R2 (AC-012 case (iv), FIFO with an open reader, and the handle-check-dropped mutant); O-b (open_other.go complement tag); O-e (ledger pointers re-collected)
  new_ids: REQ-JEVL-017, REQ-JEVL-018, AC-JEVL-017, AC-JEVL-018
  package_decision: recorder in new package internal/jevledger; no file in internal/jev edited (decision-index Q7; card t1968 owns the internal/jev client)
  red_recollected: E1-E12, B1 on 0c2d2aa4c2 (acceptance §L)
  open_gap: the 5 s / 1 s constants remain decisions until -002 measures

reduced_iter1_repair_0_2_1:
  trigger: plan-audit FAIL 0.81 at audited_sha 954adab25e7d6d83b63f8fb1c8f1aafa859f50c1 (.moai/reports/t1762/plan-audit-verdict.md, blocking D1-D4)
  version: 0.2.1
  repaired: D1 (AC-014 forbidden-import list for jev_consumers.go; jevledger stays stdlib-only; plan §B/§D file placement), D2 (threshold in REQ-009 and AC-009), D3 (§D.1 REQ-009 part row), D4 (§D.1 REQ-008 part warning passthrough)
  optional_repaired: D5 (AC-013 binding JevHook constraint), D6 (test-deletion residual risk), D7 (REQ-014 Ubiquitous positive form), D8 (E11/E12 platform-split pointer), D9 (AC-012 case (v) directory)
  ids_unchanged: REQ 12/16, AC 11/16
  red_recollected: E1-E12, B1 on 954adab25e, identical to the 0c2d2aa4c2 record (acceptance §L)

## Binding run conditions

Copied verbatim from .moai/reports/t1762/plan-audit-verdict.md (PASS-WITH-DEBT, audited_sha d6dc623eb803d061225f022598435fce61c1651e) at the plan-to-run Kickoff; the sync auditors re-read them (.claude/rules/moai/workflow/auto-semantics.md §9.1).

- debt: N1 dispose_in=run AC-JEVL-014의 import 검사는 금지 목록 방식이라 `net/rpc`, `net/smtp`, `os/exec`(두 대상 모두), 제3자 HTTP 클라이언트(`jev_consumers.go`)가 통과한다. run 단계 경계 테스트는 AC 문구를 바꾸지 않고 더 엄격하게 짤 수 있다: (a)는 `net`·`net/*`·`os/exec`도 거부, (b)는 "표준 라이브러리 + `gopkg.in/yaml.v3`만 허용"으로 판정하고 그 거부 사례를 변이 탐침으로 기록한다.
- debt: N4 dispose_in=run `JevConsumerConfig`(원시 `yaml.Node` 필드)를 어느 파일에 선언할지 run 단계에서 정하고 progress.md에 기록한다. `types.go`에 두면 `types.go`가 yaml을 import하게 되어 plan §B의 "types.go imports no yaml library" 진술과 어긋난다. 어느 AC도 깨지지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

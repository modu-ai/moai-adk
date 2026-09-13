---
id: SPEC-DECISION-AUTHORITY-001
title: "Acceptance criteria — two-cell matrix, evidence ledger, traceability"
version: "0.1.0"
created: 2026-09-13
---

# acceptance.md — SPEC-DECISION-AUTHORITY-001

Verification discipline: every release-blocking AC carries a **RED-now cell** (single-
invocation read-only command, verbatim stdout, exit code, tree SHA `62fbd6baf` — measured
before any implementation commit) and a **green-path cell** naming the milestone that flips it
(verification-completeness.md §2.1). Coordinates are content-anchored; line numbers in this
file are locating aids at `62fbd6baf` only (§4 pinning). Sweep-based ACs state their
swept-count floor (§1.1 empty-sweep visibility).

## §D AC Matrix

| AC | Asserts | REQ | Class | RED-now | Green path |
|---|---|---|---|---|---|
| AC-DA-001 | `decision_gate` key exists on `InterviewConfig` (yaml tag in `types.go`; default in `defaults.go`) | REQ-DA-001 | release-blocking | P1, P3 | M1 |
| AC-DA-002 | `ResolvedDecisionGate()`: absent/empty/`off`→`off`; `on`→`on`; unknown→`off` — test functions exist and pass | REQ-DA-002, REQ-DA-003 | release-blocking | P17 (swept 0) | M1 — swept-count floor: ≥4 `TestDecisionGate*` functions |
| AC-DA-003 | Template `interview.yaml` ships `decision_gate: off` default; absent-key behavior identical to base `62fbd6baf` | REQ-DA-004 | release-blocking | P4 | M1 |
| AC-DA-004 | Off-mode regression sweep: in the 6-file swept set (3 flow surfaces × local+template), every `decision_gate` occurrence sits inside an enclosing markdown block that carries mode conditioning (`on`-gated / `off`-no-op) | REQ-DA-002, REQ-DA-007 | release-blocking | P5, P12 (0 occurrences — nothing unconditioned exists; the floor is the post-M2 occurrence count ≥ inserted count) | M2 |
| AC-DA-005 | manager-spec body carries the decision-index authoring contract (stateless, gate-on-only) | REQ-DA-005, REQ-DA-006 | release-blocking | P7 (manager-spec row = 0) | M2 |
| AC-DA-006 | Exactly four labels, single vocabulary, present in the swept set (4 local + 3 template files; floor 7) — anchored on the distinctive token `POLICY-COVERED` | REQ-DA-008 | release-blocking | P6 | M2 |
| AC-DA-007 | `DECIDED`/`POLICY-COVERED` rows require a committed file+section authority anchor | REQ-DA-009 | release-blocking | P9 | M2 |
| AC-DA-008 | The authority register is enumerated as committed-only (product.md / completed-SPEC HISTORY / config sections / constitution; untracked reports excluded) | REQ-DA-010 | release-blocking | P14 | M2 |
| AC-DA-009 | Unverifiable anchor → `FOUNDER` (escalate, never downgrade) stated on the flow surfaces | REQ-DA-011 | release-blocking | P10 | M2 |
| AC-DA-010 | spec-assembly carries the kickoff decision-index presentation step (enrichment, fail-open, gate unchanged) | REQ-DA-013 | release-blocking | P7 (spec-assembly row = 0) | M3 |
| AC-DA-011 | Gate-count preserve: `spec-assembly.md` carries exactly the ONE `[HARD] The Implementation Kickoff Approval ... gate stays MANDATORY` clause it carries at base | — (preserve) | regression-guard | P19 (exactly 1 hit at `62fbd6baf`; assert still exactly 1 at close, same clause text) | preserve-through-close |
| AC-DA-012 | plan-auditor unchanged: `decision-index` occurrences in `plan-auditor.md` = 0 at close AND byte-identity to `62fbd6baf` holds (`git diff --stat 62fbd6baf..HEAD -- .claude/agents/moai/plan-auditor.md` prints nothing — P21) | REQ-DA-014 | regression-guard | P7 (auditor row = 0 at base), P21 (empty diff at `6732d1461`) | preserve-through-close |
| AC-DA-013 | Verdict-action vocabulary (`DECIDE` / `NEED_ANALYSIS` / `NEED_EVIDENCE` / `DEFER`) present on the kickoff surface | REQ-DA-015 | release-blocking | P13 | M3 |
| AC-DA-014 | The `product.md` ownership boundary is surfaced as a named kickoff decision (manager-docs owns project-doc scaffolding) | REQ-DA-016 | release-blocking | P15 | M3 |
| AC-DA-015 | Zero-rows ≠ approval clause present on the kickoff surface | REQ-DA-019 | release-blocking | P16 | M3 |
| AC-DA-016 | Index rows carry Detect → Explain → Ask and never an embedded preferred answer (either recommendation mode) | REQ-DA-017 | release-blocking | P11 | M2 |
| AC-DA-017 | Axis orthogonality: no resolver/default/test branch reads both keys; an independence test exists | REQ-DA-018 | release-blocking | P3, P20 | M1 |
| AC-DA-018 | Mirrored-change obligation: every file edited in M2/M3 has the same change in its `internal/template/templates/**` mirror — verified by token presence in BOTH trees per file (not `diff -q`; C1↔C2 branching is intentional) | REQ-DA-020 | release-blocking | P12 | M4 |
| AC-DA-019 | End-to-end off-mode behavioral repro: plan-phase pass with the key absent creates no `decision-index.md` and composes the kickoff unchanged | REQ-DA-002, REQ-DA-007 | regression-guard (no behavioral RED is producible at plan time — at base the flow cannot create an index; undecidable disposition, verification-completeness §2.1) | static cells P5/P12 only (necessary, not sufficient — §D.4); the M5 repro is the completing act | M5 (throwaway SPEC under `/tmp`) |

## §D.1 Severity

- **Critical (release-blocking)**: AC-DA-001..010, 013..018. Any FAIL blocks sync.
- **Guard (regression)**: AC-DA-011, AC-DA-012, AC-DA-019. Green at arrival by design (the
  off-mode behavioral repro cannot be red at plan time — at base the flow cannot create an
  index); verified by re-measurement at close with the pinned tree SHA and the re-measured
  current SHA both recorded.

## §D.2 Swept-count floors

| AC | Swept set | Floor |
|---|---|---|
| AC-DA-002 | `go test -run TestDecisionGate ./internal/config/` | ≥ 4 test functions post-M1; at base the run printed `[no tests to run]` (P17) and asserts nothing |
| AC-DA-004 | manager-spec.md, spec-assembly.md, clarity-interview.md × {local, template} | 6 files swept; every post-M2 occurrence must be block-conditioned |
| AC-DA-006 | 4 local flow files + 3 template mirrors | 7 files swept; `POLICY-COVERED` ≥ 1 in the defining surface and 0 stray second-vocabulary tokens (`URGENT`, `CRITICAL-FLAG`, `BLOCKER` as row labels) |

## §E Evidence Ledger (RED-now cells, measured at `62fbd6baf`)

Carrier: fenced ledger entries cited by id from §D, per verification-completeness §2.1.
All commands single-invocation, read-only, run 2026-09-13 from
`.claude/worktrees/t692`.

```
P1  $ grep -c "decision_gate" internal/config/defaults.go
    0
    exit=1

P2  $ grep -c "recommendation_mode" internal/config/types.go
    1
    exit=0        (positive control — yaml tag at types.go:1360)

P3  $ grep -c "decision_gate" internal/config/types.go
    0
    exit=1

P4  $ grep -c "decision_gate" internal/template/templates/.moai/config/sections/interview.yaml
    0
    exit=1

P5  $ grep -rc "decision_gate" .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    .claude/agents/moai/manager-spec.md:0
    .claude/skills/moai/workflows/plan/clarity-interview.md:0
    .claude/skills/moai/workflows/plan/spec-assembly.md:0
    exit=1

P6  $ grep -rn "POLICY-COVERED" .claude/agents/moai/manager-spec.md .claude/agents/moai/plan-auditor.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    (no output)
    exit=1

P7  $ grep -rn "decision-index" .claude/agents/moai/plan-auditor.md .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md
    (no output)
    exit=1

P8  $ grep -c "recommendation_mode" internal/config/defaults.go
    0
    exit=1        (FALSE positive control — wrong surface: the YAML key lives in
                   types.go as a struct tag; defaults.go carries only the Go field
                   name. Documented in research.md §3 so the same mistake does not
                   recur inside a release-blocking cell.)

P9  $ grep -rn "authority anchor" .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    (no output)
    exit=1

P10 $ grep -rn "Never downgrade" .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    (no output)
    exit=1

P11 $ grep -c "Detect → Explain" .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    .claude/skills/moai/workflows/plan/clarity-interview.md:0
    .claude/agents/moai/manager-spec.md:0
    .claude/skills/moai/workflows/plan/spec-assembly.md:0
    exit=1        (rebuilt at iteration 2 per plan-audit iter1 D2: single invocation,
                   no pipe, paths expanded, distinctive convention token. Measured at
                   6732d1461 — the earlier piped `<...>` form was non-conforming under
                   §2.1(a)/(b) and took the undecidable disposition.)

P12 $ grep -rc "decision_gate" internal/template/templates/.claude/agents/moai/manager-spec.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md
    internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md:0
    internal/template/templates/.claude/agents/moai/manager-spec.md:0
    internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md:0
    exit=1

P13 $ grep -n "NEED_ANALYSIS" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/agents/moai/manager-spec.md
    (no output)
    exit=1

P14 $ grep -rn "authority register" .claude/agents/moai/manager-spec.md .claude/skills/moai/workflows/plan/spec-assembly.md .claude/skills/moai/workflows/plan/clarity-interview.md
    (no output)
    exit=1

P15 $ grep -c "product.md" .claude/skills/moai/workflows/plan/spec-assembly.md
    0
    exit=1

P16 $ grep -in "zero judgment\|zero-flag\|zero flags" .claude/skills/moai/workflows/plan/spec-assembly.md .claude/agents/moai/manager-spec.md
    (no output)
    exit=1

P17 $ go test -run 'TestDecisionGate' ./internal/config/
    ok  	github.com/modu-ai/moai-adk/internal/config	0.394s [no tests to run]
    exit=0        (empty-sweep token: a green with a swept set of zero — asserts nothing, §1.1)

P18 $ grep -c "recommendation_mode" .claude/rules/moai/core/askuser-protocol.md
    4
    exit=0        (positive control — the landed pull-mode conditioning precedent;
                   the axis-under-test greps are measured against a key known present)

P19 $ grep -n "HARD] The Implementation Kickoff Approval" .claude/skills/moai/workflows/plan/spec-assembly.md
    217:[HARD] The Implementation Kickoff Approval `AskUserQuestion` gate stays MANDATORY
    exit=0        (exactly 1 hit at 62fbd6baf — AC-DA-011's preserve baseline)

P20 $ grep -cE "decision_gate.*recommendation_mode|recommendation_mode.*decision_gate" internal/config/types.go
    0
    exit=1        (coupling co-occurrence count in the resolver file — 0 co-occurring
                   lines at 6732d1461; AC-DA-017's ledger carrier. Green path: M1 adds
                   the independence test; this count must STILL be 0 at close — the
                   resolvers never reference each other's key.)

P21 $ git diff --stat 62fbd6baf..HEAD -- .claude/agents/moai/plan-auditor.md
    (no output)
    exit=0        (plan-auditor.md byte-identity vs the authoring baseline holds at
                   6732d1461 — AC-DA-012's preserve instrument per plan-audit iter1 D4;
                   green-at-arrival by design, re-measured at close per §D.5 gate 2.
                   Pinned SHAs: 62fbd6baf = authoring baseline, both ends re-read at
                   each measurement.)
```

## §D.3 Traceability

| REQ | AC(s) |
|---|---|
| REQ-DA-001 | AC-DA-001 |
| REQ-DA-002, REQ-DA-007 | AC-DA-004, AC-DA-019 |
| REQ-DA-003 | AC-DA-002 |
| REQ-DA-004 | AC-DA-003 |
| REQ-DA-005, REQ-DA-006 | AC-DA-005 |
| REQ-DA-008 | AC-DA-006 |
| REQ-DA-009 | AC-DA-007 |
| REQ-DA-010 | AC-DA-008 |
| REQ-DA-011 | AC-DA-009 |
| REQ-DA-012 | AC-DA-004 (flow text), AC-DA-019 (on-mode M5 repro complement) |
| REQ-DA-013 | AC-DA-010 |
| REQ-DA-014 | AC-DA-012 |
| REQ-DA-015 | AC-DA-013 |
| REQ-DA-016 | AC-DA-014 |
| REQ-DA-017 | AC-DA-016 |
| REQ-DA-018 | AC-DA-017 |
| REQ-DA-019 | AC-DA-015 |
| REQ-DA-020 | AC-DA-018 |
| REQ-DA-021 | AC-DA-018 (neutrality pass in M4) + template-neutrality CI guard |

## §D.4 Indirect verification

- AC-DA-019 (behavioral) indirectly verifies REQ-DA-002/007 end-to-end; its static cells
  (P5/P12) are necessary but not sufficient — the M5 repro is the completing act.
- AC-DA-011/012 (preserve) verify REQ-DA-013's "gate unchanged" and REQ-DA-014 negatively —
  the exclusions hold only if the surfaces they protect measure unchanged at close.

## §D.5 Closure gates

1. All Critical ACs PASS with their green-path milestone reached; each flipped RED cell
   re-measured and re-pinned at the closing tree SHA.
2. Both guard ACs re-measured at close: P19 shape still exactly 1; P7 auditor row still 0;
   P21 still an empty diff (`62fbd6baf..close`); AC-DA-019's M5 behavioral repro executed and
   observed.
3. `go test ./internal/config/...` green; coverage ≥ 85% (E3).
4. Template neutrality guard green on all changed template files.

## §D.6 Forward-looking checks

- The auditor-integration deferral (spec.md §E.2): a future card proposing a plan-auditor
  decision-index section must cite the SPEC-JFM §A.3 measurement and re-justify against it.
- The `product.md` reconcile automation (spec.md §F): deferred until OD-2 is settled by the
  operator at kickoff; the settled choice is recorded in the kickoff record.

# SPEC-AUDIT-CEILING-002 — Acceptance Criteria

## §A Discipline and Tree Pin

Every criterion below adopts the two-cell discipline (`.claude/rules/moai/development/verification-completeness.md` §2): a RED-now cell observed on the pre-implementation tree and a GREEN command naming the milestone that flips it. All RED cells were measured in this plan phase on tree **`e497f6936`** (branch `WT-audit-ceiling-guard`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1500`, status clean), from the worktree root. Each RED cell states why it is red: the matched identifiers name code this SPEC will create — no pre-existing file can make them green or keep them red (no wrong-reason red). New-test cells use the package-wide `go test -list` corroboration so a bare selector matching nothing cannot read as a pass.

## §B RED-Now Evidence Ledger

```
LEDGER-ACR-A
  cmd:  grep -rn "CountPlanAuditRounds" internal cmd
  out:  (no output)
  exit: 1
  why:  the counter does not exist yet; the only files that can satisfy this
        grep are the ones M1 creates.
  tree: e497f6936

LEDGER-ACR-B
  cmd:  grep -rnE "RequiredBackendFail|required_backend_fail" internal cmd
  out:  (no output)
  exit: 1
  why:  the admission predicate carries no required-backend signal; M3 creates it.
  tree: e497f6936

LEDGER-ACR-B2
  cmd:  grep -c "required_backend_fail" .moai/docs/audit-artifact-convention.md
  out:  0
  exit: 1
  cmd:  grep -c "required_backend_fail" internal/template/templates/.moai/docs/audit-artifact-convention.md
  out:  0
  exit: 1
  why:  the convention document defines no required-backend line, in either copy;
        M3 adds the same paragraph to both in one change.
  tree: e497f6936

LEDGER-ACR-C
  cmd:  grep -rn "CeilingOutcome" internal cmd
  out:  (no output)
  exit: 1
  cmd:  ls .moai/state/audit-ceiling
  out:  ls: .moai/state/audit-ceiling: No such file or directory
  exit: 1
  why:  no ceiling-outcome type, recording path, or record directory exists;
        M2 creates all three.
  tree: e497f6936

LEDGER-ACR-D
  cmd:  grep -n "this path fails open" internal/cli/mcp_worktree_root.go
  out:  121:// this path fails open, and the loud error belongs to audit_multi and the verb.
  exit: 0
  why:  the fail-open disposition REQ-ACR-006 retires is present in the tree —
        this cell is red because the defect exists, and it flips only when the
        resolver keeps the error distinct.
  tree: e497f6936

LEDGER-ACR-E
  cmd:  go test -list 'TestAdmitRequiredBackend' ./internal/auditverdict
  out:  ok  	github.com/modu-ai/moai-adk/internal/auditverdict	0.302s
  exit: 0
  why:  package-wide corroboration — the selector matches no test in the
        package, so the AC's test is genuinely new, not a pre-existing green.
  tree: e497f6936

LEDGER-ACR-F
  cmd:  grep -rn "PlanAuditTierCeilings" internal/config
  out:  (no output)
  exit: 1
  why:  the harness configuration struct carries no tier-ceiling field; M1 adds it.
  tree: e497f6936

LEDGER-ACR-G
  cmd:  go test -list 'TestCountPlanAuditRounds|TestResolvePlanAuditCeiling|TestRecordCeilingOutcome|TestEvaluatePlanAuditCeiling' ./internal/runtime
  out:  ok  	github.com/modu-ai/moai-adk/internal/runtime	0.262s
  exit: 0
  why:  package-wide corroboration — none of the AC test names exists yet.
  tree: e497f6936

LEDGER-ACR-H
  cmd:  go run ./cmd/moai spec ceiling --help
  out:  the `moai spec` help listing; COMMANDS = status, drift, view, lint,
        close, audit, archive — no `ceiling` entry
  exit: 0
  why:  the parent help exits 0; the red signal is the absent `ceiling` command
        in the listing, which only M2's registration can add.
  tree: e497f6936

LEDGER-ACR-I
  cmd:  go test -list '^(TestHarnessConfigPlanAuditCeilings|TestResolvePlanAuditCeiling)$' ./internal/config
  out:  ok  	github.com/modu-ai/moai-adk/internal/config	0.205s
  exit: 0
  why:  package-wide corroboration — the anchored selector matches no test in
        the package (the harness section has no TestStructYAMLSymmetry_* case;
        a bare `^TestStructYAMLSymmetry$` selector would match nothing and
        read as a pass — this is the enumerated-name guard against that).
  tree: e497f6936
```

## §C Acceptance Criteria

| ID | REQ | Given / When / Then | RED | GREEN (flipping milestone) |
|----|-----|---------------------|-----|----------------------------|
| AC-ACR-001 | REQ-ACR-001 | Given a fixture evidence directory holding `plan-audit.md` and `plan-audit-iter1.md`..`plan-audit-iter3.md`, when the round count is computed, then the count is 4 and nothing is read from memory or session state. | A, G | `go test -run '^TestCountPlanAuditRounds$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-002 | REQ-ACR-001 | Given a fixture directory holding `plan-audit-iterX.md` (suffix not a positive integer), when the round count is computed, then the call returns an error naming the file — never a silent skip. | G | `go test -run '^TestCountPlanAuditRoundsUnparseable$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-003 | REQ-ACR-002 | Given `.moai/config/sections/harness.yaml` as shipped, when the configuration is read through the Go structs, then `PlanAuditTierCeilings` resolves {S:1, M:2, L:3} and `PlanAuditCeilingPolicy` resolves {AutoDeltaRounds: 1, OnFinalHit: "hold-and-split"}; and given a SPEC whose frontmatter carries no `tier:`, the resolved ceiling is the L value. | F, I | `go test -run '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` exit 0 AND `go test -run '^TestResolvePlanAuditCeiling$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-004 | REQ-ACR-003 | Given a round count at or above the ceiling, a latest verdict that is not admitted, and the shipped policy value `hold-and-split`, when the recording path evaluates, then exactly one record is written to `.moai/state/audit-ceiling/<SPEC-ID>.json` with disposition `hold`, the split-proposal reference, and the count/ceiling/label/evidence fields; and `go run ./cmd/moai spec ceiling --help` lists the verb, whose `--record` writes that record; and no interactive input exists anywhere in the path. | C, H | `go test -run '^(TestRecordCeilingOutcome|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` exit 0; `go run ./cmd/moai spec ceiling --help` lists `ceiling`; `grep -c AskUserQuestion internal/runtime/audit_ceiling.go` → 0 and `grep -c AskUserQuestion internal/cli/spec_ceiling.go` → 0 (M2) |
| AC-ACR-005 | REQ-ACR-004 | Given a round count at or above the ceiling and a latest verdict labeled `PASS` passing the full shared admission predicate, when the recording path evaluates, then no record is written and the verdict admits unchanged. | C | `go test -run '^TestEvaluatePlanAuditCeiling$' ./internal/runtime` exit 0, clean-PASS arm asserting the record file does not exist (M2) |
| AC-ACR-006 | REQ-ACR-005 | Given verdict bytes carrying `required_backend_fail: codex`, when `auditverdict.Parse` + `Admit` run, then admission is refused with the backend named — for a `PASS`-labeled verdict, for a `FAIL`-labeled verdict, and unchanged behavior when the line is absent. | B, E | `go test -run '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` exit 0 (M3) |
| AC-ACR-007 | REQ-ACR-005 | Given the convention document, when either copy is read, then § What defines the `required_backend_fail: <backend>` line — its producer (multi-model convergence per-backend verdicts, or the single-backend audit's own review) and its absent-line semantics (absence refuses nothing). | B2 | per-file pair: `grep -c "required_backend_fail" .moai/docs/audit-artifact-convention.md` ≥ 1 AND the template copy ≥ 1, both exit 0 (M3) |
| AC-ACR-008 | REQ-ACR-006 | Given an audit configuration that fails to resolve, when `resolveAuditGates` runs, then the error reaches its caller distinct from an empty not-configured result, and the "this path fails open" comment no longer describes the behavior. | D | `go test -run '^TestResolveAuditGatesConfigErrorDistinct$' ./internal/cli` exit 0; `grep -c "this path fails open" internal/cli/mcp_worktree_root.go` → 0 (M3) |

## §D Traceability (AC → REQ)

| REQ | Covered by | Coverage |
|-----|------------|----------|
| REQ-ACR-001 | AC-ACR-001, AC-ACR-002 | 2 |
| REQ-ACR-002 | AC-ACR-003 | 1 |
| REQ-ACR-003 | AC-ACR-004 | 1 |
| REQ-ACR-004 | AC-ACR-005 | 1 |
| REQ-ACR-005 | AC-ACR-006, AC-ACR-007 | 2 |
| REQ-ACR-006 | AC-ACR-008 | 1 |

6 REQ covered by 8 AC — 100% in both directions (zero UNCOVERED, zero ORPHAN).

## §E Edge Cases

- Evidence directory absent → count 0; 0 is below every ceiling (all configured ceilings ≥ 1), so the ceiling path never fires on an unaudited SPEC.
- `plan-audit.md` coexisting with `iter<N>` files → each file counts (C3); the possible over-count errs toward an early ceiling, the fail-closed direction.
- A `required_backend_fail` line in a sync-audit verdict → refuses (malformed evidence; REQ-ACR-005's unconditional check).
- Repeated `required_backend_fail` lines naming the same backend → both collected; the refusal names the backend; collection does not change the refusal outcome.
- Unknown `on_final_hit` value or unreadable policy → disposition `hold` (fail-closed; REQ-ACR-003's third arm).

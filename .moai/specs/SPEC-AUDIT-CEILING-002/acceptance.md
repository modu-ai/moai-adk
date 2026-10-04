# SPEC-AUDIT-CEILING-002 — Acceptance Criteria

## §A Discipline and Tree Pin

Every criterion below adopts the two-cell discipline (`.claude/rules/moai/development/verification-completeness.md` §2): a RED-now cell observed on the pre-implementation tree and a GREEN command naming the milestone that flips it. Cells A-I were measured in the v0.1.0 plan phase on tree **`e497f6936`**; cells added by the v0.2.0 iter1 repair (J, K) and the re-run narrowed I were measured on tree **`58282d5ac`** — both on branch `WT-audit-ceiling-guard`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1500`, status clean, from the worktree root, and the code paths every cell cites are byte-identical across the two pins (the plan commit touches only `.moai/specs/**`). Each RED cell states why it is red: the matched identifiers name code this SPEC will create — no pre-existing file can make them green or keep them red (no wrong-reason red). New-test cells use the package-wide `go test -list` corroboration so a bare selector matching nothing cannot read as a pass; where a GREEN judgment depends on a swept set, the criterion requires the swept count to be non-empty.

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
  cmd:  go test -list '^TestHarnessConfigPlanAuditCeilings$' ./internal/config
  out:  ok  	github.com/modu-ai/moai-adk/internal/config	0.307s
  exit: 0
  why:  package-wide corroboration — the anchored selector matches no test in
        the package (the harness section has no TestStructYAMLSymmetry_* case;
        a bare `^TestStructYAMLSymmetry$` selector would match nothing and
        read as a pass — this is the enumerated-name guard against that).
        Narrowed in the v0.2.0 repair (D11): TestResolvePlanAuditCeiling is a
        runtime-side name and can never exist in ./internal/config, so it left
        this cell's selector; its absence is covered by LEDGER-ACR-G.
  tree: 58282d5ac (re-run this turn; the v0.1.0 measurement at e497f6936 used
        the wider selector and reproduced the same zero-match fact)

LEDGER-ACR-J
  cmd:  go run ./cmd/moai spec ceiling --help | grep -c ceiling
  out:  0
  exit: 1
  why:  the output-content RED for the M2 exit gate (D1): the verb does not
        exist, so the help listing carries no `ceiling` entry and the deciding
        stage (grep) exits 1 with count 0 — the exit code alone is vacuous
        (the help command itself exits 0 on this unstarted tree), so the gate
        keys on the listing content, and this cell records that same
        command's red value.
  tree: 58282d5ac

LEDGER-ACR-K
  cmd:  go test -list '^(TestRecordCeilingOutcomeDebtProceed|TestRecordCeilingOutcomeUnknownPolicy|TestResolvePlanAuditCeilingInvalid|TestWorktreeRootSurfacesGateError)$' ./internal/runtime ./internal/cli
  out:  ok  	github.com/modu-ai/moai-adk/internal/runtime	0.310s
        ok  	github.com/modu-ai/moai-adk/internal/cli	0.831s
  exit: 0
  why:  package-wide corroboration for the four new AC test names added by
        the v0.2.0 repair (AC-ACR-009/010, AC-ACR-003's invalid-ceiling arm,
        AC-ACR-011) — the anchored selector matches no test in either
        package.
  tree: 58282d5ac
```

## §C Acceptance Criteria

| ID | REQ | Given / When / Then | RED | GREEN (flipping milestone) |
|----|-----|---------------------|-----|----------------------------|
| AC-ACR-001 | REQ-ACR-001 | Given a fixture evidence directory holding `plan-audit.md` and `plan-audit-iter1.md`..`plan-audit-iter3.md`, when the round count is computed, then the count is 4 and nothing is read from memory or session state. | A, G | `go test -run '^TestCountPlanAuditRounds$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-002 | REQ-ACR-001 | Given a fixture directory holding `plan-audit-iterX.md` (suffix not a positive integer), when the round count is computed, then the call returns an error naming the file — never a silent skip. | G | swept-count first: `go test -list '^TestCountPlanAuditRoundsUnparseable$' ./internal/runtime` lists exactly 1 test (a non-empty swept set — a bare selector matching nothing must not read as a pass), THEN `go test -run '^TestCountPlanAuditRoundsUnparseable$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-003 | REQ-ACR-002 | Given `.moai/config/sections/harness.yaml` as shipped, when the configuration is read through the Go structs, then `PlanAuditTierCeilings` resolves {S:1, M:2, L:3} and `PlanAuditCeilingPolicy` resolves {AutoDeltaRounds: 1, OnFinalHit: "hold-and-split"}; and given a SPEC whose frontmatter carries no `tier:`, the resolved ceiling is the L value; and given a ceilings map missing that tier's key — or the L key itself missing or ≤ 0 — the resolution is a configuration error, never a ceiling of zero. | F, I (config side); G (runtime side) | `go test -run '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` exit 0 AND `go test -run '^(TestResolvePlanAuditCeiling\|TestResolvePlanAuditCeilingInvalid)$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-004 | REQ-ACR-003 | Given a round count at or above the ceiling, a latest verdict that is not admitted, and the shipped policy value `hold-and-split`, when the recording path evaluates, then exactly one record is written to `.moai/state/audit-ceiling/<SPEC-ID>.json` with disposition `hold`, the split-proposal reference, and the count/ceiling/label/evidence fields; and the `moai spec ceiling` verb exists with `--record` writing that record; and no interactive input exists anywhere in the path. | C, H, J | `go test -run '^(TestRecordCeilingOutcome\|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` exit 0; `go run ./cmd/moai spec ceiling --help \| grep -c "ceiling"` ≥ 1 (the output-content gate of D1; RED in LEDGER-ACR-J); `grep -c AskUserQuestion internal/runtime/audit_ceiling.go` → 0 and `grep -c AskUserQuestion internal/cli/spec_ceiling.go` → 0 (M2) |
| AC-ACR-005 | REQ-ACR-004 | Given a round count at or above the ceiling and a latest verdict labeled `PASS` passing the full shared admission predicate, when the recording path evaluates, then no record is written and the verdict admits unchanged. | C | `go test -run '^TestEvaluatePlanAuditCeiling$' ./internal/runtime` exit 0, clean-PASS arm asserting the record file does not exist (M2) |
| AC-ACR-006 | REQ-ACR-005 | Given verdict bytes carrying `required_backend_fail: codex`, when `auditverdict.Parse` + `Admit` run, then admission is refused with the backend named — for a `PASS`-labeled verdict, for a `FAIL`-labeled verdict, and unchanged behavior when the line is absent. | B, E | `go test -run '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` exit 0 (M3) |
| AC-ACR-007 | REQ-ACR-005 | Given the convention document, when either copy is read, then § What defines the `required_backend_fail: <backend>` line — its producer (multi-model convergence per-backend verdicts, or the single-backend audit's own review) and its absent-line semantics (absence refuses nothing). | B2 | per-file pair: `grep -c "required_backend_fail" .moai/docs/audit-artifact-convention.md` ≥ 1 AND the template copy ≥ 1, both exit 0 (M3) |
| AC-ACR-008 | REQ-ACR-006 | Given an audit configuration that errors — (a) a workflow.yaml that cannot be read or parsed (the pins-loader error class) and (b) a value the audit-plan resolver rejects (the resolver error class) — when the resolution path runs, then each surface keeps its error distinct from an empty not-configured result (`workflowAuditPins` never folds to a zero configuration; `resolveAuditGates` never folds to an empty gate set), and neither the "this path fails open" comment nor the "(N3)" fold comment describes the behavior any longer. | D, K | `go test -run '^(TestResolveAuditGatesConfigErrorDistinct\|TestWorkflowAuditPinsErrorNotFolded)$' ./internal/cli` exit 0; `grep -c "this path fails open" internal/cli/mcp_worktree_root.go` → 0; `grep -c "(N3)" internal/cli/audit_pin.go` → 0 (M3) |
| AC-ACR-009 | REQ-ACR-003 | Given a round count at or above the ceiling and a latest verdict labeled `PASS-WITH-DEBT` passing the full shared admission predicate with at least one well-formed debt, when the recording path evaluates, then the record's disposition is `debt-proceed` and the record references the verdict's debt ids; run entry proceeds. | C, K | `go test -run '^TestRecordCeilingOutcomeDebtProceed$' ./internal/runtime` exit 0 (M2) |
| AC-ACR-010 | REQ-ACR-003 | Given a round count at or above the ceiling, a latest verdict that is not admitted, and a policy value that is neither `hold-and-split` nor `split` — or an unreadable policy — when the recording path evaluates, then the record's disposition is `hold` with no split-proposal reference (fail-closed). | C, K | `go test -run '^TestRecordCeilingOutcomeUnknownPolicy$' ./internal/runtime` exit 0 (M2) |
| AC-ACR-011 | REQ-ACR-006 | Given the resolution path erroring (either error class of AC-ACR-008), when the MCP tool surface handles the result, then the tool result reports the error and never an empty "not configured" gate set. | D, K | `go test -run '^TestWorktreeRootSurfacesGateError$' ./internal/cli` exit 0 (M3) |

## §D Traceability (AC → REQ)

| REQ | Covered by | Coverage |
|-----|------------|----------|
| REQ-ACR-001 | AC-ACR-001, AC-ACR-002 | 2 |
| REQ-ACR-002 | AC-ACR-003 | 1 (three arms: shipped binding, unknown-tier→L, missing/non-positive→error) |
| REQ-ACR-003 | AC-ACR-004, AC-ACR-009, AC-ACR-010 | 3 (hold-and-split, debt-proceed, unknown/unreadable→hold; the clean-PASS exclusion is REQ-ACR-004's own AC-ACR-005) |
| REQ-ACR-004 | AC-ACR-005 | 1 |
| REQ-ACR-005 | AC-ACR-006, AC-ACR-007 | 2 |
| REQ-ACR-006 | AC-ACR-008, AC-ACR-011 | 2 (both error classes distinct; caller surfacing) |

6 REQ covered by 11 AC — 100% in both directions at REQ and clause level (zero UNCOVERED, zero ORPHAN).

## §E Edge Cases

- Evidence directory absent → count 0; 0 is below every resolved ceiling, and a ceiling that resolves missing or non-positive is a configuration error (REQ-ACR-002), so the ceiling path never fires on an unaudited SPEC or on a mis-configured ceilings map.
- `plan-audit.md` coexisting with `iter<N>` files → each file counts (C3); the possible over-count errs toward an early ceiling, the fail-closed direction.
- A `required_backend_fail` line in a sync-audit verdict → refuses (malformed evidence; REQ-ACR-005's unconditional check).
- Repeated `required_backend_fail` lines naming the same backend → both collected; the refusal names the backend; collection does not change the refusal outcome.
- Unknown `on_final_hit` value or unreadable policy → disposition `hold` (fail-closed; REQ-ACR-003's third arm).

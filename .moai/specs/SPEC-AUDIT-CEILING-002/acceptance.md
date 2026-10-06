# SPEC-AUDIT-CEILING-002 — Acceptance Criteria

## §A Discipline and Tree Pin

Every criterion below adopts the two-cell discipline (`.claude/rules/moai/development/verification-completeness.md` §2): a RED-now cell observed on the pre-implementation tree and a GREEN command naming the milestone that flips it. Cells carry per-cell tree pins — A-I measured in the v0.1.0 plan phase on **`e497f6936`**; J, K, and the re-run narrowed I on **`58282d5ac`** (v0.2.0 iter1 repair); L, M, N, O on **`e88493d3a`** (v0.3.0 iter2 repair); N re-formed and P, Q on **`9bd0d3ba3`** (v0.4.0 PASS-with-debt condition repair) — all on branch `WT-audit-ceiling-guard`, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1500`, status clean, from the worktree root, and the code paths every cell cites are byte-identical across the pins (every plan commit touches only `.moai/specs/**`). Each RED cell states why it is red: the matched identifiers name code this SPEC will create, or — for the defect-presence cells (D, L) — the defect exists; no wrong-reason red. New-test cells use the package-wide `go test -list` corroboration so a bare selector matching nothing cannot read as a pass, and EVERY test-based GREEN judgment in §C requires its swept count first — the `-list` step lists exactly the named tests (non-empty, exact count), THEN the `-run` exits 0 (the N4 rule; AC-ACR-002 is the template). Form note (N9): commands inside table cells escape pipes as `\|`; the ledger in §B and plan.md carry the unescaped canonical forms — copy from there, not from a table cell. The full-document accuracy sweep record (counts, quotes, pointers, cross-references re-verified) is progress.md §K.

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
  out:    COMMANDS
            status <SPEC-ID> <new-status> [--flags]                          Update or list SPEC document statuses
            drift [--flags]                                                  Detect SPEC status drift between frontmatter and git log
            view <SPEC-ID> [--flags]                                         View acceptance criteria in tree structure
            lint [SPEC-ID | path/to/spec.md | SPEC directory ...] [--flags]  Lint SPEC documents for EARS compliance and structural validity
            close SPEC-ID [--flags]                                          Atomic 4-phase close for a SPEC (spec.md status: completed + progress.md backfill)
            audit [--flags]                                                  Audit SPEC era classification and detect modern-era status drift
            archive [--flags]                                                Archive finished SPECs out of .moai/specs/
        (the COMMANDS block verbatim — no `ceiling` entry; the header/flags
        lines of the listing omitted here as they carry no command names)
  exit: 0
  why:  the parent help exits 0; the red signal is the absent `ceiling` command
        in the listing, which only M2's registration can add. Out field made
        raw in the v0.4.0 repair (R6) — the decisive lines verbatim instead
        of a prose description.
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
        command's red value. Form note (R6): the command is a pipe, outside
        verification-completeness §2.1's single-invocation letter — the N8
        judgment-held disposition stands per the leader's direction (four
        elements present, re-executing and reproducing every round), with
        codex's contrary reading recorded in the iter2/iter3 verdicts.
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

LEDGER-ACR-L
  cmd:  grep -c "(N3)" internal/cli/audit_pin.go
  out:  2
  exit: 0
  why:  defect-presence cell (the N3 fix's RED half for AC-ACR-008's pins
        side): the "(N3)" fold documentation the repair retires is present
        in the tree — red because the defect exists, flipping only when the
        fold is gone (AC-ACR-008's GREEN expects this count to reach 0).
  tree: e88493d3a

LEDGER-ACR-M
  cmd:  go test -list '^(TestWorkflowAuditPinsErrorNotFolded|TestResolveAuditGatesConfigErrorDistinct)$' ./internal/cli
  out:  ok  	github.com/modu-ai/moai-adk/internal/cli	0.805s
  exit: 0
  why:  package-wide corroboration for AC-ACR-008's own test names (the
        pins-loader and resolver error-distinct tests) — the anchored
        selector matches no test in the package (N3's second cell).
  tree: e88493d3a

LEDGER-ACR-N
  cmd:  grep -rn --include="*.go" --exclude="*_test.go" "workflowAuditPins(\|resolveAuditGates(" internal/cli
  out:  internal/cli/mcp_worktree_root.go:103:		return workflowAuditPins(root), false
        internal/cli/mcp_worktree_root.go:109:	return workflowAuditPins(primary), false
        internal/cli/mcp_worktree_root.go:122:func resolveAuditGates(root string) (gates config.AuditGates, assumedNote string) {
        internal/cli/mcp_convergence.go:1007:	return resolveAuditGates(root)
        internal/cli/mcp_claude.go:187:	pin := workflowAuditPins(root).Claude
        internal/cli/codex_audit_launch.go:202:	effort := workflowAuditPins(root).Codex.Effort
        internal/cli/mcp_glm.go:219:	if pin := workflowAuditPins(root).GLM; pin.Model != "" {
        internal/cli/mcp_codex.go:221:	if pin := workflowAuditPins(cwd).Codex; pin.Model != "" && codexServableModel(pin.Model) {
        internal/cli/mcp_codex.go:2050:	gates, assumedNote := resolveAuditGates(projectDir)
        internal/cli/audit_pin.go:58:func workflowAuditPins(projectDir string) config.AuditConfig {
  exit: 0
  why:  the caller-inventory baseline for plan file 18 and REQ-ACR-006's
        propagation clause (N6): six caller sites outside the two repaired
        surfaces, matching the iter2 verdict's grep-verified list; a
        re-run at implementation time that shows MORE sites means the
        inventory row must be extended before the work is done. Re-formed
        in the v0.4.0 repair (R6) as a single invocation — the former
        `| grep -v _test` pipe replaced by `--exclude="*_test.go"` — with
        the raw output lines recorded verbatim instead of a digest; the
        re-formed run reproduced the same ten lines.
  tree: e88493d3a (original form); 9bd0d3ba3 (re-formed — the lines above)

LEDGER-ACR-O
  cmd:  go test -list '^(TestRecordCeilingOutcomeSplitValue|TestLatestVerdictTieIsError)$' ./internal/runtime
  out:  ok  	github.com/modu-ai/moai-adk/internal/runtime	0.249s
  exit: 0
  why:  package-wide corroboration for the two new AC test names added by
        the v0.3.0 repair (AC-ACR-012's split arm; AC-ACR-001's tie arm) —
        the anchored selector matches no test in the package.
  tree: e88493d3a

LEDGER-ACR-P
  cmd:  go test -list '^(TestSpecCeilingRecordWritesJSON|TestGateErrorPropagatesToCallerSites)$' ./internal/cli
  out:  ok  	github.com/modu-ai/moai-adk/internal/cli	0.803s
  exit: 0
  why:  package-wide corroboration for the two new AC test names added by
        the v0.4.0 repair (AC-ACR-004's CLI-level --record arm, R1;
        AC-ACR-011's per-site propagation arm, R4) — the anchored selector
        matches no test in the package.
  tree: 9bd0d3ba3

LEDGER-ACR-Q
  cmd:  go test -list '^TestCountPlanAuditRoundsListedDirMissing$' ./internal/runtime
  out:  ok  	github.com/modu-ai/moai-adk/internal/runtime	0.240s
  exit: 0
  why:  package-wide corroboration for AC-ACR-013's test name (the
        listed-absent arm, R5) — the anchored selector matches no test in
        the package.
  tree: 9bd0d3ba3
```

## §C Acceptance Criteria

| ID | REQ | Given / When / Then | RED | GREEN (flipping milestone; every test-based judgment is swept-count-first per §A) |
|----|-----|---------------------|-----|----------------------------|
| AC-ACR-001 | REQ-ACR-001 | Given a fixture evidence directory holding `plan-audit.md` and `plan-audit-iter1.md`..`plan-audit-iter3.md`, when the round count is computed, then the count is 4 and nothing is read from memory or session state; and given the same highest iteration number present in two evidence directories, the latest-verdict selection is an error rather than a silent pick. | A, G, O | `go test -list '^(TestCountPlanAuditRounds\|TestLatestVerdictTieIsError)$' ./internal/runtime` lists exactly 2, THEN `go test -run '^(TestCountPlanAuditRounds\|TestLatestVerdictTieIsError)$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-002 | REQ-ACR-001 | Given a fixture directory holding `plan-audit-iterX.md` (suffix not a positive integer), when the round count is computed, then the call returns an error naming the file — never a silent skip. | G | swept-count first: `go test -list '^TestCountPlanAuditRoundsUnparseable$' ./internal/runtime` lists exactly 1 test, THEN `go test -run '^TestCountPlanAuditRoundsUnparseable$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-003 | REQ-ACR-002 | Given `.moai/config/sections/harness.yaml` as shipped, when the configuration is read through the Go structs, then `PlanAuditTierCeilings` resolves {S:1, M:2, L:3} and `PlanAuditCeilingPolicy` resolves {AutoDeltaRounds: 1, OnFinalHit: "hold-and-split"}; and given a SPEC whose frontmatter carries no `tier:`, the resolved ceiling is the L value; and given a ceilings map missing that tier's key — or the L key itself missing or ≤ 0 — the resolution is a configuration error, never a ceiling of zero. | F, I (config side); G (resolve-name), K (invalid-name) (runtime side) | `go test -list '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` lists exactly 1, THEN `go test -run '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` exit 0; AND `go test -list '^(TestResolvePlanAuditCeiling\|TestResolvePlanAuditCeilingInvalid)$' ./internal/runtime` lists exactly 2, THEN `go test -run '^(TestResolvePlanAuditCeiling\|TestResolvePlanAuditCeilingInvalid)$' ./internal/runtime` exit 0 (M1) |
| AC-ACR-004 | REQ-ACR-003 | Given a round count at or above the ceiling, a latest verdict (per REQ-ACR-001's selection rule) that is not admitted, and the shipped policy value `hold-and-split`, when the recording path evaluates, then exactly one record is written to `.moai/state/audit-ceiling/<SPEC-ID>.json` with disposition `hold`, the split-proposal reference, and the count/ceiling/label/evidence fields; and the `moai spec ceiling` verb exists and `--record` writes that record — observed at the CLI level by driving the compiled verb with `--record` against a temp project and asserting the written JSON (R1); and no interactive input exists anywhere in the path. | C, G, H, J | `go test -list '^(TestRecordCeilingOutcome\|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` lists exactly 2, THEN `go test -run '^(TestRecordCeilingOutcome\|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` exit 0; `go run ./cmd/moai spec ceiling --help \| grep -c "ceiling"` ≥ 1 (the output-content gate of D1; RED in LEDGER-ACR-J); `go test -list '^TestSpecCeilingRecordWritesJSON$' ./internal/cli` lists exactly 1, THEN `go test -run '^TestSpecCeilingRecordWritesJSON$' ./internal/cli` exit 0 (the R1 CLI-level arm; RED in LEDGER-ACR-P); `grep -c AskUserQuestion internal/runtime/audit_ceiling.go` → 0 and `grep -c AskUserQuestion internal/cli/spec_ceiling.go` → 0 (M2) |
| AC-ACR-005 | REQ-ACR-004 | Given a round count at or above the ceiling and the selected latest verdict labeled `PASS` passing the full shared admission predicate, when the recording path evaluates, then no record is written and the verdict admits unchanged; and given an invalid resolved ceiling (missing or non-positive), the evaluation propagates the configuration error and writes no record (the Evaluate-level arm of R2 — the resolver-level arm is AC-ACR-003's). | C, G | `go test -list '^TestEvaluatePlanAuditCeiling$' ./internal/runtime` lists exactly 1, THEN `go test -run '^TestEvaluatePlanAuditCeiling$' ./internal/runtime` exit 0, clean-PASS arm asserting the record file does not exist plus the R2 invalid-ceiling arm asserting the propagated error (M2) |
| AC-ACR-006 | REQ-ACR-005 | Given verdict bytes carrying `required_backend_fail: codex`, when `auditverdict.Parse` + `Admit` run, then admission is refused with the backend named — for a `PASS`-labeled verdict, for a `FAIL`-labeled verdict, and unchanged behavior when the line is absent. | B, E | `go test -list '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` lists exactly 1, THEN `go test -run '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` exit 0 (M3) |
| AC-ACR-007 | REQ-ACR-005 | Given the convention document, when either copy is read, then § What defines the `required_backend_fail: <backend>` line — its producer (multi-model convergence per-backend verdicts, or the single-backend audit's own review) and its absent-line semantics (absence refuses nothing). | B2 | per-file pair: `grep -c "required_backend_fail" .moai/docs/audit-artifact-convention.md` ≥ 1 AND the template copy ≥ 1, both exit 0 (M3) |
| AC-ACR-008 | REQ-ACR-006 | Given an audit configuration that errors — (a) a workflow.yaml that cannot be read or parsed (the pins-loader error class) and (b) a value the audit-plan resolver rejects (the resolver error class) — when the resolution path runs, then each surface keeps its error distinct from an empty not-configured result (`workflowAuditPins` never folds to a zero configuration; `resolveAuditGates` never folds to an empty gate set), and neither the "this path fails open" comment nor the "(N3)" fold comment describes the behavior any longer. | D, L, M (D: the resolver fail-open comment present; L: the "(N3)" fold comment present — count 2; M: both tests absent) | `go test -list '^(TestResolveAuditGatesConfigErrorDistinct\|TestWorkflowAuditPinsErrorNotFolded)$' ./internal/cli` lists exactly 2, THEN `go test -run '^(TestResolveAuditGatesConfigErrorDistinct\|TestWorkflowAuditPinsErrorNotFolded)$' ./internal/cli` exit 0; `grep -c "this path fails open" internal/cli/mcp_worktree_root.go` → 0; `grep -c "(N3)" internal/cli/audit_pin.go` → 0 (M3) |
| AC-ACR-009 | REQ-ACR-003 | Given a round count at or above the ceiling and the selected latest verdict labeled `PASS-WITH-DEBT` passing the full shared admission predicate with at least one well-formed debt, when the recording path evaluates, then the record's disposition is `debt-proceed` and the record references the verdict's debt ids; run entry proceeds. | C, K | `go test -list '^TestRecordCeilingOutcomeDebtProceed$' ./internal/runtime` lists exactly 1, THEN `go test -run '^TestRecordCeilingOutcomeDebtProceed$' ./internal/runtime` exit 0 (M2) |
| AC-ACR-010 | REQ-ACR-003 | Given a round count at or above the ceiling, a selected latest verdict that is not admitted, and a policy value that is neither `hold-and-split` nor `split` — or an unreadable policy — when the recording path evaluates, then the record's disposition is `hold` with no split-proposal reference (fail-closed). | C, K | `go test -list '^TestRecordCeilingOutcomeUnknownPolicy$' ./internal/runtime` lists exactly 1, THEN `go test -run '^TestRecordCeilingOutcomeUnknownPolicy$' ./internal/runtime` exit 0 (M2) |
| AC-ACR-011 | REQ-ACR-006 | Given the resolution path erroring (either error class of AC-ACR-008), when a consuming surface handles the result — the `worktree_root` tool AND each of the six inventoried caller sites of plan file 18 (R4: the propagation claim is verified per site, not by one surface) — then the result reports the error and never an absent configuration. | D, K, M (D: the resolver fail-open comment present; K: SurfacesGateError absent; M: the pins/resolver pair absent) | `go test -list '^(TestGateErrorPropagatesToCallerSites\|TestWorktreeRootSurfacesGateError)$' ./internal/cli` lists exactly 2, THEN `go test -run '^(TestGateErrorPropagatesToCallerSites\|TestWorktreeRootSurfacesGateError)$' ./internal/cli` exit 0 — the propagation test is table-driven with one subtest per inventoried site, and a no-discarded-error static check `grep -rnE "(_ =|= _) ?(workflowAuditPins|resolveAuditGates)\(" internal/cli --include="*.go"` matches nothing (M3) |
| AC-ACR-012 | REQ-ACR-003 | Given a round count at or above the ceiling, a selected latest verdict that is not admitted, and a policy value of exactly `split`, when the recording path evaluates, then the record's disposition is `split` (the D2 mapping's selecting arm; no shipped configuration carries the value today). | C, O | swept-count first: `go test -list '^TestRecordCeilingOutcomeSplitValue$' ./internal/runtime` lists exactly 1, THEN `go test -run '^TestRecordCeilingOutcomeSplitValue$' ./internal/runtime` exit 0 (M2) |
| AC-ACR-013 | REQ-ACR-001 | Given an evidence-directory invocation whose EXPLICITLY LISTED directory does not exist (a typo), when the round count is computed, then the call returns an error naming the missing path — while the SPEC-scoped `.moai/reports/<SPEC-ID>/` directory being absent still contributes 0 (an unaudited SPEC is not an error) (R5; the input is split `specDir, listedDirs` so the two classes cannot be confused). | Q | swept-count first: `go test -list '^TestCountPlanAuditRoundsListedDirMissing$' ./internal/runtime` lists exactly 1, THEN `go test -run '^TestCountPlanAuditRoundsListedDirMissing$' ./internal/runtime` exit 0 (M1) |

## §D Traceability (AC → REQ)

| REQ | Covered by | Coverage |
|-----|------------|----------|
| REQ-ACR-001 | AC-ACR-001, AC-ACR-002, AC-ACR-013 | 3 (count + tie arm; unparseable; the listed-absent arm) |
| REQ-ACR-002 | AC-ACR-003 | 1 (three arms: shipped binding, unknown-tier→L, missing/non-positive→error) |
| REQ-ACR-003 | AC-ACR-004, AC-ACR-009, AC-ACR-010, AC-ACR-012 | 4 — every mapping clause of REQ-ACR-003 carries its own criterion: hold-and-split (004), debt-proceed (009), unknown/unreadable→hold (010), split (012); the clean-PASS exclusion is REQ-ACR-004's own AC-ACR-005 |
| REQ-ACR-004 | AC-ACR-005 | 1 (clean-PASS arm + the Evaluate-level invalid-ceiling arm of R2) |
| REQ-ACR-005 | AC-ACR-006, AC-ACR-007 | 2 |
| REQ-ACR-006 | AC-ACR-008, AC-ACR-011 | 2 (both error classes distinct; caller surfacing with per-site propagation subtests) |

6 REQ covered by 13 AC — 100% in both directions at REQ level (zero UNCOVERED, zero ORPHAN), and at clause level every disposition arm of REQ-ACR-002/003/006 carries a criterion, with REQ-ACR-006's six-site propagation claim carried per site (AC-ACR-011's table-driven subtests) rather than descoped. Two surfaces remain explicitly DESCOPED from criterion-level coverage rather than silently uncovered: the registry-disposition note rewrite (plan file 4 — a describing-surface text edit carried by AC-ACR-003's config binding on the behavior side) and the CLI verb's no-flag read output shape (the verb's existence and `--record` write are AC-ACR-004's gates; the read listing's exact fields are implementation detail).

## §E Edge Cases

- Evidence directory absent → the SPEC-scoped `.moai/reports/<SPEC-ID>/` directory being absent contributes 0 (an unaudited SPEC is not an error); an EXPLICITLY LISTED evidence directory that does not exist is an error (N10 — a typo must not silently count 0); 0 is below every resolved ceiling, and a ceiling that resolves missing or non-positive is a configuration error (REQ-ACR-002), so the ceiling path never fires on an unaudited SPEC or on a mis-configured ceilings map.
- `plan-audit.md` coexisting with `iter<N>` files → each file counts (C3); the possible over-count errs toward an early ceiling, the fail-closed direction.
- A `required_backend_fail` line in a sync-audit verdict → refuses (malformed evidence; REQ-ACR-005's unconditional check).
- Repeated `required_backend_fail` lines naming the same backend → both collected; the refusal names the backend; collection does not change the refusal outcome.
- Unknown `on_final_hit` value or unreadable policy → disposition `hold` (fail-closed; REQ-ACR-003's third arm).
- A directory listed twice in one invocation → its rounds contribute once (identical listed paths are deduplicated before counting — `plan-audit-iter1.md` seen through a duplicate listing is still one file one round; the alternative of double-counting would violate REQ-ACR-001's one-file-one-round rule) (R6).
- Two files parsing to the same iteration number within ONE directory → impossible by construction: within a single directory the recorded file names are unique (`plan-audit.md` ranks 0 and each `plan-audit-iter<N>.md` carries its N in the name), so parsed N values collide only ACROSS directories, which is REQ-ACR-001's selection error (R6).

# SPEC-AUDIT-CEILING-002 — Plan

> Tier M. 6 REQ / 12 AC (v0.3.0 — iter2 repair added AC-ACR-012 and the tie arm in AC-ACR-001). Every code anchor below was measured on this tree at HEAD `e497f6936`, re-measured at `58282d5ac` and `e88493d3a` for the repair-round cells (branch `WT-audit-ceiling-guard`, clean). Line numbers are dated pointers — re-verify at run start, never trust across absorbs.

## §A Approach Summary

Three work items land as four milestones. Work items 1+2 share the new file `internal/runtime/audit_ceiling.go` (counter, ceiling resolution, one outcome-recording path) surfaced through one new CLI verb `moai spec ceiling` (`internal/cli/spec_ceiling.go`, registered in `internal/cli/spec.go`). Work item 3 extends the shared admission predicate (`internal/auditverdict`) with one parsed signal and one refusal arm, corrects the audit-gates resolver's fail-open, and adds the line format to the convention document (template + deployed in one change). Configuration enters through the existing harness section path (`internal/config`), additive only (C4). No prose gate texts are edited (C1).

## §B File Map

| # | File | Action | What |
|---|------|--------|------|
| 1 | `internal/config/types.go` | edit | `HarnessConfig` (struct at :1328-1353) gains `PlanAuditTierCeilings map[string]int` (yaml `plan_audit_tier_ceilings`) and `PlanAuditCeilingPolicy PlanAuditCeilingPolicyConfig` (yaml `plan_audit_ceiling_policy`), beside `PlanAuditGlobal` (:1349-1350) — the `plan_audit` member at :1401 belongs to `LevelConfig`, not `HarnessConfig` (measured; D16); new struct `PlanAuditCeilingPolicyConfig { AutoDeltaRounds int; OnFinalHit string }` |
| 2 | `internal/config/defaults.go` | edit | defaults for both: ceilings {S:1, M:2, L:3}; policy {AutoDeltaRounds: 1, OnFinalHit: "hold-and-split"} — mirroring `.moai/config/sections/harness.yaml:75-84` verbatim |
| 3 | `internal/config/loader_harness_extended_test.go` | edit | new `TestHarnessConfigPlanAuditCeilings`: the two fields bind the shipped yaml values and the defaults apply on an absent file (the harness section is a dedicated loader entry point outside `Loader.Load()` — measured, no `TestStructYAMLSymmetry_*` case covers it) |
| 4 | `internal/config/audit_registry.go` | edit | the two keys' no-Go-reader disposition retires to Go-read |
| 5 | `internal/runtime/audit_ceiling.go` | new | `CountPlanAuditRounds(evidenceDirs []string) (int, error)`; `ResolvePlanAuditCeiling(tier string, ceilings map[string]int) (int, error)` — the error carries REQ-ACR-002's configuration-error cases (missing or non-positive resolved ceiling) and `EvaluatePlanAuditCeiling` propagates it BEFORE any count/ceiling comparison; `SelectLatestVerdict` per REQ-ACR-001's rule (highest parsed N, `plan-audit.md` = 0, cross-directory tie = error); `CeilingOutcome` (Disposition, Count, Ceiling, VerdictLabel, EvidencePaths, SplitProposalRef); `RecordCeilingOutcome(specID string, outcome CeilingOutcome) error` writing `.moai/state/audit-ceiling/<SPEC-ID>.json`; the single `EvaluatePlanAuditCeiling` entry composing select → count → ceiling → disposition (REQ-ACR-003's selection order, REQ-ACR-004's clean-PASS no-record arm) |
| 6 | `internal/runtime/audit_ceiling_test.go` | new | the runtime AC tests: AC-ACR-001 (count + tie arm `TestLatestVerdictTieIsError`), AC-ACR-002 (`TestCountPlanAuditRoundsUnparseable`), AC-ACR-003's runtime arms (`TestResolvePlanAuditCeiling`, `TestResolvePlanAuditCeilingInvalid`), AC-ACR-004/005 (`TestRecordCeilingOutcome`, `TestEvaluatePlanAuditCeiling`), AC-ACR-009/010/012 (`TestRecordCeilingOutcomeDebtProceed`, `TestRecordCeilingOutcomeUnknownPolicy`, `TestRecordCeilingOutcomeSplitValue`) |
| 7 | `internal/cli/spec_ceiling.go` | new | `moai spec ceiling <SPEC-ID> [--evidence <dir>]... [--record]` — read-only by default; `--record` invokes the one recording path. No interactive input of any kind (C-HRA-008: CLI code never prompts) |
| 8 | `internal/cli/spec.go` | edit | register `newSpecCeilingCmd()` beside `newSpecLintCmd()` (:26) |
| 9 | `internal/auditverdict/verdict.go` | edit | `Fields` gains `RequiredBackendFails []string`; `Parse` (:69) collects `required_backend_fail: <name>` lines; `Admit` (:195) refuses when the set is non-empty — placed immediately after the duplicate-keys check (:196-198), before the label check, so the refusal fires regardless of label (REQ-ACR-005) |
| 10 | `internal/auditverdict/verdict_test.go` | edit | `TestAdmitRequiredBackendFail` (label PASS + line → refuse; label FAIL + line → refuse; no line → unchanged) |
| 11 | `.moai/docs/audit-artifact-convention.md` | edit | § What gains the `required_backend_fail: <backend>` line: produced by the exporting auditor from the convergence result's per-backend verdicts or its own single-backend review; absent line refuses nothing |
| 12 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | edit | same paragraph (C2 mirror discipline — same change) |
| 13 | `internal/cli/mcp_worktree_root.go` | edit | `resolveAuditGates` (:122-132) keeps a resolution error distinct from the empty result; the callers at the tool surface report the error instead of an empty "not configured" set |
| 14 | `internal/cli/audit_pin.go` | edit | `workflowAuditPins` (:58-63) keeps a load error distinct from the zero configuration — the documented "(N3)" fold (comment at :54-57) retires; both error classes the loader produces (`read workflow.yaml` :38-40, `parse workflow.yaml` :47-49) propagate to `auditSectionForRoot` (:109) |
| 15 | `internal/cli/audit_pin_test.go` (or sibling) | edit | `TestWorkflowAuditPinsErrorNotFolded` — an unreadable and an unparseable workflow.yaml each return an error, while an absent file still reads as absent (the legitimate not-configured case keeps its meaning) |
| 16 | `internal/cli/mcp_worktree_root_test.go` (or sibling) | edit | `TestResolveAuditGatesConfigErrorDistinct` + `TestWorktreeRootSurfacesGateError` (the caller-surfacing half, AC-ACR-011) |
| 17 | `.moai/config/sections/harness.yaml` + `internal/template/templates/.moai/config/sections/harness.yaml` | edit | the keys' note (lines 69-84) gains one line naming the Go reader added by this SPEC — describing-surface currency, both copies in one change (C2) |
| 18 | caller inventory (N6, no single file) | edit | the six caller sites of the two resolution surfaces each propagate the new error instead of assuming an absent configuration — `workflowAuditPins`: `mcp_claude.go:187`, `codex_audit_launch.go:202`, `mcp_glm.go:219`, `mcp_codex.go:221` (plus the in-process `mcp_worktree_root.go:103/:109`); `resolveAuditGates`: `mcp_convergence.go:1007`, `mcp_codex.go:2050`; baseline pinned as LEDGER-ACR-N (`grep -rn "workflowAuditPins(\|resolveAuditGates(" internal/cli` minus tests), re-verified at run start — absorbs before round 3 can add callers, and the inventory cell makes that visible |

PRESERVE (untouched): `internal/runtime/audit_gate.go`, `internal/runtime/audit_review.go`, `internal/auditverdict/Admit`'s existing check order after the new arm, both `plan-auditor` agent bodies, all `run/phase-execution.md` / `auto-semantics.md` gate texts, `.moai/reports/**` (read-only measurement surface), the old card tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500` (read-only reference).

## §C Milestones

### M1 — configuration + counter (work item 1)

Files 1-6. Additive config fields with defaults + the harness-loader test `TestHarnessConfigPlanAuditCeilings` (file 3 — the harness section's own test surface) + registry disposition update; `CountPlanAuditRounds` + `ResolvePlanAuditCeiling` + `SelectLatestVerdict` with table tests (fixture dirs: `plan-audit.md` + `iter1..3` → 4; empty/absent dir → 0; tier absent → L ceiling; a tier key missing from the map or resolving ≤ 0 → configuration error; the same highest iteration number in two evidence directories → selection error), plus `TestCountPlanAuditRoundsUnparseable` as its own test for the unparseable-suffix case.

**Exit (count-first, N4):** `go test -list '^(TestCountPlanAuditRounds|TestCountPlanAuditRoundsUnparseable|TestLatestVerdictTieIsError|TestResolvePlanAuditCeiling|TestResolvePlanAuditCeilingInvalid)$' ./internal/runtime` lists exactly 5 tests, THEN `go test -run '^(TestCountPlanAuditRounds|TestCountPlanAuditRoundsUnparseable|TestLatestVerdictTieIsError|TestResolvePlanAuditCeiling|TestResolvePlanAuditCeilingInvalid)$' ./internal/runtime` exit 0; `go test -list '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` lists exactly 1, THEN `go test -run '^TestHarnessConfigPlanAuditCeilings$' ./internal/config` exit 0; `go build ./...` exit 0.

### M2 — one recording path + CLI verb (work item 2)

Files 5 (outcome half), 7, 8. `CeilingOutcome` + `RecordCeilingOutcome` + the single evaluation entry; CLI verb read + `--record`. Disposition selection order in code matches REQ-ACR-003 exactly (debt-proceed / hold-and-split→hold+ref / split / other-unreadable→hold); clean PASS writes nothing (REQ-ACR-004). No `AskUserQuestion` anywhere in the new files (static grep stays 0).

**Exit (count-first, N4):** `go test -list '^(TestRecordCeilingOutcome|TestRecordCeilingOutcomeDebtProceed|TestRecordCeilingOutcomeUnknownPolicy|TestRecordCeilingOutcomeSplitValue|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` lists exactly 5 tests, THEN `go test -run '^(TestRecordCeilingOutcome|TestRecordCeilingOutcomeDebtProceed|TestRecordCeilingOutcomeUnknownPolicy|TestRecordCeilingOutcomeSplitValue|TestEvaluatePlanAuditCeiling)$' ./internal/runtime` exit 0; `go run ./cmd/moai spec ceiling --help | grep -c "ceiling"` ≥ 1 — the output-content gate (the help exit code alone is vacuous: it exits 0 on the unstarted tree; RED value recorded in LEDGER-ACR-J); `grep -c AskUserQuestion internal/runtime/audit_ceiling.go internal/cli/spec_ceiling.go` → 0 for both (exit 1 per file).

### M3 — required-backend refusal + resolution path + convention line (work item 3)

Files 9-18. Predicate signal + refusal arm; convention doc line in both copies (same change); the resolution path fail-closed end to end — resolver error distinct (file 13/16), pins-loader "(N3)" fold retired (files 14/15), caller surfacing and the six-site caller inventory (files 16/18).

**Exit (count-first, N4):** `go test -list '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` lists exactly 1, THEN `go test -run '^TestAdmitRequiredBackendFail$' ./internal/auditverdict` exit 0; `grep -c "required_backend_fail" .moai/docs/audit-artifact-convention.md` ≥ 1 AND `grep -c "required_backend_fail" internal/template/templates/.moai/docs/audit-artifact-convention.md` ≥ 1 (per-file pair, not a sum); `go test -list '^(TestResolveAuditGatesConfigErrorDistinct|TestWorkflowAuditPinsErrorNotFolded|TestWorktreeRootSurfacesGateError)$' ./internal/cli` lists exactly 3, THEN `go test -run '^(TestResolveAuditGatesConfigErrorDistinct|TestWorkflowAuditPinsErrorNotFolded|TestWorktreeRootSurfacesGateError)$' ./internal/cli` exit 0.

### M4 — verification + evidence

RED→GREEN matrix for all 11 ACs recorded into `progress.md` §E.1 with verbatim outputs; narrow selectors only (no whole-package `go test ./...` — lane discipline); `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-002 --strict` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (C5).

**Exit:** lint strict 0 findings; all AC GREEN cells carry command + verbatim output + exit code; §E.1 `audit_ready: true`.

## §D Verification Plan

- Every acceptance criterion adopts the two-cell discipline (acceptance.md §A): the RED-now cell measured at `e497f6936` (pinned in acceptance.md §B) and the GREEN command that flips it, with the milestone that flips it named.
- Go verifications run per-package with `-run` selectors; concurrency-free additions need no `-race` batch beyond the M3 predicate test.
- The report-tree measurements in §A/§B of spec.md are primary-checkout reads (worktrees excluded); re-run only if the audit disputes them, never as routine.

## §E Risks and Orderings

- M1 before M2 (the verb consumes the counter); M3 is independent of M1/M2 and may land first if the audit demands it.
- The convention-doc pair (files 11+12) must land in ONE commit — the mirror discipline lesson (one-side-only edits were a landed defect class).
- `spec.go` registration is the only edit to shared CLI wiring; the cobra duplicate-`Use` guard is satisfied by the unique verb name `ceiling` (no existing `moai spec` subcommand shares it).

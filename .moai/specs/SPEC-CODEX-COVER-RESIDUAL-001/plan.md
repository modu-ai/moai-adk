---
id: SPEC-CODEX-COVER-RESIDUAL-001
title: "Implementation plan — codex coverage residual"
version: "0.1.0"
created: 2026-09-07
---

# SPEC-CODEX-COVER-RESIDUAL-001 — Implementation Plan

Sections are ordered by decision-reversibility: the choices most likely to change under review come first (§A.6), and the mechanical steps come last (§F).

## §A Context

### A.1 Location and tree

- Project root: this worktree (`git rev-parse --show-toplevel`); all paths below are project-root-relative.
- Branch `WT-codex-cover-residual`, plan-phase base HEAD `bf779ecf2`.
- Card `t519`. Every commit message on this branch names `t519`. Evidence path: `.moai/reports/t519/`.

### A.2 Artifacts

- `.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/spec.md` — 10 requirements (REQ-CCR-001..010), documented-skip record §D, out-of-scope §F
- `.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/acceptance.md` — 12 criteria (AC-CCR-001..012), mutant ledger §D
- `.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/plan.md` — this file
- `.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/progress.md` — §E skeleton

Tier **M** (3-file set + §E skeleton). The dispatch named Tier S; the rationale for the correction is spec.md §A.1 and it is surfaced to the lead in the completion report. Consequences: plan-audit PASS threshold 0.80; the Section A-E delegation template is REQUIRED, not optional.

### A.3 Development mode

`cycle_type=tdd`. The work is test authoring, so RED-GREEN-REFACTOR degenerates to a mutant-driven variant: the RED is produced by a mutant applied to already-correct production code, not by absent implementation. spec.md REQ-CCR-008 and acceptance.md §D own that discipline.

### A.4 Existing infrastructure — PRESERVE, reuse, do not extend

Reuse as-is (no modification):

| Seam / helper | Location | Use |
|---|---|---|
| `withChangeDetector(t, bool)` | codex_review_gate_test.go:28 | drive the self-gate |
| `withCodexLookPath(t, fn)` | codex_review_gate_test.go | assert codex is not consulted |
| `withCodexRunner(t, runner)` | codex_review_gate_test.go | runner injection |
| `withCodexSession(t, script)` + `codexSessionScript(...)` | codex_review_gate_test.go | drive ALLOW / BLOCK verdicts |
| `&fakeCodexSession{startErr: errFakeCodexCrash}` + `stubCodexRunner{}` | codex_review_gate_test.go:152-164 | the handler-error arm |
| `writeWorkflowYAML(t, body)` | multi_review_gate_wiring_test.go:33 | isolated temp project root |
| `assertAllowJSON(t, stdout)` | multi_review_gate_wiring_test.go:157 | the ALLOW contract |
| `fakeCodexConn` + `fakeCodexConnPID` | mcp_codex_test.go:88, codex_jobs_test.go:31 | the pid conn arm |

### A.5 PRESERVE list — do not modify

- Every non-test `.go` file in the repository (REQ-CCR-007). Named explicitly because they are the files a coverage card is most tempted to touch: `internal/cli/codex_review_gate.go`, `internal/cli/mcp_codex.go`, `internal/cli/hook.go`.
- `internal/cli/multi_review_gate_wiring_test.go` — including `newGateCmd`, which is hard-wired to `runMultiReviewGate`. Its helpers are read and called, never edited.
- `TestCodexReviewGate_SubcommandRegistered` (codex_review_gate_test.go:324) and every other existing test.
- Runtime-managed paths: `.moai/state/`, `.moai/cache/`, `.moai/logs/`, `.moai/harness/`.
- Every other SPEC directory under `.moai/specs/`.

### A.6 Decisions taken at plan time (reversible — review these first)

These are the judgment calls. Each states what was chosen, the alternative, and what it would cost to change.

**D1 — a new test file, not an extension of `codex_review_gate_test.go`.** The new RunE tests go in `internal/cli/codex_review_gate_wiring_test.go`. Chosen because the repo already carries exactly this split on the sibling gate (`multi_review_gate.go` logic tests vs `multi_review_gate_wiring_test.go` wiring tests), and the naming symmetry is what a future reader will look for. *Alternative:* append to the existing 355-line `codex_review_gate_test.go`. *Cost to reverse:* trivial — move five functions; no test content changes.

**D2 — the command constructor is named `newCodexGateCmd`.** `newGateCmd` already exists at package scope, bound to `runMultiReviewGate`; a same-named second constructor will not compile, and modifying the existing one is forbidden by A.5. *Alternative:* `newCodexReviewGateCmd`. *Cost to reverse:* a rename.

**D3 — the pid test goes in `mcp_codex_test.go`, beside `TestRealCodexConnPid`.** The two are siblings: the same 3-arm nil-guard shape on two different receivers, one written by SPEC-CODEX-TEST-GAPS-001 REQ-CTG-012, one by this SPEC's REQ-CCR-005. Splitting them across files would hide the pairing. *Cost to reverse:* trivial.

**D4 — the S1 arm is skipped, not made reachable.** Making it reachable needs a production seam, which REQ-CCR-007 forbids. Recorded as a documented skip (spec.md §D) with the coverage ceiling derived in §D.1. *Alternative:* introduce a handler seam and reach 100%. *Cost to reverse:* this is the one decision that is NOT cheap to reverse — it would turn a tests-only card into a production-code card and invalidate AC-CCR-007. Flag to the operator rather than reversing silently.

**D5 — the coverage threshold is ≥90.0%, not 100%.** Follows from D4. The derivation (12 of 13 statements ≈ 92.3%) is recorded in spec.md §D.1 as a derivation from reading the source, explicitly not a measurement. *Cost to reverse:* none before run; after run the measured figure replaces the derivation.

**D6 — adjacent under-covered functions are declined.** `reviewableFromPorcelain` (83.3%) and `readHookInput` (85.7%) sit in the same file and are not named by the card. Declined as scope creep, recorded as residual risk (spec.md §F). *Cost to reverse:* they would add roughly two more test items and push the AC count toward the Tier M ceiling of 16.

## §B Known Issues (filtered — Section B categories that apply)

Categories B1, B2, B7, B12 do not apply (no syscall use, no retired-SPEC conflict in scope, no hook path resolution, no CHANGELOG emission at run phase). The rest:

- **B3 / B11 — subagent boundary.** No `AskUserQuestion` in `internal/cli` code, tests included. On a blocker, return a structured blocker report to the orchestrator.
- **B4 — frontmatter schema.** spec.md carries the canonical 12 fields plus `era` / `tier` / `related_specs`; `created:` and `updated:` are the canonical spellings, never the snake_case aliases.
- **B5 — CI tiers.** spec-lint, golangci-lint, and per-OS Test can each fail independently. Classify any failure as NEW or inherited-baseline before acting.
- **B6 — spec-lint heading convention.** spec.md §F carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets, not a bare H2.
- **B8 — working-tree hygiene.** Stage by explicit pathspec. Never `git add -A` / `git add .` / `git commit -a`. Re-read `git status --short` in the same call that stages.
- **B9 — commits.** This repository runs the git-flow lane protocol: card worktrees branch from `develop`, there are no card PRs, and the lane does **not** push. Commit on `WT-codex-cover-residual` with Conventional Commit subjects naming `t519`; integration is the lead's window.
- **B10 — scope discipline.** Touch nothing outside the A.5 PRESERVE boundary. Other lanes are live in sibling worktrees.

## §C Pre-flight

Run as one batch before the first edit; record the outputs in progress.md §E.2.

```
git rev-parse --show-toplevel
git rev-parse --short HEAD
git branch --show-current
git status --short
grep -rn 'func TestRunCodexReviewGate' internal/cli/
grep -rn 'func TestCodexSessionHandlePid' internal/cli/
grep -rn 'func newCodexGateCmd' internal/cli/
```

Expected: root is this worktree, HEAD descends from `bf779ecf2`, branch is `WT-codex-cover-residual`, and the three greps print nothing (the targets do not yet exist).

**A silent grep is not evidence on its own.** Pair the three empty greps with the control `grep -rn 'func TestRunMultiReviewGate' internal/cli/`, which must print three rows from `multi_review_gate_wiring_test.go`. A control that also prints nothing means the grep form or the path is wrong, not that the targets are absent.

## §D Constraints

1. **Tests only.** Zero diffs to non-test `.go` files, at every commit and at close (REQ-CCR-007, AC-CCR-007).
2. **No new production seams**, including one that would make S1 reachable (D4).
3. **No new shared test infrastructure** beyond the single new file's own constructor. The A.4 helpers are sufficient.
4. **Do not modify `newGateCmd`** or anything else in `multi_review_gate_wiring_test.go`.
5. **Verification is package-scoped**: `go test ./internal/cli/` with timeout ≥600s (baseline 473.7s; use `-timeout 700s`). Never `go test ./...` locally.
6. **The env scrub is one compound invocation**: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`. A separate `unset` call does not carry into the next command, and the `env -u` form is refused by the command guard.
7. **Never `--no-verify`, never `--amend`, never force-push.** The lane does not push at all.
8. **Every mutant is reverted before the commit that lands its test.**
9. Code and comments in English (`code_comments: en`).

## §F Milestones

Owner: `manager-develop`, `cycle_type=tdd`.

### M1 — Axis 1: the RunE wiring tests (Priority High)

Closes the 0.0%. New file `internal/cli/codex_review_gate_wiring_test.go`, mirroring `multi_review_gate_wiring_test.go` in shape and comment style.

1. Add `newCodexGateCmd(stdin string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer)` — a throwaway `&cobra.Command{Use: "codex-review-gate", RunE: runCodexReviewGate, SilenceUsage: true}` with `SetIn`/`SetOut`/`SetErr` bound to in-memory buffers.
2. `TestRunCodexReviewGate_InvalidStdinFailsOpen` (AC-CCR-001) — stdin `{not json`.
3. `TestRunCodexReviewGate_EmptyStdinFailsOpen` (AC-CCR-002) — empty stdin.
4. `TestRunCodexReviewGate_HappyPathAllow` (AC-CCR-003) — `writeWorkflowYAML` with the gate disabled, plus a `withCodexLookPath` `t.Fatal` guard.
5. `TestRunCodexReviewGate_HandlerErrorFailsOpen` (AC-CCR-004) — gate enabled, `withChangeDetector(t, true)`, `fakeCodexSession{startErr}` + `stubCodexRunner{}`; assert the stderr diagnostic.
6. `TestRunCodexReviewGate_BlockVerdictPropagates` (AC-CCR-005) — gate enabled, `withChangeDetector(t, true)`, `withCodexSession(t, codexSessionScript("- [P1] found issues"))`; decode stdout and assert `decision == "block"` with a non-empty reason.
7. Work mutants M1, M2, M3a, M3b, M4 one at a time: apply, observe RED, record verbatim output, revert, observe GREEN.
8. Commit: `test(SPEC-CODEX-COVER-RESIDUAL-001): M1 runCodexReviewGate wiring tests (t519)`.

Exit: AC-CCR-001..005 PASS; mutants M1-M4 recorded RED-then-GREEN; `git status --short` clean of production sources.

### M2 — Axis 2: the pid nil-guard arm (Priority Medium)

Append `TestCodexSessionHandlePid` to `internal/cli/mcp_codex_test.go`, beside `TestRealCodexConnPid` (mcp_codex_test.go:647).

1. Assert `(*codexSessionHandle)(nil).pid() == 0` — a typed nil receiver, which is what makes the `h == nil` short-circuit observable.
2. Assert `(&codexSessionHandle{}).pid() == 0`.
3. Assert `(&codexSessionHandle{conn: &fakeCodexConn{}}).pid() == fakeCodexConnPID`. Constructing `&fakeCodexConn{}` with a nil `sent` field is safe here: `pid()` reads no field of the conn beyond the interface dispatch, and `send` is never called.
4. Work mutants M5a and M5b: apply, observe RED, record, revert, observe GREEN.
5. Commit: `test(SPEC-CODEX-COVER-RESIDUAL-001): M2 codexSessionHandle.pid nil-guard arm (t519)`.

Exit: AC-CCR-006 PASS; M5a/M5b recorded.

### M3 — Close: measure, guard, record (Priority High)

1. Run the named-test sweep and count PASS lines (AC-CCR-009): exactly six, matched on `^--- PASS: Test` with no end anchor.
2. Re-measure coverage on the final tree with the §D.6 env-scrubbed command; extract the two target rows with `go tool cover -func`; record both alongside the package figure (AC-CCR-008).
3. Adopt AC-CCR-007 by firing mutant M6 once — append a newline to a production file, observe `git status --short` list it, revert — then observe the union clean.
4. Run the full package suite: `go test ./internal/cli/` with `-timeout 700s`.
5. Quality gate (AC-CCR-011): `go vet`, `golangci-lint run`, `gofmt -l internal/cli/`; classify findings as NEW or inherited.
6. Write progress.md §E.2 evidence and the §E.3 audit-ready signal.
7. Commit: `test(SPEC-CODEX-COVER-RESIDUAL-001): M3 close — coverage re-measurement + evidence (t519)`.

Exit: every AC in acceptance.md §B observed; §E.2/§E.3 populated.

## §G Anti-Patterns

- **Editing production code to raise a number.** The single prohibition this card exists under. A test that cannot reach a line without a production change is a documented skip, not a refactor request.
- **Adopting a test on its own absence.** "The test does not exist, so the suite is red" is false: a selector matching zero tests exits 0 and prints `ok`. Adoption is by mutant.
- **Trusting the M3-vac mutant.** Swapping `out` for `&hook.HookOutput{}` at line 196 changes nothing observable on stdout. It is pre-recorded in acceptance.md §D so it is not mistaken for adoption evidence.
- **Asserting only stdout on the handler-error arm.** Both arms emit `{}`; the stderr diagnostic is the only discriminator.
- **Reusing or editing `newGateCmd`.** It is bound to the other gate. A same-named constructor will not compile.
- **Running `go test ./...` locally.** Parallel lanes doing this drove machine load to 413 (2026-08-15). Package-scoped only; CI owns the full-suite verdict.
- **Splitting `unset` from the test command.** Each Bash call is a fresh process; a standalone `unset` scrubs nothing.
- **Sweep-staging.** `git add -A` in a tree with live sibling lanes is how another session's work is lost.
- **Fixing the wrong comment.** `multi_review_gate_wiring_test.go`'s header claims a codex counterpart that never existed. Building the counterpart is this card; editing that comment is not.

## §H Cross-References

- spec.md §B (baseline + provenance), §D (skip record + ceiling derivation), §F (out of scope)
- acceptance.md §C (criteria), §D (mutant ledger), §F (definition of done)
- `.moai/specs/SPEC-CODEX-TEST-GAPS-001/` — structural precedent (plan-audit 1.00)
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §2.1
- `.claude/rules/local/gitflow-lane-protocol.md` — lane commits, no push, lead-owned integration
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E delegation template (required at Tier M)

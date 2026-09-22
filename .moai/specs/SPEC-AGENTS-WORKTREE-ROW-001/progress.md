# progress.md — SPEC-AGENTS-WORKTREE-ROW-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-22
- plan_auditor: pending (independent audit is the next act after this commit; not run by this
  session — the lead dispatches plan-audit separately)

### Plan-phase evidence (plan-phase, this run, worktree t1071 @ `cd99336bf`, 2026-09-22)

Re-measurement of the card's investigator citations (spec.md §A, C1–C8), executed before any
artifact was written:

- **C1** `AGENTS.md:18-25` read — capability table, three rows, no worktree row. Negative
  control `grep -c '^| worktree-entry |' AGENTS.md` → `0`; positive control
  `grep -c '^| question-channel |' AGENTS.md` → `1`.
- **C2** `AGENTS.md:106-111` read — §3 launcher enumeration lists `moai cc -w`, `--spawn`,
  `EnterWorktree(<path>)` only; the `moai codex` form is absent.
- **C3** `internal/template/templates/AGENTS.md.tmpl:18-29` read — three-column table with
  **seven rows** (`question-channel`/`task-list`/`design-sync` plus `agent-spawning`,
  `output-style`, `slash-commands`, `workflow-scripts` — four rows the root copy lacks, an
  intentional-fork divergence), same absence of a worktree row. Negative control grep → `0`.
  (Corrected from the original "same three rows at `:18-25`" at plan-audit iter-1 — D1; the
  full-row enumeration was re-run and confirmed by this author before the correction landed.)
- **C4** `internal/template/templates/AGENTS.md.tmpl:291-305` read — `## 11. moai CLI Verbs`
  at `:291`, nine verb rows `:295-303`, no `moai codex` row. Negative control
  `grep -c '^| \`moai codex\`' …` → `0`; positive control `grep -c '^| \`moai init' …` → `1`.
  The dispatch's `:293-302` citation is corrected to `:293-303` (re-measured range governs).
- **C5** `internal/cli/codex_launcher.go` read — verb registration `:324` (`Use:` line, code
  declaration), `DisableFlagParsing: true` `:360`, `resolveCodexWorktreeDir` `:272-299` (L2
  validation `:279`, L1 join `:286`, absent-directory diagnostic `:289-297`), named diagnostics
  `:47`/`:52`, help text `:341-342`. Entry exists; creation is impossible.
- **C6** The dispatch's `:16-17` citation is a package-doc comment block (`codex_launcher.go:16-19`),
  not a declaration — the card's own comment-hit warning applied to the citation itself.
  Declaration anchors above are code.
- **C7** `grep -n 'moai CLI Verbs' AGENTS.md` → no matches — the root copy carries no Verbs
  table; the registration fix is mirror-only.
- **C8** `internal/template/templates/AGENTS.md.tmpl:270` prose-references `moai codex` while
  the Verbs table omits it — the command is live; only the inventory lags.

Pre-edit control batch (one compound invocation, 2026-09-22):

```
G1-root-worktree-row: 0
G2-tmpl-worktree-row: 0
G3-tmpl-codex-verb: 0
G4-positive-control-root-table: 1
G5-positive-control-verbs-table: 1
exit=0
```

Interpretation: the three zeros are evidence of absence (both positive controls fired on the
same run), and the two ones prove the row-shape patterns fire against the tables as they stand —
so the post-edit guards measuring `1` will be a real flip, not a pattern that matches nothing.

Ambiguity decisions taken (for the plan-auditor):

1. The dispatch named the root capability table and the mirror Verbs table explicitly, but
   re-measurement found the mirror's *capability table* also lacks the row (C3). Constraint (c)
   was read as licensing the fix; it is REQ-AWR-002. Quotation-of-authority note (added at
   plan-audit iter-1, D6): the lead's plan dispatch (2026-09-22) words constraint (c) as "Judge
   each copy separately and fix each accordingly", while the audit reads a different dispatch
   wording from the card body — "judge each copy separately, and template edits require `make
   build` regeneration". The two wordings diverge on the tail clause; both are retained here
   verbatim rather than adjudicated by this session. The licensing conclusion stands on the
   file-division grant either way (audit adjudication 2).
2. The dispatch's `:16-17` coordinate was replaced by code anchors (`:324`, `:272-299`) per the
   card's own comment-hit rule.
3. `tier: M` was chosen over `S` because the card requires `acceptance.md` explicitly and the
   verification chain (guards + build + lint) exceeds a two-artifact scope.

SPEC lint at plan-phase close (tree build `go run ./cmd/moai spec lint
SPEC-AGENTS-WORKTREE-ROW-001`, this run, this tree): exit 0, `0 error(s), 4 warning(s)` —
REQ collection FIRED (four REQ-AWR-* rows collected and modality-judged; REQ-AWR-004 emitted no
finding). The four findings are `ModalityMalformed` on REQ-AWR-001/002/003/005 — the same
prose-shape class card t1057 measured; they are warnings, not errors, and are left for the
plan-auditor to weigh rather than re-worded mid-flight by this plan phase.

### Plan-audit iter-1 — PASS-WITH-DEBT 0.875, rework applied (2026-09-22)

Verdict: PASS-WITH-DEBT, overall 0.875 (Tier M threshold 0.80), BLOCKING 0. Report:
`.moai/reports/t1071/plan-audit-iter1.md`. The audit's five adjudications of this author's
flagged decisions all resolved favorably (lint warnings absorbed as documented debt D9;
REQ-AWR-002 in-scope; Tier M sanctioned; the C4 `:293-303` correction verified; the C6
comment-to-code anchor substitution verified).

Rework applied this session, all in the SPEC's own four artifacts (no contract file touched,
`make build` not run — nothing under `templates/` has changed):

| Defect | Repair |
|---|---|
| D1 (SHOULD-FIX) | C3 rewritten — mirror table is **seven rows at `:21-29`**, four of which the root lacks; corrected in spec.md §A C3, REQ-AWR-002 (`:21-29`, append after last row), plan.md §A.1 item 2, plan.md M2(a) (`:29`, not the mid-table `:25`). Author's own re-enumeration of the mirror rows confirmed the auditor's count before the fix landed |
| D2 (SHOULD-FIX) | AC-AWR-004 rewritten with explicit adoption cells — green-now by construction, red constructible at run-phase (weakened-filter mutant observing `2`); AC-AWR-005 reclassified **regression-guard** per verification-completeness §2.1 (red unconstructible — no build-chain step reads `AGENTS.md.tmpl` drift, Makefile:34 measured by the audit), "proving" clause dropped |
| D3 (SHOULD-FIX) | AC-AWR-003's Then now names **both** verbs — the launch verb (`cli`) and the readout verb (`status`) — closing the missing-`cli` mutant |
| D4 (MINOR) | DoD cross-reference `§E.2` → `§E.1` |
| D5 (MINOR) | AC-AWR-004 spells both copies' full fence-file paths |
| D6 (MINOR) | The constraint-(c) quotation conflict is recorded verbatim in both wordings (lead's dispatch vs card body) — not adjudicated here; stands on the file-division grant either way |
| D7 (MINOR) | REQ-AWR-005 and AC-AWR-005 reworded — `.tmpl` is a compile-time embed the generating steps do not process; the discharge is the recompile; the make-output-dependent no-op clause is dropped |
| D8 (MINOR) | AC-AWR-001/003 "unchanged rows" claims restated as **presence counts** with the content-identity limit named (carried by plan.md §G at diff review) |
| D9 (DEBT) | The 4 lint warnings — absorbed per the audit's adjudication 1; no rewording |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Decision: serial

- Inputs: tier M · scope 2 contract files (AGENTS.md + AGENTS.md.tmpl) + SPEC artifact updates · domains 1 (documents only, zero code) · file language 100% markdown · concurrency benefit LOW (ordered M1→M2→M3 chain; M3's regeneration depends on M2's edits — no inter-file parallelism) · agent-team prereqs not requested
- Mode evaluation: direct — not selected (two contract files + guard execution + make build exceed a trivial single-line change); fanout — not selected (single domain, nothing parallelizable); sweep — not selected (~2 files, semantic doc edits, not mechanical bulk); agent-team — not selected (no explicit operator request)
- Justification: a single manager-develop (dev-swguard) spawn per milestone is the whole envelope; the coding-task-parallelism caveat and the single-domain scope both point at serial as the simpler sufficient mode
- Kickoff status: Implementation Kickoff Approval granted by the operator 2026-09-22 ("전부 승인"), relayed via the lead with the four run constraints to be carried verbatim in the dev-swguard dispatch

---

🗿 MoAI

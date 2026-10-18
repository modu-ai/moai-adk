# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — plan (v0.3.0, piece A)

Priority-ordered, no time estimates. Development mode TDD (RED first). Each milestone is its own commit with subject `feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> <what> (card t1872)` (M0 uses `docs(...)`), carrying the `🗿 MoAI` trailer and, for the agent-authored ones, `Authored-By-Agent`. Scope is piece A only (board `9dbe`); follow-ups are in `spec.md` §7 and not planned here.

## A. Review-focus order (decision reversibility)

1. **Record and verifier contracts** (manifest, attempt, verdict, producer verbs; REQ-LVM-001..006, 013, 014) — data-model decisions, highest change likelihood.
2. **The receipt commit field and the audit binding** (REQ-LVM-008..010) — a schema change in the receipt store plus a landing-policy edge.
3. **Switch discipline and prerequisites** (REQ-LVM-012, spec C5) — operator-visible landing policy; the switch is outside this branch.
4. **Lane guidance** (REQ-LVM-011).
5. Wiring edits in the merge step, `complete` and condition (d) — mechanical.

## B. Preconditions and constraints

- Git-flow window path only. Resource rule 8c1e: no `go build`, `go test`, `go vet`, `golangci-lint` or `make` while the host load (1 min) exceeds 32 or free memory is under 50%; one Go call at a time, `-p 4` or less for several packages. Heavy suites run under the `go-test-heavy` lease with one acquire, one command, immediate release.
- Work in the card worktree; explicit pathspecs; no push; nothing under `.moai/reports/` is ever force-added — the M0 baseline is a tracked file under the SPEC directory.
- **No evidence or progress backfill commit lands after the audited SHA** (board `0577` P3): bookkeeping, including any `pending-backfill` completion, is committed before the final audit, which binds the final HEAD.
- The branch never enables the gate (REQ-LVM-012). Prerequisites of the later enabling are stated in `spec.md` C5 and are not steps of any milestone.
- Rule files loaded into the session prefix are edited last (M4), per the cache-aware batching directive.

## C. Milestones in dependency order

### M0 — baseline, alone and first

Re-run the RED-now observations and the landing-interval measurement on the run tree; commit `evidence/m0-baseline.md` (tracked, never under `.moai/reports/`) as the only file of the commit.

Exit: AC-LVM-012 (M0 leg).

### M1 — gate core, producer verbs, zero-Go path

Manifest producer (`moai integration manifest derive|show|verify`), attempt records and `remeasure --call` (command binding, per-package test counts), abort/restart classes, the whole-manifest verifier with its twelve reject classes, the zero-Go evidence path, wiring into the merge step, `complete`, condition (d) and the verify verb, config key `workflow.landing_gate.enabled` (false, read from the integration root) and the narrowing allowance list. Negative tests and the positive control are written first and end green. A characterization test of today's verdicts (the oracle of AC-LVM-007) is recorded before any edit of `ValidateRemeasureRecord`.

Exit: AC-LVM-001, 002, 003, 004, 005, 006, 007, 013, 014, 015.

### M2 — receipt commit field

Additive `audited_commit` field in the receipt schema and its writer: read the audited tree's HEAD before the backends run, stamp it, leave it empty if HEAD differs at write time; never rewrite an existing receipt. The MCP server that writes receipts needs a build containing this change before a receipt counts as bound (spec C5).

Exit: AC-LVM-009.

### M3 — audit binding in `complete`

Under the gate flag, `complete` binds the SPEC card's sync-audit verdict to the current HEAD (evidence-directory difference only, no other exemption) and requires every cited receipt to be a readable, passing, unmet-gate-free codex or multi-backend receipt whose audited commit equals HEAD. Reuses `auditverdict` admission, `runtime.ResolveRequiredBackends`, and `auditreceipt.ReadReceipt`.

Exit: AC-LVM-008, 010.

### M4 — lane guidance

`factory-dispatch-mechanics.md` and its mirror state the gate, the three commands, restart as a new attempt, the no-backfill-after-audit rule, and the by-path binary and receipt-server prerequisites; Template-First: edit the template tree, `make build` (when the host admits it), local copies regenerated. Includes the recovery paragraph (the key-off one-path card).

Exit: AC-LVM-011, 012.

## D. Ordering, machine-checkable

M0 < M1 < M2 < M3 < M4 in the commit graph (AC-LVM-012); the branch's `workflow.yaml` history is empty (no enabling); the enabling commit is a later, separate commit on main made after this SPEC has landed and a by-path binary built from the landing HEAD is in use; its post-landing witness is in AC-LVM-012. Cards already merge-ready are not re-audited for the switch; each takes the HEAD-confirmation audit at landing and uses the structured form for its re-measure once the key is on.

## E. Per-milestone method

Each milestone: (1) write the AC's tests first and observe them red for the stated reason, keeping the RED output as milestone evidence per the TDD result contract; (2) implement the minimum; (3) refactor; (4) run the milestone's AC commands with the empty-sweep guard; (5) one commit.

- **M1 specifics.** Changed packages come from the diff against the merge-base, mapped through `go list` (a changed file under a package directory marks the package, embed and testdata included); importers from `Imports`, `TestImports`, `XTestImports`; embed owners from `EmbedFiles`/`TestEmbedFiles`/`XTestEmbedFiles`; a deleted package is handled by comparing the base and candidate package graphs. The structured form extends `moai integration remeasure` (keeping `RunRemeasure`'s clean-start and clean-finish checks, `integration_remeasure.go:634-720`); command comparison reuses the existing shell-segment helpers (`shellSegments`, `goSubcommand`, `stripLeadingEnv`). The verifier replaces the decision of `ValidateRemeasureRecord` where it is consulted (five call sites, `research.md` §2) only while the key is true.
- **M2 specifics.** The writer is `recordAuditReceiptAt` (`internal/cli/mcp_audit_receipt.go:60-100`); the schema is `auditreceipt.Receipt` (`store.go:166-177`). The hook-side guard (`CheckCitedReceipts`) keeps its meaning; the new field is consulted only by the landing binding.
- **M3 specifics.** Receipts are resolved from the store root the audit call used (`receiptTarget`); the binding reads the verdict file's cited ids as the guard already does.

## F. Risks

| Risk | Mitigation |
|---|---|
| Enabling later refuses merge-ready cards until their calls are in the structured form and their audit is bound | decided cost (board 0577 P2); M4 guidance precedes the switch; the leader's pre-switch list is in `research.md` §10 |
| The gate ON in a build that lacks it is silently inert | the switch is gated on the by-path binary prerequisite (C5); attempts record the judging build identity |
| Receipts from a server without the writer change are unbound | prerequisite C5; an unbound receipt refuses, fail-closed |
| `go list` derivation misses a reader of a changed non-Go file | the reader search is recorded with its command; a computed-path read stays residual (`design.md` §3) |
| Plan audit holds at the Tier L ceiling (round 2 of 3) | every blocking and recommended defect of iteration 1 is addressed in this revision; disposition in `progress.md` |

## G. Cross-references

`spec.md`, `acceptance.md` (RED-now ledger), `design.md`, `research.md`, `decision-index.md`. Rules touched at M4: `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing and its template mirror.

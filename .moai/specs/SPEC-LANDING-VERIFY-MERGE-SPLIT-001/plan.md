# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — plan (v0.9.0, piece A2)

Priority-ordered, no time estimates. Development mode TDD (RED first). Each milestone is its own commit with subject `feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> <what> (card t1872)` (M0 uses `test(...)`), carrying the `🗿 MoAI` trailer and, for the agent-authored ones, `Authored-By-Agent`. Scope is piece A2 only (boards `9dbe`, `d20e`); the completeness gate is SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875), blocked until this card is merged-local, and is not planned here.

## A. Review-focus order (decision reversibility)

1. **The receipt commit field and the verdict citation line** (REQ-LVM-009, 018) — a schema change in the receipt store and a new machine-read line: highest change likelihood.
2. **The audit binding in `complete`** (REQ-LVM-008, 010) — a landing-policy edge: what equals HEAD, which receipt qualifies, where the refusal sits.
3. **The gate-state source** (REQ-LVM-017, board 470c) — a policy edge: which committed blob decides, and the fail-modes.
4. **Switch discipline and prerequisites** (REQ-LVM-012, spec C5) — operator-visible landing policy; the switch is outside this branch. Decided (Q27, board 6404): the key turns on only after A2 and A1 have both landed; this SPEC ships the audit binding and the switch reader with the key off.
5. **Lane guidance and protocol text** (REQ-LVM-012, 018 documents).
6. Wiring in `factoryCompleteCard` and the status display — mechanical.

## B. Preconditions and constraints

- Git-flow window path only. Resource rule 8c1e: no `go build`, `go test`, `go vet`, `golangci-lint` or `make` while the host load (1 min) exceeds 32 or free memory is under 50%; one Go call at a time, `-p 4` or less for several packages. Heavy suites run under the `go-test-heavy` lease with one acquire, one command, immediate release.
- Work in the card worktree; explicit pathspecs; no push; nothing under `.moai/reports/` is ever force-added — the M0 baseline is a tracked file under the SPEC directory and the characterization test a tracked test file.
- **No evidence or progress backfill commit lands after the audited SHA** (board `0577` P3): bookkeeping, including any `pending-backfill` completion, is committed before the final audit, which binds the final HEAD.
- The branch never enables the gate (REQ-LVM-012). Prerequisites of the later enabling are stated in `spec.md` C5 and are not steps of any milestone.
- The re-measure record checks of `complete` (steps 3 and 4, `factory_card.go:2260-2300`) and of the merge step are not edited (spec C9); both pieces edit `factoryCompleteCard`, which is why piece A1 waits for this card's landing.
- Rule files loaded into the session prefix are edited last (M4), per the cache-aware batching directive.

## C. Milestones in dependency order

The order is the split brief's: baseline, receipt field, audit binding, switch reader. The binding applies only while the gate is on, so in M2 it reads the gate through a seam whose default answers off; M3 replaces that default with the real reader and adds the key.

### M0 — baseline and characterization, alone and first

Re-run the RED-now observations on the run tree and pin today's `moai factory complete` outcomes. One commit, two tracked files and nothing else: `evidence/m0-baseline.md` (never under `.moai/reports/`) and the characterization test `TestCompleteGateOffMatchesTodaysBehaviour` in `internal/cli/factory_complete_gate_off_test.go`, which runs green against the unchanged code. The test is a baseline artifact: it lands in its own commit ahead of every commit that edits `factoryCompleteCard`, because the commit graph is the only witness of that order (verification-claim-integrity §2.3); AC-LVM-012 checks it by ancestry and by the absence of the test file from the edit commit.

Exit: AC-LVM-012 (M0 leg), AC-LVM-007 (characterization leg).

### M1 — receipt commit field

Additive `audited_commit` and `audit_target` fields in the receipt schema and its writer; the entry points `mcp_codex.go` (two call sites) and `mcp_convergence.go` (one) capture the audited tree's HEAD and review target before the backends run and pass them; the writer stamps them, leaves `audited_commit` empty if HEAD differs at write time, never rewrites an existing receipt. The MCP server that writes receipts needs a build containing this change before a receipt counts as bound (spec C5).

Exit: AC-LVM-009.

### M2 — audit binding in `complete`

With the gate state entering through a package-level seam whose default reads off: the newest-verdict ranking moves to `internal/auditverdict` (one copy, `homestate` calls it); `internal/factorylane` gets the binding — locate the verdict (SPEC card: newest `sync-audit*.md`; SPEC-less card: `verdict.md`), parse the single `audit_receipt:` machine line (new parser; refusal for missing, duplicate, empty, `none`, malformed, or absent-from-store), bind the verdict and at least one qualifying receipt (looked up at `StoreRoot(audited tree)`, readable, codex or multi-backend tool, `tree_root` equal to `auditreceipt.Canonical` of the audited tree's path, unmet gate empty, verdict pass, audited commit equal to HEAD) with one rule — equals HEAD, no tolerance for an empty-diff or evidence-directory ancestor; failed extra citations are recorded and never replace the pass. `complete` calls the binding after the integration branch and tree are resolved and the three provisioning refusals have passed (`factory_card.go:2178`) and before the window is acquired (`factory_card.go:2185`), so every refusal leaves the window free and the card unchanged. Reuses `auditverdict.Parse` and `AdmitWithRequired`, `runtime.ResolveRequiredBackends` (injected by the caller) and `auditreceipt.ReadReceipt`.

Exit: AC-LVM-008, 010, 018.

### M3 — gate-state reader, key, status display

The reader (new file in `internal/factory`): the integration branch is the held window record's, else the configured target; none configured reads off; `refs/heads/<branch>^{commit}` is resolved once and the SHA is used throughout, so a same-named tag cannot shadow the branch, and a name that does not resolve there (a tag-only name included) refuses with `gate_state_integration_branch_unresolvable`; with the SHA, `git cat-file -e <sha>:.moai/config/sections/workflow.yaml` tells an absent path (off) from a failed read, then `git show <sha>:.moai/config/sections/workflow.yaml` (no working tree consulted); an unreadable present file reads on with `gate_state_unreadable` and one JSON line appended to `.moai/logs/landing-gate-audit.jsonl` under the integration lock root. The configuration type `LandingGate` (key `workflow.landing_gate.enabled`, false in the defaults and in the template). The seam of M2 now defaults to the real reader. `moai integration status` prints the gate line (text and a `gate` key in `--json`). The negative-control fixture uses a temporary repository with a detached stand-in primary and never touches the real primary checkout.

Exit: AC-LVM-007, 017.

### M4 — lane guidance

Hand-edited: `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (template first) then `.claude/rules/moai/workflow/factory-dispatch-mechanics.md`, stating the gate-state source and its `gate_state_unreadable` notice, the no-backfill-after-audit rule, and the by-path binary and receipt-server prerequisites; the sync-auditor protocol `internal/template/templates/.claude/agents/moai/sync-auditor.md` (the neutral source layer, AGENTS.local.md §2.0 copy C2) and `.claude/agents/moai/sync-auditor.md` (the local copy C1, kept byte-identical to C2 today), and `.moai/docs/audit-artifact-convention.md` (local-only, no mirror), stating the `audit_receipt machine line`. Generated, never hand-edited: after the C2 edit run `make agents-emit` (one Go call, subject to rule 8c1e), which rewrites `internal/template/templates/.codex/agents/moai/sync-auditor.toml` from C2 (the emitter reads the `.md` layer and writes only the template `.toml` tree, `internal/template/agentemit/golden_test.go`; the mapping manifest `agents-codex.yaml` is not edited); copy the regenerated file over the local `.codex/agents/moai/sync-auditor.toml` (the two are byte-identical today). `make build` then runs `agents-emit-check` green. Includes the recovery paragraph (the key-off one-path card through the gated path).

Exit: AC-LVM-012, 018.

## D. Ordering, machine-checkable

M0 < M1 < M2 < M3 < M4 in the commit graph (AC-LVM-012); the characterization test file is in the M0 commit and absent from the commit that first edits `factory_card.go`; the branch's `workflow.yaml` history is empty (no enabling); the enabling commit is a later, separate commit on main made only after this SPEC and SPEC-LANDING-COMPLETENESS-GATE-001 have both landed and a by-path binary built from a HEAD containing both is in use; its post-landing witness is in AC-LVM-012. Cards already merge-ready are not re-audited for the switch; each takes the HEAD-confirmation audit at landing, which produces the machine line and a bound receipt.

## E. Per-milestone method

Each milestone: (1) write the AC's tests first and observe them red for the stated reason, keeping the RED output as milestone evidence per the TDD result contract; (2) implement the minimum; (3) refactor; (4) run the milestone's AC commands with the empty-sweep guard; (5) one commit.

- **M1 specifics.** The writer is `recordAuditReceiptAt` (`internal/cli/mcp_audit_receipt.go:60-100`); the schema is `auditreceipt.Receipt` (`store.go:166-177`). `WriteReceipt` (`store.go:944`) assigns only the id and the creation time, so the new fields pass through. The decode is `json.Unmarshal` (`store.go:1389`), and the package uses no `DisallowUnknownFields`, so a legacy file decodes with the fields empty. The hook-side guard (`CheckCitedReceiptsSince`, `store.go:1257-1323`) keeps its meaning; the new fields are consulted only by the landing binding.
- **M2 specifics.** Receipts are resolved from the store root the audit call used (`receiptTarget`); the `audit_receipt:` line is parsed from the verdict file by new code (the existing guard parses the final-message `AUDIT-VERDICT:` line and never a file); the verdict locator reuses the stage chain's newest-iteration rule (`homestate.newestVerdictFile`, `card_evidence_readers.go:244-271`) from its new home. The SPEC-less locator reads the fixed file `.moai/reports/<card-id>/verdict.md` as `merge_specless.go:43` does today. The github-flow edge returns before the binding (`factory_card.go:2072`).
- **M3 specifics.** The configuration precedent is `Workflow.CandidateCI` read by `candidateCIEnabled` (`integration_merge.go:210-216`); the integration-branch precedent is `configuredIntegrationBranch` (`integration.go:291`) and the window record read at `integration_merge.go:55-66`; for `complete` the branch is the one it already merges into (`factory_card.go:2131-2139`).

## F. Risks

| Risk | Mitigation |
|---|---|
| Enabling (after both pieces have landed) refuses merge-ready cards until their audit is re-run with the conforming server | decided cost (board 0577 P2); M4 guidance precedes the switch; the leader's pre-switch list is in `research.md` §10 (6 of 8 sampled verdicts lack the `audit_receipt:` line and are re-audited at landing anyway) |
| The gate on before the receipt-writing server runs the M1 build refuses every landing | fail-closed and visible (the refusal names an unbound receipt); prerequisite C5 puts the by-path binary and the server check before the switch; the refusal precedes the window, so nothing wedges |
| An unreadable gate key could wedge or silently disable the gate | an unreadable present file reads on with a notice and still admits a compliant landing, including the repairing card; an unconfigured project reads off; a configured-but-missing branch refuses with a named error and needs no landing to repair (REQ-LVM-017, board 5d4d) |
| The gate on in a build that lacks it is silently inert | the switch is gated on the by-path binary prerequisite (C5) |
| The direct merge verb and merge-readiness condition (a) do not apply the binding (decision Q28 default) | a lane that skips `complete` is not bound; accepted residual R6 for A2 because the key stays off until A1 lands; the binding on the direct merge path (`merge.go:174`, `:245`, `integration_merge_step.go:321-325`) is an acceptance requirement of A1 (decision Q28, board 6404), so the gate is never on with the side door |
| A same-named tag shadows the integration branch name and turns the gate off | the reader resolves `refs/heads/<branch>^{commit}` once and uses the SHA (REQ-LVM-017; `TestSameNamedTagDoesNotShadowBranch`; probe in `research.md` §13) |
| Both pieces edit `factoryCompleteCard` | A1 is blocked until this card is merged-local (board d20e); this SPEC's edit is one call between `factory_card.go:2178` and `:2185` |
| Plan audit holds at the Tier L ceiling | the audit cycle restarts on the reduced scope (board d20e); disposition in `progress.md` |

Run-phase debt (recorded, not planned work): (1) confirm with one `make build` that `embed-manifest-check` and `gen-catalog-hashes` need no regeneration after the template edits (the embed manifest lists file names only, `internal/template/embed_manifest_gen.go:52`, and no file is added, so none is expected; not measured); (2) the AC-LVM-017 fixture uses a quoted `"yes"` and `1`; whether an unquoted YAML `yes` decodes into the boolean target was not measured; (3) the overlap-baseline script defects of `research.md` §11 (no requirement depends on its numbers).

## G. Cross-references

`spec.md`, `acceptance.md` (RED-now ledger), `design.md`, `research.md` (§12 independence proof), `decision-index.md`. Rules touched at M4: `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing and its template mirror.

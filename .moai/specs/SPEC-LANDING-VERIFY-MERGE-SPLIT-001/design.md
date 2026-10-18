# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — design (plan phase, piece A, v0.5.0)

Design intent only; identifiers are proposals, the observable contracts are `spec.md` §3 and `acceptance.md`. Code evidence: `research.md`. Scope per boards `9dbe`, `0577`, `470c` and `5d4d`.

## 1. Shape

| Job | Home |
|---|---|
| Which verification calls were required | the manifest, the stored result of a shared derivation (REQ-LVM-001, 014, 016) |
| What each execution observed and whether it ended | attempt records via `moai integration remeasure --call` (REQ-LVM-002/003) |
| Everything required passed, on this tree, by the required command, against a current manifest | the completeness verifier (REQ-LVM-004..006, 016), surfaced as `manifest verify` |
| A card with no Go change is neither wedged nor waved through | empty manifest plus derivation evidence (REQ-LVM-013) |
| The audit describes this HEAD | verdict locator, citation line, bound receipts (REQ-LVM-008..010, 018) |
| Whether the gate is on, and since when | the committed blob of the integration branch (REQ-LVM-017); the enabling is a later separate commit (REQ-LVM-012) |

The merge step remains the only writer of the integration branch. New records are machine-local state under the shared integration root's `.moai/state/remeasure/`.

## 2. Records (field sets)

- **Manifest** (`…/remeasure/manifests/<card>.json`): card id; base and card HEAD at derivation; `calls[]` with `call_id`, `command` (exact argv), `packages[]` (resolved by `go list` — the denominator for "zero missing"), `reason` (`changed` | `importer-of:<pkg>` | `embed-owner-of:<pkg>` | `non-go-reader:<path>`), `narrowing` (absent, or `35e9` and permitted only for `internal/cli`), `no_test_files` per package; `non_go_inputs[]` each with reader calls or a zero-readers record; `deleted_packages[]` with the graph comparison; `env_fingerprint`; for a zero-call manifest the derivation evidence (`changed_paths[]`, the `go list` proof, the reader-search result).
- **Attempt** (`…/remeasure/attempts/<card>/<attempt_id>.json`): the fields of REQ-LVM-002; `termination` ∈ `completed` | `aborted:signal` | `aborted:timeout` | `aborted:partial-json`.
- **Receipt** (`auditreceipt.Receipt`, additive): `audited_commit`, full 40-hex or empty. A legacy file decodes with the field empty.
- The existing `RemeasureRecord` (`integration_remeasure.go:45-62`) keeps its meaning and is judged exactly as today while the gate is not on.

## 3. Derivation is shared; currency is re-derivation (N1)

The producer and the verifier call the same derivation function over the candidate. `manifest derive` stores its result; every consumer (merge step, `complete`, condition (d), `manifest verify`) re-runs it and compares: every derived call must be present with the same command and package set, and the non-Go input records and graph comparisons must be equal. A manifest derived at an earlier HEAD that omits a package changed since, or a reader added since, is therefore refused (class 13) even though every attempt is green; a later commit that leaves the derived set unchanged does not invalidate the manifest, so lanes are not forced to re-derive after harmless commits. The zero-call evidence block is the special case of the same comparison, which also removes the earlier asymmetry between the empty and non-empty paths. The reader-search result and the `go list` proof are recomputed, not trusted. The re-derivation runs under the environment recorded in the stored manifest (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, tags, toolchain), never the verifier process's own, so a lane and a CI verifier agree for one HEAD and tree; `TestReDerivationUsesManifestEnvironmentNotVerifierEnv` pins it. Build-tag-dependent package activation is otherwise residual R1.

## 4. Completeness is observed; command binding; what derivation can miss

`go test -json` prints one terminal event per package. The gate compares the `go list`-resolved package set with the terminal events seen and counts executed test events per package: a killed run, a truncated `tee`, a build failure before any package event, a selector matching nothing, or a `TestMain` that exits 0 each leave a package without a terminal event or with zero tests, so the attempt does not pass.

**Command binding (class 12).** The attempt records its exact argv; the verifier compares the `go test` segment with the manifest call's argv after dropping leading environment assignments and an `unset … &&` scrub segment (the repository's standard env-scrubbed form, AGENTS.md §4). Any other segment, pipe or compound refuses. A `-run` is therefore accepted only where the manifest entry carries it, and the manifest may carry it only for `internal/cli` naming 35e9. The existing shell-segment helpers (`shellSegments`, `goSubcommand`, `stripLeadingEnv`, `integration_remeasure.go:236-420`) are the parsing base.

**Script-wrapped calls.** The verb classifies only a `go test -json` command (`integration_remeasure.go:194-209`); a script emits no recognized stream, so with the gate on it cannot satisfy a call. Lanes run each call in the structured form. The stdout-swallow defect stays out of scope.

| Hazard | Handling |
|---|---|
| Changed non-Go file outside a package directory that a test reads | reader search recomputed on every verification; a missing reader call and a missing zero-readers record both refuse. A test that computes the path at run time is not found — residual R1 |
| Embedded or testdata file | a file under a package directory marks that package changed; a package that embeds a file from another package's directory is derived from `go list` embed fields (`EmbedFiles`, `TestEmbedFiles`, `XTestEmbedFiles`) and gets a call |
| Deleted package | the base and candidate graphs are compared and recorded; importers of the deleted package become calls |
| Build tags / OS-specific files | fingerprint declared in the manifest and equal in each attempt; a difference refuses |
| Environment-dependent tests | declared env names with value digests are part of the fingerprint; an undeclared read is not detectable — residual R2 |
| Generated files; tests that read git history or `.moai/` state | not detectable; an attempt requires a clean tree at start and finish (`RunRemeasure`); residual R3, bounded by batch CI |
| `go list` failure | scope underivable → refuse |
| Hand-edited records | the existing trust model (build-identity field) is unchanged — an accepted residual |
| Test-only changes | importers still included (decision-index Q5); no test is ever excluded |
| A dirty working tree at audit time | `audited_commit` records HEAD only; the writer cannot see uncommitted edits — accepted residual R4 |
| A partial-scope audit stamped with the current HEAD | `audit_target` is recorded for review but not gated — accepted residual R5 (Q23) |

Abort versus restart is a property of the attempt: signal and timeout come from the child's wait status and the verb's own deadline; a cut stream from a final line that is not a complete JSON object. A restart is a new execution with a new attempt ID.

## 5. Zero-Go cards (REQ-LVM-013)

A card that changes only documentation, rules, templates or configuration still gets a manifest. The derivation runs the reader search for each changed non-Go file; any reader found becomes a call (a template file read by `internal/template` tests produces a call for that package, as lanes do today). When there are none, the manifest is empty and carries the derivation evidence; the verifier admits only when re-derivation reproduces it. The audit binding applies as for any card. A key-off card is not of this kind in this repository: tests read `.moai/config/sections/workflow.yaml` (for example `internal/config/workflow_quota_gate_test.go`, subtest `local_twin_enabled`), so its derived manifest holds reader calls (REQ-LVM-012).

## 6. The audit and receipt binding (N3, N4, O1, O4)

Facts read from the tree: the SPEC-less readiness path binds `audited_sha` to HEAD tolerating a difference only under the card's evidence directory (`factorylane/merge_specless.go:35, 108-138`); the SPEC path reads only `sync_status` (`factorylane/merge.go` `checkSyncAudit`); the stage chain reads the newest `<phase>*.md` verdict through `readAuditVerdict`/`newestVerdictFile` (highest `-iter<N>`, then latest modification) and admits it through `auditverdict.AdmitWithRequired` with `runtime.ResolveRequiredBackends` (`homestate/card_evidence_readers.go:84-260`); `auditverdict.Parse` has no `audit_receipt` key and no Go code reads receipt ids from a file; `ParseVerdictLine` (`store.go:1223`) parses the final-message `AUDIT-VERDICT:` line; `CheckCitedReceiptsSince` (`store.go:1248-1323`) checks existence, tree, start time and codex tool at verdict-write time and refuses an empty citation; measured verdicts of merge-ready SPEC cards cite receipts in mixed forms (2 of 8 with a top-level `audit_receipt:` line).

Decisions (decision-index Q14, Q19, Q21, Q22):

- **Locator.** A SPEC card: the newest `sync-audit*.md` under `.moai/reports/<card-id>/`, by the stage chain's rule. A SPEC-less card (classes A and B): the fixed file `.moai/reports/<card-id>/verdict.md` that its readiness path already reads. The receipt binding and the citation line apply to both kinds (otherwise Class A and B cards would carry no receipt binding while Q13 binds `complete`); a SPEC card with only `verdict.md`, or any card with no file, is refused while the gate is on.
- **Citation line.** The verdict must carry exactly one machine line `audit_receipt: rcpt-<id>[,rcpt-<id>…]`; ids match `rcpt-[a-z0-9]{20,40}`; a missing, duplicated, empty, `none` or malformed line, or an id absent from the store, refuses. This closes the vacuous pass of "every cited receipt" for an empty citation. The sync-auditor protocol and the audit artifact convention state the line.
- **Qualification.** At least one cited receipt must qualify. It is looked up at `auditreceipt.StoreRoot(audited tree)` — for a config-orphaned worktree that is the primary checkout (`storeroot.go:242-251`) — and must be readable; tool codex or multi-backend; **`tree_root` equal to the normalized path of the audited tree itself**, which is what the writer records (`TreeRoot: root`, `mcp_audit_receipt.go:63-73`; the existing guard compares the same quantity, `store.go:1290`); unmet gate empty; codex verdict pass; `audited_commit` equal to HEAD. The store root is where the file lives, not the compared identity: comparing `tree_root` with the store root would refuse every valid receipt of an orphaned worktree and accept the primary's receipt for the same commit. Failed or inconclusive extra citations are recorded and never replace the pass; the verdict itself is still admitted by `AdmitWithRequired`.
- **Equals HEAD, nothing else.** The verdict's audited commit and the receipt's audited commit must each equal the card's HEAD. An ancestor whose diff to HEAD is empty, patch-identical or confined to `.moai/reports/<card-id>/` does not qualify (boards 0577 P3 and 5d4d). The earlier evidence-directory tolerance is removed: `.moai/reports/*` is ignored by git in this repository (`.gitignore:235`), so it had almost nothing to tolerate, and it contradicted the no-backfill ruling. The lane guidance moves bookkeeping before the final audit.
- **Receipt stamping.** The two files that call the writer (`mcp_codex.go` at :2112 and :2163, `mcp_convergence.go` at :928) capture the audited tree's HEAD and the review target at entry and pass both to `recordAuditReceiptAt`; the writer re-reads HEAD when writing and leaves `audited_commit` empty if it moved. Existing receipts are never rewritten and are read as unbound.
- **Audit target (decision-index Q23).** New receipts also record `audit_target`, the review target the call named (for example the base-branch review versus uncommitted changes), so a reviewer can see what scope the receipt covers. It is recorded, not gated: the set of legitimate full-scope target values is not enumerated at plan time, and gating on a guess could refuse valid audits. A partial-scope audit stamped with the current HEAD is accepted residual R5, beside the dirty-tree residual R4.

## 7. Gate state, switch, binary and recovery (N2, board 470c)

- **Source.** The committed blob `<integration-branch>:.moai/config/sections/workflow.yaml`, read in the repository, consulting no working tree. The branch is the window record's when a window is held, else `configuredIntegrationBranch()` (`internal/cli/integration.go:291`). The merge verb's further fallback to `develop` for an empty name is not used for the gate: `LoadGitFlowIntegrationConfig` returns the zero configuration when git-strategy is absent or has no profile (`config/loader_integration_branch.go:95-97`), and a project that never opted in to the window must keep the gate off. Measured why not a working tree: the primary checkout is a stale detached HEAD that an enabling commit on `main` never reaches, and `integrationLockRoot()` returns it (`integration.go:34-72`); reading the main worktree would tie correctness to one local layout.
- **Fail-modes (board 5d4d).** No integration branch configured and no window held → off (the project has not opted in). A branch configured whose ref cannot be resolved → the landing is refused with the named error `gate_state_integration_branch_unresolvable`, fail-closed; this is not a deadlock, because the merge step cannot run without that branch and repairing the branch configuration or ref needs no landing. With the branch resolved: `git cat-file -e <branch>:<path>` separates an absent path (off) from a failed read; an absent key or explicit false → off; explicit true → on; a failed read of a present path, an unparseable file or a non-boolean value → on with `gate_state_unreadable` and an audit row. On is the safe state for those: it never silently disables the gate, and it does not wedge landings, because a landing that satisfies the gate (structured calls, bound audit) is admitted, including the card that repairs the file.
- **Observation.** AC-LVM-017 uses a temporary repository whose detached stand-in primary never receives the commits; the enable commit takes effect when it lands on the integration branch and the revert likewise. `manifest verify` prints the gate state with its source branch, path and commit so the effect is observable in the field.
- **Order.** This branch never enables the key. The enabling is a later, separate commit on the integration branch after the SPEC has landed and a binary built from the landing HEAD exists; the first gated landing is the next one. Cards already merge-ready are not re-audited for the switch.
- **Binary and server (prerequisites, not steps of this card).** The governed commands are invoked by path from a binary built from the landing HEAD's main; the receipt-writing MCP server must run a build containing the writer change before a receipt is bound. Neither implies reinstalling the installed binary.
- **Recovery.** The reversal is a one-path card setting the key false. Its derived manifest holds the reader calls of the configuration file; it lands through the gated path with the audit binding intact. Residual: if the gated path itself is the defect, the operator-held manual edit of the integration branch's configuration is the accepted fallback; an agent does not perform it.

## 8. Tier arithmetic

See `spec.md` §6: analogue 1,734 lines; estimate ≈ 1,400-1,750 production and 2,100-2,600 test lines; about 41 files including (the `manifest` subcommand registration in `cli/integration.go` among them) the receipt writer, schema, two audit entry points, the citation parser, the gate-state reader and the protocol text. Tier L; 18 requirements and 18 criteria.

## 9. Follow-up pieces (not designed here)

A slot piece (dead-session record removal only), a readiness-and-metering piece depending on A, and the refused reuse relaxation. Their facts are in `research.md` §4-§6 and `decision-index.md`. The overlap baseline of `research.md` §6 comes from a script with two known defects (recorded in `research.md` §11); it is quoted only as an order of magnitude for the refused reuse and must be re-measured before it is cited elsewhere.

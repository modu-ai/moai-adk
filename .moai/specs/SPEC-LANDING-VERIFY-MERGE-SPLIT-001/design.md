# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — design (plan phase, piece A, v0.3.0)

Design intent only; identifiers are proposals, the observable contracts are `spec.md` §3 and `acceptance.md`. Code evidence: `research.md`. Scope per boards `9dbe` and `0577`.

## 1. Shape

| Job | Home |
|---|---|
| Which verification calls were required | the manifest, produced by `moai integration manifest derive` (REQ-LVM-001, 014) |
| What each execution observed and whether it ended | attempt records via `moai integration remeasure --call` (REQ-LVM-002/003) |
| Everything required passed, on this tree, with the required command | the completeness verifier (REQ-LVM-004..006), surfaced as `manifest verify` |
| A card with no Go change is neither wedged nor waved through | empty-with-evidence manifest (REQ-LVM-013) |
| The audit describes this HEAD | verdict binding (REQ-LVM-008) and bound receipts (REQ-LVM-009/010) |
| When the gate becomes mandatory | a later separate commit on the integration branch (REQ-LVM-012, spec C5) |

The merge step remains the only writer of the integration branch. New records are machine-local state under the integration root's `.moai/state/remeasure/`.

## 2. Records (field sets)

- **Manifest** (`…/remeasure/manifests/<card>.json`, written by `manifest derive`, regenerated after every re-absorb): card id; base and card HEAD at derivation; `calls[]` with `call_id`, `command` (exact argv), `packages[]` (resolved by `go list` at derivation — the denominator for "zero missing"), `reason` (`changed` | `importer-of:<pkg>` | `embed-owner-of:<pkg>` | `non-go-reader:<path>`), `narrowing` (absent, or `35e9` and permitted only for `internal/cli`), `no_test_files` per package; `non_go_inputs[]` each with reader calls or a zero-readers record (search command and empty result); `deleted_packages[]` with the base-versus-candidate graph comparison; `env_fingerprint`; and, for a zero-call manifest, the evidence block (`changed_paths[]`, the `go list` proof, the reader-search result).
- **Attempt** (`…/remeasure/attempts/<card>/<attempt_id>.json`): the fields of REQ-LVM-002 including per-package executed-test counts and the exact argv; `termination` ∈ `completed` | `aborted:signal` | `aborted:timeout` | `aborted:partial-json`.
- **Receipt** (`auditreceipt.Receipt`, additive): `audited_commit` — full 40-hex, or empty. Existing fields unchanged; a legacy file decodes with the field empty.
- The existing `RemeasureRecord` (`integration_remeasure.go:45-62`) keeps its meaning and is judged exactly as today while the key is false.

## 3. Completeness is observed; command binding; what derivation can miss

`go test -json` prints one terminal event per package. The gate compares the manifest's `go list`-resolved package set with the terminal events seen, and counts executed test events per package: a killed run, a truncated `tee`, a build failure before any package event, a selector matching nothing, or a `TestMain` that exits 0 each leave a package without a terminal event or with zero tests, so the attempt does not pass.

**Command binding (REQ-LVM-005 class 12).** The attempt records its exact argv; the verifier compares the `go test` segment with the manifest call's argv after dropping leading environment assignments and an `unset … &&` scrub segment (the repository's standard env-scrubbed form, AGENTS.md §4). Any other segment, pipe or compound refuses, because it cannot be classified. A `-run` is therefore accepted only where the manifest entry itself carries it, and the manifest may carry it only for `internal/cli` naming 35e9. The existing shell-segment helpers (`shellSegments`, `goSubcommand`, `stripLeadingEnv`, `integration_remeasure.go:236-420`) are the parsing base.

**Script-wrapped calls.** The verb classifies only a `go test -json` command (`integration_remeasure.go:194-209`); a script emits no recognized stream, so with the gate on it cannot satisfy a call. Lanes run each call in the structured form. The stdout-swallow defect stays out of scope.

What derivation can miss, and the handling:

| Hazard | Handling |
|---|---|
| Changed non-Go file outside a package directory that a test reads | reader search recorded in the manifest; absence of both a reader call and a zero-readers record refuses. A test that computes the path at run time is not found — residual R1 |
| Embedded or testdata file | a file under a package directory marks that package changed; a package that embeds a file from another package's directory is derived from `go list` embed fields (`EmbedFiles`, `TestEmbedFiles`, `XTestEmbedFiles`) and gets a call |
| Deleted package | `go list` on the candidate cannot name it; the base and candidate graphs are compared and recorded, and importers of the deleted package become calls |
| Build tags / OS-specific files | fingerprint (GOOS, GOARCH, tags, CGO) declared in the manifest and equal in each attempt; a difference refuses |
| Environment-dependent tests | declared env names with value digests are part of the fingerprint; an undeclared read is not detectable — residual R2 |
| Generated files | the committed tree is what is measured; an attempt requires a clean tree at start and finish (`RunRemeasure`), so ignored generated files are not part of the claim — residual R3 |
| Tests that read git history or `.moai/` state | not detectable by search of imports or paths; same residual R3, bounded by batch CI |
| `go list` failure | scope underivable → refuse |
| Importer lists changing after a re-absorb | the manifest is regenerated; attempts bound to the superseded identity stop counting |
| Hand-edited records | the existing trust model (build-identity field, `integration_remeasure.go` header) is unchanged — an accepted residual |
| Test-only changes | importers still included (decision-index Q5); no test is ever excluded |

Abort versus restart is a property of the attempt: signal and timeout come from the child's wait status and the verb's own deadline; a cut stream from a final line that is not a complete JSON object. A restart is a new execution with a new attempt ID.

## 4. Zero-Go cards (REQ-LVM-013)

A card that changes only documentation, rules, templates or configuration still gets a manifest. The producer runs the reader search for each changed non-Go file; any reader found becomes a call (a template file read by `internal/template` tests produces a call for that package, as lanes do today). When there are none, the manifest is empty and carries the evidence block. The verifier recomputes the changed-path list from the repository (diff against the merge-base) and admits only on equality, so a path added by a later commit invalidates the block. The audit binding applies as for any card. This is the path the key-off recovery card takes.

## 5. The audit and receipt binding

Facts read from the tree: the SPEC-less readiness path binds `audited_sha` to HEAD tolerating a difference only under the card's evidence directory (`factorylane/merge_specless.go:35, 108-138`); the SPEC path reads only `sync_status` (`factorylane/merge.go` `checkSyncAudit`); stage-chain admission of a verdict goes through `auditverdict.AdmitWithRequired` and `runtime.ResolveRequiredBackends` (`homestate/card_evidence_readers.go:189-221`, `runtime/audit_gates.go:43`); receipts are validated for existence, tree, start time and codex tool at verdict-write time by `auditreceipt.CheckCitedReceiptsSince` (`store.go:1248-1323`), which never consults a commit; the writer `recordAuditReceiptAt` sets none (`cli/mcp_audit_receipt.go:60-100`).

Decision (board 0577 P4, decision-index Q14): add `audited_commit` to NEW receipts at write time. The writer reads the audited tree's HEAD before the backends run and re-reads it when writing; if they differ the field is left empty (unbound). `complete`, under the gate, reads the cited ids from the verdict file, resolves each with `ReadReceipt`, and requires `audited_commit` equal to the card's HEAD. Existing receipts are never rewritten or deleted and are read as unbound. Rejected alternative: bind only through the verdict's `audited_sha` — it would not satisfy "receipt at that HEAD" (2c51).

Post-audit commits (P3): the only tolerated difference between the audited commit and HEAD is a path under `.moai/reports/<card-id>/`. A progress or evidence backfill after the audited SHA refuses (AC-LVM-008 c). The sync flow is adjusted by the lane guidance, not by an exemption: bookkeeping is committed before the final audit.

## 6. Switch, binary and recovery

- **State source.** The key is read from the integration root's configuration (pattern of `candidateCIEnabled`, `integration_merge.go:210-216`); a card branch's copy has no effect (REQ-LVM-007).
- **Order.** This branch never enables the key. The enabling is a later, separate commit on the integration branch after the SPEC has landed and a binary built from the landing HEAD exists. The first gated landing is the one after that commit. Cards already merge-ready are not re-audited for the switch (they take the HEAD-confirmation audit at landing anyway); once the key is on their re-measure is done in the structured form.
- **Binary and server (prerequisites, not steps of this card).** The governed commands are invoked by path from a binary built from the landing HEAD's main, never a PATH-resolved installed build that may lag; the receipt-writing MCP server must run a build containing the writer change before a receipt is bound. Neither implies reinstalling the installed binary.
- **Recovery.** The gate's reversal is a one-path card that sets the key false: it lands through the ordinary chain by the zero-Go path with the audit binding intact, so no audit is bypassed. Residual: if the zero-Go path itself is the defect, the operator-held manual edit of the integration root's configuration is the accepted fallback; an agent does not perform it.

## 7. Tier arithmetic

See `spec.md` §6: analogue 1,734 lines; estimate ≈ 1,150-1,450 production and 1,700-2,100 test lines; about 33 files including the receipt writer, schema and their tests. Tier L; 15 requirements and 15 criteria.

## 8. Follow-up pieces (not designed here)

A slot piece (dead-session record removal only), a readiness-and-metering piece depending on A, and the refused reuse relaxation. Their facts are in `research.md` §4-§6 and `decision-index.md`.

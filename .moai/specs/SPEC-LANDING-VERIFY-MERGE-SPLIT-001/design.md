# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — design (plan phase, piece A)

Design intent only; identifiers are proposals for the run phase, the observable contracts are in `spec.md` §3 and `acceptance.md`. Code evidence for current behaviour: `research.md`. Scope is piece A per board `9dbe`.

## 1. Shape

Today one record carries "a command passed" (`integration_remeasure.go:135`). Piece A adds the evidence layer beside it and consults it where the record is consulted:

| Job | Home |
|---|---|
| Which verification calls were required | the manifest (REQ-LVM-001) |
| What each execution observed, and whether it ended | attempt records with package terminal events (REQ-LVM-002/003) |
| Everything required passed, on this tree | the completeness verifier (REQ-LVM-004..006) |
| The audit describes this HEAD | the verdict and receipt binding in `complete` (REQ-LVM-008/009) |
| When the gate becomes mandatory here | the ordered switch (REQ-LVM-011) |

The merge step remains the only writer of the integration branch (REQ-LVM-012). New records are machine-local state beside `.moai/state/remeasure/`.

## 2. Records (field sets)

- **Manifest** (`…/remeasure/manifests/<card>.json`): card id; base and card HEAD at derivation; `calls[]` with `call_id`, `command`, `packages[]` (resolved by `go list` at derivation — the denominator for "zero missing"), `reason` (`changed` | `importer-of:<pkg>` | `non-go-reader:<path>`), `narrowing` (absent, or `35e9`, permitted only for `internal/cli`), `no_test_files` per package; `non_go_inputs[]` each with either reader calls or a zero-readers record holding the search command and its empty result; `env_fingerprint` (GOOS, GOARCH, CGO_ENABLED, GOFLAGS, `go version`, tags, declared env names with value digests); `no_go_changes`.
- **Attempt** (`…/remeasure/attempts/<card>/<attempt_id>.json`): the fields of REQ-LVM-002; `termination` ∈ `completed` | `aborted:signal` | `aborted:timeout` | `aborted:partial-json`; `supersedes` (informational; never erases the aborted record).
- The existing `RemeasureRecord` (`integration_remeasure.go:45-62`) keeps its meaning; with the key false it is judged exactly as today (REQ-LVM-007). Whether attempt data is embedded in that struct or stored beside it is a run-phase choice constrained by AC-LVM-002.

## 3. Package completeness is observed; what the manifest can miss

`go test -json` prints one terminal event per package. The gate compares the manifest's `go list`-resolved package set with the terminal events seen: a killed run, a truncated `tee`, or a build failure before any package event leaves a package without a terminal event, so the attempt is aborted or incomplete, never a pass. A script-wrapped call emits no recognized stream to the verb (`integration_remeasure.go:194-209` classifies only a `go test -json` command), so with the gate on it cannot satisfy a call; lanes run each call in the structured form. The known stdout-swallow defect stays out of scope.

What derivation can miss, and the handling:

| Hazard | Handling |
|---|---|
| Changed non-Go file outside a package directory that a test reads | reader search over test sources recorded in the manifest (readers become calls; zero readers is an explicit recorded result); absence of both refuses (REQ-LVM-005). A test that computes the path at run time is not found by search — residual risk R1, bounded by batch CI |
| Embedded or testdata file | inside a package directory, so the package counts as changed |
| Build tags / OS-specific files | the environment fingerprint (GOOS, GOARCH, tags, CGO) is declared in the manifest and must equal each attempt's; a difference refuses |
| Environment-dependent tests | declared env names with value digests are part of the fingerprint; an undeclared env read is not detectable — residual risk R2 |
| `go list` failure | scope underivable → refuse (fail-closed) |
| Importer lists changing after a re-absorb | the manifest is regenerated; attempts bound to the superseded manifest stop counting |
| Hand-edited records | the existing trust model (build-identity field, `integration_remeasure.go` header) is unchanged — an accepted residual, not a defended boundary |
| Test-only changes | importers are still included (decision-index Q5); no test is ever excluded (REQ-LVM-012) |

Abort versus restart is a property of the attempt: signal and timeout come from the child's wait status and the verb's own deadline; a cut stream from a final line that is not a complete JSON object. A restart is a new execution with a new attempt ID; the verdict counts the most recent completed attempt per call on the candidate tree.

## 4. The audit binding (Q13) — facts and a gap

Facts read from the tree: the SPEC-less readiness path binds `audited_sha` to HEAD, tolerating a difference only under the card's evidence directory (`factorylane/merge_specless.go:35, 108-138`); the SPEC path reads only `sync_status` from progress (`factorylane/merge.go` `checkSyncAudit`); stage evidence readers consult required backends (`homestate/card_evidence_readers.go:189-221`, `runtime.ResolveRequiredBackends`, `runtime/audit_gates.go:43`); receipts live in `internal/auditreceipt`. Piece A applies the same HEAD bind to the SPEC path and requires the required backend's receipt at that HEAD.

Gap, to be closed at the start of M2 before code: this plan phase did not read how `auditreceipt` keys a receipt to a commit (by audited SHA, by tree, or by card). If the key is not a commit, REQ-LVM-009's wording is amended through manager-spec. The evidence-directory allowance is kept because a verdict file cannot name the commit that adds it; it admits no content change, so it does not let an earlier PASS stand for a later HEAD.

## 5. The switch and its transition cost

One key, `workflow.landing_gate.enabled`, governs the completeness verifier and the audit binding; it ships false in the template and is switched on in this repository's configuration by the last milestone only. Cards already `merge-ready` at that moment have script-wrapped records and are refused until their calls are re-run in the structured form; M3's guidance precedes the switch for that reason. The card's own landing is the first live use.

## 6. Tier arithmetic

See `spec.md` §6: nearest analogue 1,734 lines (record module, CLI and tests); estimate ≈ 1,000-1,300 production and 1,400-1,800 test lines; about 29 files. Tier L by LOC and files, 12 requirements and 12 criteria.

## 7. Follow-up pieces (not designed here)

A slot piece (dead-session record removal only), a readiness-and-metering piece (derived view, events, shadow counter) depending on A, and the refused reuse relaxation. Their scope, measurements and open facts are in `research.md` §4-§6 and `decision-index.md`; nothing in them is specified by this SPEC.

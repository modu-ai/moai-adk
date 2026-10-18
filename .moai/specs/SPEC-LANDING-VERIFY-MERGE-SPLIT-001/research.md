# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — research (plan phase, card t1872)

Measurement tree: `ec164d45d2e3c5a7902aa950ab212e8f385f5d08` (current main at authoring). Tool build: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`, built with `go build -o <that path> ./cmd/moai` from the same clean tree (build exit 0), so the judging build and the tree coincide (verification-claim-integrity §2.2). Go toolchain `go1.26.8 darwin/arm64`.

> **v0.7.0 note (split, board d20e).** §1-§11 were gathered for the undivided SPEC and serve both halves; the completeness-gate facts (§2, §3, §6, §8) are inherited by SPEC-LANDING-COMPLETENESS-GATE-001 and are not re-stated for A2. §12 is the independence proof of piece A2 by code and is the only new section.

This file is the plan-phase baseline. It is committed in its own commit, ahead of the spec/plan/acceptance commit (§2.3 ordering), and carries no implementation.

## 1. Which landing path this repository runs

- `.moai/config/sections/git-strategy.yaml:9` — `workflow: git-flow`, `develop_branch: main`; integration merges land locally on `main` ("Merge WT-… into main (card tNNNN, integration merge)" in `git log --first-parent`). Therefore the window path of `moai factory complete` applies, not the github-flow pull-request path.
- `internal/cli/factory_card.go:2071-2073` — `if factoryGitHubFlow(root) { return factoryCompleteGitHubFlow(...) }` bypasses the window and the re-measure record entirely. The distributed default is github-flow; **everything in this SPEC concerns the git-flow window path only** (the github-flow edge is untouched).

## 2. Where "base differs from main tip" is enforced (file:line)

| Place | Evidence |
|---|---|
| Merge step refuses when the record's base is not the tip | `internal/factory/integration_merge_step.go:321-333` — reads the record keyed by the pinned tree (`ReadRemeasureRecord`, :321), validates it (:325), then `if record.Base != tip { … MergeExitBaseMoved … "the integration branch moved past the record's base %s to %s — re-measure against the new tip, then re-acquire --wait" }` (:332-333). Exit code constant `MergeExitBaseMoved = 2` at :42. A second tip re-read inside the lock section at :464 refuses "moved mid-step". |
| `complete` step 3 | `internal/cli/factory_card.go:2281-2283` — no record for the CURRENT candidate tree → exit code `MergeExitBaseMoved`, message "no valid re-measure record for the candidate tree … run moai integration remeasure, then re-acquire --wait". Step 2 (adoption, :2229-2237) accepts an already-merged tip whose tree has a valid record. |
| Tree identity ties the landed tree to the measured tree | `internal/factorylane/merge.go:307` `checkTreeIdentityPreMerge`: `git merge-tree --write-tree <develop> <branch>` must equal the branch tree, i.e. the card branch must already CONTAIN the current tip. This is why a re-absorb is structural: a moved tip changes the merge result. |
| Record store is keyed by tree, base = merge-base | `internal/factory/integration_remeasure.go:45-62` (`RemeasureRecord`), :68-90 (store `.moai/state/remeasure/<tree>.json`), :699 `Base` is `gitMergeBase(worktree, baseBranch)` taken at record time. |
| Single verifier | `internal/factory/integration_remeasure.go:135` `ValidateRemeasureRecord` — consumed by the merge step (:325), `complete` (`factory_card.go:2234`, :2373-2377), `factory_merge.go:113`, and merge-check condition (d) (`internal/factorylane/merge.go` EvaluateMergeTriple :174, (d) block). |

## 3. Premise check of the re-measure record's evidential weight (new finding, leader 14:4x)

`ValidateRemeasureRecord` (`integration_remeasure.go:135-166`) accepts a record when `exit_code == 0` and, for non-`go test` commands, nothing else: line `StructuredRequired: isGoTestCommand(command)` (:698) is false for `sh script.sh`, and the validity rule is then command present + exit 0 (doc comment :185-193: "A tool with no recognized report … is valid on exit code zero and records command + exit only"). A script-wrapped call records no packages, no per-package terminal event, no attempt identity.

Observed real record (read-only): `/Users/goos/MoAI/moai-adk-go/.moai/state/remeasure/0f05ab3e3e2b3032bcbd650c6c4607d52beef109.json` — `"command": "sh …/t1834-gate-rest.sh"`, `"structured_required": false`, `"structured_count": false`, `exit_code 1`. The leader's note records that t1680's landed tree `1a3dc230c35ca0e2a997fc0fc3ac9444eb1f9198` was gated by a single `gate-b.sh` record with `structured_required=false` (`/tmp/adk-queue-repair-20261018/t1872-evidence-note.md`, machine-local). Consequence: nothing mechanical states which calls were required or that all of them passed on the same tree.

Standing scope rules (provenance only; they live on the machine-local decision board, not in the committed tree, so this SPEC restates the derivation it needs in `spec.md` §2): b2cd = "packages the card changed plus the packages importing them", whole package, no `-run`; the single recorded narrowing allowance 35e9 = `internal/cli` only (memory note `feedback_remeasure_scope_is_full_changed_package`). Board ids: standing rule `d-20261018T143825Z-5923`; t1709 ruling `d-20261018T133909Z-07c9`.

## 4. Slot lease (heavy slot) — current behaviour

- `internal/factory/slot_lease.go:127-142` record (one holder, `pid` = owning SESSION pid); :179 `Stale()` (pid ≤ 0 reads live); :189 `Expired`; :314 `AcquireSlotLease`; :353 takeover switch `case current.Stale()` → reason `stale`, then `expired`, then `force`; :404 `ReleaseSlotLease`. `internal/cli/slot.go:131-133` records `session.ResolveOwnerPID()` (the session, not the heavy process).
- Config: `internal/config/types.go:860-864` `SlotLeaseConfig{Enabled, DefaultMaxDuration, Resources}`; the repo's `workflow.yaml` carries no `slot_lease:` block (defaults).
- **Probe (item 4 premise).** Scratch git repo at `…/scratchpad/slotprobe`, a hand-written record with `pid 99999` (dead: `kill -0 99999` → "no such process"), far-future `expires_at`:
  - `moai slot status --resource probe-res` → `held by a session that is gone (reclaimable)`, exit 0; the record file STILL EXISTS afterwards (`ls` shows `probe-res.json`).
  - `moai slot acquire --resource probe-res --session other-session …` (no `--force`) → exit 0, `displaced: dead-session (pid 99999), reason stale`; audit row `{"event":"takeover","reason":"stale",…}`.
  So the card's premise "the record remains until its bound" is **only partly true**: a dead owner is already reclaimable without `--force` at the next acquire; what remains is (a) the record file persists until some verb rewrites it, (b) `moai slot status` and the guard report it but do not clear it, (c) liveness is bound to the SESSION pid — a live lane session whose heavy command died or was abandoned keeps the slot until its declared bound. Also: the integration window already auto-clears a dead holder (`internal/factory/integration_window_ops.go:44-52`, `holderOwnerGone`).
- Live slot now: `moai slot status` → `slot go-test-heavy: held`, lane-27, `t1680 landing remeasure (extend)`, bound 30m0s (read at authoring).
- Capacity: the lease core is strictly single-holder per resource (one JSON file per resource). Capacity 2 needs either a multi-holder record or N sibling records per resource; neither exists. `factory_slots.go` / `factory_slots_capacity_test.go` concern the LANE registry (different object).

## 5. Landing-throughput baseline (measured)

Command: `git log --first-parent --grep='integration merge' --format='%ct' -41 ec164d45d2` → consecutive intervals (minutes), n=40: `min=0.9 median=34.0 p90=158.5 max=460.3 mean=69.1`. `moai factory status` (tree-built binary) lists 16 cards in `state=merge-ready` at authoring (t1626 … t1871). Window state: `release-integration window: free`, policy open.

## 6. Overlap-eligibility baseline for item (1) (measured, approximate)

Method: `evidence/overlap_baseline.py` (committed beside this file), output `evidence/overlap_baseline.out`. Pair model: for each of the last 40 landings, the "waiting card" is landing *i+1* (its changed files = diff from the merge-base of its parents to its second parent), the "moved set" is landing *i* (first-parent diff). Verification scope ≈ directories of `go list -deps -test` closure of the Go package directories the card touched (computed ONCE at the analysis tree, an approximation) ∪ card files ∪ module files. Result (verbatim head of the output):

```
{ "pairs": 40, "decided": 37, "lenient_reusable": 10, "strict_reusable": 1, "undecidable": 3 }
```

- `lenient` ignores moved non-Go files outside the closure; `strict` treats any such file as overlap (the card's "non-Go read" hazard).
- Closure size is large: median 147 module-local package directories for the touched packages (`go list -deps -test -f '{{.Dir}}'` — `./internal/cli` 150, `./internal/factory` 24, `./internal/config` 26 directories; 1.1-2.3 s each).
- 3 of 40 landings touched no Go package and are undecidable by the closure rule (fail-closed → overlap).

**Reading:** under the card's own conservative rule (undecidable = overlap) only ≈3% of adjacent pairs keep a re-measure record; even the unsound lenient reading keeps ≈27%. Adjacent-pair is the most favourable case (one moved landing); with several landings between a record and its landing the cumulative moved set only grows. Item (1) therefore cannot be the main throughput lever on this history; this is carried to the operator as a decision (decision-index Q8) and is why instrumentation (item 5) records a **shadow eligibility counter** before any reuse relaxation opens.

## 7. Tests that exist for the touched surfaces

- Merge step: `internal/factory/integration_merge_step_test.go`, `integration_merge_step_overlap_test.go`, `mergestep_red_t1582_test.go`, `mergestep_atomic_t1582_test.go`; complete refusal: `internal/cli/factory_complete_mwq19_test.go:95` `TestMWQ19_Scenario3_MovedBaseRefusesWithReacquireCode`; remeasure: `internal/factory/integration_remeasure_test.go`, `integration_remeasure_run_test.go`, `remeasure_red_t1582_test.go`, `internal/cli/integration_remeasure_cmd_test.go`; slot: `internal/factory/slot_lease_test.go` (`TestSlotLease_Liveness`:343, `_DeclaredBound`:418, `_ForceRecordsDisplaced`:513), `slot_lease_cross_test.go`.
- Prior SPEC lineage: SPEC-MERGE-WINDOW-QUEUE-001/-002 (record, window, merge verb), SPEC-RESOURCE-SLOT-LEASE-001 and SPEC-HEAVY-TEST-SLOT-001 (slot), SPEC-VERIFY-RUN-REUSE-001 (`moai verify run`: same-tree same-command reuse — different object, cross-tree reuse here).

## 8. CLI-existence probes behave vacuously (RED-now design constraint)

`moai integration manifest|reuse|ready|metrics --help` each print the parent `integration` help and exit 0 (observed), so a help probe cannot be a RED-now cell. RED-now cells in `acceptance.md` use `go test -list` name-count assertions (the SPEC-VERIFY-RUN-REUSE-001 pattern) and `grep` exit codes.

## 9. Gaps

- Closure for §6 was computed at one tree for all pairs; per-pair trees were not checked out.
- The vouch-list (operator-declared never-read non-Go globs) variant of §6 was not measured; it needs the operator's list (decision-index Q8).
- No wall-clock measurement of a 15-25 minute re-measure was taken in this plan phase (figures are the card's, 10-18).
- The decision-board texts quoted in §3 were read through `moai decision read`; they are machine-local, not committed authority.

## 10. Transition picture for the cards that are merge-ready now (added v0.3.0, plan-audit D5)

Method: `evidence/transition_head_vs_verdict.py` (committed beside this file) reads, without any git call, each `merge-ready` card's tree HEAD (from the worktree's `.git` pointer and refs) and the `audited_sha:` line of its newest `sync-audit*.md` / `verdict.md`; output `evidence/transition_head_vs_verdict.out`, captured 2026-10-19 from `moai factory status` (19 card ids; 18 trees found, one absent).

Result (verbatim last line of the output): `cards_with_tree=18 match_current_head=8` — for 8 cards the newest verdict names the current HEAD, for 10 it does not (HEAD moved after the audit: a re-absorb, a progress or evidence commit, or a repair).

Reading, under the leader's ruling (board 0577 P2): none of these cards is re-audited because of the switch. Each takes the HEAD-confirmation audit at landing anyway (board 2c51), which writes a verdict and a receipt at its HEAD; with the writer change running (spec C5) that receipt is bound. What differs once the key is on is only the re-measure, which must be in the structured form, and the requirement that the audit is bound to the final HEAD with no backfill commit after it.

What the leader does before the switch (not the lane): build the landing binary by path from main at the landing HEAD; confirm the receipt-writing MCP server runs a build containing the receipt change (`moai doctor` compares the running server with the binary); tell the lanes the structured form (rule text from milestone M4); commit the one-path key change only after this SPEC and SPEC-LANDING-COMPLETENESS-GATE-001 have both landed (board 6404). Not measured here: how many of the 18 have a script-wrapped re-measure record (records are keyed by tree and the trees were not read), and the wall-clock cost of a structured call.

## 11. Gate-state source and receipt-citation facts (added v0.4.0, plan-audit N2/N3)

- **Primary checkout is detached.** Observed 2026-10-19: `git worktree list` first row `/Users/goos/MoAI/moai-adk-go  478285ab4c (detached HEAD)`; `git rev-parse --short main` → `03faddb705`; `/Users/goos/MoAI/moai-adk-go/.git/HEAD` holds the bare SHA `478285ab4c1fc893b020bb464d14cb69c8422523`. `integrationLockRoot()` returns `CLAUDE_PROJECT_DIR`, else the parent of `git rev-parse --git-common-dir` (`internal/cli/integration.go:34-72`), i.e. the primary; `candidateCIEnabled(root)` loads `filepath.Join(root, ".moai")` (`integration_merge.go:210-216`). A commit landing on `main` therefore never changes that file; the leader ruled (board 470c) that the key is read from the committed blob of the integration branch.
- **Integration-branch name.** `configuredIntegrationBranch()` = `config.LoadGitFlowIntegrationConfig(resolveProjectDir()).IntegrationTarget` (`integration.go:291-293`); the merge verb prefers the held window record's branch (`integration_merge.go`, "the merge target is the branch the WINDOW RECORD names").
- **Receipt citation forms.** Newest `sync-audit*.md` of 8 sampled merge-ready SPEC cards: all 8 mention at least one `rcpt-` id somewhere; a top-level machine line `audit_receipt: rcpt-…` exists in 2 (t1834, t1770; t1770 also carries `AUDIT-VERDICT: PASS … receipts=…`); the others cite ids in prose, tables or bullets (for example t1821 `Receipts: rcpt-…`, t1693 "receipt `rcpt-…`", t1832 a `- \`audit_receipt\`:` bullet). No Go parser reads a receipt id from a verdict file; `auditverdict.Parse` has no `audit_receipt` key.
- **Overlap baseline caveat (plan-audit O7, run-phase debt).** `evidence/overlap_baseline.py` has two known defects: it pairs adjacent landings in log order, and `run(..., check=False)` returns stdout even when `go list` fails, so a failed `go list` yields an empty closure counted as decided. The figures of §6 are order-of-magnitude evidence for the refused reuse relaxation only; the script is not fixed here because a rerun needs `go list` (resource rule 8c1e) and the relaxation is not built. Fix and re-measure before quoting the numbers anywhere else.

## 12. Piece A2 is independent of piece A1 — proof by code (added v0.7.0, board d20e)

**Anchor.** Every `file:line` below was read on the card branch at HEAD `49df956ac8` on 2026-10-19. Its code tree is main at `ec164d45d2` (`git merge-base HEAD main` → `ec164d45d2`; `git diff --name-only 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 HEAD -- internal .claude cmd` printed nothing, exit 0, so no source file differs from the pin). `main` has since moved (`git log -1 --format='%h %cs' main` → `0ddf5ca48b 2026-10-19`): the coordinates are an address at which a measurement was taken, not a claim about current main, and the run phase re-measures them. No Go build or test was run (resource rule 8c1e); every finding is a read of source plus the greps quoted in §12.4.

### 12.1 Surface by surface

| Surface | Observed (file:line) | What A2 does there | A1 type or behaviour needed |
|---|---|---|---|
| Receipt record | `internal/auditreceipt/store.go:166-177` (`Receipt`: eight JSON fields, `omitempty` on the optional ones) | adds `audited_commit` and `audit_target` | none |
| Receipt write | `store.go:944-959` (`WriteReceipt` assigns only `ReceiptID` and `CreatedAt`); `mcp_audit_receipt.go:60-100` (`recordAuditReceiptAt` builds the struct, `TreeRoot: root` at :71) | the new fields ride on the struct the writer builds; nothing in `WriteReceipt` overwrites them | none |
| Receipt read and legacy files | `store.go:964-971` (`ReadReceipt`), `store.go:1384-1393` (`readJSON` = `os.ReadFile` + `json.Unmarshal`); `DisallowUnknownFields` appears only in `internal/cli/mcp_claude_protocol.go:119` and `internal/cli/update/backup/retention_prune.go:248`, neither in `internal/auditreceipt` | a legacy file decodes with the new fields empty; a new file read by an older binary ignores them | none |
| Existing guard | `store.go:1258-` (`CheckCitedReceiptsSince`); its tree test is `r.TreeRoot != start.TreeRoot` at :1290 | unchanged: it reads only fields it already knows | none |
| Store root and tree identity | `storeroot.go:242-251` (`StoreRoot`: the primary checkout for a config-orphaned root); `treeroot.go:37-45` (`Canonical`: `EvalSymlinks`, else `Clean`) | the binding looks the receipt up at `StoreRoot(audited tree)` and compares `tree_root` with `Canonical(audited tree path)` | none |
| Entry points | `mcp_codex.go:2056` (`handleCodexAudit`), `recordAuditReceipt(` at :2112 and :2163 (argument `rootArg`); `mcp_convergence.go:928` (`recordAuditReceiptAt(…, cfg.ProjectRoot, …)`) | capture HEAD and target at entry, pass them to the writer | none |
| `factoryCompleteCard` | `factory_card.go:2063`; github-flow early return :2072; step-1 gates (stage, lease) :2096-2109; `initWindowLeaseOverride` :2125; branch/source resolved :2131-2139; provisioning refusals :2150-2182 (last at :2178); window acquisition `acquiredHere := false` :2185; step 2 adoption :2222; step 3 `ReadRemeasureRecord(lockRoot, candidateTree)` :2281; step 4 merge :2286 | one call between :2178 and :2185; steps 3 and 4 untouched | none: the call needs the card, the resolved branch and HEAD, all in scope at :2185 |
| Sync-audit condition | `internal/factorylane/merge.go:245` (`checkSyncAudit` reads only `sync_status`); `merge_specless.go:43` (`checkSyncAuditVerdictFile`); triple at `merge.go:174`, called by `cli/factory_merge.go:101` and `cli/factory_card_pr.go:229` | untouched under Q28's default; the binding is a separate function | none |
| Merge step | `internal/factory/integration_merge_step.go:321-325` (`ReadRemeasureRecord`, `ValidateRemeasureRecord`) | untouched | none |
| Merge verb and branch resolution | `cli/integration_merge.go:55-68` (window record branch, else `configuredIntegrationBranch()`, else `"develop"`, then `integrationMergeWorktree`); `cli/integration.go:291` (`configuredIntegrationBranch`); `config/loader_integration_branch.go:96` (zero config when git-strategy is absent) | the gate-state reader reuses the first two sources and drops the `"develop"` fallback (design §4) | none |
| `factoryResolveIntegrationBranch` | `factory_card.go:2552-2557` (configured target, else the caller's own branch with source `caller`, which `complete` refuses at :2150-2160) | `complete` gates on the branch it is about to merge into | none |
| Configuration | `config/types.go:552` (`CandidateCI`), :570 (`SlotLease`), :577 (`LocalMainIntegration`); `cli/integration_merge.go:210` (`candidateCIEnabled`, an absent key reads false) | add a `LandingGate` field beside them | none |
| Verdict parse and admission | `auditverdict.Parse`, `AdmitWithRequired` (`internal/auditverdict/verdict.go:271`); `runtime.ResolveRequiredBackends` (`runtime/audit_gates.go:43`); the SPEC-less path imports `auditverdict` already (`merge_specless.go`) | reused as is; the newest-verdict ranking (`homestate/card_evidence_readers.go:244`) moves to `auditverdict` | none |

### 12.2 The three questions

- **What A2's `complete` does when A1's manifest does not exist.** It never asks. No code in A2 reads a manifest, an attempt record or a verifier verdict, and none exists to read (§12.4). With the gate on, `complete` runs exactly today's checks — card stage and lease (:2096-2109), window, adoption, the re-measure record check (:2281), the merge step (which validates the same record at `integration_merge_step.go:321-325`) — plus one earlier step: the audit binding. With the gate off it runs only today's checks (REQ-LVM-007). When A1 lands later it replaces the decision at :2281 and :321-325 and adds its own checks under the same key; nothing A2 wrote has to change.
- **Can the gate be on with nothing behind it.** No: on, for A2 alone, is a real check — a verdict naming the current HEAD, a single machine citation line, and a receipt bound to that HEAD — and it refuses a card that lacks them (AC-LVM-008, 010, 018). It is not a re-measure completeness check, and the SPEC says so (spec C9, design §5). The one way the gate can be on with the required thing missing is a deployment one: the receipt-writing server running a build without REQ-LVM-009, in which case every receipt is unbound and every `complete` refuses; that is fail-closed, visible, ahead of the window, and is the prerequisite C5 and the cost C6. The key is not switched on between the two landings: decision Q27 (board 6404) turns it on only after A1 has also landed, because with A2 alone on the direct merge path stays a side door (design R6).
- **Whether an A2 requirement depends on an A1 concept.** Four places named one and were adjusted in v0.7.0; after the adjustments none does. (1) REQ-LVM-007 carried the narrowing-allowance list (`internal/cli` only) — A1's concept; the clause was dropped from A2. (2) REQ-LVM-012 required the key-off card's derived manifest to hold the reader calls of the configuration file — A1's concept; replaced by "lands through the gated path with the audit binding intact", and the manifest assertion is left to A1. The old REQ-LVM-011 (lane guidance) is half A1 (the three manifest commands, restart as a new attempt) and half A2 (gate-state source, `gate_state_unreadable`, no backfill after the audit, by-path binary and receipt-server prerequisite, the `audit_receipt machine line`); the A2 half is attached to REQ-LVM-012 and REQ-LVM-018. (3) REQ-LVM-017 named the merge step, merge-readiness condition (d) and `manifest verify` as consumers and `manifest verify` as the observation surface; A2 names `complete` and `moai integration status` instead, and the reader is one function A1's consumers will call. (4) The audit leg of the old REQ-LVM-015 (an earlier PASS never substitutes) is already REQ-LVM-008; the rest of REQ-LVM-015 stays with A1.

### 12.3 What A1 inherits from A2 (so the order is A2 then A1)

The gate-state reader and the key; the audit binding; both edit `factoryCompleteCard` (:2063), which is why A1 is blocked until this card is merged-local. A2 inherits nothing from A1.

### 12.4 Observed absence (greps over `internal` and `cmd`, `*.go`, at HEAD `49df956ac8`)

- `grep -rnE -e 'audit_receipt:'` → exit 0, 1 line, in a test: `internal/template/audit_plan_doc_surface_test.go:81` (the sync workflow document-surface test lists the string; no production Go code parses it). The same string is the orchestrator's binding-statement line at `.claude/skills/moai/workflows/sync.md:91` (`audit_receipt: <ids|none>`), which REQ-LVM-018 makes strict at landing; `internal/template/audit_plan_doc_surface_test.go` requires `sync.md` to keep it.
- `-e 'landing_gate'` → exit 0, 1 line, in a test file name comment (`internal/cli/candidate_landing_gate_test.go:3`, an unrelated candidate-CI test); no `LandingGate` identifier (`-e LandingGate` → exit 1, 0 lines).
- `-e 'audited_commit'` → exit 0, 1 line, a test case name in `internal/factorylane/merge_specless_test.go:130`; no production field.
- `-e 'AttemptRecord|CompletenessVerifier|RemeasureManifest|manifest derive'` → exit 0, 1 line, an unrelated skills-manifest message in `internal/cli/update_archive_guard_test.go:87`; no A1 type exists.
- `-e 'gate_state_unreadable'` → exit 1, 0 lines.
- `grep -rliE 'manifest' --include='*.go'` over `internal/factory`, `internal/factorylane`, `internal/auditreceipt`, `internal/auditverdict` and the files `cli/factory_card.go`, `cli/integration.go`, `cli/integration_merge.go`, `cli/integration_remeasure.go`, `cli/mcp_audit_receipt.go`, `cli/mcp_codex.go`, `cli/mcp_convergence.go` → exit 0, 9 files, all in `internal/factory` and all about the todo-runtime provenance record (`ManifestJSON` in `todo_runtime.go:34`, `factory_runtime.go:30-35`, and their tests); none is a re-measure manifest.

### 12.5 Gaps

No Go compile or test run (resource rule 8c1e), so "implementable without any A1 type" is established by reading call sites and imports, not by building A2 alone. The newest-verdict move to `internal/auditverdict` and the `internal/factorylane` binding were checked against the import lists (`factorylane` imports `atomicfile`, `auditverdict`, `config`, `gitenv`, `sessionmsg`; `homestate` imports `factory`, `runtime`, `auditreceipt`, among others, and is not imported by `factorylane`), not against a compiler. The run phase confirms with a build of each milestone.

## 13. Tag shadows branch — probe (added v0.9.0, plan-audit A2 round 1)

Scratch repository under the session scratchpad (no card tree touched): branch `refs/heads/main` at commit `34238e27` holding `enabled: true`, tag `refs/tags/main` at the next commit `0509e0ec` holding `enabled: false`. Observed: `git show main:c/w.yaml` printed `warning: refname 'main' is ambiguous.` then `enabled: false`, exit 0, and `git cat-file -e main:c/w.yaml` also exited 0 through the tag. With `S=$(git rev-parse "refs/heads/main^{commit}")`, `git cat-file -e "${S}:c/w.yaml"` exited 0 and `git show "${S}:c/w.yaml"` printed `enabled: true`. A tag-only name (`only_tag`): `git rev-parse --verify -q "refs/heads/only_tag^{commit}"` exited 1, so it reads as unresolvable. The reader therefore resolves the branch once through `refs/heads/` and uses the SHA (REQ-LVM-017).

## 14. Generated and hand-edited files for the sync-auditor protocol (added v0.9.0)

`AGENTS.local.md` §2.0: C1 `.claude/agents/moai/*.md` is the local hand-edited copy; C2 `internal/template/templates/.claude/agents/moai/*.md` is the neutral source layer, hand-edited; C3 `internal/template/templates/.codex/agents/moai/*.toml` is machine-emitted from C2 and never hand-edited. `Makefile:38-49`: `agents-emit` regenerates, `agents-emit-check` (ahead of `build`) only checks. `internal/template/agentemit/golden_test.go:63-116`: the emitter writes only the template `.toml` tree (`committedTOMLPath` joins `templatesDir`). Observed: no directory `internal/template/agentemit/source` exists; the mapping manifest is `agents-codex.yaml` (sync-auditor appears there as a read-only role). The root `.codex/agents/moai/sync-auditor.toml` is byte-identical to the template one (`cmp` exit 0) and is not written by the emitter, so it is refreshed by copying. `.moai/docs/audit-artifact-convention.md` is local-only (AGENTS.local.md, "Local-Only Files").

## 15. Absent path versus failed read; the merge-step pin; the catalog hash (added v0.10.0, plan-audit A2 round 2)

Scratch repository of §13, `S` = the branch commit. Observed: `git ls-tree "$S" -- c/w.yaml` printed the entry, exit 0; `git ls-tree "$S" -- c/none.yaml` printed nothing, exit 0; `git ls-tree 0000000000000000000000000000000000000001 -- c/w.yaml` printed `fatal: not a tree object`, exit 128; `git cat-file -e "${S}:c/none.yaml"` exit 128 (the same status as a failure), `git rev-parse --verify -q "${S}:c/none.yaml"` exit 1. The reader keys on `ls-tree` (design §4).

Merge-step pin (read, not run): `internal/factory/integration_merge_step.go:283-293` resolves the card branch and runs `git rev-parse refs/heads/<cardBranch>` ("pin ONE SHA", :289); `internal/cli/factory_card.go:2231` reads the card tip for the adoption comparison and :2286 calls `RunMergeStep` with no expected-SHA field in `MergeStepInput` (:2300-2309 sets Root, IntegrationWorktree, IntegrationBranch, CardID, CallerSessionID, DeferRelease). Nothing carries the bound SHA to the pin.

Catalog hash: `internal/template/catalog.yaml:188-192` (entry `sync-auditor`, path `templates/.claude/agents/moai/sync-auditor.md`, a sha256 `hash`); `Makefile:35` runs `go run ./internal/template/scripts/gen-catalog-hashes.go --all` inside `build`; the script accepts `--all`, `--entry NAME` and `--dry-run` (`internal/template/scripts/gen-catalog-hashes.go:199-203`); `internal/template/claude_audit_surface_test.go:11` (`TestClaudeAuditTemplateSurfacesAndCatalogHash`) lists the template sync-auditor and checks the catalog hash. `grep -c "factory-dispatch-mechanics\|audit-artifact" internal/template/catalog.yaml` → 0, so the rule file and the convention carry no catalog hash. Not run: the regeneration and the test (rule 8c1e, no Go call needed at plan time).


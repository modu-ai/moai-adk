---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — progress"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-14
author: manager-spec (card t708)
---

# Progress — SPEC-GATEWAY-ENVELOPE-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-GATEWAY-ENVELOPE-REPAIR-001
card: t708
phase: plan
status: completed  # §E.1 mirrors spec.md frontmatter (re-closed after sync-audit iter2 PASS 92/100)
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md, research.md]
base: local develop 7a7a08f20
branch: WT-envelope-persist
symbol_pins_verified_at_plan_time: true
security_determination: spec.md §3 (byte-exact re-injection satisfies the t672 binding)
sibling_seam: spec.md §4 (t700 REQ-WRR-007 vs REQ-WRR-008(d); Reading B adopted, adjudication gated at M0)
needs_clarification_count: 0  # plan.md §H Resolution Record — all three markers dispositioned (lead conditional-Kickoff step 1)
```

- plan_complete_at: 2026-09-13T13:44:15Z
- plan_status: audit-ready
- plan_audit: PASS 0.975 (iteration 2/2) — .moai/reports/t708/plan-audit.md

### M0 Adjudication Record (2026-09-14)

Two-input adjudication per plan.md §F M0 / §H Resolution Record row 3. Verdict-only card; zero code changes in this step.

**Input 1 — t700 seam (Reading A/B)**: t700 SPEC-GATEWAY-WEDGE-REROOT-001 remains unlanded (draft in `.claude/worktrees/t700`). Its drafted text bans envelope re-issue/synthesis/transplant with a rationale that bans MANUFACTURED envelopes ("the Opaque digest must always reference something the gateway itself issued"); verbatim gateway-issued re-injection satisfies that rationale. t708 spec.md §4 Reading B position stands with no adverse finding. Final confirmation re-rides the M1 develop absorb, where §A step 2 re-reads t700's landed status (merge order is the lead's call at landing time).

**Input 2 — t707 verdict** (`.claude/worktrees/t707/.moai/reports/t707/verdict.md`, merged to local develop as `4da5d1c4e`): no reproducible serialization defect (8 synthetic publish→stream→client-model→Check round-trips all GREEN; tamper negative tests pass — Check strength unchanged). Live CauseChain 400s localized only to a common pattern (first request ~80-90ms after an Edit tool run) with transcript≠request bytes PROVEN (prefix recompute mismatch from boundary[0] on a request pair that previously passed). Final localization awaits one live instrumentation run (t707 observability.patch, MOAI_RECEIPT_DEBUG=1) — operator decision pending, owned by t707's follow-up, NOT assigned to this SPEC. Per §H row 3: scope stays CauseReasoning (REQ-EVR-004); M3+ proceeds unblocked on this axis.

**Design directive folded (lead)**: this SPEC's design rests on client-side persistence / replay byte preservation without presupposing any server-side fix. t707's transcript≠request-bytes finding CORROBORATES the design premise: the family transcript retains the correct issued bytes (t708 research addendum: 36 complete carriers measured) while the outgoing request diverges — the launcher-side verbatim re-injection path is exactly the byte-preservation mechanism, and it presupposes no server-side change (REQ-EVR-001 validator lock).

**M0 outcome: PASSED — no adverse scope finding from either input.**

## §E.2 Run-phase Evidence

### M1 Develop Absorb + Pin Re-verification (2026-09-14, lead window)

- Absorb: `git merge develop --no-edit` — develop `4da5d1c4e` (carries t707 merge) → new HEAD `febadc784`, clean merge, no conflicts. Ancestry gate: `git merge-base --is-ancestor 4da5d1c4e HEAD` → yes.
- Pin re-verification (plan.md §C, absorbed tree, this run):
  - `go doc ./internal/gateway/receipt Manifest.Check` → resolves (func (m *Manifest) Check(authorizedUUID string, history []Observation) error)
  - `go doc ./internal/gateway/conversation Manager.Fork` → resolves (func (m *Manager) Fork(ctx context.Context, parentID string) (d Descriptor, err error))
  - `grep -n 'func (m \*Manager) refreshNative' internal/gateway/conversation/native.go` → native.go:17 (D2-corrected form) + refreshNativeProject :73
  - receipt_history.go 5/5 symbols (Check/checkObserved/Publish/replayCause/observations) → 5
  - projection.go thinking-skip → 1; codec.go Decode/BindToolID/RestoreToolID → 3; gateway_factory.go 2 → 2; gateway_session.go flow → 3
  - Result: 0 pin breaks; no blocker.
- Characterization baseline: `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1` → `ok ... 1.065s`
- Lint: branch delta `golangci-lint run --new-from-rev=7a7a08f20 internal/gateway/... internal/cli/...` → `0 issues.`, exit 0. Absorbed-tree TOTAL lint measured **0 issues** by the lead on the tree carrying `4da5d1c4e` — the earlier "997 pre-existing findings (t671 scope)" figure was a pre-absorb-tree measurement and no longer describes this tree.
- Cross-build: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Note: t707's verdict signature detail observed — `Manager.Fork` returns named `(d Descriptor, err error)`; symbol identity unchanged.

### M2 Validator Characterization Lock (2026-09-14)

- New test file: `internal/gateway/translate/receipt_history_wedge_policy_test.go` (+195, test-only; reuses t672 package-local fixture helpers). Commit `f50120ddd` (parent `6085fb131`).
- Matrix: 6 groups — trailing-unpublished chain rejection (golden), tail-re-rooted acceptance, 4 rejection-class cells (mid-drop/foreign-item/stripped-reasoning/lineage-miss, golden bodies), exact-retry acceptance, truncated-fork acceptance, `HistoryReplayError.Error()` full-literal byte-identical goldens (3 causes + zero-value default).
- GREEN-at-arrival (verbatim in agent report): `go test ./internal/gateway/translate/ -run 'TestWedgePolicy' -count=1 -v` → 6/6 PASS; affected packages `translate` + `receipt` → ok/ok. `go vet ./internal/gateway/...` → exit 0.
- PRESERVE zero-diff: only the new test file changed; all plan.md §D production files untouched (AC-EVR-011 inputs intact).
- Deviation notes: (1) foreign-item cell initially classified CauseLineage on an empty store — fixture corrected to publish the real conversation first (validator behaved as documented; not a SPEC defect). (2) Public-content-changed shape covered by t707's `TestReplayRoundTripStillRejectsTampering` (no duplicate golden cell added — lead concurred, addition not needed). (3) E8 RED n/a by design (GREEN-at-arrival).
- Gaps: no local full-suite run (CI owns it); coverage measured at a later milestone over the card's whole diff.

### M3 Launcher-Side Envelope Repair Path (2026-09-14)

- Commit `7ebf85df6` (parent `37987cc3c`), 5 files +966: `internal/gateway/conversation/repair.go` + `repair_test.go` (new), `internal/cli/gateway_repair.go` + `gateway_repair_test.go` (new), `internal/cli/gateway_session.go` (only edit to an existing file — prepareGatewayConversation hook + passthrough case).
- RED (verbatim, tree `e662531fa`): `go test ./internal/gateway/conversation/ -run 'TestRepair'` → build failed, `m.RepairEnvelope undefined`, `undefined: ErrRepairNotRepairable` / `ErrRepairAlreadyAttempted`.
- GREEN (verbatim counts, tree `7ebf85df6`): conversation repair 7/7 PASS; cli wiring 4/4 PASS; scoped gateway+repair surface 38 PASS (`-run 'Gateway|Repair'`); M2 translate net ok. `go build`+`GOOS=windows` exit 0; vet exit 0; scoped lint `--new-from-rev=4da5d1c4e` 0 issues.
- PRESERVE zero-diff vs merge-base `4da5d1c4e`: the 11 §D files → empty diff (directly verified).
- New surface: `Manager.RepairEnvelope(ctx, id) ([]RepairProvenance, error)`; `ErrRepairNotRepairable`/`ErrRepairAlreadyAttempted`; durable record `families/<FamilyID>/repair/<UUID>.json` {attempted, boundaries[{digest,position}], aside, aside_sha256}; aside `<transcript>.moai-repair-aside` (O_EXCL preimage, never deleted); launcher flag `--repair-envelope` (requires `--resume`, runs before Resume, stripped from child args).
- Key HOW decisions: marker self-attestation only (no receipt-store read — source-scan test AC-EVR-010); repair writes the TRANSCRIPT (the launcher-owned file the client re-encodes from) and guarantees the persisted source is correct/refusable — never the wire bytes (client encode-time behavior is the documented boundary, t703/t707-consistent); refusals write nothing (do not burn the single-shot); durable record written before transcript rewrite (crash burns the shot conservatively); refusal guidance = `translate.HistoryReplayError{Cause: CauseReasoning}.Error()` at CLI layer (no string duplication); `isSidechain` rows excluded.
- AC rows: AC-EVR-004..010 all PASS (per-AC evidence in agent report).
- Gaps: full internal/cli package run = default-10m timeout at 601s — pre-existing baseline (~1583s package total on this machine); CI owns the full verdict. Coverage (E3) deferred to the card's whole diff at a later milestone.
- Note: mid-milestone the agent hit a transient 429 rate limit and was resumed; no work lost.

### M1b Re-absorb — t653 landed (2026-09-14, lead window)

- Absorb: develop `93ae49ce7` (carries t653: receipt/{core,store,compact}.go + conversation/family.go) → merge HEAD `1aca63da8`, clean, no conflicts. `git merge-base --is-ancestor 93ae49ce7 HEAD` → yes.
- Pin re-verification on the absorbed tree: Manifest.Check ✓ (go doc), Manager.Fork ✓ (go doc), receipt_history 5/5, projection thinking-skip 1, codec 3, refreshNative method 2 hits. **0 pin breaks.**
- Test re-runs: conversation `ok 31.839s` (M3 repair 7/7 within), translate `ok 36.059s` (M2 wedge+receipt-history within), cli `-run 'Gateway|Repair'` `ok 13.992s` (38-surface incl. M4 live lock, 0 violations). `GOOS=windows` build exit 0, vet exit 0.
- Conclusion: t653 changed the same gateway surface but broke no pin and no test; M4's merge-base recomputation picked up `93ae49ce7` and still measures 0 violations.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-14
run_commit_sha: 3d4415676   # code close; the close-out commit (status flip + this §E.3) lands after
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 11   # §D files, all zero-diff (M4 lock enforces mechanically)
l44_pre_commit_fetch: not-run (lane-local card worktree; lane does not push — lead batch-pushes origin/develop)
l44_post_push_fetch: not-run (same lane push policy; remote landing is the lead's verification)
new_warnings_or_lints_introduced: 0   # golangci-lint --new-from-rev=4da5d1c4e internal/cli/ internal/gateway/conversation/ → 0 issues
cross_platform_build:
  native: exit 0   # go build ./...
  windows_amd64: exit 0   # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 9   # 2 test (M2) + 5 (M3: repair.go, repair_test.go, gateway_repair.go, gateway_repair_test.go, gateway_session.go) + 2 (M4/M6: gateway_preserve_test.go, gateway_repair_doc_test.go) + 1 doc (.moai/docs/gateway-envelope-repair.md)
m1_to_mN_commit_strategy: per-milestone commits (M0/M1 records → M2 test → M3 feat → M4 test → M6 docs → close-out chore); lane does not push
```

### E1 AC matrix (all evidence commands + observed output this run, tree `3d4415676` unless noted)

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-EVR-001 | PASS | `go test ./internal/gateway/translate/ ./internal/gateway/receipt/ -count=1` | `ok ... translate 33.825s` / `ok ... receipt 12.989s`; M2 wedge matrix 6/6 GREEN (`TestWedgePolicy*` PASS) |
| AC-EVR-002 | PASS | `go test -run TestWedgePolicy ./internal/gateway/translate/` | golden byte-identity cells PASS incl. zero-value chain default |
| AC-EVR-003 | PASS | same as 001 | mid-drop / foreign-item / stripped-reasoning / lineage-miss cells PASS with unchanged classes |
| AC-EVR-004 | PASS | `go test -run TestRepairEnvelope ./internal/gateway/conversation/` | restores cell PASS; `sha256(injected raw) == marker.opaque_sha256` asserted |
| AC-EVR-005 | PASS | same | head-position injection PASS; marker-less boundary → refusal zero-mod |
| AC-EVR-006 | PASS | `go test -run 'TestRepairEnvelopeFlag' ./internal/cli/` | no-flag no-op + repair-then-resume PASS (4/4 wiring) |
| AC-EVR-007 | PASS | `go test -run TestRepairEnvelope ./internal/gateway/conversation/` | source-gone / digest-mismatch / intact / incomplete refusal cells PASS, zero-mod asserted |
| AC-EVR-008 | PASS | same | `TestRepairEnvelopeSingleShotTerminatesAcrossRestart` PASS (fresh Manager = fresh process; durable single-shot) |
| AC-EVR-009 | PASS | same | aside byte-identical preimage + `{digest,position}` provenance PASS |
| AC-EVR-010 | PASS | `TestRepairPathNeverReadsReceiptStore` | source scan 0 hits for receipt symbols in repair.go |
| AC-EVR-011 | PASS | `TestGatewayRepairCardDiffTouchesNoPreservedFile` + discriminator | live: 0 violations vs merge-base `4da5d1c4e`; negative cell flags receipt/store.go, receipt_history.go, family.go |
| AC-EVR-012 | PASS | `TestGatewayEnvelopeRepairOperatorDocExistsWithSections` | `.moai/docs/gateway-envelope-repair.md` exists, section checklist green |

### E2 cross-platform build
`go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (observed on trees `7ebf85df6` and `58587fada`).

### E3 coverage (per-package figures, verbatim)
- `go test -cover ./internal/gateway/conversation/ -count=1` → `coverage: 79.5% of statements` (whole package incl. pre-existing family/native code). New `repair.go` scoped: RepairEnvelope 75.5%, repairPlan 87.0%, helpers 44.4–100%.
- `go test -cover ./internal/cli/` full-package: NOT measurable in-bounds — twice observed failing at the 10m timeout (591s/601s) against the known ~1583s package baseline; CI owns the full verdict. Scoped substitute `-run 'Gateway|Repair' -coverprofile`: new-file figures `gateway_repair.go` guidance 100.0% / repairGatewayEnvelope 87.0%, `gatewayConversationPassthrough` 100.0%; scoped-run aggregate 13.5% is meaningless (dominated by in-scope-untouched pre-existing code) — reported only for completeness.
- Caveat: neither aggregate is the strict-profile ≥85% gate reading; the new-code per-file figures above are the card-attributable measurement.

### E4 subagent boundary grep (card-scoped)
`grep -rn 'AskUserQuestion\|mcp__askuser' <card's 3 production files>` → exit 1, zero hits. (Repo-wide form hits pre-existing doc-comments in files this card never touched — harness.go, pr_watch_cmd.go, agentlint — baseline, not new.)

### E5 lint
`golangci-lint run --new-from-rev=4da5d1c4e internal/cli/ internal/gateway/conversation/` → `0 issues.`

### E6 commit inventory (branch WT-envelope-persist, unpushed; lane does not push)
- `f50120ddd` M2 wedge-policy characterization matrix (test-only)
- `7ebf85df6` M3 launcher-side envelope repair verb (feat; 5 files, +966)
- `5b10efb66` M4 non-invasiveness lock (test-only)
- `3d4415676` M6 operator documentation + AC check (docs)
- close-out commit: status flip + this §E.3 (this commit)

### E8 RED evidence pointers
- M2: GREEN-at-arrival by design (characterization lock; no RED exists — plan §F M2).
- M3: verbatim compile-failure run of `repair_test.go` against the undefined API recorded in the M3 report and §E.2 (tree `e662531fa`).
- M6: verbatim missing-doc failure of `TestGatewayEnvelopeRepairOperatorDocExistsWithSections` (tree `3d4415676^` work).

### Gap — M5 live probe (skipped by coordinator decision)
The optional live-probe milestone was skipped this card; its live-instrumentation axis is owned by card t707's follow-up. Unmeasured consequences: (1) transcript envelope-retention under real client compaction/`--continue` slicing (plan §H row-2 disposition covers it — the repair refuses on source-gone); (2) no live end-to-end probe of a repaired transcript against a real client replay — in-repo tests prove the transcript-side machinery and unchanged-Check composition, not external encode-time behavior. Not silent: recorded here, in the M6 report, and in the operator doc's non-recoverable-shapes section.

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-09-14
sync_commit_sha: "430cf4429"  # D3 backfill — real SHA of the sync commit, backfilled in this follow-up commit
sync_status: completed  # re-close
re_close: sync-audit iter1 FAIL 78 (F1/F2/F7) → repair commits b46d33271/c9e98df4c → iter2 PASS 92 (.moai/reports/t708/sync-audit.md) → this commit re-closes; sync_commit_sha 430cf4429 remains the original close, this commit is the re-close record
sync_summary: CHANGELOG [Unreleased] entry emitted (B12: pre-emission grep 0, AC count 12/12, paths verified); SPEC 3-phase close riding the single sync commit (spec.md frontmatter in-progress → completed, body untouched); docs-site skipped — maintainer-facing launcher flag documented repo-locally at .moai/docs/gateway-envelope-repair.md (M6 artifact).


---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — progress"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t708)
---

# Progress — SPEC-GATEWAY-ENVELOPE-REPAIR-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec: SPEC-GATEWAY-ENVELOPE-REPAIR-001
card: t708
phase: plan
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md, research.md]
base: local develop 7a7a08f20
branch: WT-envelope-persist
symbol_pins_verified_at_plan_time: true
security_determination: spec.md §3 (byte-exact re-injection satisfies the t672 binding)
sibling_seam: spec.md §4 (t700 REQ-WRR-007 vs REQ-WRR-008(d); Reading B adopted, adjudication gated at M0)
needs_clarification_count: 2  # plan.md §H — both bounded by the refusal path (REQ-EVR-007)
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- plan_complete_at: 2026-09-13T13:44:15Z
- plan_status: audit-ready
- plan_audit: PASS 0.975 (iteration 2/2) — .moai/reports/t708/plan-audit.md

## §E.1 Addendum — M0 Adjudication Record (2026-09-14)

Two-input adjudication per plan.md §F M0 / §H Resolution Record row 3. Verdict-only card; zero code changes in this step.

**Input 1 — t700 seam (Reading A/B)**: t700 SPEC-GATEWAY-WEDGE-REROOT-001 remains unlanded (draft in `.claude/worktrees/t700`). Its drafted text bans envelope re-issue/synthesis/transplant with a rationale that bans MANUFACTURED envelopes ("the Opaque digest must always reference something the gateway itself issued"); verbatim gateway-issued re-injection satisfies that rationale. t708 spec.md §4 Reading B position stands with no adverse finding. Final confirmation re-rides the M1 develop absorb, where §A step 2 re-reads t700's landed status (merge order is the lead's call at landing time).

**Input 2 — t707 verdict** (`.claude/worktrees/t707/.moai/reports/t707/verdict.md`, soon merged to develop): no reproducible serialization defect (8 synthetic publish→stream→client-model→Check round-trips all GREEN; tamper negative tests pass — Check strength unchanged). Live CauseChain 400s localized only to a common pattern (first request ~80-90ms after an Edit tool run) with transcript≠request bytes PROVEN (prefix recompute mismatch from boundary[0] on a request pair that previously passed). Final localization awaits one live instrumentation run (t707 observability.patch, MOAI_RECEIPT_DEBUG=1) — operator decision pending, owned by t707's follow-up, NOT assigned to this SPEC. Per §H row 3: scope stays CauseReasoning (REQ-EVR-004); M3+ proceeds unblocked on this axis.

**Design directive folded (lead)**: this SPEC's design rests on client-side persistence / replay byte preservation without presupposing any server-side fix. t707's transcript≠request-bytes finding CORROBORATES the design premise: the family transcript retains the correct issued bytes (t708 research addendum: 36 complete carriers measured) while the outgoing request diverges — the launcher-side verbatim re-injection path is exactly the byte-preservation mechanism, and it presupposes no server-side change (REQ-EVR-001 validator lock).

**M0 outcome: PASSED — no adverse scope finding from either input. M1 (develop absorb + pin re-verification) is ready; awaiting the lead's window signal so the absorb carries the t707 merge.**

## §E.2 — M1 Develop Absorb + Pin Re-verification (2026-09-14, lead window)

- Absorb: `git merge develop --no-edit` — develop `4da5d1c4e` (carries t707 merge) → new HEAD `febadc784`, clean merge, no conflicts. Ancestry gate: `git merge-base --is-ancestor 4da5d1c4e HEAD` → yes.
- Pin re-verification (plan.md §C, absorbed tree, this run):
  - `go doc ./internal/gateway/receipt Manifest.Check` → resolves (func (m *Manifest) Check(authorizedUUID string, history []Observation) error)
  - `go doc ./internal/gateway/conversation Manager.Fork` → resolves (func (m *Manager) Fork(ctx context.Context, parentID string) (d Descriptor, err error))
  - `grep -n 'func (m \*Manager) refreshNative' internal/gateway/conversation/native.go` → native.go:17 (D2-corrected form) + refreshNativeProject :73
  - receipt_history.go 5/5 symbols (Check/checkObserved/Publish/replayCause/observations) → 5
  - projection.go thinking-skip → 1; codec.go Decode/BindToolID/RestoreToolID → 3; gateway_factory.go 2 → 2; gateway_session.go flow → 3
  - Result: 0 pin breaks; no blocker.
- Characterization baseline: `go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1` → `ok ... 1.065s`
- Scoped lint delta: `golangci-lint run --new-from-rev=7a7a08f20 internal/gateway/... internal/cli/...` → `0 issues.`, exit 0. (Pre-existing repo-wide lint red, 997 findings on t707's measurement, is t671's scope — 0 NEW issues on this branch's delta.)
- Cross-build: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- Note: t707's verdict signature detail observed — `Manager.Fork` returns named `(d Descriptor, err error)`; symbol identity unchanged.

## §E.2 — M2 Validator Characterization Lock (2026-09-14)

- New test file: `internal/gateway/translate/receipt_history_wedge_policy_test.go` (+195, test-only; reuses t672 package-local fixture helpers). Commit `f50120ddd` (parent `6085fb131`).
- Matrix: 6 groups — trailing-unpublished chain rejection (golden), tail-re-rooted acceptance, 4 rejection-class cells (mid-drop/foreign-item/stripped-reasoning/lineage-miss, golden bodies), exact-retry acceptance, truncated-fork acceptance, `HistoryReplayError.Error()` full-literal byte-identical goldens (3 causes + zero-value default).
- GREEN-at-arrival (verbatim in agent report): `go test ./internal/gateway/translate/ -run 'TestWedgePolicy' -count=1 -v` → 6/6 PASS; affected packages `translate` + `receipt` → ok/ok. `go vet ./internal/gateway/...` → exit 0.
- PRESERVE zero-diff: only the new test file changed; all plan.md §D production files untouched (AC-EVR-011 inputs intact).
- Deviation notes: (1) foreign-item cell initially classified CauseLineage on an empty store — fixture corrected to publish the real conversation first (validator behaved as documented; not a SPEC defect). (2) Public-content-changed shape covered by t707's `TestReplayRoundTripStillRejectsTampering` (no duplicate golden cell added — flagged for orchestrator). (3) E8 RED n/a by design (GREEN-at-arrival).
- Gaps: no local full-suite run (CI owns it); coverage measured at a later milestone over the card's whole diff.

---
id: SPEC-GATEWAY-WEDGE-REROOT-001
title: "Wedge-safe re-rooting policy — implementation plan"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t700)
tier: M
---

# Plan — SPEC-GATEWAY-WEDGE-REROOT-001

## §A Context

- **Card / branch / base**: card t700, Class C (design change + security determination). Worktree `.claude/worktrees/t700`, branch `WT-wedge-reroot-policy`, base local develop `44e56d017`. Run phase proceeds per the git-flow lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`): card worktree branches from `develop`, integrates back into local `develop` via no-ff merge, **lane does not push** (lead batch-pushes `origin/develop`).
- **Merge order: t697 → t700 (HARD).** Sibling card t697 is concurrently modifying the same gateway area and lands first. Before the first code commit, the run phase MUST:
  1. Absorb **local `develop`** into this card worktree per the lane protocol — local develop is the lane's integration branch and the ONLY sanctioned absorb target (gitflow-lane-protocol §11). Measured at plan-audit iteration 1: local develop `e241da934` already contains the t697 merge `f45c2dddf`, while `origin/develop` (`0755cc7f5`) does NOT — absorbing `origin/develop` would complete M1 without t697. Re-measure both SHAs at M1 time; the values here are iteration-1 attribution, never pins.
  1a. **Ancestry gate**: `git merge-base --is-ancestor <t697-merge-SHA> HEAD` MUST exit 0 after the absorb — resolve the t697 merge SHA at M1 time from local develop's log (never pin a moving ref); a non-zero exit is a blocker report, never a silent proceed.
  2. **Re-verify every symbol pin** below against the absorbed tree — the SPEC pins symbols, never line numbers, precisely for this step. Pinned symbols: `receiptHistory.Check`, `checkObserved`, `replayCause`, `Publish`, `HistoryReplayError`, `historyReplayGuidance`, `NewReceiptHistory`, `NewGPTSubscriptionReceiptHistory` (`internal/gateway/translate/receipt_history.go`); `Manifest.Check`, `Manifest.Fork`, `Candidate`, `Observation` (`internal/gateway/receipt/core.go`); `conversation.Manager.Fork` (`internal/gateway/conversation/family.go`); `newGatewayHandlerFactory`, `authorizeGatewayNativeReceipt` (`internal/cli/gateway_factory.go`); the launcher conversation/fork flow (`internal/cli/gateway_session.go`); the characterization suite (`internal/gateway/translate/receipt_history_cause_test.go`).
  3. If any pin broke, return a blocker report to the orchestrator for D-NEW-1 re-delegation to manager-spec (artifact correction) — never patch the SPEC silently mid-run.
- **Evidence base**: `.moai/reports/t672/{verdict,investigation,matrix}.md` — treat as measured ground truth; do not re-derive or re-litigate its cells.
- **Methodology**: TDD (`cycle_type=tdd`) — the recovery verb is new behavior (RED-GREEN-REFACTOR); the validator lock is a characterization suite (GREEN at arrival, regression guard thereafter).

## §B Known Issues (domain-filtered from the 12-category template)

- **B4 Frontmatter schema**: spec.md carries the canonical 12 fields + `tier: M` + `related_specs`; sibling artifacts carry no `status:` field (artifact statelessness).
- **B5 CI 3-tier awareness**: distinguish pre-existing baseline failures from NEW ones. Known pre-existing environmental failure to ignore: `TestAppServerSubprocessHTTPToolContinuation` (named by t672 §D5, reproduced at pristine HEAD by t695).
- **B8 Working-tree hygiene**: commit by explicit pathspec; never `git add -A`; do not touch `.moai/state/`, `.moai/cache/`, `.moai/logs/`, or other cards' SPEC directories.
- **B9 Git discipline**: Conventional Commits with `🗿 MoAI` trailer; card id `t700` in every commit message body (traceability carrier — branch name carries the slug only); no `--no-verify`; lane does NOT push `develop` (lead batch-pushes).
- **B11 Subagent boundary**: return structured blocker reports; never prompt the user.

## §C Pre-flight

Run before the first code commit, after the t697 absorb (§A):

```bash
# 1. Branch + HEAD re-read (staleness rule — re-run immediately before every commit)
git branch --show-current
git rev-parse --short HEAD

# 2. Symbol-pin re-verification against the absorbed tree
#    Exported pins resolve via go doc. Unexported pins resolve via declaration
#    grep — `go doc` exits 1 on unexported symbols (measured, plan-audit iter-1).
#    The M2 characterization-suite compile is the compile net for every
#    unexported pin: the suite cannot build if a pinned unexported symbol moved
#    or changed signature in a breaking way.
go doc ./internal/gateway/receipt Manifest.Check
go doc ./internal/gateway/conversation Manager.Fork
grep -n "func NewReceiptHistory\|func NewGPTSubscriptionReceiptHistory" internal/gateway/translate/receipt_history.go
grep -n "func (h \*receiptHistory) Check(\|func (h \*receiptHistory) checkObserved(\|func (h \*receiptHistory) Publish(\|func replayCause" internal/gateway/translate/receipt_history.go
grep -n "func authorizeGatewayNativeReceipt\|func newGatewayHandlerFactory" internal/cli/gateway_factory.go
grep -n "families.Fork(" internal/cli/gateway_session.go

# 3. Characterization baseline — the validator suite must be green BEFORE any change
go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1

# 4. Cross-SPEC conflict pre-scan on the touched packages
grep -rn "Retired\|superseded" internal/gateway/translate/ internal/gateway/receipt/ || echo "no conflicts"

# 5. Scoped lint baseline
golangci-lint run --new-from-rev=HEAD internal/gateway/... internal/cli/... 2>&1 | tail -5
```

Record each command and its observed output in `progress.md §E.2` as the run-phase baseline attribution.

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (zero-diff files)**: `internal/gateway/receipt/core.go` (all of it); `internal/gateway/translate/receipt_history.go` — `Check`, `observations`, `Publish`, `checkObserved`, `replayCause`, `HistoryReplayError`, `historyReplayGuidance`, and the three reason-clause strings; `internal/gateway/translate/request.go` (entire file — the production authorization call site: the `NativeReceiptAuthorize` gate ordering and the `limits.History.Check` invocation); `internal/gateway/conversation/family.go` — `Manager.Fork`; `internal/cli/gateway_factory.go` — `newGatewayHandlerFactory`, `authorizeGatewayNativeReceipt`. The card's gateway-side diff MUST be empty for all of these, file-level (AC-WRR-012); the behavioral net over the frozen call ordering is `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory`, which drives a rejection through the full `Request` entrypoint.
- **Forbidden directions** (automatic design reject): item-level validation; relaxed prefix/previous matching; request-driven receipt-root selection/seeding/re-rooting; client-declared gateway truncation; reasoning-envelope re-issue or transplant; wire-body wording changes; automatic transcript surgery.
- **Live-probe discipline** (only if the optional integration AC-WRR-013 is executed): t672 §0 isolation verbatim — ephemeral loopback ports, per-run session token, isolated receipt store under `/tmp` (per-run UUID), child stopped via `ChildProcess.Stop` (bounded) on every path, real credentials read-only in place, one-line prompts with `max_tokens <= 64`, bounded call count, no listener left behind (post-run port check). **No background load.**
- **No local full-suite runs** (`go test ./...` prohibited — lane load discipline); run affected packages only; full-suite verdict is CI's (`origin/develop` push by the lead).
- Cross-platform: launcher-side changes build under `GOOS=windows GOARCH=amd64 go build ./...`.

## §E Self-Verification (delegation deliverables, E1-E8)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §E, with the attribution triple (command + verbatim output + tree SHA) on every item:

- **E1 AC matrix**: every AC-WRR-### PASS/FAIL with the verification command and verbatim output (acceptance.md §B is the SSOT).
- **E2 Cross-platform build**: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`, exit 0 both.
- **E3 Coverage**: `go test -cover ./internal/gateway/translate/... ./internal/cli/...` — ≥85% on touched packages.
- **E4 Subagent boundary grep**: for any touched package, the AskUserQuestion/mcp__askuser grep yields 0 non-test, non-comment matches.
- **E5 Lint**: `golangci-lint run --new-from-rev=HEAD` on touched packages — 0 NEW issues; pre-existing baseline reported separately.
- **E6 Branch HEAD + merge state**: commit SHAs, local develop merge SHA (post-integration), unpushed-commit count — reported to the lead; lane does not push.
- **E7 Blockers**: any pin break after the t697 absorb, or any AC unresolvable as specified — structured blocker report.
- **E8 RED evidence**: for the recovery verb's new tests, the verbatim pre-GREEN failing output (TDD).

## §F Milestones (ordered by decision-reversibility — highest-change-likelihood decisions first)

- **M1 — Sibling absorb + pin re-verification (Priority High, gate for everything)**
  Absorb develop (carries t697) per §A; run §C pre-flight; re-verify every symbol pin. Output: pre-flight record in `progress.md §E.2`. Blocker path if any pin broke.

- **M2 — Validator characterization lock (Priority High)**
  Extend the existing cause-test suite (`receipt_history_cause_test.go` pattern) with the wedge-policy matrix as characterization tests: wedge shape rejected-chain; tail-re-rooted shape accepted; mid-drop / foreign-item / stripped-reasoning / lineage-miss still rejected with unchanged classes; exact-retry accepted; truncated-fork accepted; error strings byte-identical (golden). These are GREEN at arrival — they are the regression guard that makes "the validator is unchanged" mechanically checkable for the rest of the card and for the future. The wire-level entrypoint net is `TestReceiptHistoryRejectsMidPairTruncationBeforeHistory` (already in the suite): it drives a rejection through the full `Request` entrypoint and guards the `NativeReceiptAuthorize` → `History.Check` ordering that AC-WRR-012's file lock freezes. References: REQ-WRR-001/002/007; AC-WRR-001..007.

- **M3 — Client-side recovery path (Priority High, the card's new behavior — TDD)**
  RED: failing tests for the launcher-side recovery path per REQ-WRR-003/004 — trailing-boundary-bounded removal (+dependent tool results), single-shot, non-destructive (removed content preserved aside), no gateway write, explicit-user-invoked, bounded termination on a still-rejected retry (REQ-WRR-005/006). GREEN: minimal implementation on the launcher surface (`internal/cli` conversation flow), operating on the conversation record's transcript pointer. Symbol-level surface decisions (verb name, flag shape) are made here; the WHAT is fixed by the SPEC, the HOW lands in this milestone. References: AC-WRR-008..011.

- **M4 — Gateway non-invasiveness lock (Priority Medium)**
  Mechanical no-weakening assertion: the card's diff (merge-base develop..HEAD) touches no file under `internal/gateway/receipt/` and no preserved symbol in `receipt_history.go` / `family.go` / `gateway_factory.go` (AC-WRR-012 — diff-scope assertion, scripted and recorded). If a later absorb forces a touch, that is a blocker, not an edit.

- **M5 — Optional live wedge-then-recover round trip (Priority Low, execute only with operator-visible cost note)**
  The integration proof per AC-WRR-013 under the t672 isolation discipline (§D). Skip is acceptable with the gap recorded; unit-level ACs do not depend on it.

- **M6 — Operator documentation (Priority Medium)**
  REQ-WRR-009: wedge class, recovery procedure, single-shot/non-destructive bounds, non-recoverable shapes, and the sanctioned fork path for lineage misses. Deliverable path: `.moai/docs/gateway-wedge-recovery.md` — the path AC-WRR-015 tests for. Language: repo docs convention (ko for user-facing docs per language.yaml `documentation: ko`).

- **M7 — Integration (Priority High)**
  Scoped verification (affected packages only), conventional commits with card id, local develop merge via the integration window (`moai integration acquire` → merge --no-ff → `release`), completion report to the lead with card id, branch + HEAD, local merge SHA, unpushed count, evidence path. Sync (SPEC close) completes **before** the develop merge so the landed state is not `in-progress` (gitflow-lane-protocol §7 lesson, t342).

## §G Anti-Patterns

- **Trim-until-accepted loops** — the single-shot bound is the security property, not a UX nicety; an iteration loop turns recovery into an oracle probing where the published chain ends.
- **Silent automatic repair on 400** — the chain class does not uniquely identify the wedge; automatic surgery would corrupt transcripts on unrelated rejections.
- **"Just relax the check for the trailing boundary"** — any validator relaxation is the rejected item-level direction wearing a smaller hat.
- **Reading the receipt store from the recovery path "just to be helpful"** — the client-side decision derives from client-owned content only; store-reading couples recovery to gateway internals for zero authorization benefit.
- **Hard-deleting the unpublished turn** — the user may have seen partial streamed content; non-destructive removal is required (REQ-WRR-003-3).
- **Pinning line numbers** — t697 moves lines; only symbol names survive the absorb.
- **Local `go test ./...`** — lane load discipline; CI owns the full suite.

## §H Cross-References

- `.moai/reports/t672/{verdict,investigation,matrix}.md` — measured ground truth (mechanism, 9-cell matrix, isolation discipline).
- `SPEC-MOAI-GATEWAY-001` — the gateway base SPEC (loopback gateway, launcher ingress).
- `.claude/rules/local/gitflow-lane-protocol.md` — lane integration procedure, serialization window, push policy.
- `.claude/rules/moai/core/verification-claim-integrity.md` — §E attribution requirements.
- `.moai/docs/hook-development.md` — n/a; not applicable (no hooks on this card).

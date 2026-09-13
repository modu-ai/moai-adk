---
id: SPEC-GATEWAY-ENVELOPE-REPAIR-001
title: "Reasoning-envelope repair — implementation plan"
version: "0.1.0"
created: 2026-09-13
updated: 2026-09-13
author: manager-spec (card t708)
tier: M
---

# Plan — SPEC-GATEWAY-ENVELOPE-REPAIR-001

## §A Context

- **Card / branch / base**: card t708, Class C (design change + security determination). Worktree `.claude/worktrees/t708`, branch `WT-envelope-persist`, base local develop `7a7a08f20`. Run phase proceeds per the git-flow lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`): card worktree branches from `develop`, integrates back into local `develop` via no-ff merge, **lane does not push** (lead batch-pushes `origin/develop`).
- **Merge order: t700 → t708 preferred; the HARD gate is pin re-verification.** Sibling card t700 (`SPEC-GATEWAY-WEDGE-REROOT-001`) is unlanded (draft in worktree `.claude/worktrees/t700`). Before the first code commit, the run phase MUST:
  1. Absorb **local `develop`** into this card worktree per the lane protocol — local develop is the lane's integration branch and the ONLY sanctioned absorb target. Measured at plan time: local develop `7a7a08f20` does NOT yet contain t700. Re-measure at M1 time; the values here are plan-time attribution, never pins.
  2. **Re-verify every symbol pin** below against the absorbed tree — the SPEC pins symbols, never line numbers, precisely for this step. If t700's SPEC has landed by then, also re-read its requirement IDs (`REQ-WRR-007`, `REQ-WRR-008(d)`) and the spec.md §4 seam adjudication status before proceeding past M2. In the same pass, re-read **card t707's verdict status** — its serial-turn chain-class determination (family `1f14d174`) is M0's second adjudication input, dispositioned in plan.md §H Resolution Record (row 3).
  3. If any pin broke, return a blocker report to the orchestrator for D-NEW-1 re-delegation to manager-spec (artifact correction) — never patch the SPEC silently mid-run.
- **Pinned symbols** (verified present on this tree at plan time, `git rev-parse --short HEAD` = `7a7a08f20`):
  - `internal/gateway/translate/receipt_history.go` — `receiptHistory.Check`, `checkObserved`, `replayCause`, `observations` (the `strings.HasPrefix(id, opaque.ToolPrefix)` stripped-envelope site), `Publish`, `CauseReasoning`, `classify`, `HistoryReplayError`, `historyReplayGuidance`.
  - `internal/gateway/receipt/core.go` — `Manifest.Check`, `Candidate`, `Observation`; `internal/gateway/receipt/projection.go` — `CanonicalPrefixes` (the `thinking`/`redacted_thinking` skip); `internal/gateway/receipt/store.go` — `Publish`, `os.Root` integrity discipline (referenced as the integrity bar for any NEW durable store, should the repair provenance record live in-repo state).
  - `internal/gateway/opaque/codec.go` — `Decode` (canonicality via `bytes.Equal`), `Digest`, `BindToolID`, `RestoreToolID`, the `moai_opaque_v1_`/`v2_` carrier format.
  - `internal/gateway/translate/response.go` / `stream.go` / `reasoning.go` / `request.go` — `outputEnvelopeWithRaw`, the `redacted_thinking` block emission (`content[0]`), the stream terminal re-emit (`content_block_start` full key-copy), `messageEnvelope`, the tool-marker restore + reasoning re-injection splices.
  - `internal/gateway/conversation/native.go` — `refreshNative` (the launcher's existing transcript read pattern), `transcriptModel`; `internal/gateway/conversation/family.go` — `Manager.Fork` (untouched), the family record fields (`receipt_dir`, `transcript`, `ConfigDir`).
  - `internal/cli/gateway_session.go` — the launcher conversation flow (`prepareGatewayConversation` and adjacent); `internal/cli/gateway_prepare.go` — the `ANTHROPIC_BASE_URL` handoff (evidence the launcher is not in the per-request path).
  - Characterization suites: `internal/gateway/translate/receipt_history_cause_test.go` (t672), `receipt_history_classify_test.go` (t703).
- **Evidence base** (treat as measured ground truth; do not re-derive): `.moai/reports/t708/research.md` (4-lens synthesis + orchestrator transcript-retention addendum), `.moai/reports/t703/verdict.md`, `.moai/reports/t672/{verdict,investigation,matrix}.md`.
- **Methodology**: TDD (`cycle_type=tdd`) — the repair verb is new behavior (RED-GREEN-REFACTOR); the validator lock is a characterization suite (GREEN at arrival, regression guard thereafter).
- **Design decisions fixed by the SPEC (WHAT)**: explicit-user invocation; single-shot with durable termination; byte-exact verbatim injection only; marker self-attestation decides admissibility (no receipt-store read — resolves research contradiction C5); launcher surface only; aside-before-replace; durable provenance; refusal on every unprovable precondition. **Open HOW decisions land in M3** (verb/flag shape, repair-record storage location, transcript-read plumbing) — the SPEC does not pin them.

## §B Known Issues (domain-filtered)

- **B4 Frontmatter schema**: spec.md carries the canonical 12 fields + `tier: M` + `related_specs`; sibling artifacts carry no `status:` field (artifact statelessness).
- **B5 CI 3-tier awareness**: distinguish pre-existing baseline failures from NEW ones. Known pre-existing environmental failure to ignore: `TestAppServerSubprocessHTTPToolContinuation` (named by t672 §D5). Pre-existing golangci-lint errcheck findings in untouched gateway test files (t703 verdict: 15 findings) are baseline, not new.
- **B8 Working-tree hygiene**: commit by explicit pathspec; never `git add -A`; do not touch `.moai/state/`, `.moai/cache/`, `.moai/logs/`, or other cards' SPEC directories (including `.moai/reports/t708/` — read-only evidence).
- **B9 Git discipline**: Conventional Commits with `🗿 MoAI` trailer; card id `t708` in every commit message body; no `--no-verify`; lane does NOT push `develop` (lead batch-pushes).
- **B11 Subagent boundary**: return structured blocker reports; never prompt the user.

## §C Pre-flight

Run before the first code commit, after the develop absorb (§A):

```bash
# 1. Branch + HEAD re-read (staleness rule — re-run immediately before every commit)
git branch --show-current
git rev-parse --short HEAD

# 2. Symbol-pin re-verification against the absorbed tree
#    Exported pins via go doc; unexported pins via declaration grep.
#    The M2 characterization-suite compile is the compile net for unexported pins.
go doc ./internal/gateway/receipt Manifest.Check
go doc ./internal/gateway/conversation Manager.Fork
grep -n "func Decode\|func BindToolID\|func RestoreToolID" internal/gateway/opaque/codec.go
grep -n "func (h \*receiptHistory) Check(\|func (h \*receiptHistory) checkObserved(\|func (h \*receiptHistory) Publish(\|func replayCause\|func (h \*receiptHistory) observations(" internal/gateway/translate/receipt_history.go
grep -n "thinking\", \"redacted_thinking" internal/gateway/receipt/projection.go
grep -n "func (m \*Manager) refreshNative" internal/gateway/conversation/native.go
grep -n "func prepareGatewayConversation" internal/cli/gateway_session.go

# 3. Characterization baseline — the validator suite must be green BEFORE any change
go test ./internal/gateway/translate/ -run TestReceiptHistory -count=1
go test ./internal/gateway/receipt/ -count=1

# 4. Cross-SPEC conflict pre-scan on the touched packages
grep -rn "Retired\|superseded" internal/gateway/translate/ internal/gateway/receipt/ internal/gateway/conversation/ internal/cli/ || echo "no conflicts"

# 5. Scoped lint baseline
golangci-lint run --new-from-rev=HEAD internal/gateway/... internal/cli/... 2>&1 | tail -5
```

Record each command and its observed output in `progress.md §E.2` as the run-phase baseline attribution.

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (zero-diff files)**: `internal/gateway/receipt/core.go`, `projection.go`, `store.go` (all of it); `internal/gateway/translate/receipt_history.go` — `Check`, `observations`, `Publish`, `checkObserved`, `replayCause`, `CauseReasoning` classification, `HistoryReplayError`, `historyReplayGuidance`, and the reason-clause strings; `internal/gateway/translate/request.go` (entire file — the production authorization call ordering); `internal/gateway/translate/response.go`, `stream.go`, `reasoning.go` (envelope issuance/emission — untouched); `internal/gateway/opaque/codec.go` (entire file — the canonicality the security determination rests on); `internal/gateway/conversation/family.go` — `Manager.Fork`; `internal/cli/gateway_factory.go`. The card's gateway-side diff MUST be empty for all of these, file-level (AC-EVR-011).
- **Forbidden directions** (automatic design reject): validator relaxation of any kind; item-level validation; request-driven trigger/authorization; receipt-store read or write from the repair path; envelope manufacture/re-encoding; wire-body wording changes; automatic surgery on 400; public-content edits riding along with a repair.
- **Live-probe discipline** (only if an integration scenario is executed): t672 §0 isolation verbatim — ephemeral loopback ports, per-run session token, isolated receipt store under `/tmp` (per-run UUID), child stopped via `ChildProcess.Stop` (bounded) on every path, real credentials read-only in place, bounded upstream cost (`max_tokens <= 64`). **No background load.**
- **No local full-suite runs** (`go test ./...` prohibited — lane load discipline); run affected packages only; full-suite verdict is CI's.
- Cross-platform: launcher-side changes build under `GOOS=windows GOARCH=amd64 go build ./...`.

## §E Self-Verification (delegation deliverables, E1-E8)

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §E, with the attribution triple (command + verbatim output + tree SHA) on every item:

- **E1 AC matrix**: every AC-EVR-### PASS/FAIL with the verification command and verbatim output (acceptance.md §D is the SSOT).
- **E2 Cross-platform build**: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`, exit 0 both.
- **E3 Coverage**: `go test -cover ./internal/cli/... ./internal/gateway/...` on touched packages — ≥85%.
- **E4 Subagent boundary grep**: for any touched package, the AskUserQuestion/mcp__askuser grep yields 0 non-test, non-comment matches.
- **E5 Lint**: `golangci-lint run --new-from-rev=HEAD` on touched packages — 0 NEW issues; pre-existing baseline reported separately.
- **E6 Branch HEAD + merge state**: commit SHAs, local develop merge SHA (post-integration), unpushed-commit count — reported to the lead; lane does not push.
- **E7 Blockers**: any pin break after the develop absorb; any t700 seam adjudication outcome requiring spec correction; any AC unresolvable as specified — structured blocker report.
- **E8 RED evidence**: for the repair verb's new tests, the verbatim pre-GREEN failing output (TDD).

## §F Milestones (ordered by decision-reversibility — highest-change-likelihood decisions first)

- **M0 — Seam adjudication gate (Priority High, gate for M3+) — two adjudication inputs**
  Surface spec.md §4 (Reading A vs Reading B on t700's REQ-WRR-007 / REQ-WRR-008(d)) to the orchestrator. If t700's SPEC landed in the absorbed develop, confirm its current wording still leaves Reading B viable or obtain the adjudication; if t700 remains unlanded, record that no conflict materialized and Reading B stands as this SPEC's position. **Second input — card t707's verdict** on the serial-turn chain-class shape (family `1f14d174`, spec.md §3.4/§4.1): if t707 determines the chain-class shape belongs on an envelope-adjacent repair surface, confirm this SPEC's CauseReasoning scope boundary (REQ-EVR-004) still holds or obtain the scope adjudication before proceeding. Output: adjudication record (both inputs) in `progress.md §E.1` addendum. M3+ does not start on a Reading-A adjudication or an adverse t707 scope finding (blocker → re-delegation).

- **M1 — Sibling absorb + pin re-verification (Priority High, gate for everything)**
  Absorb local develop per §A; run §C pre-flight; re-verify every symbol pin and the t700 landing status. Output: pre-flight record in `progress.md §E.2`. Blocker path if any pin broke.

- **M2 — Validator characterization lock (Priority High)**
  Extend the existing cause suites (`receipt_history_cause_test.go` / `receipt_history_classify_test.go` patterns) with the stripped-replay matrix as characterization tests: stripped-envelope+surviving-marker replay rejected `CauseReasoning`; envelope-bearing replay accepted; marker-missing / digest-mismatch / public-content-changed / source-gone shapes still rejected with unchanged classes and byte-identical error strings (golden). These are GREEN at arrival — the regression guard that makes "the validator is unchanged" mechanically checkable. References: REQ-EVR-001; AC-EVR-001..003.

- **M3 — Launcher-side repair path (Priority High, the card's new behavior — TDD; gated on M0)**
  RED: failing tests for the repair path per REQ-EVR-002/003/004 — transcript-source capture via the `refreshNative` pattern, marker self-attestation per boundary, byte-exact verbatim injection at original boundaries, refusal on every REQ-EVR-007 precondition, aside-before-replace, durable provenance + single-shot termination (REQ-EVR-006/008), explicit invocation only (REQ-EVR-005), no receipt-store access (REQ-EVR-009). GREEN: minimal implementation on the launcher surface (`internal/cli` conversation flow). The HOW decisions (verb name, flag shape, repair-record storage, transcript plumbing) are made here; the WHAT is fixed by the SPEC. References: AC-EVR-004..010.

- **M4 — Gateway non-invasiveness lock (Priority Medium)**
  Mechanical no-weakening assertion: the card's diff (merge-base develop..HEAD) touches no file in the §D PRESERVE list and reads no receipt-store symbol from the repair path (diff-scope assertion, scripted and recorded). References: AC-EVR-011.

- **M5 — Operator documentation (Priority Medium)**
  REQ-EVR-010: the stripped-replay shape, the repair procedure and bounds, refusal conditions, non-repairable shapes, and the sanctioned fork path. Deliverable path: `.moai/docs/gateway-envelope-repair.md` (ko, per language.yaml `documentation: ko`) — the path AC-EVR-012 tests for.

- **M6 — Integration (Priority High)**
  Scoped verification (affected packages only), conventional commits with card id, local develop merge via the integration window (`moai integration acquire` → merge --no-ff → `release`), completion report to the lead with card id, branch + HEAD, local merge SHA, unpushed count, evidence path. Sync (SPEC close) completes **before** the develop merge.

## §G Anti-Patterns

- **Inject-until-accepted loops** — the single-shot bound is the security property, not a UX nicety; iteration turns repair into an oracle probing where Required candidates sit.
- **Silent automatic repair on 400** — the classification identifies the shape; a human authorizes the repair (REQ-EVR-005).
- **Manufacturing or "normalizing" an envelope** — re-serializing a decoded envelope produces a byte-different carrier that `opaque.Decode`'s canonicality check (and the marker digest) must reject; only verbatim transcript bytes are ever injected (REQ-EVR-002).
- **Reading the receipt store from the repair path "just to verify"** — marker self-attestation plus the unchanged Check is the complete decision chain; a store read couples repair to gateway internals and drifts toward the t700 REQ-WRR-003-4 ban (research C5 resolved by REQ-EVR-003-4).
- **Repairing public content "while we're in there"** — the Prefix-exclusion fact (projection.go) is what makes envelope re-injection Prefix-neutral; any public-content edit rides into a different binding and must refuse instead (REQ-EVR-003-3/007).
- **Pinning line numbers** — t700 and the develop absorb move lines; only symbol names survive.
- **Local `go test ./...`** — lane load discipline; CI owns the full suite.

## §H Resolution Record

Dispositions for the three formerly-open plan questions, per the lead conditional-Kickoff directive (2026-09-13) and the acceptance.md §D.4-3 disposition clause. Each disposition follows the audited design's own bounded path — none changes the design; these ARE the §D.4-3 dispositions, so M3 entry is not blocked on them.

| # | Topic | DECISION | Basis | Resolution route |
|---|---|---|---|---|
| 1 | Public-content equality on the failing switch-turn replay (whether the t703-class client re-encode also alters public content) | **Proceed to M3 without waiting** — the repair's Prefix-class refusal (REQ-EVR-007) is the designed behavior for either answer; the open question changes only observed refusal frequency, never the design | t703 verdict.md Gaps + spec.md §3.4 | The next production occurrence's classified 400 body (t703 diagnosability closure), or an optional run-phase capture; occurrences are recorded in progress.md, never design-changing |
| 2 | Transcript envelope retention under client compaction / `--continue` slicing | **Proceed** — the repair refuses when the source is gone (REQ-EVR-007): a degradation path, not a design dependency | Research synthesis gap 5 + REQ-EVR-007 | Optional run-phase retention probe against a compacted family transcript; result recorded, design unchanged |
| 3 | Scope boundary vs the serial-turn chain-class shape (card t707, family `1f14d174`) | **This SPEC's scope stays CauseReasoning (REQ-EVR-004) until t707's verdict arrives**; the allocation is adjudicated at M0 as its second input — if t707 assigns the shape to a follow-up SPEC, M0 records that and M3+ proceeds unblocked on this SPEC's scope | spec.md Out of Scope block + §3.4 row + plan.md M0 two-input gate | t707 verdict via the lead; M0 adjudication record (both inputs) in progress.md §E.1 addendum |

## §I Cross-References

- `.moai/reports/t708/research.md` — 4-lens synthesis + orchestrator transcript-retention addendum (the Route-C content-source proof).
- `.moai/reports/t703/verdict.md` — incident ground truth (strip mechanism, `CauseReasoning` classification).
- `.moai/reports/t672/{verdict,investigation,matrix}.md` — binding invariant, rejection taxonomy, isolation discipline.
- `SPEC-GATEWAY-WEDGE-REROOT-001` (card t700, sibling worktree, unlanded) — re-rooting policy, REQ-WRR-007/008(d) seam (spec.md §4), invocation/authorization prior art for launcher-driven history intervention (prior art for the invocation pattern only — NOT for content repair; research C3).
- `SPEC-MOAI-GATEWAY-001` — the gateway base SPEC.
- Card t707 (lane-5) — serial-turn `CauseChain` 400 reproduction (family `1f14d174`); its verdict is M0's second adjudication input, dispositioned in plan.md §H Resolution Record (row 3).
- BerriAI/litellm #40288 — external prior art for reasoning-signature preservation + inbound restoration (byte stability across turns); recorded in spec.md §4.1, applicability unestablished pending t707.
- `.claude/rules/local/gitflow-lane-protocol.md` — lane integration procedure, serialization window, push policy.
- `.claude/rules/moai/core/verification-claim-integrity.md` — §E attribution requirements.

# t300 verdict — VCI §2.3 ordering attribution (baseline-first AC recurrence prevention)

Card: t300 (G2a, lane-9) · Branch: WT-baseline-order · Base: local develop b7462203a (absorbed via fast-forward)
Motivating incident: SPEC-V3R6-GRAPH-FRESHNESS-001 AC-GF-022 — the baseline artifact `.moai/reports/t250/m5-baseline.md` and the M5 implementation it measured shared commit `7f2e9e77d`; git cannot witness authoring order inside one commit, so the AC's ordering clause ("a measurement date preceding the first implementation commit") is permanently unverifiable (recorded deviation, operator decision 2026-08-27).

## Claim

The doctrine now states the recurrence-prevention rule — a baseline artifact that an acceptance criterion's ordering premise depends on MUST land in its own commit that precedes the change's commit; a commit message ASSERTS ordering, only the commit graph witnesses it — and that statement is mechanically protected in BOTH trees (repo copy + template mirror) so its deletion fails CI.

## Evidence

1. Clause landed in both trees:
   - Repo copy: `.claude/rules/moai/core/verification-claim-integrity.md` — new `### 2.3 Ordering attribution — the commit graph is the only sequencing witness` between §2.2 and §3, carrying the motivating-instance provenance (SPEC ID, AC token, artifact path, commit SHA, card id).
   - Template mirror: `internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md` — same clause, motivating-instance sentence sanitized per §25 (no SPEC-ID / AC token / SHA / internal path).
2. New guard `internal/template/vci_ordering_clause_guard_test.go` (TestVCIOrderingClausePresence):
   reads BOTH trees, requires the operative sentence in both, requires the provenance token in the repo copy only, forbids the four internal tokens in the mirror (§25), and hard-fails naming the tree when the `### 2.3` heading is absent.
3. GREEN (post-fix), verbatim: `.moai/reports/t300/guard-green.log`
   `go test ./internal/template/ -run 'TestVCIOrderingClausePresence|TestSanitizedPairParity' -count=1 -v` → both `--- PASS`, `ok ... 0.424s`. VCI sanitized pair: 12 reword pairs, net one-sided=0 (tolerance 4) — the new clause pairs across trees after token normalization.
4. RED (guard fired on a known failing input, verbatim): `.moai/reports/t300/guard-red-mutant.log`
   Mutant = mirror copy with the `### 2.3 Ordering attribution` heading line removed →
   `--- FAIL: TestVCIOrderingClausePresence` / `VCI_ORDERING_CLAUSE_MISSING: template-mirror copy ... carries no "### 2.3 Ordering attribution" heading` / exit 1.
   Mutant was applied twice (first for the conversational observation, second to capture the persisted log), restored via byte-precise re-insertion each time (NOT git restore — the clause is uncommitted work), and the final tree re-observed green (`ok ... 0.478s`).
5. Template-First build cycle: `make build` → EXIT=0 (agents-emit-check + templ-generate + gen-catalog-hashes + go build); catalog.yaml unchanged (rule files are not skill-catalog entries); post-build tracked diff is exactly the two VCI files.

## Baseline-attribution

All measurements this run, this tree (WT-baseline-order @ local develop b7462203a, absorbed before work):
- GREEN: `go test ./internal/template/ -run 'TestVCIOrderingClausePresence|TestSanitizedPairParity' -count=1 -v` → exit 0, both tests PASS (guard-green.log).
- RED: same selector on the heading-stripped mutant → exit 1, FAIL naming template-mirror (guard-red-mutant.log).
- Build: `make build` → exit 0.

## Gaps (explicitly NOT observed)

- Full `go test ./...` was NOT run — lane discipline scopes verification to touched packages (`go test ./internal/template/...` subset above); the full-suite verdict belongs to CI on the pushed develop head.
- `go vet` was not run separately on this tree (no non-test Go source changed; the new file is a `_test.go` in an existing package compiled by the test run itself).
- No observation that future actors OBEY the clause — a policy-layer doctrine's behavioral surface is the next baseline-first AC authored after this lands; what is proven here is presence + CI-enforced non-deletion.
- The escaped original commit `7f2e9e77d` remains unreachable from integrated history; nothing in this card re-witnesses its ordering (permanently unverifiable per the recorded deviation — this card prevents recurrence, it does not repair the original).

## Residual-risk

- A retitle of §2.3 that preserves the body but changes the heading string would trip the guard (by design — heading is the anchor), but a reword that keeps the heading and drops the operative sentence is caught only by the operative-literal check, which matches one exact sentence; a wholesale rewrite introducing a DIFFERENT operative sentence would pass the guard while changing the doctrine. Accepted: doctrine rewrites are reviewable events, and the guard's job is non-deletion, not prose review.
- Actors who never load VCI (non-MoAI toolchains) are outside the clause's reach — no mechanical writer-side check exists for "baseline committed before implementation" as a git-level gate; that heavier enforcement (a CI commit-topology check) was rejected here as over-engineering with a false-positive surface (reports legitimately ride implementation commits).

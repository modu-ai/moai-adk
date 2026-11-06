# t2107 progress — reproduction plan only (implementation waits for the batch FF; t2087 edits the same file)

## Why this card exists
`OwnershipTransitionRule` (`internal/spec/lint_ownership.go`) reports `OwnershipTransitionInvalid` when the commit that carried a SPEC's last `status:` change has a first `Authored-By-Agent:` trailer that is not the matrix owner. On a SQUASH merge the commit body concatenates the trailers of every squashed sub-commit, and the rule reads only the FIRST one, so the verdict depends on sub-commit order, not on who changed the status. t2105 silenced one instance with `lint.skip` on SPEC-FACTORY-COMPLETION-RECOVERY-001 (temporary); this card fixes the rule and removes that skip.

## Facts measured in this tree (HEAD at lease time)
- `lookupOwnershipTransitionFromGit` runs `git log -<gitLogWindowSize> --follow --format=%H%x00%s%x00%b%x00 -p -- <spec path>`, walks the commits newest-first, and for the first commit whose diff changes the `status:` line builds the record with `AuthoredByAgent: parseAuthoredByAgent(c.body)`.
- `parseAuthoredByAgent` uses `FindStringSubmatch` on `(?mi)^\s*Authored-By-Agent:\s*(\S+)\s*$`, i.e. the FIRST matching line only (`lint_ownership.go` ~255-266).
- Commit `55b20438fd` (`[t1963] fix(factory): completion recovery M2-B ... (#1930)`): 1 parent (squash merge), 50 `Authored-By-Agent` trailer lines in the body; the first trailer is `manager-spec`; the sub-commit that moved the SPEC `draft -> in-progress` carries `manager-develop` (audit t2105, receipt rcpt-ef8cd674f0b26d011db02307).
- Commit `5ef49c4625` (`[t1929] spec-lint common-cause repair ... (#1901)`): 1 parent (squash merge), 5 trailer lines in the body.

## Plan (RED first, after t2087 has landed)
1. Fixtures: build a scratch git repository in a test (`t.TempDir()`) that reproduces the shape, without depending on this repository's history: a SPEC `spec.md` whose status line moves `draft -> in-progress` inside ONE commit whose body contains several trailers with the first NOT the owner and the owner's trailer present later (squash-like), plus the single-trailer control (still judged by that one trailer).
2. RED test A (the false positive): that squash-shaped commit must NOT produce `OwnershipTransitionInvalid`. Expected failure on today's code: the rule reports Invalid because the first trailer is `manager-spec`.
3. RED test B (no regression): a single-trailer commit with a wrong owner still reports `OwnershipTransitionInvalid`; a no-trailer commit still reports `OwnershipTransitionUnmeasured`.
4. Decision to take before coding (mark as a recorded judgment in the card record): of the two options in the card text, prefer (b) "a multi-trailer squash body is `unmeasured`" over (a) "read the trailer of the sub-commit that carried the status change", because the sub-commit is not recoverable from the squash commit alone (it is gone from the first-parent history) and guessing it would re-introduce an unobserved attribution; (a) stays possible only if the diff hunk can be tied to a trailer, which it cannot in a squash body. The measurement state is already a first-class finding (`OwnershipTransitionUnmeasured`, Info), so no new code is needed beyond detecting "more than one distinct actor in the body" as ambiguous.
5. Detection rule to implement: collect ALL trailers in the body (not the first); if the set has exactly one distinct actor, judge as today; if it has more than one distinct actor, return an ambiguous record that the rule reports as `OwnershipTransitionUnmeasured` (never Invalid); if none, unchanged (`Unmeasured`).
6. Real-history check: run the rule (binary built from this tree) over SPEC-FACTORY-COMPLETION-RECOVERY-001 (commit `55b20438fd` in its history) and confirm Info `OwnershipTransitionUnmeasured` instead of the warning, with the t2105 `lint.skip` removed.
7. Remove the `lint.skip` block and its comment from `SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md` (frontmatter edit is a non-transition correction — owner `manager-spec` through re-delegation per `spec-frontmatter-schema.md` § Non-transition frontmatter corrections), `updated:` refreshed; baseline file `.moai/spec-lint-baseline.json` must stay unchanged; run the whole-corpus lint with the baseline (about 33 minutes at the last measurement — run once, in the background).
8. Update the schema rule text that describes the three finding codes only if the semantics of `Unmeasured` change (the SSOT `spec-frontmatter-schema.md` says it is emitted when "its commit carries no trailer"; extend that sentence for "carries trailers of more than one actor").

## Gaps (at plan time)
- Nothing implemented or run here yet; the plan is from reading `lint_ownership.go` and measuring the two commits.
- t2087 (in the batch) also edits `lint_ownership.go` (serial `git log` per SPEC); start this card only after the FF and re-read the file then, because the structure may have moved.
- The corpus-wide lint run and the real-history check need a built binary from the final tree.

## Residual-risk
Treating a multi-actor squash body as `unmeasured` reduces the rule's reach for squash-merged SPEC transitions (they become Info, not Warning). That is the honest state: ownership cannot be attributed from a squash body. A stronger option (tie the status hunk to its sub-commit trailer) would need the pre-squash history, which this repository does not keep on the first-parent line.

## Implementation (after the batch FF; main b25adf4fca absorbed --no-ff)
- RED/GREEN: `internal/spec/lint_ownership_squash_t2107_test.go` (multi-actor squash body -> Info Unmeasured naming the actors; one distinct actor repeated stays judged, wrong one stays Invalid; single trailer and no trailer unchanged; batched lookup equals the per-SPEC oracle; `distinctAuthoredByAgents` unit). Pre-fix the multi-actor body produced Invalid (first trailer rule); observed: the first RED run failed on the actor-naming assertion once record wiring existed, and the pre-wiring behaviour is the measured 55b20438fd false positive.
- Code: `distinctAuthoredByAgents` + `applyTrailerAttribution` + record field `AmbiguousActors` in `lint_ownership.go`; both lookups (per-SPEC and batched `lint_ownership_batch.go`) use the helper; `Check` reports Unmeasured for ambiguous bodies before the no-trailer branch.
- Verification: `go test ./internal/spec -count=1` exit 0 (whole package); focused `-run 'Ownership|LintGitBatch|AuthoredBy|SplitGitLog'` ok.
- Real history (binary built from this tree, by path): `moai spec lint SPEC-FACTORY-COMPLETION-RECOVERY-001` -> `INFO OwnershipTransitionUnmeasured ... commit 55b20438fd...` (no warning).
- Doc: spec-frontmatter-schema.md Unmeasured sentence extended (local + template mirror).

## Gaps (implementation)
- The `lint.skip` of t2105 is NOT in this tree (t2105 is not absorbed; it was left out of the batch), so its removal cannot be done here; it must follow t2105's landing (one-line frontmatter correction through manager-spec re-delegation). The real-history check above ran without any skip.
- Corpus-wide lint with the baseline (long run) not yet run; the baseline file is untouched.
- Template/local parity test and `go vet` not yet run on the final tree.

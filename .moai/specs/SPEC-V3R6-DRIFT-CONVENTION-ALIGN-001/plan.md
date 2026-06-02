# Implementation Plan — SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001

> Tier S — `internal/spec` drift classifier convention alignment.
> Plan-phase only; this document is the run-phase implementation oracle.

## A. Context

`moai spec drift` reports 66 DRIFT entries; ~43 are false-positives where frontmatter `status: completed` is contradicted by a git-implied status of `implemented` / `in-progress` / `planned`. `moai spec audit` reports the same SPECs as CLEAN, confirming the drift classifier — not the SPEC state — is wrong.

The defect is a **commit-convention mismatch**: the drift classifier recognizes `completed` only via a `docs(sync):` prefix, but the canonical Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md`) closes SPECs with `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close` and syncs them with `docs(SPEC-{ID}): sync-phase artifacts`. Neither matches `docs(sync):`, and the close commit is additionally swallowed by the over-broad generic `chore` prefix → misclassified `in-progress`.

Full root-cause analysis (pre-verified, not re-investigated) is in `spec.md` §A.

## B. Known Issues / Constraints in the Affected Code

- `internal/spec/transitions.go` `transitionRules` is an **ordered slice** — longer prefixes MUST precede shorter ones (Go map iteration is random; the slice is the deterministic source). Any new rule for `chore(SPEC-...)` or the close infix MUST be inserted before the generic `chore` rule (current line 35) to win the `HasPrefix` race.
- `ClassifyPRTitle()` lowercases the title (`strings.ToLower`) before `HasPrefix`. All new prefix/infix literals MUST be lowercase to match (e.g., `chore(spec-`, `4-phase close`, `mx-phase audit-ready`).
- `getGitImpliedStatus()` (`drift.go`) is the **only** call site that walks commits newest-first and stops at the first non-empty status. `ClassifyPRTitle` is also called by `StatusGitConsistencyRule` in `lint.go` — any change to `transitionRules` affects BOTH. The run-phase implementer MUST confirm the lint rule's behavior is not regressed (it should improve identically, since both consume the same classifier).
- The `shouldSkipCommitTitle()` helper in `drift.go` skips `chore(spec):` / `chore(specs):` (literal colon). The SHA-backfill commit `chore(SPEC-X-001): backfill ...` does NOT match this skip (different char after `spec`), so it is currently NOT skipped — it falls to generic `chore`. The fix MUST handle this backfill-chore case (either skip it OR classify the real close commit ahead of it).

## C. Pre-flight Checklist (run-phase entry)

- [ ] Confirm HEAD is at or descends from `e5574f1ab` (iter-1 plan-phase commit; iter-2 remediation commit will be newer) before editing.
- [ ] `go test ./internal/spec/...` is green at baseline (capture the pre-fix pass list, especially `TestClassifyPRTitle_ChoreSpecUnchanged` and the LSGF-001 word-boundary tests).
- [ ] Capture baseline `moai spec drift --count` (expected: **67**, D3-corrected) and `moai spec audit` (expected: clean) for before/after comparison.
- [ ] Re-verify the 4 AC-DCA-001 named exemplars' newest exact-token commit is still the close-infix (or a backfill chore immediately newer than it) at run-phase HEAD — if any was re-touched, substitute the backup exemplar `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001`.

## D. Constraints (DO NOT VIOLATE)

### D.0 plan-audit iter-1 verdict + iter-2 remediation (recorded per task constraint; no progress.md for Tier S)

- **iter-1 plan-auditor verdict**: **FAIL 0.62** (Tier S PASS threshold 0.75). User decision: **RESCOPE** (Tier S minimalism, honest scope).
- **D1 (BLOCKING) resolved**: infix-only fix reached only ~12-20 of the 43 false-positives. The 43 are HETEROGENEOUS (3 sub-classes — see spec.md §A.1.1). Rescoped in-scope to sub-class (1) canonically-closed ONLY; sub-classes (2) legacy-convention and (3) genuine-incomplete-close carved out in spec.md §B.
- **D2 (BLOCKING) resolved**: 3 of 4 iter-1 exemplars (`SPEC-DESIGN-001`, `SPEC-V3R2-ORC-003`, `SPEC-V3R3-CI-AUTONOMY-001`) lack the close-infix → would STILL drift post-fix. Replaced with 4 git-log-verified close-infix exemplars (spec.md §3 AC-DCA-001 table). Also rejected `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001` (newest=`fix` over close) and `SPEC-V3R6-AGENT-TEAM-REBUILD-001` (newest=generic `chore`) during verification — documented as D2 evidence.
- **D3 (SHOULD-FIX) resolved**: baseline corrected 66 → **67** (the plan commit added 1 self-drifting draft SPEC).
- **D4 (MINOR) resolved**: added AC-DCA-008 backfill-no-regression synthetic fixture (REQ-DCA-010).

### D.1 Hard constraints

- Observation-only — this fix touches the **classifier** (`transitions.go`) and the **walker** (`drift.go`), both of which already perform classification, not mutation. No file-modify primitives are introduced.
- Preserve AC-LSCSK-003 (metadata-sweep skip) and LSGF-001 (word-boundary filter) behavior exactly.
- No new commit-subject convention is introduced — the classifier is aligned to the existing SSOT.
- The narrow backfill-skip MUST be scoped to SPEC-ID-scoped chores whose subject contains `backfill` ONLY. It MUST NOT broaden to all `chore(SPEC-...)` and MUST NOT touch the `chore(spec):`/`chore(specs):` metadata-sweep skip.
- [ANTI-GOAL] The fix MUST NOT make the walker infer `completed` for any SPEC lacking a close-infix in its exact-token history (sub-class 3 genuine-incomplete-close stays drift — REQ-DCA-005 / AC-DCA-008).
- Table-driven tests per `internal/spec` convention (`[]struct{ name, content/title string; want... }` + `t.Run`).
- Go 1.23+, golangci-lint clean, coverage ≥ 85% for modified files.

## E. Self-Verification (run-phase exit gate)

The run phase is complete only when every AC in `spec.md` §3 passes:

```bash
# AC-DCA-007 — package suite green
go test ./internal/spec/...

# AC-DCA-005 / AC-DCA-006 / AC-DCA-008 — regression guards + backfill-no-regression
go test -run 'TestClassifyPRTitle_ChoreSpecUnchanged|TestCommitMatchesSPECID|LSGF|Backfill' ./internal/spec/...

# AC-DCA-001 (PRIMARY) — the 4 named verified-close-infix exemplars no longer DRIFT
moai spec drift --json | jq -r '.Records[] | select(.Drifted) | .SPECID' \
  | grep -E 'SPEC-V3R5-GIT-STRATEGY-SCHEMA-001|SPEC-V3R6-CI-FLAKY-STABILIZE-002|SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001|SPEC-V3R6-CI-FLAKY-STABILIZE-001' \
  && echo "FAIL: a named exemplar still DRIFT" || echo "PASS: all 4 exemplars aligned"

# AC-DCA-001b (SECONDARY) — count strictly decreases from baseline 67
test "$(moai spec drift --count)" -lt 67 && echo "PASS: count < 67" || echo "FAIL: count not decreased"

# AC-DCA-006 — audit stays clean (0 new MUST-FIX)
moai spec audit
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — Write failing classifier tests (RED) — Priority High

Add table-driven tests to `internal/spec/transitions_test.go` (classifier) + a focused walker test in `internal/spec/drift_test.go` (the walker test file does not yet exist — create it) that encode AC-DCA-002, AC-DCA-003, AC-DCA-004, AC-DCA-008:

- (AC-DCA-002) `chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal + 4-phase close` → `completed`.
- (AC-DCA-003) `chore(SPEC-EXAMPLE-001): backfill §E.2/§E.5 commit SHA` → SPEC-ID-scoped backfill chore: **skipped by the narrow backfill-skip** (so the walker proceeds to the real close commit). `chore(spec): frontmatter status sweep` → skip-meta (empty), UNCHANGED (AC-LSCSK-003).
- (AC-DCA-004) Newest-first walker fixture: `chore(...): backfill` (skip) → `chore(...): 4-phase close` → `docs(...): sync-phase` → `feat(...)` resolves to `completed` (close-infix wins before the `docs`→`in-progress` rule).
- (AC-DCA-008 / D4 backfill-no-regression) Newest-first walker fixture with **NO close-infix**: `chore(SPEC-Y-001): backfill §E.2 sync_commit_sha — abc1234` (skip) → `docs(SPEC-Y-001): sync-phase artifacts` → `feat(SPEC-Y-001): M1 ...`, frontmatter `implemented` → resolves to `implemented` (NOT `completed`). Asserts the backfill-skip does not invent a `completed` when no close commit exists.

Confirm these FAIL against the current classifier.

### M2 — Classifier + walker fix (GREEN) — Priority High

Two design options; the run-phase implementer selects after confirming M1 trace. **Decision recorded here, not deferred:**

**Primary design (preferred):** In `transitions.go` `transitionRules`, insert — ordered BEFORE the generic `chore` rule — a rule that recognizes the canonical close. Because the close subject is `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close`, a literal prefix `chore(spec-` is NOT specific enough (it would also catch the backfill/sync `chore(SPEC-...)` variants). Therefore the close MUST be matched by the **`4-phase close` / `mx-phase audit-ready` infix**, not a prefix. This requires extending `ClassifyPRTitle` to do an infix check for the close marker before the prefix loop:

```
if strings.Contains(lowerTitle, "4-phase close") ||
   strings.Contains(lowerTitle, "mx-phase audit-ready") {
    return "mx-close", "completed", nil
}
```

placed AFTER the revert check and BEFORE the prefix loop. This is convention-anchored (the infix comes verbatim from the Ownership Matrix close subject) and order-independent of the generic `chore` prefix.

**Disambiguation (REQ-DCA-002 / REQ-DCA-003):** The generic `chore` → `in-progress` rule remains for genuine partial-work chores, but the SHA-backfill commit `chore(SPEC-X-001): backfill ...` must not pin the walker to `in-progress` ahead of the real close commit. Two robustness sub-options — choose per M1 evidence:

- (a) **Skip the backfill chore in `shouldSkipCommitTitle`** by recognizing the `backfill` lifecycle-neutral marker for SPEC-ID-scoped chores (narrow: only when the chore is SPEC-ID-scoped AND contains `backfill`), so the walker proceeds to the real close commit. MUST keep `chore(spec):`/`chore(specs):` skip intact.
- (b) Rely on the infix close-rule alone — but note the backfill chore is NEWER than the close commit in the walker order, so without (a) the walker still hits the backfill `chore` (generic → `in-progress`) FIRST. **Therefore (a) is REQUIRED** for the GIT-STRATEGY-SCHEMA exemplar (whose newest commit is the backfill chore). The run-phase implementer MUST implement (a) — the infix rule alone is insufficient when a post-close backfill chore exists.

**No-completed-without-close anti-goal (REQ-DCA-005 / AC-DCA-008):** The fix recognizes the close-infix as a *positive* `completed` signal ONLY. It MUST NOT infer `completed` from sync/feat commits when no close-infix exists. Concretely: the narrow backfill-skip, after stepping over a backfill chore, exposes whatever commit is beneath — if that is a sync `docs` (→`implemented`) and no close-infix exists anywhere, the walker returns `implemented`, NOT `completed`. This keeps sub-class (3) genuine-incomplete-close SPECs as real drift (spec.md §B). Do NOT add a blanket `docs(SPEC-{ID}): sync-phase` → `implemented` mapping in this SPEC unless a failing in-scope fixture requires it; for the 4 named exemplars the close-infix is the resolving rule, so the existing `docs`→`in-progress` generic rule is left untouched (the close-infix wins above it in walker order). Any sync→implemented refinement belongs to the deferred legacy-convention follow-up SPEC.

### M3 — Regression + named-exemplar verification (REFACTOR + verify) — Priority High

- Run AC-DCA-005 (chore-skip), AC-DCA-006 (LSGF-001 + audit clean), AC-DCA-007 (full suite), AC-DCA-008 (backfill-no-regression).
- Run AC-DCA-001 (PRIMARY): the 4 named exemplars (`SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`, `SPEC-V3R6-CI-FLAKY-STABILIZE-002`, `SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001`, `SPEC-V3R6-CI-FLAKY-STABILIZE-001`) MUST be absent from the `moai spec drift --json` DRIFT set. Re-verify each exemplar's newest exact-token commit at run-phase HEAD before asserting; if any was re-touched, substitute the backup `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001`.
- Run AC-DCA-001b (SECONDARY): `moai spec drift --count` strictly `< 67`. Do NOT assert an exact band — the named-exemplar AC is the binding signal. If the count does not decrease, diagnose whether the close-infix/backfill-skip is actually firing (the 4 exemplars are the canary) and record findings; do NOT widen scope to sub-classes (2)/(3).

## G. Anti-Patterns to Avoid

- **AP-1**: Adding a `chore(spec-` prefix rule before generic `chore`. This over-broadly catches backfill/sync `chore(SPEC-...)` variants and mislabels them `completed`. Use the verbatim close infix instead.
- **AP-2 (genuine-incomplete-close masking — CRITICAL)**: Inferring `completed` for a SPEC that has NO close-infix in its exact-token history (e.g., via a blanket `docs(SPEC-...)→completed` or `feat→completed` mapping, or by over-broadening the backfill-skip so the walker falls through to a wrong rule). This masks sub-class (3) real drift (spec.md §B) and is an explicit anti-goal (REQ-DCA-005 / AC-DCA-008). The close-infix is the ONLY positive `completed` signal.
- **AP-3**: Broadening `shouldSkipCommitTitle` to skip ALL `chore(SPEC-...)` commits — this would skip genuine partial-work chores and the close commit itself. Scope the skip narrowly to SPEC-ID-scoped chores whose subject contains `backfill`. MUST keep `chore(spec):`/`chore(specs):` skip intact.
- **AP-4**: Editing `transitionRules` ordering without a test asserting the prefix/infix precedence (Go slice order + the pre-loop infix check are the only guards).
- **AP-5**: Touching `ExtractSPECIDs` / `commitMatchesSPECID` (LSGF-001) — out of scope, preserve verbatim.
- **AP-6**: Folding legacy close conventions (`sync(specs):`, `docs(spec): ...closure`) into this SPEC — that is the deferred `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001` scope. Re-introducing heterogeneity here is what caused iter-1 FAIL.
- **AP-7**: Lowering the coverage gate or `t.Skip()`-ing a failing regression test to make the suite green.

## H. Cross-References

- `spec.md` §A — full root-cause analysis (pre-verified).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — canonical close/sync commit subjects (SSOT this fix aligns to).
- `internal/spec/drift.go` `getGitImpliedStatus` / `shouldSkipCommitTitle` — walker + skip filter.
- `internal/spec/transitions.go` `transitionRules` / `ClassifyPRTitle` — classifier.
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 AC-LSCSK-003 — metadata-sweep skip regression guard (MUST preserve).
- SPEC-V3R4-LINT-SPECID-GREP-FIX-001 LSGF-001 — word-boundary SPEC-ID filter (MUST preserve).
- `internal/cli/spec_drift.go` — `moai spec drift --count` / `--json` surface (verification only; not modified).

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

- [ ] Confirm HEAD is at or descends from `a3dcc6b31` (plan-phase commit) before editing.
- [ ] `go test ./internal/spec/...` is green at baseline (capture the pre-fix pass list, especially `TestClassifyPRTitle_ChoreSpecUnchanged` and the LSGF-001 word-boundary tests).
- [ ] Capture baseline `moai spec drift --count` (expected: 66) and `moai spec audit` (expected: clean) for before/after comparison.

## D. Constraints (DO NOT VIOLATE)

- Observation-only lint rules — but this fix touches the **classifier** (`transitions.go`) and the **walker** (`drift.go`), both of which already perform classification, not mutation. No file-modify primitives are introduced.
- Preserve AC-LSCSK-003 (metadata-sweep skip) and LSGF-001 (word-boundary filter) behavior exactly.
- No new commit-subject convention is introduced — the classifier is aligned to the existing SSOT.
- Table-driven tests per `internal/spec` convention (`[]struct{ name, content/title string; want... }` + `t.Run`).
- Go 1.23+, golangci-lint clean, coverage ≥ 85% for modified files.

## E. Self-Verification (run-phase exit gate)

The run phase is complete only when every AC in `spec.md` §3 passes:

```bash
# AC-DCA-007 — package suite green
go test ./internal/spec/...

# AC-DCA-005 / AC-DCA-006 — regression guards
go test -run 'TestClassifyPRTitle_ChoreSpecUnchanged|TestCommitMatchesSPECID|LSGF' ./internal/spec/...

# AC-DCA-001 — drift count band (≤30 AND ≥18) + exemplars no longer DRIFT
moai spec drift --count
moai spec drift --json | jq '.Records[] | select(.Drifted) | .SPECID' | grep -E 'GIT-STRATEGY-SCHEMA-001|SPEC-DESIGN-001|V3R2-ORC-003|V3R3-CI-AUTONOMY-001' && echo "FAIL: exemplar still DRIFT" || echo "PASS: exemplars aligned"

# AC-DCA-006 — audit stays clean
moai spec audit
```

## F. Milestones (priority-ordered, no time estimates)

### M1 — Write failing classifier tests (RED) — Priority High

Add table-driven tests to `internal/spec/transitions_test.go` (and a focused walker test) that encode AC-DCA-002, AC-DCA-003, AC-DCA-004:

- `chore(SPEC-EXAMPLE-001): Mx-phase audit-ready signal + 4-phase close` → `completed`.
- `chore(SPEC-EXAMPLE-001): backfill §E.2/§E.5 commit SHA` → lifecycle-bearing (per chosen design: either skipped so the real close wins, or classified as a non-`in-progress` value).
- `chore(spec): frontmatter status sweep` → skip-meta (empty) — UNCHANGED.
- Newest-first walker fixture: close → sync `docs` → feat resolves to `completed`.

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

**Robustness mapping (REQ-DCA-005):** Additionally add the sync mapping `docs(SPEC-{ID}): sync-phase artifacts` → `implemented` only if M1 reveals a SPEC whose newest classifiable commit is the sync `docs` (i.e., closed-but-no-Mx). For `status: completed` SPECs this is not the resolving rule (the close infix wins), but it improves correctness for `status: implemented` SPECs. Implement if a failing fixture requires it; otherwise document as deferred in `progress.md`.

### M3 — Regression + count verification (REFACTOR + verify) — Priority High

- Run AC-DCA-005 (chore-skip), AC-DCA-006 (LSGF-001 + audit clean), AC-DCA-007 (full suite).
- Run AC-DCA-001 (`moai spec drift --count` in band, exemplars aligned).
- If count is outside the ≤30/≥18 band, do NOT widen the band — instead diagnose whether a residual false-positive class exists and record it in `progress.md` for a follow-up SPEC (this SPEC's scope is the ~43 `completed` false-positives only).

## G. Anti-Patterns to Avoid

- **AP-1**: Adding a `chore(spec-` prefix rule before generic `chore`. This over-broadly catches backfill/sync `chore(SPEC-...)` variants and mislabels them `completed`. Use the verbatim close infix instead.
- **AP-2**: Broadening `shouldSkipCommitTitle` to skip ALL `chore(SPEC-...)` commits — this would skip genuine partial-work chores and the close commit itself. Scope the skip narrowly to the `backfill` marker.
- **AP-3**: Editing `transitionRules` ordering without a test asserting the prefix-precedence (Go slice order is the only guard).
- **AP-4**: Touching `ExtractSPECIDs` / `commitMatchesSPECID` (LSGF-001) — out of scope, preserve verbatim.
- **AP-5**: Lowering the coverage gate or `t.Skip()`-ing a failing regression test to make the suite green.

## H. Cross-References

- `spec.md` §A — full root-cause analysis (pre-verified).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — canonical close/sync commit subjects (SSOT this fix aligns to).
- `internal/spec/drift.go` `getGitImpliedStatus` / `shouldSkipCommitTitle` — walker + skip filter.
- `internal/spec/transitions.go` `transitionRules` / `ClassifyPRTitle` — classifier.
- SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 AC-LSCSK-003 — metadata-sweep skip regression guard (MUST preserve).
- SPEC-V3R4-LINT-SPECID-GREP-FIX-001 LSGF-001 — word-boundary SPEC-ID filter (MUST preserve).
- `internal/cli/spec_drift.go` — `moai spec drift --count` / `--json` surface (verification only; not modified).

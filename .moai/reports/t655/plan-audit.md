# SPEC Review Report: SPEC-WORKTREE-KEY-WIRING-001 (card t655)

Iteration: 2/2 (FINAL — Tier M ceiling per harness.plan_audit_tier_ceilings)
Verdict: **PASS**
Overall Score: **1.00** (Tier M PASS threshold 0.80; iteration-1 was FAIL 0.75 — record preserved below)
Auditor: plan-auditor (lane-4 worktree, branch WT-worktree-keys-wiring, base `1d150a27d`)
Cross-model opinion: not invoked — `.moai/config/` carries no `audit_model` key (measured: `grep -rn 'audit_model' .moai/config/` → 0 hits), so no cross-backend fan-out is warranted by config.

All baseline facts below were re-measured in THIS tree at base `1d150a27d` on 2026-09-12. Reasoning context from the SPEC author was not supplied; audit inputs are the five artifacts under `.moai/specs/SPEC-WORKTREE-KEY-WIRING-001/` plus the code anchors they cite.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-WKW-001..013 at spec.md:58-140: sequential, zero gaps, zero duplicates, uniform zero-padding. AC-WKW-001..012 at acceptance.md:17-97 likewise sequential.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — all 13 REQ entries match GEARS patterns: Ubiquitous (REQ-001/005/009/012/013), compound `[While][Where][When]` (REQ-002, PASS-equivalent per M3), `Where`-gated (REQ-003/004), Event-driven `When...shall` (REQ-006/007/008/010), negative `shall never/shall contain no` (REQ-005/011). Judged against the `REQ-XXX` layer only; the Given-When-Then entries in acceptance.md are the verification layer and are graded under AC-1, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (spec.md:1-16): id, quoted title, quoted semver `"0.1.0"`, `status: draft` (valid enum), ISO `created`/`updated`, author, `priority: P1`, `phase: "v3.2.0"` (release target — no prohibited stage name), module, `lifecycle: spec-anchored`, comma-separated tags. Optional `tier: M` + `depends_on` well-formed. No snake_case aliases.
- **[N/A] MP-4 Section 22 language neutrality** — the SPEC is single-domain CLI/config content; it enumerates no language-specific tooling. Auto-pass per the single-language-scope rule. (Plan §D's template-neutrality constraint for the workflow.yaml comment edit correctly addresses the *programming-language* axis, §15 of CLAUDE.local.md — 16 programming languages neutral, distinct from locales.)
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 5 distinct referenced SPEC-IDs extracted (D7-1): SPEC-SESSION-WORKTREE-001 (×4), SPEC-CONFIG-KEY-HONESTY-001 (×3), SPEC-WORKTREE-BASEREF-001, SPEC-CONFIG-001, self. All 5 exist in `.moai/specs/` (D7-2) and all read `status: completed` (D7-3) — none retired/superseded/archived, so no reconciliation clause is required (D7-4 not triggered; D7-5 not triggered). `depends_on` entries both completed.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` across all five artifacts → 0/0/0/0/0 (D8-4 auto-PASS). The M5 cross-platform check (`GOOS=windows go build`) is present in plan §F M5 and acceptance DoD #2.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → 0 matches in plan.md; research.md absent (Tier M does not require it — N/A-by-tier for that file only, plan.md measured clean).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in one REQ | REQ-WKW-002 (spec.md:63-69) trigger phrasing is ambiguous against REQ-WKW-013 — see D2. All other 12 REQs single-interpretation; guard inventory and seam mechanics are pinned to file:line in design §4-§5. |
| Completeness | 1.0 | 1.0 | HISTORY (spec.md:20-24), §A problem, §B requirements, §C three `### Out of Scope — <topic>` H3 groups each with specific `-` bullets (spec.md:142-158), §D acceptance pointer. Frontmatter 12/12. |
| Testability | 0.75 | 0.75 — one AC's judge not precisely executable | AC-WKW-012's judging command is mechanically defective (D3: zero-sweep selector + inverted exit semantics). The other 11 ACs are binary Given-When-Thens with runnable judges; falsification mutants in acceptance.md §C are concrete. |
| Traceability | 0.50 | 0.50 — multiple REQs lack ACs | REQ-WKW-009 and REQ-WKW-013 have no AC citing them (D1). All 12 AC→REQ citations valid (checked one-by-one incl. the off-by-one trap: AC-WKW-009 correctly cites REQ-008, AC-WKW-010→REQ-010); no orphan ACs. |

Overall = (0.75 + 1.0 + 0.75 + 0.50) / 4 = **0.75** < 0.80 Tier M threshold. No M5 must-pass failure; the FAIL is carried by the three blocking defects below plus the aggregate score.

## Defects Found

D1. **TRAC-REQ009-013-UNCOVERED** — spec.md:108-112 (REQ-WKW-009), spec.md:138-140 (REQ-WKW-013) / acceptance.md §B — REQ-WKW-009 (non-blocking failure family + distinct notice prefix) and REQ-WKW-013 (toggle independence, four-combination matrix) have no acceptance criterion citing them. acceptance.md §A:7-10 promises the prefix-distinctness "is asserted by a test" and plan §F M2 lists the four-combination matrix, but neither is a numbered, judge-carrying AC in the verification layer that spec §D designates. Two of thirteen REQs uncovered. — Severity: major — Class: **blocking** — Required fix: add AC-WKW-013 (REQ-009: prefix distinct from `SessionExitCleanupNoticePrefix`/`PRMergeCleanupNoticePrefix`, asserted by constant-distinctness test) and AC-WKW-014 (REQ-013: 4-combination toggle matrix) with judging commands; Tier M AC ceiling is 16, current 12 → 14 is in budget. Alternatively fold explicit REQ-009/013 citations into AC-WKW-006/008 bodies — but dedicated ACs are cleaner for §E.2 adjudication.

D2. **REQ-WKW-002 DISPOSAL-TRIGGER AMBIGUITY (conflicts with REQ-WKW-013)** — spec.md:63-69 — the trigger reads "**While a session worktree is being disposed** at session exit ... shall merge ... **before disposal proceeds**." Under the literal reading, the merge fires only when disposal happens — i.e. gated on `auto_cleanup: true` — which directly contradicts REQ-WKW-013 ("each of the four combinations is reachable, and enabling one never implies the other"). The implementation intent is independence (AC-WKW-002's Given/When mention only exit, not disposal; plan M2 hooks "BEFORE `cleanupSessionWorktree`" in the caller, outside the toggle check — session_worktree.go:630 toggle check is inside the callee). Two engineers could implement the literal text differently. — Severity: major — Class: **blocking** — Required fix: reword the trigger to "While a session exits with a live session worktree, where `workflow.worktree.auto_merge` is `true`, ..." and render the ordering clause as "before the disposal step (where enabled) proceeds." One-line change; no AC edits needed (AC-WKW-002 is already correct).

D3. **AC-WKW-012 JUDGE DEFECTS (vacuous pass + exit-status inversion)** — acceptance.md:96-97 — two independent mechanical defects in the judging command `go test ./internal/config/ -run TestShippedKey -count=1 → ok && grep -c "AutoMerge has no production reader" internal/config/types.go → 0`:
  - (a) `-run TestShippedKey` matches **zero tests**. Measured: the only test function in `internal/config/shipped_key_reader_test.go` is `TestShippedConfigKeysHaveReaders` (line 70), and `grep -rn 'func TestShippedKey' internal/` returns nothing. Go's `-run` selector silently ignores unmatched names and prints `ok` with exit 0 — an empty sweep reads identically to a full pass (the repo's own recorded lesson: a `-run` selector silently drops names that do not exist). AC-WKW-012 would vacuously pass even if M4 shipped nothing.
  - (b) `grep -c <pattern> file` **exits 1 when the count is 0**, so when the criterion is MET (the stale sentence removed) the `&&` chain reports failure — the report-not-verdict inversion: the printed count says pass, the exit status says fail.
  — Severity: major — Class: **blocking** — Required fix: judge (i) `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1 -v | grep -c '^=== RUN'` ≥ 1 (or simply the correctly-named selector plus a swept-count note in acceptance §A), and judge (ii) `! grep -q 'AutoMerge has no production reader' internal/config/types.go` (exit 0 = criterion met). Also add one sentence to acceptance §A requiring every `-run` judge to record its swept count in §E.2, so the empty-sweep shape is visible rather than silent.

Optional findings (surfaced, orchestrator's discretion — not part of the FAIL basis):

O1. acceptance.md lacks RED-now/green-path adoption cells (verification-completeness §2) — acceptable at plan phase, but the OFF-baseline characterization capture (AC-WKW-001) should become an explicit pre-M1 evidence item in plan §F M1 so its "captured against the pre-M1 tree" claim carries a command+output+SHA record. — Severity: minor — Class: optional.
O2. The auto_create wording change (M3) has no user-visible migration note in the DoD — a user who set `auto_create: true` for the promised creation sees the advisory text change. Add a DoD line requiring the sync-phase CHANGELOG entry to name the wording change. — Severity: minor — Class: optional.
O3. design.md §1.3 / spec §C claim the merge notice lets "the lead's batch-push flow read what landed", but a session-exit stderr notice is ephemeral. No correctness issue — `git push origin develop` sweeps all local develop commits regardless of provenance — but design should say the batch push does NOT depend on the notice (operator-facing only). — Severity: minor — Class: optional.
O4. "Tier M+" prose (plan §A, design header) vs frontmatter `tier: M` — keep the enum value in prose or footnote the "+". — Severity: minor — Class: optional.
O5. The deliberate asymmetry between the auto path (empty develop_branch → skip+notice, REQ-WKW-003) and the manual path (caller-branch fallback + warning, integration.go:128-136 `integrationFallbackWarning`) deserves one sentence in design §2 recording the auto path as intentionally stricter. — Severity: minor — Class: optional.

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Design-Decision Adjudication (the card's three binding decisions — all SURVIVE adversarial review)

1. **Option A (auto-merge acquires the window) — SOUND.** Rationale (a) verified: the session-exit surface has no pre-existing holder, and the same-holder edge case (a session that manually holds the window then exits) resolves to skip+notice under REQ-WKW-004's blanket no-displacement rule — safe in both readings. Rationale (b) verified in-tree: `kanban.AcquireIntegrationLock(projectRoot, want, force)` (internal/kanban/integration_lock.go:236) writes the same record the PreToolUse guard and `moai integration status` read; `integrationLockRoot()` (internal/cli/integration.go:46) resolves the primary checkout in both env and git-fallback paths. Rationale (c) verified: template + `internal/config/defaults.go:876` both carry `AutoMerge: false`, so REQ-SW-008's default-manual posture is honored by the toggle, and the enabled path performing acquire→merge→release writes strictly more serialization record than a skip-ceremony variant. Never-force is load-bearing and pinned: `--force` exists only as a CLI flag (integration.go:334), and design §5 pins the API's force argument to a literal `false` constant. Skip-on-stale matches the repo's bias-to-live doctrine (gitflow-lane-protocol §3): a dead session's hold stays a human decision. Residual (accepted, not a defect): a process death between acquire and release leaves a stale hold — the identical failure mode the manual ceremony has, recoverable by human `--force`.
2. **develop_branch reuse (no new key) — SOUND; no conflict with the window-record `branch` field.** Verified: `resolveIntegrationTarget` (integration.go:121) and the auto path resolve through the same `LoadGitFlowIntegrationConfig` (loader_integration_branch.go:64; consumer integration.go:274), so the record's `branch` and the merge target cannot disagree. The one divergence case — a manual `acquire --branch <other>` hold — results in a held window → auto path skips (AC-WKW-005 covers both holder kinds) — never a cross-branch merge. All four card properties re-verified in-tree: template `git-strategy.yaml.tmpl` carries zero `develop_branch` (grep 0); every loader failure path yields "" (loader_integration_branch.go:14-16 comment + code); the reader exists (t449 note at types.go:119-121); the key IS what `acquire` records. The refinement is an improvement over the card's original scope item; the card-mandate conflict analysis the lead requested finds no conflict.
3. **auto_create explicit-scope route — SOUND, and it corrects the card premise.** The card mandate's "auto_create ... currently zero readers" is contradicted by the tree: `readWorktreeAutoCreate` (internal/cli/worktree_advisory.go) reads it, types.go:624-625 documents the read, and the inventory already classifies it W/evidence:reader (testdata/shipped_key_inventory.yaml:2952-2954). The key was never in the dead class; its defect is the false wording ("is auto-creating a worktree for isolation", worktree_advisory.go true-branch), and the SPEC fixes exactly that (REQ-WKW-011) while preserving the AC-WBG-009 observability regex (regex quote verified verbatim against the worktree_advisory.go doc comment). The rejected alternatives are genuinely worse: OR-gating `enterSessionWorktree`'s creation (session_worktree.go:158, gate `SessionWorktreeEnabled` + env override at internal/config/session_worktree.go:31-44) would let a stale `auto_create: true` re-enable what the user disabled via `workflow.session_worktree.enabled: false`; renaming breaks a live key. The migration story is thin (O2) but the changed behavior is wording-only for a promise the code never honored.

## Cross-SPEC Conflict Check (lead-requested dimension)

- **REQ-SW-008/009/010/022 (SPEC-SESSION-WORKTREE-001)** — no contradiction. `auto_merge` adds an independent third behavior with its own toggle; it does not touch `auto_cleanup`'s role as the single knob for session-exit + PR-merge removal (REQ-SW-022 intact). Clean-exit-only mirrors REQ-SW-009 (verified at session_worktree.go:637); dirty guards mirror REQ-SW-010/024 (session_worktree.go:680 `worktreeIsDirty` reused); the M8 on-touch surface stays cleanup-only, correctly excluding a merge sweep over other sessions' WT-* branches (prMergeCleanup verified at session_worktree_prmerge.go:149). Notice-prefix discipline follows the existing constants (session_worktree.go:141) with a required third distinct prefix.
- **t449/t637 develop_branch semantics** — consistent; the SPEC is a second consumer of the same loader and is stricter on the empty-value case (skip vs the manual caller-fallback warning) — the safe direction (O5).
- **Notice-family** — plan §B correctly flags `internal/github/pr_merger.go` `MergeOptions.AutoMerge` (verified at pr_merger.go:12-13) as an unrelated key domain; the collision assertion the reader test carries (shipped_key_reader_test.go:806-814) exists and M4 will invert its polarity correctly.

## Recommendation

**FAIL — fix the three blocking defects and resubmit; the design decisions stand as settled.** Numbered fixes for manager-spec:

1. (D2, spec.md:63-69) Reword REQ-WKW-002's trigger: "While a session exits with a live session worktree" and "before the disposal step (where enabled)". Do not touch AC-WKW-002.
2. (D1) Add AC-WKW-013 (REQ-WKW-009: prefix-distinctness test judge) and AC-WKW-014 (REQ-WKW-013: four-combination matrix judge) to acceptance.md §B; cite them from spec §D's invariant list.
3. (D3, acceptance.md:96-97) Fix AC-WKW-012's judge: selector → `TestShippedConfigKeysHaveReaders`; grep judge → `! grep -q 'AutoMerge has no production reader' internal/config/types.go`; add the swept-count sentence to acceptance §A.
4. (optional, cheap while editing) Fold O2's DoD line and O5's one-sentence asymmetry note.

The re-audit is scoped to this defect delta (three blocking findings + regression check over them) per the Retry Loop Contract — not a from-scratch re-audit.

---
---

# Iteration 2 — Delta Re-audit (2026-09-12, spec v0.2.0)

Verdict: **PASS** — Overall Score: **1.00** (Tier M threshold 0.80). Score progression 0.75 → 1.00, no regression. Re-audit scope: the enumerated D1-D3 delta + regression over prior-iteration defects; design decisions declared immutable by the lead and verified untouched (design.md mtime unchanged since iteration 1; all three adjudications from iteration 1 carry over unchanged).

## Defect Resolution Check

- **D1 RESOLVED** — AC-WKW-013 (acceptance.md:105-112) cites REQ-WKW-009: prefix distinctness as a constant-level assertion against both `SessionExitCleanupNoticePrefix` and `PRMergeCleanupNoticePrefix`, plus the non-blocking half (forced merge failure leaves the session-exit exit status unaffected), Given-When-Then, judge `go test ./internal/cli/ -run TestAutoMergeNoticePrefixDistinct -count=1`. AC-WKW-014 (acceptance.md:114-121) cites REQ-WKW-013: all four `auto_merge` × `auto_cleanup` combinations, merge iff merge-on / removal iff cleanup-on, judge `TestAutoMergeToggleIndependence`. All 13 REQs now covered; all 14 AC→REQ citations valid; no orphan ACs; spec.md §D updated to "AC-WKW-001..014" naming both new invariants (spec.md:162-167). 14 ≤ Tier M AC ceiling 16.
- **D2 RESOLVED** — REQ-WKW-002 (spec.md:64-69) trigger now reads "While a session exits with a live session worktree" — the disposal-coupled phrasing ("being disposed", "before disposal proceeds") is gone; no reading of the text gates the merge on `auto_cleanup`. Ordering (merge-before-disposal) remains pinned where it belongs: design §4 and plan M2. AC-WKW-002 was already consistent and is untouched.
- **D3 RESOLVED** — AC-WKW-012 judge (acceptance.md:100-103) now `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1` (the real umbrella test, verified to exist at internal/config/shipped_key_reader_test.go:70 in iteration 1 — unchanged file) `&&` `! grep -q "AutoMerge has no production reader" internal/config/types.go` → exit 0 (correct semantics: pattern absent = criterion met = exit 0). The swept-count norm is added to acceptance §A (acceptance.md:6-10): every judging command's swept count recorded verbatim in §E.2, "a PASS claimed without its recorded swept count is an unattributed claim" — this closes the vacuous-empty-sweep shape for ALL 14 judges, including the run-phase-created test selectors.

## Regression Check (iteration-1 defects)

- D1 [RESOLVED]: evidence above.
- D2 [RESOLVED]: evidence above.
- D3 [RESOLVED]: evidence above.

## Regression Sweep (must-pass + dimension re-verification on the delta)

- MP-1 PASS — REQ-WKW-001..013 unchanged and sequential; AC-WKW-001..014 sequential.
- MP-2 PASS — the reworded REQ-WKW-002 is still a GEARS compound (`While ... where ... when ... the CLI shall` — PASS-equivalent per M3); the other 12 REQ bodies verified byte-unchanged from iteration 1. Requirement layer only; ACs remain verification-layer Given-When-Then.
- MP-3 PASS — frontmatter 12/12 intact after the version bump (`"0.2.0"` quoted semver, status draft, `updated: 2026-09-12`); no aliases. HISTORY carries a proper 0.2.0 revision row (spec.md:25) citing this report and scoping the delta.
- MP-4 N/A — unchanged.
- MP-5 PASS — D7 re-run: same 5 referenced SPEC-IDs, all `status: completed`, no new references introduced by the delta.
- MP-6 PASS — D8 re-run: `syscall` count 0/0/0/0/0 across all five artifacts.
- MP-7 PASS — `[NEEDS CLARIFICATION]` grep on plan.md: no match (exit 1).
- Cross-SPEC: no contradiction with REQ-SW-008/009/010/022 or t449/t637 develop_branch semantics (iteration-1 analysis carries over; nothing in the delta touches those surfaces).

## Category Scores (iteration 2)

| Dimension | Score | Change | Evidence |
|-----------|-------|--------|----------|
| Clarity | 1.00 | +0.25 | D2 fixed — REQ-WKW-002 single-interpretation; no other ambiguous REQ found in re-read. |
| Completeness | 1.00 | = | Sections + frontmatter unchanged except the HISTORY row and §D update, both correct. |
| Testability | 1.00 | +0.25 | D3 fixed — all 14 judges executable with correct exit semantics; swept-count norm guards the empty-sweep shape document-wide. |
| Traceability | 1.00 | +0.50 | D1 fixed — 13/13 REQs covered, 14/14 ACs validly cited, no orphans. |

## Residual Notes (non-blocking, carried from iteration 1)

- O1 (RED-now two-cell records) and O2 (auto_create CHANGELOG note) were not adopted — optional per M6, no verdict effect. O2 remains worth doing at sync: the `auto_create: true` advisory wording change is user-visible.
- The three run-phase test selectors named by AC-WKW-013/014 (`TestAutoMergeNoticePrefixDistinct`, `TestAutoMergeToggleIndependence`) do not exist yet by design — they are created in M1/M2; the §A swept-count norm now makes their absence visible rather than silent if a judge were run early.
- progress.md §E.1 honestly records the iteration-1 FAIL; manager-spec (or the orchestrator) should append the iteration-2 PASS entry when persisting the gate verdict.

## Recommendation (PASS)

PASS — rationale: every must-pass criterion holds with the evidence cited above, all three iteration-1 blocking defects are mechanically resolved with no collateral change, the three binding design decisions survived adversarial review in iteration 1 and were untouched in the revision, and the aggregate 1.00 exceeds the Tier M threshold 0.80. Skip-eligibility note for the run gate: verdict PASS + score ≥ 0.80 + artifact-hash snapshot taken at this verdict (spec.md/acceptance.md/progress.md mtime 2026-09-12 19:07, plan.md/design.md unchanged) — any further artifact edit invalidates the cache per ComputeHash.


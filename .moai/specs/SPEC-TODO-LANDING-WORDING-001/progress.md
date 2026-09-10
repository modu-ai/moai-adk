# Progress — SPEC-TODO-LANDING-WORDING-001

Card t486. Plan-phase artifacts authored 2026-09-07 in worktree
`.claude/worktrees/t486` (branch `WT-t472-spec-wording`).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-TODO-LANDING-WORDING-001
tier: S
artifacts:
  spec.md: present (12-field frontmatter, status: draft, 6 REQ GEARS, 6 inline AC, exclusions per OutOfScopeRule)
  plan.md: present (§A-§H, 5 milestones)
  progress.md: present (this skeleton)
evidence_citation_policy: all figures carried from .moai/reports/t482/ (verdict.md + 3 data files); nothing re-measured
target_immutable: SPEC-TODO-LANDING-ATTRIBUTION-001 status stays completed; only its spec.md and plan.md are amended (run phase)
lint: moai spec lint .moai/specs/SPEC-TODO-LANDING-WORDING-001/spec.md → 0 errors (measured at plan-phase commit in this tree)
```

## §E.2 Run-phase Evidence

Run phase executed 2026-09-07 in worktree `.claude/worktrees/t486` (branch `WT-t472-spec-wording`),
base HEAD `e6c91da41`. Five amendments, five commits, all figures CITED from t482's committed
evidence (`.moai/reports/t482/`) — nothing re-measured, re-derived, or hand-enumerated (C-2).

| # | REQ | Commit | What changed (anchor → edit) | Evidence path cited |
|---|-----|--------|------------------------------|---------------------|
| 1 | REQ-TLW-001 | `a37ab5fd1` | target spec.md §A.4.2 `[HARD] The residual is a FLOOR, not a total` block rewritten: "at least 10" → **28 measured**, classification 38 = C 5 + M 28 + A 5, 28 still a floor (5 judgment-deferred unopened; opening raises M to at most 33), history line extended `1 → 19 → 7 → ≥10 → 28 measured`; live-queue clause updated to t482's measurement | `.moai/reports/t482/verdict.md` §4 + `residual-evidence.txt` (§2.5 for queue status) |
| 3 | REQ-TLW-003 | `a37ab5fd1` | same editing pass — non-closure property paragraph appended after the floor block: three grounds (open population t409/t460; [HARD] branch-name rule collision, `(branch WT-t80)` direction; ancestor propagation refuted via `git log d2ad26c90^2 --not d2ad26c90^1`) + under-estimation sentence (18 / 4 accidental / 1 → 19 → 7 → ≥10) + no-reduction-target [HARD] | `.moai/reports/t482/verdict.md` §5.1, §2.6, §2.7; `s1-reconcile.txt` |
| 2 | REQ-TLW-002 | `28df117f3` | S1 paragraph inserted between floor block and property paragraph: release-integrate shape definition + example, BOTH figures as distinct measures (18 subjects / 18 ids, 10 `WT-` / 8 `worktree-`; residual contribution 14, other 4 = `t119 t130 t145 t146`), 3-version/3-iteration anonymity, operator rejection with ground (branch-name attribution; positional exception) — **naming only, no form added** | `.moai/reports/t482/s1-reconcile.txt`; verdict.md §4.1 + §5.1 cause 2 + §9 판정 1 (plan-audit F2) |
| 4 | REQ-TLW-004 | `63b7fbf70` | target spec.md §D `### Out of Scope — axis C, landing evidence storage` gains the termination bullet: enumeration is a transitional instrument, t359 the sole termination path, further rounds not the plan | `.moai/reports/t482/verdict.md` §5.2(1) + §6 recommendation 4 |
| 5 | REQ-TLW-005 | `84498f8c0` | target spec.md §A.4: form-table row 3b `Observed` 76 → 77; single-token paragraph's count 76 → 77; new casing-latitude paragraph (`^[Mm]erge` 77 vs `^Merge` 76, delta subject `merge: WT-ci-test-observability into develop (t358)`) with residual-invariant note (347/309/38 + 28-floor unmoved). §A.7's "76" deliberately untouched — it quotes the iter-2 audit's own recorded inference | `.moai/reports/t482/form3b-delta.txt`; verdict.md §2.4 |
| 1t | REQ-TLW-001 tail (plan-audit F1) | `2e50ee34a` | target plan.md §D residual risk row: BOTH figure occurrences replaced — the literal (`→ **at least 10** at v0.4.0` → `→ ≥10 at v0.4.0 → **28 measured** …`) and the bare figure (`10 is a floor, not a total` → 28-floor with the at-most-33 bound); adjacent live-queue "unmeasured" clause corrected to t482's measurement for cross-file consistency | `.moai/reports/t482/verdict.md` §4, §2.5 |

Delivery vehicle per spec.md §A.3 (plan-audit F3): amendments land as t486's own commits; the
target's frontmatter was NOT touched (status `completed`, version `0.4.0`, `updated: 2026-09-06`,
no `amendment_of:`, no HISTORY row, no version bump). New SPEC frontmatter transitioned
`draft → in-progress` on the first run commit (`a37ab5fd1`).

§E self-verification (plan §E), verbatim outputs in `.moai/reports/t486/run-evidence.md`:

- Load-bearing tokens present: `28 measured` (spec.md ×2, plan.md ×1), `S1` ×5,
  `residual contribution` ×1, `casing` ×2, `termination` ×2, `t359` ×3, form-table `| 77 |` ×1.
- `grep "at least 10"` in target: 2 hits, both quoted history (HISTORY 0.4.0 row; §A.4.2's own
  quotation of the replaced figure) — 0 live claims. Bare forms (`10 is a floor` / `floor of 10`):
  0 hits.
- `git diff --name-only e6c91da41..HEAD` → exactly the two target files + this SPEC's spec.md
  (AC-TLW-006 parenthetical).
- Target frontmatter after amendment: `status: completed` unchanged.
- Residual invariant: the 347/309/38 measurement block (§A.4.2 lines 309-311) untouched by the
  diff; only changed lines naming those figures are the new invariant note itself and the
  classification statement.
- No reduction-target framing (`reduce/shrink the residual`): 0 hits.
- `moai spec lint .moai/specs/SPEC-TODO-LANDING-WORDING-001/spec.md` → `✓ No findings — all SPEC
  documents are valid` (rc=0).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
spec_id: SPEC-TODO-LANDING-WORDING-001
run_complete_at: 2026-09-07
evidence_head: 2e50ee34a
commits:
  - a37ab5fd1  # M1: 28-floor + non-closure property (REQ-TLW-001 + REQ-TLW-003) + draft->in-progress
  - 28df117f3  # M2: S1 naming with rejection recorded (REQ-TLW-002)
  - 63b7fbf70  # M3: t359 sole termination path (REQ-TLW-004)
  - 84498f8c0  # M4: form 3b 76->77 + casing latitude (REQ-TLW-005)
  - 2e50ee34a  # M5: plan.md risk row both figure occurrences (REQ-TLW-001 tail, F1)
target_status_unchanged: true   # SPEC-TODO-LANDING-ATTRIBUTION-001 status: completed, version 0.4.0
evidence_export: .moai/reports/t486/run-evidence.md   # uncommitted per card convention
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: completed
spec_id: SPEC-TODO-LANDING-WORDING-001
sync_complete_at: 2026-09-07
sync_commit_sha: 82c55077c
sync_subject: docs(SPEC-TODO-LANDING-WORDING-001): sync-phase — 3-phase close, completed (t486)
frontmatter_status_transitions:
  - in-progress -> implemented -> completed   # full transition rides the single sync commit
changelog_entry: none   # SPEC-documents-only card; no shipped artifact, no user-facing surface
b12_self_test_a:
  pre_emission_grep: not-applicable   # no CHANGELOG entry emitted (B12 halted at skip-decision)
b12_self_test_c:
  file_path_verification: not-applicable   # no file paths claimed in a CHANGELOG entry
```


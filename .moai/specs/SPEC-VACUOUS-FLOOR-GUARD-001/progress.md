# Progress — SPEC-VACUOUS-FLOOR-GUARD-001 (card t378)

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **S** (one file, one function; 7 REQ / 8 AC, both inside the 8/8 Tier S ceiling).
- Artifact set: `spec.md` + `plan.md` (AC inline in `spec.md` §C) + this `progress.md`.
- Base measured against: `3f03d9c36` (= `origin/develop`), branch `WT-vacuous-floor-guard`.
- Repair direction: **deletion** of the unreachable `budget < floor` branch, argued in
  `spec.md` §A.4 from a REQ-BLB-002 coverage analysis rather than assumed. M1 of `plan.md` is a
  stop-and-blocker gate that re-verifies that argument against the tree before any edit.
- Plan-phase produced artifacts only: no implementation code, no push, no PR, no CI, no load.

## §E.2 Run-phase Evidence

Full evidence record: `.moai/reports/t378/run-evidence.md`, with companions
`repair-direction.md` (M1), `census.md` (M2), `mutants.md` (M3), `negative-evidence.md` (M4).
Base `3f03d9c36`; measurements taken at HEAD `226bdd0dc` on branch `WT-vacuous-floor-guard`.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-VFG-001 | PASS | `grep -n 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go`; `grep -c 'floor :=' …`; `go vet ./internal/kanban/...` | 1 match (`122:	if boardLockWaitBudget < floor {`, t372's guard), baseline 2; `floor :=` count `1` at line 120, baseline 2; vet exit 0, no output |
| AC-VFG-002 | PASS | `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/` | `=== RUN` ×1 (non-zero selector), `--- PASS`, `ok …0.452s`; four assertions at lines 28 / 65 / 74 / 79 |
| AC-VFG-003 | PASS | M1 (cost 33ms→20ms) → `go test -timeout 600s -count=1 ./internal/kanban/` → revert → re-run | RED, sole failure: `board_lock_wait_test.go:55: per-mutation cost 20ms is below the CI-class observation of 33ms`; post-revert `ok …15.975s` |
| AC-VFG-004 | PASS | M2 (headroom 5→1) → whole-package run → revert → re-run | RED: `board_lock_wait_test.go:60: headroom factor 1 states no headroom` (+3 further REDs, each named and attributed in `mutants.md`); post-revert `ok …16.074s` |
| AC-VFG-005 | PASS | M3 (writers 10→8) → whole-package run → revert → re-run | RED: `board_lock_wait_test.go:46: supported writers = 8, want 10 (Factory mode's ten lanes against one queue)` (+t372's guard, predicted); post-revert `ok …16.076s` |
| AC-VFG-006 | PASS | M4 form A `1650ms` then form B `1400ms` → revert | Form A GREEN (both guards `--- PASS` — the stated gap, observed); form B RED: `board_lock_wait_test.go:29: budget 1.4s is not the product of its named inputs (…)`; post-revert `ok …15.995s` |
| AC-VFG-007 | PASS | Branch present (pre-edit) + M2 planted → scoped `-v` run | `headroom factor 1 states no headroom` PRESENT; `< headroom floor` ABSENT from the complete output, at a 330ms budget vs the 660ms composed floor |
| AC-VFG-008 | PASS (qualified) | `git diff --stat`; `git diff` per file; `grep -rn 'go test' .moai/reports/t378/`; `./bin/moai spec lint --strict` | 1 file changed, 27 insertions(+), 7 deletions(-); `board_store.go` diff empty; 12/12 recorded invocations package-scoped and serial; `0 error(s), 1096 warning(s)` |

**Invariants**

| Invariant | Status | Evidence |
|---|---|---|
| t372's `TestBoardLockWaitBudgetCoversSerializedMutations` untouched (AC-SIV-013 window open) | HELD | Single diff hunk spans original lines 32-44 inside `TestBoardLockWaitBudgetDerivedFromNamedInputs`; t372's guard begins at original line 95 and appears in no hunk |
| No constant retuned (REQ-VFG-006) | HELD | `git diff -- internal/kanban/board_store.go` → empty; every mutant reverted and confirmed clean before the next was planted |
| No production behaviour changed | HELD | Only `board_lock_wait_test.go` in `git diff --stat` |
| No new same-terms inequality (REQ-VFG-003) | HELD | Removed block replaced by comment lines only; `boardLockWaitBudget <` count 1, in t372's guard |
| Verification load (REQ-VFG-007) | HELD | 12 serial invocations, all `./internal/kanban/`-scoped; zero backgrounded (`grep -rn 'go test.*&\s*$'` → no output); no full-suite, no race-detector invocation |

**AC-VFG-008 qualification, stated rather than glossed.** The criterion asks for "zero occurrences"
of two tokens in the `go test` grep. As a literal token count that is not satisfied: both appear on
`plan-audit.md` lines 71-72, inside the plan-auditor's own prose *stating the prohibition* and
proposing this criterion's wording. Verified precisely — `grep -rn -- '-race' .moai/reports/t378/`
→ 2 matches, both `plan-audit.md:71,72`; `grep -rn 'go test \./\.\.\.' .moai/reports/t378/` → the
same 2 lines. No recorded invocation carries either. The substantive requirement holds without
exception; the literal phrasing is defeated by a document describing what is forbidden.

**Gaps (not observed):** cross-platform build (not run — one `_test.go` file, no build tags);
full-suite verdict (CI-owned per REQ-VFG-007); coverage percentage (no new executable line);
`golangci-lint` (`gofmt -l` and `go vet` were run); every-assignment unreachability (static
argument only — AC-VFG-007 corroborates on one assignment); repository-wide sweep for sibling
vacuous guards (SPEC-excluded). A deletion carries no positive mutant evidence of its own.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: c3208b08f
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 1   # TestBoardLockWaitBudgetCoversSerializedMutations, byte-identical
l44_pre_commit_fetch: not-applicable   # no push in this card; lead holds the integration window
l44_post_push_fetch: not-applicable    # no push in this card
new_warnings_or_lints_introduced: 0    # gofmt -l empty; go vet exit 0; spec lint 0 error(s)
cross_platform_build:
  darwin: not-run
  linux: not-run
  windows: not-run
  note: >-
    Single _test.go file, no build tags, no platform-conditional code. Not run, and
    recorded as a gap rather than claimed.
total_run_phase_files: 1               # internal/kanban/board_lock_wait_test.go
m1_to_mN_commit_strategy: single-commit   # Tier S: one file, one hunk, plus evidence records
```

## §E.4 Sync-phase Audit-Ready Signal

Full sync verdict, in five evidence-bearing sections with per-measurement party attribution
(`[MD]` manager-develop / `[ORCH]` orchestrator / `[DOCS]` manager-docs):
`.moai/reports/t378/verdict.md`.

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: 7c555c220   # sync-phase commit (t378); merge e79272713 is its integration point
sync_status: complete
b12_self_test_a: pass    # grep -c 'SPEC-VACUOUS-FLOOR-GUARD-001' CHANGELOG.md -> 0, no duplicate
b12_self_test_b: pass    # 8 own AC (AC-VFG-001..008); the domain-agnostic pattern returns 9
                         # because spec.md cites t372's AC-SIV-013 as a cross-reference. Non-zero
                         # and plausible, so the test is satisfied rather than vacuous.
b12_self_test_c: pass    # every path named in the entry resolves at HEAD (verdict.md §2.3)
changelog_entry_position: "[Unreleased] -> ### Changed, first entry (section is newest-first)"
frontmatter_status_transitions:
  spec_md: in-progress -> implemented
  plan_md: not-applicable        # Tier S: plan.md carries no status frontmatter
  acceptance_md: not-applicable  # Tier S: AC inline in spec.md §C, no acceptance.md
  progress_md: not-applicable    # no status frontmatter
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking policy
docs_surfaces_checked:
  readme_4_locale: no-change-warranted
  docs_site: no-change-warranted
  template_mirror: no-change-warranted
  evidence: >-
    grep -rln 'boardLockWaitBudget|boardLockHeadroom|BudgetDerivedFromNamedInputs|
    VACUOUS-FLOOR-GUARD' README.md README.*.md docs-site/ -> rc=1 (no hits);
    git diff --name-only 3f03d9c36..HEAD | grep -c 'internal/template/templates/' -> 0.
    Checked, not assumed.
sync_phase_reverification:
  guards_green: "go test -timeout 600s -count=1 -v -run 'TestBoardLockWaitBudget'
    ./internal/kanban/ -> 2 === RUN, both --- PASS, ok 0.393s [DOCS] at c3208b08f"
  lint: "./bin/moai spec lint -> 0 error(s), 1096 warning(s); tree binary,
    strings bin/moai | grep -c c3208b08f -> 4 [DOCS]"
  scope_hold: "git diff --stat 3f03d9c36..HEAD -- internal/ -> 1 file, +27/-7;
    git diff -- internal/kanban/board_store.go -> empty [DOCS]"
coordinate_reverification:
  groups_checked: 19        # 95 raw prose hits, 3 of them the 'baseline 2' false positive
  drifted: 1                # spec.md §A.3 table Line column (45/54/59/28 -> 65/74/79/28 at HEAD)
  corrected: 0              # the drifted group is spec.md §A body, outside manager-docs' boundary
  blocker: B-1              # see verdict.md §6 — re-delegate to manager-spec if the fix is wanted
status_closes_at: implemented   # NOT completed: branch unpushed, no CI verdict on any commit
```

**Non-claims carried forward (verdict.md §5), restated here so they survive a reader who reads
only this file:** a deletion carries no positive mutant evidence of its own; AC-VFG-007's negative
evidence is one observation on one assignment and corroborates the static argument rather than
proving it; "660ms is looser than 1650ms" is an illusion the vacuous branch created, since 1650ms
was the same expression as the budget and never functioned as a bound; AC-VFG-008 is **not** a
clean zero (both prohibited tokens appear on `plan-audit.md` lines 71-72, inside prose stating the
prohibition — all 12 recorded invocations comply); AC-VFG-006's form-A hole is documented, not
closed; and one census prediction (`TestConcurrencyStress` under M2) missed, recorded because a
miss must stay visible.

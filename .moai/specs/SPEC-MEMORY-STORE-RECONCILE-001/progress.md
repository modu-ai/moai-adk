# SPEC-MEMORY-STORE-RECONCILE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

**Iteration 3 (post plan-audit FAIL 0.78, up from 0.72).** N1-N8 addressed; optional N9-N11, N13
applied; N12 (AC ceiling) resolved without a merge. Per-defect mapping in the closing report.

- **Tier: M**, re-checked rather than inherited. Non-artifact files touched: **8** — four sources
  (`moai-memory.md`, `moai-constitution.md`, `moai.md`, `token_budget_guard.go`), **three** mirrors
  (`internal/config/` has none), and one test file. Inside Tier M's 5-15 band. The earlier
  "four sources plus four mirrors" phrasing was wrong and is corrected in plan.md §B.4, which
  carries the enumerated table this line must stay consistent with.
- Requirements: **15** (ceiling 16). Acceptance criteria: **16** (ceiling 16, zero headroom).
  N2 / N3 / N5 / N6 / N11 were absorbed as **strengthenings of existing criteria**, so no new slot
  was needed and no merge was forced. Any further criterion now requires a scope split or a tier
  change — stated so the ceiling is not quietly exceeded later.
- Evidence base: `.moai/reports/t383/measurements.md` §M5a (revision 2 — records **both** wrong
  relations) and §M5b, plus `.moai/reports/t383/measure-n1.sh`, an independent re-measurement run
  in this worktree at iteration 3. No figure in the SPEC is re-derived; each is cited, expressed in
  R4 re-measurement form, or explicitly withdrawn as unattributed.
- **A live-mutation observation that decides which figures may be pinned.** The index moved three
  times on 2026-08-31 — 26,280 → 26,290 → **26,577** bytes, 123 → **124** entry lines, 189 → **190**
  unique targets — with nobody here touching it. Across the same interval the defect figures held
  exactly (58 / 44 / 14 / 40). Size metrics are dated references; defect metrics may be asserted.
- SPEC ID regex check executed (Bash), verbatim output: `PASS`.
- Open decisions: none. G5 (the copy set) is decided in spec.md §A.2.2 — **all 58**.
- No `[NEEDS CLARIFICATION]` markers.
- **Two gaps recorded, not closed** (plan.md §I): G2 — `moai spec lint` is unmeasured for this SPEC
  at plan time (installed binary 20 commits behind for `internal/spec/`; the PATH run's 0 findings
  is recorded as weak evidence from a stale build, and AC-MSR-016 requires `./bin/moai` by path).
  G4 — doctrine-vs-code divergence on store derivation, now enumerated across **three** surfaces
  including `moai.md:165`, the line M1 itself edits.

## §E.2 Run-phase Evidence

Full evidence with verbatim output: `.moai/reports/t383/verdict.md`. Store readings:
`reconcile-before.json` / `reconcile-after.json`. Sampling gate: `m0-sample.md`.

`$D` = the active store resolved by `moai memory doctor` (path in `preflight.md`).

### AC matrix — 16 of 16 PASS

| AC | Status | Actual output |
|---|---|---|
| AC-MSR-001 | PASS | `0` / `0` / `0` (baselines at HEAD: `3` / `1` / `2`) |
| AC-MSR-002 | PASS | cond.1 `4` / `4` in local+mirror; cond.2 no numeric cap value in either copy |
| AC-MSR-003 | PASS | `131:#### Compressing the index means making entries shorter — never fewer` |
| AC-MSR-004 | PASS | 4 matches; § "Two stores, and only one of them is loaded" carries the `--dir` + `exists` rule |
| AC-MSR-005 | PASS | `head -1` → `# MoAI Constitution` (always-loaded); clause offers no drop branch |
| AC-MSR-006 | PASS | first grep exit `1`; `MEMORY.md` only at lines 109/114/127, all inside the TOMBSTONE |
| AC-MSR-007 | PASS | mutation → `--- FAIL: TestFixedSlotsExistInRepoTree`; revert → `--- PASS` |
| AC-MSR-008 | PASS | `18d17 < …/MEMORY.md`; same-tree A/B total `76009` both ways → slot contributed `0` |
| AC-MSR-009 | PASS | gate exit `0`; BEFORE `58`; AFTER `0` |
| AC-MSR-010 | PASS | entry `139 → 139`; targets `205 → 205`; bytes `31366 → 31366` |
| AC-MSR-011 | PASS | legacy `1098 → 1098`; recent-mtime `0 → 0`; `skipped_exists: 0` |
| AC-MSR-012 | PASS | `git status --short` shows no path under either store |
| AC-MSR-013 | PASS | `rc1=0 rc2=0 rc3=0` |
| AC-MSR-014 | PASS | existence gate ok ×3; scan `exit=1` exactly; planted red `exit=0` |
| AC-MSR-015 | PASS | `m0-sample.md`: 0 of 12 superseded vs threshold ≥4 → PROCEED |
| AC-MSR-016 | PASS | anchor `0`; `lint_exit=0`; liveness `0`/`0`; `0 error(s), 1096 warning(s)`; count `0` |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| No store file is a commit target (REQ-MSR-015) | HOLDS | `git status --short` — no store path, staged per-file by explicit pathspec |
| Copy-only, never-overwrite (REQ-MSR-010) | HOLDS | `cp -n`; log `copied 58 / skipped_exists 0 / no_source 0`; legacy count and mtimes unmoved |
| Neither reachability metric decreased (REQ-MSR-011) | HOLDS | `139 → 139`, `205 → 205` |
| Index not edited by this card (REQ-MSR-012/013) | HOLDS | bytes `31366 → 31366` |
| Mirror parity (REQ-MSR-014) | HOLDS | three `diff -q` rc=0 after `make build` |
| Defect figures stable across 5 readings | HOLDS | dangling `58` at every pre-M3 reading while every size figure moved |

### Debt items discharged

| Debt | How |
|---|---|
| 1 — stale `123` denominator | Removed from spec.md:92 (the [HARD] sentence), :102, :306, :312 **before** M1 wrote doctrine; cascade-corrected REQ-MSR-011 and AC-MSR-010's stale `124/135/190`. Verified: **zero** stale figures reached any of the six doctrine files. Vindicated in-flight — the denominator moved 123 → 139 during this card |
| 2 — AC-MSR-016 misfires | Anchor + liveness assertion added; both misfires reproduced first (`6` from the SPEC dir; `0` from an empty file). The first anchor form was itself un-runnable and was corrected (see below) |
| 3 — prefix pathspec sweeping runtime state | AC-MSR-012 narrowed to the four named artifacts; staged per-file; `git status --short` re-read in the staging call |
| 4 — commit `measure-n1.sh` | Committed with `derive-missing.sh` and `reconcile.sh`; two caveats recorded (needs bash, word-splits on filenames) |
| 5 — re-run what you rewrote | Rule written into plan.md §G with the base-rate table, now **five for five** including this round's two |
| 6 / G8 — orphan count | Recorded before **and** after: `46 → 46`, unmoved as predicted but now measured rather than assumed |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-01
run_commit_sha: <backfill — commits created at close of this run-phase>
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed (origin/develop 26 ahead; NOT absorbed per lead instruction)
l44_post_push_fetch: n/a — push is out of scope for this run-phase
new_warnings_or_lints_introduced: 0
  spec_lint: "0 error(s), 1096 warning(s)" via ./bin/moai by path; 0 findings name this SPEC
  go_vet: clean
  gofmt: clean
cross_platform_build:
  darwin_arm64: pass (make build)
  other: not attempted locally — CI owns the matrix
total_run_phase_files: 8 tracked modified + 2 untracked trees (SPEC artifacts, evidence reports)
m1_to_mN_commit_strategy: three commits (doctrine+mirrors / guard removal / SPEC+evidence)
budget_change:
  always_loaded_token_budget: 76000 -> 76210
  reason: REQ-MSR-004 clause must be always-loaded (C5); headroom was 201 tokens
  measured_before: 75799
  measured_after: 76009
  note: raised by exactly this card's 210-token addition so prior headroom is preserved, not inflated
gaps_open: [G4, G6, G7]
gaps_closed: [G2]
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

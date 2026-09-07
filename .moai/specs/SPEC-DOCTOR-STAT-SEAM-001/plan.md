# Plan — SPEC-DOCTOR-STAT-SEAM-001

Card t563 · worktree `.claude/worktrees/t563` · branch `WT-doctor-stat-shim` · base `ef10a2524` · Tier S · Class C

## §A Context

`internal/cli/doctor_codex.go` bypasses the package's stat seam (`osStatFn`,
`update_preserve_inventory.go:59`) at exactly two sites — `:459` (mirror-entry symlink follow,
site A) and `:857` (`codexStaleSkillFinding`, site B) — while every sibling consumer in the same
package routes through the seam. The defect is a testability gap, not a behavior defect: tests
cannot observe which paths were stat'ed or inject stat failures portably. `codexStaleSkillFinding`
has zero direct tests today. This SPEC routes both sites through the seam with ZERO behavior
change, adding characterization tests first and observation tests after.

Development mode: this is a behavior-preserving transformation — the ANALYZE-PRESERVE-IMPROVE
(DDD) cycle fits; PRESERVE carries the weight (characterization + output-identity proof), IMPROVE
is the two-line seam swap.

## §B Known Issues

- `codexStaleSkillFinding` has zero name-invoked (unit-level) tests (verified: defined `:815`,
  called `:319`, no `*_test.go` calls it by name) — but every observable bucket is ALREADY pinned
  indirectly through `checkCodexWiring` by real-fixture tests: `TestCodexSkillPath_{HomeRelative
  ExistingNotMissing:896, HomeRelativeMissingStillCounted:917, RelativeNotMissingDistinct
  Classification:942, RelativeOnlyNonDestructiveFinding:970, BackslashAndOtherUserHomeNotMissing
  :993, AbsoluteExistingAndMissing:1020, RealMissingWithSymlinkLoopIndeterminate:1051,
  ExpansionUsesUserHomeSeam:1090}` and the `TestCheckCodexWiring_{StaleHomeSkillsReported:247,
  EmptyPathEntryNotCountedMissing:311, UnspecifiedEnabledReportedSeparately:363,
  NonBooleanEnabledCountedSeparatelyInStaleSplit:399, IndeterminateStatNotMissing:498,
  DirectoryPathNotMissing:535}` family, plus `TestCheckCodexWiring_Mirror*` (:1259-1569) for site
  A. What is genuinely unpinned is exactly what the seam adds: (a) stat-ARGUMENT observability —
  no test sees which path was stat'ed — and (b) portable stat-failure injection. M1 therefore adds
  STRUCT-level direct pinning (finer-grained than the existing rendered-Message/Detail assertions,
  which can coincide while the `codexFinding` struct differs); it does NOT duplicate the existing
  bucket fixtures — M2 reuses them as the pre/post identity comparison set.
- The seam is a package-level variable with an `@MX:WARN` (`update_preserve_inventory.go:53`):
  reassigning tests MUST NOT call `t.Parallel()` — a concurrent reassignment is a data race.
- The mirror-check half of site A has existing indirect coverage only (via `doctor_codex_test.go`
  fixtures); no test asserts the stat path argument today.

## §C Pre-flight

1. Confirm worktree LINEAGE before first commit: `git merge-base --is-ancestor ef10a2524 HEAD`
   (exit 0 required) and `git branch --show-current` → `WT-doctor-stat-shim`. The recorded base is
   an ancestor assertion, NOT a literal HEAD equality — HEAD has already advanced past the base
   (plan-phase artifacts landed; facts re-confirmed at `45b590c24`; origin/develop tip `9dddac882`).
   Re-read HEAD immediately before every commit per AGENTS.md §2.
2. Confirm measured facts still hold on THIS tree: `grep -n 'os\.Stat(' internal/cli/doctor_codex.go`
   → exactly `:459` and `:857`; `grep -cn 'osStatFn' internal/cli/doctor_codex.go` → 0;
   `grep -n 'os\.Lstat(' internal/cli/doctor_codex.go` → `:441` (must survive the swap).
3. Confirm scoped baseline: `go test ./internal/cli/ -run 'Codex' -count=1 -timeout 1800s` passes
   before any change (baseline attribution for the output-identity proof).

## §D Constraints

- C1 — OBSERVABILITY CHANGE ONLY: zero diagnostic-output delta (REQ-004). If any output changes,
  the change is wrong by definition.
- C2 — COMMIT ORDERING: the characterization-test commit precedes the seam-swap commit. The commit
  graph is the only sequencing witness (ordering-attribution doctrine). Both commits carry `t563`
  in the message.
- C3 — Verification discipline (REQ-007): scoped runs with `-timeout 1800s`; NO local full suite
  (10+ lanes); CI owns the full verdict. Affected-package lint: `go vet ./internal/cli/...` +
  `golangci-lint run internal/cli/...`.
- C4 — Seam-override tests: no `t.Parallel()`; save → replace → `t.Cleanup` restore, exactly the
  `codex_skills_prune_test.go:55-57` pattern.
- C5 — Repo Go conventions: `fmt.Errorf("op: %w", err)` wrapping; English comments/godoc;
  table-driven tests preferred; `t.TempDir()` everywhere; no `t.Setenv` of OTEL vars.
- C6 — `os.Lstat` (`:441`) stays direct; no second seam primitive (spec.md §3.2).

## §E Self-Verification

Run-phase evidence lands in `progress.md` §E.2/§E.3. The verification batch re-runs the AC-SEAM-001
grep assertions, the scoped test commands, and the ordering witness (`git log`) with verbatim
output persisted under `.moai/state/verify/` or `.moai/reports/t563/`.

## §F Milestones

Priority order (no time estimates). M1 first — it is the change-likelihood-heavy half (new test
code pinning unpinned behavior); M2 is the mechanical two-line swap it de-risks.

### M1 — Unit-level characterization tests (priority: High, commit 1)

- Add DIRECT unit tests for `codexStaleSkillFinding` (new file, e.g.
  `doctor_codex_stale_skill_test.go`), names pinned to the `TestCodexStaleSkillFinding_*`
  convention: real fixtures in `t.TempDir()` asserting the classification buckets at the STRUCT
  level (the `codexFinding` fields) — resolves (file AND directory, since directories resolve by
  design), missing (per-`Enabled`-state arms), home-relative expanded, unresolvable-home →
  indeterminate, relative → `relativeCount` (never stat'ed), oddly-formed → `oddlyFormed` (never
  stat'ed), empty path skipped. This is struct-level direct pinning, FINER-grained than the
  existing indirect `checkCodexWiring` suite's rendered Message/Detail assertions (rendered
  strings can coincide while the `codexFinding` struct differs). It SUPPLEMENTS the existing
  `TestCodexSkillPath_*` / `TestCheckCodexWiring_*` fixtures — it does not duplicate them; the
  existing suite (already covered by the pre-flight `-run 'Codex'` baseline) is the M2
  identity-proof comparison set (AC-SEAM-005), which strengthens it.
- The unresolvable-home → indeterminate arm is driven by overriding the existing `userHomeDirFn`
  seam (`stubCodexHome`, `doctor_codex_test.go:103`; serial per C4) — real fixtures alone cannot
  reach it, because a failed home makes `codexUserSkillConfig` skip the entry loop entirely
  (`ok=false` before any stat). The plan's no-injection property is scoped to the STAT seam only;
  the home seam is fair game in M1.
- Add `TestInspectSkillMirror_*` characterization for any mirror-check state the M2
  output-identity proof needs (resolve / dangling / indeterminate via `inspectSkillMirror`).
- Tests must PASS on the unmodified tree — that is their definition. Commit message carries `t563`.
- Exit: `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_|TestInspectSkillMirror_'
  -count=1 -timeout 1800s` green AND the evidence records the swept-test COUNT with the test names
  (> 0) — an empty-sweep green (`ok ... 0.00s`, 0 tests matched) is NOT exit success.

### M2 — Seam swap + observation tests + identity proof (priority: High, commit 2)

- Swap both sites in ONE commit: `:459` and `:857` call `osStatFn` instead of `os.Stat`. Diff
  touches ONLY the two call-site lines (plus no import change — `os` remains in use for `Lstat`
  and types).
- Add per-site observation tests (REQ-006): a recording `osStatFn` override (serial, save →
  replace → `t.Cleanup` restore) asserting the recorded argument per site (mirror-relative joined
  path for A; classified absolute/expanded path for B) and the bucket each injected result drives
  (`ErrNotExist` → dangling/missing; `nil` → resolves; other error → indeterminate).
- Output-identity proof (REQ-004): re-run the M1 characterization suite plus the pre-flight scoped
  baseline selection on the swapped tree — identical finding results on the shared fixtures.
- Exit: `grep -n 'os\.Stat(' internal/cli/doctor_codex.go` → no matches; `os\.Lstat` still at
  `:441`; scoped runs green.

### M3 — Evidence + sync handoff (priority: Medium)

- Persist verbatim verification output (grep assertions, scoped test runs, `git log` ordering
  witness) to `.moai/reports/t563/` and `progress.md` §E.2/§E.3.
- Update frontmatter per transition ownership (`draft → in-progress` owned by manager-develop on
  the M1 commit). Sync phase (manager-docs) closes the SPEC.

## §G Anti-Patterns

- Do NOT "improve" the bucketing, the `classifyCodexSkillPath` switch, or any diagnostic string
  while in the file (scope discipline; REQ-004 makes any such change a defect).
- Do NOT add `t.Parallel()` to any test that reassigns `osStatFn` (data race — `@MX:WARN`).
- Do NOT run `go test ./...` locally (lane load incident 2026-08-15; C3).
- Do NOT introduce an `osLstatFn` (spec.md §3.2) or touch `:441`.
- Do NOT write the characterization tests as RED-first TDD — they pin current behavior and must
  pass BEFORE the swap exists (REQ-005).
- Do NOT claim the t540 gap closed anywhere in commit messages, code comments, or docs
  (spec.md §3.3).

## §H Cross-References

- spec.md / acceptance.md: `.moai/specs/SPEC-DOCTOR-STAT-SEAM-001/`
- Seam doc comment + consumers: `internal/cli/update_preserve_inventory.go:44-59,437-439`,
  `internal/cli/codex_skills_prune.go:96`, `internal/cli/codex_skills_disable.go:149,168`
- Override pattern: `internal/cli/codex_skills_prune_test.go:55-57`,
  `internal/cli/update_preserve_partial_test.go:84-86`
- Motivation only: t540 spec.md §G gap 5

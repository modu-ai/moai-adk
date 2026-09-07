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

- `codexStaleSkillFinding` has ZERO direct tests (verified: defined `:815`, called `:319`, no test
  references in `internal/cli/*_test.go`). Everything the swap changes in site B is currently
  unpinned.
- The seam is a package-level variable with an `@MX:WARN` (`update_preserve_inventory.go:53`):
  reassigning tests MUST NOT call `t.Parallel()` — a concurrent reassignment is a data race.
- The mirror-check half of site A has existing indirect coverage only (via `doctor_codex_test.go`
  fixtures); no test asserts the stat path argument today.

## §C Pre-flight

1. Confirm worktree state before first commit: `git rev-parse --short HEAD` → base `ef10a2524`
   (re-read immediately before every commit per AGENTS.md §2).
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

### M1 — Characterization tests (priority: High, commit 1)

- Add `codexStaleSkillFinding` characterization tests (new file, e.g.
  `doctor_codex_stale_skill_test.go`): real fixtures in `t.TempDir()` covering every bucket —
  resolves (file AND directory, since directories resolve by design), missing (per-`Enabled`-state
  arms), home-relative expanded, unresolvable-home → indeterminate, relative → `relativeCount`
  (never stat'ed), oddly-formed → `oddlyFormed` (never stat'ed), empty path skipped. Assertions
  come from REAL fixture states — no seam injection exists yet (that is the point of M1).
- Add any missing mirror-check characterization the M2 output-identity proof needs (fixture states
  driving resolve / dangling / indeterminate at site A, asserted through `inspectSkillMirror`'s
  current results).
- Tests must PASS on the unmodified tree — that is their definition. Commit message carries `t563`.
- Exit: scoped run green (`go test ./internal/cli/ -run 'StaleSkill|SkillMirror' -count=1 -timeout 1800s`).

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

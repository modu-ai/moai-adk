# Progress — SPEC-DOCTOR-STAT-SEAM-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t563 · worktree `.claude/worktrees/t563` · branch `WT-doctor-stat-shim` · base `ef10a2524`
- Tier: S · Class C (scope decision: both stat sites, settled by card t563 — encoded in spec.md §3.1)
- Artifact set: spec.md + plan.md + acceptance.md (+ this progress.md); `tier: S` in frontmatter;
  acceptance.md included per the orchestrator's explicit deliverable list
- Requirements: 7 (REQ-001..REQ-007, GEARS) · Acceptance criteria: 8 (AC-SEAM-001..008)
- SPEC ID regex check executed as Bash: `PASS` (verbatim output cited in the plan-phase session)
- ID uniqueness: no `SPEC-DOCTOR-STAT-SEAM-001` in `.moai/specs/` (checked via directory listing)
- Measured facts re-confirmed on this tree: `os.Stat` at `doctor_codex.go:459,857`; `os.Lstat` at
  `:441`; `osStatFn` 0 occurrences in `doctor_codex.go`; seam at `update_preserve_inventory.go:59`
- Status: `draft` — awaiting plan audit and Implementation Kickoff Approval

### Repair iteration 1 (spec.md 0.2.0 · plan.md revised)

Resolves the four blocking findings of `.moai/reports/t563/plan-audit.md` (PASS-WITH-CONDITIONS
0.94): D1 — site-B coverage narrative corrected in spec.md §1 and plan.md §B (indirect
`TestCodexSkillPath_*` / `TestCheckCodexWiring_*` suite acknowledged; the real gap re-scoped to
stat-argument observability + portable stat-failure injection; M1 reframed as struct-level direct
pinning, folding existing fixtures into the M2 identity baseline); D2 — plan.md §C.1 preflight
rewritten as an ancestry assertion (`merge-base --is-ancestor`) after HEAD advanced past the
recorded base; D3 — REQ-005 / M1 no-injection claim scoped to the STAT seam (the `userHomeDirFn`
home seam may be overridden to reach the unresolvable-home arm); D4 — M1 exit selector re-pinned
to the named `TestCodexStaleSkillFinding_|TestInspectSkillMirror_` convention with a required
swept-test count. D5 (optional) — t540 pending-ID note added to spec.md §5. D6 — recorded, no
action.

## §E.2 Run-phase Evidence

### M1 — characterization tests (commit `08361d0ee`, this tree @ base `c07fb0dac`)

- Deliverable: `internal/cli/doctor_codex_stale_skill_test.go` — 20 struct-level characterization
  tests pinning CURRENT behavior on the UNMODIFIED tree (12 `TestCodexStaleSkillFinding_*` + 8
  `TestInspectSkillMirror_*`), plus evidence file `.moai/reports/t563/m1-exit-gate.txt`.
  `doctor_codex.go` untouched (seam swap is M2).
- **Swept count: 20** (12 stale-skill + 8 mirror-inspection) — selector
  `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_|TestInspectSkillMirror_' -count=1
  -timeout 1800s -v` → `PASS / ok github.com/modu-ai/moai-adk/internal/cli 0.935s`; all 20 PASS.
  Count > 0 per plan.md §F M1 exit gate (empty-sweep green rejected).
- Scoped verification: `go vet ./internal/cli/` → exit 0; `golangci-lint run internal/cli/...` →
  exit 0 ("0 issues."); scoped baseline `go test ./internal/cli/ -run 'Codex' -count=1
  -timeout 1800s` → `ok ... 149.381s` (indirect suite incl. new file green). NO local full suite
  (REQ-007 — CI owns the full verdict).
- Pinning approach: struct-level direct assertions on `codexFinding` fields (summary exact-match
  where static, detail ordered-phrase, severity pinned to the ADVISORY zero value) and on
  `skillMirrorState` fields (per-counter). Fixture reuse: `writeCodexHomeConfig`, `stubCodexHome`,
  `t468HomeSkillFile`, `absentSkillPath`, `liveSkillFile`, `writeSkillMirror` from
  `doctor_codex_test.go` — the M2 identity proof (REQ-004) should reuse these same builders plus
  this file's fixtures as the pre/post comparison set.
- Seam discipline honored: `osStatFn` NOT overridden anywhere in M1 (stat observation is M2's
  REQ-006); `codexUserHomeDir` overridden serially only (`stubCodexHomeUnresolvable`, plus
  `t.Setenv(codexHomeEnvVar, ...)` — no `t.Parallel()` in the file).
- M1 → M2 handoff: M2 should (a) swap both stat sites to `osStatFn`, (b) reuse the fixture
  builders above for the output-identity proof — re-run this file's 20 tests + the `-run 'Codex'`
  scoped baseline on the swapped tree and require identical results, (c) add the REQ-006
  observation tests asserting the recorded stat argument per site. Note for M2: `CODEX_HOME`
  points at the CONFIG DIR (the home's `.codex`), not the home root — config lookup joins only
  the file's base name (`codexUserSkillConfig`); observation fixtures for site B should set
  `codexHomeEnvVar` to `filepath.Join(home, ".codex")`.

### M2 — seam swap + observation tests + output-identity proof (commits `bca046150`, `02d8be597`)

- **Seam commit `bca046150`** (AC-SEAM-001/002/003, minimal): `internal/cli/doctor_codex.go`
  diff = exactly 2 lines — `:459` `os.Stat(entryPath)` → `osStatFn(entryPath)` (site A),
  `:857` `os.Stat(statPath)` → `osStatFn(statPath)` (site B). No comment rewrites, no import
  churn, no logic edits.
- **Grep assertion numbers (AC-SEAM-001, on the swapped tree)**: `grep -c 'os\.Stat('`
  doctor_codex.go → **0**; `grep -n 'os\.Lstat('` → **`:441` only (1 match — survives, spec
  §3.2)**; `grep -c 'osStatFn'` → **2**.
- **Observation tests (REQ-006)** — 6 new, pinned prefixes, all serial (`osStatFn` save →
  replace → `t.Cleanup` restore, no `t.Parallel`):
  `TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths` (recorded args = declared
  absolute + seam-expanded home-relative, in order; relative / oddly-formed / path-less
  entries record ZERO stat calls) · `..._InjectedErrNotExistDrivesMissing` (path EXISTS,
  injected ENOENT → missing bucket) · `..._InjectedNilDrivesResolves` (path ABSENT, injected
  nil → resolves, silent) · `..._InjectedOtherErrorDrivesIndeterminate` (other error →
  indeterminate, never missing) · `TestInspectSkillMirror_StatSeamRecordsMirrorJoinedPath`
  (recorded args = `filepath.Join(mirrorDir, name)`, lexical order) ·
  `TestInspectSkillMirror_InjectedResultsDriveBuckets` (ENOENT → dangling by name, other
  error → indeterminate).
- **Output-identity proof (AC-SEAM-005, REQ-004)** — artifact
  `.moai/reports/t563/m2-identity-proof.txt` (+ retained pre-seam dump
  `m2-identity-pre.txt`; post dump byte-identical, removed after verification):
  - Method: untracked capture harness, identical bytes on both sides, dumping the observed
    surface (codexFinding ok/severity/summary/detail; skillMirrorState counters;
    checkCodexWiring Status/Message/Detail) with temp paths normalized to stable tokens.
    PRE captured @ `7d57ff2f5`, POST captured @ `bca046150`.
  - **(a) fixture/staging count: 25** (12 site-B direct + 8 site-A direct + 5 wiring-level
    covering the existing families' bucket states incl. the fatal non-boolean split, rendered
    mirror-absent/dangling findings, copy-mode/unmirrored detail, dir-not-missing).
  - **(b) ok=true traversal count (HARD-NEW-1): 6** (b02, b03, b05, b06, b07, b12 — live
    findings both sides; site-A analog: 4 fixtures carry non-zero dangling/indeterminate
    counters). NOT vacuous.
  - **(c) pre/post diff: EMPTY** (`diff` exit 0, no output — byte-identical).
  - **Verdict: identity HOLDS — no DISCOVERY to report.** The existing
    `TestCodexSkillPath_*` / `TestCheckCodexWiring_*` families additionally re-ran green on
    the swapped tree (below), completing the comparison set.
- **Scoped verification (post-swap regression)**: `go test ./internal/cli/ -run
  'TestCodexStaleSkillFinding_|TestInspectSkillMirror_|TestCodexSkillPath_|TestCheckCodexWiring'
  -count=1 -timeout 1800s` → `ok ... 0.848s` (26 seam tests + both existing families, green on
  the swapped tree); `go vet ./internal/cli/` → exit 0; `golangci-lint run internal/cli/...` →
  exit 0 ("0 issues."). Tree quiescence checked before and after the batch — only this SPEC's
  test file modified; no foreign writes. NO local full suite (REQ-007).
- Harness footnote: the capture harness initially triggered the local ast-grep
  `go-error-ignored-blank` rule on two-value call assignments (`f, ok = fn()`) and was
  restructured through a result-struct helper; the blocking write also exposed a fixture defect
  mid-proof (w3 initially created the canonical dangling target, so the link resolved and the
  dangling finding never rendered) — fixed before the PRE baseline was accepted.

### M1 observed-behavior notes for the lead (nothing fixed — pinned as-is)

1. `codexStaleSkillFinding`'s Detail leads with the RESOLVED config path (a temp/fixture path at
   test time); the display form (`~/.codex/config.toml` / `$CODEX_HOME/config.toml`) appears in
   the SUMMARY only. Characterization reflects this split; no action implied.
2. A mirror symlink loop contributes to `st.indeterminate` once PER LOOP ENTRY (2 for a 2-link
   loop), since each entry is stat'ed independently. Pinned as current behavior.
3. Regular files occupying a mirror path are counted as NOTHING (copyMode 0, indeterminate 0) per
   the source comment — an intentionally-unclassified shape. Pinned.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated at M-final>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

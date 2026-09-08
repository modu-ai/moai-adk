# SPEC-HARNESS-EVIDENCE-WRITE-001 — Implementation Plan

Tier: S (single package `internal/spec`, ~4 files: 2 t362 harnesses + 1 t528 probe + 1 new guard
test; well under 300 LOC delta). Artifact set: spec.md + plan.md; AC inline in spec.md §3.

## §A Context

See spec.md §A for the full premise record (lane-10 re-verification, HEAD `3ac58b5a1`,
2026-09-08). In short: two gated t362 harnesses write repo-tracked evidence paths by default and
one ungated t528 probe writes into `.moai/reports/t528/probe/out/` on every test run. The fix
retargets all three defaults to per-run `t.TempDir()` and adds a same-layer guard test so the
invariant "no test in this package writes into the repo by default" is mechanically enforced
without an allowlist carve-out.

## §B Known Issues

- The defect's write sites are exact: `lint_req_widen_corpus_test.go` lines ~239-247 (constant
  line 19) and `lint_req_widen_decompose_test.go` lines ~541-548 (constant line 15); probe
  default in `t528ProbeOutDir` (`zz_t528_anchor_probe_test.go` ~lines 56-65).
- The two harnesses' reports embed a `# produced by:` header naming the old invocation; both
  headers must be updated to describe the new default (TempDir) + the `MOAI_T362_EVIDENCE_OUT`
  durable-capture procedure, and must name NO repo-relative default path.
- `t.TempDir()` directories are deleted when the test completes — a post-run copy from the
  `t.Logf`-announced path is impossible (plan-audit review-1 D1). Durable capture selects its
  destination BEFORE the run, via the single env override; implement it with loud-fail
  no-clobber semantics (`os.Stat` existence check → `t.Fatalf` on existing target).
- `findRepoRoot(t)` itself stays — it is used elsewhere for reads; only its use as a WRITE anchor
  is removed.
- The guard's scan surface is `internal/spec/*_test.go` sources. Reads of `.moai/reports/**`
  must not be flagged (evidence-as-input is legitimate): the discriminant is the write primitive
  (`os.WriteFile` / `os.MkdirAll` / equivalents), not the path string alone.

## §C Pre-flight

- Confirm worktree HEAD matches the branch tip immediately before the first edit (staleness
  rule) and that `git status --porcelain` shows no pre-existing modification under
  `.moai/reports/t362/`.
- Confirm both pinned files are currently tracked: `git ls-files .moai/reports/t362/` (12 files
  expected, including both report files).
- Record a before-image of the two pinned files (e.g. `git rev-parse HEAD:<path>`) so AC-007's
  zero-diff claim is checkable against a named baseline.

## §D Constraints

- Do NOT modify any file under `.moai/reports/**`.
- Do NOT touch `ac_count_clause_test.go` or `zz_t528_overacceptance_test.go`.
- Do NOT touch the `MOAI_T362_CORPUS_SCAN` gating semantics — the scan stays opt-in; only its
  output destination changes.
- `T528_PROBE_OUT` env override behavior is preserved byte-for-byte (explicit dir wins).
- No new dependency; stdlib only. Go comments in English.

## §E Self-Verification

Run the affected package only (no full-suite local run, per repo discipline):

Every test invocation carries `-count=1` (a cached PASS replay is vacuous green — it never
exercises the current write path) and runs in a scrubbed shell so an inherited
`MOAI_T362_CORPUS_SCAN` / `T528_PROBE_OUT` cannot silently exercise the wrong branch
(scrub + command in ONE compound invocation):

```bash
go vet ./internal/spec/
unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run 'TestCorpusREQWideningMeasurement|TestCorpusRejectedREQIDDecomposition' -count=1 -v   # without gate: both SKIP (gate preserved)
unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/ -run 'TestCorpusREQWideningMeasurement|TestCorpusRejectedREQIDDecomposition' -count=1 -v
unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run TestT528Anchor -count=1 -v
unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run TestNoTestWritesRepoTree -count=1 -v   # guard test, name TBD at run phase
unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 MOAI_T362_EVIDENCE_OUT=<tmp-out> go test ./internal/spec/ -run TestCorpusREQWideningMeasurement -count=1 -v   # AC-008a override honored
git status --porcelain .moai/reports/
```

The no-clobber negative (AC-008b) re-runs the AC-008a invocation against the same `<tmp-out>`
and expects FAIL with the target file's `shasum` unchanged.

Expected: gated runs PASS with paths under a `/var/folders/...` (or OS-equivalent) TempDir
announced via `t.Logf`; the override run lands at `<tmp-out>`; `git status --porcelain
.moai/reports/` prints NOTHING.

Mutant evidence (AC-006, run inside the same session, reverted immediately after):

1. Temporarily add a repo-anchored write (e.g. `os.WriteFile(filepath.Join(root,
   ".moai/reports/mutant-probe.txt"), ...)` in a scratch test).
2. Run the guard → observe RED with file+line finding (capture verbatim output to the §E.2
   evidence path).
3. Revert the mutant; re-run the guard → GREEN.
4. Record both outputs side by side in progress.md §E.2 — a guard that has never printed RED on
   its own mutant is unadopted per this repo's discipline (an unreached mutant and a real
   survivor print the same `ok`).

## §F Milestones

Ordered by decision-reversibility — the retargeting decisions come first (highest change
likelihood), the mechanical sweep last.

- **M1 (High) — Retarget the t362 pair.** In `lint_req_widen_corpus_test.go` and
  `lint_req_widen_decompose_test.go`: default write target becomes `t.TempDir()`; delete or
  repurpose `corpusMeasurementRelPath` / `decomposeReportRelPath` so no repo-anchored write path
  remains constructible; keep `t.Logf` path reporting. Add the single durable-capture override
  `MOAI_T362_EVIDENCE_OUT=<dir>` (pre-run destination selection, loud-fail no-clobber — existing
  target file → `t.Fatalf`, nothing written); update the in-report `# produced by:` headers to
  document the TempDir default + `MOAI_T362_EVIDENCE_OUT` capture procedure and to name NO
  repo-relative default path. Gate semantics unchanged.
- **M2 (High) — Retarget the t528 probe default.** `t528ProbeOutDir`: unset `T528_PROBE_OUT` now
  resolves to `t.TempDir()`; set env still wins. Update the probe's header comment if it names
  the old default path.
- **M3 (High) — Same-layer guard test.** New `internal/spec` test (table-driven where natural,
  per package convention): scans `internal/spec/*_test.go` sources for repo-tree write targets —
  a `.moai/`-anchored path reaching a write primitive — while allowing reads. Must produce the
  demonstrated mutant evidence per §E (RED on mutant, GREEN after revert) before adoption.
- **M4 (Medium) — Verification sweep.** §E command batch; confirm zero diff under
  `.moai/reports/`; confirm read-only consumer tests untouched (`git status` on those paths);
  record evidence paths in progress.md §E.2.

## §G Anti-Patterns

- Do NOT "fix" the harness by making the write conditional on an env var and keeping the repo
  path as the default — the default must be TempDir.
- Do NOT add the guard as a lint `Rule` inside the package's rule set — the observation-only
  clause binds Rules, and a Rule cannot test-test-sources cleanly; this is a same-package test.
- Do NOT weaken the guard with an allowlist of "known good" repo writers — the invariant is
  uniform precisely because there are none after M1/M2.
- Do NOT re-run or "improve" the pinned evidence files; deliberate capture is an operator act
  (REQ-007), outside this SPEC.
- Do NOT scan non-test sources or other packages in the guard (scope discipline).

## §H Cross-References

- spec.md §A (premise record), §B (REQ-001..007), §C (AC-001..008)
- `internal/spec/CLAUDE.md` (package conventions)
- CLAUDE.local.md §6 (t.TempDir test isolation), §14 (hardcoding prevention)
- Lane-10 premise re-verification record (card t569 dispatch, 2026-09-08)

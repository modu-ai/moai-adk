---
id: SPEC-HARNESS-EVIDENCE-WRITE-001
title: "Measurement harnesses must not overwrite tracked evidence files in the repository tree"
version: "1.0.1"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tier: S
tags: "test-hygiene, evidence-integrity, guard-test, t569"
---

# SPEC-HARNESS-EVIDENCE-WRITE-001 — Measurement harnesses must not overwrite tracked evidence files

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-08 | 1.0.0 | Initial draft (card t569; plan-phase artifacts by manager-spec) |
| 2026-09-08 | 1.0.1 | Delta revision per plan-audit review-1 (D1-D3): REQ-007 rewritten to a pre-run `MOAI_T362_EVIDENCE_OUT` override with no-clobber semantics (post-run copy from `t.TempDir()` is impossible — Go deletes it at test end); AC-001/002/004 given environment-scrub + `-count=1` pins; AC-008 added for the durable-capture path |

---

## §A Context

A gated measurement harness inside `internal/spec` overwrites another card's TRACKED evidence
files when re-run. Two `internal/spec` test harnesses and one ungated probe default to writing
into the repository's evidence tree (`.moai/reports/**`), silently rewriting verdict evidence of
CLOSED SPECs. The lane (lane-10) re-verified the following premises against this worktree
(HEAD `3ac58b5a1`, 2026-09-08); the lane's reproduction runs are cited as provided evidence and
are NOT re-run by this SPEC's plan phase:

1. `internal/spec/lint_req_widen_decompose_test.go` — `TestCorpusRejectedREQIDDecomposition`
   (gate `MOAI_T362_CORPUS_SCAN=1`) writes `findRepoRoot(t)` + `.moai/reports/t362/m2-gate0-decomposition.txt`
   via `os.MkdirAll` + `os.WriteFile` (constant `decomposeReportRelPath`, line 15; write site
   ~lines 541-548). Lane repro: PASS + `git diff --stat` 387 insertions / 93 deletions on the
   TRACKED file (244 → 538 lines).
2. `internal/spec/lint_req_widen_corpus_test.go` — `TestCorpusREQWideningMeasurement` (same gate)
   writes `.moai/reports/t362/m1-corpus-measurement.txt` (constant `corpusMeasurementRelPath`,
   line 19; write site ~lines 239-247). Lane repro: PASS + 291 insertions / 36 deletions.
3. Both evidence files are git-tracked (`git ls-files .moai/reports/t362/` lists 12 files
   including both). They pin the verdict evidence of CLOSED SPECs (SPEC-COVERAGE-RULE-SCOPE-001,
   SPEC-SPEC-LINT-BLIND-AXES-001) — a re-run silently rewrites that evidence.
4. Attribution: the repo-path write was introduced by commit `130846ab2` (t362), verified an
   ancestor of t518's `fc02d2542` via `git merge-base --is-ancestor`. t518's audit (F-A2)
   discovered the defect; it did not create it.
5. Adjacent instance: `internal/spec/zz_t528_anchor_probe_test.go` — `TestT528Anchor` runs
   UNGATED on every `go test ./internal/spec/` and writes 8 files to
   `../../.moai/reports/t528/probe/out/` by default (`t528ProbeOutDir`, env override
   `T528_PROBE_OUT`). The out/ directory holds 0 tracked files, so no evidence is LOST — but the
   probe writes into the repo's evidence tree on every test run, and its own header states the
   intent "run-scoped directory, NOT the plan-phase artifact directory". Retargeting its default
   to `t.TempDir()` FULFILLS that stated intent and lets a same-layer guard carry NO exceptions.
6. READ-ONLY usages that must NOT be touched (legitimate evidence-as-input):
   `internal/spec/ac_count_clause_test.go` reads tracked `.moai/reports/t338/ac-count-baseline.txt`;
   `internal/spec/zz_t528_overacceptance_test.go` reads tracked
   `../../.moai/reports/t528/probe/nondecl-bullets.txt`. `internal/graph/*` test fixtures already
   write under `t.TempDir()` — verified safe.

### Root Cause (Five-Whys summary)

The t362 harnesses chose a repo-relative output path so the gate-0 measurement would sit beside
the card's other evidence. That conflated two concerns: run output (ephemeral, per-run) and
durable evidence (an operator-curated, append-named artifact). The default became the durable
location, so every re-run mutated pinned history.

## §B Requirements (GEARS)

- **REQ-001 (Ubiquitous)** — The `internal/spec` test suite shall not write any file into the
  repository working tree by default; every default harness output target shall resolve to a
  per-run `t.TempDir()` directory.
- **REQ-002 (Event-driven)** — When `TestCorpusREQWideningMeasurement` executes, the harness
  shall write `m1-corpus-measurement.txt` only under the per-run `t.TempDir()` directory and
  shall report the exact output path via `t.Logf`.
- **REQ-003 (Event-driven)** — When `TestCorpusRejectedREQIDDecomposition` executes, the harness
  shall write `m2-gate0-decomposition.txt` only under the per-run `t.TempDir()` directory and
  shall report the exact output path via `t.Logf`.
- **REQ-004 (Event + Where)** — When `TestT528Anchor` executes while `T528_PROBE_OUT` is unset,
  the probe shall write its eight output files into the per-run `t.TempDir()` directory;
  **Where** the operator sets `T528_PROBE_OUT`, the probe shall redirect output to that directory
  (the explicit escape hatch is preserved unchanged).
- **REQ-005 (Unwanted)** — The harness source shall not carry a code path that joins a
  repository-root-resolved path with a `.moai/**` destination for writing (`os.WriteFile`,
  `os.MkdirAll`, or file-modifying equivalents). The repo-relative write-path constants
  (`decomposeReportRelPath`, `corpusMeasurementRelPath`) shall be deleted or repurposed so no
  such path remains constructible.
- **REQ-006 (Event-detected)** — When any `internal/spec/*_test.go` re-introduces a repo-tree
  write target (a `.moai/`-anchored path used with a file-write primitive), the same-package
  guard test shall FAIL with a finding naming the offending file and line, while continuing to
  allow READS of tracked evidence files (evidence-as-input is legitimate).
- **REQ-007 (Where + When)** — Where an operator wants durable evidence capture and has set
  `MOAI_T362_EVIDENCE_OUT=<dir>` BEFORE the run, the t362 harness shall write its report into
  `<dir>`; **When** the target file already exists at that path, the harness shall fail loudly
  and write nothing (no-clobber — a deliberate-capture path that silently overwrites would
  recreate the very defect this SPEC exists to kill, so durable capture is always to a NEW
  file). Runs without the override keep the per-run `t.TempDir()` default unchanged; the
  in-report `# produced by:` header shall document the override procedure and name no
  repo-relative default path. (A post-run copy from the `t.Logf`-announced TempDir path is NOT
  a viable capture mechanism — Go deletes `t.TempDir()` directories when the test completes.)

## §C Acceptance Criteria (Given-When-Then)

- **AC-001** — Given a clean working tree and a scrubbed shell (`MOAI_T362_CORPUS_SCAN` and
  `T528_PROBE_OUT` unset) with the gate re-set for the single invocation, When
  `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/ -run TestCorpusREQWideningMeasurement -count=1`
  runs, Then it PASSes non-cached, the report lands under `t.TempDir()` (path announced via
  `t.Logf`), and `git status --porcelain .moai/reports/t362/` prints nothing.
- **AC-002** — Given a clean working tree and the same scrubbed shell, When
  `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/ -run TestCorpusRejectedREQIDDecomposition -count=1`
  runs, Then it PASSes non-cached, the decomposition lands under `t.TempDir()`, and
  `git status --porcelain .moai/reports/t362/` prints nothing.
- **AC-003** — Given a scrubbed shell (`MOAI_T362_CORPUS_SCAN` and `T528_PROBE_OUT` unset), When
  `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run TestT528Anchor -count=1`
  runs, Then the eight probe files land under `t.TempDir()` and no new or modified file appears
  under `.moai/reports/t528/`.
- **AC-004** — Given `T528_PROBE_OUT=<fresh dir>` set and `MOAI_T362_CORPUS_SCAN` scrubbed, When
  `unset MOAI_T362_CORPUS_SCAN && T528_PROBE_OUT=<fresh dir> go test ./internal/spec/ -run TestT528Anchor -count=1`
  runs, Then the probe output lands in `<fresh dir>` and the test PASSes non-cached.
- **AC-005** — Given the fixed package sources, When the guard test scans `internal/spec/*_test.go`
  for repo-anchored write targets, Then it finds none and PASSes (guard GREEN on the fixed tree).
- **AC-006** — Given a mutant that temporarily re-introduces a repo write target (e.g. a
  `.moai/reports/`-anchored `os.WriteFile` in any `internal/spec` test), When the guard test
  runs against the mutant, Then it FAILS (RED) naming the file and line. The demonstrated mutant
  is the discriminative evidence — an unreached mutant and a real survivor print the same `ok`.
- **AC-007** — Given the run-phase diff, When inspected, Then the pinned evidence files under
  `.moai/reports/t362/` show ZERO diff, and `ac_count_clause_test.go` /
  `zz_t528_overacceptance_test.go` (read-only evidence consumers) are unmodified.
- **AC-008a** — Given `MOAI_T362_EVIDENCE_OUT=<dir>` set, the gate on, and NO pre-existing file
  at the target path, When
  `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 MOAI_T362_EVIDENCE_OUT=<dir> go test ./internal/spec/ -run TestCorpusREQWideningMeasurement -count=1`
  runs, Then the report lands at `<dir>` and the run PASSes non-cached.
- **AC-008b** — Given the same `<dir>` now holds the previous report, When the identical
  invocation runs again, Then it FAILS with a loud no-clobber error and the existing file is
  byte-identical before and after (`shasum` unchanged) — nothing is overwritten.
- **AC-008c** — Given any produced report (TempDir default or override), When its in-report
  `# produced by:` header is read, Then it documents the `MOAI_T362_EVIDENCE_OUT` capture
  procedure and names NO repo-relative default path.

## §D Non-Functional Constraints

- Go code, comments, and godoc in English (language.yaml `code_comments: en`).
- Tests use `t.TempDir()` per CLAUDE.local.md §6 (test isolation) — no `os.TempDir()`-joined
  project paths.
- Package conventions per `internal/spec/CLAUDE.md` apply; its "Rules are observation-only,
  never write files" clause binds lint Rule implementations, NOT test harnesses — the guard is a
  test, not a lint Rule.
- The guard reads test SOURCES (or compile-anchored path constants); it must not execute the
  tests it scans.
- Durable capture for the t362 pair uses exactly ONE env override —
  `MOAI_T362_EVIDENCE_OUT=<dir>` (named consistently with the existing `MOAI_T362_CORPUS_SCAN`
  gate), with loud-fail no-clobber semantics. No dual-env scheme; `T528_PROBE_OUT` remains the
  t528 probe's separate, unchanged override.
- The gated scans (`MOAI_T362_CORPUS_SCAN=1` runs) are NOT re-run at plan phase; the lane's
  reproduction numbers are the provided evidence.
- Go's `t.TempDir()` directories are removed when the test and its subtests complete
  (stdlib contract) — nothing announced via `t.Logf` survives the run. Durable capture therefore
  MUST select its output location BEFORE the run (REQ-007's `MOAI_T362_EVIDENCE_OUT` override),
  never via a post-run copy from the TempDir path.

## §E Out of Scope

### Out of Scope — Other packages' write behavior

- `internal/graph/*` test fixtures already write under `t.TempDir()`; no changes there.

### Out of Scope — Pinned evidence migration or regeneration

- The 12 tracked files under `.moai/reports/t362/` (and `.moai/reports/t528/probe/` pinned
  before-images) are NOT modified, renamed, or regenerated by this SPEC. Deliberate re-capture
  is a future operator act per REQ-007.

### Out of Scope — Read-only evidence-as-input tests

- `ac_count_clause_test.go` and `zz_t528_overacceptance_test.go` READ tracked evidence files;
  this behavior is legitimate and stays untouched.

### Out of Scope — Repo-target rejection hardening for existing overrides

- An operator pointing `T528_PROBE_OUT` at a repo-internal (pinned-evidence) directory can
  still overwrite tracked before-images — status-quo behavior this SPEC preserves byte-for-byte.
  Rejecting repo-resolved override targets is follow-up-SPEC material (plan-audit review-1 D5),
  out of scope here.

## §F Cross-References

- Card t569 (operator-approved direction: redirect measurement output to `t.TempDir()` or an
  explicitly operator-chosen location, never into the repo by default).
- Origin incident: t518 audit finding F-A2 (discovery); defect introduced by `130846ab2` (t362).
- Closed SPECs whose pinned evidence is at risk: SPEC-COVERAGE-RULE-SCOPE-001,
  SPEC-SPEC-LINT-BLIND-AXES-001.
- Package conventions: `internal/spec/CLAUDE.md`.
- Related discipline: this repo's mutation-check doctrine — a demonstrated mutant is the
  discriminative evidence for any guard test.

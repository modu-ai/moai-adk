---
id: SPEC-CODEX-MIRROR-DOCTOR-001
title: "Implementation plan — .agents/skills mirror-state diagnostic"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
---

# Implementation Plan — SPEC-CODEX-MIRROR-DOCTOR-001

Ordered by decision-reversibility: the decisions most likely to change sit first, the mechanical
work last.

## §A Context

Card t498, worktree `.claude/worktrees/t498`, branch `WT-codex-mirror-doctor`, base `ace1c5440`.
Authoritative input: `.moai/reports/t498/root-cause.md` (4 measured claims, 3 gaps, residual risk).
No measurement in that report is re-derived here.

## §B Tier

**Tier M.** Files affected: 2 (`internal/cli/doctor_codex.go`,
`internal/cli/doctor_codex_test.go`) — a Tier S signal. Estimated LOC: ~120 production +
~200 test = ~320, which crosses the Tier S 300-LOC ceiling. The tie is broken toward M for a
third reason: the change carries a named regression risk (the un-nagging invariant,
`doctor_codex.go:96-103`) whose acceptance criterion deserves its own artifact rather than an
inline `§3` bullet. Artifact set therefore spec.md + plan.md + acceptance.md (3 files), plan-auditor
PASS threshold 0.80, REQ ceiling 16 (11 used), AC ceiling 16.

## §C Known issues / constraints inherited

1. `.agents/` is gitignored in this repository, so no mirror exists in any tree here (root-cause.md
   Claim 3). Every test must therefore build its own mirror fixture under `t.TempDir()`; there is no
   in-repo mirror to read.
2. The check's panel-width bound is a *shared* resource: `joinCodexSummaries` drops summaries from
   the tail, so a new mirror summary appended after existing findings can be dropped. That is the
   intended behaviour, not a defect — the acceptance criteria must not assert unconditional Message
   presence when other findings are also raised.
3. Windows copy-fallback behaviour is unexercised (root-cause.md Gaps). The plan does not add a
   Windows-only branch; the copy-mode state is counted and reported the same way on every host.

## §D Milestones

### M1 — Mirror state model (highest reversibility: the classification could change)

The decision most likely to be revisited is *which states warrant a Message finding*. It is
therefore built first and in isolation, as a pure inspection function that returns a state struct,
so the finding/detail split is one call site away from being re-tuned.

- Add a small unexported result type carrying: `dirPresent bool`, `dangling []string`,
  `copyMode int`, `unmirrored int`, `indeterminate int`.
- Add the inspector: reads `.agents/skills` with `os.ReadDir`; per entry uses `Lstat` to separate a
  symlink from a real directory, `Readlink` + `Stat` against the mirror directory to decide
  dangling; reads `.claude/skills` once to compute `unmirrored`.
- Every read error is folded into `indeterminate` and never into a count that drives a finding
  (REQ-CMD-010).
- The inspector performs no writes of any kind (REQ-CMD-002).

### M2 — Wiring into `checkCodexWiring` (the un-nagging invariant lives here)

- Call the inspector **after** the `!wired && !codexInstalled` early return, so the claude-only path
  is untouched by construction (REQ-CMD-003).
- Call it only in the `wired` branch (REQ-CMD-008): the `!wired` branch already emits
  `initCodexAdvice`, and a second directive there would double-nag a project that has not opted in.
- Append at most one absent-mirror finding and at most one dangling finding to `problems`; append
  the copy-mode / unmirrored / indeterminate counts to `extraDetail`.

### M3 — Directive constant and width

- Add a `mirrorRedeployAdvice` constant carrying `run moai update --templates-only --force --yes`,
  in the shape of the existing `reTrustAdvice` / `initCodexAdvice` constants.
- Confirm each new summary is inside 113 runes standalone (REQ-CMD-009).

### M4 — Tests (mechanical; follows the file's existing seam pattern)

- Reuse `stubCodexLookup`, `stubCodexHome`, `wireProjectForDoctor`; add a mirror-fixture helper that
  builds `.agents/skills` entries of each shape under the wired project root.
- One test per acceptance criterion in acceptance.md, including the un-nagging regression test.

### M5 — Verification

- `go test ./internal/cli/...` (package scope; no full local suite per CLAUDE.local.md §4).
- `go vet ./internal/cli/...`, `golangci-lint run ./internal/cli/...`.
- `GOOS=windows GOARCH=amd64 go build ./...` (cross-platform path handling).
- Golden-fixture check: `internal/cli/testdata/doctor-nocolor.golden` must be unchanged — the
  claude-only render gains no row.

## §E Anti-patterns to avoid

- **Repairing on read.** Any `os.Symlink`, `os.Remove`, or `MkdirAll` in the doctor path violates
  REQ-CMD-002 outright.
- **Warning on copy mode.** It is the correct materialization on symlink-less hosts; warning would
  make a permanent row out of a working mirror.
- **Deriving a "missing skills" finding from a `.claude/skills` count.** The denominator is
  unobservable from doctor (spec §3, row 4).
- **A new top-level `DiagnosticCheck` row.** That would give every claude-only user a permanent line
  — the un-nagging invariant defeated at the registration layer rather than inside the function.
- **Asserting Codex behaviour.** Nothing observed says how Codex reacts to a dangling link or a
  partial mirror; every message stays an action directive, never a claim about Codex.

## §F Cross-references

- `.moai/reports/t498/root-cause.md` — Claims 1-4, Gaps, Residual risk.
- `internal/cli/doctor_codex.go` — host function, finding shape, width bound, seams.
- `internal/template/skill_mirror.go` — mirror layout, link body, mode enum, fail-open contract.
- `internal/cli/doctor_codex_test.go` — the seam-stubbing test pattern this change follows.

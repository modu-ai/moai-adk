# M1 verdict — SPEC-SPECLINT-GATE-SIGNAL-001 (card t525)

The three-axis verdict (spec.md §1.4), each row evidence-gated per REQ-SLGS-002.
All measurements: tree-build `go run ./cmd/moai` (NOT the installed binary — see
m1-demographics.md provenance), tree `b6efc874f`, 2026-09-08, corpus 804 SPEC dirs.
Companion measurement record: `m1-demographics.md` + raw `census-b6efc874f.json`.

## Axis (a) — the warnings are real debt, the gate is right → 성립, 지연 (deferred per REQ-SLGS-011)

- **Claim**: the standing warning stock is real debt, and paying it is NOT this card's job —
  gated on t518 landing (REQ-SLGS-011).
- **Evidence**:
  ```
  $ jq -r '[.[] | select(.severity=="warning")] | group_by(.advisory // false)[] | "advisory=\(.[0].advisory // false)\t\(length)"' census-b6efc874f.json
  advisory=false	2
  advisory=true	4376
  ```
  The 4,376 advisory-marked findings (12 rules, breakdown in m1-demographics.md) exist and
  are demoted by `applyEraDemotion` (`internal/spec/lint.go:345`; cause `dd644b5e0`,
  2026-07-21 — file-content-based, not git-dependent). They are debt in the (a) sense:
  standing, uncounted by the gate, owned by post-t518 cards.
- **Baseline-attribution**: this run, tree `b6efc874f`, census JSON committed with this report.
- **Disposition**: no (a)-axis debt payment in this card (REQ-SLGS-011 interlock; t518
  measured NOT landed twice this milestone — `git merge-base --is-ancestor 6cfcfef00
  origin/develop` → non-ancestor, pre- and post-absorb-merge).

## Axis (b) — exiting 1 on warnings is the wrong policy → 구조 입증됨 + 동기 갱신 (motivation corrected by measurement)

- **Claim**: `--strict` turns a warning-only corpus red — structurally proven by re-execution.
  The card's premise about the red's SIZE is corrected: today's red is carried by exactly 2
  findings, both inside M4's scope.
- **Evidence** (the AC-SLGS-002 rc=1 half, re-executed this milestone):
  ```
  $ go run ./cmd/moai spec lint --strict
  0 error(s), 4378 warning(s)
  exit status 1
  RC_STRICT=1
  ```
  0 errors, exit 1 — escalation of non-advisory warnings alone. The 2 non-advisory
  findings (census, `advisory=false	2`) are `SpecsDirMissingSpecFile` on
  SPEC-V3R4-CC2X-ADOPT-001/002 — precisely M4's work item (REQ-SLGS-012).
- **Mutation half (AC-SLGS-002)**: NOT observed in M1 — per acceptance.md, the baseline-
  absorption mutation belongs to M2's observation window (the `--baseline` path does not
  exist yet). Recorded as pending, not passed.
- **Baseline-attribution**: this run, tree `b6efc874f`.
- **Motivation correction (material finding for the lead)**: the card's premise "advisory 가
  아닌 경고가 수천 건" (spec.md §1) is falsified at this tree — the mass stock is already
  advisory (demotion landed `dd644b5e0`, 2026-07-21, before the historical sample tree
  `0b1e27877`). The "permanently red" condition since 09-03 (CI record in spec.md §1) is
  carried by the 2-finding M4 pair, which t365 (`85a783c9e`, 2026-09-02) surfaced from
  silence. Consequence, measured-conditional: once M4 closes the pair **with recorded
  reasons** (AC-SLGS-012), bare `--strict` reads 0 non-advisory warnings → green, and any
  NEW non-advisory warning (any rule firing on a modern active SPEC) reddens it again —
  the gate is then a functioning signal without further mechanism. M2's surviving value
  (already kickoff-approved as (iii)): per-rule delta visibility (REQ-SLGS-006) and the
  t518 population-move absorption via the M3-gated re-baseline (REQ-SLGS-011). This row
  reports the premise shift; the (i)-(iv) mechanism decision stays with the kickoff record.

## Axis (c) — the warnings themselves are false positives → 소관 외 (t518's axis)

- **Claim**: whether the advisory boundary is drawn correctly is SPEC-SPEC-LINT-BLIND-AXES-001's
  (t518) subject; this SPEC does not touch it (plan.md §G anti-pattern: advisory 둔갑 금지).
- **Evidence** (measured scope of what t518 owns — the boundary it would redraw):
  ```
  $ jq -r 'group_by([.code, .severity, (.advisory // false)])[] | "\(.[0].code)\t\(.[0].severity)\t\(.[0].advisory // false)\t\(length)"' census-b6efc874f.json
  CoverageIncomplete	warning	true	3620
  ModalityMalformed	warning	true	412
  MovingRefUnpinned	warning	true	115
  StatusTransitionInvalid	warning	true	102
  LegacyEARSKeyword	warning	true	48
  MissingExclusions	warning	true	27
  StatusGitConsistency	warning	true	18
  FrontmatterInvalid	warning	true	14
  StatusTokenUnrecognized	warning	true	7
  InvalidREQID	warning	true	6
  SyncSHASlotFormat	warning	true	6
  SpecsDirMissingSpecFile	warning	false	2
  OwnershipTransitionInvalid	warning	true	1
  ```
  4,376 findings across 12 rules carry `advisory=true` — that population and its demotion
  boundary (era/terminal in `applyEraDemotion`, heuristic in StatusGitConsistency) are what
  t518's re-classification would move. Nothing fixed here.
- **Baseline-attribution**: this run, tree `b6efc874f`.

## M1 AC matrix (AC-SLGS-001..004)

| AC | Verdict | Command | Observed |
|---|---|---|---|
| AC-SLGS-001 | **PASS** | `go run ./cmd/moai spec lint --json > census-b6efc874f.json`; `jq length` | rc=0, 4378 findings; command verbatim + tree SHA `b6efc874f` recorded in m1-demographics.md; SHA matches `git rev-parse --short HEAD` at measurement |
| AC-SLGS-002 | **PASS (rc=1 half)** | `go run ./cmd/moai spec lint --strict` | `0 error(s), 4378 warning(s)`, rc=1. Mutation half = M2 window (explicitly pending, per acceptance.md observation-window note) |
| AC-SLGS-003 | **PASS** | this file | all three axes carry command + verbatim output (or, for (c), the measured t518 scope); no unsupported "구조상 그렇다" phrasing |
| AC-SLGS-004 | **PASS** | `/usr/bin/grep -rnE '4[,.]?3(44|68)' .moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001 internal/spec internal/cli` | 12 hits, all in the 4 SPEC docs, **0** in internal/spec · internal/cli; every hit line is a sourced historical citation (run id / tree SHA / report path named) or the freeze-prohibition's own illustration (plan.md §G, acceptance.md failure-mode example). Binary used: `/usr/bin/grep` (per counting discipline) |

## Gaps

- AC-SLGS-002's mutation half (baseline absorption → rc=0) is unobserved — M2's window by
  the AC's own observation-window split; recorded as pending, never as pass.
- CI-side advisory split is inferred-equal (demotion is file-content-based, git-independent)
  but was not measured inside a CI runner — no CI measurement exists for the advisory column.
- The 12 AC-SLGS-004 hits were classified by reading each line's bullet/paragraph context;
  that classification is a reading, not a mechanical verdict.

## Residual-risk

- The corpus moves on every SPEC landing (spec.md §1.1): today's 4378/2 split expires at the
  next landing; M2's baseline must be produced from a fresh re-derivation at M2 time, not
  from this census.
- M4 closing the pair flips the gate green — if M4's "why" records are skipped and the
  findings are suppressed by any other means (AC-SLGS-012 failure mode), the green would be
  false signal-restoration. M4's own AC guards this.
- t518's advisory widening, when it lands, moves the 4,376 boundary wholesale — the M3
  interlock (re-baseline gate) is the designed absorption; skipping it invalidates M2's
  baseline rather than this verdict.

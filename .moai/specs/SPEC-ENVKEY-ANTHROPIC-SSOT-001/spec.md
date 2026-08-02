---
id: SPEC-ENVKEY-ANTHROPIC-SSOT-001
title: Single-source the ANTHROPIC_* environment-variable name family through internal/config/envkeys.go
version: 0.2.1
status: completed
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P2
phase: "v3.0.2"
module: internal/config
lifecycle: spec-anchored
tags: "envkeys, hardcoding, refactor, guard-test, ssot"
tier: M
---

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-31 | Initial plan-phase authoring. Baseline measured at `76d9a8f3b`: 83 bare literals across 10 production files; guard-test scope defect identified. | manager-spec |
| 0.2.0 | 2026-07-31 | Audit iteration 2 (plan-auditor FAIL 0.72). `lifecycle` corrected to the canonical enum value; REQ-EAS-007 relabelled Event-driven; A.1 extended with the `pkg/`+`cmd/` zero-literal fact; A.5 census re-measurement gate added. Companion fixes in plan.md (M2 guard design, M6 falsification targets, symbol citations) and acceptance.md (runtime banned-set assertion, lint exit-code capture, positive scope assertion, mechanized AC-EAS-004, recorded baselines). | manager-spec |
| 0.2.1 | 2026-07-31 | Sync-phase close (3-phase plan-run-sync). Frontmatter `in-progress` -> `completed`; `progress.md` section E.4 sync-phase audit-ready signal populated; CHANGELOG `[Unreleased]` -> `### Changed` entry added. No REQ text, no AC row, and no `.go` file changed. | manager-docs |

---

## A. Context

### A.1 Problem

The `ANTHROPIC_*` environment-variable **name** family is referenced as bare string
literals throughout the production Go source. `internal/config/envkeys.go` is the
declared single source of truth (SSOT) for environment-variable names
(CLAUDE.local.md section 14: environment-variable names must be defined as
constants in `internal/config/envkeys.go` and referenced from there), but the
ANTHROPIC family only partially honours it.

Measured at base commit `76d9a8f3b`:

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' --exclude='*_test.go' \
  | grep -v '^internal/config/envkeys.go' | wc -l
# observed: 83
```

Distribution across 10 production files:

| File | Count |
|------|------:|
| `internal/cli/glm.go` | 32 |
| `internal/hook/session_end.go` | 12 |
| `internal/hook/glm_tmux.go` | 9 |
| `internal/hook/session_start.go` | 8 |
| `internal/cli/launcher.go` | 7 |
| `internal/cli/settings.go` | 6 |
| `internal/tmux/cg_detect.go` | 4 |
| `internal/statusline/metrics.go` | 3 |
| `internal/sandbox/env.go` | 1 |
| `internal/cli/worktree/tmux_integration.go` | 1 |
| **Total** | **83** |

Nine distinct literal values are in use:

```
"ANTHROPIC_"                    (bare prefix, used for prefix-matching - see OQ-1)
"ANTHROPIC_API_KEY"             (NO constant exists)
"ANTHROPIC_AUTH_TOKEN"
"ANTHROPIC_BASE_URL"
"ANTHROPIC_DEFAULT_FABLE_MODEL" (NO constant exists)
"ANTHROPIC_DEFAULT_HAIKU_MODEL"
"ANTHROPIC_DEFAULT_OPUS_MODEL"
"ANTHROPIC_DEFAULT_SONNET_MODEL"
"ANTHROPIC_REASONING_EFFORT"
```

**Distribution is entirely under `internal/`.** Measured at the same commit:

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' pkg/ cmd/ --include='*.go' --exclude='*_test.go' | wc -l
# observed: 0
find pkg cmd -name '*.go' -not -name '*_test.go' | wc -l
# observed: 7
```

`pkg/` and `cmd/` hold **zero** bare literals today across their 7 production Go
files. REQ-EAS-005 nonetheless claims those roots as guarded surface, so their
coverage cannot be demonstrated by the natural corpus - there is nothing there to
find. It is demonstrated instead by **deliberate plants** at M6 -- one under `cmd/`
(acceptance.md AC-EAS-012 step (d)) and one under `pkg/` (step (f)). One plant per
root, because plan.md F/M2 item 6 enumerates the three scan roots as
independently-failing list elements: a guard whose root list silently omits or
typos `pkg/` would still pass a `cmd/`-only probe. Without both plants the `pkg/`
and `cmd/` portions of REQ-EAS-005's claimed surface would be asserted and never
proven.

`internal/config/envkeys.go` currently defines exactly **6** ANTHROPIC constants
(`EnvAnthropicBaseURL`, `EnvAnthropicAuthToken`, `EnvAnthropicDefaultHaikuModel`,
`EnvAnthropicDefaultSonnetModel`, `EnvAnthropicDefaultOpusModel`,
`EnvAnthropicReasoningEffort`). Two names in active use have **no constant at all**.

### A.2 The load-bearing finding: the existing guard cannot see most of the surface

`internal/cli/glm_env_parity_test.go:66`
`TestNoBareGLMEnvVarLiteralsInCLIProduction` is the mechanical gate that was
supposed to prevent this class of drift. It carries **two independent defects**
that make it structurally incapable of covering the ANTHROPIC family:

1. **Wrong banned list.** Its `banned` slice contains only three `CLAUDE_CODE_*`
   names. No `ANTHROPIC_*` name is checked at all.

2. **Wrong walk root.** Its walk root is `os.Getwd()`, which under `go test`
   resolves to the **package directory** `internal/cli`, not the repository
   root. Every occurrence outside `internal/cli/` is therefore structurally
   unreachable.

Measured reachability split at `76d9a8f3b`:

```bash
grep -rn '"ANTHROPIC_[A-Z_]*"' internal/cli/ --include='*.go' --exclude='*_test.go' | wc -l
# observed: 46   (reachable by the current walk root)

grep -rn '"ANTHROPIC_[A-Z_]*"' internal/ pkg/ cmd/ --include='*.go' --exclude='*_test.go' \
  | grep -v '^internal/config/envkeys.go' \
  | grep -v '^internal/cli/' | wc -l
# observed: 37   (structurally unreachable)
```

46 of 83 (55.4 percent) are reachable; **37 are not**. Naively reusing the
existing pattern would therefore produce a guard that silently passes on 45
percent of the surface. This SPEC treats the guard repair (walk root AND banned
list AND a demonstrated falsification) as a first-class deliverable, not a
by-product.

### A.3 New hazard introduced by widening the walk root

A repo-root-scoped walk reaches `internal/config/envkeys.go`, which by design
contains 6 bare `"ANTHROPIC_*"` literals - they are the constant definitions,
the SSOT itself:

```bash
grep -c '"ANTHROPIC_[A-Z_]*"' internal/config/envkeys.go
# observed: 6
```

The current `internal/cli`-rooted guard never faced this because `envkeys.go`
lives outside its walk root. A repo-root guard MUST exclude the SSOT definition
file, or it flags the very constants it exists to enforce. This is a new
requirement created by the widening, not an inherited one.

### A.4 Why now

This is mechanical name-indirection debt. It carries no user-visible symptom, but
it is the same drift class that `SPEC-CLIFIX-HYGIENE-001` closed for the
`CLAUDE_CODE_*` family, and the guard that closed it is the one shown above to be
scope-blind. Leaving the ANTHROPIC family unguarded means the next env-var rename
silently diverges across `internal/cli`, `internal/hook`, `internal/tmux`,
`internal/statusline`, and `internal/sandbox`.

### A.5 The census is a point-in-time snapshot - re-measurement is a run-phase gate

Every count in this SPEC (83 / 46 / 37 / 9 / 6 / 291 / 12) and every acceptance
baseline in `acceptance.md` was measured at `76d9a8f3b`. Any main-ward commit
touching the 10 files in the A.1 table, `internal/config/envkeys.go`, or the
guard-precedent files invalidates those numbers and, with them, the M3-M5
milestone arithmetic.

**[HARD] Run-phase entry precondition.** Before starting M1, the implementer MUST
re-run the A.1 census commands (the 83 / 46 / 37 totals and the 9-value unique
inventory) against the run-phase HEAD and compare them to the values recorded
here. **If any count differs, halt**: do not proceed into M1. Re-baseline the
affected ACs in `acceptance.md`, record the new numbers and the base commit in the
HISTORY table, and only then resume. Proceeding on a stale census would make every
downstream AC verdict unattributable to an observed baseline
(`.claude/rules/moai/core/verification-claim-integrity.md` section 2).

The exact commands and their expected values are enumerated as pre-flight items in
`plan.md` section C.

---

## B. Requirements (GEARS)

### B.1 SSOT constants

**REQ-EAS-001** (Ubiquitous)
The `internal/config/envkeys.go` file shall be the single definition site for
every `ANTHROPIC_*` environment-variable name referenced by production Go source
under `internal/`, `pkg/`, and `cmd/`.

**REQ-EAS-002** (Ubiquitous)
The `envkeys.go` file shall define a named constant for `ANTHROPIC_API_KEY` and a
named constant for `ANTHROPIC_DEFAULT_FABLE_MODEL`, following the existing
`EnvAnthropic<CamelCase>` naming pattern with Go initialism casing.

**REQ-EAS-003** (Ubiquitous)
The `envkeys.go` file shall define a constant for the bare namespace prefix
`ANTHROPIC_`, documented as a prefix, distinct from any full variable name.

**REQ-EAS-004** (Unwanted)
Production Go source under `internal/`, `pkg/`, and `cmd/`, excluding
`internal/config/envkeys.go` and excluding `_test.go` files, shall not contain a
bare `ANTHROPIC_*` string literal.

### B.2 Mechanical guard

**REQ-EAS-005** (Ubiquitous)
The guard test shall resolve its walk root to the repository root, so that every
production Go file under `internal/`, `pkg/`, and `cmd/` is reachable, including
files outside `internal/cli/`.

> Coverage-proof note: `pkg/` and `cmd/` hold zero bare literals today (A.1), so
> the natural corpus cannot demonstrate their reachability. AC-EAS-012 step (d)
> plants one literal under `cmd/moai/main.go` and observes RED, which is the only
> evidence this clause is satisfied rather than merely asserted.

**REQ-EAS-006** (Ubiquitous)
The guard test shall exclude `internal/config/envkeys.go` (the SSOT definition
site) and all `_test.go` files from its scan.

**REQ-EAS-007** (Event-driven)
**When** a bare `ANTHROPIC_*` string literal is introduced into any production Go
file reachable from the repository root, including a package outside
`internal/cli/`, the guard test shall fail and name the offending file and
position.

**REQ-EAS-008** (Ubiquitous)
The guard test's banned set shall cover every `ANTHROPIC_*` name defined in
`envkeys.go`, so that adding a constant without adding it to the guard is a
visible omission rather than a silent gap.

### B.3 Behaviour preservation (non-functional)

**REQ-EAS-009** (Ubiquitous)
The transition shall be a mechanical name-indirection refactor with no intended
change to runtime behaviour: every constant's value shall be byte-identical to
the literal it replaces.

**REQ-EAS-010** (Unwanted)
The change shall not modify any file under `internal/template/templates/`
(template neutrality, CLAUDE.local.md sections 15 and 25).

**REQ-EAS-011** (Unwanted)
The change shall not modify `ANTHROPIC_*` literals inside `_test.go` files, which
are a declared hardcoding-allowed zone (CLAUDE.local.md section 14).

---

## C. Success Criteria

| # | Criterion | Baseline (observed at `76d9a8f3b`) | Target |
|---|-----------|-----------------------------------:|-------:|
| 1 | ANTHROPIC constants in `envkeys.go` | 6 | 9 |
| 2 | Bare literals in production source (excl. `envkeys.go`, excl. `_test.go`) | 83 | 0 |
| 3 | Guard walk root reaches outside `internal/cli` | no (37 unreachable) | yes (0 unreachable) |
| 4 | Guard falsifiable by deliberate re-introduction outside `internal/cli` | n/a (guard absent) | proven RED then GREEN |
| 5 | `go test ./...` | not observed at base (see acceptance.md AC-EAS-013 skip note) | green |
| 6 | ANTHROPIC refs under `internal/template/templates/` | 12 | unchanged (12) |
| 7 | `ANTHROPIC_*` occurrences in `_test.go` files | 291 | unchanged (291) |

---

## D. Exclusions

### Out of Scope - the CLAUDE_CODE_* family

- Widening `TestNoBareGLMEnvVarLiteralsInCLIProduction` to repo-root scope for the
  three `CLAUDE_CODE_*` names it already guards. Those names remain unguarded
  outside `internal/cli/` after this SPEC lands. This is a known, recorded
  residual gap and a candidate follow-up SPEC; it is deliberately excluded here
  to keep this SPEC's blast radius to one env-var family.
- Auditing any other env-var family (`MOAI_*`, `GLM_*`, `CLAUDE_*`) for the same
  drift class.

### Out of Scope - test files

- `_test.go` files contain 291 `ANTHROPIC_*` occurrences. These are **not**
  migrated and **not** guarded. CLAUDE.local.md section 14 declares `_test.go` an
  allowed-hardcoding zone. A future reader MUST NOT treat these 291 occurrences
  as residual debt from this SPEC; their exclusion is intentional and recorded
  here so it is not "fixed" later by mistake.

### Out of Scope - template surface

- `internal/template/templates/` carries 12 `ANTHROPIC_` references. These are
  distributed user-facing template assets, governed by template neutrality
  (CLAUDE.local.md sections 15 and 25) and outside the Go compilation surface.
  They are untouched.

### Out of Scope - behaviour and configuration

- Any change to which environment variables are read, written, cleared, or
  injected; any change to their values, precedence, or lifecycle. This SPEC
  changes only how the **names** are spelled in Go source.
- Adding, removing, or renaming any `ANTHROPIC_*` environment variable.
- `make build` and template regeneration: no template file changes, so no rebuild
  is entailed by this SPEC.

---

## E. Open Questions

Recorded with a recommendation each; resolved at the Implementation Kickoff
Approval gate, not silently during run-phase.

### OQ-1 - the bare `ANTHROPIC_` prefix literal

**Finding.** Exactly one production usage, at
`internal/cli/worktree/tmux_integration.go:238`:

```go
// Only include ANTHROPIC_* vars
if strings.HasPrefix(key, "ANTHROPIC_") {
    cfg.GLMEnvVars[key] = value
}
```

It filters `.env.glm` entries down to the ANTHROPIC namespace. It is a **prefix**,
semantically distinct from a variable name.

**Recommendation.** Define `EnvAnthropicPrefix = "ANTHROPIC_"` in `envkeys.go`
with a doc comment stating it is a namespace prefix, not a variable name, and
route the call site through it. Rationale: it is the same SSOT concern (a rename
of the namespace must land in one place), and leaving it bare would require an
awkward carve-out in the guard.

**Guard interaction, verified safe.** The guard matches on **exact literal
equality** (`bannedSet[val]`), not substring containment, so banning the exact
string `"ANTHROPIC_"` does not collide with `"ANTHROPIC_API_KEY"` or any other
member. Confirmed by reading `glm_env_parity_test.go:103-107`.

### OQ-2 - `_test.go` exclusion

**Finding.** 291 `ANTHROPIC_*` occurrences live in `_test.go` files.
CLAUDE.local.md section 14 declares `_test.go` a hardcoding-allowed zone, and the
existing guard already skips `_test.go`.

**Recommendation.** Exclude test files from **both** the transition and the
guard, and state the exclusion explicitly in section D (done above) so a later
reader does not interpret the 291 as leftover debt. No change required to the
existing suffix check.

### OQ-3 - naming of the new constants

**Recommendation.** Follow the existing `EnvAnthropic<CamelCase>` pattern with Go
initialism casing (consistent with the existing `EnvAnthropicBaseURL`, which uses
`URL` not `Url`):

| Literal | Constant |
|---------|----------|
| `ANTHROPIC_API_KEY` | `EnvAnthropicAPIKey` |
| `ANTHROPIC_DEFAULT_FABLE_MODEL` | `EnvAnthropicDefaultFableModel` |
| `ANTHROPIC_` | `EnvAnthropicPrefix` |

### OQ-4 - extend the existing guard, or add a new one?

**Recommendation: add a NEW repo-root-scoped guard; leave the existing test
intact.**

Rationale:

1. **The existing name would become a lie.**
   `TestNoBareGLMEnvVarLiteralsInCLIProduction` asserts CLI scope in its
   identifier. Widening it to repo-root scope without renaming misleads every
   future reader; renaming it churns a test owned by a different, already-closed
   SPEC.
2. **Different owners, different families.** The existing test guards three
   `CLAUDE_CODE_*` names for `SPEC-CLIFIX-HYGIENE-001`. Its narrow scope is
   recorded behaviour, not an accident to be silently overwritten.
3. **Clean falsification target.** A separate test gives this SPEC an independent
   RED/GREEN signal that cannot be confounded with the CLAUDE_CODE_* assertions.

**Proposed home and name.** `internal/config/anthropic_env_ssot_test.go`,
`TestNoBareAnthropicEnvVarLiteralsInProduction`. Hosting it in `internal/config`
places the guard beside the SSOT it enforces.

**Consequence to accept.** The three `CLAUDE_CODE_*` names remain unguarded
outside `internal/cli/`. Recorded in section D as an explicit out-of-scope
residual gap and a follow-up candidate.

---
id: SPEC-FALSE-ALLCLEAR-GUARD-001
title: "Block two false all-clear signals: scope slog suppression to the moai hook path, and make ast-grep report an unavailable scanner"
version: "0.3.0"
status: in-progress
created: 2026-07-27
updated: 2026-07-27
author: manager-spec
priority: P0
phase: "v3.0.0-rc-stabilization"
module: "internal/cli, internal/astgrep, internal/hook/quality, internal/hook"
lifecycle: spec-anchored
tags: "false-all-clear, slog-suppression, ast-grep, sentinel-error, quality-gate, doctor, observability"
tier: M
era: V3R6
related_specs: [SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001]
---

# SPEC-FALSE-ALLCLEAR-GUARD-001 — False All-Clear Guard (slog scoping + ast-grep scanner sentinel)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-27 | manager-spec | Initial plan-phase authoring. Scope = census P1-C (slog scoping) + P1-A (ast-grep sentinel), merged into one SPEC. |
| 0.2.0 | 2026-07-27 | manager-spec | Plan-audit iteration 1 delta-fix (FAIL 0.66 → revision; Testability 0.45 was the driver). `acceptance.md` rewritten with a measured baseline per criterion. spec.md changes: GEARS label drift corrected (REQ-FAG-004 → capability gate; 7× "Event-detected" → "Event-driven"); §A.3 sequencing rationale restated to the real dependency direction; REQ-FAG-019 restated to the frame where it is observable; REQ-FAG-030 narrowed to the three constructible revert round trips and re-traced. plan.md changes: G-6 retracted as a false premise, R-8 mitigation corrected, production file count 8 → 9. |
| 0.3.0 | 2026-07-27 | manager-spec | Plan-audit iteration 2 delta-fix (CONDITIONAL PASS 0.82; 2 blocking + 2 minor conditions). N1: AC-FAG-018's `gofmt` half re-scoped to this SPEC's explicit file list — the repo-wide expectation of 0 was unsatisfiable against a measured 107, and `root.go` (an M1 file) is one of them; the 106 others are recorded as an out-of-scope observation in §C. N2: AC-FAG-006's discriminator replaced — the old round trip depended on a conditional gopls warn that does not fire on a healthy host; it now asserts the `deps` global directly (host-independent). N3: all three falsification round trips moved to detached probe worktrees (Shape 8). N4: REQ-FAG-032 added to back AC-FAG-012's named-constant prescription. |

> Frontmatter note: `related_specs` is accepted by `moai spec lint` (no findings) but is **not** listed in the Optional Fields table of `.claude/rules/moai/development/spec-frontmatter-schema.md`. It is used here as a non-blocking reference with no schema-defined semantics — a documentation gap in the schema SSOT, not a validity failure of this SPEC.

## §A Context and Problem

A codebase census recorded at `.moai/reports/census-2026-07-27-handoff.md` identified 13 **false all-clear** defects — surfaces that report success while not functioning. This SPEC covers exactly two of them, and merges them into a single SPEC because one is the verification precondition of the other.

### A.1 Defect 1 (census SLOG-01, recommended unit P1-C) — global slog suppression

`internal/cli/deps.go` (`InitDependencies`) and `internal/cli/root.go` (`initLightDeps`) each install `slog.New(slog.NewTextHandler(io.Discard, nil))` as the process-wide default logger, unconditionally, for **every** `moai` subcommand. Measured callsite counts across `internal/` + `cmd/`, excluding `_test.go`:

| Level | Callsites |
|-------|----------:|
| `slog.Warn` | 198 |
| `slog.Error` | 20 |
| `slog.Info` | 78 |
| `slog.Debug` | 84 |

All 218 warn-and-above callsites are silent on every code path. The in-code justification is a comment on `deps.go`: *"prevent slog output from hook handlers leaking to stderr"*. That rationale covers the `moai hook <event>` path — where stdout carries the hook's structured JSON contract and stderr is consumed by the Claude Code runtime — but it does not cover `moai update`, `moai doctor`, `moai ast-grep`, or any other operator-facing subcommand.

`MOAI_LOG_LEVEL` (constant `EnvLogLevel`, `internal/config/envkeys.go`) is read into `cfg.System.LogLevel` at `internal/config/manager.go` and validated at `internal/config/validation.go`, but a repo-wide search finds **no code that wires `cfg.System.LogLevel` to any slog handler**. The env var is documented and accepted, and changes nothing. That is itself a second false signal: a user who sets `MOAI_LOG_LEVEL=debug` to diagnose a problem receives the same silence.

### A.2 Defect 2 (census C-1, recommended unit P1-A) — ast-grep reports a clean scan when the scanner is absent

`internal/astgrep/scanner.go` `Scan()` returns `([]Finding{}, nil)` when `isSGAvailable()` is false. An empty finding slice with a nil error is **indistinguishable from a clean scan**. The accompanying `slog.Warn` is swallowed by Defect 1, so the user observes zero bytes on stderr.

The empty slice reaches both production callers of `Scanner.Scan`:

- `internal/cli/astgrep.go` — the `moai ast-grep` CLI prints `no findings` and exits 0.
- `internal/hook/quality/astgrep_gate.go` — `RunAstGrepGateV2` returns `(true, "")`; the commit-time PreToolUse quality gate passes silently.

A user who runs `moai init` receives ast-grep rule files under `.moai/config/astgrep-rules/`, but receives no `sg` install guidance from the README, from `moai doctor`, or from the CLI itself. A CI pipeline gating on `moai ast-grep ./` passes green with unscanned code.

The correctly-shaped sibling already exists in this repository: `internal/cli/astedit.go` checks `IsSGAvailable` and prints `ast-grep (sg) is not installed; nothing to apply.` before returning.

### A.3 Why one SPEC, and why M1 precedes M2

The two defects share a surface (`Scanner.Scan`'s skip path and the `slog.Warn` inside it) and one is the other's verification lever, so they are sequenced as M1 → M2 inside one SPEC rather than run as two independent SPECs.

The ordering rationale is stated precisely, because a looser version of it is not supported by this SPEC's own criteria. M2 is **directly verifiable without M1**: REQ-FAG-010..013 are exercised by Go tests independent of the logging path, and the CLI's stderr guidance (REQ-FAG-014) is a new explicit write, not a `slog` record. The real dependency runs the other direction — the M1 falsification uses the *pre-existing* `slog.Warn` at `internal/astgrep/scanner.go:236` as its observable, and that observable is superseded once M2 changes what the CLI writes to stderr. M1 first is what preserves the window in which that falsification is cleanly evaluable.

### A.4 Reachability hazard (discovered during plan-phase code reading, not present in the census)

The census describes the gate fix as a change to `astgrep_gate.go`. Reading the call chain shows that is necessary but **not sufficient**. `RunAstGrepGateV2` returns `(bool, string)`, and both downstream frames discard the string on the pass path:

- `internal/hook/quality/gate.go` — `if ok, out := RunAstGrepGateV2(...); !ok { return false, out }`; when `ok` is true, `out` is dropped and `Run` later returns `(true, "")`.
- `internal/hook/pre_tool.go` — `passed, output := gate.Run(ctx)`; when `passed` is true, `output` is dropped and no hook output field carries it.

A reason string returned by `astgrep_gate.go` alone is therefore **inert**. Surfacing it requires propagating a pass-path reason through `QualityGate.Run` and emitting it on the PreToolUse allow path.

The emission channel must be `HookOutput.SystemMessage` (an existing field, already used by `post_tool.go`, `auto_update.go`, and `post_compact.go`), **not** `slog`: the `moai hook` path retains `io.Discard` under this SPEC's own M1 decision, so a `slog.Warn` added in `pre_tool.go` would be silent by construction.

## §B Requirements (GEARS)

### B.1 M1 — Logging scope (Defect 1)

- **REQ-FAG-001** — Ubiquitous. The CLI shall install exactly one process-wide default `slog` handler, chosen at a single decision site, for every `moai` invocation.
- **REQ-FAG-002** — Capability gate. **Where** the invoked subcommand is `hook`, the CLI shall install a handler that discards all records, preserving the current hook-path behavior.
- **REQ-FAG-003** — Capability gate. **Where** the invoked subcommand is not `hook`, the CLI shall install a handler that writes records to stderr.
- **REQ-FAG-004** — Capability gate. **Where** the invoked subcommand is not `hook` and `MOAI_LOG_LEVEL` is unset, the handler shall admit records at `warn` and above and shall suppress records below `warn`.
- **REQ-FAG-005** — State-driven. **While** `MOAI_LOG_LEVEL` holds a recognized level name, the non-`hook` handler shall admit records at that level and above, whether that level is above or below the `warn` default.
- **REQ-FAG-006** — Event-driven. **When** `MOAI_LOG_LEVEL` holds a value that is not a recognized level name, the CLI shall fall back to the `warn` default rather than failing the invocation.
- **REQ-FAG-007** — Ubiquitous. The CLI shall read the `MOAI_LOG_LEVEL` variable name from the existing `internal/config/envkeys.go` constant and shall not restate the literal string at the handler-construction site.
- **REQ-FAG-008** — Unwanted. The CLI shall not write `slog` records to stdout on any path, so that stdout remains reserved for machine-readable output.
- **REQ-FAG-009** — Ubiquitous. The trivial-subcommand fast path (`--version`, `version`, `-v`, `help`, `--help`, `-h`, `completion`) shall continue to avoid full dependency initialization after this change.

### B.2 M2 — ast-grep scanner sentinel (Defect 2)

- **REQ-FAG-010** — Ubiquitous. The `internal/astgrep` package shall export a sentinel error value that identifies "the `sg` scanner binary is unavailable" as a condition distinct from both a clean scan and a scan failure.
- **REQ-FAG-011** — Event-driven. **When** `Scanner.Scan` detects that the `sg` binary is not resolvable, it shall return an error that satisfies `errors.Is` against the sentinel, instead of returning a nil error.
- **REQ-FAG-012** — Ubiquitous. The error returned under REQ-FAG-011 shall name the binary that was sought and shall carry install guidance.
- **REQ-FAG-013** — Unwanted. `Scanner.Scan` shall not report a scanner-unavailable condition as an empty-and-successful scan result.

### B.3 M2 — `moai ast-grep` CLI behavior (Defect 2, detector surface)

- **REQ-FAG-014** — Event-driven. **When** `moai ast-grep` receives the scanner-unavailable sentinel from `Scan`, it shall write an install-guidance message to stderr and shall terminate with a non-zero exit status.
- **REQ-FAG-015** — Unwanted. `moai ast-grep` shall not print its no-findings result line when the scan did not run.
- **REQ-FAG-016** — Ubiquitous. The scanner-unavailable message shall be written to stderr only, leaving stdout free of the message under every `--format` value.
- **REQ-FAG-017** — Ubiquitous. `moai ast-edit` shall retain its current behavior of printing a notice and terminating with a zero exit status when `sg` is unavailable, because it is a mutator for which "nothing to apply" is a genuine no-op, whereas `ast-grep` is a detector used as a gate.

### B.4 M2 — PreToolUse quality gate (Defect 2, gate surface)

- **REQ-FAG-018** — Event-driven. **When** the ast-grep quality-gate step receives the scanner-unavailable sentinel, it shall report the step as passing and shall accompany that pass with a non-empty reason naming the skip.
- **REQ-FAG-019** — Ubiquitous. The ast-grep quality-gate step shall select its pass-path reason from two distinct, non-empty values — one naming a scanner-unavailable skip, one naming a degraded scan — and shall never emit an empty reason on a pass that followed a scan error. (Scope note: the gate constructs its scanner with a fixed binary name, so only the scanner-unavailable branch is reachable end-to-end through the gate; the two-class *classification* is verified at the `Scanner` frame, where both are constructible. See `acceptance.md` AC-FAG-012 for the split and its explicitly-uncovered residue.)
- **REQ-FAG-020** — Unwanted. The ast-grep quality-gate step shall not deny, block, or fail a commit because `sg` is absent.
- **REQ-FAG-021** — Ubiquitous. The quality gate shall propagate a pass-path reason from the ast-grep step through its own return value, so that the reason reaches the gate's caller rather than being discarded at the pass branch.
- **REQ-FAG-022** — Event-driven. **When** the PreToolUse handler completes a passing quality-gate run that carried a non-empty reason, it shall emit that reason on the hook output's system-message channel.
- **REQ-FAG-023** — Ubiquitous. The reason emitted under REQ-FAG-022 shall reach the caller through the hook's structured output rather than through the logging subsystem, because the `hook` path discards log records under REQ-FAG-002.
- **REQ-FAG-032** — Ubiquitous. The two pass-path reason values named in REQ-FAG-019 shall be package-level named constants rather than inline string literals, so that their distinctness is assertable without exercising a code path that is unreachable through the gate. This is a deliberate testability constraint on the implementation shape, recorded as a requirement rather than left as an acceptance-criterion-only prescription; its cost is one named constant per reason, and its benefit is that the collapse-to-one-string failure is mechanically detectable.

### B.5 M2 — Diagnostics and documentation (Defect 2, discoverability)

- **REQ-FAG-024** — Ubiquitous. `moai doctor` shall report whether the `sg` binary is resolvable, as one check among its existing check set.
- **REQ-FAG-025** — Event-driven. **When** `moai doctor` finds `sg` unresolvable, the check shall report a non-OK status and shall carry install guidance in its message.
- **REQ-FAG-026** — Ubiquitous. The `sg` check shall follow the structure of the existing optional-external-binary check in the same file, rather than introducing a new check shape.
- **REQ-FAG-027** — Unwanted. The `sg` check shall not report a failing status that would misrepresent an optional tool as a broken installation.
- **REQ-FAG-028** — Ubiquitous. The published CLI reference page for `ast-grep` / `ast-edit` shall describe the two commands' differing behavior when `sg` is absent, and shall carry install guidance.
- **REQ-FAG-029** — Ubiquitous. The documentation change under REQ-FAG-028 shall be applied to all four published locales in the same change set.

### B.6 Cross-cutting

- **REQ-FAG-030** — Ubiquitous. The three behaviors named here shall each be demonstrated by an executed revert-and-observe round trip, in which removing the production change flips the observation: (a) the logging-scope change (the scanner warning becomes visible, and a binary built from the base commit shows zero stderr bytes); (b) the trivial-command fast path (deleting the branch makes `--version` emit warn records); (c) the gate reason's three-frame propagation (both the call-deleted and the body-neutered form make the reachability test fail). Scope note: this requirement is deliberately narrowed to these three, because they are the behaviors for which a revert round trip is actually constructible — a blanket "every behavior" claim would trace to criteria that cannot observe revert-behaviour.
- **REQ-FAG-031** — Ubiquitous. The repository's existing test suite shall pass after this SPEC's changes, with no test newly skipped as a means of accommodating the changes.

## §C Scope Exclusions

This section is the SPEC's exclusions section: what this SPEC does **not** build.

### Out of Scope — the other eleven false all-clear defects

- The census enumerated 13 false all-clear defects. This SPEC covers exactly two (its rows 1 and the SLOG-01 entry). The remaining eleven — including the PreToolUse 5-second timeout that destroys a `deny` verdict, the `moai agent lint` roster gap, the `@MX` sidecar write path, the `moai doctor` Skills Allowlist and MCP Scope checks, `moai constitution guard`, unknown-subcommand exit codes, and the shell-gate `t.Skipf` tests — are out of scope and remain open.
- In particular the PreToolUse hook-timeout defect is **not** addressed here. This SPEC's gate change makes an absent scanner visible; it does not make the gate's verdict survive the hook budget.

### Out of Scope — `MOAI_LOG_FORMAT` wiring

- `MOAI_LOG_FORMAT` (constant `EnvLogFormat`) is read into `cfg.System.LogFormat` and is equally unwired to any handler. It is an instance of the same defect class and is deliberately excluded: the decision recorded for this SPEC names `MOAI_LOG_LEVEL` only. Wiring the format selector is a candidate follow-up, not a requirement here.
- **Residual risk, sharpened**: the consequence is worse than "stays inert". Today every record is discarded, so the knob's inertness is **invisible** — nobody can tell it does nothing. After M1 the non-hook handler is a fixed text handler, so a user setting `MOAI_LOG_FORMAT=json` will see text output and may reasonably conclude the setting applied. This SPEC therefore converts an invisible inert knob into a **visibly misleading** one. Accepted as scoped-out, recorded so the follow-up is not forgotten.

### Out of Scope — hook-path log observability

- Under REQ-FAG-002 the `moai hook` path discards log records unconditionally; `MOAI_LOG_LEVEL` does not re-enable them there. Providing a supported way to observe hook-handler logs is a separate concern.

### Out of Scope — `moai doctor` exit-code semantics

- `moai doctor` currently returns a nil error regardless of how many checks report a failing status, so the process exits 0 even when the rendered output shows failures. This SPEC adds a check to that command but does not change its exit-code behavior; that repair belongs to the census's separate diagnostics unit.

### Out of Scope — shipping or vendoring the `sg` binary

- `sg` is an external tool. This SPEC makes its absence observable and actionable; it does not install, bundle, download, or vendor it, and it does not add a package-manager dependency.

### Out of Scope — ast-grep rule content

- The rules under `.moai/config/astgrep-rules/` are untouched. Neither the shipped rule set nor the local dogfood rule tree is added to, removed from, or rewritten by this SPEC.

### Out of Scope — the 106 pre-existing `gofmt`-unclean files

- `gofmt -l internal cmd` reports **107** unclean files on the unmodified tree. Exactly one of them — `internal/cli/root.go` — is a file this SPEC edits, and bringing that one to clean is in scope. The other **106** are pre-existing debt and MUST NOT be reformatted as part of this SPEC's diff.
- Root cause recorded for the follow-up: the repository's actual formatter is `gofumpt` (`Makefile:61`), and `.github/workflows/` contains no format guard at all, so the debt accumulated silently. `gofumpt` is not installed on the authoring host, so no acceptance criterion here depends on it.
- Adding a CI format guard and cleaning the 106 files is a separate, mechanical follow-up — deliberately not bundled here, because a 106-file reformat would swamp this SPEC's reviewable diff.

### Out of Scope — the remaining ~380 slog callsites' content

- This SPEC changes which records are *emitted*, not what any individual callsite says. No `slog.Warn` / `Info` / `Debug` call is added, removed, re-worded, or re-levelled as part of this work.

## §D Acceptance Criteria

Acceptance criteria are enumerated in `acceptance.md` (Tier M artifact set). Each criterion is a runnable command paired with its expected observation, and each states what would have to break for it to fail.

## §E Traceability

| Requirement group | Milestone | Primary surfaces |
|-------------------|-----------|------------------|
| REQ-FAG-001 .. 009 | M1 | `internal/cli/root.go`, `internal/cli/deps.go` |
| REQ-FAG-010 .. 013 | M2 | `internal/astgrep/scanner.go` |
| REQ-FAG-014 .. 017 | M2 | `internal/cli/astgrep.go` (`internal/cli/astedit.go` unchanged, asserted) |
| REQ-FAG-018 .. 023, 032 | M2 | `internal/hook/quality/astgrep_gate.go`, `internal/hook/quality/gate.go`, `internal/hook/pre_tool.go` (REQ-FAG-032 → AC-FAG-012 half 2) |
| REQ-FAG-024 .. 029 | M2 | `internal/cli/doctor.go`, `docs-site/content/{en,ko,ja,zh}/cli-reference/ast-grep.md` |
| REQ-FAG-030 | M1 + M2 | falsification round trips — AC-FAG-004 (logging scope), AC-FAG-006 (trivial fast path), AC-FAG-011 Form A + Form B (three-frame propagation) |
| REQ-FAG-031 | M1 + M2 | AC-FAG-017 (suite green, skip count not increased) |

> Traceability note (plan-audit iter-1 D11): REQ-FAG-030 previously traced only to AC-FAG-018, whose commands (`go vet`, Windows cross-build, `gofmt`) cannot observe revert-behaviour — a row that looked complete and verified nothing. It now traces to the three criteria that actually execute revert round trips, and REQ-FAG-030's own wording is narrowed to match (§B.6).

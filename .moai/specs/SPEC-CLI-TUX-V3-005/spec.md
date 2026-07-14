---
id: SPEC-CLI-TUX-V3-005
title: "AC-TUX3-020 Printer Migration — fmt.Print* Ratchet Succession"
version: "0.1.0"
status: draft
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, printer, ratchet, debt, tux-v3"
depends_on: [SPEC-CLI-TUX-V3-001]
tier: M
era: V3R6
---

# SPEC-CLI-TUX-V3-005 — AC-TUX3-020 Printer Migration

## HISTORY

| Date | Version | Author | Summary |
|------|---------|--------|---------|
| 2026-07-14 | 0.1.0 | manager-spec | Plan-phase artifact creation. Debt split from SPEC-CLI-TUX-V3-003 (amendment re-close). AC-TUX3-020 ratchet succession. |

**Provenance**: AC-TUX3-020 was originally REQ-TUX3-020 in SPEC-CLI-TUX-V3-003 (spec.md:98). During SPEC-003's amendment re-close (sync bc20d5e2d), the Printer-migration scope was explicitly excluded from the amendment and split to this standalone SPEC. The original REQ wording is carried verbatim; the acceptance criterion (grep-count ratchet) is inherited unchanged.

**Dependency chain**: SPEC-CLI-TUX-V3-001 (completed) introduced the `internal/cli/printer` package with @MX:DEBT annotations marking the migration trigger. SPEC-CLI-TUX-V3-002 (ModeTTY spinner animation) and SPEC-CLI-TUX-V3-003 (5-subpackage decomposition + radio cleanup) are siblings. This SPEC fulfills the @MX:UPGRADE trigger recorded in SPEC-001.

---

## §A. User Story

**As a** moai-adk-go maintainer,
**I want** all direct `fmt.Print*` calls in non-test `internal/cli` sources migrated to the existing `printer.Printer` interface,
**so that** CLI output flows through a single gateway with consistent channel discipline (stdout=data, stderr=status), render-mode awareness (TTY/Plain/JSON), and NO_COLOR/reduced-motion support.

---

## §B. Requirements (GEARS)

### REQ-TUX3-020 — Ratchet succession (inherited from SPEC-003)

**Where** the pre-flight baseline of direct `fmt.Printf`/`fmt.Println`/`fmt.Print(` calls in non-test `internal/cli` sources is 38 (measured 2026-07-14), the count of such calls **shall** be strictly lower than the re-measured baseline as update-surface call sites migrate to the Printer (ratchet succession — REQ-CTX-017).

### REQ-TUX3-021 — Printer-method usage at migrated sites

**Where** a call site in `internal/cli` previously used `fmt.Print*` for status, informational, or error output, the migrated code **shall** route through the appropriate `printer.Printer` method (`Info`, `Warn`, `Error`, `Success`, `Data`, `Step`, `Spinner`, `Progress`) as determined by the output classification in research.md §C.

### REQ-TUX3-022 — Behavior preservation (characterization)

**While** migrating `fmt.Print*` call sites to Printer methods, the observable output (byte content + channel routing: stdout for data, stderr for status) **shall** be preserved as characterized by pre-migration characterization tests, with the sole intentional change being channel re-routing from stdout-mixed to the Printer's channel-disciplined output (stdout=data, stderr=status).

### REQ-TUX3-023 — Printer package coverage floor

The `internal/cli/printer` package **shall** maintain test coverage ≥ 85.0% across the migration, and each migrated file (`state.go`, `migration.go`, `worktree/tmux_integration.go`, and any others in scope per the M1 architecture decision) **shall** maintain or improve its pre-migration coverage.

### REQ-TUX3-024 — Dead-code and gap-site disposition

**When** a call site is classified as a gap (interactive prompt, TUI render — see research.md §D), the SPEC **shall** record an explicit disposition for each: migrate via Printer interface extension, exclude as out-of-scope, or defer to a follow-up SPEC. No gap site **shall** be silently left on `fmt.Print*` without a recorded rationale.

---

## §C. Scope

### In Scope

- Migration of `fmt.Print*` call sites in non-test `internal/cli/*.go` and `internal/cli/worktree/*.go` to `printer.Printer` methods.
- Characterization tests capturing pre-migration output for behavior-preservation verification.
- Channel-discipline correction: stdout = data (machine-consumable), stderr = status (human progress).
- Ratchet verification: the grep-count AC (strictly below baseline 38).
- Disposition recording for gap sites (interactive prompts, TUI render).

### Out of Scope — New Printer interface design

- The `printer.Printer` interface (`internal/cli/printer/printer.go`) is ALREADY DESIGNED and IMPLEMENTED (SPEC-CLI-TUX-V3-001). This SPEC migrates call sites onto the EXISTING interface; it does not redesign the interface.
- New Printer methods (e.g. for styled stdout or interactive prompts) are proposed ONLY if the M1 architecture gate determines a gap site requires interface extension; the default disposition is exclusion or deferral.

### Out of Scope — Test-file migration

- `*_test.go` files are excluded from the ratchet grep (`grep -v _test.go`) and from migration. Test files may legitimately use `fmt.Print*` for debug output or test fixtures.

### Out of Scope — Non-CLI packages

- `internal/` packages outside `internal/cli/` (e.g. `internal/hook/`, `internal/template/`, `internal/tui/`) are not scanned or migrated by this SPEC. The ratchet grep targets `internal/cli/` only.

### Out of Scope — `internal/cli/update/*` subpackages

- `internal/cli/update/`, `internal/cli/update/backup/`, `internal/cli/update/deploy/` already route output through Printer-backed reporters (SPEC-CLI-TUX-V3-003 REQ-CTX-015 `newPrinterReporter`). These subpackages are confirmed clean of direct `fmt.Print*` status calls (the `fmt.Fprintln(out, tui.CheckLine(...))` pattern in `update.go` uses `io.Writer` not bare `fmt.Print*`). Not migration targets.

---

## §D. Constraints

| Constraint | Value | Source |
|-----------|-------|--------|
| Pre-flight baseline | 38 direct `fmt.Print*` calls (non-test, `internal/cli/`) | Measured 2026-07-14 via canonical grep |
| Ratchet direction | Strictly decreasing (succession) | REQ-CTX-017 |
| Channel discipline | stdout = data, stderr = status/info/warn/error | printer.go §Channel discipline (HARD), internal/cli/CLAUDE.md |
| Subagent boundary | CLI code MUST NOT call AskUserQuestion | C-HRA-008, internal/cli/CLAUDE.md |
| Render modes | ModeTTY, ModePlain, ModeJSON | printer.go REQ-CTX-013 |
| Coverage floor | ≥ 85.0% (printer + migrated files) | TRUST 5 Tested pillar |

---

## §E. Dependencies

| SPEC | Status | Relationship |
|------|--------|-------------|
| SPEC-CLI-TUX-V3-001 | completed | **depends_on** — introduced Printer interface + @MX:DEBT trigger this SPEC fulfills |
| SPEC-CLI-TUX-V3-003 | completed | Parent — AC-TUX3-020 was split from here (debt) |
| SPEC-CLI-TUX-V3-002 | exists | Sibling — spinner animation in Printer (no blocking dependency) |
| SPEC-CLI-TUX-V3-004 | exists | Sibling — policy SPEC (no blocking dependency) |

---

## §F. Assumptions

1. The `printer.Printer` interface as declared in `internal/cli/printer/printer.go` (lines 64-87) is the STABLE migration target — no interface redesign is in scope unless the M1 architecture gate identifies an unavoidable gap.
2. The existing wiring pattern (`printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))`, observed in `clean.go:33` and `update.go:996`) is the canonical injection point for cobra `RunE` closures.
3. The ratchet AC counts the RAW grep output (including commented-out lines matched by the grep), not just executable lines. Baseline 38 includes 1 commented-out godoc example line (`worktree/new.go:433`).

---

## §G. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Channel re-routing breaks scripted consumers of stdout output | Medium | High | Characterization tests capture exact pre-migration stdout/stderr split; channel changes are intentional and documented |
| `uikit/banner.go` migration requires Printer interface extension | Medium | Medium | M1 architecture gate decides before any code migration; default = exclude as TUI render |
| `state.go` human-format output has no clean Printer method (multi-line stdout text) | Medium | Medium | Characterization tests + channel decision in research.md §C; Data() with string argument is the likely mapping |
| Baseline drift if unrelated `fmt.Print*` calls are added during migration | Low | Medium | Re-measure baseline at run-phase start; record in progress.md §E.2 |

---

## §H. Cross-References

- `internal/cli/printer/printer.go` — Printer interface (lines 64-112), the migration target
- `internal/cli/CLAUDE.md` — "Output streams" convention (stdout/stderr discipline)
- SPEC-CLI-TUX-V3-003 spec.md:98 — original REQ-TUX3-020 wording (verbatim carry-over)
- SPEC-CLI-TUX-V3-003 acceptance.md:33 — original AC-TUX3-020 grep command
- `internal/cli/clean.go:33` — canonical Printer wiring pattern for cobra commands
- `internal/cli/reporter.go:24-33` — printerReporter adapter pattern (ProgressReporter → Printer)
- REQ-CTX-017 (SPEC-CLI-TUX-V3-001) — ratchet succession original requirement

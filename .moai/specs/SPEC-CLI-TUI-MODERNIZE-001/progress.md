---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Progress — interactive TUI surface modernization"
version: "0.1.3"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: "internal/cli/update, internal/merge, internal/cli, internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, tui, tux, bubbletea, huh, theme, preview, dead-code, refactor"
tier: L
---

# Progress — SPEC-CLI-TUI-MODERNIZE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts (Tier L, 6-file set)**: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` under `.moai/specs/SPEC-CLI-TUI-MODERNIZE-001/`. `design.md` + `research.md` added at v0.1.2 (plan-audit D4 — the v0.1.1 4-file set was short of the Tier L requirement).
- **Counts**: 41 GEARS requirements (Groups A-E) / 46 acceptance criteria (Groups A-F, 16 MUST-FIX). Verified: `grep -c '^- \*\*REQ-TUIM-' spec.md` → `41`.
- **Tier**: L. **Milestones**: M1 (preview redesign, independently landable), M2 (dead confirm removal), M3 (huh theme unification).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | CLI ✓ | TUI ✓ | MODERNIZE ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, executed).
- **Vocabulary decision**: TUX / TUI / huh form runtime — user-confirmed, bound in `spec.md` §A.2.
- **Measured baselines** (plan.md §B, all executed this session): file sizes; `ConfirmMerge` production callers = 0; hex literals outside `internal/tui/` = 4 (2 real code in `internal/merge/confirm.go:454-455`, 2 comment-only); coverage `internal/cli` 74.9% / `internal/cli/update` 70.2% / `internal/merge` 86.3% / `internal/cli/wizard` 95.2% / `internal/tui` 93.6%; `GOOS=windows` build exit 0; `wizard.go` theme anchor at line 483.
- **Open clarifications**: **0**. Both prior markers were resolved by the user at v0.1.1 and recorded in plan.md §I — D-4 = HYBRID residue policy (table inline / diff alternate-screen / `esc` restores / single-line exit summary); D-5 = SPLIT verification (`NO_COLOR` structure golden + token-application unit tests, self-trip bound to the token tests).
- **Renderer verification (D-4 precondition, executed)**: the mid-run `AltScreen` toggle is first-class in bubbletea v2 — `cursed_renderer.go:320` compares `s.lastView.AltScreen != view.AltScreen` per frame and emits the DEC mode 1049 pair (`:513-525`). `esc`-restores-scrollback is a mode-1049 guarantee, so **no fallback was needed and none is recorded**. Verification additionally surfaced a load-bearing constraint: the close path (`:167-171`) discards the final frame when the last view is an alt-screen view, so a quit reachable from the diff sub-view would silently drop the exit summary → REQ-TUIM-022a + AC-TUIM-014d (MUST-FIX).
- **plan-audit iteration 1**: FAIL 0.74 (Tier L threshold 0.85) — 4 MUST-FIX, 9 SHOULD-FIX. All 13 repaired at v0.1.2; every numeric baseline and the full M2 DELETE inventory were independently reproduced by the auditor and are unchanged. Repairs were confined to the acceptance-criteria layer plus the two missing Tier L artifacts.
- **v0.1.2 repair evidence (commands executed this session)**: lipgloss v2 renders `#d97757` as `\x1b[38;2;217;119;87m`, `containsHexToken=false` (D1); escaped-`\|` ERE returns exit 1 against a positive control while the unescaped form matches (D3/S1); `grep -c $'\x1b\['` errors with `brackets ([ ]) not balanced` exit 2, corrected form returns `0` (S5); `internal/cli/update/plan/plan.go:63,64,88` confirmed as a fourth `merge.FileAnalysis` consumer with a named-field composite literal (S3); `go.mod:6` pins `bubbles/v2 v2.1.1` (S8); REQ count 41 (S9).
- **plan-audit iteration 2 (delta)**: CONDITIONAL **0.90** — Tier L threshold 0.85 cleared. 12 of 13 v0.1.2 repairs confirmed genuinely fixed; all 7 must-pass criteria PASS; AC-TUIM-030a/030c confirmed carried through byte-for-byte. Three residual findings (NEW-1 MUST-FIX, NEW-2 SHOULD-FIX, NEW-3 NICE-TO-HAVE) plus 3 cosmetics, all repaired at v0.1.3.
- **v0.1.3 NEW-1 evidence (executed before applying the fix, per the auditor's calibration note)**: combined-attribute render against pinned `charm.land/lipgloss/v2@v2.0.5` — `fgOnly = "\x1b[38;2;191;101;71mx\x1b[m"`, `bold+fg = "\x1b[1;38;2;191;101;71mx\x1b[m"`. The v0.1.2 full-CSI prefix is a substring of `fgOnly` (true) but **not** of `bold+fg` (false); the parameter run `38;2;191;101;71` is a substring of all three. The prediction was confirmed, so the fix was applied: SGR-parameter-substring method, recorded with evidence in acceptance.md §C.2.
- **Status**: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

# SPEC-HOOK-ZONE-SHELL-ESCAPE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: authored (plan-audit pending — the card's plan-audit gate promotes this to audit-ready after PASS)
plan_complete_at: 2026-10-09
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M set)

### RED evidence (the reproduction the card was gated on)

- Record: `.moai/reports/t1585/red-repro.md` — committed in the baseline-first
  commit `9dbe40c0a` (test(t1585), branch `WT-zone-gate-defects`), which also
  carries the instrument `internal/hook/protected_zone_shell_repro_test.go`.
  Ordering attribution satisfied: the RED baseline precedes every repair
  commit (verification-claim-integrity §2.3).

### Reproduced defect claims (each with its evidence line)

1. **① P1 NUL-truncation protected-path bypass** — the guard judges
   `zone_dir\x00/sub` while bash truncates the ANSI-C part at the NUL and
   removes `zone_dir` itself. Evidence (red-repro.md §Defect ①, exit 1):
   `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected directory is GONE (rm output: "")`
2. **② P2 `\x` raw-byte vs code-point confusion** — a raw-byte spelling of a
   protected path decodes to re-encoded text no zone entry matches. Evidence
   (red-repro.md §Defect ②, exit 1):
   `BYPASS — decision="allow" reason="", want deny; real bash ran the allowed command and the protected marker is GONE (rm output: "")`
3. **③ P2 no-digit hex panic + mangling** — `\x`/`\u`/`\U` with no digit
   panics (slice bounds) at end-of-string and mangles text after a non-digit.
   Evidence (red-repro.md §Defect ③, exit 1): `"\x": PANICKED — slice bounds
   out of range on the no-digit escape` (×3 prefixes) + `"\xZ": decoded
   "xZZ", want the bash literal "\xZ"` (×2 shapes).

### Decision record

decided_by: lane-19 — RED-first gate satisfied: the three defects were
UNREPRODUCED at dispatch and the RED reproduction was measured 2026-10-08 in
the card worktree (bash ground truth first, then the instrument; exit 1 on
all three defect rows), evidence `.moai/reports/t1585/red-repro.md`, pinned
baseline-first as commit `9dbe40c0a`. Run phase may not repair beyond what
the RED rows measure.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

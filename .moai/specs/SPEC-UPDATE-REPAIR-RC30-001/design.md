# SPEC-UPDATE-REPAIR-RC30-001 — Design

Design decisions behind spec.md and plan.md. Every decision names its requirement and its item.

## 1. Data model: shared roots remain the install source

- Install source: the four `$HOME` roots at `internal/userassets/paths.go:36-43` (REQ-RC30-001). Operator direction D1 keeps this model. SPEC-USER-ASSET-INSTALL-001 (completed, on main) decided it.
- Project-side copies (`.claude/skills/moai/**`, `.claude/agents/moai/**`) are not the install source. They are removed only after the profile-visibility check passes (REQ-RC30-003). Until then they stay.
- The profile directory is not written by default. Writing into it conflicts with REQ-002 of SPEC-USER-ASSET-INSTALL-001 (spec.md:255). The operator decides this (decision Q2).

## 2. Profile reach

- Reach has two candidate mechanisms: an install into the profile directory (blocked by REQ-002 unless Q2 amends it), or a runnable verification that the profile session sees the shared assets (AC-RC30-01).
- The design does not choose between them. It makes AC-RC30-01 the gate for every removal, so a wrong choice cannot delete a working project copy.

## 3. Migration ordering

- Order: back up every file the removal will delete into the run backup (REQ-RC30-004), then run the removal. The removal set is the same set the migration would delete, so no file leaves without a copy.
- The emptied skill directories are removed up to the first non-empty parent (U-10).

## 4. Config writes

- `ConfigManager.Save` must not drop keys it does not model, must keep `deployment_mode`, and must keep comments (REQ-RC30-007, U-05). A yaml-node patch is the planned route. The typed struct alone cannot hold unmodelled keys, which is why the struct is not the whole answer.
- A record-less project is never re-stamped silently (REQ-RC30-008, U-06).

## 5. Restore and retention

- Restore applies the backed-up `settings.json` and lists the paths the backup does not cover (REQ-RC30-009, U-07).
- Retention counts per run, not by directory-name shape (REQ-RC30-010). The run identifier is taken from the run's own backup record.

## 6. Output

- One renderer: `internal/cli/update_tux.go`. Inputs are measured values only (REQ-RC30-011).
- Layout order: banner once (REQ-RC30-013), step progress (REQ-RC30-014), conflicts and warnings (REQ-RC30-017), one-screen summary with one restore command (REQ-RC30-016), file lists collapsed (REQ-RC30-015).
- Glyphs: ✓, !, ✗ from the existing whitelist (REQ-TUXIU-004). Colour and motion follow REQ-TUXIU-041 and REQ-TUXIU-042.
- Fallback: `NO_COLOR`, non-TTY, `CI`, and narrow terminals use plain lines with the same counts (REQ-RC30-018).
- Thresholds are named constants in the `internal/tui` defaults file (hardcoding rule). The width threshold is decision Q7.

## 7. Version and skip

- The skip decision compares a template tree hash, not a version string (U-13). The stamp is checked against the deployed bytes (U-14).
- A user root is never refreshed by an older binary (REQ-RC30-012, U-11). Both branches touch the install code that enforces this, so the change lands after absorption (M1).

## 8. Out of design scope

- Factory slug and inbox defects (t1621, t1622): see spec.md § H.
- Init-side presentation: not in this item register; queue overlap with t1527 is recorded with disposition undecided (spec.md § F).

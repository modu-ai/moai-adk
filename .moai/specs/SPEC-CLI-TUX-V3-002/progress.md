# SPEC-CLI-TUX-V3-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §E.2 Run-phase Evidence

### M2a — huh v2 spike verdict (REQ-TUX2-005)

**huh v2 spike verdict: SUCCESS** — the YOffset scroll defect is RESOLVED under the
huh v2.0.3 + bubbletea v2 pair. M2c takes the unified multi-group form path
(REQ-TUX2-006); plan B (REQ-TUX2-007) is NOT taken.

Reproduction evidence (spike executed in `/tmp/huh-v2-spike`, isolated module
`spike`, `charm.land/huh/v2@v2.0.3` + `charm.land/bubbletea/v2@v2.0.2`):

1. **Source-level**: the v1 defect mechanism — `updateViewportHeight()` with the
   unconditional `s.viewport.YOffset = s.selected` reset (huh v1.0.0
   `field_select.go:543` and `:203`, forced into effect by the OptionsFunc
   `s.height = defaultHeight` path at `:235-236`) — is REMOVED in v2.0.3.
   Replacement: `ensureVisible()` (v2.0.3 `field_select.go:656-668`) scrolls the
   viewport "the minimum amount so that the region [offset, offset+height) is
   within the visible area" — cursor-out-of-view clamping only, never an
   unconditional snap-to-top.
2. **Reproduction-level**: a programmatic multi-group form (2 groups, 9 fields —
   input + selects + confirm, matching the wizard's 7-9 visible-question
   envelope) driven with `tea.KeyPressMsg`/`tea.WindowSizeMsg` at 80x40 AND
   80x12. After `KeyDown` inside a 3-option select (cursor high→medium), the
   option ABOVE the cursor ("high") **remains visible in the rendered frame in
   both terminal sizes**. The v1 defect (options above cursor hidden) did not
   reproduce. Frames archived at `/tmp/huh-v2-spike/frames.txt` (run exit=0).
3. **API compatibility probe** (all wizard usage patterns compile + behave):
   `TitleFunc`/`DescriptionFunc(fn, binding)` OK; `Validate`-as-save OK (saved
   answers harvested: `map[development_mode:tdd model_policy:medium
   plan_type:subscription project_name:spike]`); `Group.WithHideFunc` OK;
   `huh.ErrUserAborted` OK (v2 `form.go:55`); `WithAccessible` OK;
   `huh.NewOption` OK. **API delta**: `WithTheme` now takes a `huh.Theme`
   interface (`Theme(isDark bool) *Styles`) — adapt via `huh.ThemeFunc`; the v2
   `Styles`/`FieldStyles` field set matches the v1 `Theme.Focused/Blurred`
   fields the wizard theme sets; lipgloss v2 drops `AdaptiveColor` — the
   `isDark` parameter supplies the light/dark axis (maps directly onto
   `tui.LightTheme()`/`tui.DarkTheme()` tokens).
4. **Known v2 behavior (not the defect)**: on height-constrained terminals the
   group viewport anchors the FOCUSED FIELD to the viewport top
   (`group.go buildView()` → `SetYOffset(focused-field offset)`), so fields
   above the focused field scroll out when the page does not fit. This is
   standard scroll-into-view behavior (viewport clamps to 0 when content fits),
   not the v1 defect.

### M2a — bubbletea/bubbles v2 adoption (REQ-TUX2-012)

Pinned `charm.land/bubbletea/v2 v2.0.8` + `charm.land/bubbles/v2 v2.1.1` as
direct dependencies (I-3 animated-spinner prerequisite; import-usage lands in
M2d printer backend). huh major follows the verdict above: v2 (M2c).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

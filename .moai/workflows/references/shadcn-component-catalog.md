# shadcn/ui Component Catalog — Design-System Bundle Reference

> Bundled reference for the design-phase D2 step (design-system generation).
> Loaded on demand from `workflows/design.md` (progressive disclosure — not
> always-loaded). Canonical source: https://ui.shadcn.com/docs/components.

## Purpose

When `manager-design` bundles a design system from code and pushes it to the
Claude Design project (D2), the component set MUST cover the full shadcn/ui
catalog below. A partial bundle is a D2 defect: Claude Design cannot render a
screen for a component whose tokens were never imported, and the gap surfaces
only after the user reaches for that component on the canvas — late, expensive,
and avoidable.

## Full component set (mandatory)

The design-system bundle MUST include every component in this list. Names match
the shadcn/ui registry.

| | | | |
|---|---|---|---|
| Accordion | Alert | Alert Dialog | Aspect Ratio |
| Attachment | Avatar | Badge | Breadcrumb |
| Bubble | Button | Button Group | Calendar |
| Card | Carousel | Chart | Checkbox |
| Collapsible | Combobox | Command | Context Menu |
| Data Table | Date Picker | Dialog | Direction |
| Drawer | Dropdown Menu | Empty | Field |
| Hover Card | Input | Input Group | Input OTP |
| Item | Kbd | Label | Marker |
| Menubar | Message | Message Scroller | Native Select |
| Navigation Menu | Pagination | Popover | Progress |
| Radio Group | Resizable | Scroll Area | Select |
| Separator | Sheet | Sidebar | Skeleton |
| Slider | Spinner | Switch | Table |
| Tabs | Textarea | Toast | Toggle |
| Toggle Group | Tooltip | Typography |

## Theme completeness rule — light AND dark, both mandatory

Every catalog component MUST be bundled in **both** a light-theme and a
dark-theme token variant. A single-theme bundle is a D2 defect of the same
severity as a missing component:

- The user toggles theme modes on the Claude Design canvas. A component whose
  other-theme variant was never imported renders with wrong/default tokens
  after the toggle — a silent visual regression that `report_validate` may not
  catch (it checks render validity, not theme parity).
- Implementation downstream (run-phase) consumes BOTH variants from the
  handoff; a missing variant forces an ad-hoc guess during code generation.

When brand tokens (`.moai/project/brand/`) define only one theme, synthesize
the counterpart (invert the light/dark mapping) before `write_files` — do not
ship a single-theme bundle and defer the second theme.

## Usage in D2

1. Enumerate the full set above as the bundle checklist.
2. For each component, resolve its token spec in BOTH light and dark variants
   (brand tokens first; synthesize the missing variant per the rule above).
3. `write_files(localPath)` pushes per-component-unit (256 KiB ceiling per
   file); keep each component + its two theme variants in one push unit so a
   partial write cannot split a component from one of its themes.
4. After the push, cross-check the pushed set against this catalog — a missing
   component or a missing theme variant blocks D2 completion (H6 drift check).

## Maintenance

This catalog mirrors the shadcn/ui registry. When shadcn/ui adds a component,
append it to the table above in the same edit — the "every catalog component"
clause in `design.md` / `manager-design.md` D2 picks it up automatically; no
separate instruction edit is needed. Re-verify the canonical source before
adding.

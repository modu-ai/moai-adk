---
id: SPEC-DESKTOP-NATIVE-E2E-001
title: "Desktop-native E2E automation lane: replace the /moai e2e desktop-native deferral with OS-accessibility toolchains (macOS + Windows + Linux)"
version: "0.1.1"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/template/templates/.claude (skills/moai/workflows/e2e.md, agents/moai/e2e-specialist.md, commands/moai/e2e.md.tmpl) + local .claude siblings"
lifecycle: spec-anchored
tags: "e2e, desktop-native, accessibility, axcli, appium-mac2, flaui, pywinauto, dogtail, at-spi2, template, token-minimization"
era: V3R6
tier: M
related_specs: [SPEC-E2E-REVIVAL-001]
---

# SPEC-DESKTOP-NATIVE-E2E-001 — Desktop-native E2E automation lane (OS-accessibility toolchains)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.1 | 2026-07-13 | manager-spec | Audit-fix pass (plan-audit iter-1 FAIL 0.81, defects D1-D9 + §D-map nit). D1 BLOCKING: CMD-DNE-001/003/006 section windows rewritten from the self-terminating `/start/,/^### /` awk range (empirically emits only the heading) to flag-awk (`{f=1;next} f&&/^#{2,3} /{exit} f`); CMD-DNE-006 window now ends at the next heading (`### MCP Escalation Ladder`). D2: CMD-DNE-009A re-anchored from vacuous bare macOS/Windows/Linux/Accessibility tokens to per-OS recipe-heading regex (`^#### .*(macOS\|Windows\|Linux)`, 0 today) + TCC×Accessibility compound anchor (0 today). D3: bare `EXPERIMENTAL` grep (matches the Electron caveat today) replaced with a FlaUI×EXPERIMENTAL co-occurrence grep (0 today). D4: CMD-DNE-002 ordering check moved inside the flag-awk Detection Matrix window (removes first-match drift from Supported Flags / host-OS rule lines). D5: AC-DNE-021 REQ column gains 303. D6: CMD-DNE-015 agent-count check switched from alias-sensitive `ls \| wc -l` (13) to `find … -name '*.md' \| wc -l` (10, both trees). D7: CMD-DNE-012 tees actual counts into EVID (was "see stdout above"). D8: REQ-DNE-008 extended — Phase 5 report-template platform enum `{web \| mobile \| desktop \| mixed}` gains `desktop-native`; new AC-DNE-008c + report-enum check in CMD-DNE-008. D9: REQ-DNE-305 GEARS label corrected to (Ubiquitous — invariance) matching its affirmative shall-phrasing. Nit: §D Group A AC range corrected to name 008a/b/c. All rewritten commands re-run on the current tree: new anchors 0 (fail-today-pass-after), preservation checks PASS, windows emit section bodies. |
| 0.1.0 | 2026-07-13 | manager-spec | Plan-phase artifact set authored (Tier M, 4 artifacts: spec/plan/acceptance/progress). Origin: SPEC-E2E-REVIVAL-001 §E Exclusions deferred native-desktop automation to a follow-up SPEC (user decision 2026-07-13; former REQ-E2E-502 removed at parent v0.1.1). User decisions drained via orchestrator AskUserQuestion (2026-07-13): (1) **Scope = 3-OS full** — macOS + Windows + Linux recipes all documented; execution probes run only on the current host OS; macOS is locally verifiable, Windows/Linux recipes are declarative (verified by prose/structure ACs, not live runs). (2) **Tier M** — spec/plan/acceptance/progress only, plan-auditor standard depth. Toolchain matrix from verified 2026-07 research (see §A.2). Baselines measured live on this tree (see §A.3). |

---

## §A Context & Problem

SPEC-E2E-REVIVAL-001 (completed) revived `/moai e2e` for web + mobile + desktop, but its desktop coverage is limited to web-hybrid shells (Electron via Playwright `_electron`, Tauri via WebdriverIO + tauri-service). Projects built on **native desktop toolkits** — pure AppKit macOS apps, WinUI/Win32, Qt, GTK — classify as `desktop-native` and today route to the No-Target Graceful Exit branch with a deferral notice: "There is no opt-in automation path for `desktop-native`." (workflow skill, No-Target Graceful Exit section).

This SPEC is the promised follow-up (E2E-REVIVAL deferred item 2/2). It replaces the deferral with a real **OS-accessibility automation lane**: per-OS default + fallback toolchains driven through the existing e2e-specialist agent, CLI-first and token-minimized per the parent's Group B doctrine. No new agent, no new Go runtime logic — the deliverables are markdown edits to the two byte-identical file pairs (workflow skill + agent, local and template trees) plus the command argument-hint surfaces.

### §A.1 Extension points (codebase investigation, verified 2026-07-13)

1. **Workflow skill** — `.claude/skills/moai/workflows/e2e.md` + template sibling (currently byte-identical; edits MUST stay lockstep):
   - Detection Matrix: the `desktop-native` row currently carries "Automation not yet provided — see the graceful branch below". It gains POSITIVE per-toolkit markers and routes to the new lane. Most-specific-first ordering is preserved (Electron/Tauri desktop rows stay checked before `desktop-native`).
   - No-Target Graceful Exit: the deferral paragraph (containing the verbatim sentence "There is no opt-in automation path for `desktop-native`.") is replaced with routing into the automation lane; the graceful branch REMAINS for genuinely no-target projects (pure libraries) — parent REQ-E2E-007 semantics preserved.
   - Supported Flags: `--tool` gains the desktop-native toolchain tokens; `--platform` enum gains `desktop-native`.
   - Tool Matrix + Toolchain Probe table: gain desktop-native rows.
   - Execution Summary: the "(incl. `desktop-native` deferral notice)" mention is dropped.
2. **Agent** — `.claude/agents/moai/e2e-specialist.md` + template sibling (byte-identical): the `### desktop-native (non-Electron/non-Tauri)` stub ("Automation for native desktop toolkits is not provided by this agent…") is replaced with per-OS recipes (default + fallback, install commands, version probes, accessibility-permission prerequisites, artifact conventions); the Artifact Directory Conventions table is extended.
3. **Command** — `internal/template/templates/.claude/commands/moai/e2e.md.tmpl`: `argument-hint` `--platform` values gain `desktop-native`. Body must stay <20 non-empty LOC (thin-command audit). The local rendered sibling `.claude/commands/moai/e2e.md` is updated with the same argument-hint delta (render-pattern parity — command pairs are NOT byte-identical by design).
4. **Parent SPEC linkage**: SPEC-E2E-REVIVAL-001 spec.md (REQ-E2E-007, the v0.1.1 removal note, and the §E deferral entry) references this follow-up. The parent SPEC is completed and is NOT edited by this SPEC (SPEC bodies are immutable post-completion) — see §E.

### §A.2 Toolchain matrix (verified 2026-07 research — normative for the recipes)

| OS | Default | Fallback | Probe |
|----|---------|----------|-------|
| macOS | `axcli` (cargo install axcli; AXUIElement tree snapshot + background-safe actions, Playwright-like selectors; young project → pin version in the recipe) | appium-mac2-driver + WebdriverIO (v3.5.0 2026-05, healthy; reuses the existing Tauri WDIO lane; requires Xcode) | `axcli --version` / `appium driver list --installed` |
| Windows | FlaUI.WebDriver + WebdriverIO (W3C WebDriver2, UIA3; v0.4.0 EXPERIMENTAL → pin version + `GET /status` smoke probe) | pywinauto (`pip install pywinauto`; `print_control_identifiers()` as UIA tree dump) | server `/status` / python import probe |
| Linux | dogtail 2.x (AT-SPI2, PyPI 2.0.4 2025-06; requires at-spi2 packages; `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` for Qt; Wayland support GNOME-only via ponytail) | ydotool (Wayland) / xdotool (X11) blind input injectors + screenshot verification | python import probe / `ydotool --version` |
| All (last resort) | AX-tree text snapshot loop (hundreds of tokens per read) | computer-use screenshot loop (~1.1-1.6K tokens/frame; non-deterministic — NOT for CI-repeatable AC evidence) | n/a |

WinAppDriver and appium-windows-driver are EXCLUDED (abandoned since 2020) — see §E; the recipes do not reference them at all.

**Token-cost ordering doctrine** (aligns with parent REQ-E2E-100..104): filtered AX-tree text snapshot ≪ full tree JSON < single screenshot < screenshot loop. Screenshots are reserved for final visual evidence artifacts only. The file-redirect + bounded-tail contract applies (`agent-common-protocol.md` § Parallel Execution: ≤50 lines OR ≤2KB, whichever smaller).

### §A.3 Measured baseline (verified 2026-07-13, this tree)

| Surface | Current state | Post-SPEC state |
|---------|---------------|-----------------|
| Skill pair `workflows/e2e.md` (local vs template) | byte-identical (`diff` exit 0) | byte-identical after edits |
| Agent pair `e2e-specialist.md` (local vs template) | byte-identical (`diff` exit 0) | byte-identical after edits |
| Deferral-wording family (CMD-DNE-011 regex) in the 4 skill/agent files | skill: 3 matching lines per tree; agent: 1 matching line per tree | 0 matches in all 4 files |
| Verbatim sentence "There is no opt-in automation path for `desktop-native`." | 1 per tree (skill) | 0 in both trees |
| `--platform` enum (flags + argument-hints, both command surfaces) | `web\|mobile\|desktop` | `web\|mobile\|desktop\|desktop-native` |
| Command template body | 6 non-empty LOC | <20 non-empty LOC (unchanged pattern) |
| `desktop-native` token count | skill 3, agent 2 (per tree) | grows with the lane (reachability-anchored, not raw-count, per acceptance.md) |
| CI count pins | `expectedAgentCount = 10`, `expectedTotal = 38` | UNCHANGED (no new agent, no catalog changes) |

---

## §B Requirements (GEARS)

### Group A — Workflow-skill lane (detection, routing, flags; both trees)

- **REQ-DNE-001** (Ubiquitous): The e2e workflow Detection Matrix `desktop-native` row shall carry POSITIVE per-toolkit markers — AppKit (`.xcodeproj` / `Package.swift` with a macOS app target and no electron/tauri dependencies), WinUI/Win32 (`.vcxproj` / WinUI 3 project files), Qt (`CMakeLists.txt` with a Qt `find_package` / `.pro` files), GTK (meson or CMake with gtk dependencies) — replacing the "Automation not yet provided" note, in BOTH trees.
- **REQ-DNE-002** (Ubiquitous): The Detection Matrix shall preserve most-specific-first ordering: the Electron and Tauri desktop rows shall remain ABOVE (checked before) the `desktop-native` row, so web-hybrid shells never misroute into the native lane.
- **REQ-DNE-003** (Event-driven): **When** detection classifies a surface as `desktop-native`, the workflow shall route it into the desktop-native automation lane (Phase 0.5 toolchain selection among the §A.2 per-OS toolchains) instead of the graceful-exit branch.
- **REQ-DNE-004** (Ubiquitous): The No-Target Graceful Exit branch shall be preserved for genuinely no-target projects (pure libraries with no e2e-able surface — parent REQ-E2E-007 semantics), while the verbatim deferral sentence "There is no opt-in automation path for `desktop-native`." shall be removed from BOTH trees.
- **REQ-DNE-005** (Ubiquitous): The Supported Flags section shall extend `--platform` with `desktop-native` and `--tool` with the desktop-native toolchain tokens (`axcli`, `appium-mac2`, `flaui-webdriver`, `pywinauto`, `dogtail`), in BOTH trees. (The Linux blind injectors ydotool/xdotool remain recipe-level fallbacks, not first-class `--tool` values.)
- **REQ-DNE-006** (Ubiquitous): The Tool Matrix shall gain per-OS desktop-native rows with tier and token-cost classification consistent with the §A.2 token-cost ordering (default rows CLI-class; AX-tree snapshot cost named), in BOTH trees.
- **REQ-DNE-007** (Ubiquitous): The Toolchain Probe + Installation table shall gain the desktop-native probe rows (axcli version probe; appium mac2 driver installed-list probe; FlaUI.WebDriver `GET /status` smoke probe; pywinauto and dogtail python import probes; ydotool version probe), in BOTH trees.
- **REQ-DNE-008** (Ubiquitous): The Execution Summary shall drop the `desktop-native` deferral-notice mention — a detected `desktop-native` surface flows through the standard Detection → Selection → … pipeline — and the Phase 5 report-template platform enum `{web | mobile | desktop | mixed}` shall gain `desktop-native`, in BOTH trees.
- **REQ-DNE-009** (Event-driven): **When** a documented recipe's target OS differs from the host OS, the workflow shall treat that recipe as declarative documentation for this host (no live probe/execution) and shall run execution probes only for the host OS (3-OS-full docs, current-OS-only probes — user decision 2026-07-13).

### Group B — e2e-specialist per-OS recipes (both trees)

- **REQ-DNE-100** (Ubiquitous): The e2e-specialist `### desktop-native` stub shall be replaced with per-OS recipe subsections (macOS / Windows / Linux), each carrying: default + fallback toolchain, install commands, version probes, accessibility-permission prerequisites, and artifact conventions, in BOTH trees.
- **REQ-DNE-101** (Ubiquitous): The macOS recipe's default shall be `axcli` (install via `cargo install axcli`; AXUIElement tree snapshot + background-safe actions; Playwright-like selectors), with the version PINNED in the recipe (young project). Probe: `axcli --version`.
- **REQ-DNE-102** (Ubiquitous): The macOS recipe's fallback shall be appium-mac2-driver + WebdriverIO (reuses the existing Tauri WDIO lane; requires Xcode). Probe: `appium driver list --installed`.
- **REQ-DNE-103** (Event-driven): **When** the macOS Accessibility (TCC) permission is not granted to the executing terminal/host process, the e2e-specialist shall surface the System Settings grant path and return a structured blocker report — it shall not silently fail and shall not prompt the user.
- **REQ-DNE-104** (Ubiquitous): The Windows recipe's default shall be FlaUI.WebDriver + WebdriverIO (W3C WebDriver2 over UIA3), documented declaratively with the EXPERIMENTAL caveat, a pinned version, and the `GET /status` smoke probe.
- **REQ-DNE-105** (Ubiquitous): The Windows recipe's fallback shall be pywinauto (`pip install pywinauto`), with `print_control_identifiers()` documented as the UIA tree dump. Probe: a python import probe.
- **REQ-DNE-106** (Ubiquitous): The Linux recipe's default shall be dogtail 2.x (AT-SPI2), documented declaratively with: the at-spi2 package prerequisite, `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` for Qt apps, and the Wayland caveat (GNOME-only via ponytail). Probe: a python import probe.
- **REQ-DNE-107** (Ubiquitous): The Linux recipe's fallback shall be the blind input injectors ydotool (Wayland) / xdotool (X11) PAIRED with screenshot verification (blind injection without verification is not a recipe). Probe: `ydotool --version`.
- **REQ-DNE-108** (Ubiquitous): The recipes shall document the universal last resort: an AX-tree text snapshot loop (hundreds of tokens per read) as the cross-OS floor, and the computer-use screenshot loop (~1.1-1.6K tokens/frame, non-deterministic) as reserved for final visual evidence artifacts — explicitly NOT acceptable as CI-repeatable AC evidence.
- **REQ-DNE-109** (Unwanted): The recipes shall not reference WinAppDriver or appium-windows-driver (abandoned since 2020) — zero mentions in the 4 skill/agent files across both trees.
- **REQ-DNE-110** (Ubiquitous): The recipes shall encode the token-cost ordering doctrine (filtered AX-tree text snapshot ≪ full tree JSON < single screenshot < screenshot loop) and the bounded-tail/file-redirect contract (≤50 lines OR ≤2KB; verbose output redirected under `e2e/.runs/` with citable paths), mirroring parent REQ-E2E-100..104.
- **REQ-DNE-111** (Ubiquitous): The Artifact Directory Conventions table shall gain the desktop-native rows: scripts/flows under `e2e/desktop-native/`; AX-tree snapshots ride the existing `e2e/.runs/` timestamped-log convention.
- **REQ-DNE-112** (Event-driven): **When** a required desktop-native toolchain is absent (version probe fails), the missing-toolchain sequence shall follow the parent REQ-E2E-006 pattern: probe → the ORCHESTRATOR surfaces the exact install command(s) for approval → install → re-probe. The e2e-specialist never prompts the user (blocker reports only).

### Group C — Command surfaces

- **REQ-DNE-200** (Ubiquitous): The template command `e2e.md.tmpl` `argument-hint` shall extend the `--platform` values with `desktop-native`, and the body shall remain <20 non-empty lines (Thin Command Pattern preserved).
- **REQ-DNE-201** (Ubiquitous): The local rendered command `.claude/commands/moai/e2e.md` `argument-hint` shall receive the same `--platform` delta in the same change (render-pattern parity; byte parity is NOT asserted for command pairs).

### Group D — Cross-tree integrity & guards

- **REQ-DNE-300** (Ubiquitous): The workflow-skill pair and the agent pair shall remain byte-identical across the local and template trees after all edits (`diff` exit 0 per pair).
- **REQ-DNE-301** (Event-driven): **When** template-tree authoring completes, `make build` shall be run so the embedded FS (`//go:embed all:templates`) recompiles with the edited artifacts.
- **REQ-DNE-302** (Unwanted): The change shall not add any agent file, shall not modify `catalog.yaml`, `model_policy.go`, or any CI count pin (`expectedAgentCount` stays 10; `expectedTotal` stays 38) — body-only edits to existing artifacts.
- **REQ-DNE-303** (Unwanted): Template-tree content shall not contain internal SPEC IDs, REQ/AC tokens, audit citations, internal work dates, commit SHAs, or CLAUDE.local references (template neutrality) — the existing neutrality CI guard shall pass unchanged.
- **REQ-DNE-304** (Ubiquitous): The CI guards shall stay green: `TestCommandsThinPattern`, `TestCommandsFrontmatterConsistency`, the template neutrality guard, and `go test ./internal/template/...` exit 0.
- **REQ-DNE-305** (Ubiquitous — invariance): After the edits, the deferral-wording family regex (`deferral notice|deferred to a follow-up|not yet provided|no opt-in automation path|not provided by this agent`) shall return 0 matches across the 4 skill/agent files in BOTH trees (measured baseline: skill 3 + agent 1 matching lines per tree), while the no-target graceful-exit text for genuinely target-less projects remains present.

---

## §C Non-functional Constraints

- **C-1** [HARD] Template-First: the template sibling is authored in the same change; the skill and agent pairs stay byte-identical across trees; `make build` runs after template edits.
- **C-2** [HARD] NO new agent file — reuse e2e-specialist. Adding an agent breaks the CI pins (`expectedAgentCount = 10` in `catalog_tier_audit_test.go`) and the 66-cell tier matrix (`model_policy_test.go`). Body-only edits require no `catalog.yaml` / `model_policy.go` changes.
- **C-3** [HARD] Thin Command Pattern preserved: command body <20 non-empty LOC.
- **C-4** [HARD] Subagent boundary: e2e-specialist never prompts users; missing prerequisites (Accessibility/TCC permission not granted, toolchain absent) produce blocker reports / orchestrator-surfaced install commands per the parent REQ-E2E-006 pattern.
- **C-5** Token minimization first-class: CLI-output-first, bounded tail ≤50 lines/≤2KB, artifacts under project-local `e2e/` with citable paths (mirror parent REQ-E2E-100..104).
- **C-6** macOS-only live verification: ACs are structured so Windows/Linux recipes are verified by grep/structure checks (declarative); the macOS path MAY additionally be smoke-probed locally (non-gating); no CI dependency on Windows/Linux runners is introduced.

---

## §D Acceptance Criteria Map

The full AC matrix, Given-When-Then scenarios, edge cases, Executable Command Block, and Definition of Done live in `acceptance.md`. Mapping summary:

| REQ group | ACs |
|-----------|-----|
| A (workflow-skill lane) | AC-DNE-001 … AC-DNE-007, AC-DNE-008a/b/c |
| B (per-OS recipes) | AC-DNE-009 … AC-DNE-017 |
| C (command surfaces) | AC-DNE-018 … AC-DNE-019 |
| D (integrity & guards) | AC-DNE-020 … AC-DNE-024 |

---

## §E Exclusions

The exclusions below are out of scope for this SPEC.

### Out of Scope — WinAppDriver / appium-windows-driver

- Both are abandoned upstream (no releases since 2020) and are EXCLUDED from the Windows recipes entirely. The recipes do not mention them even as anti-recommendations (REQ-DNE-109 pins zero references); this SPEC §E is the sole record of the exclusion decision.

### Out of Scope — Windows/Linux live execution & CI runners

- No Windows or Linux CI runner, VM matrix, or live execution path is introduced. Windows/Linux recipes are declarative documentation verified by structure/grep ACs (user decision 2026-07-13, scope option "3-OS full"); only the macOS path is locally probe-able, and even that probe is non-gating (C-6).

### Out of Scope — Parent SPEC edits

- SPEC-E2E-REVIVAL-001 is `completed` and its body is immutable post-completion. Its REQ-E2E-007 text, the v0.1.1 REQ-E2E-502-removal note, and the §E native-desktop deferral entry all reference "a follow-up SPEC" — this SPEC IS that follow-up, and the linkage is recorded here (plus `related_specs`) rather than by editing the closed parent.

### Out of Scope — New agent / Go runtime changes

- No new agent file, no `catalog.yaml` entry, no `model_policy.go` row, no CI count-pin change, no new Go control flow. The `moai` binary changes only by re-embedding the edited markdown templates via `make build`.

### Out of Scope — MCP computer-use server provisioning

- The screenshot-loop last resort is documented as a cost class, but no MCP server registration, provisioning automation, or `.mcp.json` editing ships here (parent §E pattern; provisioning stays with the `/moai project` MCP flow).

### Out of Scope — docs-site documentation

- adk.mo.ai.kr 4-locale documentation of the desktop-native lane is deferred to a future docs-site sync (the existing `/moai e2e` docs-site page from SPEC-DOCSITE-E2E-001 is not edited here).

### Out of Scope — Device farms / remote desktop grids

- No cloud desktop grids or remote-machine execution integration; local host execution only.

---
description: >
  Multi-platform end-to-end testing workflow for web, mobile, and desktop
  applications. Auto-detects the project platform type, selects a CLI-first
  toolchain (Playwright, Maestro, Playwright-Electron, WebdriverIO + tauri-service),
  maps user journeys, creates and executes test scripts with token-minimized
  output, and reports results with citable artifact paths.
user-invocable: false
metadata:
  version: "3.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-07-14"
  tags: "e2e, end-to-end, testing, web, mobile, desktop, playwright, maestro, appium, detox, electron, tauri, user-journey"
  docs-libraries: "microsoft/playwright"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["e2e", "end-to-end", "e2e test", "browser test", "mobile test", "desktop test", "playwright", "maestro", "user journey"]
  agents: ["e2e-tester"]
  phases: ["e2e"]
---

# Workflow: E2E — Multi-Platform End-to-End Testing

Purpose: Create and run end-to-end tests that validate complete user flows through web, mobile, and desktop applications. Detects the project platform type first, then selects the appropriate toolchain per platform, with token-minimized CLI-first execution as a hard rule throughout.

Flow: Detection → Selection → Journey Mapping → Script Creation → Execution → Recording (optional) → Report

Execution owner: the **e2e-tester** subagent performs detection probes, journey mapping, script creation, execution, and recording. The ORCHESTRATOR owns every user-facing question in this workflow — the e2e-tester receives all selections via its spawn prompt and never prompts the user.

## Supported Flags

- `--tool TOOL`: Force toolchain selection, skipping the selection question. Options: playwright, agent-browser, chrome-devtools-mcp, claude-in-chrome, maestro, appium, detox, playwright-electron, wdio-tauri, axcli, appium-mac2, flaui-webdriver, pywinauto, dogtail (default: ask via the orchestrator's AskUserQuestion)
- `--platform web|mobile|desktop|desktop-native`: Force the platform classification when markers are ambiguous
- `--record`: Record runs via the selected toolchain's NATIVE trace/recording facility
- `--url URL`: Target URL for web testing (default: auto-detect from project config)
- `--journey NAME`: Run a specific named user journey only
- `--headless`: Run in headless mode (default: true; web/desktop-web toolchains)
- `--browser BROWSER`: Browser for Playwright (default: chromium). Options: chromium, firefox, webkit
- `--timeout N`: Test timeout in seconds (default: 30)
- `--retry N`: Retries for failed tests (default: 1). Retries re-run ONLY the failed specs, never the full suite silently
- `--autofix`: Enable autonomous fix delegation. On Phase 3 failure/improvement findings, the orchestrator groups findings (parallel for independent, sequential for dependent), delegates each group to `manager-develop` (cycle_type=autofix), and re-runs Phase 3 — looping up to 3 iterations until green or user escalation. One Implementation Kickoff Approval gates the whole loop

## Hard Rules — Token Minimization

- [HARD] **CLI-first**: every capability achievable via CLI invocation uses the CLI path. MCP tools are used ONLY for capabilities the selected CLI cannot provide (see the Tool Matrix token-cost column).
- [HARD] **Bounded output**: the e2e-tester redirects verbose run output to files under `e2e/.runs/` and surfaces only exit code + bounded tail in context, citing the file path.
- [HARD] **No MCP hard dependency**: every default platform path is fully executable with CLI-only tools. MCP-tier tools are conditional additions, never prerequisites.
- [HARD] **Artifacts by path**: reports, traces, screenshots, and recordings are persisted under project-local `e2e/` directories and cited by path — never inlined into context.

## Phase 0: Project-Type Detection

[HARD] Delegate detection to the **e2e-tester** subagent (read-only marker scan via Glob/Read).

### Detection Matrix

The marker scan treats all supported project ecosystems equally — detection is marker-driven, and no language or framework receives privileged treatment. Rows are ordered most-specific-first; a project matching multiple rows classifies as `mixed`.

| Platform class | Markers (any of) | Notes |
|----------------|------------------|-------|
| `desktop` (electron) | `electron` in package.json dependencies; `electron-builder` / Forge config files | Checked before generic web — an Electron repo also carries package.json |
| `desktop` (tauri) | `src-tauri/tauri.conf.json`; `tauri` in dependencies or Cargo.toml | Rust + web hybrid |
| `mobile` (react-native) | `react-native` in dependencies; `ios/` + `android/` directories; `app.json` with `expo` (managed workflow — `ios/`/`android/` may be absent) | Detox becomes an RN-conditional option |
| `mobile` (flutter) | `pubspec.yaml` containing `flutter:`; `lib/main.dart` | Maestro supports Flutter apps |
| `mobile` (native) | `*.xcodeproj` / `Package.swift` with iOS targets; `build.gradle` with `com.android.application` | Maestro/Appium capable |
| `web` | Web framework configs (next/nuxt/vite/astro/sveltekit/angular); `index.html` servers; any HTTP-serving app in any supported language ecosystem (Django, Rails, Spring, Fiber, Phoenix, Laravel, ...) | Broadest class; the framework list is exemplary — detection is marker-driven |
| `desktop-native` | Native toolkit markers WITHOUT Electron/Tauri — AppKit (`.xcodeproj` / `Package.swift` with a macOS app target, no electron/tauri dependencies), WinUI/Win32 (`.vcxproj` / WinUI 3 project files), Qt (`CMakeLists.txt` with Qt `find_package` / `.pro` files), GTK (meson or CMake with gtk dependencies) | Routes to the desktop-native automation lane — per-OS accessibility toolchain selection in Phase 0.5 |
| `mixed` | Two or more platform classes matched | Enumerate matched surfaces; per-surface selection in Phase 0.5 |
| none | No markers above | Graceful exit — see below |

### No-Target Graceful Exit

When NO e2e-able surface is detected (for example a pure library with no web/mobile/desktop entry point), report "no e2e target detected" listing the marker evidence consulted, and exit WITHOUT creating any `e2e/` artifacts.

A detected `desktop-native` surface does NOT take this branch: it routes into the desktop-native automation lane — Phase 0.5 toolchain selection over the per-OS accessibility recipes owned by the e2e-tester agent (axcli on macOS; FlaUI.WebDriver on Windows; dogtail on Linux). The graceful branch remains reserved for genuinely target-less projects.

### Host-OS Rule (desktop-native recipes)

The desktop-native lane documents recipes for all three OSes (macOS, Windows, Linux). When a documented recipe's target OS differs from the host OS, that recipe is treated as declarative documentation for this host — no live probe or execution is attempted, and the report states the host-OS/target-OS mismatch. Execution probes run only for the host OS.

### Toolchain Probe + Installation

After classification, the e2e-tester probes the DEFAULT toolchain for each detected platform:

| Toolchain | Version probe | Install command |
|-----------|---------------|-----------------|
| Playwright CLI | `npx playwright --version` | `npm i -D @playwright/test && npx playwright install --with-deps chromium` |
| Maestro | `maestro --version` | `curl -fsSL "https://get.maestro.mobile.dev" \| bash` |
| Playwright `_electron` | `npx playwright --version` (same package) | Same as Playwright CLI (no extra browser needed — the Electron binary is the target) |
| WebdriverIO + tauri-service | `npx wdio --version` | `npm i -D @wdio/cli @wdio/tauri-service && npx wdio config` |
| agent-browser (alternative) | `agent-browser --version` | `npm i -g agent-browser && agent-browser install` |
| Appium (fallback) | `appium --version` | `npm i -g appium && appium driver install uiautomator2` (Android) / `appium driver install xcuitest` (iOS) |
| Detox (RN only) | `npx detox --version` | `npm i -D detox @config-plugins/detox` + per-app native build configuration |
| axcli (desktop-native macOS) | `axcli --version` | `cargo install axcli` — PIN the version (young project); record the pinned version in the flow header |
| appium-mac2-driver (desktop-native macOS fallback) | `appium driver list --installed` | `npm i -g appium && appium driver install mac2` (requires Xcode) |
| FlaUI.WebDriver (desktop-native Windows) | `GET /status` smoke probe against the running server | Download the PINNED FlaUI.WebDriver release (v0.4.0); start the server, then probe `/status` |
| pywinauto (desktop-native Windows fallback) | `python -c "import pywinauto"` | `pip install pywinauto` |
| dogtail (desktop-native Linux) | `python -c "import dogtail"` | `pip install dogtail` (requires distribution at-spi2 packages) |
| ydotool (desktop-native Linux fallback) | `ydotool --version` | Install via the distribution package manager (Wayland); `xdotool` for X11 |

Missing-toolchain sequence (per selected toolchain):

1. **Probe**: run the version probe. On success, proceed.
2. **Surface**: on failure, the ORCHESTRATOR presents the exact install command(s) to the user for approval via AskUserQuestion.
3. **Install**: upon approval, the e2e-tester performs the installation.
4. **Re-probe**: re-run the version probe and confirm the version BEFORE Phase 1 begins.

Mobile probes must distinguish "CLI missing" from "no booted device/simulator" — each failure gets its own remedy (install command vs `xcrun simctl boot <device>` / `emulator -avd <name>`).

## Phase 0.5: Toolchain Selection

If `--tool` is provided: bypass the selection question entirely and use the named toolchain directly (verify with a version probe; run the missing-toolchain sequence if absent). Skip to Phase 1.

If `--tool` is NOT provided: the ORCHESTRATOR presents the toolchain options for the detected platform via AskUserQuestion — one question per platform surface when the classification is `mixed` (never one global toolchain forced across surfaces). Option rules:

- The first option carries the locale-appropriate Recommended label per the defaults in the Tool Matrix below
- Every option description states install status (from the Phase 0 probe) + factual trade-offs in neutral language (bias-prevention rule)
- The e2e-tester NEVER presents these questions; it receives the final selection via its spawn prompt

Recommendation modifiers:

- `CI=true` environment detected → bias toward Playwright CLI (web) and headless flags throughout; MCP-tier tools marked unavailable
- `--record` → prefer toolchains with a native trace/recording facility (Playwright trace, Maestro recording)
- Explicit performance/Lighthouse ask → chrome-devtools-mcp becomes the recommended row FOR THAT CAPABILITY only
- React Native markers detected → Detox appears as an RN-conditional alternative with a factual trade-off description
- `desktop-native` surface detected → the host-OS accessibility default is the recommended row (axcli on macOS; FlaUI.WebDriver on Windows; dogtail on Linux); recipes for other OSes are surfaced as declarative documentation only (Host-OS Rule)

## Tool Matrix (per-capability CLI-vs-MCP classification)

| Platform | Tool | Tier | Token cost | Use for |
|----------|------|------|------------|---------|
| web | **Playwright CLI** (default) | CLI | Low (exit code + bounded tail; JSON reporter on disk) | Deterministic cross-browser suites (chromium/firefox/webkit); CI pipelines |
| web | agent-browser | CLI | Low (accessibility-tree snapshots with deterministic element refs) | AI-exploratory journeys where selectors are unknown; Chromium-family only |
| web | chrome-devtools-mcp | MCP (conditional) | High (per-call round-trips) | ONLY live performance traces/insights, Lighthouse-class audits — capabilities with no CLI equivalent |
| web | Claude in Chrome | MCP (conditional) | High (requires visible Chrome; no CI path) | ONLY interactive visual debugging when the user explicitly asks |
| mobile | **Maestro** (default) | CLI | Low (declarative YAML flows; plain CLI output) | iOS/Android/Flutter flows; deterministic single-binary execution |
| mobile | Appium 3.x | CLI + server | Mid (session-based scripts; more verbose output) | Fallback: widest device/driver matrix (W3C WebDriver) when Maestro's declarative surface cannot express a flow |
| mobile | Detox | CLI | Low-mid | React Native ONLY (gray-box RN synchronization); offered when RN markers detected |
| desktop (electron) | **Playwright `_electron`** (default) | CLI | Low (reuses the web Playwright install) | Electron apps. API is EXPERIMENTAL — carry the caveat in reports. Native OS dialogs bypass Playwright: mock them in the Electron MAIN process via `evaluate()` |
| desktop (tauri) | **WebdriverIO + `@wdio/tauri-service`** (default) | CLI | Low-mid (CLI runner output) | Tauri apps. Embedded-WebDriver mode is cross-platform INCLUDING macOS; the native tauri-driver route is Windows/Linux only — never steer macOS projects there |
| desktop-native (macOS) | **axcli** (default) | CLI | Low (filtered AX-tree text snapshot — hundreds of tokens per read) | AppKit / native macOS apps via the AXUIElement accessibility tree; background-safe actions, Playwright-like selectors; version PINNED in the recipe |
| desktop-native (macOS) | appium-mac2 + WebdriverIO (fallback) | CLI + server | Mid (session-based specs) | Fallback when axcli cannot express a flow; reuses the Tauri WDIO lane; requires Xcode |
| desktop-native (Windows) | **FlaUI.WebDriver + WebdriverIO** (default, declarative) | CLI + server | Mid (W3C WebDriver2 over UIA3) | WinUI/Win32/Qt-on-Windows apps; EXPERIMENTAL — pin the version, `GET /status` smoke probe |
| desktop-native (Windows) | pywinauto (fallback, declarative) | CLI | Mid (`print_control_identifiers()` UIA tree dump) | Python UIA scripting fallback |
| desktop-native (Linux) | **dogtail** (default, declarative) | CLI | Low-mid (AT-SPI2 tree queries) | GTK/Qt apps via AT-SPI2; Qt needs `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1`; Wayland is GNOME-only (ponytail) |
| desktop-native (Linux) | ydotool / xdotool + screenshot verification (fallback, declarative) | CLI | High (blind input + screenshot verification) | Non-GNOME Wayland (ydotool) / X11 (xdotool) blind injection PAIRED with screenshot verification |

Every platform's DEFAULT row is CLI-class. MCP rows are conditional-only: they require `.mcp.json` registration + session restart and are never a prerequisite for the default path.

### MCP Escalation Ladder (embedded in the e2e-tester as a hard rule)

1. **Rung 1 — CLI + bounded tail**: full output → `e2e/.runs/<timestamp>-<slug>.log`; context gets exit code + bounded tail; log path cited.
2. **Rung 2 — CLI structured reporters**: JSON-class reporter output parsed selectively (failed specs only) instead of verbose re-runs.
3. **Rung 3 — MCP, batched, capability-gated**: only for capabilities the matrix marks CLI-impossible; snapshot/batch reads over per-element round-trips; never polling loops.

## Phase 1: Journey Mapping

[HARD] Delegate journey mapping to the **e2e-tester** subagent.

If `--journey` is provided: load the specified journey definition and skip to Phase 2.

Journey discovery (e2e-tester, read-only):

- Read project documentation (`.moai/project/product.md`) for feature descriptions
- Analyze route definitions (routes.ts, urls.py, router.go, navigation graphs, deep-link manifests) for available paths
- Identify forms, authentication flows, and CRUD operations
- Map critical user paths (login, main feature, error handling); for mobile: launch → onboarding → core flow; for desktop: window open → menu actions → dialog flows

The e2e-tester returns the discovered journey list; the ORCHESTRATOR presents it via AskUserQuestion:

- Test all journeys (Recommended): most comprehensive coverage, longest execution
- Select specific journeys: focus on recently changed features
- Define custom journey: the user describes a flow; the e2e-tester scripts it from the description

Journey Definition Format:

```markdown
Journey: User Login
Steps:
1. Navigate to /login (web) | Launch app to login screen (mobile/desktop)
2. Enter email in the email field
3. Enter password in the password field
4. Submit
5. Verify redirect to /dashboard (web) | dashboard screen visible (mobile/desktop)
6. Verify welcome message displayed
```

## Phase 2: Script Creation

[HARD] Delegate test script creation to the **e2e-tester** subagent.

Per-toolchain file conventions:

| Toolchain | Artifact | Location |
|-----------|----------|----------|
| Playwright (web + `_electron`) | `.spec.ts` / `.spec.js` / `test_*.py` specs | `e2e/<journey>.spec.ts` |
| agent-browser | Task definitions (natural-language steps + assertions) | `e2e/<journey>.agent.ts` |
| Maestro | Declarative flow YAML | `e2e/flows/<journey>.yaml` |
| Appium / WebdriverIO | WDIO/W3C WebDriver specs + `wdio.conf.ts` | `e2e/<journey>.e2e.ts` |
| Detox | Detox specs + `.detoxrc.js` | `e2e/<journey>.test.js` |

Script quality expectations:

- Playwright: Page Object Model where the suite exceeds a handful of specs; setup/teardown fixtures; step-scoped assertions; screenshots at key verification points
- Playwright `_electron`: launch fixture with executable path/args; `firstWindow()`; main-process `evaluate()` mocks for native dialogs
- Maestro: one flow file per journey; `assertVisible`/`assertTrue` outcomes per step; built-in timing tolerance preferred over manual sleeps
- WebdriverIO + tauri-service: embedded mode config (macOS-safe); Tauri command mocking where the journey crosses native commands
- All toolchains: each journey step maps to a verifiable outcome — no assertion-free navigation scripts

## Phase 3: Execution

[HARD] Delegate test execution to the **e2e-tester** subagent — CLI-first, bounded output.

Execution pattern (Rung 1, all CLI toolchains):

```bash
# Example: Playwright (the same redirect pattern applies to maestro test,
# npx wdio run, npx detox test, and agent-browser invocations)
npx playwright test e2e/ > e2e/.runs/$(date +%Y%m%d-%H%M%S)-suite.log 2>&1; \
  echo "exit=$?"; tail -50 e2e/.runs/*-suite.log
```

- Parse results from exit code + bounded tail; full log stays on disk with a citable path
- On failure, escalate to Rung 2 (JSON reporter parsed for failed specs only) before any verbose re-run
- `--headless` / `--timeout` / `--retry` flags map to the selected toolchain's native options
- Retries (`--retry N`) re-run ONLY failed specs
- MCP-tier execution (Rung 3) applies only to the conditional capabilities in the Tool Matrix, batched

## Phase 3.5: Autofix Delegation (--autofix only)

Activates ONLY when `--autofix` is set AND Phase 3 produced failures or improvement findings. Skipped entirely when Phase 3 is green (proceeds straight to Phase 4/5) or when `--autofix` is absent (Phase 3 failures route to the normal Phase 5 next-step question).

### Kickoff Approval (one-time gate)

[HARD] Before the FIRST fix delegation of the loop, the ORCHESTRATOR obtains a single Implementation Kickoff Approval via AskUserQuestion — "enter autonomous autofix loop (manager-develop cycle_type=autofix, max 3 iterations, parallel-where-safe)". This one approval covers the entire loop; subsequent iterations do NOT re-ask. Declining falls back to the standard Phase 5 manual next-step.

### Finding Grouping (orchestrator, blast-radius analysis)

The orchestrator groups Phase 3 failures/improvement findings by blast radius:

- **Independent group → parallel fan-out**: findings touching disjoint files/modules with zero overlap spawn as concurrent `manager-develop` delegations (Mode 4 ceiling: 3-5 concurrent).
- **Dependent group → sequential**: findings touching the same module/file (cascade risk) are bundled into ONE `manager-develop` delegation processed in order.

[HARD] Write-capable agents never run concurrently on overlapping scope (`agent-common-protocol.md` § Background Agent Execution). Orchestrator work concurrent with a write-capable agent stays read-only.

### Delegation Contract

Each `manager-develop` spawn (cycle_type=autofix):
- **Skill injection** (skill-routing.md §1): inject `At start, invoke Skill("moai-workflow-ddd") for the autofix repair cycle.` plus 0-3 domain `moai-ref-*` skills matched to the failing journey's domain (e.g. frontend → moai-ref-react-patterns, backend → moai-ref-api-patterns; per `.moai/config/sections/delegation.yaml` domain_skills)
- **Input**: failing journey(s) + bounded failure excerpt + artifact path (`e2e/.runs/<log>`) + exact reproduction
- **Cycle**: localize → repair → validate (manager-develop autofix)
- **Validate**: MUST re-run the relevant e2e spec locally (not the full suite) to confirm the fix before returning

### Loop Control

```
iteration = 0
while iteration < 3 and not green:
    delegate grouped fixes → manager-develop (autofix)   # parallel-where-safe
    re-run Phase 3 (CLI-first, bounded output)
    iteration++
green             → Phase 5 (success report)
iteration == 3    → escalate to user (remaining failures + artifact paths + AskUserQuestion)
```

Max 3 iterations mirrors `ci-autofix-protocol.md`. On exhaustion the orchestrator reports remaining failures with citable artifact paths and asks the user how to proceed (manual fix / re-run with adjusted scope / abort).

## Phase 4: Recording (optional)

Applies when `--record` is set. [HARD] Delegate recording to the **e2e-tester** subagent using the selected toolchain's NATIVE facility — never MCP screenshot loops:

| Toolchain | Native facility | Output |
|-----------|-----------------|--------|
| Playwright (web + `_electron`) | `npx playwright test --trace on` | `e2e/traces/*.zip` (view: `npx playwright show-trace`) |
| agent-browser | `--trace` flag | `e2e/recordings/` |
| Maestro | `maestro record <flow>.yaml` | `e2e/recordings/` |
| WebdriverIO | wdio video/trace reporter services | `e2e/recordings/` |

Recording conventions: descriptive filenames (`login_flow`, `checkout_process`), timestamps for versioning, all stored under `e2e/recordings/` or `e2e/traces/` — cited by path in the report, never inlined.

## Phase 5: Report

The ORCHESTRATOR renders the report in the user's conversation_language from the e2e-tester's return contract (per-journey status, artifact paths, bounded failure excerpts):

```markdown
## E2E Test Report

### Platform: {web | mobile | desktop | desktop-native | mixed} — Tool: {selected toolchain}

### Results Summary
| Journey | Status | Duration | Artifacts |
|---------|--------|----------|-----------|
| Login | PASS | 2.3s | 3 screenshots |
| Checkout | FAIL | 5.1s | 4 screenshots, 1 trace |

### Failures
- Checkout (Step 4): expected /confirmation, got /error
  - Log: e2e/.runs/20260101-120000-suite.log (bounded excerpt above; full log at path)
  - Screenshot: e2e/screenshots/checkout-step4.png

### Recordings (if --record)
- e2e/recordings/login_flow (path cited — content not inlined)

### Environment
- Toolchain: {name + probed version} | Device/simulator state (mobile) | headless: {true|false}
```

Next steps (ORCHESTRATOR AskUserQuestion): Fix failing tests (Recommended) / Rerun failed tests only / Add more journeys / Switch toolchain.

## Task Tracking

[HARD] Task management tools are mandatory:

- Each user journey tracked as a pending task via TaskCreate
- Before execution: in_progress via TaskUpdate
- After pass: completed via TaskUpdate
- Failed journeys remain in_progress with failure details

## Agent Chain Summary

- Phase 0: e2e-tester (detection probes + toolchain probe/install)
- Phase 0.5: MoAI orchestrator (AskUserQuestion selection; `--tool` bypass)
- Phase 1: e2e-tester (journey mapping)
- Phase 2: e2e-tester (script creation)
- Phase 3: e2e-tester (CLI-first execution)
- Phase 3.5 (--autofix only): orchestrator (grouping + Kickoff Approval) → manager-develop autofix (parallel-where-safe)
- Phase 4: e2e-tester (native-facility recording)
- Phase 5: MoAI orchestrator (report + next-step question)

## Execution Summary

1. Parse arguments (--tool, --platform, --record, --url, --journey, --headless, --browser, --timeout, --retry)
2. Phase 0: delegate detection to the e2e-tester; classify platform; probe defaults (host OS only for `desktop-native`); graceful exit when no target
3. Phase 0.5: orchestrator AskUserQuestion selection (per-surface on `mixed`); `--tool` bypasses
4. Missing toolchain: probe → surface install command for approval → install → re-probe
5. Phase 1: delegate journey mapping; orchestrator presents journey options
6. Phase 2: delegate script creation per toolchain conventions
7. Phase 3: delegate execution — CLI-first, bounded tail, file-redirect, selective JSON triage
7.5. (if --autofix and Phase 3 not green) Phase 3.5: group findings → Kickoff Approval (1회) → manager-develop autofix (parallel independent / sequential dependent) → re-run; max 3 iterations or green; else escalate
8. Phase 4: if --record, native-facility recording only
9. TaskCreate/TaskUpdate for all journeys
10. Phase 5: report in conversation_language with citable artifact paths

---

Version: 3.0.0

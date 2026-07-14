# SPEC-E2E-REVIVAL-001 — Research

## §A Method & Evidence Discipline

- External tool facts were live-verified on **2026-07-13** via WebFetch against the sources listed in §Sources — NOT asserted from training memory (knowledge cutoff predates several releases below).
- **Caveat (recorded per verification-claim integrity)**: the fetch digests for the Playwright and Maestro releases pages returned internally inconsistent YEAR labels (an artifact of the summarizing fetch layer). VERSION NUMBERS are cited as verified; exact release dates are NOT relied upon anywhere in this SPEC. Where a date matters at run-phase, re-fetch the specific release page.
- Repo-internal facts were measured with executed commands on this tree (2026-07-13); see plan.md §C for the re-verification checklist.

## §B Retired-Baseline Analysis

Source: `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md` (452 lines; frontmatter v2.7.0, body v2.1.0; removed by commit c6b04d39c).

### Inherit (proven structure)

- 7-stage flow: Tool Selection → Installation → Journey Mapping → Test Script Creation → Execution → Recording → Report
- Auto-detection → orchestrator AskUserQuestion selection with a recommendation matrix (condition → recommended tool → rationale)
- Token-cost comparison column ("CLI output = Low, MCP round-trips = High") — the load-bearing principle
- Journey Definition Format (numbered steps, verifiable outcomes)
- Artifact conventions: `e2e/` test files, `e2e/traces/`, `e2e/recordings/`, `e2e/screenshots/`
- Flags: `--tool`, `--record`, `--url`, `--journey`, `--headless`, `--browser`, `--timeout`, `--retry`
- TaskCreate/TaskUpdate journey tracking
- Report template in conversation_language with per-journey status table

### Strengthen

- Token-cost principle: was a descriptive table column → becomes [HARD] requirements (CLI-first REQ-E2E-100, bounded-tail REQ-E2E-101, no-MCP-hard-dependency REQ-E2E-105)
- Delegation target: was manager-develop (+ per-spawn general-purpose) → becomes the dedicated `e2e-tester`
- Platform scope: web-only → web + mobile + desktop with project-type detection ahead of tool detection

### Drop / correct (stale facts)

- "Chrome DevTools MCP … 29 tools across 6 categories" → v1.5.0 exposes ~60+ tools (verified §C3); all counts in the new artifacts come from §C, never ported prose
- Retired references to archived-agent-era delegation phrasing are re-grounded on the current retained catalog

## §C Web Toolchain (verified 2026-07-13)

### C1 — Playwright (DEFAULT, CLI-tier)

- **Verified**: v1.61.1 latest on the releases page. Notable agent-era features: `browser.bind()` (launched browsers accept playwright-cli/agent connections), `--debug=cli` (agents attach and debug over playwright-cli), screencast + action annotations with real-time frame streaming for AI vision, WebAuthn virtual authenticator, WebStorage API.
- **Fit**: deterministic cross-browser (chromium/firefox/webkit) suites; CLI output parsing; headless default; first-class CI. `npx playwright test` + JSON/HTML reporters → lowest token cost per assertion.
- **Trade-offs**: ~200MB install with browsers; script authoring required (mitigated: the specialist authors specs).

### C2 — agent-browser (Vercel Labs) (AI-EXPLORATORY, CLI-tier)

- **Verified**: v0.31.1 (92 releases; actively maintained). Native Rust CLI; headless by DEFAULT (`--headed` to show); `snapshot` command emits agent-readable accessibility trees with deterministic element refs (`@e1`, `@e2`); nav/interaction/screenshot/network-interception/cookies/JS-exec; install via npm (`npm i -g agent-browser` + `agent-browser install`), Homebrew, or cargo; uses Chrome for Testing.
- **Fit**: AI-driven exploratory journeys where selectors are unknown; the deterministic-ref snapshot is markedly cheaper than MCP DOM round-trips.
- **Trade-offs**: Chromium-family only; younger ecosystem than Playwright; no cross-browser matrix.

### C3 — chrome-devtools-mcp (CONDITIONAL, MCP-tier)

- **Verified**: v1.5.0 (2026-07-03 per the repo page), ~60+ tools (input automation, navigation, performance traces + insights, network debugging, memory/heap snapshots, screencast experimental), `--headless` supported, very active (46.8k stars, 56 releases).
- **Fit**: ONLY for capabilities with no CLI equivalent in the selected stack — live performance traces/insights, Lighthouse-class audits, interactive CSS/DOM forensics.
- **Trade-offs**: MCP round-trip token cost; requires `.mcp.json` registration + session restart; classified conditional-tier by REQ-E2E-100/105.

### C4 — Claude in Chrome (INTERACTIVE-DEBUG ONLY, MCP-tier)

- Built-in MCP; requires visible Chrome + extension; no CI path; highest token cost. Retained solely for interactive visual debugging when the user explicitly asks.

## §D Mobile Toolchain (verified 2026-07-13)

### D1 — Maestro (RECOMMENDED DEFAULT, CLI-tier)

- **Verified**: CLI 2.6.1. iOS + Android + web-flow support signals on the releases page (iOS passcode-screen interaction fixes, iframe hierarchy access for web flows).
- **Fit**: single-binary CLI; declarative YAML flows (`maestro test flow.yaml`); built-in tolerance for flakiness/timing; plain CLI output → the lowest-token mobile option and the closest mobile analogue of "Playwright-CLI-class" determinism. Flow YAML is also the easiest format for an agent to author correctly on the first pass.
- **Trade-offs**: less low-level device control than Appium; younger driver surface; cloud features out of scope here.

### D2 — Appium (FALLBACK, CLI+server-tier)

- **Verified**: appium@3.5.2 (2026-06 series active; storage-plugin fixes visible). Architecture detail did not render in the fetch digest — re-verify driver specifics (XCUITest/UiAutomator2 driver versions) at run-phase before writing install prose.
- **Fit**: widest platform/driver matrix (W3C WebDriver standard; native, hybrid, mobile-web); the escape hatch when Maestro's declarative surface can't express a flow.
- **Trade-offs**: server + driver + client-binding setup (heaviest); session-based scripts; more verbose output → higher token cost. Classified fallback, not default.

### D3 — Detox (RN-CONDITIONAL, CLI-tier)

- **Verified**: 20.51.3 (active; iOS 26+ and RN 0.83 compatibility work visible; semantic `by.type()` matching added in 20.47.0).
- **Fit**: gray-box synchronization for React Native — auto-waits on RN internals, so offered as the alternative ONLY when RN markers are detected.
- **Trade-offs**: React Native only; requires native build configuration per app.

## §E Desktop Toolchain (verified 2026-07-13)

### E1 — Playwright `_electron` (ELECTRON DEFAULT, CLI-tier)

- **Verified**: documented as **experimental**; supports Electron v12.2.0+/v13.4.0+/v14+; `_electron.launch()` with executable path/args; `firstWindow()`; `evaluate()` runs in the Electron MAIN process (the sanctioned pattern for mocking native dialogs, which bypass Playwright at the OS level).
- **Fit**: reuses the web-default toolchain (one install, one reporter format, one skill section); CLI execution.
- **Trade-offs**: experimental status → the workflow prose must carry the experimental caveat + the dialog-mocking pattern (edge case E-2).

### E2 — WebdriverIO + `@wdio/tauri-service` (TAURI DEFAULT, CLI-tier)

- **Verified** (v2.tauri.app): `@wdio/tauri-service` is maintained under the WebdriverIO project and is the RECOMMENDED route; the embedded-WebDriver mode supports **Windows, Linux, AND macOS**; the native `tauri-driver` route is **Windows/Linux only** (no WKWebView driver tooling on macOS); features include Tauri API access, command mocking, log capture, multiremote; Selenium possible via tauri-driver directly but manual.
- **Fit**: cross-platform including macOS via embedded mode; auto-detects the app binary; CLI runner output.
- **Trade-offs**: WebdriverIO config surface is heavier than Playwright's; macOS restriction applies ONLY to the native-driver route (edge case E-1 steers macOS to embedded mode).

### E3 — OS-level accessibility / computer-use (NATIVE-DESKTOP — DEFERRED)

- For non-Electron/non-Tauri native apps there is no dominant CLI-first cross-platform driver. OS-accessibility/computer-use approaches (screenshot + accessibility-tree + synthetic input) work but are token-expensive and non-deterministic. RESOLVED (user decision 2026-07-13, plan.md D-7): DEFERRED to a follow-up SPEC — former REQ-E2E-502 removed at v0.1.1. This SPEC's detection still classifies `desktop-native` and reports the deferral via the REQ-E2E-007 graceful branch (spec.md §E Exclusions).

## §F Token-Cost Model (per-tool classification)

| Tier | Cost profile | Tools |
|------|--------------|-------|
| CLI (low) | one process spawn; parse exit code + bounded tail; reporter files on disk | Playwright CLI, agent-browser, Maestro, Detox, Appium (runner output), WebdriverIO runner |
| CLI-structured (low-mid) | JSON reporter parsed selectively | Playwright `--reporter=json`, WDIO json reporter |
| MCP (high) | per-call round-trips ride the conversation | chrome-devtools-mcp, Claude in Chrome, computer-use fallback |

Escalation ladder (design.md §F): rung 1 CLI bounded-tail → rung 2 CLI structured reporters → rung 3 MCP, batched, only for CLI-impossible capabilities (live perf insights, Lighthouse, interactive debugging).

## §G Repo Baseline Measurements (executed 2026-07-13)

- Retired workflow recoverable: `git show c6b04d39c~1:...e2e.md | wc -l` → 452
- `e2e` absent from both SKILL.md trees (grep → 0 hits)
- `expectedSkillCount = 28` / `expectedAgentCount = 9` (`catalog_tier_audit_test.go`); `expectedTotal = 37` (`catalog_loader_test.go`)
- 9 agent files per tree; 13 command files per tree; `CLAUDE.md` §4 "exactly 10 retained agents (9 MoAI-custom + 1 Explore)" in BOTH trees (line 62 at measurement — content-anchor, not line-anchor)
- `gen-catalog-hashes.go` (invoked by `make build`) iterates EXISTING catalog entries only — new entries are manual-add-then-hash
- Command pairs are NOT byte-identical across trees (local `.md` is rendered; template is `.md.tmpl` — e.g. plan.md 254B vs plan.md.tmpl 918B): replicate the sibling render pattern, don't assert byte parity
- Agent files ARE byte-identical across trees except known sanitized pairs (manager-spec.md, builder-harness.md) — the new agent targets byte-identical
- `.moai/specs/` dedup: only `SPEC-HARNESS-EXECUTE-E2E-001` (unrelated harness-telemetry bugfix, completed) shares the e2e token

## §H Open Questions — RESOLVED (orchestrator AskUserQuestion round, 2026-07-13)

H-1 and H-2 were carried as clarification markers in plan.md §B (D-7 / D-8); H-3 was carried in THIS file only (an Out-of-Scope timing question, never a plan.md marker — the prior claim that all three were plan.md markers was inaccurate, corrected per audit iter-1 D11). All three are resolved and the markers removed at v0.1.1:

1. **H-1 (= D-7)**: desktop-native fallback → **DEFERRED to a follow-up SPEC** (REQ-E2E-502 removed; `desktop-native` detection routes to the REQ-E2E-007 graceful branch with a deferral notice).
2. **H-2 (= D-8)**: mobile default → **Maestro CONFIRMED** (Appium fallback, Detox RN-conditional), as drafted.
3. **H-3**: docs-site 4-locale documentation → **deferred to a follow-up** (stays Out of Scope; spec.md §E).

## Sources (fetched 2026-07-13)

- https://github.com/microsoft/playwright/releases — Playwright 1.61.1; browser.bind / --debug=cli / screencast features
- https://github.com/vercel-labs/agent-browser — agent-browser v0.31.1; Rust CLI; headless default; snapshot refs
- https://github.com/ChromeDevTools/chrome-devtools-mcp — v1.5.0; ~60+ tools; --headless; active maintenance
- https://github.com/mobile-dev-inc/maestro/releases — Maestro CLI 2.6.1; iOS/Android/web-flow signals
- https://github.com/appium/appium/releases — appium@3.5.2 (driver-architecture detail did not render; re-verify at run-phase)
- https://github.com/wix/Detox/releases — Detox 20.51.3; RN 0.83 / iOS 26+ compatibility
- https://v2.tauri.app/develop/tests/webdriver/ — tauri-driver platform matrix; @wdio/tauri-service recommendation; macOS embedded-only
- https://playwright.dev/docs/api/class-electron — Electron support experimental; firstWindow/evaluate main-process pattern

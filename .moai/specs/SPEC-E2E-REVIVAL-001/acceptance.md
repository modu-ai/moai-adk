# SPEC-E2E-REVIVAL-001 — Acceptance Criteria

> Every AC verifies REACHABILITY (existence + cross-file reference + executable gate), never token presence alone. Each AC names its executable check; run-phase evidence cites the executed command + verbatim output path. Baselines were measured 2026-07-13 (spec.md §A table) — re-verify at run entry.

## §D AC Matrix

### Group A — Detection & selection (content ACs on the workflow skill body)

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-001 | REQ-E2E-001 | Workflow skill contains a project-marker detection matrix covering web, mobile, desktop, and mixed classifications, with ≥2 marker examples per platform class | Read `internal/template/templates/.claude/skills/moai/workflows/e2e.md` §Detection; matrix table present with 4 platform rows; markers treat ecosystems equally (no "primary" language) |
| AC-E2E-002 | REQ-E2E-002 | Platform-toolchain matrix maps each detected type to a default: web→Playwright CLI, mobile→Maestro, desktop→Playwright-Electron/WebdriverIO+tauri-service | Matrix table present; defaults match design.md §C; each default's install + version-probe command included |
| AC-E2E-003 | REQ-E2E-003, REQ-E2E-004 | All selection questions are specified as ORCHESTRATOR AskUserQuestion instructions; zero instructions direct the e2e-specialist to prompt | `grep -n 'AskUserQuestion' <workflow>` — every hit sits in orchestrator-addressed prose; `grep -c 'specialist.*AskUserQuestion'` semantic review → 0 agent-prompts directives |
| AC-E2E-004 | REQ-E2E-005 | `--tool` flag documented in Supported Flags and short-circuits selection | Flag row present; Phase 0.5 contains the bypass branch |
| AC-E2E-005 | REQ-E2E-006 | Missing-toolchain path: version probe → install command surface → approval → re-probe | Phase 0 install section contains probe-install-reprobe sequence for each default toolchain |
| AC-E2E-006 | REQ-E2E-007, REQ-E2E-502 | No-target graceful exit AND desktop-native opt-in (with token-cost warning) both specified | Workflow contains "no e2e target" report branch; `desktop-native` section marked explicit opt-in |

### Group B — Token minimization

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-007 | REQ-E2E-100 | CLI-first rule stated as [HARD] in BOTH the workflow skill and the agent body, with the per-capability CLI-vs-MCP tool matrix in the workflow | `grep -n '\[HARD\]' <workflow> <agent>` includes the CLI-first rule; matrix has token-cost column |
| AC-E2E-008 | REQ-E2E-101 | Agent body carries the bounded-tail + file-redirect contract (≤50 lines OR ≤2KB; artifacts dir; citable paths) | `grep -n '50 lines\|2KB' <agent>` ≥1; redirect example present |
| AC-E2E-009 | REQ-E2E-102 | MCP batching rule (snapshot/batch over per-element round-trips) present in agent body | Section present in agent §token-minimization ladder rung 3 |
| AC-E2E-010 | REQ-E2E-103, REQ-E2E-104 | Report/trace/recording artifacts persist under `e2e/` dirs, paths cited not inlined; `--record` uses native toolchain facility | Workflow Phase 4/5 specify artifact dirs + path-citation rule; no MCP-screenshot-loop recording path |
| AC-E2E-011 | REQ-E2E-105 | Every default platform path executable CLI-only; no MCP server hard dependency | Tool matrix: each platform's DEFAULT row is CLI-class; MCP rows all marked conditional |

### Group C — Deliverable artifacts (existence + CI + boundary)

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-012 | REQ-E2E-200 | Thin command exists in both trees and passes thin-pattern CI | `test -f internal/template/templates/.claude/commands/moai/e2e.md.tmpl && test -f .claude/commands/moai/e2e.md`; `go test -run TestCommandsThinPattern ./internal/template/` exit 0 (log shows 28 command files: 14 per tree-representation baseline 26→28... measure: pre-count 26, post-count 28 in embedded FS is template-only → pre 13, post 14 `.md.tmpl` audited) |
| AC-E2E-013 | REQ-E2E-201 | Workflow skill exists in both trees, `user-invocable: false`, identical content | `diff .claude/skills/moai/workflows/e2e.md internal/template/templates/.claude/skills/moai/workflows/e2e.md` exit 0; `grep -c 'user-invocable: false'` = 1 |
| AC-E2E-014 | REQ-E2E-202 | Agent file exists in both trees, byte-identical, frontmatter complete (name/description/tools CSV/model/effort/color/permissionMode/memory/skills) | `diff` exit 0; `grep -c '^tools:'` = 1 and value is CSV (no leading `-`); `go test -run TestAgentFrontmatterAudit ./internal/template/` exit 0 |
| AC-E2E-015 | REQ-E2E-203 | Agent tools line excludes Agent and AskUserQuestion; body contains the `## Missing Inputs` blocker-report contract | `grep '^tools:' <agent> \| grep -cE '\bAgent\b\|AskUserQuestion'` → 0; `grep -c 'Missing Inputs' <agent>` ≥1 |
| AC-E2E-016 | REQ-E2E-204 | Cross-file reachability chain: workflow delegates to `e2e-specialist` by name in ≥3 phase sections AND the agent file exists at the resolving path in both trees AND catalog registers it | `grep -c 'e2e-specialist' <workflow>` ≥3; `test -f` both agent paths; `go test -run 'TestAllAgentsInCatalog\|TestCatalogReferencesValid' ./internal/template/` exit 0 |

### Group D — Router & catalog reachability (baseline-delta greps, both trees)

| AC | REQ | Criterion | Executable check (baseline → target) |
|----|-----|-----------|--------------------------------------|
| AC-E2E-017 | REQ-E2E-300 | Priority 1 router row restored in both SKILL.md trees | `grep -cE '^- \*\*e2e\*\*' <both SKILL.md>` : 0 → 1 each |
| AC-E2E-018 | REQ-E2E-301 | Frontmatter description enumeration + CLAUDE.md §3 Subcommands line include `e2e`, both trees | `grep -c 'e2e' <SKILL.md frontmatter block>` 0 → ≥1; `grep 'Subcommands:' CLAUDE.md internal/template/templates/CLAUDE.md \| grep -c e2e` : 0 → 2 |
| AC-E2E-019 | REQ-E2E-302 | CLAUDE.md §4 both trees: count text "11 retained agents (10 MoAI-custom", catalog table row `e2e-specialist`, Selection Decision Tree entry | `grep -c '11 retained agents'` 0 → 2 (across both files); `grep -c 'e2e-specialist' <both CLAUDE.md>` 0 → ≥2 each (table row + decision tree) |
| AC-E2E-020 | REQ-E2E-303 | Priority 3 semantic-classification cue line for e2e-testing intent added, both trees | `grep -n 'e2e' <SKILL.md P3 section>` ≥1 each; cue line phrased as semantic exemplar (not literal-match requirement) |
| AC-E2E-021 | REQ-E2E-304 | catalog.yaml core.agents entry with real (non-placeholder) 64-hex hash | `grep -A4 'name: e2e-specialist' internal/template/catalog.yaml` shows tier/path/hash/version; hash matches `^[0-9a-f]{64}$`; `go test -run TestAllAgentsInCatalog ./internal/template/` exit 0 |

### Group E — Distribution & CI reconciliation

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-022 | REQ-E2E-400 | Template-First provenance: every new local file has a template sibling; commit order/content shows template tree authored in the same commit as (or before) local | `git show --stat <run commits>` — no local-only additions; parity diffs (AC-013/014) green |
| AC-E2E-023 | REQ-E2E-401 | Count constants reconciled with provenance comments | `grep -n 'expectedAgentCount = 10' catalog_tier_audit_test.go` = 1; `grep -n 'expectedTotal = 38' catalog_loader_test.go` = 1; `grep -n 'expectedSkillCount = 28'` STILL = 1; ledger comment lines added adjacent |
| AC-E2E-024 | REQ-E2E-402, REQ-E2E-404 | Full template CI + build green after `make build` | `make build` exit 0; `go test ./internal/template/...` exit 0 |
| AC-E2E-025 | REQ-E2E-403 | Zero internal-content leaks in template artifacts | `grep -rn 'SPEC-E2E-REVIVAL\|REQ-E2E-\|AC-E2E-' internal/template/templates/` → 0 matches; neutrality CI test names pass within `go test ./internal/template/...` |

### Group F — Boundaries

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-026 | REQ-E2E-500, REQ-E2E-501 | No design-pack resurrection; skill dir count unchanged; agent respects subagent boundary | `ls internal/template/templates/.claude/skills \| grep -cE 'moai-domain-(ideation\|research\|brand-design\|copywriting\|design-handoff)\|moai-workflow-(design\|gan-loop)'` → 0; `expectedSkillCount = 28` test passes; AC-E2E-015 green |

## Given-When-Then Scenarios

### S1 — Web project detection (happy path)
- **Given** a user project with `playwright.config.ts` absent but `package.json` + `next.config.js` present
- **When** the user runs `/moai e2e`
- **Then** detection classifies `web`, the orchestrator presents toolchain options with "Playwright CLI (Recommended)" first, and upon selection the e2e-specialist receives the choice via spawn prompt (never prompting itself).

### S2 — React Native mobile project
- **Given** a project with `ios/` + `android/` + `react-native` in package.json dependencies
- **When** `/moai e2e` runs detection
- **Then** classification is `mobile`, Maestro is the recommended default, and Detox appears as an RN-conditional alternative in the option list with a factual trade-off description.

### S3 — Mixed monorepo
- **Given** a monorepo containing a Next.js web app and a Tauri desktop app
- **When** detection completes
- **Then** classification is `mixed`, and the orchestrator enumerates per-surface selections (web → Playwright default; desktop → WebdriverIO + tauri-service default) rather than forcing one global toolchain.

### S4 — Toolchain not installed
- **Given** Maestro selected but `maestro --version` probe fails
- **When** Phase 0 installation runs
- **Then** the exact install command is surfaced for approval, installation executes only after approval, and a re-probe confirms the version before Phase 1 begins.

### S5 — No e2e surface
- **Given** a pure Go library project (no web/mobile/desktop markers)
- **When** `/moai e2e` runs
- **Then** the workflow reports "no e2e target detected" with the marker evidence consulted, and exits without creating `e2e/` artifacts.

### S6 — Verbose output containment
- **Given** a Playwright run producing a 4,000-line failure log
- **When** the e2e-specialist reports results
- **Then** the in-context report carries exit code + ≤50-line tail + the full log's citable file path under the artifacts dir — never the full log inline.

### S7 — Router precedence
- **Given** the input `/moai e2e checkout flow on staging`
- **When** the Intent Router processes it
- **Then** Priority 1 matches `e2e` on the FIRST WORD after the command, routes to the e2e workflow, and "checkout flow on staging" passes through as context (not a routing signal).

### S8 — Agent boundary under missing input
- **Given** the e2e-specialist is spawned without a target URL for a web journey
- **When** it reaches the navigation step
- **Then** it returns a structured `## Missing Inputs` blocker report naming the parameter and stops — no free-form question, no AskUserQuestion attempt.

## Edge Cases

- **E-1 Tauri on macOS**: native `tauri-driver` route unsupported on macOS — the workflow must steer macOS Tauri projects to the embedded/`@wdio/tauri-service` mode (research.md §E).
- **E-2 Electron native dialogs**: OS-level dialogs bypass Playwright — the workflow's Electron section must include the main-process `evaluate()` mocking pattern.
- **E-3 Expo-managed RN**: `ios/`/`android/` dirs absent in managed workflow — detection must also read `app.json`/`expo` dependency markers.
- **E-4 CI/headless environment**: `CI=true` detection biases recommendation to Playwright CLI (web) and headless flags throughout; MCP-tier tools marked unavailable.
- **E-5 Flaky retry**: `--retry N` bounded (default 1); retries never re-run the full suite silently — only failed specs.
- **E-6 Simulator/emulator absence (mobile)**: Maestro/Appium probes must distinguish "CLI missing" from "no booted device/simulator" and report each with its own remedy.

## Quality Gates

- G1: `go test ./internal/template/...` exit 0 AND `go test ./...` exit 0 (full suite — no partial-suite success claims)
- G2: `golangci-lint run` exit 0
- G3: `moai spec lint --strict .moai/specs/SPEC-E2E-REVIVAL-001` → 0 findings
- G4: Template neutrality: AC-E2E-025 grep → 0 AND neutrality CI tests green
- G5: Subagent boundary: AC-E2E-015 greps → 0 violations
- G6: Both-tree parity: AC-E2E-013/014 diffs → exit 0

## Definition of Done

1. All 26 ACs PASS with executed-command evidence (verification-claim integrity: command + verbatim output per row; unexecuted rows are Gaps, not passes).
2. G1–G6 green in a single final verification batch.
3. `make build` completed after the last template edit; embedded FS carries all three artifacts (proven transitively by G1's catalog tests).
4. progress.md §E.2/§E.3 populated by manager-develop with the evidence table.
5. No modifications outside the declared surface: 3 template artifacts + 3 local siblings + 2 SKILL.md + 2 CLAUDE.md + catalog.yaml + 2 test constants (+ optional mirror-test registration per pre-flight #9).

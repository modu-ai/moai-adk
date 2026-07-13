# SPEC-E2E-REVIVAL-001 — Design

## §A Architecture — Three-Artifact Topology

```
/moai e2e <args>
   │  (thin command: .claude/commands/moai/e2e.md — frontmatter + Skill routing, <20 lines)
   ▼
moai SKILL.md Intent Router (Priority 1: **e2e** row)
   │
   ▼
workflow skill: .claude/skills/moai/workflows/e2e.md   ← orchestrator-facing playbook
   │  Phase 0   detection (delegated: e2e-specialist, read-only probe)
   │  Phase 0.5 selection (ORCHESTRATOR AskUserQuestion — never the agent)
   │  Phase 1   journey mapping (e2e-specialist)
   │  Phase 2   script creation (e2e-specialist)
   │  Phase 3   execution (e2e-specialist, CLI-first)
   │  Phase 4   recording — optional (e2e-specialist, native facility)
   │  Phase 5   report (orchestrator, conversation_language)
   ▼
agent: .claude/agents/moai/e2e-specialist.md           ← execution owner
   │  returns: bounded results + artifact paths | blocker report
   ▼
project artifacts: e2e/{*.spec.*, flows/*.yaml, traces/, recordings/, screenshots/}
```

Separation of concerns: the COMMAND is routing-only; the WORKFLOW owns UX flow + matrices + orchestrator instructions; the AGENT owns execution mechanics + token discipline. Distribution: all three authored in `internal/template/templates/` first (Template-First), synced to local, embedded via `//go:embed all:templates` at `make build`.

## §B Project-Type Detection Matrix (Phase 0)

Marker scan is read-only (Glob/Read), ecosystem-equal (no privileged language), and ordered most-specific-first. A project may match multiple rows → `mixed`.

| Platform class | Markers (any) | Notes |
|----------------|---------------|-------|
| `desktop` (electron) | `electron` in package.json deps; `electron-builder`/`forge` config | checked before generic web (an Electron repo also has package.json) |
| `desktop` (tauri) | `src-tauri/tauri.conf.json`; `tauri` in deps/Cargo.toml | Rust+web hybrid |
| `mobile` (react-native) | `react-native` in deps; `ios/` + `android/`; `app.json` with `expo` (managed workflow — edge E-3) | Detox becomes RN-conditional option |
| `mobile` (flutter) | `pubspec.yaml` with `flutter:` | flutter (canonical name) — Maestro supports Flutter; note in option description |
| `mobile` (native) | `*.xcodeproj`/`Package.swift` + iOS targets; `build.gradle` with `com.android.application` | Maestro/Appium capable |
| `web` | web framework configs (next/nuxt/vite/astro/sveltekit/angular), `index.html` servers, or any HTTP-serving app in the 16-language matrix (Django/Rails/Spring/Fiber/…) | broadest class; framework list is exemplary, detection is marker-driven not framework-privileged |
| `desktop-native` | native toolkit markers WITHOUT Electron/Tauri (e.g., pure `.xcodeproj` mac app, WinUI, Qt/GTK builds) | REQ-E2E-502 opt-in path |
| none | no markers above | REQ-E2E-007 graceful exit |

`mixed` handling: enumerate matched surfaces; per-surface selection questions (REQ-E2E-003).

## §C Platform-Toolchain Matrix + Recommendation Logic

| Platform | Default (Recommended) | Alternatives | MCP-tier (conditional only) |
|----------|----------------------|--------------|------------------------------|
| web | Playwright CLI 1.61+ | agent-browser (AI-exploratory) | chrome-devtools-mcp (perf/Lighthouse), Claude in Chrome (interactive debug) |
| mobile | Maestro 2.6+ | Appium 3.x (complex native flows), Detox 20.x (RN detected only) | — |
| desktop (electron) | Playwright `_electron` (experimental — caveat in prose) | — | — |
| desktop (tauri) | WebdriverIO + `@wdio/tauri-service` (embedded mode; macOS-safe) | Selenium via tauri-driver (Win/Linux only) | — |
| desktop-native | (opt-in) OS accessibility / computer-use | — | computer-use class, token-cost warning |

Recommendation modifiers (inherited from retired baseline, re-grounded):
- `CI=true` env → bias Playwright CLI / headless everywhere; MCP-tier marked unavailable
- `--record` → prefer toolchains with native trace/recording (Playwright trace, Maestro recording)
- explicit perf/Lighthouse ask → chrome-devtools-mcp becomes the recommended row FOR THAT CAPABILITY only
- All questions ride the orchestrator's AskUserQuestion; first option carries the locale-appropriate Recommended label; descriptions carry install-state + factual trade-offs (bias-prevention rule).

## §D e2e-specialist Agent Contract

Frontmatter draft (template-neutral prose; final wording at M1):

```yaml
---
name: e2e-specialist
description: |
  End-to-end test execution specialist for web, mobile, and desktop applications.
  Owns toolchain probing, journey script authoring, CLI-first test execution with
  bounded output, and artifact management under e2e/ directories.
  Use PROACTIVELY when the e2e workflow delegates detection, script creation, or execution.
  NOT for: implementation-cycle code changes (manager-develop), SPEC authoring (manager-spec),
  unit/integration test authoring within a TDD cycle (manager-develop), documentation (manager-docs).
tools: Read, Write, Edit, Bash, Grep, Glob, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill
model: inherit
effort: high
color: cyan
permissionMode: default
memory: project
skills:
  - moai-workflow-testing
---
```

Contract points:
- `tools:` CSV EXCLUDES `Agent` (no nesting — flat-hierarchy invariant) and `AskUserQuestion` (Frozen-zone subagent boundary). `mcp__context7` deliberately omitted: version lookups ride Skill("moai-workflow-testing") references and Bash probes.
- `model: inherit` (cache-aware spawn rule); `effort: high` (execution-heavy but not architecture-reasoning); `memory: project` for cross-session toolchain/journey recall.
- Body sections: (1) scope + phase responsibilities; (2) toolchain execution recipes per §C (probe/install/run/report commands); (3) [HARD] token-minimization ladder (§F); (4) blocker-report protocol (`## Missing Inputs` table, verbatim contract from agent-common-protocol); (5) artifact-directory conventions; (6) return contract (result summary schema: per-journey status, artifact paths, failure excerpts ≤ bounded tail).
- Body carries NO internal SPEC/REQ tokens (template neutrality).

## §E Workflow Skill Design (delta vs retired baseline)

Frontmatter: `user-invocable: false`; triggers keywords (e2e, end-to-end, browser test, mobile test, maestro, playwright, user journey…); `agents: ["e2e-specialist"]`; fresh version lineage.

| Phase | Retired (web-only) | Revived |
|-------|--------------------|---------|
| 0 | tool detection (4 web tools) | PROJECT-TYPE detection first (§B), then per-platform toolchain probe |
| 0.5 | AskUserQuestion tool selection | unchanged pattern; per-surface questions on `mixed`; `--tool` bypass |
| 1 | journey mapping via manager-develop | journey mapping via e2e-specialist; journey format inherited |
| 2 | script creation (spec.ts / agent tasks) | + Maestro flow YAML, Appium/WDIO specs, `_electron` fixtures, tauri-service config |
| 3 | execution (CLI or MCP) | CLI-first [HARD]; MCP escalation ladder; bounded-tail output contract |
| 4 | recording (traces/GIF) | native facilities only; MCP screenshot-loop recording removed |
| 5 | report | inherited template + artifact-path citation rule |

Flags: inherit all 8 retired flags; `--browser` applies to web only; add `--platform web|mobile|desktop` (forces classification when markers ambiguous).

Delegation directives: Phases 0, 1, 2, 3, 4 each carry "Delegate to the e2e-specialist subagent" (AC-E2E-016 requires ≥3 named references; design places 5).

## §F Token-Minimization Protocol (embedded in agent body as [HARD])

1. **Rung 1 — CLI + bounded tail**: every run command redirects full output to `e2e/.runs/<timestamp>-<slug>.log`; context receives `exit=<code>` + `tail -50` (and ≤2KB). The log path is cited in the report.
2. **Rung 2 — structured reporters**: on failure triage, prefer `--reporter=json`-class output parsed selectively (failed specs only) over re-running with verbose flags.
3. **Rung 3 — MCP, batched, capability-gated**: only for capabilities the matrix marks CLI-impossible (live perf insight, Lighthouse-class audit, interactive debug). Batch calls; prefer snapshot/aggregate tools; never per-element polling loops.
- Artifacts (HTML reports, traces, screenshots, recordings) are NEVER inlined — paths only.
- Ladder violations are the agent's own §self-check items before returning.

## §G Distribution Design

Order of operations (single logical change-set; M1→M4):
1. Template tree: agent → workflow → command (`.md.tmpl` replicating sibling render pattern)
2. `catalog.yaml`: manual `core.agents` entry — `name: e2e-specialist / tier: core / path: templates/.claude/agents/moai/e2e-specialist.md / hash: <placeholder> / version: 1.0.0`
3. CI constants: `expectedAgentCount` 9→10 (+ ledger comment), `expectedTotal` 37→38 (+ comment); `expectedSkillCount` untouched
4. `make build` → gen-catalog-hashes `--all` fills the real hash + refreshes the changed moai-skill hash; binary re-embeds
5. Local sync: byte-identical copies for agent + workflow; rendered command; SKILL.md + CLAUDE.md edits applied to BOTH trees in the same commit
6. Mirror-registration decision per plan.md pre-flight #9 (follow sibling precedent)

## §H Test/CI Mapping

| Gate | Proves |
|------|--------|
| `TestCommandsThinPattern` / `TestCommandsFrontmatterConsistency` | REQ-E2E-200 (thin body, CSV tools, no deprecated fields) |
| `TestAgentFrontmatterAudit` | REQ-E2E-202 frontmatter cleanliness |
| `TestAllAgentsInCatalog` + `TestCatalogReferencesValid` | REQ-E2E-304 catalog membership + path resolution (embedded FS) |
| `expectedAgentCount`/`expectedTotal` assertions | REQ-E2E-401 reconciliation |
| `expectedSkillCount = 28` still passing | REQ-E2E-500 (no skill-dir additions/resurrections) |
| neutrality CI tests + AC-E2E-025 grep | REQ-E2E-403 |
| boundary greps (AC-E2E-015) | REQ-E2E-203/501 |
| parity diffs (AC-E2E-013/014) | REQ-E2E-400 both-tree consistency |
| `go test ./...` + `golangci-lint run` | repo-wide regression freedom |

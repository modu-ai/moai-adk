# SPEC-E2E-REVIVAL-001 — Acceptance Criteria

> Every AC verifies REACHABILITY (existence + cross-file reference + executable gate), never token presence alone. Each AC names its executable check; run-phase evidence cites the executed command + verbatim output path. Baselines were measured 2026-07-13 (spec.md §A table) — re-verify at run entry.
>
> **Table-cell commands are executed VERBATIM.** Commands containing pipes (shell `|` or regex alternation) MUST NOT live inside markdown table cells — cell escaping (`\|`) renders regex alternation and shell pipes vacuous (empirically confirmed, audit iter-1 D3). All pipe-bearing commands live in § Executable Command Block below and are referenced by CMD-ID from the cells.

## §D AC Matrix

### Group A — Detection & selection (content ACs on the workflow skill body)

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-001 | REQ-E2E-001 | Workflow skill contains a project-marker detection matrix covering web, mobile, desktop, and mixed classifications, with ≥2 marker examples per platform class | Read `internal/template/templates/.claude/skills/moai/workflows/e2e.md` §Detection; matrix table present with 4 platform rows; markers treat ecosystems equally (no "primary" language) |
| AC-E2E-002 | REQ-E2E-002 | Platform-toolchain matrix maps each detected type to a default: web→Playwright CLI, mobile→Maestro, desktop→Playwright-Electron/WebdriverIO+tauri-service | Matrix table present; defaults match design.md §C; each default's install + version-probe command included |
| AC-E2E-003 | REQ-E2E-003, REQ-E2E-004 | All selection questions are specified as ORCHESTRATOR AskUserQuestion instructions; zero instructions direct the e2e-specialist to prompt | `grep -n 'AskUserQuestion' <workflow>` — every hit sits in orchestrator-addressed prose; `grep -c 'specialist.*AskUserQuestion'` semantic review → 0 agent-prompts directives |
| AC-E2E-004 | REQ-E2E-005 | `--tool` flag documented in Supported Flags and short-circuits selection | Flag row present; Phase 0.5 contains the bypass branch |
| AC-E2E-005 | REQ-E2E-006 | Missing-toolchain path: version probe → install command surface → approval → re-probe | Phase 0 install section contains probe-install-reprobe sequence for each default toolchain |
| AC-E2E-006 | REQ-E2E-007 | No-target graceful exit specified; a detected `desktop-native` surface routes to the SAME graceful branch with a deferral notice (native-desktop automation deferred per user decision — former REQ-E2E-502 removed at v0.1.1) | Workflow contains the "no e2e target" report branch incl. the `desktop-native` deferral notice; NO opt-in automation path present for `desktop-native` |

### Group B — Token minimization

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-007 | REQ-E2E-100 | CLI-first rule stated as [HARD] in BOTH the workflow skill and the agent body, with the per-capability CLI-vs-MCP tool matrix in the workflow | `grep -n '\[HARD\]' <workflow> <agent>` includes the CLI-first rule; matrix has token-cost column |
| AC-E2E-008 | REQ-E2E-101 | Agent body carries the bounded-tail + file-redirect contract (≤50 lines OR ≤2KB; artifacts dir; citable paths) | Run **CMD-008** → ≥1; redirect example present |
| AC-E2E-009 | REQ-E2E-102 | MCP batching rule (snapshot/batch over per-element round-trips) present in agent body | Section present in agent §token-minimization ladder rung 3 |
| AC-E2E-010 | REQ-E2E-103, REQ-E2E-104 | Report/trace/recording artifacts persist under `e2e/` dirs, paths cited not inlined; `--record` uses native toolchain facility | Workflow Phase 4/5 specify artifact dirs + path-citation rule; no MCP-screenshot-loop recording path |
| AC-E2E-011 | REQ-E2E-105 | Every default platform path executable CLI-only; no MCP server hard dependency | Tool matrix: each platform's DEFAULT row is CLI-class; MCP rows all marked conditional |

### Group C — Deliverable artifacts (existence + CI + boundary)

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-012 | REQ-E2E-200 | Thin command exists in both trees and passes thin-pattern CI | `test -f internal/template/templates/.claude/commands/moai/e2e.md.tmpl && test -f .claude/commands/moai/e2e.md`; `go test -run TestCommandsThinPattern ./internal/template/` exit 0. NOTE: the test walks the EMBEDDED template FS only — the audited `commands/moai` set goes 13 → 14 `.md.tmpl` files; the local `.md` sibling is proven by the `test -f` check only, never by the CI test |
| AC-E2E-013 | REQ-E2E-201 | Workflow skill exists in both trees, `user-invocable: false`, identical content | `diff .claude/skills/moai/workflows/e2e.md internal/template/templates/.claude/skills/moai/workflows/e2e.md` exit 0; `grep -c 'user-invocable: false'` = 1 |
| AC-E2E-014 | REQ-E2E-202 | Agent file exists in both trees, byte-identical, frontmatter complete (name/description/tools CSV/model/effort/color/permissionMode/memory/skills) | `diff` exit 0; `grep -c '^tools:'` = 1 and value is CSV (no leading `-`); frontmatter COMPLETENESS verified by MANUAL field-by-field checklist — no CI test proves completeness; `TestAgentFrontmatterAudit` (run additionally, exit 0) guards retired-field cleanliness ONLY (audit iter-1 D12) |
| AC-E2E-015 | REQ-E2E-203 | Agent tools line excludes Agent and AskUserQuestion; body contains the `## Missing Inputs` blocker-report contract | Run **CMD-015** → 0; `grep -c 'Missing Inputs' <agent>` ≥1 |
| AC-E2E-016 | REQ-E2E-204 | Cross-file reachability chain: workflow delegates to `e2e-specialist` by name in ≥3 phase sections AND the agent file exists at the resolving path in both trees AND catalog registers it | `grep -c 'e2e-specialist' <workflow>` ≥3; `test -f` both agent paths; run **CMD-016** → ≥2 `=== RUN` lines AND exit 0 (both tests provably executed — a non-matching `-run` pattern exits 0 vacuously) |

### Group D — Router & catalog reachability (baseline-delta greps, both trees)

| AC | REQ | Criterion | Executable check (baseline → target) |
|----|-----|-----------|--------------------------------------|
| AC-E2E-017 | REQ-E2E-300 | Priority 1 router row restored in both SKILL.md trees | `grep -cE '^- \*\*e2e\*\*' <both SKILL.md>` : 0 → 1 each |
| AC-E2E-018 | REQ-E2E-301 | Frontmatter description enumeration + CLAUDE.md §3 Subcommands line include `e2e`, both trees | `grep -c 'e2e' <SKILL.md frontmatter block>` 0 → ≥1; run **CMD-018** : 0 → 2 |
| AC-E2E-019 | REQ-E2E-302 | FULL in-scope count-literal surface updated per the spec.md REQ-E2E-302 rings: ring 1 doctrine (12 files/24 sites, both trees) + ring 2 template-skill modules (4 template files/8 sites, incl. the agents-reference.md `e2e-specialist` table row) with the two-generations-stale LOCAL siblings (8/7-era, 7 sites) normalized + ring 3 README.md (3 sites; ko/ja/zh locale-language review) | `grep -c '11 retained agents'` 0 → 1 per CLAUDE.md (2 across both trees); `grep -c 'e2e-specialist' <both CLAUDE.md>` 0 → ≥2 each (table row + decision tree); INVARIANCE: run **CMD-019-INV** (19-file widened surface) → 0 (measured baseline 2026-07-13: 38) AND **CMD-019-INV-B** (local skill siblings, 8/7-era family) → 0 (measured baseline: 7) |
| AC-E2E-020 | REQ-E2E-303 | Priority 3 semantic-classification cue line for e2e-testing intent added, both trees | `grep -n 'e2e' <SKILL.md P3 section>` ≥1 each; cue line phrased as semantic exemplar (not literal-match requirement) |
| AC-E2E-021 | REQ-E2E-304 | catalog.yaml core.agents entry with real (non-placeholder) 64-hex hash | `grep -A4 'name: e2e-specialist' internal/template/catalog.yaml` shows tier/path/hash/version; hash matches `^[0-9a-f]{64}$`; `go test -run TestAllAgentsInCatalog ./internal/template/` exit 0 |
| AC-E2E-028 | REQ-E2E-305 | Go tier-profile display surface includes `e2e-specialist` (renders in the `moai web` model-policy preview) with all pins reconciled | `grep -c 'e2e-specialist' internal/template/model_policy.go` ≥1 (order list + profile entries); run **CMD-028** → ≥2 `=== RUN` lines AND exit 0 with the updated pins (length 11; 66-cell assertion; per-plan rows 11); stale "10 retained agents" comments in both test files are covered by CMD-019-INV (they are in its 19-file scope) |

### Group E — Distribution & CI reconciliation

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-022 | REQ-E2E-400 | Template-First provenance: every new local file has a template sibling; commit order/content shows template tree authored in the same commit as (or before) local | `git show --stat <run commits>` — no local-only additions; parity diffs (AC-013/014) green |
| AC-E2E-023 | REQ-E2E-401 | Count constants reconciled with provenance comments | `grep -n 'expectedAgentCount = 10' catalog_tier_audit_test.go` = 1; `grep -n 'expectedTotal = 38' catalog_loader_test.go` = 1; `grep -n 'expectedSkillCount = 28'` STILL = 1; ledger comment lines added adjacent |
| AC-E2E-024 | REQ-E2E-402, REQ-E2E-404 | Full template CI + build green after `make build` | `make build` exit 0; `go test ./internal/template/...` exit 0 |
| AC-E2E-025 | REQ-E2E-403 | Zero internal-content leaks in template artifacts | Run **CMD-025** → 0 matches; neutrality CI test names pass within `go test ./internal/template/...` |
| AC-E2E-027 | REQ-E2E-405 | Detection-matrix ecosystem equality: no language/framework presented as privileged; marker coverage even across platform classes | Manual matrix review: ≥2 marker examples per platform class AND the web class documented as marker-driven (not framework-privileged); run **CMD-027** → 0 privileging-phrase matches in the workflow detection section |

### Group F — Boundaries

| AC | REQ | Criterion | Executable check |
|----|-----|-----------|------------------|
| AC-E2E-026 | REQ-E2E-500, REQ-E2E-501 | No design-pack resurrection; skill dir count unchanged; agent respects subagent boundary | Run **CMD-026** → 0 (supportive absence check); the LOAD-BEARING guard is the `expectedSkillCount = 28` test passing; AC-E2E-015 green |

## Executable Command Block (verbatim — referenced by CMD-ID from table cells)

Pipe-bearing commands cannot live inside markdown table cells: cell escaping (`\|`) turns regex alternation into a literal-pipe match and shell pipes into broken arguments, making the checks vacuous (audit iter-1 D3, empirically confirmed). Run these verbatim from the repo root:

```bash
# CMD-008 — bounded-tail contract presence in agent body, both trees (expected output: >=1 per file)
grep -cnE '50 lines|2KB' .claude/agents/moai/e2e-specialist.md \
  internal/template/templates/.claude/agents/moai/e2e-specialist.md

# CMD-015 — agent tools-line boundary (expected output: 0)
grep -h '^tools:' .claude/agents/moai/e2e-specialist.md \
  internal/template/templates/.claude/agents/moai/e2e-specialist.md \
  | grep -cE '\bAgent\b|AskUserQuestion'

# CMD-016 — catalog reachability tests provably RUN and pass
# (expected: first command prints >=2; second prints exit=0)
go test -v -run 'TestAllAgentsInCatalog|TestCatalogReferencesValid' ./internal/template/ | grep -c '^=== RUN'
go test -run 'TestAllAgentsInCatalog|TestCatalogReferencesValid' ./internal/template/; echo "exit=$?"

# CMD-018 — CLAUDE.md §3 Subcommands line gains e2e in both trees (expected output: 2)
grep 'Subcommands:' CLAUDE.md internal/template/templates/CLAUDE.md | grep -c e2e

# CMD-019-INV — stale count-literal invariance over the 19-file widened REQ-E2E-302 surface
# (rings 1-3 + Go test files; measured baseline 2026-07-13: 38; expected post-change output: 0)
grep -rEn '10 retained agents|9 MoAI-custom|10-agent' \
  CLAUDE.md internal/template/templates/CLAUDE.md \
  .claude/rules/moai/development/agent-authoring.md \
  internal/template/templates/.claude/rules/moai/development/agent-authoring.md \
  .claude/rules/moai/development/agent-patterns.md \
  internal/template/templates/.claude/rules/moai/development/agent-patterns.md \
  .claude/rules/moai/development/model-policy.md \
  internal/template/templates/.claude/rules/moai/development/model-policy.md \
  .claude/rules/moai/workflow/spec-workflow.md \
  internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \
  .claude/agents/moai/manager-design.md \
  internal/template/templates/.claude/agents/moai/manager-design.md \
  internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md \
  internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md \
  internal/template/templates/.claude/skills/moai-foundation-core/modules/INDEX.md \
  internal/template/templates/.claude/skills/moai-foundation-quality/SKILL.md \
  README.md \
  internal/template/model_policy_test.go \
  internal/web/modelpolicy_test.go \
  | wc -l

# CMD-019-INV-B — local skill siblings, two-generations-stale 8/7-era family
# (measured baseline 2026-07-13: 7; expected post-change output: 0)
grep -rEn '8 retained agents|7 MoAI-custom|8-agent' \
  .claude/skills/moai-foundation-core/SKILL.md \
  .claude/skills/moai-foundation-core/modules/agents-reference.md \
  .claude/skills/moai-foundation-core/modules/INDEX.md \
  .claude/skills/moai-foundation-quality/SKILL.md \
  | wc -l

# CMD-025 — template neutrality leak scan (expected output: 0)
grep -rn 'SPEC-E2E-REVIVAL|REQ-E2E-|AC-E2E-' internal/template/templates/ -E | wc -l

# CMD-026 — retired design-pack absence (expected output: 0; supportive — expectedSkillCount=28 is the real guard)
ls internal/template/templates/.claude/skills \
  | grep -cE 'moai-domain-(ideation|research|brand-design|copywriting|design-handoff)|moai-workflow-(design|gan-loop)'

# CMD-027 — detection-matrix privileging-phrase scan (expected output: 0)
# Scope: the Detection section of the workflow skill (both trees are identical per AC-E2E-013)
sed -n '/Phase 0/,/Phase 0.5/p' internal/template/templates/.claude/skills/moai/workflows/e2e.md \
  | grep -icE 'primary (language|framework)|first-class (language|framework)|enabled.*planned'

# CMD-028 — Go tier-profile display pins provably RUN and pass with e2e-specialist row
# (expected: first command prints >=2; second prints exit=0)
go test -v -run 'TierProfile' ./internal/template/ ./internal/web/ | grep -c '^=== RUN'
go test -run 'TierProfile' ./internal/template/ ./internal/web/; echo "exit=$?"
```

Note: `grep -c` prints `0` and exits 1 on no-match — the EXPECTED OUTPUT VALUE is the criterion, not the exit code, for the absence checks (CMD-015, CMD-019-INV, CMD-019-INV-B, CMD-025, CMD-026, CMD-027). CMD-008 is a presence check (≥1 per file).

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
- G3: `moai spec lint --strict .moai/specs/SPEC-E2E-REVIVAL-001/spec.md` → 0 ERRORS (FILE argument — a directory argument fails with `ParseFailure … is a directory`; iter-1 D7). The pre-existing `StatusGitConsistency` WARNING — frontmatter `draft`/`in-progress` vs git-implied status from this SPEC's own `feat()` commits — is EXPECTED until sync close and does NOT fail this gate (iter-2 D14). Do NOT suppress it via `lint.skip`.
- G4: Template neutrality: AC-E2E-025 grep → 0 AND neutrality CI tests green
- G5: Subagent boundary: AC-E2E-015 greps → 0 violations
- G6: Both-tree parity: AC-E2E-013/014 diffs → exit 0

## Definition of Done

1. All 28 ACs PASS with executed-command evidence (verification-claim integrity: command + verbatim output per row; unexecuted rows are Gaps, not passes).
2. G1–G6 green in a single final verification batch.
3. `make build` completed after the last template edit; embedded FS carries all three artifacts (proven transitively by G1's catalog tests).
4. progress.md §E.2/§E.3 populated by manager-develop with the evidence table.
5. No modifications outside the declared surface: 3 template artifacts + 3 local siblings + 2 SKILL.md + 2 CLAUDE.md + the ring-1 count-literal rule/agent files ×2 trees (agent-authoring.md, agent-patterns.md, model-policy.md, spec-workflow.md, manager-design.md) + the ring-2 skill-module files ×2 trees (moai-foundation-core SKILL.md / modules/agents-reference.md / modules/INDEX.md + moai-foundation-quality SKILL.md) + README.md (+ ko/ja/zh only if locale-language count claims are found) + the REQ-E2E-305 Go display surface (model_policy.go + model_policy_test.go + internal/web/modelpolicy_test.go) + catalog.yaml + 2 test constants (+ optional mirror-test registration per pre-flight #9 — audit-verified precedent: sibling top-level workflow files are NOT in `workflowOptMirroredPaths`, so the default resolution is no-register).

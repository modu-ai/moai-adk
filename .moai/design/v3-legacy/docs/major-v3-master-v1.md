# MoAI-ADK v3.0.0 — Master Design Document

> Status: DRAFT (Wave 3 — architect synthesis)
> Based on: Claude Code source analysis (1,903 files, 33MB) + moai-adk-go current state snapshot
> Date: 2026-04-22
> Inputs: `.moai/design/v3-research/gap-matrix.md` (194 gaps), `priority-roadmap.md` (56 tiered items), `v3-themes.md` (9 themes), 6 Wave 1 findings files
> Author directive: Go-idiomatic, evidence-grounded, SPEC-ready

---

## Executive Summary

MoAI-ADK v3.0.0 is a capability-parity and foundation release. It closes the largest delta between moai-adk-go and Claude Code's native subsystems (Hook Protocol v2, 4-type Memory, layered settings, plugin manifest) while laying the structural foundations — formal schemas, versioned migrations, discriminated-union team mailbox — that make every subsequent release safer. v3.0.0 ships nine architectural themes (detailed in §3), anchored on 28 proposed SPECs carrying 545 REQs (§8), and intentionally rejects thirteen tempting-but-wrong scopes (§10) to keep the release shippable.

The release is necessary now because moai-adk-go has crossed the scale threshold where ad-hoc configuration no longer works: 104 active SPECs, 50 skills, 22 agents, 502 embedded template files, 35 rule files, and 22 YAML config sections. Claude Code itself ships a versioned migration framework (`CURRENT_MIGRATION_VERSION` counter, `preAction` hook — W1.5 §5.1) and JSON-Schema-validated settings (Zod v4, W1.5 §6.1); moai-adk validates nothing. The cost of introducing schemas and migrations grows linearly with every new SPEC shipped. v3.0.0 is the last cheap moment.

Breaking change scope is deliberately narrow: eight named breaking changes (BC-001 through BC-008, §4), every one paired with a dual-parse shim or opt-out env var, and all concentrated in low-cardinality surfaces (hook authors, team-mode users, config maintainers). User-visible migration effort is one command: `moai migrate v2-to-v3 --dry-run` (§5.1). v3.0 preserves moai's identity verbatim — SPEC-First DDD, TRUST 5, TAG system, @MX protocol, the 50-skill / 22-agent catalog, harness-based quality routing, and the hybrid design system all ship unchanged in behavior.

---

## 1. Vision & Goals

### 1.1 Product Vision

MoAI-ADK continues to be the strategic orchestrator layer for Claude Code, specialized for SPEC-First DDD and TDD engineering teams. v3.0.0 reframes this vision around three commitments:

1. **CC-native integration**: moai runs inside Claude Code v2.1.111+ and speaks every authoritative CC subsystem (hooks, memory, settings, plugins, output) with full protocol fidelity. Where CC has a schema, moai validates against it. Where CC has a hook reply contract, moai emits the reply shape.
2. **Schema-first configuration**: every `.moai/config/sections/*.yaml` file, every hook declaration, every migration step has a formal Go struct + validator/v10 tags + auto-generated JSON Schema. No more silent typos.
3. **Additive moai differentiation**: SPEC-to-SPEC chaining, SPEC templates, SPEC lifecycle transitions, and the harness-based quality routing are moai-unique — v3 formalizes them as first-class primitives, not afterthoughts.

### 1.2 Design Principles (v3 refined)

Anchored in W1.1–W1.5 observed CC patterns and moai's existing constitution (`.claude/rules/moai/core/moai-constitution.md`):

- **Evidence over inference** — every behavior choice cites a Wave 1 finding (file:line) or a moai-adk source anchor.
- **Go-idiomatic, minimal-deps** — continue the 9-direct-dep philosophy (W1.6 §14.1). Reject solutions requiring net-new heavy dependencies (OTEL counters, CUE, jsonrpc2 direct). Exception: `go-playground/validator/v10` for schemas (Tier 1 critical).
- **One-turn fully-loaded agent prompts** — inherit the Opus 4.7 prompt philosophy already ratified (SPEC-OPUS47-COMPAT-001, W1.6 §16 git log).
- **Reversible changes** — every breaking change has a dual-parse shim or opt-out env var for one minor version.
- **Verify don't assume** — ship `moai doctor config --fix`, `moai doctor migration --dry-run`, `moai doctor hook --validate` for all v3 additions.

### 1.3 Success Metrics (measurable)

- **MIG-PASS**: `moai migrate v2-to-v3 --dry-run` completes with zero destructive operations on a corpus of 10 representative v2.12 projects.
- **HOOK-PARITY**: 100% of CC's 27 hook events have moai handlers returning schema-valid `HookOutput` (target: match W1.1 §1 catalog 1:1).
- **SCHEMA-COVERAGE**: 22/22 YAML config sections have Go struct + validator tags + exported JSON Schema.
- **AGENT-FRONTMATTER**: 22/22 existing agents compile under v2 frontmatter schema with zero validation errors.
- **TEMPLATE-DRIFT-ZERO**: Local `.claude/` and template tree byte-identical except for explicitly-protected runtime-managed files (per CLAUDE.local.md §2).
- **BINARY-SIZE**: `bin/moai` grows by less than 4 MB net from v2.12.0 baseline.
- **TEST-COVERAGE**: `internal/` packages hold 85%+ line coverage (moai-constitution HARD rule).
- **DOCS-4LOCALE**: `docs-site/content/{ko,en,ja,zh}/` section parity — zero missing sections per locale (fixes gap matrix #194).

---

## 2. Current State Analysis

### 2.1 Strengths to Preserve

From Wave 1.6 (`findings-wave1-moai-current.md`):

- **Strong agent catalog** (W1.6 §6): 22 agents organized into Managers (8) / Experts (8) / Builders (3) / Evaluators (2) / Researcher (1); consistent frontmatter fields (`name`, `description`, `tools`, `model`, `permissionMode`, `memory`, `skills`, `hooks`) across all 22.
- **Skill ecosystem depth** (W1.6 §7): 50 skills spanning `moai-foundation-*`, `moai-workflow-*`, `moai-domain-*`, `moai-platform-*`, `moai-framework-*`, `moai-library-*`, `moai-ref-*`, `moai-design-*`, `moai-docs-*`. 320 skill markdown files total — Progressive Disclosure Level 2 works.
- **SPEC-First DDD maturity** (W1.6 §11): 104 active SPECs, triple-file pattern (`spec.md` / `plan.md` / `acceptance.md`), EARS requirements format, YAML frontmatter with `spec_id`/`phase`/`lifecycle`.
- **Thin command pattern** (W1.6 §8, CLAUDE.local.md §3): 13 `/moai` commands all under 20 LOC body, enforced by `internal/template/commands_audit_test.go`.
- **Hook registry core** (W1.6 §5.1): `internal/hook/registry.go` with 27 `EventType` constants, lazy TraceWriter initialization, `@MX:ANCHOR fan_in=20+`.
- **Template system robustness** (W1.6 §4): `go:embed all:templates`, 502 embedded files, provenance-aware deployer (`TemplateManaged` / `UserCreated` / `UserModified`), `validateDeployPath` security.
- **Hybrid design system** (`.claude/rules/moai/design/constitution.md` v3.3.0): FROZEN/EVOLVABLE zones, GAN Loop contract, Sprint Contract Protocol.
- **Multi-locale docs** (W1.6 §12): Hugo+Hextra, 4 locales (`ko/en/ja/zh`), `docs-site/scripts/translate.mjs` AI translation, Vercel deployment.
- **Minimal dependency budget** (W1.6 §14): 9 direct deps, 23 indirect — intentional simplicity.

### 2.2 Weaknesses to Address

From gap-matrix.md and Wave 1.6 §15:

- **Hook protocol poverty** (gm#4, #5, #6 — Critical): exit-code-only output; no `hookSpecificOutput`, `additionalContext`, `updatedInput`, `systemMessage`.
- **Zero formal schemas** (gm#156, #157, #163 — Critical): `.moai/config/sections/*.yaml` parsed via `gopkg.in/yaml.v3` with struct tags only; typos silently ignored.
- **One-shot migrations** (gm#149 — Critical): only `moai migrate agency` exists; no `CURRENT_MIGRATION_VERSION` counter; no `preAction` runner.
- **Flat settings tier** (gm#164 — High): single `.moai/config/sections/` level; CC has 6-tier precedence (user / project / local / flag / policy / managed).
- **Ad-hoc team JSON** (gm#71 — Medium): `TeamCreate`/`SendMessage` payloads are untyped; CC has 10 Zod-validated message types.
- **MEMORY.md unbounded** (gm#44 — Critical): CLAUDE.md has 40K char cap but MEMORY.md does not; CC enforces 200 lines / 25KB with warning appendix.
- **Memory freshness missing** (gm#50 — High): stale memories (>1 day) propagate without staleness marker.
- **Agent frontmatter gaps** (gm#56–#63 — High): missing `initialPrompt`, `omitClaudeMd` (5–15 Gtok/week savings per CC BQ data, W1.3 §4.2), `requiredMcpServers`, `maxTurns`.
- **No plugin system** (gm#140 — High): zero third-party extension path.
- **Output rendering fidelity** (gm#175, #176, #177 — Medium): moai emits plain text; CC has StructuredDiff, ValidationErrorsList, ProgressBar components ready to render structured output.

### 2.3 Technical Debt Inventory

Self-identified in Wave 1.6 §15:

1. **project.yaml.template_version drift** (W1.6 §15.4, §15.8, gm#183): `project.yaml:template_version: v2.7.22` vs `system.yaml:moai.version: v2.12.0` — 12 minor versions stale. `moai update` silently skips this field.
2. **Template/local skill drift** (W1.6 §7.1, §15.1, gm#184): 3 template-only skills (`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`) not deployed to local.
3. **Hook wrapper drift** (W1.6 §5.5, §15.7, gm#185): `handle-permission-denied.sh` missing from local (`.tmpl` exists in template).
4. **Pre-refactor backups** (W1.6 §15.10): `internal/cli/glm.go.bak` (28,567 bytes), `internal/cli/worktree/new_test.go.bak` (13,700 bytes). Checked in; should be removed.
5. **Stale coverage artifacts** (W1.6 §13.2, §15.2, gm#188): `coverage.out` / `coverage.html` dated 2026-03-11. Predates many SPEC changes.
6. **ADR-011 comment drift** (W1.6 §4.6, §15.3, gm#189): `internal/template/embed.go:8-12` claims runtime-generated files excluded, but `.claude/settings.json.tmpl` IS embedded and rendered.
7. **.agency/ vestigial** (W1.6 §15.6, gm#190): stub redirects in `.claude/commands/agency/*.md` (8 files) + `.claude/rules/agency/constitution.md` (695-byte stub) + `.moai-backups/` folder (April migration snapshot).
8. **MCP pencil unconfigured in project scope** (W1.6 §16.2, gm#192): `expert-frontend.md` declares 14 `mcp__pencil__*` tools but `.mcp.json` does not list `pencil`.
9. **Handler count drift** (W1.6 §5.3, §15.5): `InitDependencies()` registers 28 handler calls vs 27 `EventType` constants. `AutoUpdateHandler` is second SessionStart handler via compose pattern; intentional but undocumented.
10. **docs-site locale lag** (W1.6 §12.2, gm#194): `docs-site/content/en/` lacks `contributing/` and `multi-llm/` sections present in `ko/`.

---

## 3. v3 Architecture Themes

Each theme below includes: problem statement, design approach, API/schema sketch (Go-idiomatic), breaking change impact (BC-IDs defined in §4), migration path, and SPEC IDs (defined in §8).

### 3.1 Theme 1 — Hook Protocol v2

#### Problem statement

moai-adk-go hook handlers today emit only exit codes (0 / 2 / other — W1.6 §5.1, `internal/hook/types.go:19-114`). Claude Code's authoritative contract (W1.1 §4, `utils/hooks.ts:747-1335` + `types/hooks.ts:211-226`) expects a rich `HookJSONOutput` with `decision`, `hookSpecificOutput`, `additionalContext`, `updatedInput`, `updatedMCPToolOutput`, `systemMessage`, `suppressOutput`, `stopReason`, `continue`, `watchPaths`. Without this, moai hooks cannot:
- Inject model-turn context (`additionalContext` — W1.1 §14.3)
- Rewrite tool input before execution (`updatedInput`)
- Participate in permission dialogs (`PermissionRequest.decision` — W1.1 §2.1)
- Register `FileChanged` watches (`SessionStart.watchPaths` — W1.1 §2.1)
- Block compaction or config changes (`PreCompact`, `ConfigChange` — W1.1 §2.1)

#### Design approach

1. Extend `internal/hook/types.go` `HookOutput` struct to match CC's `HookJSONOutput` shape (already partially exists — W1.6 §5.7 lists `Continue, StopReason, SystemMessage, SuppressOutput, Decision, Reason, HookSpecificOutput, UpdatedInput, Retry, ExitCode`). Add missing: `AdditionalContext`, `UpdatedMCPToolOutput`, `WatchPaths`, `InitialUserMessage`, `NewCustomInstructions`, `UserDisplayMessage`.
2. Introduce `HookSpecificOutput` as discriminated union tagged by `hookEventName` (Go: interface with `HookEventName() string` method + type-switched marshalers).
3. Add hook source precedence pipeline (3-tier v3.0: user / project / local). Defer policy / plugin / skill / session / builtin tiers to v3.2 (see §9 open question #4).
4. Add `if` condition filter pre-spawn (permission-rule syntax `Bash(git *)`, `Read(*.ts)` — W1.1 §3.8). Implemented as a lightweight condition evaluator in `internal/hook/condition.go` reusing existing permission rule parsing.
5. Add `async: true` / `asyncRewake: true` hook flags with a Go `AsyncHookRegistry` equivalent (`internal/hook/async_registry.go`). Completion signaled via `continue` field in stdout JSON.
6. Add `once: true` self-removing hook flag with source-of-truth bookkeeping in `.moai/state/hook-once.json`.
7. Add `CLAUDE_ENV_FILE` mechanism: SessionStart / Setup / CwdChanged / FileChanged hook handlers write to `$CLAUDE_ENV_FILE` temp file; moai's BashTool wrapper sources the file before invocation (W1.1 §4.3).
8. Upgrade 6 observational handlers (PermissionRequest, PermissionDenied, PreCompact, ConfigChange, InstructionsLoaded, WorktreeCreate, Elicitation/ElicitationResult) to emit structured outputs per CC schema (W1.1 §2.1).
9. `WorktreeCreate` moves from observational to PROVIDER contract: stdout MUST contain absolute path to created worktree, non-zero exit = failed (W1.1 §14.4, gm#24).

#### API/schema sketch

```go
// internal/hook/types.go (v3)

// HookOutput is the rich JSON reply from a moai hook wrapper.
// Maps 1:1 to CC's HookJSONOutput (entrypoints/sdk/coreSchemas.ts:806-935).
type HookOutput struct {
    // Flow control
    Continue      *bool   `json:"continue,omitempty"`
    StopReason    string  `json:"stopReason,omitempty"`
    SystemMessage string  `json:"systemMessage,omitempty"`
    SuppressOutput bool   `json:"suppressOutput,omitempty"`

    // Legacy: inherit from v2 (deprecated in v3.2, removed v4.0)
    Decision string `json:"decision,omitempty"`  // allow | deny | ask | block
    Reason   string `json:"reason,omitempty"`
    ExitCode int    `json:"-"`  // synthesized from Decision for backcompat

    // Event-specific payload (discriminated by HookEventName)
    HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput is the discriminated union (per event).
type HookSpecificOutput interface {
    HookEventName() string
}

// Concrete variant for PreToolUse.
type PreToolUseOutput struct {
    EventName                string                `json:"hookEventName"` // "PreToolUse"
    PermissionDecision       string                `json:"permissionDecision,omitempty"` // allow|deny|ask
    PermissionDecisionReason string                `json:"permissionDecisionReason,omitempty"`
    UpdatedInput             map[string]any        `json:"updatedInput,omitempty"`
    AdditionalContext        string                `json:"additionalContext,omitempty"`
}

func (PreToolUseOutput) HookEventName() string { return "PreToolUse" }

// SessionStart — watchPaths is UNIQUE to this event.
type SessionStartOutput struct {
    EventName          string   `json:"hookEventName"` // "SessionStart"
    AdditionalContext  string   `json:"additionalContext,omitempty"`
    InitialUserMessage string   `json:"initialUserMessage,omitempty"`
    WatchPaths         []string `json:"watchPaths,omitempty"` // abs paths
}

// PermissionRequest — decision is a nested allow/deny object.
type PermissionRequestOutput struct {
    EventName string                     `json:"hookEventName"` // "PermissionRequest"
    Decision  *PermissionRequestDecision `json:"decision,omitempty"`
}

type PermissionRequestDecision struct {
    Behavior           string           `json:"behavior"` // allow | deny
    UpdatedInput       map[string]any   `json:"updatedInput,omitempty"`       // allow-only
    UpdatedPermissions []PermissionRule `json:"updatedPermissions,omitempty"` // allow-only
    Message            string           `json:"message,omitempty"`            // deny-only
    Interrupt          bool             `json:"interrupt,omitempty"`          // deny-only
}
```

Hook declaration schema (settings):

```yaml
# .claude/settings.json (hooks section)
hooks:
  PreToolUse:
    - matcher: "Write|Edit"
      hooks:
        - type: command
          command: "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh"
          if: "Write(*.go) | Edit(*.go)"   # permission-rule syntax
          async: false
          once: false
          timeout: 5
```

Source precedence merge (v3.0 scope — 3 tiers):

```
resolve(eventType) =
  merge(
    userHooks[eventType],      // ~/.moai/hooks.yaml
    projectHooks[eventType],   // .claude/settings.json
    localHooks[eventType],     // .claude/settings.local.json
  )
```

Dedup key: `{shell}\0{command}\0{if}` (W1.1 §3.9). Duplicate hooks across tiers retained only once; outermost tier wins.

#### Breaking change impact

- **BC-001** — Hook output protocol (exit-code-only deprecated)
- **BC-003** — Hook settings source layering (flat single-tier → 3-tier)
- **BC-005** — WorktreeCreate provider contract (observational → stdout-is-path)
- **BC-007** — PermissionRequest decision semantics
- **BC-008** — PermissionDenied retry semantics

#### Migration path

- **v3.0 dual-parse**: `internal/hook/protocol.go` attempts JSON first; on parse error or empty stdout, falls back to exit-code semantics. `HookOutput.ExitCode` synthesizes from `Decision` field for backward compatibility.
- **v3.0 warn on legacy**: if a hook returns only exit codes with no JSON, emit a `systemMessage` to the user: "Hook {wrapper-name} uses deprecated exit-code protocol; see docs-site/migration/v3-hooks.md."
- **v3.2 (next minor)**: the warn becomes an error for new hook declarations; existing hooks with `api_version: 2` (optional frontmatter) retain dual-parse.
- **v4.0**: exit-code fallback removed entirely.
- **Opt-out**: `MOAI_HOOK_LEGACY=1` env var disables warn for CI/airgap.

#### SPEC IDs

- SPEC-V3-HOOKS-001 — Hook Protocol v2 — Rich JSON IO
- SPEC-V3-HOOKS-002 — Hook Type System — 4 Types (command, prompt, agent, http) [absorbed the previously-planned type:prompt, type:agent, and type:http SPECs into a single type-system SPEC]
- SPEC-V3-HOOKS-003 — Async Hook Execution — async, asyncRewake, once
- SPEC-V3-HOOKS-004 — Hook Matcher & Filter System — if condition
- SPEC-V3-HOOKS-005 — Missing Hook Event Handlers — 14 Events
- SPEC-V3-HOOKS-006 — Hook Scoping Hierarchy — 3-tier

### 3.2 Theme 2 — Migration Framework

#### Problem statement

moai-adk has one one-shot migration (`moai migrate agency` per SPEC-AGENCY-ABSORB-001; `internal/cli/migrate_agency.go:569`). With 104 SPECs, 50 skills, 22 agents, and 22 YAML sections, every future breaking change now incurs an ad-hoc script. Wave 1.6 §15.4 also found `project.yaml.template_version` silently stale by 12 minor versions — a defect that a versioned framework would prevent.

Claude Code's `CURRENT_MIGRATION_VERSION` counter (W1.5 §5.1) with ordered, idempotent migration runner firing in `preAction` hook (W1.5 §5.2) is the reference pattern.

#### Design approach

1. Introduce `CURRENT_MIGRATION_VERSION` counter in `.moai/config/sections/system.yaml` (`migration_version: 1`). Counter increments by 1 per landed migration step.
2. Each migration is a `MigrationStep` implementing:
```go
type MigrationStep interface {
    Version() int           // monotonic, starts at 1
    ID() string             // e.g., "M01-template-version-sync"
    Description() string    // human-readable
    IsIdempotent() bool     // safe to re-run?
    PreConditionsMet(*MigrationContext) (bool, error)
    DryRun(*MigrationContext) (MigrationDiff, error)
    Apply(*MigrationContext) error
    Rollback(*MigrationContext) error
}
```
3. Migration runner (`internal/core/migration/runner.go`) fires on selected commands only (conservative, see §9 open question #10): `moai init`, `moai update`, `moai doctor`, `moai migrate`. Explicitly NOT on every cobra command (avoids surprise on `moai version`, `moai status`, etc.).
4. Opt-out: `MOAI_DISABLE_MIGRATIONS=1` environment variable for CI and air-gapped installs.
5. Every migration snapshots affected files to `.moai/backups/{ISO-8601-timestamp}/` before applying (rollback path).
6. Initial migration set (dogfood fixes for §2.3 debt):
   - **M01**: Auto-sync `project.yaml.template_version` with `system.yaml.moai.version` (fixes W1.6 §15.8, gm#183).
   - **M02**: Archive `.agency/` redirects and `.moai-backups/` to `~/.moai/history/{version}/` (addresses gm#190, gm#191).
   - **M03**: Hook wrapper drift resolver — deploy missing `handle-permission-denied.sh` (gm#185).
   - **M04**: Skill drift resolver — deploy 3 template-only skills to `.claude/skills/` (gm#184).
   - **M05**: Remove `.go.bak` files and stale coverage artifacts (gm#186, gm#187, gm#188).
7. Post-migration: log to `.moai/reports/migration-{timestamp}.md` with before/after diff summary.

#### API/schema sketch

```go
// internal/core/migration/context.go
type MigrationContext struct {
    ProjectRoot  string
    ConfigMgr    *config.Manager
    TemplateSrc  fs.FS             // EmbeddedTemplates()
    Logger       *slog.Logger
    DryRun       bool
    BackupDir    string            // .moai/backups/{timestamp}/
}

// internal/core/migration/runner.go
type Runner struct {
    steps   []MigrationStep   // registered in init()
    getVer  func() int        // reads system.yaml.migration_version
    setVer  func(int) error
}

func (r *Runner) Run(ctx *MigrationContext) (*RunReport, error) {
    current := r.getVer()
    pending := filter(r.steps, func(s MigrationStep) bool { return s.Version() > current })
    sort.Slice(pending, func(i, j int) bool { return pending[i].Version() < pending[j].Version() })

    var applied []AppliedStep
    for _, step := range pending {
        ok, err := step.PreConditionsMet(ctx)
        if !ok || err != nil { continue }

        diff, err := step.DryRun(ctx)
        if err != nil { return &RunReport{Applied: applied}, err }

        if ctx.DryRun {
            applied = append(applied, AppliedStep{Step: step, Diff: diff, DryRun: true})
            continue
        }

        if err := step.Apply(ctx); err != nil {
            _ = step.Rollback(ctx)
            return &RunReport{Applied: applied}, err
        }
        if err := r.setVer(step.Version()); err != nil { return nil, err }
        applied = append(applied, AppliedStep{Step: step, Diff: diff})
    }
    return &RunReport{Applied: applied}, nil
}
```

#### Breaking change impact

- **BC-004** — Migrations auto-run on init/update/doctor/migrate (was: explicit `moai migrate <name>` only)

#### Migration path

- v3.0 first run: user sees "5 migrations pending. Run with --yes to apply, --dry-run to preview. Or set MOAI_DISABLE_MIGRATIONS=1 to skip."
- v3.0 `moai init` on fresh project: no migrations run (`migration_version` starts at current count).
- v3.0 `moai update` on v2.12 project: migrations M01–M05 run in order after template refresh.

#### SPEC IDs

- SPEC-V3-MIG-001 (Versioned migration framework)
- SPEC-V3-MIG-002 (Initial migration set M01-M05)

### 3.3 Theme 3 — Schema & Validation Layer

#### Problem statement

moai has zero formal schemas (W1.6 §10.1). All 22 `.moai/config/sections/*.yaml` files are parsed via `gopkg.in/yaml.v3` with struct tags only; typos like `development_mode: tddd` fail silently and default to zero-value. Claude Code uses Zod v4 with `lazySchema()` and full round-trip safety (W1.5 §6.1, §6.7). Every Tier 1 and Tier 2 theme depends on schemas existing.

#### Design approach

1. Adopt `github.com/go-playground/validator/v10` (Go-native, minimal dep) for struct-tag-based validation. Rejected alternative: CUE (external language, cross-file dependency mismatch with moai's Go-first philosophy; see §9 open question #6).
2. Introduce `internal/config/schema/` with one file per YAML section (`system_schema.go`, `workflow_schema.go`, etc.). Each defines:
   - Go struct with `validate:"..."` tags
   - `Validate() error` method
   - `JSONSchema() []byte` method (auto-gen via `invopop/jsonschema` or hand-rolled)
3. Schema registration in `internal/config/schema/registry.go` — maps section name → validator.
4. `moai doctor config` uses the registry: loads each YAML, validates, reports violations with file:line and a fix suggestion when deterministic.
5. Settings source layering (v3.0 scope — 3 tiers; 6-tier deferred to v3.2 per §9 open question #4):
   - `userSettings` — `~/.moai/config/sections/*.yaml`
   - `projectSettings` — `.moai/config/sections/*.yaml`
   - `localSettings` — `.moai/config/sections/*.local.yaml` (gitignored)
6. Deep-merge semantics: maps merged key-by-key; arrays replaced wholesale; scalars overridden. Precedence (last-wins): local > project > user.
7. JSON Schema export: `make schemas` generates `docs-site/static/schemas/*.json` for editor integration (`$schema` field in YAML heads).

#### API/schema sketch

```go
// internal/config/schema/quality_schema.go
package schema

type QualityConfig struct {
    DevelopmentMode    string  `yaml:"development_mode" validate:"required,oneof=tdd ddd"`
    TestCoverageTarget int     `yaml:"test_coverage_target" validate:"min=0,max=100"`
    LSPQualityGates    struct {
        Plan LSPGateLevel `yaml:"plan" validate:"required,dive"`
        Run  LSPGateLevel `yaml:"run" validate:"required,dive"`
        Sync LSPGateLevel `yaml:"sync" validate:"required,dive"`
    } `yaml:"lsp_quality_gates" validate:"required"`
    Principles struct {
        Simplicity struct {
            MaxParallelTasks int `yaml:"max_parallel_tasks" validate:"min=1,max=50"`
        } `yaml:"simplicity"`
    } `yaml:"principles"`
}

type LSPGateLevel struct {
    MaxErrors    int `yaml:"max_errors" validate:"gte=0"`
    MaxWarnings  int `yaml:"max_warnings" validate:"gte=0"`
}

func (c QualityConfig) Validate() error {
    return validator.New().Struct(c)
}
```

Doctor command behavior:

```
$ moai doctor config
Checking .moai/config/sections/quality.yaml...
  ERROR development_mode: "tddd" is not one of [tdd ddd] (line 3)
    suggestion: did you mean "tdd"?
  ERROR test_coverage_target: 150 must be <= 100 (line 7)
    suggestion: must be between 0 and 100

Run 'moai doctor config --fix' to auto-repair recoverable issues.
2 errors in 1 file.
```

#### Breaking change impact

- **BC-002** — Settings source layering (flat → 3-tier)
- **BC-006** — Strict config validation (typos previously silent now fail)

#### Migration path

- v3.0 first `moai update` or `moai init`: runs `doctor config --fix` in auto-repair mode. Unrecoverable errors require user intervention.
- v3.0 dual-parse: if any YAML fails schema validation, fall back to v2 lenient parsing BUT emit `systemMessage` warning: "Config section {name} failed v3 schema; using v2 best-effort parse. See docs-site/migration/v3-schemas.md."
- v3.2: strict mode default on; lenient fallback opt-in via `MOAI_CONFIG_STRICT=0`.

#### SPEC IDs

- SPEC-V3-SCH-001 (Formal config schemas + validator/v10 + JSON Schema export)
- SPEC-V3-SCH-002 (Settings source layering, 3-tier)

### 3.4 Theme 4 — Agent Frontmatter Expansion

#### Problem statement

moai agents support `name`, `description`, `tools`, `model`, `permissionMode`, `memory`, `skills`, `hooks` (W1.6 §6.3). Claude Code's agent model supports additionally: `memory: user|project|local`, `initialPrompt`, `requiredMcpServers`, `omitClaudeMd`, `maxTurns`, `criticalSystemReminder_EXPERIMENTAL`, `background`, `isolation` (W1.3 §4.2, §8.1). Notably, `omitClaudeMd` alone saves 5–15 Gtok/week per CC BigQuery telemetry (W1.3 §4.2), extremely relevant for moai's high-volume Explore-style research agents.

#### Design approach

1. Extend agent frontmatter parser (`internal/template/deployer.go` + agent-loading path) to accept the new fields.
2. Validate per `internal/config/schema/agent_schema.go`:
```go
type AgentFrontmatter struct {
    Name          string `yaml:"name" validate:"required,slug"`
    Description   string `yaml:"description" validate:"required"`
    Tools         string `yaml:"tools"` // CSV string; parsed to []string
    Model         string `yaml:"model" validate:"oneof=opus sonnet haiku inherit"`
    PermissionMode string `yaml:"permissionMode" validate:"oneof=default plan acceptEdits bypassPermissions"`

    // v3 additions
    Memory             string   `yaml:"memory" validate:"omitempty,oneof=user project local"`
    InitialPrompt      string   `yaml:"initialPrompt"`
    RequiredMcpServers []string `yaml:"requiredMcpServers"`
    OmitClaudeMd       bool     `yaml:"omitClaudeMd"`
    MaxTurns           int      `yaml:"maxTurns" validate:"min=0"`
    CriticalSystemReminderExperimental string `yaml:"criticalSystemReminder_EXPERIMENTAL"`
    Background         bool     `yaml:"background"`
    Isolation          string   `yaml:"isolation" validate:"omitempty,oneof=none worktree"`
    Skills             []string `yaml:"skills"`                            // already partial
    Hooks              map[string][]AgentHook `yaml:"hooks"`               // already partial
    Effort             string   `yaml:"effort" validate:"omitempty,oneof=low medium high xhigh max"`
}
```
3. Default `omitClaudeMd: false` (§9 open question #9 — opt-in, never default-on). Candidates for opt-in in v3.0: `researcher`, `plan-auditor`, `evaluator-active`.
4. `requiredMcpServers` availability check fires at agent spawn time; 30-second timeout matching CC (W1.3 §4.2). Fails-fast with blocker report if server unavailable.
5. `background` as frontmatter field complements spawn-time `Agent(background: true)`. Frontmatter value is a DEFAULT; spawn-time overrides.
6. `isolation` as frontmatter field formalizes current convention (`.claude/rules/moai/workflow/worktree-integration.md`). Spawn-time override remains authoritative.
7. Fork subagent primitive (Tier 2, simplified scope for v3.0): `Agent()` call without `subagent_type` inherits parent's system prompt (NOT cache-identical prefix — that's v3.1). Recursion depth cap: 2 (parent → fork → leaf). Enforced by `internal/cli/deps.go` agent spawn dispatcher.

#### API/schema sketch

```yaml
# .claude/agents/moai/researcher.md (v3 frontmatter)
---
name: researcher
description: Read-only research agent. Use for codebase discovery.
tools: Read, Grep, Glob, WebSearch, WebFetch
model: opus
permissionMode: plan

# v3 additions
memory: project                 # project-scoped persistent memory
omitClaudeMd: true              # skip CLAUDE.md to save ~15Ktok/spawn
maxTurns: 30                    # hard cap
requiredMcpServers: [context7]  # block if context7 unavailable
effort: high                    # Opus 4.7 effort level
skills:
  - moai-foundation-cc
  - moai-foundation-core
---
```

#### Breaking change impact

None — all additions are optional. Existing 22 agents work unchanged under v3 schema.

#### Migration path

- v3.0 `moai update` re-parses existing agent definitions against v3 schema. Validation errors reported but non-blocking (legacy allowed).
- Per-agent opt-in: edit frontmatter to add new fields as needed.

#### SPEC IDs

- SPEC-V3-AGT-001 (Agent frontmatter v2 bundle: memory / initialPrompt / requiredMcpServers / omitClaudeMd / maxTurns / critical / background / isolation / effort)
- SPEC-V3-AGT-002 (Built-in moai agents: Explore / Plan moai-augmented versions)
- SPEC-V3-AGT-003 (Fork subagent primitive, simplified scope)

### 3.5 Theme 5 — Memory 2.0 Alignment

#### Problem statement

CC's `memdir` contract (W1.2 §6) mandates 4-type taxonomy (user/feedback/project/reference), 200-line / 25KB entrypoint truncation, freshness markers via `<system-reminder>` for memories older than 1 day, path security validation (non-absolute reject, tilde-only reject, UNC reject, null byte reject, NFKC attack reject — W1.2 §6.3). moai already uses the 4-type TAXONOMY implicitly (via `~/.claude/projects/-*/memory/MEMORY.md` format) but does NOT enforce truncation, freshness, or path security. Current MEMORY.md is unbounded; CLAUDE.md has 40K cap (`.claude/rules/moai/development/coding-standards.md`) but MEMORY.md does not.

#### Design approach

1. `internal/core/memory/truncate.go`: enforce `MAX_ENTRYPOINT_LINES=200` and `MAX_ENTRYPOINT_BYTES=25000` with human-readable warning appendix: `> [Truncated at {N} lines / {M} bytes — see .moai/reports/memory-truncation-{timestamp}.md for full content]`.
2. `internal/core/memory/freshness.go`: compute memory age from `mtime`. Files >1 day old get a wrapping `<system-reminder>Memory file {path} last modified {relative-age}; content may be stale.</system-reminder>`.
3. `internal/core/memory/validate.go`: port CC's `validateMemoryPath` rules faithfully (W1.2 §6.3).
4. `internal/core/memory/taxonomy.go`: enforce 4-type classification. Each MEMORY.md has YAML frontmatter:
```yaml
---
memory_type: project  # user | feedback | project | reference
last_validated: 2026-04-22T12:00:00Z
---
```
5. LLM-based relevance selection (T2-MEM-02, opt-in): `memory.yaml.llm_relevance.enabled: false` default. When enabled, runs a Haiku-class side query with memory candidates and picks top-5 relevant (configurable via `top_k`). Cost: ~$0.01-0.03 per turn; telemetry-driven evaluation for default-on in v3.2.

#### API/schema sketch

```go
// internal/core/memory/types.go
type MemoryEntry struct {
    Path          string    `yaml:"-"`
    Type          string    `yaml:"memory_type" validate:"oneof=user feedback project reference"`
    LastValidated time.Time `yaml:"last_validated"`
    Content       string    `yaml:"-"`
    SizeBytes     int       `yaml:"-"`
    LineCount     int       `yaml:"-"`
    AgeHours      float64   `yaml:"-"`
}

func Truncate(content string) (truncated string, truncationMark string) {
    // MAX_LINES=200, MAX_BYTES=25000; return truncation marker on overflow
}

func FreshnessPreamble(entry MemoryEntry) string {
    if entry.AgeHours <= 24 {
        return ""
    }
    return fmt.Sprintf("<system-reminder>Memory file %s last modified %s; content may be stale.</system-reminder>\n",
        entry.Path, humanize.Time(time.Now().Add(-time.Duration(entry.AgeHours)*time.Hour)))
}
```

Config:

```yaml
# .moai/config/sections/memory.yaml
memory:
  truncation:
    max_lines: 200
    max_bytes: 25000
  freshness:
    stale_threshold_hours: 24
  llm_relevance:
    enabled: false         # default off (opt-in for v3.0)
    model: haiku           # Haiku for cost
    top_k: 5
    timeout_seconds: 5
```

#### Breaking change impact

None — all additions are non-breaking. Users with oversized MEMORY.md files see a one-time truncation notice.

#### Migration path

- Existing MEMORY.md files >200 lines or >25KB: on first load, full content snapshotted to `.moai/reports/memory-truncation-{timestamp}.md`, in-context view truncated. User sees one-time `systemMessage`.

#### SPEC IDs

- SPEC-V3-MEM-001 (MEMORY.md truncation + freshness + 4-type enforcement + path validation)
- SPEC-V3-MEM-002 (LLM-based memory relevance, opt-in)

### 3.6 Theme 6 — Plugin Ecosystem Parity

#### Problem statement

moai has no plugin system today (gm#140, W1.5 §9.4). CC's plugin system (W1.5 §4) supports 3 plugin kinds (built-in, marketplace, session inline), 5 marketplace source types (github, git, url, directory, file), 6 capabilities (agents, skills, commands, hooks, mcpServers, outputStyles), install scopes (user/project/local), and hot-reload. Full parity is massive (Tier 4 defer); v3.0 ships a reduced-scope v1.

#### Design approach (v3.0 scope — reduced)

1. Plugin manifest at `.moai-plugin/plugin.json`:
```json
{
  "name": "example-moai-plugin",
  "version": "1.0.0",
  "description": "Example plugin",
  "author": "you@example.com",
  "capabilities": {
    "agents": ["agents/my-agent.md"],
    "skills": ["skills/my-skill/"],
    "commands": ["commands/my-command.md"]
  },
  "engines": {
    "moai": "^3.0.0"
  }
}
```
2. **v1 scope**: agents + skills + commands ONLY. NO hooks, NO mcpServers, NO outputStyles (defer to v3.2+).
3. Marketplace source types v1: `github` + `directory` only. Defer git/url/file.
4. Install scopes: `user` (`~/.moai/plugins/`), `project` (`.moai/plugins/`), `local` (`.moai/plugins/local/`, gitignored).
5. CLI surface:
   - `moai plugin install {source}` (source = `github:owner/repo@tag` or local path)
   - `moai plugin uninstall {name}`
   - `moai plugin enable {name}` / `moai plugin disable {name}`
   - `moai plugin update {name}`
   - `moai plugin list`
   - `moai plugin marketplace add {url}` / `list` / `remove` / `update`
   - `moai plugin validate {path}`
6. Validation (`moai plugin validate`): manifest schema check + content walk (agents valid per schema, skills have SKILL.md, commands are thin routers).
7. Dependency conflicts: two plugins declaring the same agent `name` → error at install time (no silent override).
8. Trust: plugins from `github:*` require explicit `--trust` flag on first install; `directory:` plugins require source-pinned commit hash.

#### API/schema sketch

```go
// internal/plugin/manifest.go
type PluginManifest struct {
    Name         string              `json:"name" validate:"required,slug"`
    Version      string              `json:"version" validate:"required,semver"`
    Description  string              `json:"description"`
    Author       string              `json:"author"`
    Capabilities PluginCapabilities  `json:"capabilities" validate:"required"`
    Engines      PluginEngines       `json:"engines" validate:"required"`
}

type PluginCapabilities struct {
    Agents   []string `json:"agents,omitempty"`   // paths to .md files
    Skills   []string `json:"skills,omitempty"`   // paths to skill dirs
    Commands []string `json:"commands,omitempty"` // paths to .md files
}

type PluginEngines struct {
    Moai string `json:"moai" validate:"required,semverConstraint"` // e.g., "^3.0.0"
}

// internal/plugin/installer.go
type Installer interface {
    Install(ctx context.Context, source string, scope InstallScope, opts InstallOpts) (*InstallReport, error)
    Uninstall(ctx context.Context, name string, scope InstallScope) error
    List(scope InstallScope) ([]InstalledPlugin, error)
    Validate(path string) (*ValidationReport, error)
}
```

#### Breaking change impact

None — plugin system is entirely additive.

#### Migration path

N/A — no prior state to migrate.

#### SPEC IDs

- SPEC-V3-PLG-001 (Plugin system v1: manifest + install + marketplace skills/agents/commands)

### 3.7 Theme 7 — Output Style Contract

#### Problem statement

moai emits plain-text diffs, errors, and progress (W1.4 §8.4). Claude Code's Ink renderer has 146 components ready to render structured content (StructuredDiff, ValidationErrorsList, StatusIcon, ProgressBar, HighlightedCode) — but only if moai emits the right output formats. Zero TUI code required on moai's side — pure output-fidelity theme.

#### Design approach

1. `internal/output/diff.go`: emit `diff --git` format for all file modifications. Replaces plain-text `<<<<<<< / =======` style.
2. `internal/output/errors.go`: structured errors emitted as YAML lists:
```yaml
validation_errors:
  - severity: error
    path: .moai/config/sections/quality.yaml
    line: 3
    message: 'development_mode: "tddd" is not one of [tdd ddd]'
    suggestion: 'did you mean "tdd"?'
```
3. `internal/output/progress.go`: progress prefixes during long operations. `Progress: 3/10 — running gopls diagnostics`. Integrates with `moai doctor`, `moai update`, `moai migrate`.
4. `internal/output/code.go`: every code block gets a language fence (` ```go `, ` ```python `, ` ```yaml `).
5. `internal/output/file_link.go`: absolute paths emitted with OSC-8 hyperlinks when stdout is TTY (detected via `isatty` — already a dep, W1.6 §14.1). Graceful fallback when non-TTY (plain absolute path).
6. `internal/output/status_icon.go`: `✓` (success), `✗` (error), `⚠` (warning), `ℹ` (info), `○` (pending), `…` (loading). ANSI-colored when TTY.
7. Output style files under `.claude/output-styles/` use CC-compat frontmatter keys (`name`, `description`, `keep-coding-instructions`, `force-for-plugin`). moai-specific metadata goes under `moai:` prefix:
```yaml
---
name: moai-default
description: Default MoAI output style
keep-coding-instructions: true
moai:
  spec_id: SPEC-V3-OUT-001
---
```

#### API/schema sketch

```go
// internal/output/renderer.go
type Renderer struct {
    isTTY bool
    color bool
}

func (r *Renderer) Diff(oldPath, oldContent, newPath, newContent string) string {
    // emit "diff --git a/{path} b/{path}" with unified context
}

func (r *Renderer) ValidationError(errs []ValidationError) string {
    // YAML list with severity/path/line/message/suggestion
}

func (r *Renderer) Progress(step, total int, message string) string {
    return fmt.Sprintf("Progress: %d/%d — %s", step, total, message)
}

func (r *Renderer) FilePath(path string, line int) string {
    if r.isTTY {
        return fmt.Sprintf("\x1b]8;;file://%s#L%d\x1b\\%s:%d\x1b]8;;\x1b\\", path, line, path, line)
    }
    return fmt.Sprintf("%s:%d", path, line)
}
```

#### Breaking change impact

None — output-only changes; downstream CC rendering unchanged.

#### Migration path

N/A — additive.

#### SPEC IDs

- SPEC-V3-OUT-001 (Output contract v2: diff + errors + progress + language hints + OSC-8)

### 3.8 Theme 8 — Team Protocol v2 (Mailbox + In-process)

#### Problem statement

moai team mode today uses ad-hoc JSON for `SendMessage` payloads (W1.6 feedback memory, gm#71). CC defines 10 Zod-validated message types (W1.3 §6.8): `shutdown_request`, `shutdown_approved`, `shutdown_rejected`, `plan_approval_request`, `plan_approval_response`, `permission_request`, `permission_response`, `sandbox_permission_request`, `sandbox_permission_response`, `task_assignment`. This prevents a whole class of team-mode bugs (silent payload-shape drift).

In-process teammate backend (via goroutine + `context.Context` — gm#68) is deferred to v3.1 (§9 open question #7). v3.0 ships the mailbox schema only.

#### Design approach

1. `internal/team/mailbox/types.go`: 10 typed message structs, each implementing `Message` interface with `Type() string` method.
2. `SendMessage` adds strict mode: payloads validated against schema before enqueue. Invalid messages rejected with error.
3. Legacy ad-hoc JSON: accepted with warning log for one minor version (v3.0). Removed in v3.2.
4. Plan-approval flow (T3-TEAM-03): team lead receives `plan_approval_request` from teammate, sends `plan_approval_response` with `feedback` payload. Implementation in `internal/team/approval/`.
5. `TEAMMATE_MESSAGES_UI_CAP = 50`: bounded message history per teammate session (prevents 36GB whale sessions per W1.3 §7.2).

#### API/schema sketch

```go
// internal/team/mailbox/types.go
type Message interface {
    Type() string
    RequestID() string  // for request/response pairing
}

type ShutdownRequest struct {
    ReqID     string `json:"request_id" validate:"required,uuid4"`
    TeamName  string `json:"team_name" validate:"required"`
    Initiator string `json:"initiator" validate:"required"`
    Reason    string `json:"reason"`
}
func (ShutdownRequest) Type() string        { return "shutdown_request" }
func (m ShutdownRequest) RequestID() string { return m.ReqID }

type PlanApprovalRequest struct {
    ReqID        string `json:"request_id" validate:"required,uuid4"`
    TeammateName string `json:"teammate_name" validate:"required"`
    PlanPath     string `json:"plan_path" validate:"required"`
    Summary      string `json:"summary" validate:"required"`
}
func (PlanApprovalRequest) Type() string { return "plan_approval_request" }

type PlanApprovalResponse struct {
    ReqID    string `json:"request_id" validate:"required,uuid4"`
    Approved bool   `json:"approved"`
    Feedback string `json:"feedback"`
}
func (PlanApprovalResponse) Type() string { return "plan_approval_response" }

// ... (7 more types)

// internal/team/mailbox/registry.go
var registry = map[string]func() Message{
    "shutdown_request":      func() Message { return &ShutdownRequest{} },
    "plan_approval_request": func() Message { return &PlanApprovalRequest{} },
    // ... 10 types total
}

func Decode(raw []byte) (Message, error) {
    var envelope struct{ Type string `json:"type"` }
    if err := json.Unmarshal(raw, &envelope); err != nil { return nil, err }
    factory, ok := registry[envelope.Type]
    if !ok {
        // legacy ad-hoc payload — return LegacyMessage{RawBytes: raw}
    }
    msg := factory()
    if err := json.Unmarshal(raw, msg); err != nil { return nil, err }
    if err := validator.New().Struct(msg); err != nil { return nil, err }
    return msg, nil
}
```

#### Breaking change impact

- **BC-008** — ad-hoc JSON accepted for 1 minor version with warning, removed v3.2.

#### Migration path

- v3.0 legacy messages: `LegacyMessage{RawBytes: ...}` wrapper. Handlers receive raw bytes; must self-deserialize. Warning logged.
- v3.2: unknown `type` rejected outright.

#### SPEC IDs

- SPEC-V3-TEAM-001 (Teammate mailbox v2 schemas, 10 message types, plan-approval flow)

### 3.9 Theme 9 — Internal Cleanup & Template Drift Resolution

#### Problem statement

Eleven self-identified issues from Wave 1.6 §15 (listed in §2.3 above). None of these add capability; they restore integrity. All are low-risk XS-effort fixes that ship as part of the M01–M05 migration set (§3.2 Theme 2).

#### Design approach

1. **T1-CLN-01** (Template drift resolution): migration M03 (hook wrapper) + M04 (skill) deploy the 4 missing files. Byte-identical to template source.
2. **T1-CLN-02** (`template_version` sync): migration M01 runs one-line YAML patch reading `system.yaml.moai.version` and writing `project.yaml.template_version`. Idempotent.
3. **T1-CLN-03** (Legacy code removal): migration M05 removes:
   - `internal/cli/glm.go.bak` (28,567 bytes)
   - `internal/cli/worktree/new_test.go.bak` (13,700 bytes)
   - `coverage.out` / `coverage.html` (if older than 30 days)
4. **Fix ADR-011 comment drift** (gm#189): not a migration — direct source edit in `internal/template/embed.go:8-12`. Ships in v3.0 first PR.
5. **.agency/ archival** (gm#190, gm#191): migration M02 moves `.claude/commands/agency/*.md` (8 redirect files), `.claude/rules/agency/constitution.md` (stub), `.moai-backups/` folder to `~/.moai/history/v2.12/`. Never deletes.
6. **docs-site 4-locale sync** (T3-DOC-01, gm#194): delegate to `manager-docs` subagent; uses existing `docs-site/scripts/translate.mjs`. Part of Phase 7 user-facing docs, NOT a migration.
7. **`.mcp.json` pencil addition** (gm#192): direct source edit in `internal/template/templates/.mcp.json.tmpl` to include `pencil` config block with commentary note about user-scope fallback.
8. **Handler count doc** (gm#193): add ADR-style comment in `internal/cli/deps.go:151-186` explaining that `AutoUpdateHandler` is a second SessionStart handler via compose pattern.

#### API/schema sketch

```go
// internal/core/migration/steps/m01_template_version.go
type M01TemplateVersion struct{}

func (m M01TemplateVersion) Version() int     { return 1 }
func (m M01TemplateVersion) ID() string       { return "M01-template-version-sync" }
func (m M01TemplateVersion) Description() string {
    return "Sync .moai/config/sections/project.yaml:template_version with system.yaml:moai.version"
}
func (m M01TemplateVersion) IsIdempotent() bool { return true }

func (m M01TemplateVersion) PreConditionsMet(ctx *MigrationContext) (bool, error) {
    projectCfg, err := loadYAML(ctx.ProjectRoot + "/.moai/config/sections/project.yaml")
    if err != nil { return false, err }
    systemCfg, err := loadYAML(ctx.ProjectRoot + "/.moai/config/sections/system.yaml")
    if err != nil { return false, err }
    return projectCfg["template_version"] != systemCfg["moai"].(map[string]any)["version"], nil
}

func (m M01TemplateVersion) DryRun(ctx *MigrationContext) (MigrationDiff, error) { /* ... */ }
func (m M01TemplateVersion) Apply(ctx *MigrationContext) error                    { /* in-place YAML edit */ }
func (m M01TemplateVersion) Rollback(ctx *MigrationContext) error                 { /* restore from backup */ }
```

#### Breaking change impact

- Minor: users who manually pinned `template_version` may see it change. Mitigated by:
  - Backup to `.moai/backups/{timestamp}/project.yaml.bak`
  - One-time notice with rollback instructions

#### Migration path

- All cleanup lives in M01–M05. Users run `moai migrate v2-to-v3` or `moai update` on v3.0 first encounter.

#### SPEC IDs

- SPEC-V3-CLN-001 (Template drift resolution — M03/M04)
- SPEC-V3-CLN-002 (Legacy code removal — M05 + ADR-011 comment fix)

Note: M02 agency archival + docs-site 4-locale sync absorbed into SPEC-V3-MIG-002. The third CLN slot (previously planned as a separate SPEC covering `.agency/` archival and docs-site locale sync) was rolled into MIG-002 to keep migration step Go implementations in a single SPEC (see §3.9 Ownership Split below).

#### Ownership Split

**M01–M05 migration step ownership (resolves MIG-002 ↔ CLN-001/002 overlap)**:

- **SPEC-V3-MIG-002** owns ALL migration step Go implementations in `internal/core/migration/steps/m0*.go` (including M02 agency archival + docs-site 4-locale sync absorbed from the retired third CLN slot).
- **SPEC-V3-CLN-001** owns surrounding tooling: `moai doctor template-drift`, `moai doctor skill-drift` CLI commands, diagnostic reporting, and user-facing cleanup orchestration. CLN-001 MUST NOT implement migration step Go files directly — those live in MIG-002.
- **SPEC-V3-CLN-002** owns `moai doctor legacy-cleanup` CLI, direct source edits in files OUTSIDE `internal/core/migration/steps/`, and removal of .go.bak / stale ADR-011 comments. CLN-002 MUST NOT implement migration step Go files.

This split ensures mutually exclusive file scopes. Wave 5/6 schedulers assign migration step files to MIG-002; diagnostic tooling files to CLN-001/002.

### 3.10 Theme 10: Command Extension Parity (CMDS-001..003)

**Scope**: Extend moai command frontmatter to match Claude Code conventions: new fields (model, argument-hint array, context:fork, paths glob, skills array), named args + indexed $ARGUMENTS[N], inline !`cmd` bash execution.

**Breaking**: None (additive).

**SPECs**: SPEC-V3-CMDS-001 (frontmatter extensions), SPEC-V3-CMDS-002 (args system), SPEC-V3-CMDS-003 (inline bash).

**Rationale**: Claude Code ships 68+ built-in commands with rich frontmatter; moai's command system currently lacks named args and inline execution. CMDS theme closes this parity gap.

---

## 4. Breaking Changes Catalog

| BC-ID | Description | v2 Behavior | v3 Behavior | Migration |
|-------|-------------|-------------|-------------|-----------|
| **BC-001** | Hook output protocol | Exit codes (0/2/other); minimal JSON | Rich JSON: `decision`, `hookSpecificOutput`, `additionalContext`, `updatedInput`, `watchPaths`, `stopReason`, `systemMessage`, `continue` | Auto: v3.0 dual-parse shim; legacy warn; removed in v4.0. Env escape: `MOAI_HOOK_LEGACY=1` |
| **BC-002** | Settings source layering | Single tier (`.moai/config/sections/*.yaml`) | 3-tier: user / project / local with deep-merge precedence | Auto: migration M01 preserves existing project-tier values. Opt-in: users create `~/.moai/config/` and `*.local.yaml` as needed |
| **BC-003** | Hook settings source layering | Hook declarations only in project `.claude/settings.json` | Same 3-tier as BC-002, with per-event dedup key `{shell}\0{command}\0{if}` | Auto: existing hooks become project-tier. No action required |
| **BC-004** | Migration auto-run | Only explicit `moai migrate <name>` | Auto-run on `moai init`, `moai update`, `moai doctor`, `moai migrate` | Auto: dry-run preview + user confirmation by default. Escape: `MOAI_DISABLE_MIGRATIONS=1` |
| **BC-005** | WorktreeCreate provider contract | Observational only (log and return) | Provider: stdout MUST contain absolute path; non-zero exit = failed | Manual: users with custom WorktreeCreate hooks must write absolute path to stdout. Detected by `moai doctor hook --validate`; migration M06 optional |
| **BC-006** | Config schema strict validation | Typos silently ignored (YAML best-effort) | validator/v10 enforcement; unknown fields warn; out-of-range values error | Auto: v3.0 first run invokes `moai doctor config --fix` auto-repair. Unrecoverable errors logged with fix suggestions. Env escape: `MOAI_CONFIG_STRICT=0` |
| **BC-007** | PermissionRequest decision semantics | Pass-through (handler observational) | Handler can return `decision: {behavior: allow\|deny, updatedInput?, updatedPermissions?, message?, interrupt?}` | Auto: default handler returns neutral (behaves as pass-through). Opt-in via custom wrapper |
| **BC-008** | Team mailbox structured schema + PermissionDenied retry hint | Observational handler; ad-hoc JSON in team mode | `{retry: boolean}` reply; 10 typed team messages (shutdown, plan_approval, permission, sandbox_permission, task_assignment) | Auto: v3.0 accepts legacy ad-hoc team JSON with warning; strict in v3.2 |

### Blast radius analysis

- BC-001: affects hook authors. Current moai-adk-go hook surface = 26 shell wrappers + `internal/hook/registry.go` handlers. All internal wrappers will be rewritten in Phase 2. External hook authors (plugin users, custom scripts) have grace period.
- BC-002 / BC-003: affects config maintainers. Migration M01 touches `project.yaml:template_version` — one field. No user action required.
- BC-004: affects `moai init` / `moai update` / `moai doctor` / `moai migrate` command users. Dry-run default; explicit confirmation for apply.
- BC-005: affects users with custom `handle-worktree-create.sh`. Detection via `moai doctor hook --validate`. Current moai-adk-go handler is observational (W1.6 §5.7) — needs rewrite in Phase 2.
- BC-006: affects users whose configs have typos. `moai doctor config --fix` auto-repairs.
- BC-007 / BC-008: affects permission rule authors (rare) and team mode users (experimental).

---

## 5. Migration Strategy (v2 → v3)

### 5.1 `moai migrate v2-to-v3` Tool Design

Single entry point for users on v2.x upgrading to v3.0.

```
$ moai migrate v2-to-v3 --help
Usage: moai migrate v2-to-v3 [flags]

Perform the full v2.x → v3.0 migration. Equivalent to running:
  moai migrate (runs M01..M05)
  moai update  (refreshes templates)
  moai doctor config --fix (auto-repair schema errors)

Flags:
  --dry-run          Show what would change without applying (default: true for interactive)
  --yes              Skip confirmation prompts
  --no-backup        Skip backup snapshot (NOT recommended)
  --only <step>      Run a single migration step (e.g., "M01")
  --rollback <ts>    Roll back to a timestamped backup
```

Behavior:

1. Pre-flight check: moai-adk-go v3.0.0+ binary is current (else error "upgrade moai binary first").
2. Snapshot `.moai/backups/{ISO-8601-timestamp}/` containing changed files.
3. Run migration steps M01–M05 in order.
4. Refresh templates via `moai update` logic.
5. Run `moai doctor config --fix` in auto-repair mode.
6. Generate summary report `.moai/reports/migration-v2-to-v3-{timestamp}.md`.
7. Prompt user: "Migration complete. Review changes? (y/N)"

Dry-run mode outputs:

```
Planned changes:

M01-template-version-sync (idempotent)
  .moai/config/sections/project.yaml
    - template_version: v2.7.22
    + template_version: v2.12.0

M02-agency-archival (one-shot)
  .claude/commands/agency/ (8 files) → ~/.moai/history/v2.12/commands/agency/
  .claude/rules/agency/constitution.md → ~/.moai/history/v2.12/rules/agency/
  .moai-backups/ (folder) → ~/.moai/history/v2.12/backups/

M03-hook-wrapper-drift
  + .claude/hooks/moai/handle-permission-denied.sh (new)

M04-skill-drift (3 skills added)
  + .claude/skills/moai-domain-db-docs/ (from template)
  + .claude/skills/moai-workflow-design-context/ (from template)
  + .claude/skills/moai-workflow-pencil-integration/ (from template)

M05-legacy-cleanup
  - internal/cli/glm.go.bak (28,567 bytes)
  - internal/cli/worktree/new_test.go.bak (13,700 bytes)
  - coverage.out (dated 2026-03-11, stale)
  - coverage.html (dated 2026-03-11, stale)

Backup: .moai/backups/2026-04-22T14:30:00Z/
Apply? (y/N)
```

### 5.2 Backward-Compat Shims (deprecation window)

| Shim | v3.0 | v3.2 | v4.0 |
|------|------|------|------|
| Hook exit-code fallback | dual-parse | warn | remove |
| Config schema lenient fallback | dual-parse + auto-repair | strict default; opt-out env | strict only |
| Team mailbox ad-hoc JSON | accept with warning | reject unknown `type` | N/A |
| `.agency/` stub redirects | retain for backward `moai` command routing | remove | N/A |

### 5.3 User-facing Migration Guide Sketch

Location: `docs-site/content/{ko,en,ja,zh}/migration/v3.md`

Sections:

1. **Why v3?** — 1-paragraph summary of capabilities gained.
2. **Who needs to migrate?** — anyone running moai-adk-go v2.x.
3. **Before you start** — backup checklist, recommended: commit pending changes.
4. **The one command** — `moai migrate v2-to-v3 --dry-run && moai migrate v2-to-v3 --yes`.
5. **What changes** — table of affected files per migration step.
6. **Hook authors** — BC-001/005/007/008 details with before/after examples.
7. **Config maintainers** — BC-002/003/006 details, schema docs link.
8. **Team mode users** — BC-008 structured mailbox details.
9. **Rollback** — `moai migrate v2-to-v3 --rollback {timestamp}`.
10. **Troubleshooting** — common error recovery paths.

Each locale maintained per CLAUDE.local.md §17 rules (canonical = ko, translation SLA 48h for en, 72h for zh/ja).

---

## 6. Release Plan

### 6.1 Phase Structure

Phases are dependency-ordered, not time-boxed. Each phase closes with a release artifact (tag).

- **Phase 1 — Foundation**
  Schema (T1-SCH-01, T1-SCH-02) + Migration framework (T1-MIG-01, T1-MIG-02 with M01–M05) + ADR-011 comment fix.
  Artifact: `v3.0.0-alpha.1`.

- **Phase 2 — Hook Protocol v2 core**
  Phase 2 lands 5 HOOKS SPECs: HOOKS-001 (JSON output), HOOKS-003 (async/asyncRewake/once + CLAUDE_ENV_FILE), HOOKS-004 (matcher & `if` condition), HOOKS-005 (event handler richness), HOOKS-006 (3-tier scoping). HOOKS-002 (Hook Type System — 4 types: command/prompt/agent/http) is deferred to Phase 6a because its prompt/agent/http dispatch requires plugin and team primitives.
  Artifact: `v3.0.0-alpha.2`.

- **Phase 3 — Agent Runtime v2**
  T1-AGT-01 (frontmatter bundle), T1-AGT-02 (skills preload formalization), T1-AGT-03 (background frontmatter), T1-SKL-01 (skill frontmatter bundle).
  Artifact: `v3.0.0-alpha.3`.

- **Phase 4 — Memory 2.0 (parallel with Phase 3)**
  T1-MEM-01 (truncation + freshness + 4-type enforcement + path validation).
  Artifact: folded into next alpha tag.

- **Phase 5 — Internal Cleanup**
  T1-CLN-01 (template drift), T1-CLN-02 (template_version), T1-CLN-03 (legacy code), T2-SKL-03 (skill drift detection).
  All implemented as migrations in Phase 1's M01–M05 (owned by SPEC-V3-MIG-002 for step Go files); Phase 5 is the documentation + validation of the cleanup via SPEC-V3-CLN-001/002 tooling plus SPEC-V3-SKL-002 skill drift detection (per its frontmatter `phase: "v3.0.0 — Phase 5 Internal Cleanup"`). The retired third CLN slot (M02 agency archival + docs-site 4-locale sync) is absorbed into SPEC-V3-MIG-002.
  Artifact: `v3.0.0-beta.1`.

- **Phase 6a — Tier 2 Strategic Differentiators**
  T2-HOOK-10/11/12/13/14 (prompt/agent/http hook types + CLAUDE_ENV_FILE + full 8-tier source precedence), T2-PLG-01/02 (plugin system + marketplace), T2-MEM-02 (LLM relevance), T2-AGT-04/05 (fork subagent + built-in agents), T2-TEAM-01 (team mailbox), T2-DIFF-01 (SPEC-to-SPEC chaining).
  Artifact: `v3.0.0-beta.2`.

- **Phase 6b — Tier 2 Polish**
  T2-CLI-01 (startup profiler), T2-OUT-01 (output contract). Note: T2-SKL-02 (context:fork) is now delivered as part of SPEC-V3-SKL-001 (Phase 3 skill frontmatter bundle), and T2-SKL-03 (skill drift detection) ships via SPEC-V3-SKL-002 in Phase 5.
  Artifact: `v3.0.0-rc.1`.

- **Phase 7 — Migration Tool + User Docs**
  `moai migrate v2-to-v3` tool wiring, 4-locale migration guide (docs-site), release notes, CHANGELOG.
  Artifact: `v3.0.0-rc.2`.

- **Phase 8 — Release Rollout**
  Final QA on a corpus of 10 v2.12 projects, binary tagging, GitHub Release, announcement to Discord.
  Artifact: `v3.0.0` (stable).

### 6.2 Phase → SPEC → PR mapping

| Phase | SPECs | Expected PR count |
|-------|-------|--------------------|
| 1 | SPEC-V3-SCH-001, SPEC-V3-SCH-002, SPEC-V3-MIG-001, SPEC-V3-MIG-002 | 4–6 PRs (schema split per section, migration per step) |
| 2 | SPEC-V3-HOOKS-001..006 | 6–8 PRs |
| 3 | SPEC-V3-AGT-001, SPEC-V3-SKL-001 | 2–3 PRs |
| 4 | SPEC-V3-MEM-001 | 1–2 PRs |
| 5 | SPEC-V3-CLN-001, SPEC-V3-CLN-002, SPEC-V3-SKL-002 | Folded into Phase 1 migration PRs (retired third CLN slot absorbed into SPEC-V3-MIG-002); SKL-002 (Skill Drift Detection) is cleanup tooling per SPEC frontmatter `phase: "v3.0.0 — Phase 5 Internal Cleanup"` |
| 6a | SPEC-V3-HOOKS-002 (absorbs type:prompt/agent/http), SPEC-V3-PLG-001, SPEC-V3-MEM-002, SPEC-V3-AGT-002/003, SPEC-V3-TEAM-001, SPEC-V3-SPEC-001 | 7–9 PRs |
| 6b | SPEC-V3-CLI-001, SPEC-V3-OUT-001 | 2–3 PRs |
| 7 | SPEC-V3-MIGRATE-001 | 2–3 PRs |
| 8 | none (release workflow) | 1 PR (CHANGELOG bump, tag) |

Target total: ~28–39 PRs across 28 SPECs.

### 6.3 Rollout Strategy (alpha → beta → stable)

- **Alpha**: tagged per phase close (`-alpha.1`..`-alpha.3`). Released to early adopters via Discord opt-in list.
- **Beta**: `v3.0.0-beta.1` (after cleanup) and `-beta.2` (after Tier 2 strategic). Public but flagged.
- **RC**: `-rc.1` and `-rc.2`. Feature-frozen; only bug fixes and docs.
- **Stable**: `v3.0.0`. Tagged and released via goreleaser (`.goreleaser.yml` existing per W1.6 §13.4).

Homebrew / install.sh / install.bat / install.ps1 updated at `v3.0.0` stable (W1.6 §13.4).

---

## 7. Risk Register

| Risk ID | Description | Probability | Impact | Mitigation |
|---------|-------------|-------------|--------|------------|
| **R-001** | BC-001 hook backward-compat breakage — users with exit-code-only wrappers see regressions | Medium | High | Dual-parse shim for 2 minor versions; `MOAI_HOOK_LEGACY=1` env escape; `moai doctor hook --validate` detects legacy hooks |
| **R-002** | Migration M01-M05 corrupts user config | Low | Critical | Automatic backup to `.moai/backups/{timestamp}/`; dry-run default; rollback via `moai migrate v2-to-v3 --rollback` |
| **R-003** | BC-006 schema validation breaks configs with typos | Medium | Medium | `moai doctor config --fix` auto-repair; lenient fallback with warning for 1 minor version; escape via `MOAI_CONFIG_STRICT=0` |
| **R-004** | Plugin system scope creep (adds hooks/MCP/outputStyles mid-phase) | Medium | High | Explicit scope charter locks v1 to 3 capabilities (agents/skills/commands); full parity deferred to v3.2+; PR reviews reject scope additions |
| **R-005** | `omitClaudeMd: true` default regression — agents lose rules | Very Low | High | Default is `false`; opt-in per agent; 22 existing agents retain default behavior |
| **R-006** | Fork subagent recursion — parent spawns child which spawns child, etc. | Low | Medium | Depth cap of 2 enforced in `internal/cli/deps.go` agent dispatcher; detected by `moai doctor agent --validate` |
| **R-007** | HTTP hook SSRF (T2-HOOK-12) — user writes hook targeting internal network | Low | Critical | Faithful port of CC's `utils/hooks/ssrfGuard.ts` (W1.1 §12); URL allowlist via `policy_settings.allowed_http_hook_urls`; env-var interpolation allowlist |
| **R-008** | BC-004 migration surprise — users don't expect migrations on `moai update` | Medium | Low | Conservative trigger list (init/update/doctor/migrate only, NOT every command); dry-run default; one-time notice on first upgrade |
| **R-009** | BC-008 team mailbox legacy JSON breakage | Low | Medium | Accept both shapes for 1 minor version with warning log; strict in v3.2 |
| **R-010** | docs-site 4-locale translation lag — en/zh/ja behind ko at release | Medium | Low | Use `docs-site/scripts/translate.mjs`; SLA per CLAUDE.local.md §17.3 (ko merge → en within 48h → zh/ja within 72h); block release if locale parity less than 95% |
| **R-011** | Binary size balloon — plugin + schema + migration adds >4 MB | Low | Medium | Embedded schemas as minified JSON; plugin manifest parsing via `encoding/json` (stdlib); size target enforced by CI check |
| **R-012** | Fork subagent cache-identical prefix not working in v3.0 | Accepted | Low | Explicitly deferred to v3.1; v3.0 ships inherit-system-prompt only; documented in §9 open question #5 |

---

## 8. SPEC Index (Wave 4 inputs)

The Wave 4 SPEC writer will produce 28 SPECs organized across 8 groupings below. Each line gives SPEC ID, one-sentence scope, and traceability to Tier 1/2 items and gap matrix row numbers.

### 8.1 Hooks/Commands SPECs (9 SPECs: 6 HOOKS + 3 CMDS)

Hooks (6 SPECs — matches on-disk reality; the previously-planned separate type:prompt, type:agent, and type:http SPECs were absorbed into HOOKS-002 Hook Type System as a single unified type-dispatch SPEC):

- **SPEC-V3-HOOKS-001** — Hook Protocol v2 — Rich JSON IO (T1-HOOK-01; gm#4, gm#5, gm#6, gm#7). Extend `internal/hook/types.go HookOutput` with `HookSpecificOutput` discriminated union; emit `additionalContext`, `updatedInput`, `watchPaths`; dual-parse backward-compat shim.
- **SPEC-V3-HOOKS-002** — Hook Type System — 4 Types (command, prompt, agent, http) (T2-HOOK-10, T2-HOOK-11, T2-HOOK-12; gm#1, gm#2, gm#3). Unified hook type dispatcher covering: command (existing shell wrapper), prompt (Haiku-class LLM gate returning `{ok, reason?}` JSON with cost budget), agent (multi-turn subagent verifier respecting `ALL_AGENT_DISALLOWED_TOOLS`, depth cap 2), http (SSRF-guarded webhook with URL allowlist + env-var interpolation allowlist + CRLF sanitization).
- **SPEC-V3-HOOKS-003** — Async Hook Execution — async, asyncRewake, once (T1-HOOK-03, T1-HOOK-04, T2-HOOK-13; gm#9, gm#10, gm#11, gm#20). `AsyncHookRegistry` in Go; `once: true` self-remove; CLAUDE_ENV_FILE mechanism for 4 events (SessionStart/Setup/CwdChanged/FileChanged).
- **SPEC-V3-HOOKS-004** — Hook Matcher & Filter System — if condition (T1-HOOK-02, T3-HOOK-09; gm#8, gm#13). Add permission-rule-syntax `if` evaluator (`Bash(git *)`, `Read(*.ts)`); upgrade matcher to support exact / pipe-separated / regex / `*`.
- **SPEC-V3-HOOKS-005** — Missing Hook Event Handlers — 14 Events (T1-HOOK-06, T1-HOOK-07; gm#24, gm#27, gm#28, gm#29, gm#30, gm#31, gm#32). Upgrade PreCompact / ConfigChange / InstructionsLoaded / Elicitation / ElicitationResult / WorktreeCreate and remaining event handlers to emit structured outputs; PermissionRequest emits `decision: {behavior, updatedInput?, updatedPermissions?, message?, interrupt?}`; PermissionDenied emits `{retry: boolean}`.
- **SPEC-V3-HOOKS-006** — Hook Scoping Hierarchy — 3-tier (T1-HOOK-05, T3-HOOK-08; gm#15 scope-reduced, gm#14). Introduce user / project / local layering; dedup key `{shell}\0{command}\0{if}`; source precedence pipeline in `internal/hook/registry.go`.

Commands (3 SPECs — CMDS theme, see §3.10):

- **SPEC-V3-CMDS-001** — Command frontmatter extensions (model, argument-hint array, context:fork, paths glob, skills array).
- **SPEC-V3-CMDS-002** — Command args system (named args + indexed $ARGUMENTS[N] substitution).
- **SPEC-V3-CMDS-003** — Command inline `!`cmd`` bash execution.

### 8.2 Query/Context/Memory SPECs (2 SPECs)

- **SPEC-V3-MEM-001** — MEMORY.md 4-type + truncation + freshness + path validation (T1-MEM-01, T3-MEM-03; gm#44, gm#45, gm#50, gm#53). Enforce taxonomy, 200-line/25KB cap, freshness `<system-reminder>` wrapping, path security rules.
- **SPEC-V3-MEM-002** — LLM-based memory relevance (opt-in) (T2-MEM-02; gm#48, gm#49, gm#52). Haiku side-query returns top-k relevant; config-gated; default off.

### 8.3 Agent/Team SPECs (4 SPECs)

- **SPEC-V3-AGT-001** — Agent frontmatter v2 bundle (T1-AGT-01, T1-AGT-02, T1-AGT-03; gm#56, gm#57, gm#58, gm#59, gm#60, gm#61, gm#62, gm#63). Add memory/initialPrompt/requiredMcpServers/omitClaudeMd/maxTurns/criticalSystemReminder/background/isolation/effort fields.
- **SPEC-V3-AGT-002** — Built-in moai agents (T2-AGT-05; gm#66). Ship Explore / Plan moai-augmented definitions (SPEC-aware variants of CC built-ins).
- **SPEC-V3-AGT-003** — Fork subagent primitive (simplified) (T2-AGT-04; gm#67). Omit `subagent_type` → child inherits parent's system prompt; depth cap 2.
- **SPEC-V3-TEAM-001** — Teammate mailbox v2 schemas (T2-TEAM-01, T3-TEAM-03; gm#71, gm#72, gm#75, gm#76). 10 typed messages + plan-approval flow + TEAMMATE_MESSAGES_UI_CAP 50.

### 8.4 Skill SPECs (2 SPECs)

- **SPEC-V3-SKL-001** — Skill frontmatter v2 bundle + `context: fork` (T1-SKL-01, T2-SKL-02; gm#87, gm#88, gm#89, gm#90, gm#99, gm#100, gm#101, gm#104, gm#105, gm#106, gm#107). Add paths/effort conditional activation; $ARGUMENTS[N] / $N / $name substitution; ${CLAUDE_SKILL_DIR} / ${CLAUDE_SESSION_ID} body substitution; `context: fork` sub-agent execution with walk-up nested `.claude/skills/` discovery and realpath dedup.
- **SPEC-V3-SKL-002** — Skill Drift Detection & Resolution (T2-SKL-03; gm#184). `moai doctor skill --drift` detection + `moai migrate` M04 auto-resolution (template → local byte-identical copy of 3 missing skills) + drift-recurrence warning. Scheduled in Phase 5 Internal Cleanup per SPEC frontmatter.

### 8.5 UI/UX SPECs (1 SPEC)

- **SPEC-V3-OUT-001** — Output contract v2 (T2-OUT-01, T3-OUT-02, T3-OUT-03, T3-OUT-04; gm#175, gm#176, gm#177, gm#178, gm#179, gm#180, gm#181, gm#182). diff --git emission + structured validation errors + Progress: N/M + language-fenced code blocks + OSC-8 file-path links + StatusIcon + CC-compat output-style frontmatter.

### 8.6 Bootstrap / CLI / Plugin / Migration / Schema SPECs (6 SPECs)

- **SPEC-V3-SCH-001** — Formal config schemas (T1-SCH-01; gm#156, gm#157, gm#163). validator/v10 tags + JSON Schema export + `moai doctor config --fix`.
- **SPEC-V3-SCH-002** — Settings source layering 3-tier (T1-SCH-02; gm#138, gm#164 scope-reduced). user / project / local with deep-merge precedence.
- **SPEC-V3-MIG-001** — Versioned migration framework (T1-MIG-01; gm#149, gm#150, gm#151). `CURRENT_MIGRATION_VERSION` counter + ordered runner + conservative trigger list + opt-out env.
- **SPEC-V3-MIG-002** — Initial migration set M01–M05 (T1-MIG-02; gm#183, gm#184, gm#185, gm#190, gm#191). template_version sync + .agency/ archival + hook wrapper drift + skill drift + legacy cleanup.
- **SPEC-V3-PLG-001** — Plugin system v1 (T2-PLG-01, T2-PLG-02; gm#140, gm#141 scope-reduced, gm#142 scope-reduced, gm#143, gm#144, gm#145). Manifest + install scopes (user/project/local) + CLI surface (install/uninstall/update/list) + marketplace (github + directory only) + validation.
- **SPEC-V3-CLI-001** — Startup profiler + --bare mode (T2-CLI-01, T3-CLI-02, T3-CLI-03, T3-CLI-04, T3-CLI-05; gm#127, gm#128, gm#133, gm#135, gm#136). profileCheckpoint markers + `--bare` minimal mode + cobra PersistentPreRunE migration wiring + `moai completion <shell>`.

### 8.7 Internal Cleanup + moai-unique SPECs (3 SPECs)

Internal Cleanup (2 SPECs):

- **SPEC-V3-CLN-001** — Template drift resolution (T1-CLN-01; gm#184, gm#185). Deploy 3 missing skills + handle-permission-denied.sh via M03/M04 (tooling layer — `moai doctor template-drift` / `skill-drift` CLIs; migration step Go files live in SPEC-V3-MIG-002).
- **SPEC-V3-CLN-002** — Legacy code removal (T1-CLN-03; gm#186, gm#187, gm#188, gm#189). Delete .go.bak + stale coverage + fix ADR-011 comment via M05 + direct source edit (tooling layer — `moai doctor legacy-cleanup` CLI; migration step Go files live in SPEC-V3-MIG-002).

moai-unique (1 SPEC):

- **SPEC-V3-SPEC-001** — SPEC-to-SPEC chaining (T2-DIFF-01; moai-unique, not in CC gap matrix). SPEC inheritance (`inherits:` field) + SPEC templates + dependency graph validation + lifecycle transitions (`spec-first` → `spec-anchored` → `spec-as-source`).

Note: The retired third CLN slot (`.agency/` archival + docs-site locale sync, previously proposed) has been absorbed into SPEC-V3-MIG-002 per §3.9 Ownership Split. Only CLN-001 and CLN-002 remain active in v3.

### 8.8 Migration Tool SPEC (1 SPEC)

- **SPEC-V3-MIGRATE-001** — `moai migrate v2-to-v3` tool (Phase 7). CLI wiring + dry-run / yes / rollback / only flags + 4-locale migration guide generation.

### SPEC count summary

| Grouping | SPECs |
|----------|------:|
| Hooks/Commands | 9 |
| Memory | 2 |
| Agent/Team | 4 |
| Skill | 2 |
| UI/UX | 1 |
| Bootstrap/CLI/Plugin/Migration/Schema | 6 |
| Internal Cleanup + moai-unique | 3 |
| Migration tool | 1 |
| **Total SPECs** | **28** |
| **Total REQs** | **545** |

(Exceeds the 20-25 target upper bound slightly at 28 because the Hooks/Commands grouping naturally decomposes into 9 discrete units (6 HOOKS + 3 CMDS) — the hook subsystem is deep and the CMDS theme covers three orthogonal extensions (frontmatter, args, inline bash). Wave 4 may consolidate adjacent SPECs if scope permits.)

*(REQ counts auto-derived via `grep -c '^- \*\*REQ-' .moai/specs/SPEC-V3-*/spec.md`; pre-Stage 2 baseline 545.)*

---

## 9. Open Questions (defer to implementation)

Each question carries a recommended default. Wave 4 SPEC writers may override with explicit justification.

1. **Hook exit-code deprecation window**
   - Pros: Shorter = less legacy code; Longer = more user adoption time.
   - Cons: Shorter = breaks users who skip minor versions; Longer = carrying dead code.
   - **Recommended default**: v3.0 dual-parse → v3.2 warn-only (both shapes accepted) → v4.0 JSON-only.

2. **Plugin system v1 scope**
   - Pros (reduce): ship faster; lower risk; no source-precedence complexity.
   - Cons (reduce): plugin authors can't ship hooks, reducing appeal.
   - **Recommended default**: Lock v1 to skills + agents + commands ONLY. No hooks/MCP/outputStyles in v3.0. Re-evaluate for v3.2 based on plugin author feedback.

3. **Memory LLM relevance default**
   - Pros (opt-in): predictable cost; no surprise API bills.
   - Cons (opt-in): users miss out on relevance improvement; telemetry slow.
   - **Recommended default**: opt-in (default off) in v3.0. Telemetry-driven evaluation for default-on in v3.2.

4. **Settings source layering v1 tier count**
   - Pros (3-tier): enough for most users; simple to explain.
   - Cons (3-tier): doesn't match CC's 6-tier; enterprise policy/managed settings deferred.
   - **Recommended default**: 3-tier (user/project/local) in v3.0; extend to 6-tier (adds policy/flag/managed) in v3.2.

5. **Fork subagent scope**
   - Pros (simplified): ships in v3.0; low recursion risk; minimal new infra.
   - Cons (simplified): doesn't achieve cache-identical prefix sharing (CC's key win).
   - **Recommended default**: Simplified (inherit system prompt) in v3.0; full cache-identical prefix in v3.1.

6. **Schema technology choice**
   - Pros (validator/v10): Go-native; minimal new deps; familiar.
   - Cons (validator/v10): no cross-language reuse; no declarative DSL.
   - Pros (CUE): declarative; cross-language capable.
   - Cons (CUE): new language to learn; external tooling; adds substantial build complexity.
   - **Recommended default**: validator/v10 in v3.0. Revisit CUE if cross-language schema sharing becomes requirement.

7. **In-process teammate timing**
   - Pros (v3.0): ship parity faster.
   - Cons (v3.0): requires goroutine + context.Context identity isolation; test matrix heavy.
   - **Recommended default**: Defer to v3.1. v3.0 ships mailbox schemas only.

8. **CG Mode vs CC Bridge**
   - Pros (adopt Bridge): name-compat with CC internals.
   - Cons (adopt Bridge): CC's Bridge is Anthropic-cloud specific (W1.3 §1).
   - **Recommended default**: Reject CC Bridge transport; adopt `SDKControlRequest/Response` message NAMING for internal CG channel schemas (provides naming consistency without implementation coupling).

9. **`omitClaudeMd` default**
   - Pros (default-true): massive token savings across 22 agents.
   - Cons (default-true): agents lose CLAUDE.md context silently; regression risk.
   - **Recommended default**: Default `false`. Opt-in per agent via frontmatter. Explicit candidates: `researcher`, `plan-auditor` (adversarial, context less relevant), `evaluator-active` (scoring role).

10. **Migration trigger aggression**
    - Pros (aggressive — every cobra command): migrations can't be skipped.
    - Cons (aggressive): surprise on `moai version`, `moai status`.
    - **Recommended default**: Conservative — fires only on `moai init`, `moai update`, `moai doctor`, `moai migrate`. `moai version` and `moai status` stay side-effect-free.

11. **docs-site 4-locale translation SLA**
    - Pros (synchronous, block release): no locale lag.
    - Cons (synchronous): release cadence slows; small doc changes block release.
    - **Recommended default**: Per CLAUDE.local.md §17.3 — canonical ko → en within 48h → zh/ja within 72h. Release gated by 95%+ locale parity (not 100%) for major releases.

12. **v3.0 dual-parse enforcement**
    - Pros (strict): cleaner upgrade; fewer moving parts.
    - Cons (strict): hook authors must update wrappers in v3.0 cycle.
    - **Recommended default**: Dual-parse active in v3.0; warn phase v3.2; removal v4.0. Telemetry via `moai doctor hook --validate` counts legacy wrappers for graduation decision.

---

## 10. Non-Goals (what v3 explicitly does NOT do)

These items are explicitly OUT OF SCOPE for v3.0.0. Most are Tier 4 from `priority-roadmap.md`.

- **T4-BRIDGE-01 — Remote Control Bridge** (gm#86; W1.3 §1.7). Reject. CC's bridge subsystem (33 files, 500KB+) is tied to `api.anthropic.com` OAuth and cannot be reused. moai's tmux-based CG Mode is sufficient.
- **T4-BUDDY-01 — Buddy Sprite**. Reject. Gamified companion sprite — zero business value for a professional dev tool.
- **T4-SDK-01 — Agent SDK re-export barrel** (gm#171). Reject. moai is a CLI, not a library; no SDK consumers exist.
- **T4-MCP-01 — MCP server entrypoint (`moai mcp serve`)** (gm#170). Defer. moai's tools are Go binary subcommands, not MCP-exposed. Revisit if demand materializes.
- **T4-REPL-01 — REPL TUI** (gm#173). Reject. moai is non-interactive by design.
- **T4-PRINT-01 — Headless mode `--print`** (gm#172). Reject. moai is always headless; concept doesn't apply.
- **T4-AUTH-01 — OAuth subcommands**. Reject. moai relies on CC's auth.
- **T4-POLICY-01 — Policy limits / managed remote settings**. Defer. Enterprise-gated; revisit when enterprise contracts materialize.
- **T4-AUTOMODE-01 — auto-mode classifier command**. Reject. moai uses its own permissionMode config via `.moai/config/sections/quality.yaml`.
- **T4-REMOTE-01 — `isolation: remote`** (gm#84). Reject. Anthropic-internal CCR infrastructure.
- **T4-PLG-FULL — Full CC-parity plugin system** (gm#142, gm#146, gm#147, gm#148). Defer to v3.2+. v3.0 ships reduced scope (skills+agents+commands only).
- **T4-COMPACT-01 — 5-layer compaction pipeline** (gm#36, gm#37). Reject. snip → microcompact → context-collapse → autocompact → reactive is CC-runtime's QueryEngine concern.
- **T4-COST-01 — Cost tracker with OpenTelemetry counters** (gm#42, gm#43). Defer. Adds OTEL deps; conflicts with moai's 9-dep philosophy.
- **T4-COORD-01 — Coordinator Mode** (gm#85). Defer. moai already has manager-strategy + agent orchestration.

Additionally rejected mid-design:

- Full 6-tier settings in v3.0 (reduced to 3-tier per §9 question #4).
- Cache-identical fork subagent prefix (deferred to v3.1 per §9 question #5).
- In-process teammate backend (deferred to v3.1 per §9 question #7).
- Plugin hooks / MCP / outputStyles capabilities (deferred to v3.2 per §9 question #2).
- `omitClaudeMd: true` default across all agents (default-false per §9 question #9).
- KAIROS daily-log mode with `/dream` distillation (deferred per v3-themes §2 open question).

---

## Appendix A: Design Principles for v3

Consolidated from §1.2 and §3 rationale sections. These principles govern every SPEC, PR, and migration in the v3 release.

1. **Evidence-grounded**: every design decision cites a Wave 1 finding (file:line) or gap-matrix row ID. No speculation.
2. **Reversible**: every breaking change has dual-parse or opt-out env escape for one minor version.
3. **Go-idiomatic**: struct tags over external DSLs; stdlib over heavy frameworks; 9-direct-dep budget preserved.
4. **Opus 4.7 one-turn**: agent prompts deliver intent + constraints + completion criteria in a single message.
5. **CC parity where it matters**: emit output CC's renderer can parse; reply to hook events with schema-valid payloads; validate configs with schemas.
6. **moai identity preserved**: SPEC-First DDD, TRUST 5, TAG system, @MX protocol, 50-skill / 22-agent catalog stay.
7. **Conservative migrations**: fire only on explicit user actions (init/update/doctor/migrate), never on read-only commands.
8. **Additive > breaking**: prefer optional frontmatter fields over required changes; grow schemas, don't rewrite them.
9. **Template-first changes**: new .claude/.moai/.agency/ content added to `internal/template/templates/` FIRST (per CLAUDE.local.md §2).
10. **Language-neutral templates**: treat all 16 supported languages equally in `internal/template/templates/`; bias allowed only in `CLAUDE.local.md` / `settings.local.json` / `_test.go` (per CLAUDE.local.md §15).

---

## Appendix B: Glossary of new terms

| Term | Definition |
|------|-----------|
| **Hook Protocol v2** | The rich-JSON hook output contract introduced in v3.0 (superset of v2's exit-code-only). See §3.1. |
| **HookSpecificOutput** | Discriminated union of per-event payloads, tagged by `hookEventName`. Implemented as Go interface with per-event concrete type. |
| **CLAUDE_ENV_FILE** | Temp-file mechanism where SessionStart/Setup/CwdChanged/FileChanged hooks write bash exports for subsequent BashTool commands. Matches CC (W1.1 §4.3). |
| **MigrationStep** | Interface implemented by each versioned migration (Version/ID/Description/IsIdempotent/PreConditionsMet/DryRun/Apply/Rollback). |
| **Source precedence pipeline** | Ordered merge of hook/config sources (user → project → local in v3.0). |
| **Fork subagent** | Sub-agent spawned without `subagent_type`, inheriting parent's system prompt. v3.0 ships inherit-prompt; v3.1 adds cache-identical prefix. |
| **`omitClaudeMd`** | Agent frontmatter field; when true, agent skips loading CLAUDE.md hierarchy. Saves 5–15Gtok/week per CC BQ data. Default false in v3.0. |
| **Mailbox v2** | Structured discriminated-union team message schema (10 types); replaces v2 ad-hoc JSON. |
| **Teammate mailbox** | Message queue between team lead and teammates in team mode; typed messages with validator/v10 schemas. |
| **Plan-approval flow** | Team coordination mechanism where lead approves/rejects teammate plans via `plan_approval_request` / `plan_approval_response` messages. |
| **SPEC chaining** | moai-unique mechanism for SPEC inheritance (`inherits:` field), SPEC templates, and lifecycle transitions (spec-first → spec-anchored → spec-as-source). |
| **Harness-based quality routing** | Existing (preserved) 3-level (minimal/standard/thorough) quality routing in `.moai/config/sections/harness.yaml`. |
| **4-type memory taxonomy** | CC-native memory classification: user / feedback / project / reference. moai enforces this in v3 MEMORY.md frontmatter. |
| **Memory freshness** | Computed age of MEMORY.md file via mtime; >24h wraps content in `<system-reminder>` staleness warning. |
| **Dual-parse shim** | Transitional parser that tries v3 JSON first, falls back to v2 exit-code semantics on parse error. Active v3.0 → v4.0. |
| **Tier 1 / Tier 2 / Tier 3 / Tier 4** | Priority roadmap tiers from `priority-roadmap.md`. Tier 1 = Critical must-ship (18 items in v3.0). Tier 2 = Strategic differentiators (14 items, Phase 6a/6b). Tier 3 = Nice-to-have polish (11 items, opportunistic). Tier 4 = Rejected or deferred (13 items, see §10 Non-Goals). |
| **Harness level** | Quality-depth routing setting from `.moai/config/sections/harness.yaml`. Values: `minimal` (fast validation for simple changes), `standard` (default checks), `thorough` (full evaluator-active + TRUST 5 for complex SPECs). Auto-determined by Complexity Estimator. |
| **Sprint Contract** | Pre-iteration negotiation artifact in the GAN Loop (`.claude/rules/moai/design/constitution.md` §11). evaluator-active issues acceptance checklist + priority dimension + test scenarios + pass conditions; expert-frontend accepts or requests adjustment. Stored in `.moai/sprints/` per `design.yaml` `sprint_contract.artifact_dir`. Required at `thorough` harness level. |
| **BC-ID** | Breaking Change identifier format `BC-###` (e.g., BC-001..BC-008 in §4). Each BC-ID names one discrete breaking change, pairs with a dual-parse shim or opt-out env var, and cross-references via SPEC frontmatter `bc_id: [BC-XXX]`. |

---

## Appendix C: References (Wave 1 findings file:section index)

Every substantive claim in this master document traces to one of the Wave 1 findings files or a moai-adk-go source file.

### Wave 1 findings (6 files)

- **W1.1** `.moai/design/v3-research/findings-wave1-hooks-commands.md` (1,148 lines)
  - §1 27-event hook catalog
  - §2 4 hook types
  - §3 hook configuration schema
  - §4 hook output / env / timeout / bidir
  - §6 skill discovery
  - §7 frontmatter parser
  - §8 argument substitution
  - §9 skill body substitution
  - §11 skill `context: inline` vs `fork`
  - §12 security primitives (ssrfGuard)
  - §14 v2.1.111 summary
- **W1.2** `findings-wave1-query-context.md` (1,083 lines)
  - §1 QueryEngine (CC-runtime, not moai)
  - §2 Compaction pipeline (CC-runtime)
  - §3 System prompt assembly
  - §5 Cost accounting
  - §6 Memory (4-type, truncation, freshness, path validation, LLM relevance)
  - §7 Migration patterns
  - §9 Integration summary
- **W1.3** `findings-wave1-agent-team.md` (1,015 lines)
  - §1 Bridge subsystem (REJECT)
  - §2 Coordinator mode (DEFER)
  - §3 Buddy (REJECT)
  - §4 Agent frontmatter fields
  - §6 Team mailbox message types
  - §7 Team backends (tmux / iterm2 / in-process)
  - §8 Integration summary
- **W1.4** `findings-wave1-ui-ux.md` (1,304 lines)
  - §2 Component catalog
  - §8 Output contract gaps (diff / validation / progress / icons / links)
- **W1.5** `findings-wave1-bootstrap-cli.md` (1,202 lines)
  - §1 CLI shim + init
  - §2 Commander + preAction
  - §3 Entry points
  - §4 Plugin system
  - §5 Migration framework
  - §6 Schema (Zod v4)
  - §7 Output styles
  - §9 Recommendation summary
- **W1.6** `findings-wave1-moai-current.md` (909 lines)
  - §1 CLI subcommand inventory
  - §2 internal/ packages
  - §3 pkg/ surface
  - §4 Template system (go:embed)
  - §5 Hook implementations
  - §6 Agent catalog
  - §7 Skill catalog
  - §8 Command catalog
  - §9 Rules hierarchy
  - §10 Config schema
  - §11 SPEC state
  - §12 docs-site
  - §13 Tests & coverage
  - §14 Dependencies
  - §15 Self-identified weaknesses
  - §16 Version & release

### Wave 2 synthesis (3 files)

- `.moai/design/v3-research/gap-matrix.md` — 194 gap rows across 7 sections (Hook / Memory / Agent+Team / Command+Skill / Bootstrap+CLI+Plugin+Migration / UI+UX / moai self-identified).
- `.moai/design/v3-research/priority-roadmap.md` — 56 tiered items (18 Tier-1 / 14 Tier-2 / 11 Tier-3 / 13 Tier-4).
- `.moai/design/v3-research/v3-themes.md` — 9 themes + summary matrix + cross-theme dependencies.

### Key moai-adk-go source anchors

- `CLAUDE.md` (project instructions, 17 sections, 19 HARD rules)
- `CLAUDE.local.md` (private, §17 docs-site 4-locale sync rules)
- `.claude/rules/moai/core/moai-constitution.md` (Agent Core Behaviors, Opus 4.7 prompt philosophy)
- `.claude/rules/moai/design/constitution.md` v3.3.0 (FROZEN/EVOLVABLE zones)
- `.claude/rules/moai/development/coding-standards.md` (language policy, file-size limits)
- `internal/hook/types.go:19-114` (27 EventType constants)
- `internal/hook/types.go:167-311` (HookInput/HookOutput schema — partial)
- `internal/hook/registry.go:18-325` (registry implementation)
- `internal/cli/deps.go:151-186` (handler registrations — 28 calls)
- `internal/template/embed.go:25` (go:embed all:templates)
- `internal/template/deployer.go:21-30` (Deployer interface)
- `internal/cli/migrate_agency.go:569-595` (existing agency migration)
- `pkg/version/version.go:7` (Version = "v2.12.0" default)
- `.moai/config/sections/system.yaml` (moai.version, template_version)
- `.moai/config/sections/project.yaml:template_version: v2.7.22` (drift source)

---

End of Master Design Document. Wave 4 input: SPEC IDs in §8 are the authoritative set for Wave 4 SPEC authoring. Wave 4 may split or merge SPECs for implementation-scope reasons with explicit justification.

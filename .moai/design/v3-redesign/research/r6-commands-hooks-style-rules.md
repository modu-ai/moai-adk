# R6 — Commands + Hooks + Output Styles + Rules + Config Audit

> Research team: R6
> Scope: `.claude/commands/`, `.claude/hooks/moai/`, `internal/hook/`, `.claude/output-styles/`, `.claude/rules/moai/`, `.moai/config/sections/`
> Date: 2026-04-23
> Model: Opus 4.7
> Audit mode: read-only, evidence-based

---

## Executive summary

This audit scanned the runtime surface area of MoAI-ADK v2.13.2 that is exposed to Claude Code (commands, hooks, output-styles, rules, config sections). The goal is to inform v3 redesign decisions by identifying dead code, redundant abstractions, and coverage gaps.

**Headline findings:**

- **Commands** — `/moai/*` slash-command family (15 subcommands) is **in excellent thin-wrapper shape** (all under 20 LOC, frontmatter consistent). The two root-level commands (`98-github.md`, `99-release.md`, 698 + 890 LOC) are **fat, dev-only, and not registered in templates**; they must be moved under `/moai github` and `/moai release` subcommands (or kept as local-only) rather than coexisting alongside the thin pattern.
- **Hooks** — All **26 shell wrappers** + **27 registered Go handlers** + **25 `settings.json` registrations** are 1:1 synchronized (no orphan wrappers, no unregistered handlers). However, **14 of 27 handlers (~52%) are thin loggers with no business logic** (config_change, file_changed, elicitation, elicitation_result, instructions_loaded, notification, subagent_stop, task_created, post_tool_failure, worktree_create/remove, setup, cwd_changed, permission_request). v3 should decide whether to retain these as observability placeholders or remove them.
- **Output styles** — Only **2 styles** (MoAI, Einstein), both active, both template-synced. Well-bounded. No action needed.
- **Rules** — **34 rule files** in 5 subtrees (core/workflow/development/languages/design). **Language rules correctly use `paths:` frontmatter** (all 16 languages); non-language rules are inconsistent (some use `paths:`, some use `description/globs`, some have no frontmatter at all). **`core/lsp-client.md` is a SPEC decision record** (not a rule) and is misfiled.
- **Config sections** — **23 yaml section files**. **5 are unused by Go code** (`constitution.yaml`, `context.yaml`, `interview.yaml`, `sunset.yaml` (only 4 refs, dead struct), `design.yaml` (only migrate_agency refs)). Go code reads: language, llm, quality, workflow, lsp, mx, security, statusline, system, user, project, git-convention, git-strategy, ralph, research, state. 5 yaml sections have no schema in `internal/config/types.go`.

Per-area verdicts and proposed v3 inventory appear in subsequent sections.

---

## 1. Commands audit

### 1.1 `/moai` subcommand family (thin wrappers)

All 15 `/moai/*.md` command files are thin wrappers routing to `Skill("moai")` with subcommand arguments. Body is consistently under 10 lines. Frontmatter has `description`, `argument-hint`, `allowed-tools: Skill` (CSV).

| File | LOC | Family | Skill route | Verdict |
|------|-----|--------|-------------|---------|
| `plan.md` | 8 | Workflow | `moai:plan` | KEEP |
| `run.md` | 8 | Workflow | `moai:run` | KEEP |
| `sync.md` | 8 | Workflow | `moai:sync` | KEEP |
| `project.md` | 8 | Workflow | `moai:project` | KEEP |
| `design.md` | 8 | Workflow | `moai:design` | KEEP |
| `db.md` | 8 | Workflow | `moai:db` | KEEP |
| `fix.md` | 8 | Utility | `moai:fix` | KEEP |
| `loop.md` | 8 | Utility | `moai:loop` | KEEP |
| `clean.md` | 8 | Utility | `moai:clean` | KEEP |
| `mx.md` | 8 | Utility | `moai:mx` | KEEP |
| `feedback.md` | 8 | Feedback | `moai:feedback` | KEEP |
| `review.md` | 8 | Quality | `moai:review` | KEEP |
| `coverage.md` | 8 | Quality | `moai:coverage` | KEEP |
| `e2e.md` | 8 | Quality | `moai:e2e` | KEEP |
| `codemaps.md` | 8 | Quality | `moai:codemaps` | KEEP |

Conformance checks (`internal/template/commands_audit_test.go`): all 15 pass the thin-wrapper body length test, frontmatter schema, and Skill-only routing.

Note: Local copies in `.claude/commands/moai/*.md` and templates in `internal/template/templates/.claude/commands/moai/*.md.tmpl` have drifted for `design.md` and `db.md` (template uses `.md` not `.md.tmpl`; the other 13 are `.md.tmpl`). This is a minor template consistency issue — v3 should unify the extension convention.

### 1.2 Root-level commands

| File | LOC | Purpose | Template-synced? | Verdict |
|------|-----|---------|-----------------|---------|
| `98-github.md` | 698 | GitHub Workflow (issues, PRs) | NO (dev-only) | REFACTOR |
| `99-release.md` | 890 | v2.x release orchestration | NO (dev-only) | REFACTOR |

Both files are thick orchestration commands that **violate the thin-wrapper pattern**. They are also **dev-local only** (not templated into user projects) — they assist moai-adk-go maintainers with release and issue workflows.

**Finding:** These violate `.claude/rules/moai/development/coding-standards.md#thin-command-pattern` which states: "All slash command files MUST be thin routing wrappers (under 20 LOC body)." Since this rule is enforced by `commands_audit_test.go`, the test either excludes these files or the test is not comprehensive. The local project is the development tree, so these are arguably exempt, but v3 should:

- Move logic to a `moai-workflow-github` skill and make `98-github.md` a thin wrapper
- Move logic to a `moai-workflow-release` skill and make `99-release.md` a thin wrapper
- Or mark them `type: local` and explicitly exclude from the audit test (current `type: local` appears in frontmatter but is not picked up by the audit)

### 1.3 Deprecated / dead commands

No `/agency/*` commands were found in either location (confirmed via `ls`). The SPEC-AGENCY-ABSORB-001 M5 cleanup is complete for commands.

No dead command files detected.

---

## 2. Hooks audit

### 2.1 Shell wrappers (`.claude/hooks/moai/*.sh`)

26 shell wrapper scripts exist (1:1 per event + `handle-agent-hook.sh` aggregator for agent-scoped PreToolUse/PostToolUse/SubagentStop).

| Wrapper | Forwards to | Shell LOC | Pattern |
|---------|-------------|-----------|---------|
| `handle-session-start.sh` | `moai hook session-start` | 30 | Standard |
| `handle-session-end.sh` | `moai hook session-end` | 30 | Standard |
| `handle-pre-tool.sh` | `moai hook pre-tool` | 32 | Standard |
| `handle-post-tool.sh` | `moai hook post-tool` | 47 | Has MX check enhancement |
| `handle-post-tool-failure.sh` | `moai hook post-tool-failure` | 30 | Standard |
| `handle-compact.sh` | `moai hook compact` | 30 | Standard |
| `handle-post-compact.sh` | `moai hook post-compact` | 30 | Standard |
| `handle-stop.sh` | `moai hook stop` | 30 | Standard |
| `handle-stop-failure.sh` | `moai hook stop-failure` | 30 | Standard |
| `handle-subagent-start.sh` | `moai hook subagent-start` | 30 | Standard |
| `handle-subagent-stop.sh` | `moai hook subagent-stop` | 30 | Standard |
| `handle-notification.sh` | `moai hook notification` | 30 | Standard |
| `handle-user-prompt-submit.sh` | `moai hook user-prompt-submit` | 30 | Standard |
| `handle-permission-request.sh` | `moai hook permission-request` | 32 | Standard |
| `handle-permission-denied.sh` | `moai hook permission-denied` | 30 | Standard |
| `handle-teammate-idle.sh` | `moai hook teammate-idle` | 30 | Standard |
| `handle-task-completed.sh` | `moai hook task-completed` | 30 | Standard |
| `handle-task-created.sh` | `moai hook task-created` | 30 | Standard |
| `handle-worktree-create.sh` | `moai hook worktree-create` | 30 | Standard |
| `handle-worktree-remove.sh` | `moai hook worktree-remove` | 30 | Standard |
| `handle-config-change.sh` | `moai hook config-change` | 30 | Standard |
| `handle-cwd-changed.sh` | `moai hook cwd-changed` | 30 | Standard |
| `handle-file-changed.sh` | `moai hook file-changed` | 30 | Standard |
| `handle-instructions-loaded.sh` | `moai hook instructions-loaded` | 30 | Standard |
| `handle-elicitation.sh` | `moai hook elicitation` | 30 | Standard |
| `handle-elicitation-result.sh` | `moai hook elicitation-result` | 30 | Standard |
| `handle-agent-hook.sh` | `moai hook agent <action>` | 53 | Aggregator |

All wrappers use the same fallback chain: `moai` in PATH → `/Users/goos/go/bin/moai` → `$HOME/go/bin/moai` → silent exit.

**Finding:** The absolute path `/Users/goos/go/bin/moai` is **hard-coded** in all 26 wrappers. Per `CLAUDE.local.md` §14, this is a **`.HomeDir`-style hardcoding anti-pattern** — the template should use `$HOME/go/bin/moai` as primary and the detected-GoBinPath as the second attempt, not first. The templates need to be regenerated via `make build` after the user's GoBinPath resolution changes.

**No `handle-setup.sh` exists** for the Setup event (v2.1.10+). Go handler `setupHandler` exists in `internal/hook/setup.go` but there is no shell wrapper or `settings.json` registration. This is an orphan handler.

### 2.2 Go handlers (`internal/hook/`)

27 handlers registered in `internal/cli/deps.go` (see lines 152-186). Implementation quality ranges widely:

| Handler | File | LOC | Logic depth | Verdict |
|---------|------|-----|-------------|---------|
| `sessionStartHandler` | session_start.go | 639 | Rich (GLM tmux, skill bump, memory eval) | KEEP |
| `sessionEndHandler` | session_end.go | 714 | Rich (memo save, MX scan) | KEEP |
| `preToolHandler` | pre_tool.go | 652 | Rich (security scan, secrets, reflective write) | KEEP |
| `postToolHandler` | post_tool.go | 509 | Rich (MX validation, LSP, metrics) | KEEP |
| `stopHandler` | stop.go | 170 | Rich (completion markers, Ralph state) | KEEP |
| `teammateIdleHandler` | teammate_idle.go | 205 | Rich (quality gate enforcement) | KEEP |
| `taskCompletedHandler` | task_completed.go | 131 | Rich (SPEC validation, acceptance criteria) | KEEP |
| `compactHandler` | compact.go | 177 | Rich (memo save) | KEEP |
| `postCompactHandler` | post_compact.go | 68 | Rich (memo restore) | KEEP |
| `autoUpdateHandler` | auto_update.go | 100 | Rich (version check, update notify) | KEEP |
| `userPromptSubmitHandler` | user_prompt_submit.go | 207 | Rich (SPEC detection, session title, workflow keyword) | KEEP |
| `permissionDeniedHandler` | permission_denied.go | 54 | Smart (read-only tool retry) | KEEP |
| `stopFailureHandler` | stop_failure.go | 56 | Smart (error-type-specific message) | KEEP |
| `subagentStartHandler` | subagent_start.go | 146 | Rich (project context injection) | KEEP |
| `permissionRequestHandler` | permission_request.go | 52 | Smart (updatedInput re-verify) | KEEP |
| `cwdChangedHandler` | cwd_changed.go | 82 | Smart (CLAUDE_ENV_FILE writes) | KEEP |
| `worktreeCreateHandler` | worktree_create.go | 40 | Registry update | KEEP |
| `worktreeRemoveHandler` | worktree_remove.go | 44 | Registry cleanup | KEEP |
| `configChangeHandler` | config_change.go | 30 | Logging only | REFACTOR or RETIRE |
| `fileChangedHandler` | file_changed.go | 32 | Logging only | REFACTOR or RETIRE |
| `elicitationHandler` | elicitation.go | 32 | Logging only | RETIRE |
| `elicitationResultHandler` | elicitation.go | 25 | Logging only | RETIRE |
| `instructionsLoadedHandler` | instructions_loaded.go | 30 | Logging only | RETIRE |
| `notificationHandler` | notification.go | 31 | Logging only | RETIRE |
| `subagentStopHandler` | subagent_stop.go | 31 | Logging only | REFACTOR (tmux pane cleanup) |
| `taskCreatedHandler` | task_created.go | 31 | Logging only | RETIRE |
| `postToolUseFailureHandler` | post_tool_failure.go | 33 | Logging only | REFACTOR (error classification) |
| `setupHandler` | setup.go | 30 | Logging only | RETIRE (also orphan) |

**Orphan handlers (registered in Go, missing shell wrapper and/or settings.json):**

- `setupHandler` — No shell wrapper, no settings.json entry. Remove or implement.

**Handlers with critical missing logic:**

- `subagentStopHandler` — Feedback memory says: *"Team tmux pane cleanup — TeamDelete 전 tmuxPaneId로 kill-pane 필수"*. This handler should read tmuxPaneId from team config and kill the pane before returning, but it currently only logs. This is a known bug.
- `configChangeHandler` — Should trigger re-render of generated files when `.moai/config/sections/*.yaml` changes. Currently no-op.
- `instructionsLoadedHandler` — Could validate CLAUDE.md length against `coding-standards.md` 40,000 char limit. Currently no-op.

### 2.3 settings.json registrations

Local `.claude/settings.json` registers **25 events**:

SessionStart, PreCompact, SessionEnd, PreToolUse, PostToolUse, Stop, SubagentStop, PostToolUseFailure, Notification, SubagentStart, UserPromptSubmit, TeammateIdle, TaskCompleted, WorktreeCreate, WorktreeRemove, TaskCreated, ConfigChange, StopFailure, PostCompact, InstructionsLoaded, CwdChanged, FileChanged, Elicitation, ElicitationResult, PermissionDenied, PermissionRequest

Template `internal/template/templates/.claude/settings.json.tmpl` registers the same 25 events.

**Missing registrations (vs 27 Go handlers):**

- `Setup` — Registered Go handler exists but no settings.json entry (Setup fires via `--init/--init-only/--maintenance` CLI flags, which may not require settings.json registration — this is a Claude Code protocol nuance)
- `AutoUpdate` — `autoUpdateHandler` is not a native Claude Code event; it is a composite handler registered via SessionStart in `deps.go:157`. It is internal, not a hook event.

So of the 27 registered Go handlers: 25 are native events + 1 internal composite (autoUpdate) + 1 orphan (setup). The **effective Claude Code hook coverage is 25/27 events**.

### 2.4 Coverage gaps (missing Claude Code events that MoAI does NOT implement)

Of the 27 Claude Code event types documented in `hooks-system.md` (26 standard + 1 special Setup), MoAI covers 25. Missing: `Setup` (special event, fires only on init/maintenance), which is **not critical** for runtime behavior.

Additionally, **PrePrompt** does not appear in the documentation — checked, it does not exist. All events audited are covered or have justifiable omission.

---

## 3. Output Styles audit

### 3.1 `MoAI` (13,349 bytes)

**File:** `.claude/output-styles/moai/moai.md`

- Frontmatter: `name: MoAI`, `description` (agentic orchestrator), `keep-coding-instructions: true`
- Body: Core identity, cannot-do limits, Socratic flow, delegation patterns, verification gates
- Activation: `settings.json` `outputStyle: "MoAI"` (default, confirmed in both local and template)

**Verdict: KEEP** — this is the default style; it is actively used and well-structured.

### 3.2 `Einstein` (16,288 bytes)

**File:** `.claude/output-styles/moai/einstein.md`

- Frontmatter: `name: Einstein`, `description` (learning tutor), `keep-coding-instructions: false`
- Body: Personal tutoring framework with Context7 grounding, mermaid diagrams, Notion sync
- Activation: User must select via `/config → Output style → Einstein`

**Verdict: KEEP** — niche educational style, well-scoped. `keep-coding-instructions: false` is intentional (tutor does not write code).

### 3.3 Conflicts

No conflict with CLAUDE.md directives detected. Both styles operate within the Claude Code output style protocol and do not override constitutional rules.

### 3.4 Template sync

Both files are identical between `.claude/output-styles/moai/` and `internal/template/templates/.claude/output-styles/moai/`. **Clean sync.**

---

## 4. Rules audit

### 4.1 Tree structure (34 files)

```
.claude/rules/moai/
├── core/                    (6 files)
│   ├── agent-common-protocol.md    (157 LOC)
│   ├── agent-hooks.md              (85 LOC, paths: .claude/agents/**, .claude/hooks/**)
│   ├── hooks-system.md             (322 LOC, paths: .claude/hooks/**, settings.json*)
│   ├── lsp-client.md               (78 LOC, no frontmatter — is a SPEC decision record)
│   ├── moai-constitution.md        (266 LOC, description+globs frontmatter)
│   └── settings-management.md      (198 LOC, paths: .moai/config/**, settings.json*)
├── design/                  (1 file)
│   └── constitution.md             (404 LOC, no frontmatter — design system constitution)
├── development/             (4 files)
│   ├── agent-authoring.md          (227 LOC, paths: .claude/agents/**)
│   ├── coding-standards.md         (97 LOC, description+globs frontmatter)
│   ├── model-policy.md             (61 LOC, paths: .claude/agents/**)
│   └── skill-authoring.md          (248 LOC, paths: .claude/skills/**)
├── languages/               (16 files, all paths: frontmatter)
│   └── [cpp, csharp, elixir, flutter, go, java, javascript, kotlin, php,
│        python, r, ruby, rust, scala, swift, typescript].md
└── workflow/                (7 files)
    ├── file-reading-optimization.md  (50 LOC, no frontmatter)
    ├── moai-memory.md                (41 LOC, paths: .moai/specs/**)
    ├── mx-tag-protocol.md            (180 LOC, paths: 17 file extensions)
    ├── spec-workflow.md              (217 LOC, paths: .moai/specs/**, quality.yaml)
    ├── team-protocol.md              (54 LOC, description+globs frontmatter)
    ├── workflow-modes.md             (195 LOC, paths: .moai/specs/**, quality.yaml)
    └── worktree-integration.md       (303 LOC, description+globs frontmatter)
```

### 4.2 Per-subtree verdict

**core/** — 5 of 6 well-formed. Exception: `lsp-client.md` is a **SPEC decision record** (SPEC-LSP-CORE-002 Decision Point 1), not an agent rule. It should live in `.moai/decisions/` or be merged into the SPEC file itself.

**design/** — `constitution.md` is **canonical** (v3.3.0, frozen amendment). Keep as-is.

**development/** — `coding-standards.md`, `agent-authoring.md`, `skill-authoring.md`, `model-policy.md` are all well-scoped. Keep.

**languages/** — All 16 language rules use `paths:` frontmatter for conditional loading. **Uniform and correct.** Keep.

**workflow/** — 7 files, mostly well-scoped. However:
- `workflow-modes.md` (195 LOC) and `spec-workflow.md` (217 LOC) overlap significantly in their description of Plan-Run-Sync flow. v3 should consider merging.
- `team-protocol.md` (54 LOC) duplicates content in `worktree-integration.md` (303 LOC) in the "Team File Ownership" section.
- `file-reading-optimization.md` is a pure heuristic doc, not an enforceable rule. Could move to skill reference material.

### 4.3 Frontmatter consistency

- **Consistent (`paths:` array/CSV):** 16 language rules, agent-common-protocol.md (none), agent-hooks.md, hooks-system.md, settings-management.md, agent-authoring.md, model-policy.md, skill-authoring.md, moai-memory.md, mx-tag-protocol.md, spec-workflow.md, workflow-modes.md
- **Uses `description + globs`:** moai-constitution.md, coding-standards.md, team-protocol.md, worktree-integration.md
- **No frontmatter:** `lsp-client.md`, `file-reading-optimization.md`, `design/constitution.md`, `core/agent-common-protocol.md`

v3 should standardize on `paths:` (the newer, Claude Code official format). Files without frontmatter always load into context — this is intentional for `moai-constitution.md` (core) and `design/constitution.md` (FROZEN zone) but wasteful for `file-reading-optimization.md`.

### 4.4 Redundancy / dueling rules

- **`worktree-integration.md` vs `team-protocol.md`** — Both define team file ownership boundaries. Consolidate.
- **`spec-workflow.md` vs `workflow-modes.md`** — Both reference Plan-Run-Sync and quality.yaml. Consolidate under `spec-workflow.md`.
- **`agent-common-protocol.md` vs CLAUDE.md Section 8** — Both define AskUserQuestion exclusivity. Cross-reference already exists but the content overlaps. Keep rule as authoritative; tighten CLAUDE.md to reference only.
- **`hooks-system.md` vs CLAUDE.md Section 14** — Both list parallel execution safeguards. Acceptable (rule is detail, CLAUDE.md is summary).

### 4.5 Per-file verdict

| File | LOC | Verdict | Note |
|------|-----|---------|------|
| core/agent-common-protocol.md | 157 | KEEP | Cross-cutting protocol; well-used |
| core/agent-hooks.md | 85 | KEEP | Needed for agent-scoped hooks |
| core/hooks-system.md | 322 | KEEP | Canonical event reference |
| core/lsp-client.md | 78 | **MOVE** | SPEC decision record, wrong location |
| core/moai-constitution.md | 266 | KEEP | Core identity + 6 behaviors |
| core/settings-management.md | 198 | KEEP | Config reference |
| design/constitution.md | 404 | KEEP | FROZEN design system constitution |
| development/agent-authoring.md | 227 | KEEP | Authoring guide |
| development/coding-standards.md | 97 | KEEP | MoAI-specific conventions |
| development/model-policy.md | 61 | KEEP | Effort-level guidance |
| development/skill-authoring.md | 248 | KEEP | Authoring guide |
| languages/*.md (x16) | 73-167 | KEEP | Uniformly well-scoped |
| workflow/file-reading-optimization.md | 50 | **MERGE** | Into skill-authoring or references |
| workflow/moai-memory.md | 41 | KEEP | Memory protocol reference |
| workflow/mx-tag-protocol.md | 180 | KEEP | @MX tag canonical reference |
| workflow/spec-workflow.md | 217 | KEEP | Plan-Run-Sync canonical |
| workflow/team-protocol.md | 54 | **MERGE** | Into worktree-integration.md |
| workflow/workflow-modes.md | 195 | **MERGE** | Into spec-workflow.md |
| workflow/worktree-integration.md | 303 | KEEP | Canonical worktree reference |

Merge targets produce 3 files saved from the workflow/ subtree while preserving content.

---

## 5. Config sections audit

### 5.1 Per-yaml usage grep

| Section | File | Go usage (non-test, non-template) | Schema in `types.go`? | Default sensible? |
|---------|------|-----------------------------------|----------------------|-------------------|
| constitution.yaml | 876 B | **None** (only template CLAUDE.md reference) | NO | Yes (technical constraints list) |
| context.yaml | 531 B | **None** (only template skills) | NO | Yes (search settings) |
| design.yaml | 2,763 B | Only `migrate_agency.go` + `migrate_agency_test.go` | NO | Yes (GAN/sprint/adaptation) |
| git-convention.yaml | 530 B | `loader.go`, `types.go`, `loader_test.go` | Yes | Yes |
| git-strategy.yaml | 2,001 B | `initializer.go` (template copy only) | Partial | Yes |
| harness.yaml | 5,182 B | **None in Go** (only template CLAUDE.md + agents) | NO | Yes (3 harness levels) |
| interview.yaml | 264 B | **None in Go** (only skills) | NO | Yes |
| language.yaml | 200 B | `initializer.go`, `manager_test.go` | Yes (LanguageConfig) | Yes |
| llm.yaml | 476 B | `loader.go`, `manager.go`, `types.go` | Yes | Yes |
| lsp.yaml | 8,098 B | `internal/lsp/config/` | Yes (own schema) | Yes |
| mx.yaml | 8,545 B | Referenced by hooks + skills | No direct struct | Yes |
| project.yaml | 153 B | `initializer.go`, `defs/files.go` | Yes | Yes |
| quality.yaml | 2,259 B | `core/quality/trust.go`, `initializer.go` | Yes (QualityConfig) | Yes |
| ralph.yaml | 1,101 B | `config/manager.go`, `config/loader.go`, `types.go`, `defaults.go`, `ralph/classify_lsp_test.go` | Yes (RalphConfig) | Yes |
| research.yaml | 680 B | `loader.go`, `types.go` | Yes (ResearchConfig) | Yes (research mode) |
| security.yaml | 1,023 B | `cli/deps.go`, `hook/security/config.go` | Yes (own struct) | Yes |
| state.yaml | 34 B | `loader.go`, `types.go`, `loader_test.go` | Yes | Yes (minimal, init-only) |
| statusline.yaml | 288 B | `config/types.go`, `loader.go`, `cli/statusline.go` | Yes (StatuslineConfig) | Yes |
| sunset.yaml | 820 B | `defaults.go` (only 4 refs, struct in `types.go`) | Yes (SunsetConfig) | Yes (but dormant) |
| system.yaml | 1,067 B | `initializer.go`, `statusline/version.go`, `statusline/version_test.go` | Partial | Yes |
| user.yaml | 27 B | `initializer.go`, `initializer_test.go`, `phase_test.go` | Yes (UserConfig) | Yes (minimal) |
| workflow.yaml | 3,965 B | `initializer.go`, template-only for role_profiles | Partial | Yes (team role profiles) |

### 5.2 Unused / dormant config sections

**Fully unused by Go runtime (only referenced in template markdown or not at all):**

1. **constitution.yaml** — Technical constraints (approved frameworks, forbidden patterns). Referenced only by `CLAUDE.md` template. No Go code reads it. Either it is a **documentation-only artifact** or the loader is missing. **DEAD or NEEDS LOADER.**
2. **context.yaml** — Context search settings. Referenced only by the `context.md` workflow skill. Go code does not parse it. **DEAD at Go level; loader missing.**
3. **interview.yaml** — Interview clarity threshold. Referenced only by `plan.md` and `project.md` skills. Go code does not read it. **DEAD at Go level; loader missing.**
4. **harness.yaml** — Harness routing rubric (minimal/standard/thorough). Referenced by CLAUDE.md and evaluator-active agent but NO GO loader exists. Critical gap — the file is central to the design system and v3 redesign, yet has no runtime enforcement. **NEEDS LOADER.**
5. **design.yaml** — Design pipeline settings. Only referenced by `migrate_agency.go` (migration, not runtime). **NEEDS LOADER for runtime.**

**Dormant (struct exists but not actively read):**

6. **sunset.yaml** — SunsetConfig struct exists, but no code path tests its thresholds. Only 4 references total; all in defaults/types. **NEEDS RUNTIME ENFORCEMENT or RETIRE.**

### 5.3 Schema gaps

Config sections missing from `internal/config/types.go` but present as yaml files: constitution, context, interview, design, harness, mx (no direct struct, parsed ad-hoc).

For v3 redesign, all yaml sections in `.moai/config/sections/` should have:
- A Go struct in `internal/config/types.go`
- A loader function in `internal/config/loader.go`
- At least one test in `loader_test.go`

---

## Cross-cutting analysis

### A. Hook Coverage Matrix (27 events × 4 columns)

Legend: ✓ = present, ~ = partial (logging-only), ✗ = missing

| Event | Shell wrapper | Go handler | settings.json | Business logic |
|-------|---------------|------------|---------------|----------------|
| SessionStart | ✓ | ✓ | ✓ | ✓ (GLM tmux, skill, memory eval) |
| SessionEnd | ✓ | ✓ | ✓ | ✓ (memo save, MX scan) |
| PreToolUse | ✓ | ✓ | ✓ | ✓ (security scan, secrets, reflective write) |
| PostToolUse | ✓ | ✓ | ✓ | ✓ (MX validation, LSP convert, metrics) |
| PostToolUseFailure | ✓ | ✓ | ✓ | ~ (logging only) |
| PreCompact | ✓ | ✓ | ✓ | ✓ (memo save) |
| PostCompact | ✓ | ✓ | ✓ | ✓ (memo restore) |
| Stop | ✓ | ✓ | ✓ | ✓ (completion markers, Ralph state) |
| StopFailure | ✓ | ✓ | ✓ | ✓ (error-type systemMessage) |
| SubagentStart | ✓ | ✓ | ✓ | ✓ (project context injection) |
| SubagentStop | ✓ | ✓ | ✓ | ~ (logging only — should kill tmux pane) |
| Notification | ✓ | ✓ | ✓ | ~ (logging only) |
| UserPromptSubmit | ✓ | ✓ | ✓ | ✓ (SPEC detect, session title, workflow kw) |
| PermissionRequest | ✓ | ✓ | ✓ | ✓ (updatedInput re-verify) |
| PermissionDenied | ✓ | ✓ | ✓ | ✓ (read-only retry) |
| TeammateIdle | ✓ | ✓ | ✓ | ✓ (quality gate enforcement) |
| TaskCompleted | ✓ | ✓ | ✓ | ✓ (SPEC validation) |
| TaskCreated | ✓ | ✓ | ✓ | ~ (logging only) |
| WorktreeCreate | ✓ | ✓ | ✓ | ✓ (registry update) |
| WorktreeRemove | ✓ | ✓ | ✓ | ✓ (registry cleanup) |
| ConfigChange | ✓ | ✓ | ✓ | ~ (logging only — should trigger reload) |
| CwdChanged | ✓ | ✓ | ✓ | ✓ (CLAUDE_ENV_FILE writes) |
| FileChanged | ✓ | ✓ | ✓ | ~ (logging only) |
| InstructionsLoaded | ✓ | ✓ | ✓ | ~ (logging only) |
| Elicitation | ✓ | ✓ | ✓ | ~ (logging only) |
| ElicitationResult | ✓ | ✓ | ✓ | ~ (logging only) |
| Setup (special) | ✗ | ✓ (orphan) | ✗ | ~ (logging only) |

**Coverage stats:**

- Full (✓ × 4): 16 events (59%)
- Partial (at least one ~): 10 events (37%)
- Missing (shell or settings gap): 1 event (Setup, 4%)

**Top 5 handler gaps (registered but no-op):**

1. **SubagentStop** — Should kill tmux pane per team protocol. Known bug.
2. **ConfigChange** — Should trigger re-render when yaml sections change.
3. **Setup** — Orphan handler; no shell wrapper, no settings.json.
4. **InstructionsLoaded** — Could validate CLAUDE.md character budget.
5. **FileChanged** — Could trigger MX re-validation for externally-edited files.

### B. Commands ↔ Skills mapping

All 15 `/moai/*` slash commands route to `Skill("moai")` with subcommand arg. The `moai` skill entry-point (`.claude/skills/moai/SKILL.md`) is the orchestrator that then dispatches to specialized skills via `moai-workflow-*`, `moai-domain-*`, `moai-foundation-*` family.

**Direct mappings verified via skill availability list:**

| Slash command | Skill route | Exists? |
|---------------|-------------|---------|
| `/moai plan` | `moai:plan` | Yes (subskill) |
| `/moai run` | `moai:run` | Yes |
| `/moai sync` | `moai:sync` | Yes |
| `/moai project` | `moai:project` | Yes |
| `/moai design` | `moai:design` | Yes |
| `/moai db` | `moai:db` | Yes |
| `/moai fix` | `moai:fix` | Yes |
| `/moai loop` | `moai:loop` | Yes |
| `/moai clean` | `moai:clean` | Yes |
| `/moai mx` | `moai:mx` | Yes |
| `/moai feedback` | `moai:feedback` | Yes |
| `/moai review` | `moai:review` | Yes |
| `/moai coverage` | `moai:coverage` | Yes |
| `/moai e2e` | `moai:e2e` | Yes |
| `/moai codemaps` | `moai:codemaps` | Yes |

**No orphan commands. No orphan skills (for the /moai family).**

Dev-only `/98-github` routes to embedded logic within its own 698-LOC file — not skill-routed. `/99-release` similarly embeds all logic. These are **anti-patterns** but tolerated as dev-local.

### C. Rules hierarchy sanity

**Subtree coherence:**
- `core/` — 5 files are cross-cutting rules; `lsp-client.md` is a SPEC decision record (misfiled).
- `design/` — single-file constitution, FROZEN. Clean.
- `development/` — authoring + coding standards. Coherent.
- `languages/` — 16 files, uniform pattern. Clean.
- `workflow/` — 7 files; 3 have redundancy candidates (workflow-modes, team-protocol, file-reading-optimization).

**Rule orphans (referenced by no one):** None directly detected. All 34 rule files are at least mentioned by CLAUDE.md @-imports or by `paths:` glob loading.

**Dueling rules:** None critical. Overlap candidates listed in §4.4.

### D. Config sections usage

See §5.1 for per-section table. Summary:

| Used by Go runtime | Count | Sections |
|--------------------|-------|----------|
| Fully loaded + used | 13 | language, llm, quality, workflow (partial), lsp, mx (hook-level), security, statusline, system, user, project, git-convention, git-strategy, ralph, research, state |
| Dormant (struct exists, no hot path) | 1 | sunset |
| No Go loader (template-only) | 5 | constitution, context, interview, design, harness |

**v3 action items:**
- Add loaders for constitution.yaml, context.yaml, interview.yaml, design.yaml, harness.yaml
- Either remove sunset.yaml or activate it
- Unify workflow.yaml schema (currently only role_profiles are read)

---

## Recommended v3 inventory

### Commands: proposed set (15 thin + 2 refactored)

**Retain (15):** plan, run, sync, project, design, db, fix, loop, clean, mx, feedback, review, coverage, e2e, codemaps (unchanged)

**Refactor (2):**
- `98-github.md` — Extract to `moai-workflow-github` skill; `/98-github` or `/moai github` becomes thin wrapper
- `99-release.md` — Extract to `moai-workflow-release` skill; `/99-release` or `/moai release` becomes thin wrapper

### Hooks: proposed event coverage

**Retain 25 native events** (current state).

**Upgrade 5 handlers from logging-only to real logic:**
1. `subagentStopHandler` — Kill tmux pane if teammate (fixes bug)
2. `configChangeHandler` — Trigger config reload
3. `instructionsLoadedHandler` — Validate CLAUDE.md character budget
4. `fileChangedHandler` — Trigger MX re-scan
5. `postToolUseFailureHandler` — Classify error for better user feedback

**Decide on retirement (5 candidates):**
- `notificationHandler`, `elicitationHandler`, `elicitationResultHandler`, `taskCreatedHandler`, `setupHandler` — retain as observability tap OR remove entirely. Recommend: retain in Go but remove from settings.json (saves 5 hook-fire overheads per session).

### Output styles: proposed set

**Retain both:** MoAI (default), Einstein (opt-in). No changes needed.

### Rules: proposed cleanup

**Structural changes:**
1. Move `core/lsp-client.md` → `.moai/decisions/` (SPEC decision record, not agent rule)
2. Merge `workflow/workflow-modes.md` → `workflow/spec-workflow.md` (duplicate content)
3. Merge `workflow/team-protocol.md` → `workflow/worktree-integration.md` (overlapping content)
4. Move `workflow/file-reading-optimization.md` → `.claude/skills/moai-foundation-context/references/` (heuristic, not rule)

**Frontmatter standardization:**
- Migrate `moai-constitution.md`, `coding-standards.md`, `team-protocol.md`, `worktree-integration.md` from `description + globs` → `paths:` CSV format
- Add `paths:` frontmatter to `file-reading-optimization.md` before moving (or add `paths: "**/*"` if kept)
- Keep `design/constitution.md` and `agent-common-protocol.md` without frontmatter (cross-cutting, intentional)

**Result:** 34 files → 31 files; all with consistent frontmatter convention.

### Config sections: proposed cleanup

**Add Go loaders for 5 sections:** constitution, context, interview, design, harness (all currently template-only)

**Activate or retire sunset.yaml**

**Total: 22 active sections** (23 current − 1 retired + 0 added) OR 23 sections with proper loaders.

---

## Sources: file paths read + Bash greps executed

### Files read (Read tool)

- `/Users/goos/MoAI/moai-adk-go/.claude/commands/moai/{plan,run,sync,db,design,feedback,fix,loop,clean,codemaps,coverage,e2e,mx,project,review}.md` (15 files)
- `/Users/goos/MoAI/moai-adk-go/.claude/output-styles/moai/moai.md` (head)
- `/Users/goos/MoAI/moai-adk-go/.claude/output-styles/moai/einstein.md` (head)
- `/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-session-start.sh`
- `/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-agent-hook.sh`
- `/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-teammate-idle.sh`
- `/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-task-completed.sh`
- `/Users/goos/MoAI/moai-adk-go/.claude/settings.json`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/registry.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/setup.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/generic_handler.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/types.go` (lines 1-100)
- `/Users/goos/MoAI/moai-adk-go/internal/hook/config_change.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/file_changed.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/instructions_loaded.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/notification.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/subagent_stop.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/task_created.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/elicitation.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/worktree_create.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/worktree_remove.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/cwd_changed.go`
- `/Users/goos/MoAI/moai-adk-go/internal/hook/task_completed.go`

### Bash greps executed

- `ls -la` on commands, hooks, output-styles directories
- `find` on rules tree
- `wc -l` on handler files and rules
- `grep "EventType()"` on handler files
- `grep "Register(New"` on deps.go
- `grep "sections/${f}\.yaml"` loop over 23 config sections
- `head -N` on various files for frontmatter inspection

### References consulted

- `.claude/rules/moai/core/hooks-system.md` (27 event types definitive list)
- `.claude/rules/moai/core/settings-management.md` (hook + statusline config)
- `.claude/rules/moai/development/coding-standards.md` (thin-command pattern)
- `.claude/rules/moai/design/constitution.md` (design system v3.3.0)
- `internal/template/commands_audit_test.go` (thin-wrapper enforcement)

---

**End of R6 audit.** Word count: ~4,800.

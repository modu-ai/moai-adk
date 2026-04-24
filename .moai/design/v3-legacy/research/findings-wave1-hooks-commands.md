# Wave 1.1 — Claude Code Hooks & Commands Inventory

**Research scope**: Claude Code TypeScript source dump at `/Users/goos/MoAI/AgentOS/claude-code-source-map/`
**Wave owner**: Wave 1.1 researcher
**Purpose**: Catalog Claude Code's authoritative hook + slash-command subsystems so moai-adk-go v3 planners can compute parity gaps.

---

## 0. Scope Note and Directory Orientation

Important clarification: the top-level `hooks/` directory in the source dump contains **React UI hooks** (e.g., `useCanUseTool.tsx`, `useReplBridge.tsx`, `useTypeahead.tsx`). These are NOT the settings-configurable lifecycle hooks moai-adk cares about.

The **authoritative lifecycle hook system** lives in:
- `entrypoints/sdk/coreTypes.ts` — `HOOK_EVENTS` constant (27 events)
- `entrypoints/sdk/coreSchemas.ts` — Zod schemas for every hook input/output
- `schemas/hooks.ts` — Settings-level hook command schemas (4 hook types)
- `types/hooks.ts` — TypeScript types + JSON output schema
- `utils/hooks.ts` — 5022-line runner (spawn, dedup, match, dispatch per event)
- `utils/hooks/hooksConfigManager.ts` — metadata per event
- `utils/hooks/hooksConfigSnapshot.ts` — snapshot + policy gating
- `utils/hooks/hooksSettings.ts` — load from user/project/local settings
- `utils/hooks/hookEvents.ts` — runtime event emission to SDK consumers
- `utils/hooks/sessionHooks.ts` — ephemeral session-scoped hooks
- `utils/hooks/registerFrontmatterHooks.ts` — agent/skill frontmatter hooks
- `utils/hooks/registerSkillHooks.ts` — skill-invoked hooks
- `utils/hooks/execPromptHook.ts` — prompt-type hook runner (LLM call)
- `utils/hooks/execAgentHook.ts` — agent-type hook runner (multi-turn)
- `utils/hooks/execHttpHook.ts` — http-type hook runner (axios + SSRF guard)
- `utils/hooks/AsyncHookRegistry.ts` — async/backgrounded hook tracking
- `utils/hooks/ssrfGuard.ts` — HTTP hook SSRF protection
- `utils/plugins/loadPluginHooks.ts` — plugin hook registration

The **slash-command system** lives under `commands/` (100+ sub-directories + top-level `.ts` files) plus:
- `commands.ts` — top-level registration, getCommands(), lazy load, filters
- `types/command.ts` — Command type discriminated union
- `skills/loadSkillsDir.ts` — load user/project `.claude/skills/`, `.claude/commands/`
- `skills/bundledSkills.ts` — bundled skills (init, security-review, etc.)
- `utils/plugins/loadPluginCommands.ts` — plugin command loader
- `utils/argumentSubstitution.ts` — `$ARGUMENTS` / `$0` / named args
- `utils/frontmatterParser.ts` — YAML frontmatter + paths split + brace expansion
- `utils/markdownConfigLoader.ts` — markdown file discovery, tools parsing
- `utils/promptShellExecution.ts` — `!` `command` ``` and ```!``` block execution inside skill prompts

---

## 1. Hook Events — Complete Catalog (27 events)

Source: `entrypoints/sdk/coreTypes.ts:25-53`, `entrypoints/sdk/coreSchemas.ts:355-383`.

All 27 events (verbatim order, `as const`):

```
PreToolUse
PostToolUse
PostToolUseFailure
Notification
UserPromptSubmit
SessionStart
SessionEnd
Stop
StopFailure
SubagentStart
SubagentStop
PreCompact
PostCompact
PermissionRequest
PermissionDenied
Setup
TeammateIdle
TaskCreated
TaskCompleted
Elicitation
ElicitationResult
ConfigChange
WorktreeCreate
WorktreeRemove
InstructionsLoaded
CwdChanged
FileChanged
```

### 1.1 BaseHookInput (applies to all 27)

Source: `entrypoints/sdk/coreSchemas.ts:387-411`.

```
session_id         string   (required) — getSessionId()
transcript_path    string   (required) — getTranscriptPathForSession(sessionId)
cwd                string   (required) — current cwd via getCwd()
permission_mode    string?  (optional) — current permission mode
agent_id           string?  (optional) — subagent ID when fired from a subagent
agent_type         string?  (optional) — agent type (general-purpose, code-reviewer...)
```

Enrichment source: `utils/hooks.ts:301-328` `createBaseHookInput()`.

### 1.2 Per-event schema detail

Each entry gives: (a) payload fields added on top of BaseHookInput, (b) matcher field used by `getMatchingHooks` (`utils/hooks.ts:1616-1669`), (c) exit-code semantics from `utils/hooks/hooksConfigManager.ts:28-264`, (d) hookSpecificOutput reply shape from `entrypoints/sdk/coreSchemas.ts:806-935`.

---

#### PreToolUse
- **Payload** (`coreSchemas.ts:414-423`): `tool_name:string`, `tool_input:unknown`, `tool_use_id:string`
- **Matcher**: `tool_name` (supports `*`, `|`-separated list, and regex — `utils/hooks.ts:1346-1381`)
- **Exit codes** (`hooksConfigManager.ts:31-37`):
  - `0` — stdout/stderr not shown
  - `2` — show stderr to model and block tool call
  - Other — stderr to user only, continue
- **hookSpecificOutput reply**: `{hookEventName:'PreToolUse', permissionDecision?: 'allow'|'deny'|'ask', permissionDecisionReason?: string, updatedInput?: object, additionalContext?: string}`
- **Trigger file**: `utils/hooks.ts:3394` `executePreToolHooks`
- **Uses `if` condition**: YES — filtered via `prepareIfConditionMatcher` using permission-rule syntax e.g. `Bash(git *)` (`utils/hooks.ts:1390-1421`)

#### PostToolUse
- **Payload** (`coreSchemas.ts:436-446`): `tool_name`, `tool_input`, `tool_response`, `tool_use_id`
- **Matcher**: `tool_name`
- **Exit codes** (`hooksConfigManager.ts:40-46`): `0`/stdout shown in transcript mode (ctrl+o); `2`/show stderr to model immediately; other/to user only
- **Reply**: `{hookEventName:'PostToolUse', additionalContext?, updatedMCPToolOutput?: unknown}` — the updatedMCPToolOutput field is unique to PostToolUse and lets hooks rewrite MCP tool outputs
- **Trigger**: `utils/hooks.ts:3450` `executePostToolHooks`

#### PostToolUseFailure
- **Payload** (`coreSchemas.ts:448-459`): `tool_name`, `tool_input`, `tool_use_id`, `error:string`, `is_interrupt?:boolean`
- **Matcher**: `tool_name`
- **Reply**: `{hookEventName:'PostToolUseFailure', additionalContext?}`
- **Trigger**: `utils/hooks.ts:3492`
- **Not in moai-adk-go** — moai-adk currently ships PostToolUse but does not distinguish PostToolUseFailure

#### Notification
- **Payload** (`coreSchemas.ts:473-482`): `message:string`, `title?:string`, `notification_type:string`
- **Matcher**: `notification_type` — enumerated values from `hooksConfigManager.ts:71-78`: `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog`, `elicitation_complete`, `elicitation_response`
- **Exit codes**: `0` (stdout/stderr not shown); other (stderr to user only)
- **Trigger**: `utils/hooks.ts:3570`

#### UserPromptSubmit
- **Payload** (`coreSchemas.ts:484-491`): `prompt:string`
- **Matcher**: none (fires on every prompt)
- **Exit codes** (`hooksConfigManager.ts:83-85`): `0` stdout shown to Claude; `2` block processing, erase original prompt; other/to user only
- **Reply**: `{hookEventName:'UserPromptSubmit', additionalContext?}` — additionalContext is injected into the model turn
- **Trigger**: `utils/hooks.ts:3826`

#### SessionStart
- **Payload** (`coreSchemas.ts:493-502`): `source: 'startup'|'resume'|'clear'|'compact'`, `agent_type?`, `model?`
- **Matcher**: `source`
- **Exit codes** (`hooksConfigManager.ts:87-89`): `0` stdout shown to Claude; blocking errors ignored; other/to user only
- **Reply**: `{hookEventName:'SessionStart', additionalContext?, initialUserMessage?, watchPaths?: string[]}` — **watchPaths** registers absolute paths with the FileChanged watcher (unique feature)
- **Trigger**: `utils/hooks.ts:3867`
- **ALWAYS_EMITTED** — fires regardless of `includeHookEvents` SDK flag (`utils/hooks/hookEvents.ts:18`)

#### SessionEnd
- **Payload** (`coreSchemas.ts:758-765`): `reason: 'clear'|'resume'|'logout'|'prompt_input_exit'|'other'|'bypass_permissions_disabled'` (EXIT_REASONS, `coreSchemas.ts:747-754`)
- **Matcher**: `reason`
- **Exit codes**: `0` normal; other/to user only
- **Special timeout** (`utils/hooks.ts:175-182`): **1500 ms default** (tight bound via `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` env override)
- **Trigger**: `utils/hooks.ts:4097`

#### Stop
- **Payload** (`coreSchemas.ts:513-527`): `stop_hook_active:boolean`, `last_assistant_message?:string`
- **Matcher**: none
- **Exit codes** (`hooksConfigManager.ts:96-99`): `0` not shown; `2` show stderr to model and continue; other/to user only
- **Trigger**: `utils/hooks.ts:3639`

#### StopFailure
- **Payload** (`coreSchemas.ts:529-538`): `error: SDKAssistantMessageError`, `error_details?`, `last_assistant_message?`
- **Matcher**: `error` — values enumerated in `hooksConfigManager.ts:104-115`: `rate_limit`, `authentication_failed`, `billing_error`, `invalid_request`, `server_error`, `max_output_tokens`, `unknown`
- **Semantics**: Fire-and-forget — output and exit codes ignored (`hooksConfigManager.ts:103`)
- **Trigger**: `utils/hooks.ts:3594`
- **Not in moai-adk-go**

#### SubagentStart
- **Payload** (`coreSchemas.ts:540-548`): `agent_id:string`, `agent_type:string`
- **Matcher**: `agent_type`
- **Trigger**: `utils/hooks.ts:3932`
- **Not in moai-adk-go** — moai has SubagentStop but not SubagentStart

#### SubagentStop
- **Payload** (`coreSchemas.ts:550-567`): `stop_hook_active:boolean`, `agent_id`, `agent_transcript_path`, `agent_type`, `last_assistant_message?`
- **Matcher**: `agent_type`
- **Exit codes**: `0` not shown; `2` show stderr to subagent, continue its run; other/to user only
- **Trigger**: `utils/hooks.ts:3639` (shared with Stop when agentId provided)

#### PreCompact
- **Payload** (`coreSchemas.ts:569-577`): `trigger: 'manual'|'auto'`, `custom_instructions:string|null`
- **Matcher**: `trigger`
- **Exit codes** (`hooksConfigManager.ts:137-140`): `0` stdout appended as custom compact instructions; `2` block compaction; other/continue
- **Trigger**: `utils/hooks.ts:3961` (returns `{newCustomInstructions?, userDisplayMessage?}`)

#### PostCompact
- **Payload** (`coreSchemas.ts:579-589`): `trigger`, `compact_summary:string` (the summary produced by compaction)
- **Matcher**: `trigger`
- **Trigger**: `utils/hooks.ts:4034`
- **Not in moai-adk-go** — moai has PreCompact but not PostCompact

#### PermissionRequest
- **Payload** (`coreSchemas.ts:425-434`): `tool_name`, `tool_input`, `permission_suggestions?: PermissionUpdate[]`
- **Matcher**: `tool_name`
- **Reply**: `{hookEventName:'PermissionRequest', decision: {behavior:'allow', updatedInput?, updatedPermissions?} | {behavior:'deny', message?, interrupt?}}`
- **Trigger**: `utils/hooks.ts:4157`
- **Not in moai-adk-go** — a powerful primitive allowing programmatic permission decisions before a dialog is shown

#### PermissionDenied
- **Payload** (`coreSchemas.ts:461-471`): `tool_name`, `tool_input`, `tool_use_id`, `reason:string`
- **Matcher**: `tool_name`
- **Reply**: `{hookEventName:'PermissionDenied', retry?:boolean}` — allows the model to be told it may retry
- **Trigger**: `utils/hooks.ts:3529`
- **Not in moai-adk-go**

#### Setup
- **Payload** (`coreSchemas.ts:504-511`): `trigger: 'init'|'maintenance'`
- **Matcher**: `trigger`
- **ALWAYS_EMITTED** alongside SessionStart (`utils/hooks/hookEvents.ts:18`)
- **Trigger**: `utils/hooks.ts:3902`
- **Not in moai-adk-go** — this is a per-repo init/maintenance hook

#### TeammateIdle
- **Payload** (`coreSchemas.ts:591-599`): `teammate_name:string`, `team_name:string`
- **Matcher**: none
- **Exit codes** (`hooksConfigManager.ts:183-184`): `0` not shown; `2` show stderr to teammate AND prevent idle (teammate continues working); other/to user only
- **Trigger**: `utils/hooks.ts:3709`

#### TaskCreated / TaskCompleted
- **Payload** (`coreSchemas.ts:601-625`): `task_id`, `task_subject`, `task_description?`, `teammate_name?`, `team_name?`
- **Matcher**: none
- **Exit codes** (`hooksConfigManager.ts:189, 194`): `0` not shown; `2` show stderr to model and prevent task creation/completion; other/to user only
- **Trigger**: `utils/hooks.ts:3745` / `3789`

#### Elicitation
- **Payload** (`coreSchemas.ts:627-643`): `mcp_server_name:string`, `message:string`, `mode?:'form'|'url'`, `url?`, `elicitation_id?`, `requested_schema?`
- **Matcher**: `mcp_server_name`
- **Reply**: `{hookEventName:'Elicitation', action?: 'accept'|'decline'|'cancel', content?: object}` — hooks can auto-respond to MCP elicitations
- **Trigger**: `utils/hooks.ts:4470`
- **Not in moai-adk-go**

#### ElicitationResult
- **Payload** (`coreSchemas.ts:645-660`): `mcp_server_name`, `elicitation_id?`, `mode?`, `action: 'accept'|'decline'|'cancel'`, `content?`
- **Matcher**: `mcp_server_name`
- **Reply**: same shape; hook can override user response before it's sent to the server
- **Trigger**: `utils/hooks.ts:4525`
- **Not in moai-adk-go**

#### ConfigChange
- **Payload** (`coreSchemas.ts:670-678`): `source: 'user_settings'|'project_settings'|'local_settings'|'policy_settings'|'skills'`, `file_path?`
- **Matcher**: `source`
- **Exit codes** (`hooksConfigManager.ts:216-217`): `0` allow the change; `2` block the change from being applied to the session; other/to user only
- **Trigger**: `utils/hooks.ts:4214`
- **Not in moai-adk-go** — would let moai react to settings file edits mid-session

#### InstructionsLoaded
- **Payload** (`coreSchemas.ts:695-706`): `file_path:string`, `memory_type: 'User'|'Project'|'Local'|'Managed'`, `load_reason: 'session_start'|'nested_traversal'|'path_glob_match'|'include'|'compact'`, `globs?:string[]`, `trigger_file_path?`, `parent_file_path?`
- **Matcher**: `load_reason`
- **Semantics** (`hooksConfigManager.ts:231-232`): Observability-only — **does not support blocking**
- **Trigger**: `utils/hooks.ts:4335`
- **Not in moai-adk-go** — excellent for auditing which CLAUDE.md rules got loaded

#### WorktreeCreate
- **Payload** (`coreSchemas.ts:709-716`): `name:string` (suggested slug)
- **Matcher**: none
- **Semantics** (`hooksConfigManager.ts:245-247`): stdout MUST contain absolute path to created worktree; non-zero exit = creation failed. This is a PROVIDER hook — Claude Code calls it to create a worktree and uses its stdout.
- **Reply**: `{hookEventName:'WorktreeCreate', worktreePath:string}`
- **Trigger**: `utils/hooks.ts:4928`
- **moai-adk has WorktreeCreate**, but it's currently observational — not implementing the "provider" contract (stdout = absolute path)

#### WorktreeRemove
- **Payload** (`coreSchemas.ts:718-725`): `worktree_path:string`
- **Matcher**: none
- **Trigger**: `utils/hooks.ts:4967`

#### CwdChanged
- **Payload** (`coreSchemas.ts:727-735`): `old_cwd:string`, `new_cwd:string`
- **Matcher**: none
- **Semantics** (`hooksConfigManager.ts:255-257`): CLAUDE_ENV_FILE is set — write bash exports there to apply env to subsequent BashTool commands. Reply can include `watchPaths` to register with FileChanged watcher.
- **Reply**: `{hookEventName:'CwdChanged', watchPaths?:string[]}`
- **Not in moai-adk-go**

#### FileChanged
- **Payload** (`coreSchemas.ts:737-745`): `file_path:string`, `event: 'change'|'add'|'unlink'`
- **Matcher**: `basename(file_path)` — e.g. matcher `.envrc|.env` matches by base filename
- **Semantics** (`hooksConfigManager.ts:260-262`): `CLAUDE_ENV_FILE` is set; hook can write bash exports. Reply can include `watchPaths` to dynamically update the watch list.
- **Watcher implementation**: `utils/hooks/fileChangedWatcher.ts` (5309 bytes)
- **Not in moai-adk-go**

---

## 2. Hook Types (Settings-level — 4 user-configurable + 2 internal)

Source: `schemas/hooks.ts`.

### 2.1 `type: 'command'` — Bash/PowerShell hook (most common)

`schemas/hooks.ts:32-65`:

```yaml
type: command
command: string           # required — shell command
if: string?               # permission-rule syntax: "Bash(git *)", "Read(*.ts)"
shell: 'bash'|'powershell'?  # default 'bash'
timeout: number?          # positive, seconds
statusMessage: string?    # shown in spinner
once: boolean?            # remove after first execution
async: boolean?           # run backgrounded
asyncRewake: boolean?     # backgrounded + wakes model on exit 2 (implies async)
```

Execution: `utils/hooks.ts:747-1335` `execCommandHook`. Spawns child process with stdin=JSON input, reads stdout/stderr.

### 2.2 `type: 'prompt'` — LLM prompt hook

`schemas/hooks.ts:67-95`:

```yaml
type: prompt
prompt: string            # required — LLM prompt, uses $ARGUMENTS for hook input
if: string?
timeout: number?
model: string?            # e.g. "claude-sonnet-4-6"; default small-fast model
statusMessage: string?
once: boolean?
```

Execution: `utils/hooks/execPromptHook.ts:21-80`. Runs a one-shot LLM query with system-prompt requiring JSON response `{ok:boolean, reason?:string}`. Haiku-class by default.

### 2.3 `type: 'agent'` — Agentic verifier hook

`schemas/hooks.ts:128-163`:

```yaml
type: agent
prompt: string            # "Verify that unit tests ran and passed."
if: string?
timeout: number?          # default 60 seconds
model: string?            # default Haiku
statusMessage: string?
once: boolean?
```

Execution: `utils/hooks/execAgentHook.ts`. Multi-turn query via `query()` — actual subagent with tools, NOT just a one-shot prompt. Respects `ALL_AGENT_DISALLOWED_TOOLS`.

### 2.4 `type: 'http'` — HTTP webhook

`schemas/hooks.ts:97-126`:

```yaml
type: http
url: string               # required, validated as URL
if: string?
timeout: number?
headers: Record<string,string>?   # values may use $VAR / ${VAR}
allowedEnvVars: string[]?         # explicit allowlist for header interpolation
statusMessage: string?
once: boolean?
```

Execution: `utils/hooks/execHttpHook.ts`. POST hook input JSON; reads JSON response as the hook output. Protected by:
- SSRF guard (`utils/hooks/ssrfGuard.ts`, 8732 bytes) — blocks internal network resolution
- URL allowlist via policy setting `allowedHttpHookUrls` (`execHttpHook.ts:49-58`)
- Env-var interpolation allowlist `httpHookAllowedEnvVars`
- Header value sanitization strips `\r\n\x00` to prevent CRLF injection

### 2.5 `type: 'callback'` — SDK/plugin native JS hook (internal only)

Not user-configurable. Registered via `registerHookCallbacks()` from `bootstrap/state.js`. Used by:
- SDK hosts (via `Options.hooks` callback interface)
- Plugins shipping native JS hooks
- Internal observability (see `utils/sessionFileAccessHooks.ts`, `utils/commitAttribution.ts`)

Shape (`types/hooks.ts:211-226`):
```ts
{
  type: 'callback'
  callback: (input, toolUseID, abort, hookIndex?, context?) => Promise<HookJSONOutput>
  timeout?: number
  internal?: boolean   // excluded from tengu_run_hook metrics
}
```

### 2.6 `type: 'function'` — Session-scoped in-memory hook (internal only)

Used for structured-output enforcement. Never persisted to settings. Cannot be deduplicated (no stable identifier). Source: `utils/hooks/sessionHooks.ts`.

---

## 3. Hook Configuration Schema & Registration

### 3.1 Top-level HooksSettings shape

`schemas/hooks.ts:192-213`:

```yaml
hooks:
  PreToolUse:
    - matcher: "Write|Edit"            # string pattern
      hooks:
        - type: command
          command: "~/format.sh"
          timeout: 5
    - matcher: "Bash"
      hooks:
        - type: command
          command: "~/audit-bash.sh"
          if: "Bash(git *)"            # only runs for git bash commands
  PostToolUse:
    - matcher: "*"
      hooks:
        - type: http
          url: "https://my-webhook.example.com/cc"
          timeout: 3
          allowedEnvVars: ["MY_TOKEN"]
          headers: { Authorization: "Bearer $MY_TOKEN" }
  SessionStart:
    - hooks:
        - type: prompt
          prompt: "Summarize last session: $ARGUMENTS"
          model: "claude-haiku-4-5"
```

### 3.2 Hook sources and precedence

`utils/hooks/hooksSettings.ts:15-228`:

```
Priority order (highest first):
1. policySettings          (managed-settings.json / MDM)
2. userSettings            (~/.claude/settings.json)
3. projectSettings         (.claude/settings.json)
4. localSettings           (.claude/settings.local.json)
5. pluginHook              (plugin hooks.json)
6. skillHook               (skill frontmatter hooks)
7. sessionHook             (in-memory, ephemeral)
8. builtinHook             (registered internally)
```

### 3.3 Hook loading pipeline

`utils/hooks.ts:1492-1566` `getHooksConfig`:

1. Start with `getHooksConfigFromSnapshot()` — captured at startup from allowed sources
2. Append `getRegisteredHooks()` (plugin hooks + SDK callback hooks)
3. Merge `getSessionHooks()` scoped to current session (agent frontmatter + skill frontmatter)
4. Merge `getSessionFunctionHooks()` — non-persistable structured-output hooks

### 3.4 Policy gates (`utils/hooks/hooksConfigSnapshot.ts`)

- `policySettings.disableAllHooks === true` → returns `{}` (no hooks, including managed)
- `policySettings.allowManagedHooksOnly === true` → only `policySettings.hooks`
- `mergedSettings.disableAllHooks === true` (non-managed) → only managed hooks run
- `isRestrictedToPluginOnly('hooks')` → only policy hooks (plus plugin-registered)

### 3.5 Plugin hook registration

`utils/plugins/loadPluginHooks.ts:91-157` `loadPluginHooks` (memoized):
- Reads `plugin.hooksConfig` from each enabled plugin
- Converts to `PluginHookMatcher` with `pluginRoot`, `pluginName`, `pluginId`
- Atomic clear-then-register (`utils/plugins/loadPluginHooks.ts:147-148`)
- Hot-reloads when `policySettings` changes (`setupPluginHookHotReload`)
- Plugin hooks run with `${CLAUDE_PLUGIN_ROOT}` and `${CLAUDE_PLUGIN_DATA}` and `${user_config.X}` substitution available in command strings

### 3.6 Skill/agent frontmatter hook registration

Skills: `utils/hooks/registerSkillHooks.ts:20-64`. When a skill is invoked, its frontmatter `hooks:` block registers session-scoped hooks.

Agents: `utils/hooks/registerFrontmatterHooks.ts:18-67`. Same mechanism, but for agents. Critically, `Stop` hooks in agent frontmatter are rewritten to `SubagentStop` because that's what fires when a subagent completes (`registerFrontmatterHooks.ts:38-45`).

### 3.7 `once: true` semantics

Per-hook `once: true` (defined on command/prompt/agent/http) — the hook is auto-removed from session hooks after its first successful execution (`registerSkillHooks.ts:36-43`).

### 3.8 Hook matcher semantics

`utils/hooks.ts:1346-1381`:

- `undefined` or `*` → matches everything
- Simple string (`[a-zA-Z0-9_|]+`) → exact match with legacy-tool-name normalization
- Pipe-separated list `Write|Edit|Bash` → any exact match
- Any other pattern → compiled as RegExp (tried against current name + all legacy aliases)

The matchQuery is resolved from the hook input payload based on event (`utils/hooks.ts:1616-1669`):

| Event | matchQuery |
|-------|-----------|
| PreToolUse / PostToolUse / PostToolUseFailure / PermissionRequest / PermissionDenied | `tool_name` |
| SessionStart | `source` |
| Setup / PreCompact / PostCompact | `trigger` |
| Notification | `notification_type` |
| SessionEnd | `reason` |
| StopFailure | `error` |
| SubagentStart / SubagentStop | `agent_type` |
| TeammateIdle / TaskCreated / TaskCompleted | none |
| Elicitation / ElicitationResult | `mcp_server_name` |
| ConfigChange | `source` |
| InstructionsLoaded | `load_reason` |
| FileChanged | `basename(file_path)` |
| CwdChanged / Stop / UserPromptSubmit / WorktreeCreate / WorktreeRemove | none |

### 3.9 Hook dedup (`utils/hooks.ts:1712-1801`)

Hooks are deduplicated within the same source context (pluginRoot / skillRoot / settings):
- `command` key: `{shell}\0{command}\0{if}`
- `prompt` key: `{prompt}\0{if}`
- `agent` key: `{prompt}\0{if}`
- `http` key: `{url}\0{if}`
- `callback` / `function`: no dedup (each unique)

`hookDedupKey` uses `pluginRoot ?? skillRoot ?? ''` prefix so cross-plugin template collisions don't drop hooks.

---

## 4. Hook Output Protocol

### 4.1 Wire format (`types/hooks.ts:29-166`, `entrypoints/sdk/coreSchemas.ts:799-935`)

Every hook receives JSON on stdin, writes response to stdout. Two reply modes:

**Sync reply** (must be single JSON object, hook blocks until exit):
```json
{
  "continue": true,             // false = halt Claude
  "suppressOutput": false,       // hide stdout from transcript
  "stopReason": "optional msg shown when continue is false",
  "decision": "approve" | "block",
  "reason": "string",
  "systemMessage": "warning shown to user",
  "hookSpecificOutput": { "hookEventName": "...", ... }
}
```

**Async reply** (`{"async":true, "asyncTimeout":NUMBER}` on FIRST LINE of stdout):
- Hook is backgrounded, registered in `AsyncHookRegistry`
- Default timeout 15000 ms
- `asyncRewake` variant wakes the model on exit code 2 via `task-notification` queue

### 4.2 Exit-code semantics (summary)

The full matrix per event is in `utils/hooks/hooksConfigManager.ts:31-262`, but the general pattern is:
- `0`: success, stdout interpretation varies (shown to Claude / stored / ignored)
- `2`: **blocking** — stderr is injected as a blocking error the model sees
- Any other non-zero: stderr shown to user only, Claude continues

### 4.3 Environment variables injected into hook processes

`utils/hooks.ts:882-926`:

| Env var | When set |
|---------|----------|
| `CLAUDE_PROJECT_DIR` | Always — resolves to stable project root (never worktree) |
| `CLAUDE_PLUGIN_ROOT` | Plugin hooks AND skill hooks (reuses same var name for migration) |
| `CLAUDE_PLUGIN_DATA` | Plugin hooks only — per-plugin data dir |
| `CLAUDE_PLUGIN_OPTION_*` | Plugin hooks — each plugin option as env var |
| `CLAUDE_ENV_FILE` | `SessionStart`, `Setup`, `CwdChanged`, `FileChanged` (bash only) — path to a .sh file where hook writes bash exports for subsequent BashTool commands |

On Windows (bash only): all paths go through `windowsPathToPosixPath()` for Git Bash compatibility.

### 4.4 Command substitution

In the `command` string itself (`utils/hooks.ts:822-857`):
- `${CLAUDE_PLUGIN_ROOT}` → plugin root path (before env subst)
- `${CLAUDE_PLUGIN_DATA}` → plugin data dir
- `${user_config.KEY}` → plugin-option value

### 4.5 Shell selection

`utils/hooks.ts:790-792, 957-984`. Per-hook `shell: 'bash'|'powershell'`. PowerShell uses `pwsh -NoProfile -NonInteractive -Command`. No `bash` prepend, no POSIX path conversion, no `CLAUDE_CODE_SHELL_PREFIX`.

### 4.6 Trust gating

`utils/hooks.ts:286-296` `shouldSkipHookDueToTrust()`: ALL hooks require workspace trust (except in non-interactive/SDK mode where trust is implicit).

### 4.7 SessionEnd bound

`utils/hooks.ts:175-182`: SessionEnd uses **1500 ms** timeout (not `TOOL_HOOK_EXECUTION_TIMEOUT_MS = 10*60*1000`). Override via `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS`.

### 4.8 Parallel execution

Hooks for a single event run in parallel by default (one spawn per matched hook). The runner collects results into `AggregatedHookResult`. Ordering is not preserved — callers must not depend on hook execution order within an event.

### 4.9 Hook progress streaming (SDK)

`utils/hooks/hookEvents.ts:22-191`. Emits `HookStartedEvent`, `HookProgressEvent` (1s interval), `HookResponseEvent` to SDK consumers. Progress gated:
- `SessionStart` + `Setup` always emitted
- Other events require `setAllHookEventsEnabled(true)` (SDK `includeHookEvents` option or `CLAUDE_CODE_REMOTE`)

### 4.10 Prompt elicitation from inside a hook

`types/hooks.ts:28-47`: A hook command can write a `{"prompt":"id","message":"...","options":[...]}` JSON line to stdout. The runner will parse it, call `requestPrompt` (bridging to `AskUserQuestion` UI), and write the user's response back to the hook's stdin as `{"prompt_response":"id","selected":"..."}`. Fully interactive command-line hook conversations are possible.

---

## 5. Slash-Command Catalog (built-in)

Source: enumerated via `commands.ts:258-346` and each `commands/<dir>/index.ts`.

### 5.1 Command type taxonomy (`types/command.ts:205-206`)

```
Command = CommandBase & (PromptCommand | LocalCommand | LocalJSXCommand)
```

- **`prompt`**: Text sent to model (e.g., `/review`, `/security-review`, `/init`, `/insights`, built-in skills). Content injected as a user message.
- **`local`**: Node-executed function returning text/compact/skip — runs outside REPL, no model interaction. Example: `/clear`, `/cost`, `/files`, `/rewind`.
- **`local-jsx`**: Renders Ink UI modal — user-interactive. Example: `/config`, `/model`, `/agents`, `/hooks`, `/plugin`, `/permissions`.

### 5.2 Built-in commands (full list — alphabetical)

From `commands.ts:258-346` + individual index files:

| name | aliases | type | description | argumentHint |
|------|---------|------|-------------|--------------|
| add-dir | — | local-jsx | Add a new working directory | `<path>` |
| advisor | — | local-jsx | Advisor tool | — |
| agents | — | local-jsx | Manage agent configurations | — |
| branch | fork* | local-jsx | Create a branch of the current conversation at this point | `[name]` |
| btw | — | local-jsx | Quick note | `<question>` |
| chrome | — | local-jsx | Claude in Chrome (Beta) settings | — |
| clear | reset, new | local | Clear conversation history and free up context | — |
| color | — | local-jsx | Set the prompt bar color for this session | `<color\|default>` |
| compact | — | local | Clear conversation history but keep a summary in context | `<optional custom summarization instructions>` |
| config | settings | local-jsx | Open config panel | — |
| context | — | local-jsx / local | Visualize / show current context usage | — |
| copy | — | local-jsx | Copy last message | — |
| cost | — | local | Show the total cost and duration of the current session | — |
| desktop | app | local-jsx | Continue the current session in Claude Desktop | — |
| diff | — | local-jsx | View uncommitted changes and per-turn diffs | — |
| doctor | — | local-jsx | Diagnose and verify Claude Code installation | — |
| effort | — | local-jsx | Set effort level for model usage | `[low\|medium\|high\|max\|auto]` |
| exit | quit | local-jsx | Exit the REPL | — |
| export | — | local-jsx | Export the current conversation | `[filename]` |
| extra-usage | — | local-jsx / local | Configure extra usage to keep working when limits are hit | — |
| fast | — | local-jsx | — | `[on\|off]` |
| feedback | bug | local-jsx | Submit feedback about Claude Code | `[report]` |
| files | — | local | List all files currently in context | — |
| heapdump | — | local | Dump the JS heap to ~/Desktop | — |
| help | — | local-jsx | Show help and available commands | — |
| hooks | — | local-jsx | View hook configurations for tool events | — |
| ide | — | local-jsx | Manage IDE integrations and show status | `[open]` |
| init | — | prompt | Initialize CLAUDE.md (see NEW_INIT_PROMPT in `commands/init.ts`) | — |
| insights | — | prompt (lazy) | Generate a report analyzing Claude Code sessions | — |
| install-github-app | — | local-jsx | Set up Claude GitHub Actions for a repository | — |
| install-slack-app | — | local | Install the Claude Slack app | — |
| keybindings | — | local | Open or create keybindings configuration file | — |
| login | — | local-jsx | Sign in | — |
| logout | — | local-jsx | Sign out from Anthropic account | — |
| mcp | — | local-jsx | Manage MCP servers | `[enable\|disable [server-name]]` |
| memory | — | local-jsx | Edit Claude memory files | — |
| mobile | ios, android | local-jsx | Show QR code to download Claude mobile app | — |
| model | — | local-jsx | Set the AI model | `[model]` |
| output-style | — | local-jsx (hidden) | Deprecated: use /config | — |
| passes | — | local-jsx | — | — |
| permissions | allowed-tools | local-jsx | Manage allow & deny tool permission rules | — |
| plan | — | local-jsx | Enable plan mode or view the current session plan | `[open\|<description>]` |
| plugin | plugins, marketplace | local-jsx | Manage Claude Code plugins | — |
| pr-comments | — | prompt | Get comments from a GitHub pull request | — |
| privacy-settings | — | local-jsx | View and update privacy settings | — |
| rate-limit-options | — | local-jsx | Show options when rate limit is reached | — |
| release-notes | — | local | View release notes | — |
| reload-plugins | — | local | Activate pending plugin changes in current session | — |
| remote-env | — | local-jsx | Configure default remote environment for teleport | — |
| rename | — | local-jsx | Rename the current conversation | `[name]` |
| resume | continue | local-jsx | Resume a previous conversation | `[conversation id or search term]` |
| review | — | prompt | Review a pull request (see `commands/review.ts:33`) | — |
| ultrareview | — | local-jsx | ~10–20 min remote bughunter via Claude Code on the web | — |
| rewind | checkpoint | local | Restore the code and/or conversation to a previous point | — |
| sandbox | — | local-jsx | Sandbox controls | `exclude "command pattern"` |
| security-review | — | prompt | Complete a security review of the pending changes (see `commands/security-review.ts`) | — |
| session | remote | local-jsx | Show remote session URL and QR code | — |
| skills | — | local-jsx | List available skills | — |
| stats | — | local-jsx | Show Claude Code usage statistics | — |
| status | — | local-jsx | — | — |
| statusline | — | ? | Status line toggle | — |
| stickers | — | local | Order Claude Code stickers | — |
| tag | — | local-jsx | Toggle a searchable tag on the current session | `<tag-name>` |
| tasks | bashes | local-jsx | List and manage background tasks | — |
| terminal-setup | — | local-jsx | — | — |
| theme | — | local-jsx | Change the theme | — |
| think-back | — | local-jsx | Your 2025 Claude Code Year in Review | — |
| upgrade | — | local-jsx | Upgrade to Max | — |
| usage | — | local-jsx | Show plan usage limits | — |
| vim | — | local | Toggle between Vim and Normal editing modes | — |
| voice | — | local | Toggle voice mode | — |

### 5.3 Internal-only commands (`INTERNAL_ONLY_COMMANDS`)

Visible only when `process.env.USER_TYPE === 'ant'`. `commands.ts:225-254`:
`backfill-sessions`, `break-cache`, `bughunter`, `commit`, `commit-push-pr`, `ctx_viz`, `good-claude`, `issue`, `init-verifiers`, `force-snip`, `mock-limits`, `bridge-kick`, `version`, `ultraplan`, `subscribe-pr`, `reset-limits`, `onboarding`, `share`, `summary`, `teleport`, `ant-trace`, `perf-issue`, `env`, `oauth-refresh`, `debug-tool-call`, `agents-platform`, `autofix-pr`.

### 5.4 Feature-flagged commands

`commands.ts:60-123`:
- `proactive` — `PROACTIVE` or `KAIROS`
- `brief` — `KAIROS` or `KAIROS_BRIEF`
- `assistant` — `KAIROS`
- `bridge` / `remote-control` — `BRIDGE_MODE`
- `voice` — `VOICE_MODE`
- `force-snip` — `HISTORY_SNIP`
- `workflows` — `WORKFLOW_SCRIPTS`
- `web-setup` — `CCR_REMOTE_SETUP`
- `skill-index-cache` — `EXPERIMENTAL_SKILL_SEARCH`
- `subscribe-pr` — `KAIROS_GITHUB_WEBHOOKS`
- `ultraplan` — `ULTRAPLAN`
- `torch` — `TORCH`
- `peers` — `UDS_INBOX`
- `fork` — `FORK_SUBAGENT`
- `buddy` — `BUDDY`

### 5.5 Remote/bridge command filtering

`commands.ts:619-675`:

- `REMOTE_SAFE_COMMANDS` (`:619-637`): commands safe in `--remote` mode (session, exit, clear, help, theme, color, vim, cost, usage, copy, btw, feedback, plan, keybindings, statusline, stickers, mobile)
- `BRIDGE_SAFE_COMMANDS` (`:651-660`): commands safe when triggered from mobile/web bridge (compact, clear, cost, summary, release-notes, files)
- `isBridgeSafeCommand` (`:672-676`): prompt-type always OK; local-jsx always blocked; local commands need explicit BRIDGE_SAFE_COMMANDS opt-in

### 5.6 Command availability gating

`commands.ts:417-443` `meetsAvailabilityRequirement` — separate from `isEnabled()`:
- `availability: ['claude-ai']` → only claude.ai subscribers
- `availability: ['console']` → only direct Console API key users
- No `availability` → everyone

`isEnabled()` is dynamic (called every `getCommands()`) so auth changes and feature flags take effect immediately.

---

## 6. User-defined Commands / Skills

### 6.1 Loading pipeline

`commands.ts:449-517` `loadAllCommands(cwd)` + `getCommands(cwd)`:

1. `getSkills(cwd)` — load from `/skills/` dirs (project+user), plugin skills, bundled skills, built-in plugin skills
2. `getPluginCommands()` — plugin-defined commands
3. `getWorkflowCommands(cwd)` — `WORKFLOW_SCRIPTS` feature
4. Combined list: `[bundled, builtinPlugin, skillDir, workflow, plugin, pluginSkills, COMMANDS]`

### 6.2 Directory locations

`skills/loadSkillsDir.ts:78-94` `getSkillsPath`:

| Source | skills path | commands path |
|--------|-------------|---------------|
| policySettings (MDM) | `<managed>/.claude/skills` | `<managed>/.claude/commands` |
| userSettings | `~/.claude/skills` | `~/.claude/commands` |
| projectSettings | `.claude/skills` (walked upward to home) | `.claude/commands` |
| plugin | `<pluginRoot>` | `<pluginRoot>` |

`utils/markdownConfigLoader.ts:29-36` `CLAUDE_CONFIG_DIRECTORIES`:
```
commands, agents, output-styles, skills, workflows, [templates (feature-flagged)]
```

### 6.3 Skill directory format

`skills/loadSkillsDir.ts:407-480` — ONLY directory-based skills are supported in `/skills/`:
```
.claude/skills/
  my-skill/
    SKILL.md        <- required, holds frontmatter + body
    helper.sh       <- referenced via ${CLAUDE_SKILL_DIR}
```

Single `.md` files in `/skills/` are rejected.

### 6.4 Legacy `/commands/` format

`skills/loadSkillsDir.ts:566-623` `loadSkillsFromCommandsDir`:
- Supports single `.md` files AND `<name>/SKILL.md` directory form
- Namespace from subdir path (e.g., `.claude/commands/ops/deploy.md` → `ops:deploy`)
- `loadedFrom: 'commands_DEPRECATED'` — marked for removal

### 6.5 Deduplication

Post-load dedup via `realpath` canonicalization (`skills/loadSkillsDir.ts:743-763`). Handles symlinks and overlapping parent directories (e.g., nested git repos within a workspace).

### 6.6 Dynamic skill discovery (`skills/loadSkillsDir.ts:861-915`)

When the model touches file paths, Claude Code walks UP from each file path to cwd, looking for nested `.claude/skills/` directories. Skipped if the containing dir is gitignored. Deepest-first ordering — skills closer to the file take precedence.

### 6.7 Conditional (path-filtered) skills

`skills/loadSkillsDir.ts:773-796, 997-1040`. Skills with `paths:` frontmatter are stored in `conditionalSkills` map and only activated when the model touches matching paths (gitignore-style patterns via `ignore` library).

---

## 7. Frontmatter Schema (full field reference)

`utils/frontmatterParser.ts:10-59` + `skills/loadSkillsDir.ts:185-264`:

```yaml
---
name: string                            # optional display name override
description: string                     # short description (falls back to first line of .md)
when_to_use: string                     # detailed usage scenarios (for SkillTool model hints)
version: string                         # semver-ish
model: string                           # 'inherit' | 'haiku' | 'sonnet' | 'opus' | specific ID
allowed-tools: string | string[]        # tool names, comma list or YAML array
argument-hint: string                   # gray hint shown after command name
arguments: string | string[]            # named arg list for $name substitution
effort: 'low'|'medium'|'high'|'max'|'auto' | integer
disable-model-invocation: boolean
user-invocable: boolean                 # default true for /commands/, true for /skills/ 
hide-from-slash-command-tool: string    # boolean-string: hide from SlashCommand tool
skills: string                          # comma list of skill names to preload (agents only)
context: 'inline' | 'fork'              # inline expands into current; fork runs as sub-agent
agent: string                           # sub-agent type when context: fork
paths: string | string[]                # conditional activation globs (e.g. "**/*.py,**/pyproject.toml")
shell: 'bash' | 'powershell'            # shell for !`cmd` and ```! blocks in the .md body
hooks:                                  # Session-scoped hooks registered when skill invoked
  PostToolUse:
    - matcher: "Write|Edit"
      hooks:
        - type: command
          command: "~/format.sh"
type: 'user'|'feedback'|'project'|'reference'   # for memory files only
---
```

`paths` parsing (`utils/frontmatterParser.ts:189-232`): comma-separated with brace expansion, e.g. `src/*.{ts,tsx}` → `[src/*.ts, src/*.tsx]`.

Special YAML preprocessing (`utils/frontmatterParser.ts:85-121`): auto-quotes values with YAML special chars (`{ } [ ] * & # ! | > % @ \`` and `: ` pattern) so glob patterns parse correctly.

---

## 8. Argument Substitution

`utils/argumentSubstitution.ts`:

- `$ARGUMENTS` — full args string
- `$ARGUMENTS[N]` — Nth arg (0-indexed)
- `$N` (shorthand) — same as `$ARGUMENTS[N]`, but only if N is an integer and not followed by a word char
- `$name` (when frontmatter `arguments: "name1 name2 name3"` is set) — maps to position
- Arguments parsed via `tryParseShellCommand` (shell-quote) — preserves quoted strings as one token

`generateProgressiveArgumentHint(argNames, typedArgs)` (`:76-83`) → remaining hint like `[arg2] [arg3]` shown in typeahead as user types.

Behavior when NO placeholders found AND `appendIfNoPlaceholder` is true: content is appended with `\n\nARGUMENTS: {args}` (used by skill prompts to convey unused args to the model).

---

## 9. In-skill Shell Execution

`utils/promptShellExecution.ts`:

Two syntaxes inside skill/command `.md` body content:
- **Block**: <code>```!</code> command <code>```</code>
- **Inline**: `` !`command` `` (requires whitespace or line-start before `!`)

Executed via BashTool (default) or PowerShellTool (when `shell: powershell` in frontmatter AND `isPowerShellToolEnabled()`). Permission checked first — hits normal tool permission flow. Output is substituted back into the prompt.

**MCP skills are blocked from this** (`skills/loadSkillsDir.ts:373-396`) — remote, untrusted; never execute inline shell from their bodies.

Built-in replacements applied to skill body before shell exec:
- `${CLAUDE_SKILL_DIR}` → skill's own directory (normalized to forward slashes on Windows)
- `${CLAUDE_SESSION_ID}` → current session ID

---

## 10. Model-Invocable Skills (SkillTool routing)

`commands.ts:561-608`:

- `getSkillToolCommands(cwd)` — filter to `type: 'prompt'` + not `disableModelInvocation` + source ≠ builtin + (bundled OR skills OR commands_DEPRECATED OR hasUserSpecifiedDescription OR whenToUse)
- `getSlashCommandToolSkills(cwd)` — stricter: also requires either `loadedFrom in (skills, plugin, bundled)` OR `disableModelInvocation: true`

This is the mechanism by which the model chooses to invoke a "skill" — the model sees the skill's `name`, `description`, `whenToUse` and can call it via the SkillTool.

---

## 11. Command-to-Skill routing (how /moai-style commands work)

The pattern Claude Code uses for "thin routing commands":

1. User-defined markdown file at `.claude/commands/<name>.md` with frontmatter `description: "..."` and body that invokes a skill via `Skill("skill-name")` (or equivalent routing logic).
2. This is exactly the pattern moai-adk's `/moai` command uses — a thin 20-LOC markdown file that calls `Skill("moai")`.
3. Claude Code has no native "command alias → skill" primitive; it's a convention.

Relevant: `types/command.ts:25-57` `PromptCommand` fields include:
- `source: SettingSource | 'builtin' | 'mcp' | 'plugin' | 'bundled'`
- `pluginInfo` (plugin manifest + repository)
- `hooks?: HooksSettings` — hooks registered when this skill is invoked
- `skillRoot?: string` — used for CLAUDE_PLUGIN_ROOT env var
- `context?: 'inline' | 'fork'` — inline expands into the current conv; fork runs as a sub-agent
- `agent?: string` — agent type when forked
- `effort?: EffortValue`
- `paths?: string[]` — glob gating

---

## 12. Hook Source References (selected claims)

| Claim | File:Lines |
|-------|-----------|
| 27 hook events enumerated | `entrypoints/sdk/coreTypes.ts:25-53` |
| HOOK_EVENTS also exported from coreSchemas | `entrypoints/sdk/coreSchemas.ts:355-383` |
| BaseHookInput schema | `entrypoints/sdk/coreSchemas.ts:387-411` |
| PreToolUse schema | `entrypoints/sdk/coreSchemas.ts:414-423` |
| PostToolUse schema (updatedMCPToolOutput) | `entrypoints/sdk/coreSchemas.ts:436-446` |
| SessionStart schema + watchPaths reply | `entrypoints/sdk/coreSchemas.ts:493-502, 823-830` |
| SessionEnd reason enum | `entrypoints/sdk/coreSchemas.ts:747-754` |
| TaskCreated / TaskCompleted schemas | `entrypoints/sdk/coreSchemas.ts:601-625` |
| Elicitation hook schemas | `entrypoints/sdk/coreSchemas.ts:627-660` |
| ConfigChange schema + CONFIG_CHANGE_SOURCES | `entrypoints/sdk/coreSchemas.ts:662-678` |
| InstructionsLoaded schema | `entrypoints/sdk/coreSchemas.ts:695-706` |
| WorktreeCreate/Remove schemas | `entrypoints/sdk/coreSchemas.ts:709-725` |
| CwdChanged schema | `entrypoints/sdk/coreSchemas.ts:727-735` |
| FileChanged schema | `entrypoints/sdk/coreSchemas.ts:737-745` |
| PermissionRequest decision shape | `entrypoints/sdk/coreSchemas.ts:875-891` |
| Hook JSON output schema | `types/hooks.ts:50-166` |
| PromptRequest inline elicitation schema | `types/hooks.ts:28-47` |
| Hook types (4 user-facing): command/prompt/agent/http | `schemas/hooks.ts:32-189` |
| HooksSchema top-level | `schemas/hooks.ts:211-213` |
| Hook matcher patterns (exact, pipe, regex, `*`) | `utils/hooks.ts:1346-1381` |
| Per-hook `if` condition with permission syntax | `utils/hooks.ts:1390-1421`, `schemas/hooks.ts:19-27` |
| Hook dedup keys | `utils/hooks.ts:1723-1801` |
| HookSource priority enum | `utils/hooks/hooksSettings.ts:15-21, 230-270` |
| Plugin hook registration | `utils/plugins/loadPluginHooks.ts:28-157` |
| Skill hook registration (session-scoped) | `utils/hooks/registerSkillHooks.ts:20-64` |
| Agent frontmatter hooks with Stop→SubagentStop rewrite | `utils/hooks/registerFrontmatterHooks.ts:38-45` |
| Workspace trust gating | `utils/hooks.ts:286-296` |
| TOOL_HOOK_EXECUTION_TIMEOUT_MS = 10 min | `utils/hooks.ts:166` |
| SessionEnd timeout = 1500 ms | `utils/hooks.ts:175-182` |
| Async hook protocol (first-line JSON) | `utils/hooks.ts:1112-1165` |
| asyncRewake wakes model on exit 2 | `utils/hooks.ts:205-245` |
| CLAUDE_ENV_FILE for SessionStart/Setup/CwdChanged/FileChanged | `utils/hooks.ts:917-926` |
| Plugin var substitution in command string | `utils/hooks.ts:822-857` |
| PowerShell shell selection | `utils/hooks.ts:790-792, 957-984` |
| SSRF guard for HTTP hooks | `utils/hooks/ssrfGuard.ts` (entire) |
| HTTP hook URL+env allowlist | `utils/hooks/execHttpHook.ts:43-58` |
| PreCompact returns newCustomInstructions | `utils/hooks.ts:3961-4025` |
| WorktreeCreate stdout = worktreePath | `utils/hooks.ts:4928-4958` |
| WorktreeRemove boolean signature | `utils/hooks.ts:4967-5003` |
| executePermissionRequestHooks | `utils/hooks.ts:4157` |
| executeInstructionsLoadedHooks | `utils/hooks.ts:4335` |
| executeElicitationHooks | `utils/hooks.ts:4470` |
| executeConfigChangeHooks | `utils/hooks.ts:4214` |
| SDK hook-event emission gating (SessionStart/Setup always) | `utils/hooks/hookEvents.ts:18, 83-91` |
| Hook telemetry definition | `utils/hooks.ts:5005-5022` |

## 13. Command Source References (selected claims)

| Claim | File:Lines |
|-------|-----------|
| 68+ built-in commands registered | `commands.ts:258-346` |
| INTERNAL_ONLY_COMMANDS | `commands.ts:225-254` |
| Feature-flagged commands | `commands.ts:60-123` |
| loadAllCommands composition order | `commands.ts:449-469` |
| REMOTE_SAFE_COMMANDS / BRIDGE_SAFE_COMMANDS | `commands.ts:619-675` |
| Command availability gating | `commands.ts:417-443` |
| Command type discriminated union | `types/command.ts:205-206` |
| PromptCommand fields (hooks, skillRoot, context, agent, paths, effort) | `types/command.ts:25-57` |
| CLAUDE_CONFIG_DIRECTORIES | `utils/markdownConfigLoader.ts:29-36` |
| getSkillDirCommands pipeline | `skills/loadSkillsDir.ts:638-804` |
| loadSkillsFromSkillsDir (dir-only) | `skills/loadSkillsDir.ts:407-480` |
| loadSkillsFromCommandsDir (legacy, namespaced) | `skills/loadSkillsDir.ts:566-623` |
| realpath dedup | `skills/loadSkillsDir.ts:743-763` |
| Dynamic skill discovery walk | `skills/loadSkillsDir.ts:861-915` |
| Conditional skills (paths frontmatter) | `skills/loadSkillsDir.ts:773-796, 997-1040` |
| Argument substitution impl | `utils/argumentSubstitution.ts:94-145` |
| Frontmatter schema | `utils/frontmatterParser.ts:10-59` |
| YAML special-char quoting | `utils/frontmatterParser.ts:79-121` |
| Path split + brace expansion | `utils/frontmatterParser.ts:189-232` |
| !`cmd` and ```! block executor | `utils/promptShellExecution.ts:48-143` |
| MCP skills blocked from shell exec | `skills/loadSkillsDir.ts:373-396` |
| ${CLAUDE_SKILL_DIR} and ${CLAUDE_SESSION_ID} substitution | `skills/loadSkillsDir.ts:358-369` |
| SkillTool filter (getSkillToolCommands) | `commands.ts:561-608` |
| `review` prompt command (inline template) | `commands/review.ts:33-43` |
| `security-review` prompt command | `commands/security-review.ts` |
| `init` command NEW_INIT_PROMPT pattern | `commands/init.ts:28-100+` |
| `plan` mode command | `commands/plan/index.ts` |
| `hooks` UI command | `commands/hooks/index.ts` |
| `clear` / `compact` / `resume` / `rewind` | `commands/{clear,compact,resume,rewind}/index.ts` |

---

## 14. Gap Matrix — Claude Code HAS / moai-adk-go LACKS

Cross-referenced against moai-adk-go shell hook wrappers in `.claude/hooks/moai/` and its Go `hook` subcommand handlers.

### 14.1 Hook events present in Claude Code but NOT in moai-adk-go

Based on the moai-adk-go doc inventory (SessionStart, UserPromptSubmit, Stop, SubagentStop, PreToolUse, PostToolUse, Notification, SessionEnd, WorktreeCreate, WorktreeRemove, TeammateIdle, TaskCompleted, PreCompact), the following Claude Code events are **NOT implemented**:

| Missing event | What it enables |
|---------------|-----------------|
| **PostToolUseFailure** | React to tool errors (e.g., auto-rollback on failed Edit, notify on Bash failure) |
| **StopFailure** | Handle API errors mid-turn (rate_limit, auth_failed, billing_error, max_output_tokens) |
| **SubagentStart** | Pre-subagent setup (inject context, configure tools, seed memory) |
| **PostCompact** | React to new compacted summary (persist summary, re-seed memory) |
| **PermissionRequest** | Programmatic permission decisions BEFORE a dialog is shown (auto-allow internal tools, auto-deny risky patterns) |
| **PermissionDenied** | React to denied tool calls (offer retry via `retry:true`, log anti-patterns) |
| **Setup** | Repo init/maintenance tasks that fire once per repo setup |
| **TaskCreated** | Block/modify task creation (audit, classify) |
| **Elicitation** | Auto-respond to MCP elicitation dialogs (form/URL prompts) |
| **ElicitationResult** | Observe/override user responses before they reach MCP servers |
| **ConfigChange** | React to user/project/local/policy/skills settings edits mid-session; can BLOCK changes |
| **InstructionsLoaded** | Audit which CLAUDE.md, rule, or memory file was loaded, with load_reason + globs |
| **CwdChanged** | React to cwd changes; inject env via CLAUDE_ENV_FILE |
| **FileChanged** | Watch specified files (.env, .envrc, config YAMLs) — fires on change/add/unlink |

### 14.2 Hook types missing in moai-adk-go

moai-adk-go only supports `type: 'command'` (shell scripts). Claude Code supports FOUR user-facing types:

| Missing hook type | Use case for moai |
|-------------------|-------------------|
| **`type: 'prompt'`** | LLM-evaluated gating (e.g., "Is this SPEC file valid?"). Runs against Haiku-class model. Returns `{ok, reason?}` JSON. Cheap. |
| **`type: 'agent'`** | Full sub-agent verifier with tools (e.g., "Verify tests ran and passed"). Multi-turn, bounded by timeout. |
| **`type: 'http'`** | Webhook integration (e.g., notify Slack/Discord on commit, post to observability service). With SSRF guard + env-var allowlisting. |

### 14.3 Hook features missing

| Feature | File:Lines | moai-adk status |
|---------|-----------|-----------------|
| `if` condition (permission-rule syntax: `Bash(git *)`) — avoids spawning hook when tool input doesn't match | `utils/hooks.ts:1390-1421`, `schemas/hooks.ts:19-27` | Not implemented — hooks always spawn |
| `shell: 'bash'` vs `shell: 'powershell'` per-hook selection | `utils/hooks.ts:790-792` | bash-only |
| `async: true` — backgrounded hook | `utils/hooks.ts:995-1030` | Not implemented |
| `asyncRewake: true` — backgrounded + wakes model on exit 2 | `utils/hooks.ts:205-245` | Not implemented |
| `once: true` — auto-remove after first success | `utils/hooks/registerSkillHooks.ts:36-43`, `schemas/hooks.ts:51-54` | Not implemented |
| Hook JSON output protocol (`decision`, `hookSpecificOutput`, `additionalContext`, `systemMessage`, `suppressOutput`, `stopReason`, `continue`) | `types/hooks.ts:50-166` | Partial — only exit codes |
| `hookSpecificOutput.additionalContext` injection into model turn | `utils/hooks.ts:622-628, 644-652` | Not implemented |
| `hookSpecificOutput.updatedInput` — hooks rewrite tool input | `utils/hooks.ts:618-620, 668-672` | Not implemented |
| `hookSpecificOutput.updatedMCPToolOutput` — hooks rewrite MCP tool output | `utils/hooks.ts:645-649` | Not implemented |
| Prompt elicitation from hook stdout (bidirectional) | `utils/hooks.ts:1066-1109`, `types/hooks.ts:28-47` | Not implemented |
| `CLAUDE_ENV_FILE` bash-export mechanism (SessionStart/Setup/CwdChanged/FileChanged) | `utils/hooks.ts:917-926` | Not implemented — env stays static |
| Hook progress streaming to SDK consumers | `utils/hooks/hookEvents.ts` | Not applicable (moai has no SDK) |
| Hook dedup across multiple settings sources | `utils/hooks.ts:1723-1801` | Not implemented |
| Workspace-trust gating on hooks | `utils/hooks.ts:286-296` | Not implemented |
| Plugin hook hot-reload on policySettings change | `utils/plugins/loadPluginHooks.ts:249-287` | Not applicable (no plugin system yet) |
| `policySettings.disableAllHooks` / `allowManagedHooksOnly` | `utils/hooks/hooksConfigSnapshot.ts:18-76` | Not implemented |
| `strictPluginOnlyCustomization` policy | `utils/hooks/hooksConfigSnapshot.ts:39-41` | Not applicable |
| Session-scoped hooks (in-memory, temporary) via `addSessionHook` | `utils/hooks/sessionHooks.ts` | Not implemented — all hooks are settings-file-based |

### 14.4 Hook lifecycle — moai MUST re-examine

- **WorktreeCreate as a PROVIDER hook** — Claude Code's WorktreeCreate reads stdout as the worktree path. moai-adk treats it as observational. If moai wants to integrate with `claude --worktree`, it should return `{hookEventName:'WorktreeCreate', worktreePath:...}` or print the absolute path on stdout.

### 14.5 Command features missing in moai-adk-go

| Feature | CC source | moai status |
|---------|-----------|-------------|
| `type: 'local'` vs `'local-jsx'` vs `'prompt'` taxonomy (execution context distinction) | `types/command.ts:16-155` | moai has only one form — markdown → skill routing |
| `context: 'inline' \| 'fork'` in frontmatter — run a skill as a sub-agent with isolated context/token budget | `types/command.ts:43-47` | Not implemented |
| `agent:` frontmatter — pick the sub-agent type when forked | `types/command.ts:47-48` | Not implemented |
| `effort:` frontmatter — per-skill Opus 4.7 effort level override | `types/command.ts:49` | moai has effortLevel in settings but not per-skill |
| `paths:` frontmatter — conditional skill activation based on touched files | `skills/loadSkillsDir.ts:159-178` | Not implemented |
| `disableModelInvocation` — hide skill from SkillTool | `types/command.ts:189` | Not implemented |
| `user-invocable` — allow/disallow user typing /skill-name | `types/command.ts:190` | Implicit (all moai skills are user-invocable) |
| `argumentHint` — gray hint in typeahead | `types/command.ts:186` | Partial — moai has it but no progressive-filling UI |
| `arguments` named-arg list — `$name` substitution | `utils/argumentSubstitution.ts:50-67` | Not implemented (moai uses only $ARGUMENTS) |
| `$ARGUMENTS[N]` indexed substitution | `utils/argumentSubstitution.ts:124-127` | Not implemented |
| Progressive argument hint as user types | `utils/argumentSubstitution.ts:76-83` | Not implemented |
| `hide-from-slash-command-tool` | frontmatter | Not implemented |
| `!`cmd` and ```! block execution inside skill body | `utils/promptShellExecution.ts:48-143` | Not implemented |
| `${CLAUDE_SKILL_DIR}` and `${CLAUDE_SESSION_ID}` substitution | `skills/loadSkillsDir.ts:358-369` | Not implemented |
| Dynamic skill discovery (walk up from touched file paths) | `skills/loadSkillsDir.ts:861-915` | Not implemented |
| Conditional skill activation on file paths | `skills/loadSkillsDir.ts:997-1040` | Not implemented |
| `commands_DEPRECATED` vs `skills` distinction (legacy migration) | `types/command.ts:191-197` | Not applicable |
| `source:` tracking — builtin / mcp / plugin / bundled / userSettings / projectSettings / managed | `types/command.ts:32-36` | Partial |
| `hooks:` in skill frontmatter — register session-scoped hooks when skill invoked | `types/command.ts:39, utils/hooks/registerSkillHooks.ts` | Not implemented |
| Plugin-sourced commands (with `pluginInfo.pluginManifest` metadata) | `types/command.ts:33-36` | Not applicable |
| `availability: ['claude-ai'\|'console']` gating | `types/command.ts:169-173` | Not applicable |
| `REMOTE_SAFE_COMMANDS` / bridge-safe filtering | `commands.ts:619-675` | Not applicable |
| `isSensitive: true` — redact args from conversation history | `types/command.ts:200` | Not implemented |
| `immediate: true` — execute without waiting for stop point, bypass queue | `types/command.ts:199` | Not implemented |
| Command versioning (`version:` frontmatter) | `types/command.ts:188` | Not implemented |
| Lazy command loading (`load: () => import(...)`) | most `commands/*/index.ts` | Not applicable (Go binary) |

### 14.6 Built-in commands Claude Code ships but moai-adk-go does NOT

Only listing ones moai might want to adopt/adapt:

- `/rewind` (checkpoint) — restore code and/or conversation to a previous point
- `/branch` (fork) — create a branch of the current conversation
- `/diff` — view uncommitted changes + per-turn diffs
- `/context` — visualize context usage as colored grid
- `/compact [instructions]` — keep summary in context
- `/export [filename]` — export conversation
- `/memory` — edit Claude memory files (moai has auto-memory but no edit command)
- `/permissions` (aliases: allowed-tools) — manage allow/deny rules
- `/doctor` — diagnose installation (moai has `moai doctor` CLI, but no `/doctor` slash)
- `/skills` — list skills (moai has Skill() programmatic but no /skills browser)
- `/agents` — manage agent configurations
- `/hooks` — view hook configurations (moai has `moai hook` CLI, but no `/hooks` browser)
- `/plugin` / `/marketplace` — plugin management
- `/mcp [enable\|disable server]` — manage MCP servers
- `/reload-plugins` — activate pending plugin changes mid-session
- `/init` — initialize CLAUDE.md (moai's `/moai project` serves similar role but scope differs — `/init` also optionally sets up skills and hooks)
- `/ide` — manage IDE integrations
- `/session` / `/remote` — show remote session URL and QR code
- `/theme` / `/color` / `/vim` / `/keybindings` — UX personalization
- `/usage` / `/cost` / `/stats` / `/insights` — usage visibility
- `/mobile` / `/desktop` — cross-device handoff
- `/resume` / `/continue` — resume previous conversation (moai's context-search serves part of this)
- `/files` — list files in context
- `/tasks` (alias: bashes) — background task management (moai has TaskList but no `/tasks` viewer)
- `/model` — per-session model switch
- `/effort` — set effort level (moai has effortLevel in settings, not a slash command)
- `/output-style` — deprecated but CC supports named output styles

### 14.7 Unique Claude Code primitives moai could adopt

- **SDK-level hook event streaming** (`utils/hooks/hookEvents.ts`) — stream hook execution to external consumers. moai's `moai glm` CG architecture could benefit.
- **Hook-driven prompt elicitation** (`types/hooks.ts:28-47`) — hooks can dynamically ask the user questions via stdin/stdout JSON protocol. Would replace some AskUserQuestion flows with deterministic scripts.
- **Sprint Contract–equivalent via `hookSpecificOutput`** — hooks can inject additionalContext that the model sees. This is a lighter-weight alternative to moai's evaluator-active Sprint Contract for some cases.
- **Conditional skill activation via `paths:`** — directly parallels moai's `paths:` frontmatter for CLAUDE.md rules, but applied to SKILLS — a skill only appears to the model after the model touches a matching file.
- **`context: fork` skills** — a skill that runs as a sub-agent with fresh context. moai has `Skill()` but it always expands into the current conversation.
- **CLAUDE_ENV_FILE bash-export protocol** — allow hooks to inject env vars into subsequent BashTool commands. Very powerful for session-level env management (.envrc, direnv integration).

---

## 15. Key Architectural Differences (CC vs moai-adk)

1. **Go binary vs embedded TypeScript**: moai-adk's hook handlers are Go functions (`moai hook <event>`). CC's hooks are shell commands + LLM prompts + HTTP endpoints + native JS callbacks. CC's handler selection is in TS runtime; moai's in Go bin.
2. **Multi-type hooks**: CC supports command/prompt/agent/http/callback/function. moai has only shell commands.
3. **Async hook protocol**: CC has two async variants (async, asyncRewake). moai runs hooks synchronously.
4. **Hook sources + policy**: CC merges user/project/local/policy/plugin/skill/session/builtin. moai loads from project `.claude/settings.json` only.
5. **Session-scoped hooks**: CC's `sessionHooks` layer is ephemeral (driven by skill/agent frontmatter at invocation time). moai has no equivalent.
6. **Hook output protocol**: CC hooks speak a rich JSON protocol with `additionalContext`, `updatedInput`, `systemMessage`, `permissionDecision`, `watchPaths`, etc. moai hooks communicate only via exit codes.
7. **Skill execution model**: CC skills have `inline` / `fork` execution contexts, with fork running as a sub-agent. moai skills all expand inline.
8. **Dynamic skill discovery**: CC walks up from file paths to load nested `.claude/skills/` dirs. moai loads once at session start.
9. **Trust + policy**: CC has workspace-trust + MDM policy gates. moai is single-user developer tool — simpler.
10. **Plugin system**: CC has full marketplace + plugin hot-reload. moai has no plugin primitive.

---

## 16. Recommendations for v3 Planners

Prioritization hints (not a decision — for the architect):

**High value, moderate effort**:
- Implement hook JSON output protocol — unlocks `additionalContext`, `permissionDecision`, `updatedInput`, `stopReason`, `systemMessage`. Small API surface change, large capability uplift.
- Add `type: 'prompt'` hooks — LLM-gated quality checks at near-zero cost.
- Add `async`/`asyncRewake` support — enables long-running Ralph-style loops without blocking.
- Add ConfigChange + InstructionsLoaded — for debuggability and audit of mid-session changes.

**High value, higher effort**:
- Implement `context: 'fork'` skills — align with `isolation: worktree` agent pattern. Enables cheap "verifier sub-agents" as skill invocations.
- Implement `paths:` conditional skill activation — matches moai's existing paths-frontmatter for rules.
- Add `type: 'http'` hooks (with SSRF guard) — for webhook integration.

**Low priority but strategic**:
- WorktreeCreate as provider hook — integrate with `claude --worktree` ecosystem.
- `/context`, `/rewind`, `/diff` as slash commands — UX parity.
- Session-scoped hook registration from skill frontmatter — makes moai skills self-contained.

**Avoid / defer**:
- Full plugin system (Claude Code's plugin/marketplace surface is huge) — compose around `Skill()` and rules instead.
- SDK hook event streaming — out of scope for a CLI-first tool.

---

End of findings.

# Wave 1.3 — Claude Code Agent/Team/Bridge Inventory

Research target: `/Users/goos/MoAI/AgentOS/claude-code-source-map/` — TypeScript dump of the Claude Code binary (April 10 2026 snapshot).

Scope: `bridge/`, `coordinator/`, `remote/`, `buddy/`, `native-ts/`, `assistant/`, `moreright/`, `schemas/`, plus any file with `team`/`agent`/`subagent` in the name (which pulled in `tools/AgentTool/`, `tools/TeamCreateTool/`, `tools/TeamDeleteTool/`, `tools/SendMessageTool/`, `tools/Task*Tool/`, `tasks/`, `utils/swarm/`, `utils/teammateMailbox.ts`).

---

## 0. Executive Summary

Claude Code's agent/team/bridge stack is NOT one system — it is three orthogonal systems fused at the top-level orchestrator:

| System | Purpose | Activation | State location |
|---|---|---|---|
| **AgentTool (sub-agent dispatch)** | Spawn one-shot or persistent sub-agents that run in-process or in the background | Always enabled | `AppState.tasks[agentId]`, `~/.claude/projects/{hash}/` transcripts |
| **Agent Teams / Swarm** (TeamCreate, SendMessage, Task*) | Multi-agent coordination with a shared task list and inboxes; teammates spawn in separate tmux panes or in-process AsyncLocalStorage contexts | `USER_TYPE=ant` OR `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` + `tengu_amber_flint` killswitch | `~/.claude/teams/{name}/config.json`, `~/.claude/teams/{name}/inboxes/{agent_name}.json`, `~/.claude/tasks/{name}/` |
| **Remote Control Bridge** (`bridge/`, `remote/`) | Control a Claude Code session from another machine via claude.ai web app or another CLI | `claude remote-control` command or `/remote-control` slash command | Anthropic cloud WebSocket + HTTP POST `/v1/sessions` |

The `coordinator/` directory adds a fourth overlay: "Coordinator mode" — a single sub-agent type (`worker`) that the main session delegates ALL substantive work to (feature-gated via `CLAUDE_CODE_COORDINATOR_MODE`).

`buddy/` is unrelated — it is a gamified companion sprite (duck/goose/blob with rarity stats) shown in the status line. Not an agent system.

`native-ts/` is unrelated — pure-TS fallback ports of Rust NAPI modules (file-index, color-diff, yoga-layout) for runtimes without native deps.

`moreright/` is a stub for an internal-only hook.

---

## 1. `bridge/` Subsystem

### 1.1 Identity and Purpose

The `bridge/` directory implements **Remote Control** — NOT agent-to-agent bridging. Remote Control lets a user:

- Control their local `claude` session from `claude.ai/code` (web browser)
- Control their local session from a second `claude` CLI on another machine
- Create remote CCR (Claude Cloud Runner) sessions and attach from local

Two architectures coexist:

1. **Env-based bridge** (`replBridge.ts`, ~100KB) — registers the local process as an "environment" via `POST /v1/environments/bridge`, polls for work via `GET /v1/environments/{id}/work`, dispatches sessions on-demand. Used by `claude remote-control` (standalone daemon).

2. **Env-less bridge** (`remoteBridgeCore.ts`) — directly creates a session via `POST /v1/code/sessions`, fetches worker credentials via `POST /bridge`, and maintains a persistent SSE+CCRClient connection. Used by the REPL's `/remote-control` slash command.

Evidence:
- bridge/types.ts:63-69 — SpawnMode union: `'single-session' | 'worktree' | 'same-dir'`
- bridge/remoteBridgeCore.ts:1-28 — architectural comment: "Env-less = no Environments API layer"
- bridge/replBridge.ts:70-81 — `ReplBridgeHandle` public type

### 1.2 Transport Mechanisms

Three transport layers:

| Transport | Read | Write | Used by |
|---|---|---|---|
| **HybridTransport** (v1) | WebSocket | HTTP POST to Session-Ingress | `createV1ReplTransport` in replBridge.ts |
| **SSETransport + CCRClient** (v2) | Server-Sent Events | HTTP POST to CCR `/worker/events` | `createV2ReplTransport` in replBridgeTransport.ts |
| **SessionsWebSocket** (remote client) | WebSocket | Same WebSocket | `remote/RemoteSessionManager.ts` |

Evidence:
- bridge/replBridgeTransport.ts:11-22 — comment: "v1: HybridTransport (WS reads + POST writes) ; v2: SSETransport (reads) + CCRClient (writes)"
- bridge/replBridgeTransport.ts:23-70 — `ReplBridgeTransport` interface unifying v1 and v2
- remote/SessionsWebSocket.ts:82-111 — WebSocket endpoint: `wss://api.anthropic.com/v1/sessions/ws/{sessionId}/subscribe?organization_uuid=...`
- remote/SessionsWebSocket.ts:108-118 — Auth via `Authorization: Bearer {accessToken}` header; `anthropic-version: 2023-06-01`

### 1.3 Message Schemas

Control request/response over the bridge WebSocket uses Anthropic's SDK control protocol:

```
SDKControlRequest: { type: 'control_request', request_id, request: { subtype: 'initialize' | 'interrupt' | 'set_model' | 'set_permission_mode' | 'set_max_thinking_tokens' | 'can_use_tool' , ... } }
SDKControlResponse: { type: 'control_response', response: { subtype: 'success' | 'error', request_id, response?, error? } }
SDKControlCancelRequest: { type: 'control_cancel_request', request_id }
```

Evidence:
- remote/RemoteSessionManager.ts:187-213 — `handleControlRequest` routes `'can_use_tool'` subtype to a permission-request callback
- bridge/bridgeMessaging.ts:246-391 — `handleServerControlRequest` handles all 5 control subtypes
- bridge/bridgeMessaging.ts:269-283 — outbound-only mode rejects mutating requests with `'error'` subtype

### 1.4 Bridge File-by-File Role

| File | LOC | Role |
|---|---|---|
| `bridgeMain.ts` | 115K | Orchestrator for `claude remote-control` standalone daemon |
| `replBridge.ts` | 100K | Orchestrator for `/remote-control` slash command (lives inside REPL) |
| `remoteBridgeCore.ts` | 39K | Env-less bridge init (direct /v1/code/sessions + /bridge) |
| `bridgeApi.ts` | 18K | HTTP client for Environments API (register, poll, ack, stop, deregister, heartbeat) |
| `bridgeMessaging.ts` | 15K | Shared ingress parsing + control_request response logic (used by both cores) |
| `replBridgeTransport.ts` | 15K | v1/v2 transport abstraction |
| `bridgeUI.ts` | 16K | Terminal UI for bridge status line |
| `initReplBridge.ts` | 23K | Startup path selection (env-based vs env-less, via `tengu_bridge_repl_v2` gate) |
| `createSession.ts` | 12K | POST /v1/sessions for bridge session creation |
| `sessionRunner.ts` | 18K | Spawns Claude child processes per bridge session |
| `jwtUtils.ts` | 9K | Worker JWT decoding + proactive refresh scheduler |
| `trustedDevice.ts` | 8K | `X-Trusted-Device-Token` header resolution |
| `inboundMessages.ts` | 3K | Convert inbound SDKMessage → prompt submission |
| `inboundAttachments.ts` | 6K | Resolve attached-file references from inbound messages |
| `workSecret.ts` | 5K | Decode base64url JWT work secret from pollForWork response |
| `bridgePointer.ts` | 8K | `~/.claude/projects/{hash}/bridge-pointer.json` for crash recovery (4h mtime TTL) |
| `capacityWake.ts` | 2K | Wake bridge poll loop when capacity frees up |
| `codeSessionApi.ts` | 5K | `POST /v1/code/sessions`, `POST /bridge` (env-less path) |
| `flushGate.ts` | 2K | Coordinate initial message flush order |
| `bridgeEnabled.ts` | 8K | Feature flag resolution (`USER_TYPE=ant` ∧ OAuth ∧ gate checks) |
| `bridgeConfig.ts` | 2K | Parse `--bridge-*` CLI flags |
| `envLessBridgeConfig.ts` | 7K | Env-less path config (heartbeat intervals, HTTP timeouts) |
| `bridgeDebug.ts` | 5K | Fault injection harness for bridge tests |
| `bridgePermissionCallbacks.ts` | 1K | Permission callback registration |
| `sessionIdCompat.ts` | 3K | Translate between `session_*` (v1) and `cse_*` (v2) session IDs |
| `types.ts` | 10K | All bridge protocol type definitions |
| `replBridgeHandle.ts` | 1K | Global handle getter for active bridge |
| `pollConfig.ts` / `pollConfigDefaults.ts` | 8K | Exponential backoff for env-based polling |
| `debugUtils.ts` | 4K | Shared debug-logging helpers |

### 1.5 Bridge Auth Flow

1. User runs `claude` with OAuth tokens from prior `/login` (stored in keychain)
2. REPL calls `utils/auth.ts::getClaudeAIOAuthTokens()` → `{ accessToken, orgUUID }`
3. Bridge sends `Authorization: Bearer {accessToken}`, `anthropic-version: 2023-06-01`, `anthropic-beta: environments-2025-11-01` OR `ccr-byoc-2025-07-29`
4. Server validates: subscription tier, organization membership, trusted device token (`X-Trusted-Device-Token` when `tengu_sessions_elevated_auth_enforcement` is on)
5. On 401: `withOAuthRetry` calls `onAuth401` callback (closure over `utils/auth.ts::handleOAuth401Error`) → refresh OAuth → retry once

Evidence: bridge/bridgeApi.ts:91-139 (withOAuthRetry), bridge/trustedDevice.ts (trusted device).

### 1.6 Remote Control Peer Messaging (`SendMessage` bridge: target)

`bridge:{session-id}` address scheme in `SendMessageTool` reaches a peer Claude via the bridge cloud server.

Evidence:
- tools/SendMessageTool/SendMessageTool.ts:586-600 — `parseAddress(input.to).scheme === 'bridge'` triggers `ask`-mode permission (cross-machine prompt injection requires explicit consent)
- tools/SendMessageTool/SendMessageTool.ts:741-770 — `postInterClaudeMessage(addr.target, input.message)` via `bridge/peerSessions.js`
- tools/SendMessageTool/prompt.ts:7-36 — feature('UDS_INBOX') gated (not all builds)

### 1.7 Why Bridge ≠ moai-adk-go Use Case

Claude Code's bridge is specifically for Anthropic's cloud (WebSocket to `api.anthropic.com`) + OAuth subscription. It is NOT a general agent-to-agent transport and cannot be reused directly for moai-adk without running against Anthropic infrastructure. See Section 8 Gap.

---

## 2. Coordinator & Remote

### 2.1 Coordinator Mode (`coordinator/coordinatorMode.ts`)

Coordinator mode is a feature-gated UX where the main session acts as a pure orchestrator and delegates ALL work to background workers via `AgentTool(subagent_type='worker')`.

Activation: `feature('COORDINATOR_MODE')` build flag + `CLAUDE_CODE_COORDINATOR_MODE` env var.

Evidence:
- coordinator/coordinatorMode.ts:36-41 — gate check
- coordinator/coordinatorMode.ts:29-34 — `INTERNAL_WORKER_TOOLS` set excludes TeamCreate/TeamDelete/SendMessage from the worker tool-list description (workers can't spawn other workers / chat)
- coordinator/coordinatorMode.ts:111-369 — `getCoordinatorSystemPrompt()` defines the full coordinator prompt (380-line system prompt heavy on "parallelism is your superpower", task-notification XML, continue-vs-spawn-fresh decision matrix)

Key concept: when coordinator mode is on, EVERY `AgentTool` spawn is forced to `shouldRunAsync = true` (see AgentTool.tsx:567: `isCoordinator || forceAsync || ...`), and ALL results arrive as `<task-notification>` XML inside user-role messages.

### 2.2 `remote/` — Remote CCR Sessions

Separate from bridge — `remote/` lets the REPL ATTACH to an already-running remote CCR session (e.g., via `claude --session-id`). 

Files:
- `RemoteSessionManager.ts` (9.3KB): Manages a single CCR session. Uses `SessionsWebSocket` for inbound SDK messages + control requests, and `sendEventToRemoteSession` (HTTP POST) for outbound user messages. Exposes: `connect()`, `sendMessage()`, `respondToPermissionRequest()`, `cancelSession()`, `disconnect()`, `reconnect()`.
- `SessionsWebSocket.ts` (12.5KB): WebSocket client for `wss://api.anthropic.com/v1/sessions/ws/{sessionId}/subscribe`. Implements ping (30s interval), reconnect logic (5 attempts, 2s delay), permanent close codes (4003 unauth), and limited retries for 4001 (session-not-found during compaction).
- `sdkMessageAdapter.ts` (9KB): Converts SDKMessage (cloud format) → internal Message (REPL format) for rendering.
- `remotePermissionBridge.ts` (2.4KB): Creates synthetic `AssistantMessage` for tool uses running on remote CCR (the local REPL renders the permission prompt but the tool executes remotely).

Evidence:
- remote/SessionsWebSocket.ts:17-27 — constants: `RECONNECT_DELAY_MS=2000, MAX_RECONNECT_ATTEMPTS=5, PING_INTERVAL_MS=30000, MAX_SESSION_NOT_FOUND_RETRIES=3`
- remote/SessionsWebSocket.ts:34-36 — `PERMANENT_CLOSE_CODES = new Set([4003])`
- remote/RemoteSessionManager.ts:189-214 — control-request dispatch (currently only `can_use_tool` subtype supported)

### 2.3 Remote Agent Task (`tasks/RemoteAgentTask/`)

When `AgentTool` is invoked with `isolation: 'remote'` (ant-only), the agent spawn is teleported to a remote CCR container:

Evidence:
- tools/AgentTool/AgentTool.tsx:435-482 — `"external" === 'ant' && effectiveIsolation === 'remote'` branch calls `teleportToRemote({ initialMessage: prompt, description, signal })` then `registerRemoteAgentTask({ remoteTaskType: 'remote-agent', session, command, context })`
- Returns `{ status: 'remote_launched', taskId, sessionUrl, description, prompt, outputFile }`
- The remote session runs on Anthropic CCR; local session only holds the task state + output file polling

---

## 3. Buddy (Pair Agent)

**FINDING: `buddy/` is NOT a pair-agent system.**

`buddy/` implements a gamified companion sprite — a virtual pet (duck, goose, blob, cat, dragon, octopus, owl, ...) that appears in the Claude Code UI. It has rarity tiers (common/uncommon/rare/epic/legendary), stats (DEBUGGING, PATIENCE, CHAOS, WISDOM, SNARK), eyes, hats, and a shiny flag. The species is hash-derived from the user's OAuth account UUID (deterministic) + a SALT.

Evidence:
- buddy/types.ts:1-148 — full type inventory
- buddy/companion.ts:106-133 — `roll(userId)` deterministic generator from `hashString(userId + SALT)`
- buddy/CompanionSprite.tsx — 45KB React component rendering ASCII art of the companion
- buddy/useBuddyNotification.tsx — notification UI integration
- buddy/prompt.ts (1.4KB) — a short system-prompt snippet mentioning the buddy (if any)

**Conclusion**: `buddy/` has no overlap with moai-adk team/subagent concerns and should not be studied for v3 design.

---

## 4. Agent Frontmatter Schema

The canonical agent frontmatter schema is defined in `tools/AgentTool/loadAgentsDir.ts:73-99` (AgentJsonSchema) and parsed for markdown files at `tools/AgentTool/loadAgentsDir.ts:541-755` (parseAgentFromMarkdown).

### 4.1 Full Schema (JSON + markdown)

```yaml
---
name: expert-backend          # required, agentType
description: ...              # required, 1-char min
tools: [Read, Write, ...]     # optional, CSV in markdown
disallowedTools: [Bash, ...]  # optional, subtracted from allowlist
prompt: "..."                 # JSON only; markdown uses body content
model: sonnet|opus|haiku|"inherit"|custom-string    # optional
effort: low|medium|high|xhigh|max|<integer>          # optional
permissionMode: default|acceptEdits|bypassPermissions|plan|auto|bubble    # optional
mcpServers: [name-ref | {name: config}, ...]        # optional
hooks: { SessionStart: [...], ... }   # optional, full HooksSchema
maxTurns: <positive int>              # optional
skills: [skill-name-1, skill-name-2]  # optional, comma-separated in markdown
initialPrompt: "..."                  # optional, prepended to first user turn
memory: user|project|local            # optional, scope for auto-memory
background: true                      # optional, boolean; forces async execution
isolation: worktree | remote          # optional; remote is ant-only
color: <AgentColorName>               # markdown only
---

Body is the system prompt (markdown files).
```

Evidence: loadAgentsDir.ts:73-99 (AgentJsonSchema), loadAgentsDir.ts:541-755 (parseAgentFromMarkdown)

### 4.2 Extra fields not in moai-adk today

| Field | Purpose | Where defined |
|---|---|---|
| `skills` | Preload skills in agent's tool pool | loadAgentsDir.ts:90, `parseSlashCommandToolsFromFrontmatter` |
| `initialPrompt` | Injected as first user turn; slash commands work | loadAgentsDir.ts:91, runAgent.ts prepends |
| `memory` | Persistent agent memory scope (user/project/local) | `agentMemory.ts`, `agentMemorySnapshot.ts` |
| `criticalSystemReminder_EXPERIMENTAL` | Short re-injected message at every turn | loadAgentsDir.ts:121 |
| `requiredMcpServers` | Agent only available when listed MCP servers have tools | loadAgentsDir.ts:122 |
| `omitClaudeMd` | Skip CLAUDE.md hierarchy (read-only agents like Explore/Plan) | loadAgentsDir.ts:128-132 |
| `background` | Always run as background task | loadAgentsDir.ts:93, AgentTool.tsx:567 |
| `isolation: remote` | Run in remote CCR (ant-only) | AgentTool.tsx:435-482 |

### 4.3 Permission Modes

| Mode | Behavior |
|---|---|
| `default` | Ask for permission on each non-auto-approved tool call |
| `acceptEdits` | Auto-approve file edits |
| `bypassPermissions` | Skip all permission prompts (DANGEROUS) |
| `plan` | Plan mode — must `ExitPlanMode` before any write |
| `auto` | Classifier-based auto-approval (gated by `tengu_auto_mode`) |
| `bubble` | Surface permission prompts to parent (used by fork subagents) |

Evidence: `utils/permissions/PermissionMode.ts:PERMISSION_MODES` referenced at loadAgentsDir.ts:32-34.

### 4.4 Built-in Agents (`tools/AgentTool/built-in/`)

| Agent | Model | Permission | Purpose |
|---|---|---|---|
| `general-purpose` | default (getDefaultSubagentModel) | acceptEdits | General research + implementation |
| `Explore` | ant: `inherit`, ext: `haiku` | disallow {Agent, ExitPlanMode, FileEdit, FileWrite, NotebookEdit} | Fast read-only code search |
| `Plan` | `inherit` | disallow (same as Explore) | Architectural plan design, read-only |
| `Code Guide` (`CLAUDE_CODE_GUIDE_AGENT`) | default | - | User-facing Claude Code help (non-SDK only) |
| `Statusline Setup` | default | - | Configure statusline during onboarding |
| `Verification` | (gated by `tengu_hive_evidence`) | - | Independent QA verification |
| `worker` | default | - | Coordinator mode's single worker type (see coordinator/workerAgent.js) |
| `fork` | `inherit` | `bubble` | Fork subagent — inherits parent's context (feature('FORK_SUBAGENT') gated) |

Evidence: `builtInAgents.ts:45-72` (getBuiltInAgents), `built-in/generalPurposeAgent.ts`, `built-in/exploreAgent.ts`, `built-in/planAgent.ts`, `forkSubagent.ts:60-71`.

### 4.5 Hooks Schema

Agents may declare hooks in their frontmatter (schemas/hooks.ts:165-213). Four hook types supported:

| Type | Format | Purpose |
|---|---|---|
| `command` | `{ type: 'command', command: '...', shell: bash|powershell, timeout, async, asyncRewake, once, if, statusMessage }` | Bash/PowerShell command execution |
| `prompt` | `{ type: 'prompt', prompt: 'eval with $ARGUMENTS', model, timeout, once, if, statusMessage }` | LLM prompt evaluation |
| `http` | `{ type: 'http', url, headers, allowedEnvVars, timeout, once, if, statusMessage }` | POST JSON to URL |
| `agent` | `{ type: 'agent', prompt: 'verify X', model, timeout, once, if, statusMessage }` | Agentic verifier (another LLM call that decides pass/fail) |

All four support:
- `if` (permission-rule syntax like `Bash(git *)`) — evaluated against tool_name + tool_input to filter
- `once` — hook runs once then self-removes
- `statusMessage` — custom spinner text

`command` uniquely supports:
- `async` — fire-and-forget
- `asyncRewake` — async + wake the model on exit code 2

Evidence: schemas/hooks.ts:31-171 (buildHookSchemas).

---

## 5. Sub-agent Dispatch Flow

### 5.1 `AgentTool` input schema (tools/AgentTool/AgentTool.tsx:82-125)

```ts
{
  description: string          // required, 3-5 word summary
  prompt: string               // required, task for the agent
  subagent_type?: string       // optional; defaults to 'general-purpose' (or FORK when gate is on)
  model?: 'sonnet'|'opus'|'haiku'  // optional override
  run_in_background?: boolean  // optional; forces async
  name?: string                // optional (teams only): makes SendMessage({to: name}) addressable
  team_name?: string           // optional (teams only): inherits from context if absent
  mode?: PermissionMode        // optional: spawn in 'plan' mode requires plan approval
  isolation?: 'worktree' | 'remote'  // optional; 'remote' is ant-only
  cwd?: string                 // optional (KAIROS): explicit working dir override
}
```

### 5.2 Dispatch decision tree (tools/AgentTool/AgentTool.tsx:239-765)

```
AgentTool.call(input, toolUseContext):
  1. Check team_name gate — if set, require isAgentSwarmsEnabled()
  2. Teammate-in-process cannot spawn more teammates (flat roster invariant)
  3. Teammate-in-process cannot spawn background agents
  4. If teamName AND name present:
     → spawnTeammate(...)  // Agent Teams path (tmux pane or in-process)
     → return { status: 'teammate_spawned', ... }
  5. Resolve subagent_type (or FORK_AGENT if fork gate is on + subagent_type omitted)
  6. Validate required MCP servers; wait up to 30s for pending connections
  7. If isolation === 'remote' (ant-only):
     → teleportToRemote() → registerRemoteAgentTask()
     → return { status: 'remote_launched', taskId, sessionUrl }
  8. Build system prompt (fork path: inherit parent's rendered bytes; normal: selectedAgent.getSystemPrompt())
  9. If isolation === 'worktree':
     → createAgentWorktree(slug) → worktreeInfo
  10. Assemble worker tool pool via assembleToolPool(workerPermissionContext, mcp.tools)
  11. Determine shouldRunAsync:
     run_in_background === true || selectedAgent.background === true
     || isCoordinator || forceAsync || assistantForceAsync || proactive
  12a. If shouldRunAsync:
     → registerAsyncAgent() → runAsyncAgentLifecycle() (detached)
     → return { status: 'async_launched', agentId, outputFile, canReadOutputFile }
  12b. Else (sync):
     → registerAgentForeground() with optional auto-background timer
     → stream runAgent() messages through `for await` loop
     → race against backgroundPromise; if backgrounded mid-flight, switch paths
     → return { status: 'completed', content: AgentToolResult, prompt }
  13. On completion: cleanupWorktreeIfNeeded (idempotent, skip if hook-based)
```

Evidence: tools/AgentTool/AgentTool.tsx:239-1200 (main `call` function).

### 5.3 Sub-agent context isolation

Sub-agents run in the SAME Node.js process but with isolated "context" provided by:

| Isolation Type | Mechanism |
|---|---|
| **Conversation history** | Sub-agent starts with its own `promptMessages[]`; parent's messages are NOT visible except in fork path |
| **System prompt** | Sub-agent's `selectedAgent.getSystemPrompt()` replaces parent's (fork path inherits via `override.systemPrompt`) |
| **Tool pool** | Sub-agent's `workerTools = assembleToolPool(workerPermissionContext, ...)` — independent of parent restrictions |
| **Permission mode** | `selectedAgent.permissionMode ?? 'acceptEdits'` (overrides parent's current mode) |
| **CWD** | `isolation: 'worktree'` → `runWithCwdOverride(worktreePath, fn)` wraps all filesystem operations |
| **Agent identity** | `runWithAgentContext({ agentId, agentType: 'subagent', subagentName, ... })` via AsyncLocalStorage |
| **MCP servers** | Agent's frontmatter `mcpServers: [...]` additively merged with parent's (Section 4.3) |
| **Messages/file state** | `cloneFileStateCache()` for Read tool's per-agent cache |
| **Transcript** | `setAgentTranscriptSubdir(agentId)` isolates JSONL transcript file |

Evidence: tools/AgentTool/AgentTool.tsx:570-637 (runAgentParams construction), runAgent.ts:1-150 (sub-agent entry).

### 5.4 Result return schema

Sync return (`status: 'completed'`):
```ts
{
  content: AgentToolResult,   // { content: ContentBlock[], totalToolUseCount, totalDurationMs, ... }
  prompt: string              // original user prompt
}
```

Async return (`status: 'async_launched'`):
```ts
{
  agentId: string,
  description: string,
  prompt: string,
  outputFile: string,         // ~/.claude/{project-hash}/tasks/{taskId}.json
  canReadOutputFile: boolean  // true if caller has Read or Bash
}
```

When background agent completes, a `<task-notification>` XML block is injected into the PARENT session's conversation as a user-role message (AgentTool.tsx:978-991 calls `enqueueAgentNotification`).

Format (from coordinator/coordinatorMode.ts:144-160):
```xml
<task-notification>
  <task-id>{agentId}</task-id>
  <status>completed|failed|killed</status>
  <summary>{human-readable status}</summary>
  <result>{agent's final text response}</result>
  <usage>
    <total_tokens>N</total_tokens>
    <tool_uses>N</tool_uses>
    <duration_ms>N</duration_ms>
  </usage>
  <worktree>
    <worktreePath>...</worktreePath>
    <worktreeBranch>...</worktreeBranch>
  </worktree>
</task-notification>
```

### 5.5 Parallel spawn handling

- Parent can emit N tool_use blocks in a single assistant turn; `StreamingToolExecutor` runs them concurrently
- Each spawn produces an independent `taskId` (`createAgentId()` uses random UUID)
- Spawns with `run_in_background: true` return immediately; results arrive as separate turns
- Max parallel: no hard cap, but permission-prompt serialization limits practical parallelism
- The fork path encourages parallel: "launch parallel forks in one message" (AgentTool prompt line 86)

Evidence: tools/AgentTool/AgentTool.tsx:686-764 (async path setup), AgentTool prompt.ts:264 ("You MUST send a single message with multiple Agent tool use content blocks").

### 5.6 Timeout / cancellation

- **Child AbortController**: each agent gets its own `agentBackgroundTask.abortController` (AgentTool.tsx:688-698). Background agents do NOT link to parent's abort controller — they survive Esc on the main thread (AgentTool.tsx:694).
- **TaskStopTool**: provides the `task_id` from the launch result → `stopTask(id)` → `taskImpl.kill(taskId, setAppState)` → calls `abortController.abort()`.
- **Parent interrupt**: when parent is stopped, sync sub-agents abort (they share the parent's running context); async sub-agents continue.
- **Max turns**: agent frontmatter `maxTurns` is enforced by `runAgent` — the while-loop breaks when turn count exceeds.

Evidence: Task.ts:72-76 (Task interface with `kill`), tasks/stopTask.ts:38-100 (stopTask logic).

---

## 6. Team API Implementation

### 6.1 Activation gate

`utils/agentSwarmsEnabled.ts:1-44`:
```
USER_TYPE === 'ant' → enabled
OR (CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS truthy OR --agent-teams CLI flag)
   AND tengu_amber_flint GrowthBook gate (killswitch)
```

### 6.2 TeamCreate (`tools/TeamCreateTool/TeamCreateTool.ts:74-240`)

Input:
```
{ team_name: string, description?: string, agent_type?: string }
```

Side effects:
1. Validates: only one team per session (`appState.teamContext?.teamName` check) — cannot spawn nested teams
2. `generateUniqueTeamName(team_name)` — if name collides, generates a new word slug via `generateWordSlug()`
3. Computes deterministic lead agent ID: `formatAgentId(TEAM_LEAD_NAME, finalTeamName)` → `"team-lead@{teamName}"`
4. Resolves leader model from `appState.mainLoopModelForSession ?? mainLoopModel ?? getDefaultMainLoopModel()`
5. Writes team file to `~/.claude/teams/{sanitized-name}/config.json`:
   ```
   TeamFile {
     name, description, createdAt,
     leadAgentId, leadSessionId,
     members: [{ agentId, name: 'team-lead', agentType, model, joinedAt, tmuxPaneId: '', cwd, subscriptions: [] }]
   }
   ```
6. Calls `registerTeamForSessionCleanup(finalTeamName)` — team gets cleaned up on SIGINT
7. Resets task list directory at `~/.claude/tasks/{sanitized-name}/` (task numbering starts at 1)
8. Registers leader team name so `getTaskListId()` returns it for the leader (instead of falling through to session ID)
9. Updates `AppState.teamContext` with teamName, leadAgentId, teammates map (leader only)
10. Emits `tengu_team_created` analytics event with `teammate_mode` (tmux|in-process)
11. Returns `{ team_name, team_file_path, lead_agent_id }`

Notes:
- Leader does NOT get `CLAUDE_CODE_AGENT_ID` env var (breaks `isTeammate()` detection which leader depends on)
- `teammate_mode` is resolved via `getResolvedTeammateMode()` (in-process vs tmux)

### 6.3 TeamDelete (`tools/TeamDeleteTool/TeamDeleteTool.ts:74-139`)

No input. Side effects:
1. Read team file. If any non-lead members are still `isActive !== false`, REJECT with message: "Use requestShutdown to gracefully terminate teammates first"
2. If all teammates idle/dead: `cleanupTeamDirectories(teamName)` — destroys worktrees, team dir, tasks dir
3. `unregisterTeamForSessionCleanup` (prevent double-cleanup on gracefulShutdown)
4. `clearTeammateColors()` — global color pool reset
5. `clearLeaderTeamName()` — `getTaskListId()` now falls back to session ID
6. Clear `AppState.teamContext` and `AppState.inbox`
7. Emit `tengu_team_deleted` analytics
8. Return `{ success, message, team_name }`

### 6.4 SendMessage (`tools/SendMessageTool/SendMessageTool.ts:67-917`)

Input:
```
{
  to: string,           // teammate name | "*" | "uds:/path" | "bridge:session-id"
  summary?: string,     // 5-10 word preview; REQUIRED when message is plain string
  message: string | StructuredMessage
}

StructuredMessage = 
  | { type: 'shutdown_request', reason? }
  | { type: 'shutdown_response', request_id, approve: boolean, reason? }
  | { type: 'plan_approval_response', request_id, approve: boolean, feedback? }
```

Routing logic (SendMessageTool.ts:741-913):

| Target format | Handler | Delivery |
|---|---|---|
| `"*"` (broadcast, string only) | `handleBroadcast` | Write plain message to every non-self teammate's mailbox file |
| Teammate name, string message | (1) Check `agentNameRegistry` for in-process subagent → `queuePendingMessage` or `resumeAgentBackground` (2) Else `handleMessage` → `writeToMailbox` |
| `"uds:/path"` (same-machine cross-session) | `sendToUdsSocket` (via UDS_INBOX feature gate) | Unix domain socket write |
| `"bridge:session-id"` (cross-machine) | `postInterClaudeMessage` | Anthropic cloud relay (requires `ask` permission) |
| Structured shutdown_request | `handleShutdownRequest` | `createShutdownRequestMessage` → mailbox |
| Structured shutdown_response (approve) | `handleShutdownApproval` | Abort in-process task OR `gracefulShutdown(0, 'other')` |
| Structured shutdown_response (reject) | `handleShutdownRejection` | `createShutdownRejectedMessage` → team-lead mailbox |
| Structured plan_approval_response | `handlePlanApproval`/`handlePlanRejection` | Team-lead only; writes `plan_approval_response` JSON to recipient mailbox |

Permission flow:
- Default: `behavior: 'allow'`
- `bridge:` scheme: `behavior: 'ask'` with `safetyCheck` decision reason — cross-machine bridge messaging requires explicit user consent and is bypass-immune

Validation rules (SendMessageTool.ts:604-718):
- `summary` required when message is a plain string
- No `@` in `to` field (bare name only, agents have one-team-per-session)
- `shutdown_response` must target `TEAM_LEAD_NAME` ("team-lead")
- `shutdown_response` with `approve: false` requires `reason`
- Structured messages cannot be broadcast to `"*"` or across UDS/bridge
- Bridge `to:` supports only string messages

### 6.5 Task* Tools (V2 task list)

All Task* tools are gated by `isTodoV2Enabled()` and write to `~/.claude/tasks/{taskListId}/`:

**TaskCreate** (tools/TaskCreateTool/TaskCreateTool.ts:48-138):
```
Input: { subject, description, activeForm?, metadata? }
Output: { task: { id, subject } }
```
- Runs `TaskCreated` hooks; if any returns blocking error → deletes task + throws
- Auto-expands the task-list view in UI

**TaskUpdate** (tools/TaskUpdateTool/TaskUpdateTool.ts:88-406):
```
Input: {
  taskId, subject?, description?, activeForm?,
  status?: 'pending'|'in_progress'|'completed'|'failed'|'deleted',
  addBlocks?, addBlockedBy?, owner?, metadata?
}
Output: { success, taskId, updatedFields, error?, statusChange?, verificationNudgeNeeded? }
```
- Auto-sets `owner` to current agent when teammate marks in_progress without owner
- On `status: 'completed'`: runs `TaskCompleted` hooks; blocking error → return failure
- On owner change: writes `task_assignment` JSON to new owner's mailbox
- On deletion: file removed from disk
- "Verification nudge": if main-thread agent completes 3+ tasks with none titled "verif*", appends nudge message

**TaskList** (tools/TaskListTool/TaskListTool.ts:33-116):
- No input
- Returns all tasks with `metadata._internal` filtered out
- `blockedBy` filtered to exclude resolved tasks
- Rendered as `#id [status] subject (owner) [blocked by #x, #y]`

**TaskGet** (tools/TaskGetTool/TaskGetTool.ts:38-127):
```
Input: { taskId }
Output: { task: { id, subject, description, status, blocks, blockedBy } | null }
```

**TaskStop** (tools/TaskStopTool/TaskStopTool.ts:39-131):
```
Input: { task_id?, shell_id? (deprecated KillShell compat) }
Output: { message, task_id, task_type, command? }
```
- Dispatches to `Task.kill(taskId, setAppState)` via `getTaskByType(task.type)`
- Suppresses "exit 137" notification for bash tasks; agent tasks send partial-result notification

### 6.6 Task List shared storage

- Each team has 1:1 task list correspondence: Team = TaskList
- `getTaskListId()` returns `setLeaderTeamName()` value for leader, `getSessionId()` for teammates/solo
- Storage path: `~/.claude/tasks/{taskListId}/` with individual JSON files per task
- `listTasks()`, `createTask()`, `updateTask()`, `deleteTask()`, `blockTask()`, `claimTask()` all atomic via file locks

Evidence: tools/TeamCreateTool/TeamCreateTool.ts:184-191, TaskCreate/Update/List/Get tools all call `getTaskListId()`.

### 6.7 Task Framework (`Task.ts`, `tasks.ts`, `tasks/`)

Task types (Task.ts:6-14):
```
'local_bash' | 'local_agent' | 'remote_agent' | 'in_process_teammate' | 'local_workflow' | 'monitor_mcp' | 'dream'
```

Task ID prefixes (Task.ts:79-87): b/a/r/t/w/m/d + 8 random base36 chars.

Task statuses (Task.ts:15-29): `pending | running | completed | failed | killed`. Terminal = last 3.

Each Task type implements `{ name, type, kill(taskId, setAppState): Promise<void> }` (Task.ts:72-76).

Registered tasks (tasks.ts:22-32):
- LocalShellTask (bash/powershell background processes)
- LocalAgentTask (sub-agent backgrounded or launched with `run_in_background`)
- RemoteAgentTask (`isolation: 'remote'` spawns)
- DreamTask (gated)
- LocalWorkflowTask (feature: WORKFLOW_SCRIPTS)
- MonitorMcpTask (feature: MONITOR_TOOL)
- InProcessTeammateTask (registered separately, not in getAllTasks) — team mode in-process teammate

### 6.8 Teammate Mailbox System (`utils/teammateMailbox.ts`)

**Storage**: `~/.claude/teams/{team_name}/inboxes/{agent_name}.json`

**Concurrency**: File-backed with `proper-lockfile` retries (10 retries, 5-100ms backoff).

**Message shape** (teammateMailbox.ts:43-50):
```ts
TeammateMessage = { from, text, timestamp, read: boolean, color?, summary? }
```

**Structured sub-types** (Zod schemas):
- `PlanApprovalRequestMessage` — `{ type: 'plan_approval_request', from, timestamp, planFilePath, planContent, requestId }`
- `PlanApprovalResponseMessage` — `{ type: 'plan_approval_response', requestId, approved, feedback?, timestamp, permissionMode? }`
- `ShutdownRequestMessage` — `{ type: 'shutdown_request', requestId, from, reason?, timestamp }`
- `ShutdownApprovedMessage` — `{ type: 'shutdown_approved', requestId, from, timestamp, paneId?, backendType? }`
- `ShutdownRejectedMessage` — `{ type: 'shutdown_rejected', requestId, from, reason, timestamp }`
- `PermissionRequestMessage` / `PermissionResponseMessage` — worker tool permission flow
- `SandboxPermissionRequestMessage` / `SandboxPermissionResponseMessage` — network host allowlisting

Evidence: utils/teammateMailbox.ts:682-860.

**Poller**: `hooks/useSwarmPermissionPoller.ts` (evidenced by import in inProcessRunner.ts:17-22) — teammates poll their mailbox on a timer; on receipt of structured messages (shutdown_request, permission_response, etc.), they route through `processMailboxPermissionResponse` or similar handlers.

---

## 7. Isolation & Background Execution

### 7.1 Three Teammate Backends (`utils/swarm/backends/`)

| Backend | Implementation | Isolation |
|---|---|---|
| `tmux` | TmuxBackend.ts (registered dynamically via `registerTmuxBackend`) | Separate OS process in a tmux pane |
| `iterm2` | ITermBackend.ts (via `it2` CLI) | Separate OS process in iTerm2 native split |
| `in-process` | InProcessBackend.ts | Same Node.js process, AsyncLocalStorage for identity isolation |

Detection priority (registry.ts:128-254):
1. Inside tmux → always `tmux`
2. Inside iTerm2 with it2 CLI → `iterm2` (unless user preferTmux)
3. Inside iTerm2 without it2 + tmux available → fallback tmux (+ flag `needsIt2Setup: true` unless preferTmux)
4. Outside both + tmux available → external tmux session (`claude-swarm-{pid}` socket)
5. Nothing → throw error with install instructions

`CLAUDE_CODE_TEAMMATE_MODE` env var (resolved by `getTeammateModeFromSnapshot()`): `'auto' | 'tmux' | 'in-process'`. In non-interactive sessions, always `in-process`.

### 7.2 In-Process Teammate (`InProcessTeammateTask`)

`tasks/InProcessTeammateTask/types.ts`:
```ts
InProcessTeammateTaskState {
  type: 'in_process_teammate',
  identity: { agentId, agentName, teamName, color?, planModeRequired, parentSessionId },
  prompt, model?, selectedAgent?,
  abortController?, currentWorkAbortController?,
  awaitingPlanApproval, permissionMode,
  messages?: Message[] (capped at 50),
  inProgressToolUseIDs?, pendingUserMessages,
  isIdle, shutdownRequested,
  onIdleCallbacks?,
  lastReportedToolCount, lastReportedTokenCount,
  ...
}
```

**Context isolation** happens via `runWithTeammateContext(context, () => ...)` — an AsyncLocalStorage provides: `getAgentName()`, `getAgentId()`, `getTeamName()`, `isTeammate()`, `getTeammateColor()`, `isInProcessTeammate()`.

Evidence: utils/teammateContext.ts (referenced by swarm/inProcessRunner.ts:88-89).

**Message cap**: `TEAMMATE_MESSAGES_UI_CAP = 50` (task.messages for transcript UI). Full conversation on disk at `~/.claude/projects/{hash}/` sidechain transcript.

Rationale for cap: BigQuery analysis showed ~20MB RSS per agent at 500+ turns; 292-agent whale session reached 36.8GB (types.ts:92-99).

### 7.3 Tmux Teammate Spawn (`tools/shared/spawnMultiAgent.ts`)

Two spawn modes:

**Split-pane** (default): `handleSpawnSplitPane` (spawnMultiAgent.ts:305-539):
- Inside tmux: splits current window (leader left, teammates right)
- iTerm2 with it2: native iTerm2 split
- External: creates `claude-swarm-{pid}` tmux session with tiled teammates

**Separate window** (legacy): `handleSpawnSeparateWindow` (spawnMultiAgent.ts:545+):
- Each teammate gets own tmux window

Spawn command assembly:
```
cd {workingDir} && env {inherited-env} {binaryPath} \
  --agent-id {teammateId} --agent-name {sanitizedName} \
  --team-name {teamName} --agent-color {color} \
  --parent-session-id {leaderSessionId} \
  [--plan-mode-required] [--agent-type {type}] \
  [--dangerously-skip-permissions | --permission-mode {mode}] \
  [--model {model}] [--settings {path}] [--plugin-dir ...] \
  [--chrome | --no-chrome]
```

Evidence: spawnMultiAgent.ts:403-440, 614-650.

**Initial instructions**: after spawn, the leader writes the prompt to the teammate's mailbox via `writeToMailbox(sanitizedName, { from: 'team-lead', text: prompt, ... }, teamName)`. The teammate's inbox poller picks it up and submits as its first turn.

### 7.4 Background Agent Execution

Forced async when any of (AgentTool.tsx:567):
- `run_in_background === true`
- `selectedAgent.background === true` (frontmatter)
- `isCoordinator` (coordinator mode)
- `forceAsync` (fork subagent gate)
- `assistantForceAsync` (KAIROS feature)
- `proactiveModule?.isProactiveActive()`

AND NOT:
- `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS` truthy

When async:
- `registerAsyncAgent({ agentId, description, prompt, selectedAgent, setAppState, toolUseId })` returns `{ agentId, abortController }`
- `void runWithAgentContext(asyncAgentContext, () => wrapWithCwd(() => runAsyncAgentLifecycle({ taskId, abortController, makeStream: runAgent(...), ... })))` — detached
- Parent's abort does NOT cancel async (AgentTool.tsx:694: "background agents should survive when user presses ESC")

Auto-background timer (AgentTool.tsx:72-77): sync sub-agents auto-background after 2 minutes if `CLAUDE_AUTO_BACKGROUND_TASKS` truthy OR GrowthBook `tengu_auto_background_agents` (currently 120_000ms).

### 7.5 Worktree Isolation

`isolation: 'worktree'` triggers `createAgentWorktree(slug)` (utils/worktree.ts, referenced at AgentTool.tsx:42). Result:
```
{ worktreePath, worktreeBranch?, headCommit?, gitRoot?, hookBased? }
```

Cleanup logic (AgentTool.tsx:644-685):
- `hookBased: true` → always keep (VCS changes undetectable)
- If `headCommit` matches and no changes → remove worktree + clear metadata
- Else → keep (emit `worktreePath + worktreeBranch` in result for caller)

Cleanup is idempotent — `worktreeInfo = null` guard prevents double-call.

---

## 8. Gap vs moai-adk (CC HAS / moai-adk LACKS)

### 8.1 Major features CC has but moai-adk lacks

| Feature | CC Location | Impact |
|---|---|---|
| **Built-in agent types** `Explore`, `Plan`, `Code Guide`, `Verification`, `Statusline Setup` | tools/AgentTool/built-in/ | moai has `Explore` and `Plan` as concepts but no canonical built-in definitions shipped with the Go binary |
| **`fork` subagent** | tools/AgentTool/forkSubagent.ts | Omit `subagent_type` → child inherits parent's full conversation context + system prompt (byte-identical for cache sharing). moai has nothing equivalent |
| **Agent Memory** (`memory: user|project|local`) | tools/AgentTool/agentMemory.ts, agentMemorySnapshot.ts | Agent-scoped persistent memory with snapshot-based project→user initialization. moai has `auto-memory/lessons.md` but no agent-type-scoped memory |
| **Agent `hooks:` frontmatter** | schemas/hooks.ts + loadAgentsDir.ts:711 | Per-agent session-scoped hooks; 4 hook types including `agent` (LLM verifier) |
| **Agent `requiredMcpServers`** | loadAgentsDir.ts:122 | Agent unavailable if required MCP server not connected. moai has no such gating |
| **Agent `initialPrompt`** | loadAgentsDir.ts:91 | Prepended to first user turn; slash commands work. moai has no direct equivalent |
| **Agent `criticalSystemReminder_EXPERIMENTAL`** | loadAgentsDir.ts:121 | Re-injected at every user turn |
| **Agent `maxTurns`** | loadAgentsDir.ts:649 | Hard cap on agentic turns |
| **Agent `skills:` preload** | loadAgentsDir.ts:684 | Comma-separated skill names auto-included in tool pool |
| **Agent `omitClaudeMd`** | loadAgentsDir.ts:128-132 | Saves ~5-15 Gtok/week across 34M+ Explore spawns. Applied to Explore + Plan |
| **Agent `permissionMode: bubble`** | forkSubagent.ts:67, PERMISSION_MODES | Fork-specific mode: permission prompts bubble to parent terminal |
| **Agent `permissionMode: auto`** | PERMISSION_MODES | Classifier-based auto-approval of tool calls |
| **Task Framework registry** (7 types) | Task.ts, tasks.ts | Unified kill semantics across bash, agent, workflow, dream, monitor. moai has specific task handling per domain but no unified Task abstraction |
| **TaskStopTool** | tools/TaskStopTool/ | LLM-invokable stop with deprecated `KillShell` alias. moai has no equivalent tool |
| **In-process teammate backend** | utils/swarm/backends/InProcessBackend.ts + inProcessRunner.ts | Single-process teammates via AsyncLocalStorage. moai is tmux-only — no in-process alternative |
| **iTerm2 native-split backend** | utils/swarm/backends/ITermBackend.ts | Via `it2` CLI. moai is tmux-only |
| **teammateMailbox schemas** | utils/teammateMailbox.ts:682-860 | 7 Zod-validated structured message types (shutdown_request, shutdown_approved, shutdown_rejected, plan_approval_request, plan_approval_response, permission_request, permission_response, sandbox_permission_request, sandbox_permission_response). moai uses ad-hoc JSON |
| **Plan-approval flow** | SendMessageTool.ts:434-518 | Team lead approves/rejects teammate plans via structured message; rejection carries `feedback` |
| **Per-teammate permission mode** (`TeamFile.members[].mode`) | utils/swarm/teamHelpers.ts:85-89 | Each teammate has independent cyclable permission mode |
| **Team-allowed paths** (`TeamFile.teamAllowedPaths`) | utils/swarm/teamHelpers.ts:57-63 | Directory allowlist all teammates can edit without asking |
| **`TEAMMATE_MESSAGES_UI_CAP = 50`** | InProcessTeammateTask/types.ts:101 | Message history cap to prevent 36GB whale sessions |
| **Word-slug team-name collision** | tools/TeamCreateTool/TeamCreateTool.ts:64-72 | Duplicate team name auto-generates new slug via `generateWordSlug()` |
| **Session cleanup registry** | utils/swarm/teamHelpers.ts:560-590 | Teams auto-cleaned on SIGINT; teammate panes force-killed first |
| **`auto-background` timer** | AgentTool.tsx:72-77, 819-832 | Sync agents auto-background after 2 min (gated) |
| **Agent name registry** | AgentTool.tsx:700-712 | `appState.agentNameRegistry` maps `name` → `agentId` for SendMessage routing. Stopped agents auto-resume via `resumeAgentBackground` |
| **Resume agent from transcript** | tools/AgentTool/resumeAgent.ts + SendMessageTool.ts:822-872 | Send message to evicted/stopped agent → reads disk transcript → resumes in background |
| **Coordinator mode** | coordinator/coordinatorMode.ts | 380-line coordinator system prompt with task-notification XML protocol |
| **Remote agent isolation** (`isolation: 'remote'`) | AgentTool.tsx:435-482 | ant-only; teleports agent to Anthropic CCR container |
| **Remote Control bridge** (`bridge/`) | All 33 files | Control local Claude from web/another CLI via Anthropic cloud |
| **Cross-session UDS messaging** (`uds:/path/to.sock`) | SendMessageTool.ts:775-798 | Same-machine inter-session messaging via Unix domain socket |
| **Plan-approval schema** `permissionMode` inheritance | SendMessageTool.ts:448-457 | Team lead's `plan` mode maps to teammate's `default` mode on approval |
| **Agent `isConcurrencySafe`** flag | TaskListTool.ts:57, TaskGetTool.ts:62, TaskUpdateTool.ts:111 | Tool flag marking tools that can be called in parallel |
| **Task `metadata._internal` filter** | TaskListTool.ts:68 | Hidden internal-use tasks |

### 8.2 Schemas CC has but moai-adk lacks

- **`StructuredMessage` discriminated union** in SendMessageTool (3 types: shutdown_request, shutdown_response, plan_approval_response)
- **Mailbox message schemas** (10+ types via Zod)
- **`SDKControlRequest`/`SDKControlResponse`** for control-channel messages (initialize, interrupt, set_model, set_permission_mode, set_max_thinking_tokens, can_use_tool)
- **Agent hooks schema** (4 hook types, lazy-evaluated)
- **`TeamFile` schema** with typed `members[]` array including `isActive`, `mode`, `backendType`

### 8.3 Orchestration patterns CC has but moai-adk lacks

- Parent → sub-agent returns `<task-notification>` XML injected as user-role message (async lifecycle)
- Teammate → team lead goes idle → automatic idle notification with token delta + peer DM summary
- SendMessage auto-routes: teammate-name → mailbox; known in-process agent name → queue pending message; stopped agent → auto-resume
- Fork subagent: cache-identical API prefix through `buildForkedMessages` — all forks share parent's prompt cache

### 8.4 Items moai-adk has that CC does NOT

| Feature | moai-adk | CC |
|---|---|---|
| **SPEC-based workflow** (Plan-Run-Sync pipeline) | Yes, core | Not a primitive |
| **TRUST 5 framework** | Yes | No |
| **MX code annotations** (@MX:NOTE/WARN/ANCHOR/TODO) | Yes | No |
| **CG Mode** (Claude leader + GLM teammates) | Yes, tmux-based | No (single-provider) |
| **Worktree persistent per-SPEC workspaces** | `~/.moai/worktrees/{Project}/{SPEC}/` | CC has ephemeral `.claude/worktrees/` only |
| **Project-level docs** (product.md, structure.md, tech.md) | Yes | No |
| **16-language toolchain detection** (go/python/ts/js/rust/java/kotlin/csharp/ruby/php/elixir/cpp/scala/r/flutter/swift) | Yes | CC is language-agnostic |
| **`/moai` unified slash command router** | Yes | CC has separate commands |
| **Evolvable/Frozen zone Design System constitution** | Yes (design/constitution.md) | No |

---

## 9. Source References (file:line)

### Bridge / Remote Control

- bridge/types.ts:18-115 — BridgeConfig, WorkData, WorkResponse, WorkSecret schemas
- bridge/types.ts:133-176 — BridgeApiClient interface (registerBridgeEnvironment, pollForWork, acknowledgeWork, stopWork, deregisterEnvironment, sendPermissionResponseEvent, archiveSession, reconnectSession, heartbeatWork)
- bridge/bridgeApi.ts:68-197 — createBridgeApiClient factory
- bridge/bridgeApi.ts:106-139 — withOAuthRetry (401 refresh logic)
- bridge/bridgeMessaging.ts:132-208 — handleIngressMessage (routes ingress to SDKMessage/control_response/control_request handlers)
- bridge/bridgeMessaging.ts:243-391 — handleServerControlRequest (responds to initialize/set_model/set_max_thinking_tokens/set_permission_mode/interrupt)
- bridge/bridgeMessaging.ts:429-461 — BoundedUUIDSet for echo/re-delivery dedup
- bridge/createSession.ts:33-180 — createBridgeSession (POST /v1/sessions)
- bridge/createSession.ts:263-317 — archiveBridgeSession
- bridge/remoteBridgeCore.ts:140-250 — initEnvLessBridgeCore (env-less path)
- bridge/replBridgeTransport.ts:23-70 — ReplBridgeTransport interface
- bridge/replBridgeTransport.ts:78-103 — createV1ReplTransport (HybridTransport wrapper)
- bridge/bridgePointer.ts:40-202 — Crash recovery pointer (4h TTL)
- remote/RemoteSessionManager.ts:95-324 — RemoteSessionManager class
- remote/RemoteSessionManager.ts:329-343 — createRemoteSessionConfig factory
- remote/SessionsWebSocket.ts:17-27 — connection constants
- remote/SessionsWebSocket.ts:82-404 — SessionsWebSocket class
- remote/sdkMessageAdapter.ts:145-278 — convertSDKMessage (CCR SDKMessage → REPL Message)
- remote/remotePermissionBridge.ts:12-78 — createSyntheticAssistantMessage, createToolStub

### Coordinator

- coordinator/coordinatorMode.ts:29-34 — INTERNAL_WORKER_TOOLS excludes TeamCreate/TeamDelete/SendMessage from worker tools list
- coordinator/coordinatorMode.ts:36-41 — isCoordinatorMode gate
- coordinator/coordinatorMode.ts:80-109 — getCoordinatorUserContext (worker tools description)
- coordinator/coordinatorMode.ts:111-369 — getCoordinatorSystemPrompt (full 380-line coordinator prompt)

### AgentTool / Sub-agent Dispatch

- tools/AgentTool/AgentTool.tsx:82-125 — full input schema (baseInputSchema + multiAgentInputSchema + isolation)
- tools/AgentTool/AgentTool.tsx:141-156 — outputSchema (sync vs async_launched)
- tools/AgentTool/AgentTool.tsx:161-191 — private InternalOutput types (TeammateSpawnedOutput, RemoteLaunchedOutput)
- tools/AgentTool/AgentTool.tsx:239-280 — dispatch entry (team/teammate checks)
- tools/AgentTool/AgentTool.tsx:282-316 — spawnTeammate path
- tools/AgentTool/AgentTool.tsx:320-356 — fork vs subagent_type resolution
- tools/AgentTool/AgentTool.tsx:435-482 — remote isolation branch (ant-only)
- tools/AgentTool/AgentTool.tsx:567 — shouldRunAsync computation
- tools/AgentTool/AgentTool.tsx:573-577 — workerPermissionContext + assembleToolPool
- tools/AgentTool/AgentTool.tsx:590-593 — createAgentWorktree on isolation:worktree
- tools/AgentTool/AgentTool.tsx:686-764 — async path (registerAsyncAgent)
- tools/AgentTool/AgentTool.tsx:765-1200 — sync path (with auto-backgrounding mid-flight)
- tools/AgentTool/built-in/generalPurposeAgent.ts:25-34 — GENERAL_PURPOSE_AGENT definition
- tools/AgentTool/built-in/exploreAgent.ts:64-83 — EXPLORE_AGENT (disallowedTools: Agent, ExitPlanMode, FileEdit, FileWrite, NotebookEdit; model: ant→inherit, ext→haiku; omitClaudeMd: true)
- tools/AgentTool/built-in/planAgent.ts:73-92 — PLAN_AGENT (same disallowedTools; model: inherit; omitClaudeMd: true)
- tools/AgentTool/builtInAgents.ts:22-72 — getBuiltInAgents (conditional registration)
- tools/AgentTool/loadAgentsDir.ts:73-99 — AgentJsonSchema (Zod)
- tools/AgentTool/loadAgentsDir.ts:106-133 — BaseAgentDefinition type
- tools/AgentTool/loadAgentsDir.ts:296-393 — getAgentDefinitionsWithOverrides (memoized loader)
- tools/AgentTool/loadAgentsDir.ts:445-516 — parseAgentFromJson
- tools/AgentTool/loadAgentsDir.ts:541-755 — parseAgentFromMarkdown
- tools/AgentTool/forkSubagent.ts:32-38 — isForkSubagentEnabled (FORK_SUBAGENT gate; mutually exclusive with coordinator + non-interactive)
- tools/AgentTool/forkSubagent.ts:60-71 — FORK_AGENT synthetic definition
- tools/AgentTool/forkSubagent.ts:73-89 — isInForkChild (recursive-fork guard via FORK_BOILERPLATE_TAG)
- tools/AgentTool/prompt.ts:66-287 — full Agent tool prompt generator

### Team Coordination

- tools/TeamCreateTool/TeamCreateTool.ts:37-49 — TeamCreate input schema
- tools/TeamCreateTool/TeamCreateTool.ts:64-72 — generateUniqueTeamName
- tools/TeamCreateTool/TeamCreateTool.ts:74-240 — TeamCreate implementation
- tools/TeamCreateTool/prompt.ts:1-113 — Full TeamCreate prompt doc (team workflow, task ownership, automatic message delivery, idle state, discovering team members)
- tools/TeamDeleteTool/TeamDeleteTool.ts:32-139 — TeamDelete implementation (active-member check, cleanupTeamDirectories, clear AppState)
- tools/SendMessageTool/SendMessageTool.ts:46-65 — StructuredMessage discriminated union schema
- tools/SendMessageTool/SendMessageTool.ts:67-87 — SendMessage input schema (feature('UDS_INBOX') gates `uds:`/`bridge:` schemes)
- tools/SendMessageTool/SendMessageTool.ts:149-189 — handleMessage (writeToMailbox)
- tools/SendMessageTool/SendMessageTool.ts:191-266 — handleBroadcast
- tools/SendMessageTool/SendMessageTool.ts:268-399 — handleShutdownRequest, handleShutdownApproval, handleShutdownRejection
- tools/SendMessageTool/SendMessageTool.ts:434-518 — handlePlanApproval, handlePlanRejection
- tools/SendMessageTool/SendMessageTool.ts:585-602 — checkPermissions (bridge: ask-mode)
- tools/SendMessageTool/SendMessageTool.ts:604-718 — validateInput
- tools/SendMessageTool/SendMessageTool.ts:741-913 — main call dispatch

### Task* Tools

- tools/TaskCreateTool/TaskCreateTool.ts:18-138 — TaskCreate (TaskCreated hooks gate)
- tools/TaskUpdateTool/TaskUpdateTool.ts:88-406 — TaskUpdate (TaskCompleted hooks, auto-owner, mailbox task_assignment)
- tools/TaskListTool/TaskListTool.ts:33-116 — TaskList (blockedBy filter)
- tools/TaskGetTool/TaskGetTool.ts:38-127 — TaskGet
- tools/TaskStopTool/TaskStopTool.ts:39-131 — TaskStop (deprecated KillShell alias)
- tasks/stopTask.ts:38-100 — stopTask shared logic
- Task.ts:6-29 — TaskType, TaskStatus, isTerminalTaskStatus
- Task.ts:44-57 — TaskStateBase
- Task.ts:72-76 — Task interface (kill only)
- Task.ts:79-106 — TaskId alphabet + generateTaskId
- tasks.ts:22-39 — getAllTasks, getTaskByType

### Tasks / Backends

- tasks/InProcessTeammateTask/types.ts:22-76 — InProcessTeammateTaskState
- tasks/InProcessTeammateTask/types.ts:92-121 — appendCappedMessage + TEAMMATE_MESSAGES_UI_CAP
- tasks/InProcessTeammateTask/InProcessTeammateTask.tsx:24-30 — Task impl
- tasks/InProcessTeammateTask/InProcessTeammateTask.tsx:34-125 — requestTeammateShutdown, appendTeammateMessage, injectUserMessageToTeammate, findTeammateTaskByAgentId, getAllInProcessTeammateTasks, getRunningTeammatesSorted
- tasks/LocalAgentTask/LocalAgentTask.tsx:23-150 — AgentProgress + LocalAgentTaskState
- tasks/LocalAgentTask/LocalAgentTask.tsx:197+ — enqueueAgentNotification, queuePendingMessage, drainPendingMessages, etc.

### Swarm / Utility

- utils/agentSwarmsEnabled.ts:7-44 — full isAgentSwarmsEnabled gate
- utils/swarm/constants.ts:1-34 — TEAM_LEAD_NAME, SWARM_SESSION_NAME, TMUX_COMMAND, env vars
- utils/swarm/teamHelpers.ts:19-42 — input schema for spawnTeam/cleanup operations
- utils/swarm/teamHelpers.ts:64-90 — TeamFile + TeamAllowedPath schemas
- utils/swarm/teamHelpers.ts:100-112 — sanitizeName, sanitizeAgentName
- utils/swarm/teamHelpers.ts:131-182 — readTeamFile/writeTeamFile (sync + async)
- utils/swarm/teamHelpers.ts:188-227 — removeTeammateFromTeamFile
- utils/swarm/teamHelpers.ts:260-348 — removeMemberFromTeam, removeMemberByAgentId
- utils/swarm/teamHelpers.ts:357-445 — setMemberMode, syncTeammateMode, setMultipleMemberModes
- utils/swarm/teamHelpers.ts:454-485 — setMemberActive
- utils/swarm/teamHelpers.ts:560-590 — registerTeamForSessionCleanup, cleanupSessionTeams
- utils/swarm/teamHelpers.ts:598-683 — killOrphanedTeammatePanes, cleanupTeamDirectories
- utils/swarm/backends/types.ts:9-312 — BackendType, PaneBackend, TeammateExecutor interfaces
- utils/swarm/backends/registry.ts:74-254 — ensureBackendsRegistered, detectAndGetBackend
- utils/swarm/backends/registry.ts:335-390 — isInProcessEnabled
- utils/swarm/backends/registry.ts:396-451 — getTeammateExecutor, getPaneBackendExecutor
- utils/teammateMailbox.ts:43-50 — TeammateMessage type
- utils/teammateMailbox.ts:56-66 — getInboxPath
- utils/teammateMailbox.ts:84-131 — readMailbox, readUnreadMessages
- utils/teammateMailbox.ts:134-192 — writeToMailbox (file-locked)
- utils/teammateMailbox.ts:201-271 — markMessageAsReadByIndex
- utils/teammateMailbox.ts:279-340 — markMessagesAsRead
- utils/teammateMailbox.ts:488-536 — createPermissionRequest/ResponseMessage
- utils/teammateMailbox.ts:576-678 — SandboxPermissionRequest/Response
- utils/teammateMailbox.ts:682-767 — PlanApprovalRequest/Response schemas, ShutdownRequest/Approved/Rejected schemas
- utils/teammateMailbox.ts:771-820 — createShutdownRequestMessage, createShutdownApprovedMessage, createShutdownRejectedMessage
- utils/teammateMailbox.ts:831-897 — sendShutdownRequestToMailbox, isShutdownRequest, isPlanApprovalRequest
- utils/swarm/inProcessRunner.ts:1-150 — in-process teammate runner, permission callbacks

### Shared spawn logic

- tools/shared/spawnMultiAgent.ts:72-82 — getDefaultTeammateModel
- tools/shared/spawnMultiAgent.ts:93-101 — resolveTeammateModel (`inherit` → leader)
- tools/shared/spawnMultiAgent.ts:193-197 — getTeammateCommand
- tools/shared/spawnMultiAgent.ts:208-260 — buildInheritedCliFlags
- tools/shared/spawnMultiAgent.ts:305-539 — handleSpawnSplitPane
- tools/shared/spawnMultiAgent.ts:545-705+ — handleSpawnSeparateWindow

### XML Constants

- constants/xml.ts:28-38 — TASK_NOTIFICATION_TAG, TASK_ID_TAG, TOOL_USE_ID_TAG, TASK_TYPE_TAG, OUTPUT_FILE_TAG, STATUS_TAG, SUMMARY_TAG, REASON_TAG, WORKTREE_TAG, WORKTREE_PATH_TAG, WORKTREE_BRANCH_TAG
- constants/xml.ts:52-55 — TEAMMATE_MESSAGE_TAG, CHANNEL_MESSAGE_TAG, CHANNEL_TAG
- constants/xml.ts:59 — CROSS_SESSION_MESSAGE_TAG (UDS_INBOX)
- constants/xml.ts:62-66 — FORK_BOILERPLATE_TAG, FORK_DIRECTIVE_PREFIX

### Hooks Schema

- schemas/hooks.ts:11 — imports HOOK_EVENTS from entrypoints/agentSdkTypes.js
- schemas/hooks.ts:19-27 — IfConditionSchema (permission-rule syntax)
- schemas/hooks.ts:31-171 — buildHookSchemas (BashCommand + Prompt + Http + Agent)
- schemas/hooks.ts:176-213 — HookCommandSchema (discriminated union on `type`), HookMatcherSchema, HooksSchema

### Buddy (out of scope)

- buddy/types.ts:1-148 — Rarity/Species/Eye/Hat/Stat types
- buddy/companion.ts:106-133 — roll/rollWithSeed, companionUserId, getCompanion

### native-ts (out of scope)

- native-ts/color-diff/index.ts:1-50 — pure-TS port rationale (replaces syntect/bat NAPI)
- native-ts/file-index/index.ts:1-50 — pure-TS fuzzy file search (replaces nucleo NAPI)
- native-ts/yoga-layout/index.ts — pure-TS layout engine (replaces yoga NAPI)

### moreright (out of scope — stub)

- moreright/useMoreRight.tsx:1-25 — stub for external builds (real hook is internal-only)

### assistant (minimal)

- assistant/sessionHistory.ts:1-88 — viewer-only session history paging (`claude assistant` mode)

---

## 10. Follow-up Questions for Design

1. **Does moai v3 want a `fork` subagent primitive** (omit subagent_type → child inherits parent context)? CC treats this as a first-class citizen when `feature('FORK_SUBAGENT')` is on.

2. **Should moai adopt the `<task-notification>` XML envelope** for async agent completion messages? This is a load-bearing UX primitive in coordinator mode and async sub-agents.

3. **Is in-process teammate** (AsyncLocalStorage-based, same Node.js process) worth implementing in Go, given moai's tmux-first architecture? The Go equivalent would be goroutine-local context via `context.Context` — semantically equivalent.

4. **Should moai add `isolation: 'remote'`** for delegating SPECs to remote CCR-like infrastructure? Currently ant-only in CC.

5. **Should moai expose `TaskStop` as an LLM-invokable tool**? Today moai has tmux pane-kill via admin paths; CC exposes `TaskStop` directly to the model.

6. **Do we need the 10+ Zod-validated mailbox message types**, or is ad-hoc JSON sufficient? CC uses strict schemas (shutdown_request, shutdown_approved, shutdown_rejected, plan_approval_request/response, permission_request/response, sandbox_permission_request/response, task_assignment).

7. **Does moai want to adopt `agentNameRegistry` auto-resume**? SendMessage to a stopped/evicted agent triggers transcript-based resume in the background. Without it, dead agent references silently fail.

8. **Should moai ship built-in agent definitions** as Go-embedded resources (like Explore, Plan, general-purpose)? Current moai approach is user-land agent templates.

9. **Does moai want the `skills:` preload frontmatter field**? Lets agents bundle pre-required skills in their definition.

10. **Is `omitClaudeMd` worth implementing**? CC saves ~5-15 Gtok/week across 34M+ Explore spawns by excluding CLAUDE.md from read-only agent contexts.

11. **Adopt Claude Code bridge subsystem or not?** The bridge is tied to `api.anthropic.com` OAuth and cannot be reused for arbitrary transports. For moai's CG Mode (GLM+Claude), a purely local variant would need a different design.

12. **Permission-mode `bubble`** — does it fit moai's sub-agent model where fork subagents inherit parent system prompt and need parent-terminal permission prompts?

---

Version: 1.0
Generated: 2026-04-22
Source corpus: /Users/goos/MoAI/AgentOS/claude-code-source-map/ (snapshot 2026-04-10)
Research scope: bridge/, coordinator/, remote/, buddy/, native-ts/, assistant/, moreright/, schemas/, plus tools/AgentTool/, tools/Team*Tool/, tools/SendMessageTool/, tools/Task*Tool/, tasks/, utils/swarm/, utils/teammateMailbox.ts

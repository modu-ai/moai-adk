# Wave 1.2 — Claude Code Query/Context/Memory Inventory

Scope: QueryEngine + query loop, auto-compaction, context assembly, history/resume, cost tracker, memdir (memory), migrations, constants. Evidence pulled from the dumped TypeScript source at `/Users/goos/MoAI/AgentOS/claude-code-source-map/`. Every claim below anchors to a file:line.

---

## 1. Query Lifecycle

### 1.1 Two Entry Points

- `QueryEngine` class — stateful, multi-turn conversation owner. Defined at `QueryEngine.ts:184`.
- `ask()` generator — one-shot convenience wrapper at `QueryEngine.ts:1186`.
- Inside a single turn, `QueryEngine.submitMessage()` (`QueryEngine.ts:209`) calls `query()` (`query.ts:219`) which is a thin trampoline around `queryLoop()` (`query.ts:241`).

### 1.2 QueryEngine per-conversation state

Defined in `QueryEngine.ts:185-207`:

- `mutableMessages: Message[]` — full transcript (append-only within a turn; truncated on compact boundary at `QueryEngine.ts:928`).
- `abortController: AbortController` — interrupt handle, exposed via `.interrupt()` (`QueryEngine.ts:1158`).
- `permissionDenials: SDKPermissionDenial[]` — per-turn denial log for final result.
- `totalUsage: NonNullableUsage` — aggregated tokens.
- `readFileState: FileStateCache` — "what files has Claude read" cache (used by memory-dedup and file-change attachments).
- `discoveredSkillNames: Set<string>` — turn-scoped skill-discovery tracking (`QueryEngine.ts:197`).
- `loadedNestedMemoryPaths: Set<string>` — cross-turn nested CLAUDE.md memoization.
- `hasHandledOrphanedPermission` — one-shot flag for permission recovery (`QueryEngine.ts:398`).

### 1.3 submitMessage pipeline (per turn)

Sequential steps, each anchored to `QueryEngine.ts`:

1. **L238 clear discoveredSkillNames** — per-turn reset.
2. **L244-271 wrap canUseTool** — track denials for final result.
3. **L274-282 resolve model + thinking config** — `getMainLoopModel()` or user override; `shouldEnableThinkingByDefault()` returns `{ type: 'adaptive' }` or `{ type: 'disabled' }`.
4. **L288-308 fetch system prompt parts** via `fetchSystemPromptParts()` (`utils/queryContext.ts:44`). Returns `{ defaultSystemPrompt, userContext, systemContext }`.
5. **L316-325 assemble final system prompt** = `[customPrompt | defaultSystemPrompt, memoryMechanicsPrompt?, appendSystemPrompt?]`.
6. **L331-333 register structured-output hook** (when `jsonSchema` present).
7. **L335-395 build `processUserInputContext`** — first pass, mutable.
8. **L398-408 handle orphanedPermission** — once per engine lifetime.
9. **L410-428 processUserInput** — runs slash commands, expands `@` imports, returns `{ messages, shouldQuery, allowedTools, model, resultText }`.
10. **L450-463 persist transcript** — `recordTranscript(messages)`, optionally `flushSessionStorage()` when `CLAUDE_CODE_EAGER_FLUSH` or `CLAUDE_CODE_IS_COWORK`.
11. **L477-486 update permissions context** from user-input-allowed tools.
12. **L492-527 rebuild processUserInputContext** with updated model.
13. **L534-538 parallel skill+plugin cache load** — `Promise.all([getSlashCommandToolSkills, loadAllPluginsCacheOnly])`.
14. **L540-551 yield `buildSystemInitMessage`** — the first SDK message announcing tools/model/agents/skills/plugins.
15. **L556-638 local-command fast path** — if `!shouldQuery`, stream local outputs, write transcript, yield `result subtype=success`, return.
16. **L641-654 fileHistoryMakeSnapshot** — when `fileHistoryEnabled()` and persistSession true.
17. **L675-1049 main query loop** — `for await (const message of query(...))`. Each message type is dispatched through a giant switch (assistant, user, stream_event, tombstone, progress, attachment, system, tool_use_summary).
18. **L971-1002 per-iteration budget check** — `maxBudgetUsd` exhaustion ⇒ yield `error_max_budget_usd`.
19. **L1004-1048 structured-output retry limit** — `MAX_STRUCTURED_OUTPUT_RETRIES` env, default 5.
20. **L1058-1117 result selection** — `findLast assistant||user`, failure path emits `error_during_execution`.
21. **L1121-1154 success result** — extract text, yield terminal `result subtype=success`.

### 1.4 queryLoop (`query.ts:241-1728`)

Loop state (`query.ts:204-217`): `State = { messages, toolUseContext, autoCompactTracking, maxOutputTokensRecoveryCount, hasAttemptedReactiveCompact, maxOutputTokensOverride, pendingToolUseSummary, stopHookActive, turnCount, transition }`.

Per-iteration pipeline (all in `query.ts`):

1. **L301-304 startRelevantMemoryPrefetch** — starts async memory recall side-query; `using pendingMemoryPrefetch` disposes on any exit path.
2. **L331-335 startSkillDiscoveryPrefetch** — per-iteration skill prefetch.
3. **L365 getMessagesAfterCompactBoundary** — trim to post-compaction slice.
4. **L379-394 applyToolResultBudget** — enforce per-message tool-result budget (writes overflow to disk, replaces with previews).
5. **L401-410 HISTORY_SNIP** — `snipCompactIfNeeded()` — zombie-message removal; tracks `snipTokensFreed`.
6. **L414-426 microcompact** — lightweight compaction (`deps.microcompact`). If `CACHED_MICROCOMPACT`, defers boundary until API reports `cache_deleted_input_tokens`.
7. **L440-447 CONTEXT_COLLAPSE** — `contextCollapse.applyCollapsesIfNeeded()` — read-time projection of collapsed history.
8. **L449-451 appendSystemContext** — full system prompt = `[systemPrompt, ...systemContext keys]`.
9. **L453-467 autocompact** — `deps.autocompact()`; see §2 below.
10. **L470-543 post-compact handling** — log event, capture `taskBudgetRemaining`, reset tracking, build post-compact messages, yield them.
11. **L561-568 StreamingToolExecutor** — concurrent tool execution when `streamingToolExecution` gate on.
12. **L572-578 getRuntimeMainLoopModel** — with plan-mode 200k guard `doesMostRecentAssistantMessageExceed200k`.
13. **L588-590 createDumpPromptsFetch** — ant-only request-body dump; cached per-query session to avoid memory retention.
14. **L615-620 CONTEXT_COLLAPSE ownership check** — when collapse active AND autocompact enabled, collapse owns overflow handling; blocking limit check skipped.
15. **L628-648 hard blocking limit** — `calculateTokenWarningState` ⇒ if `isAtBlockingLimit`, yield `PROMPT_TOO_LONG_ERROR_MESSAGE` and return `{ reason: 'blocking_limit' }`.
16. **L653-954 model streaming** — `deps.callModel()` wraps `queryModelWithStreaming`. Handles fallback model via `FallbackTriggeredError`.
17. **L999-1009 executePostSamplingHooks** — fire-and-forget after model response.
18. **L1015-1052 abort handling** — drain StreamingToolExecutor, chicago MCP cleanup, yield interruption, return.
19. **L1062-1356 no-follow-up branch** — terminal path:
    - **L1085-1117 prompt-too-long (PTL) recovery** — collapse drain retry first.
    - **L1119-1183 reactiveCompact recovery** — post-PTL/media retry.
    - **L1185-1256 max_output_tokens recovery** — escalate to `ESCALATED_MAX_TOKENS=64_000` once, then 3 retry nudge messages (`MAX_OUTPUT_TOKENS_RECOVERY_LIMIT=3` at `query.ts:164`).
    - **L1262-1265 skip stop hooks on API error** — avoid death spiral.
    - **L1267-1306 handleStopHooks** — Stop/SubagentStop/TeammateIdle/TaskCompleted hooks; blocking errors produce continuation.
    - **L1308-1355 TOKEN_BUDGET** — `checkTokenBudget()` at 90% completion threshold; max 3 continuations then diminishing returns (see §8).
    - **L1357 return { reason: 'completed' }**.
20. **L1366-1408 tool execution** — `StreamingToolExecutor.getRemainingResults()` OR `runTools()`.
21. **L1415-1482 generateToolUseSummary** (Haiku, async, non-blocking). Surfaces in `claude ps`.
22. **L1567-1578 queue-command drain** — per-agent scoping, priority-based.
23. **L1580-1590 getAttachmentMessages** — inline injection of attachments (edited files, task-notifications, queued prompts).
24. **L1599-1614 memory prefetch consume** — awaits settled prefetch, dedups via `readFileState`.
25. **L1620-1628 skill prefetch consume**.
26. **L1685-1702 BG_SESSIONS task summary** — `maybeGenerateTaskSummary` for `claude ps`.
27. **L1705-1712 max turns check** — `maxTurns && nextTurnCount > maxTurns` ⇒ emit `max_turns_reached` attachment, return `{ reason: 'max_turns', turnCount }`.
28. **L1714-1727 state = next; continue**.

### 1.5 Tool invocation dispatch

- **Streaming path** (`StreamingToolExecutor`, gated on `tengu_streaming_tool_execution2` Statsig flag at `query/config.ts:33`) — kicks off tool calls as `tool_use` blocks arrive during streaming. Tool results consumed via `getCompletedResults()` (`query.ts:851`) and `getRemainingResults()`.
- **Batch path** — `runTools(toolUseBlocks, assistantMessages, canUseTool, toolUseContext)` at `query.ts:1382`. Tool permission check applied via `canUseTool` wrapper in `QueryEngine.ts:244-271`.
- `StreamingToolExecutor.discard()` called on fallback (`query.ts:734, 913`) to avoid orphan tool_results.

### 1.6 Return paths (`Terminal` reasons)

Enum reasons from `query.ts` return statements:

- `'blocking_limit'` (L646) — hard token ceiling hit without autocompact.
- `'aborted_streaming'` (L1051), `'aborted_tools'` (L1515).
- `'image_error'` (L977, L1175) — ImageSizeError / ImageResizeError.
- `'model_error'` (L996) — API/runtime error.
- `'prompt_too_long'` (L1175, L1182) — recovery exhausted.
- `'stop_hook_prevented'` (L1279), `'hook_stopped'` (L1520).
- `'completed'` (L1264, L1357) — normal terminal.
- `'max_turns'` (L1711, L1514 via aborted) — max-turns ceiling.

---

## 2. Auto-Compaction Algorithm (exact)

### 2.1 Thresholds (`services/compact/autoCompact.ts`)

- `AUTOCOMPACT_BUFFER_TOKENS = 13_000` (L62).
- `WARNING_THRESHOLD_BUFFER_TOKENS = 20_000` (L63).
- `ERROR_THRESHOLD_BUFFER_TOKENS = 20_000` (L64).
- `MANUAL_COMPACT_BUFFER_TOKENS = 3_000` (L65).
- `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES = 3` (L70) — circuit breaker. BQ 2026-03-10 rationale: "1,279 sessions had 50+ consecutive failures (up to 3,272), wasting ~250K API calls/day globally."
- `MAX_OUTPUT_TOKENS_FOR_SUMMARY = 20_000` (L30) — based on p99.99 compact summary output of 17,387 tokens.

### 2.2 Context window derivation

`getEffectiveContextWindowSize(model)` at `autoCompact.ts:33-49`:
```
effectiveWindow = contextWindow(model, betas) - min(maxOutputForModel, MAX_OUTPUT_TOKENS_FOR_SUMMARY)
```

Where `contextWindow` (`utils/context.ts:51-98`) resolves:
1. `CLAUDE_CODE_MAX_CONTEXT_TOKENS` env override (ant-only).
2. `[1m]` model suffix ⇒ `1_000_000`.
3. `cap.max_input_tokens` if ≥ 100_000 and not disabled.
4. `CONTEXT_1M_BETA_HEADER` beta + `modelSupports1M()`.
5. Sonnet 1M experiment.
6. Default `MODEL_CONTEXT_WINDOW_DEFAULT = 200_000` at `utils/context.ts:9`.

### 2.3 Threshold formula

`getAutoCompactThreshold(model)` at `autoCompact.ts:72-91`:
```
threshold = effectiveContextWindow - AUTOCOMPACT_BUFFER_TOKENS (13_000)
```
Testing override: `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` (0-100 percent).

### 2.4 Blocking-limit formula

`calculateTokenWarningState()` at `autoCompact.ts:93-145`:
```
blockingLimit = actualEffectiveContextWindow - MANUAL_COMPACT_BUFFER_TOKENS (3_000)
```
Overridable via `CLAUDE_CODE_BLOCKING_LIMIT_OVERRIDE`.

Returns `{ percentLeft, isAboveWarningThreshold, isAboveErrorThreshold, isAboveAutoCompactThreshold, isAtBlockingLimit }`.

### 2.5 Enable gating

`isAutoCompactEnabled()` at `autoCompact.ts:147-158`:
1. `DISABLE_COMPACT` env ⇒ false.
2. `DISABLE_AUTO_COMPACT` env ⇒ false.
3. `userConfig.autoCompactEnabled` from global config.

### 2.6 shouldAutoCompact decision (`autoCompact.ts:160-239`)

Guards:
- `querySource === 'session_memory' | 'compact'` ⇒ false (forked agent deadlock guard).
- `CONTEXT_COLLAPSE` + `marble_origami` querySource ⇒ false (avoid destroying main thread log).
- `REACTIVE_COMPACT` + `tengu_cobalt_raccoon` Statsig ON ⇒ false (reactive-only mode).
- `CONTEXT_COLLAPSE` + `isContextCollapseEnabled()` ⇒ false (collapse owns headroom).

Decision:
```
tokenCount = tokenCountWithEstimation(messages) - snipTokensFreed
return isAutoCompactEnabled() && tokenCount >= autoCompactThreshold
```

### 2.7 autoCompactIfNeeded (`autoCompact.ts:241-351`)

1. `DISABLE_COMPACT` env ⇒ skip.
2. Circuit breaker: `tracking.consecutiveFailures >= MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES (3)` ⇒ skip.
3. `shouldAutoCompact()` check.
4. **Session memory path first** (L287-310): `trySessionMemoryCompaction()` — experimental. Resets `lastSummarizedMessageId`, calls `runPostCompactCleanup`, `notifyCompaction` (cache-break detection), `markPostCompaction`.
5. **Legacy compaction fallback** (L312-333): `compactConversation(messages, toolUseContext, cacheSafeParams, suppressUserQuestions=true, undefined, isAutoCompact=true, recompactionInfo)`.
6. On failure: increment `consecutiveFailures` (L334-350); log when circuit trips.

### 2.8 Compaction strategies (layered)

In loop order (`query.ts:401-467`):

1. **Snip compaction** (HISTORY_SNIP feature) — removes zombie messages and stale markers. Runs first; reports `tokensFreed` to autocompact.
2. **Microcompact** (always) — cache-level trimming via `microcompactMessages()`. CACHED_MICROCOMPACT variant defers boundary message until API reports `cache_deleted_input_tokens`.
3. **Context collapse** (CONTEXT_COLLAPSE feature) — projection-based, commits incremental compactions; acts as 90/95% gate when enabled.
4. **Autocompact** (above) — proactive summary compaction.
5. **Reactive compact** (REACTIVE_COMPACT feature, `query.ts:1119-1166`) — triggered by API 413/media errors that the streaming loop withholds until recovery resolved. Single-shot (`hasAttemptedReactiveCompact` guard).
6. **Collapse drain retry** (`query.ts:1085-1117`) — drains staged context-collapse on withheld 413 before reactive compact.

---

## 3. Context Loading Order

### 3.1 System prompt assembly (`utils/queryContext.ts:44-74`)

`fetchSystemPromptParts()` returns in parallel via `Promise.all`:
- `defaultSystemPrompt: string[]` — from `constants/prompts.ts getSystemPrompt(tools, model, additionalDirs, mcpClients)`.
- `userContext: { [k: string]: string }` — from `context.ts getUserContext()`.
- `systemContext: { [k: string]: string }` — from `context.ts getSystemContext()`.

When `customSystemPrompt` is provided, `defaultSystemPrompt` and `systemContext` are skipped entirely (empty arrays/objects).

### 3.2 userContext (`context.ts:155-189`)

Memoized. Contents:
- `claudeMd` (CLAUDE.md tree) — from `getClaudeMds(filterInjectedMemoryFiles(await getMemoryFiles()))` at `context.ts:170-172`.
- `currentDate` — `Today's date is ${getLocalISODate()}.` at `context.ts:186`.

Disable conditions (`context.ts:165-167`):
- `CLAUDE_CODE_DISABLE_CLAUDE_MDS` env truthy.
- `isBareMode() && getAdditionalDirectoriesForClaudeMd().length === 0`.

Side effect: `setCachedClaudeMdContent(claudeMd || null)` at L176 — primes `yoloClassifier` without creating import cycle.

### 3.3 systemContext (`context.ts:116-150`)

Memoized. Contents:
- `gitStatus` (if git repo and not CCR and `shouldIncludeGitInstructions()`) — see §3.4.
- `cacheBreaker` (ant-only, `BREAK_CACHE_COMMAND` feature) — when injection set.

### 3.4 Git status injection (`context.ts:36-111`)

Parallel git commands: `getBranch()`, `getDefaultBranch()`, `git status --short`, `git log --oneline -n 5`, `git config user.name`.

`MAX_STATUS_CHARS = 2000` (`context.ts:20`). Truncated with notice: `"... (truncated because it exceeds 2k characters. If you need more information, run \"git status\" using BashTool)"`.

Output lines (L96-103):
```
This is the git status at the start of the conversation. Note that this status is a snapshot in time, and will not update during the conversation.
Current branch: {branch}
Main branch (you will usually use this for PRs): {mainBranch}
Git user: {userName}
Status:\n{truncatedStatus || '(clean)'}
Recent commits:\n{log}
```

### 3.5 Final system prompt layering (`QueryEngine.ts:321-325`)

```
systemPrompt = asSystemPrompt([
  ...(customPrompt ?? defaultSystemPrompt),
  ...(memoryMechanicsPrompt ? [memoryMechanicsPrompt] : []),
  ...(appendSystemPrompt ? [appendSystemPrompt] : []),
])
```

Memory-mechanics prompt is injected when `customPrompt` set AND `CLAUDE_COWORK_MEMORY_PATH_OVERRIDE` env present (`QueryEngine.ts:316-319`). This is the SDK caller "I wired up memory, now tell Claude how to use it" signal.

### 3.6 systemContext append timing (`query.ts:449-451`)

Inside the loop, every iteration:
```
const fullSystemPrompt = asSystemPrompt(appendSystemContext(systemPrompt, systemContext))
```

`userContext` is prepended to messages (not systemPrompt) via `prependUserContext(messagesForQuery, userContext)` at `query.ts:660`.

### 3.7 Cache-key prefix

The three pieces `(systemPrompt, userContext, systemContext)` form the API cache-key prefix — the comment at `utils/queryContext.ts:31` makes this explicit. Ordering is load-bearing for prompt caching.

### 3.8 Multi-directory support (`QueryEngine.ts:295-297`)

`additionalWorkingDirectories` sourced from `initialAppState.toolPermissionContext.additionalWorkingDirectories` (a Map).

---

## 4. History & Resume

### 4.1 Global prompt history (`history.ts`)

- Single file: `${CLAUDE_CONFIG_HOME}/history.jsonl` (L115, L299).
- Append-only JSONL, mode 0o600 (L319).
- Locked writes via `lock()` at L308 with 10s stale, 3 retries, 50ms min timeout.
- `MAX_HISTORY_ITEMS = 100` (L19) — window cap for reads.
- `MAX_PASTED_CONTENT_LENGTH = 1024` (L20) — inline threshold; large pastes go to paste store with hash reference.

### 4.2 LogEntry schema (`history.ts:219-225`)

```
{
  display: string,
  pastedContents: Record<number, StoredPastedContent>,
  timestamp: number,
  project: string,
  sessionId?: string,
}
```

StoredPastedContent (L25-32): `{ id, type: 'text'|'image', content?, contentHash?, mediaType?, filename? }`.

### 4.3 Reading ordering (`history.ts:190-217 getHistory`)

Current-session entries yielded first (newest-first), then other sessions. Windowed to `MAX_HISTORY_ITEMS`. Prevents concurrent-session interleave in Up-arrow UX.

### 4.4 Reference parsing (`history.ts:62-75`)

Reference pattern: `/\[(Pasted text|Image|\.\.\.Truncated text) #(\d+)(?: \+\d+ lines)?(\.)*\]/g`.

`formatPastedTextRef(id, numLines)` (L51-56): `[Pasted text #{id} +{numLines} lines]` or `[Pasted text #{id}]` when 0.

`expandPastedTextRefs()` (L81-100) inlines text (not images — they become content blocks) by splicing at match offsets in reverse order.

### 4.5 Undo support

`removeLastFromHistory()` at `history.ts:453-464`: fast path pops from pending buffer; slow path uses `skippedTimestamps: Set<number>` to filter at read time.

### 4.6 Flush lifecycle

`flushPromptHistory(retries)` at `history.ts:329-353`:
- `isWriting` lock prevents concurrent flush.
- Max 5 retries per burst (L335-337).
- 500ms backoff between retries (L347).
- `registerCleanup()` drains pending on process exit (L421-430).

### 4.7 Session transcript (separate from history)

Session transcripts at `${projectDir}/${sessionId}.jsonl` (`utils/sessionStorage.ts:202-205`).

- `MAX_TRANSCRIPT_READ_BYTES = 50 * 1024 * 1024` (50 MB) at `sessionStorage.ts:229`. Session JSONL can grow to GB; callers must bail out above this.
- `isTranscriptMessage(entry)` type-guard at `sessionStorage.ts:139`.
- `isChainParticipant(m)` at `sessionStorage.ts:154` — for parentUuid walks.
- `isEphemeralToolProgress(dataType)` at `sessionStorage.ts:194` — filters `bash_progress`, `powershell_progress`, `mcp_progress`, conditionally `sleep_progress`.
- Per-agent subdirs (`agentTranscriptSubdirs: Map`) at `sessionStorage.ts:234` — workflow runs grouped under `subagents/workflows/<runId>/`.
- `writeAgentMetadata` / `readAgentMetadata` / `writeRemoteAgentMetadata` at L283-346.

### 4.8 Recording in query loop (`QueryEngine.ts:450-463`)

User messages written BEFORE the query loop so `--resume` works even if the process is killed before API response:
- `persistSession && messagesFromUserInput.length > 0` ⇒ `await recordTranscript(messages)`.
- `isBareMode()` ⇒ fire-and-forget instead of await.
- `CLAUDE_CODE_EAGER_FLUSH` or `CLAUDE_CODE_IS_COWORK` ⇒ also `flushSessionStorage()` immediately.

Mid-turn writes (`QueryEngine.ts:687-735`):
- Assistant messages fire-and-forget (`void recordTranscript`) — write queue is order-preserving.
- User/compact_boundary messages `await recordTranscript`.
- Before compact boundary: writes through to `tailUuid` of `preservedSegment` so `applyPreservedSegmentRelinks` can walk tail→head on resume.

### 4.9 Compact boundary GC (`QueryEngine.ts:925-933`)

After `compact_boundary` yielded:
- `this.mutableMessages.splice(0, mutableBoundaryIdx)` — drop everything before the boundary for GC.
- `messages.splice(0, localBoundaryIdx)` — same for the turn-local snapshot.

---

## 5. Cost Tracker

### 5.1 Persistence (`cost-tracker.ts`)

Costs stored in project config (not settings). Key fields (L71-80):
```
StoredCostState = {
  totalCostUSD, totalAPIDuration, totalAPIDurationWithoutRetries,
  totalToolDuration, totalLinesAdded, totalLinesRemoved,
  lastDuration, modelUsage
}
```

### 5.2 Session restore (`cost-tracker.ts:87-137`)

`getStoredSessionCosts(sessionId)`: returns `undefined` unless `projectConfig.lastSessionId === sessionId` (L92-95). Rebuilds model usage by merging stored data with live `contextWindow`/`maxOutputTokens` (L99-110).

`restoreCostStateForSession(sessionId)` calls `setCostStateForRestore(data)` — **session-specific restore is the only supported path**.

### 5.3 Save point (`cost-tracker.ts:143-175`)

`saveCurrentSessionCosts(fpsMetrics?)` writes:
- `lastCost, lastAPIDuration, lastAPIDurationWithoutRetries, lastToolDuration, lastDuration`.
- `lastLinesAdded, lastLinesRemoved`.
- `lastTotalInputTokens, lastTotalOutputTokens`.
- `lastTotalCacheCreationInputTokens, lastTotalCacheReadInputTokens`.
- `lastTotalWebSearchRequests`.
- `lastFpsAverage, lastFpsLow1Pct` (UI perf metrics).
- `lastModelUsage: { [model]: { inputTokens, outputTokens, cacheReadInputTokens, cacheCreationInputTokens, webSearchRequests, costUSD } }`.
- `lastSessionId`.

### 5.4 Display formatting (`cost-tracker.ts:177-244`)

`formatCost(cost, 4)` — two-decimal `$X.XX` for > $0.50, else 4-decimal.
`formatModelUsage()` — accumulates per-canonical-model-name via `getCanonicalName(model)`.
`formatTotalCost()` — text block with `Total cost`, `Total duration (API)`, `Total duration (wall)`, `Total code changes`, `Usage by model`.

### 5.5 Per-request accounting (`cost-tracker.ts:278-323`)

`addToTotalSessionCost(cost, usage, model)`:
1. `addToTotalModelUsage(cost, usage, model)` → merges into per-model ModelUsage (L250-276):
   - `inputTokens, outputTokens, cacheReadInputTokens, cacheCreationInputTokens, webSearchRequests, costUSD, contextWindow, maxOutputTokens`.
2. OpenTelemetry counters: `getCostCounter()`, `getTokenCounter()` per token type (input/output/cacheRead/cacheCreation).
3. Fast-mode attribute: `{ model, speed: 'fast' }` when `isFastModeEnabled() && usage.speed === 'fast'` (L287-289).
4. **Advisor recursion** (L303-321): `getAdvisorUsage(usage)` returns advisor-model sub-usage; each gets `calculateUSDCost(advisorUsage.model, ...)` and recursively `addToTotalSessionCost`. Event logged as `tengu_advisor_tool_token_usage` with `cost_usd_micros: round(cost * 1_000_000)`.

### 5.6 USD calculation

`calculateUSDCost(model, usage)` lives in `utils/modelCost.ts` (not inspected here but referenced at `cost-tracker.ts:48, 305`). Table-based: per-model price per 1K tokens by token type (input, output, cache-read, cache-write).

### 5.7 Budget enforcement

**Hard budget**: `maxBudgetUsd` in `QueryEngineConfig`. Checked after every message inside the query loop at `QueryEngine.ts:972`:
```
if (maxBudgetUsd !== undefined && getTotalCost() >= maxBudgetUsd) {
  yield { type: 'result', subtype: 'error_max_budget_usd', ... }
  return
}
```

**Soft budget**: Token budget (§8) is a separate system with nudge messages, not cost-based.

---

## 6. Memory System (memdir)

### 6.1 Core concepts (`memdir/paths.ts`)

- `isAutoMemoryEnabled()` L30-55: priority chain:
  1. `CLAUDE_CODE_DISABLE_AUTO_MEMORY` env.
  2. `CLAUDE_CODE_SIMPLE` (`--bare`) ⇒ OFF.
  3. CCR without `CLAUDE_CODE_REMOTE_MEMORY_DIR` ⇒ OFF.
  4. `settings.autoMemoryEnabled`.
  5. Default ON.

- `isExtractModeActive()` L69-77: `tengu_passport_quail` Statsig + not non-interactive (or `tengu_slate_thimble`).

- `getMemoryBaseDir()` L85-90: `CLAUDE_CODE_REMOTE_MEMORY_DIR` env override, else `getClaudeConfigHomeDir()` (~/.claude).

- `AUTO_MEM_DIRNAME = 'memory'` (L92), `AUTO_MEM_ENTRYPOINT_NAME = 'MEMORY.md'` (L93).

### 6.2 Path resolution (`memdir/paths.ts:223-235 getAutoMemPath`)

Memoized on `getProjectRoot()`. Resolution order:
1. `CLAUDE_COWORK_MEMORY_PATH_OVERRIDE` env (full-path override, Cowork).
2. `autoMemoryDirectory` from settings (policy/flag/local/user sources, not projectSettings).
3. Default: `{memoryBase}/projects/{sanitizePath(getAutoMemBase())}/memory/`.

`getAutoMemBase()` (L203-205) prefers canonical git root via `findCanonicalGitRoot()` so worktrees share memory (gh-24382 rationale).

### 6.3 Path validation (`memdir/paths.ts:109-150 validateMemoryPath`)

SECURITY — rejects:
- Non-absolute paths.
- Length < 3 (root-like `"/"`, `"/a"`).
- Windows drive-root `C:\` regex.
- UNC paths `\\server\share`.
- Null byte `\0`.
- Tilde-only (`~/`, `~`, `~/..`) — would expand to $HOME or ancestor.
- NFKC Unicode normalization attacks (teamMemPaths.ts:43-54).

Always returns path with exactly one trailing separator, NFC-normalized.

### 6.4 Daily log path (`memdir/paths.ts:246-251`)

KAIROS (assistant mode) uses append-only daily log: `{autoMemPath}/logs/YYYY/MM/YYYY-MM-DD.md`. A nightly `/dream` skill distills these into topic files + MEMORY.md.

### 6.5 Entrypoint truncation (`memdir/memdir.ts`)

- `MAX_ENTRYPOINT_LINES = 200` (L35).
- `MAX_ENTRYPOINT_BYTES = 25_000` (L38) — "p97 today; catches long-line indexes that slip past the line cap (p100 observed: 197KB under 200 lines)".

`truncateEntrypointContent(raw)` at L57-103:
1. Line-truncate to 200 first (natural boundary).
2. Then byte-truncate at last newline before 25K.
3. Append warning: `"WARNING: MEMORY.md is {reason}. Only part of it was loaded. Keep index entries to one line under ~200 chars; move detail into topic files."`

### 6.6 Memory type taxonomy (`memdir/memoryTypes.ts:14-19`)

Fixed four types: `['user', 'feedback', 'project', 'reference']` — `MEMORY_TYPES` const tuple. `parseMemoryType(raw)` degrades gracefully for unknown/missing types.

Each type has full `<description>`, `<when_to_save>`, `<how_to_use>`, `<body_structure>`, `<examples>` XML sections (memoryTypes.ts:37-106 COMBINED, L113-* INDIVIDUAL).

Key distinction: **content derivable from project state (code, git, file structure) is explicitly excluded**. The comment at L1-12 makes this a constitutional rule.

### 6.7 Frontmatter format

Memory files are `.md` with YAML frontmatter:
```
---
name: ...
description: ...
type: user|feedback|project|reference
---
(body)
```

Stored in the auto-memory directory. MEMORY.md itself has no frontmatter — it's an index of `- [Title](file.md) — one-line hook` lines (see memdir.ts:227).

### 6.8 Save protocol (memdir.ts:219-234)

Two-step save:
1. Write memory content to its own file (e.g. `user_role.md`).
2. Add pointer line to MEMORY.md.

Skip-index variant (L205-217, when `tengu_moth_copse` Statsig on): step 2 omitted; only write the memory file. Used when MEMORY.md is maintained by a separate process (KAIROS daily-log mode).

### 6.9 System prompt integration (`memdir/memdir.ts:419-507 loadMemoryPrompt`)

Dispatches by feature flags:
- `KAIROS` + autoEnabled + `getKairosActive()` ⇒ daily-log prompt.
- `TEAMMEM` + `isTeamMemoryEnabled()` ⇒ combined prompt (auto + team dirs).
- Auto-only ⇒ `buildMemoryLines('auto memory', autoDir, extraGuidelines, skipIndex)`.

`ensureMemoryDirExists()` creates the directory before prompt build so the model can write without `mkdir` (L129-147 DIR_EXISTS_GUIDANCE).

### 6.10 Relevance selection (`memdir/findRelevantMemories.ts`)

`findRelevantMemories(query, memoryDir, signal, recentTools, alreadySurfaced)`:
1. `scanMemoryFiles(memoryDir, signal)` returns all `.md` headers except MEMORY.md.
2. Filters out `alreadySurfaced` paths.
3. Calls `selectRelevantMemories()` — a **Sonnet sideQuery** with system prompt:
   > "You are selecting memories that will be useful to Claude Code as it processes a user's query..." (L18-24). Returns up to 5 filenames.
4. Filters `recentTools` — excludes tools Claude is already exercising (noise reduction via keyword overlap prevention).
5. Returns `RelevantMemory[] = { path, mtimeMs }[]`.

The output shape + JSON schema enforced via `output_format: json_schema` with `selected_memories: string[]` (L109-120). Max 256 tokens.

### 6.11 Memory scan (`memdir/memoryScan.ts`)

- `MAX_MEMORY_FILES = 200` (L21) — cap on returned headers.
- `FRONTMATTER_MAX_LINES = 30` (L22) — read depth per file.
- Single-pass: `readFileInRange()` returns mtime so no separate stat (rationale L29-34: "halves syscalls vs stat-sort-read").
- Sorts newest-first by mtime, slices to MAX.
- `formatMemoryManifest(memories)` outputs `[type] filename (timestamp): description` lines for injection into selector prompt.

### 6.12 Memory age (`memdir/memoryAge.ts`)

- `memoryAgeDays(mtime)` L6-8: floor-rounded days, clamped to 0.
- `memoryAge(mtime)` L15-20: human string (`"today"`, `"yesterday"`, `"{N} days ago"`).
- `memoryFreshnessText(mtime)` L33-42: returns staleness caveat for memories > 1 day old:
  > "This memory is {d} days old. Memories are point-in-time observations, not live state — claims about code behavior or file:line citations may be outdated. Verify against current code before asserting as fact."
- `memoryFreshnessNote(mtime)` L49-53: wraps in `<system-reminder>` tags.

Rationale at L27-30: "user reports of stale code-state memories being asserted as fact — citation makes the stale claim sound more authoritative, not less."

### 6.13 Team memory (`memdir/teamMemPaths.ts`)

- Always a subdirectory of auto memory: `{autoMemPath}/team/` (L84-86).
- Same entrypoint (`MEMORY.md`) but in team dir.
- Gated on `tengu_herring_clock` Statsig + auto memory enabled (L73-78).
- Extensive path-key sanitization (L22-64): null bytes, URL-encoded traversals, Unicode NFKC attacks, backslashes, absolute paths all rejected via `PathTraversalError`.

### 6.14 Prefetch integration

In query loop (`query.ts:301-304`):
```
using pendingMemoryPrefetch = startRelevantMemoryPrefetch(state.messages, state.toolUseContext)
```
- Starts concurrently with model streaming (5-30s window) so the Sonnet selector call (~1s) is hidden.
- `consumedOnIteration` tracks whether already consumed (`query.ts:1602-1614`).
- `filterDuplicateMemoryAttachments` dedups against `readFileState` (files Claude already Read/Wrote/Edited in any iteration).

---

## 7. Migration Framework

### 7.1 Runner (`main.tsx:325-352`)

- `CURRENT_MIGRATION_VERSION = 11` (L325). Bumped when adding any sync migration so all users re-run.
- Stored in `getGlobalConfig().migrationVersion`.
- Ordered sequential execution (L326-342):
  1. `migrateAutoUpdatesToSettings`
  2. `migrateBypassPermissionsAcceptedToSettings`
  3. `migrateEnableAllProjectMcpServersToSettings`
  4. `resetProToOpusDefault`
  5. `migrateSonnet1mToSonnet45`
  6. `migrateLegacyOpusToCurrent`
  7. `migrateSonnet45ToSonnet46`
  8. `migrateOpusToOpus1m`
  9. `migrateReplBridgeEnabledToRemoteControlAtStartup`
  10. `resetAutoModeOptInForDefaultOffer` (only with `TRANSCRIPT_CLASSIFIER` feature)
  11. `migrateFennecToOpus` (only when `USER_TYPE === 'ant'`)
- Finalizer (L343-346): `saveGlobalConfig(prev => prev.migrationVersion === 11 ? prev : { ...prev, migrationVersion: 11 })` (no-op when already current).
- Async migration fire-and-forget: `migrateChangelogFromConfig()` at L349-351.

### 7.2 Idempotency patterns

Two dominant patterns:

**Pattern A — flag-based in-config state** (e.g. `migrateAutoUpdatesToSettings.ts`):
- Check pre-migration source state (`autoUpdates === false && !protectedForNative`).
- Apply change to new source (`settings.env.DISABLE_AUTOUPDATER = '1'`).
- Delete old field from globalConfig (L48-53).

**Pattern B — value-match idempotency** (e.g. `migrateSonnet45ToSonnet46.ts`, `migrateFennecToOpus.ts`):
- Only act if value matches old canonical form.
- Writes same source it reads (userSettings) — keeps idempotent without completion flag.
- Rationale (migrateSonnet45ToSonnet46.ts:27): "Reads userSettings specifically (not merged) so we only migrate what /model wrote — project/local pins are left alone."

### 7.3 Skip conditions

Migrations skip based on:
- Provider check: `getAPIProvider() !== 'firstParty'` (sonnet45ToSonnet46:30).
- Subscriber check: `!isProSubscriber() && !isMaxSubscriber() && !isTeamPremiumSubscriber()` (sonnet45ToSonnet46:34).
- User type: `USER_TYPE !== 'ant'` (fennecToOpus:19).
- Build channel: feature flag gating (`feature('TRANSCRIPT_CLASSIFIER')`).

### 7.4 Scope rules (constitutional)

- Never touch `projectSettings` (could be written by malicious repos).
- Prefer `userSettings` direct source — not merged settings (avoids infinite re-runs, silent global promotion).
- Log via `logEvent('tengu_migrate_*')` with migration-specific metadata.

### 7.5 Logging

Every migration logs success and error cases:
- Success: `tengu_migrate_{name}` with metadata.
- Error: `tengu_migrate_{name}_error` with `has_error: true`.
- Rollback: nothing. Migrations are forward-only; `saveGlobalConfig` only removes old fields after successful write.

### 7.6 Notification metadata

`sonnet45To46MigrationTimestamp: Date.now()` stored in global config (migrateSonnet45ToSonnet46.ts:55-58) — UI shows a banner to users for N days after migration. Skipped for brand-new users (`numStartups <= 1`).

---

## 8. Key Constants & Thresholds

### 8.1 Context window

| Constant | Value | File |
|----------|-------|------|
| `MODEL_CONTEXT_WINDOW_DEFAULT` | 200_000 | `utils/context.ts:9` |
| `COMPACT_MAX_OUTPUT_TOKENS` | 20_000 | `utils/context.ts:12` |
| `MAX_OUTPUT_TOKENS_DEFAULT` | 32_000 | `utils/context.ts:15` |
| `MAX_OUTPUT_TOKENS_UPPER_LIMIT` | 64_000 | `utils/context.ts:16` |
| `CAPPED_DEFAULT_MAX_TOKENS` | 8_000 | `utils/context.ts:24` |
| `ESCALATED_MAX_TOKENS` | 64_000 | `utils/context.ts:25` |
| `1M context (via [1m] suffix or beta)` | 1_000_000 | `utils/context.ts:71, 86` |

### 8.2 Auto-compaction

| Constant | Value | File |
|----------|-------|------|
| `MAX_OUTPUT_TOKENS_FOR_SUMMARY` | 20_000 | `services/compact/autoCompact.ts:30` |
| `AUTOCOMPACT_BUFFER_TOKENS` | 13_000 | `services/compact/autoCompact.ts:62` |
| `WARNING_THRESHOLD_BUFFER_TOKENS` | 20_000 | `services/compact/autoCompact.ts:63` |
| `ERROR_THRESHOLD_BUFFER_TOKENS` | 20_000 | `services/compact/autoCompact.ts:64` |
| `MANUAL_COMPACT_BUFFER_TOKENS` | 3_000 | `services/compact/autoCompact.ts:65` |
| `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES` | 3 | `services/compact/autoCompact.ts:70` |

### 8.3 Tool results

| Constant | Value | File |
|----------|-------|------|
| `DEFAULT_MAX_RESULT_SIZE_CHARS` | 50_000 | `constants/toolLimits.ts:13` |
| `MAX_TOOL_RESULT_TOKENS` | 100_000 | `constants/toolLimits.ts:22` |
| `BYTES_PER_TOKEN` | 4 | `constants/toolLimits.ts:28` |
| `MAX_TOOL_RESULT_BYTES` | 400_000 | `constants/toolLimits.ts:33` |
| `MAX_TOOL_RESULTS_PER_MESSAGE_CHARS` | 200_000 | `constants/toolLimits.ts:49` |
| `TOOL_SUMMARY_MAX_LENGTH` | 50 | `constants/toolLimits.ts:56` |

### 8.4 API limits

| Constant | Value | File |
|----------|-------|------|
| `API_IMAGE_MAX_BASE64_SIZE` | 5 MB | `constants/apiLimits.ts:22` |
| `IMAGE_TARGET_RAW_SIZE` | 3.75 MB | `constants/apiLimits.ts:29` |
| `IMAGE_MAX_WIDTH` / `HEIGHT` | 2000 | `constants/apiLimits.ts:42-43` |
| `PDF_TARGET_RAW_SIZE` | 20 MB | `constants/apiLimits.ts:54` |
| `API_PDF_MAX_PAGES` | 100 | `constants/apiLimits.ts:59` |
| `PDF_EXTRACT_SIZE_THRESHOLD` | 3 MB | `constants/apiLimits.ts:66` |
| `PDF_MAX_EXTRACT_SIZE` | 100 MB | `constants/apiLimits.ts:72` |
| `PDF_MAX_PAGES_PER_READ` | 20 | `constants/apiLimits.ts:77` |
| `PDF_AT_MENTION_INLINE_THRESHOLD` | 10 | `constants/apiLimits.ts:83` |
| `API_MAX_MEDIA_PER_REQUEST` | 100 | `constants/apiLimits.ts:94` |

### 8.5 Memory

| Constant | Value | File |
|----------|-------|------|
| `MAX_ENTRYPOINT_LINES` | 200 | `memdir/memdir.ts:35` |
| `MAX_ENTRYPOINT_BYTES` | 25_000 | `memdir/memdir.ts:38` |
| `MAX_MEMORY_FILES` | 200 | `memdir/memoryScan.ts:21` |
| `FRONTMATTER_MAX_LINES` | 30 | `memdir/memoryScan.ts:22` |

### 8.6 History

| Constant | Value | File |
|----------|-------|------|
| `MAX_HISTORY_ITEMS` | 100 | `history.ts:19` |
| `MAX_PASTED_CONTENT_LENGTH` | 1024 | `history.ts:20` |
| `MAX_TRANSCRIPT_READ_BYTES` | 50 MB | `utils/sessionStorage.ts:229` |
| `MAX_STATUS_CHARS` (git status in context) | 2000 | `context.ts:20` |

### 8.7 Token budget

| Constant | Value | File |
|----------|-------|------|
| `COMPLETION_THRESHOLD` | 0.9 | `query/tokenBudget.ts:3` |
| `DIMINISHING_THRESHOLD` | 500 | `query/tokenBudget.ts:4` |
| `MAX_OUTPUT_TOKENS_RECOVERY_LIMIT` | 3 | `query.ts:164` |
| `MAX_STRUCTURED_OUTPUT_RETRIES` (env override) | 5 | `QueryEngine.ts:1012` |

### 8.8 Beta headers (`constants/betas.ts`)

All dated string constants:
- `CLAUDE_CODE_20250219_BETA_HEADER` = `'claude-code-20250219'`
- `INTERLEAVED_THINKING_BETA_HEADER` = `'interleaved-thinking-2025-05-14'`
- `CONTEXT_1M_BETA_HEADER` = `'context-1m-2025-08-07'`
- `CONTEXT_MANAGEMENT_BETA_HEADER` = `'context-management-2025-06-27'`
- `STRUCTURED_OUTPUTS_BETA_HEADER` = `'structured-outputs-2025-12-15'`
- `WEB_SEARCH_BETA_HEADER` = `'web-search-2025-03-05'`
- `TOOL_SEARCH_BETA_HEADER_1P` = `'advanced-tool-use-2025-11-20'` (Claude API/Foundry)
- `TOOL_SEARCH_BETA_HEADER_3P` = `'tool-search-tool-2025-10-19'` (Vertex/Bedrock)
- `EFFORT_BETA_HEADER` = `'effort-2025-11-24'`
- `TASK_BUDGETS_BETA_HEADER` = `'task-budgets-2026-03-13'`
- `PROMPT_CACHING_SCOPE_BETA_HEADER` = `'prompt-caching-scope-2026-01-05'`
- `FAST_MODE_BETA_HEADER` = `'fast-mode-2026-02-01'`
- `REDACT_THINKING_BETA_HEADER` = `'redact-thinking-2026-02-12'`
- `TOKEN_EFFICIENT_TOOLS_BETA_HEADER` = `'token-efficient-tools-2026-03-28'`
- `SUMMARIZE_CONNECTOR_TEXT_BETA_HEADER` (CONNECTOR_TEXT feature) = `'summarize-connector-text-2026-03-13'`
- `AFK_MODE_BETA_HEADER` (TRANSCRIPT_CLASSIFIER feature) = `'afk-mode-2026-01-31'`
- `CLI_INTERNAL_BETA_HEADER` (ant-only) = `'cli-internal-2026-02-09'`
- `ADVISOR_BETA_HEADER` = `'advisor-tool-2026-03-01'`

Bedrock extra-params set (L38-42): `INTERLEAVED_THINKING`, `CONTEXT_1M`, `TOOL_SEARCH_3P`.
Vertex countTokens allowed betas (L48-52): `CLAUDE_CODE_20250219`, `INTERLEAVED_THINKING`, `CONTEXT_MANAGEMENT`.

### 8.9 Query sources (QuerySource enum)

Used for gating and recursion prevention. Observed values in `query.ts`:
- `'sdk'` — headless SDK.
- `'repl_main_thread'` / `'repl_main_thread_*'` — interactive REPL.
- `'compact'` — forked compact agent.
- `'session_memory'` — forked session memory agent.
- `'agent:...'` — subagent-spawned query.
- `'memdir_relevance'` — findRelevantMemories sideQuery (`findRelevantMemories.ts:121`).
- `'marble_origami'` — ctx-agent (CONTEXT_COLLAPSE).

### 8.10 Environment variables (seen in this scope)

Behavioral:
- `CLAUDE_CODE_EAGER_FLUSH` — force session flush after every persist.
- `CLAUDE_CODE_IS_COWORK` — cowork integration mode.
- `CLAUDE_CODE_SIMPLE` / `--bare` — minimal mode (skip CLAUDE.md autodiscovery, auto-memory, skill prefetch).
- `CLAUDE_CODE_SKIP_PROMPT_HISTORY` — skip history writes (used by Tungsten test sessions).
- `CLAUDE_CODE_REMOTE` (CCR) — remote/server mode.
- `CLAUDE_CODE_EMIT_TOOL_USE_SUMMARIES` — enable tool summary generation.
- `CLAUDE_CODE_DISABLE_CLAUDE_MDS` — hard off CLAUDE.md autodiscovery.
- `CLAUDE_CODE_DISABLE_AUTO_MEMORY` — hard off memdir.
- `CLAUDE_CODE_DISABLE_1M_CONTEXT` — HIPAA/compliance.
- `CLAUDE_CODE_DISABLE_FAST_MODE` — disable fast-mode routing.
- `CLAUDE_CODE_MAX_CONTEXT_TOKENS` (ant-only) — cap effective context window.
- `CLAUDE_CODE_AUTO_COMPACT_WINDOW` — cap autocompact window.
- `CLAUDE_CODE_MAX_OUTPUT_TOKENS` — override max output cap.
- `CLAUDE_CODE_BLOCKING_LIMIT_OVERRIDE` — test override.
- `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` — percentage-based compact override.
- `DISABLE_COMPACT` / `DISABLE_AUTO_COMPACT` — disable compaction paths.
- `DISABLE_AUTOUPDATER` — block auto-updater (migration-target).
- `MAX_STRUCTURED_OUTPUT_RETRIES` — structured output retry limit (default 5).
- `CLAUDE_CODE_REMOTE_MEMORY_DIR` — remote memory mount override.
- `CLAUDE_COWORK_MEMORY_PATH_OVERRIDE` — full-path memory override.
- `CLAUDE_COWORK_MEMORY_EXTRA_GUIDELINES` — extra memory policy text.
- `CLAUDE_JOB_DIR` — dispatched job state dir (TEMPLATES feature).
- `CLAUDE_CODE_ENTRYPOINT` — attribution header entrypoint name.
- `CLAUDE_CODE_ATTRIBUTION_HEADER` — kill switch for attribution.
- `IS_DEMO` — demo mode (disables onboarding).
- `USER_TYPE` = `'ant'` — internal Anthropic user build gate.
- `NODE_ENV` = `'test'` — test mode (disables getGitStatus).

---

## 9. Gap vs moai-adk (CC HAS / moai-adk LACKS)

### 9.1 Query lifecycle

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Stateful conversation engine | `QueryEngine` class with mutable messages, read-file cache, session state | None — moai is an orchestrator that delegates; no persistent query engine |
| Multi-stage compaction (snip → microcompact → collapse → autocompact → reactive) | 5-layer compaction pipeline in `query.ts:396-467, 1062-1184` | None — no compaction at all |
| Streaming tool execution | `StreamingToolExecutor` fires tools mid-stream | N/A — moai runs on top of Claude Code which handles this |
| Fallback model support | `FallbackTriggeredError` switches model mid-turn with history strip (`query.ts:894-952`) | None |
| Max-output-tokens recovery | 3-attempt nudge + escalate 8k→64k (`query.ts:1185-1256`) | None |
| Structured-output retry cap | Env-configurable per-turn limit (`QueryEngine.ts:1010-1048`) | None |
| Hard budget enforcement (USD) | Per-iteration cost check in QueryEngine | No; no cost tracker at all |
| Max-turns ceiling | `maxTurns` in params, checked per iteration | No explicit ceiling |
| Stop/TeammateIdle/TaskCompleted hooks with blocking-error re-injection | Full pipeline in `query/stopHooks.ts` — hooks can inject blocking errors that continue the loop | TeammateIdle + TaskCompleted hooks exist in MoAI docs but without the continuation-loop semantics |

### 9.2 Context loading

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Memoized system prompt (cached per conversation) | `getSystemContext`, `getUserContext` via `memoize()` | Static template files; no memoization |
| Git-status injection with 2K truncation | `context.ts:86-103` | Manual mention in CLAUDE.md (this very file shows git status was injected by CC into moai's context) |
| Multi-directory (`additionalWorkingDirectories`) | `AppState.toolPermissionContext.additionalWorkingDirectories` map | Single project directory assumed |
| System prompt injection for cache breaking (`BREAK_CACHE_COMMAND`) | `setSystemPromptInjection` | None |
| Append-system-prompt vs custom-system-prompt distinction | Both first-class `QueryEngineConfig` fields; affect `getSystemContext` skip | moai Skills use only prepend (via YAML frontmatter) |
| `$CLAUDE_COWORK_MEMORY_PATH_OVERRIDE` ⇒ auto-inject memory-mechanics prompt | `QueryEngine.ts:316-319` | None |

### 9.3 History & resume

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Global prompt history file (`~/.claude/history.jsonl`) locked, 100-entry window | `history.ts:106-217, 291-327` | None — moai uses Claude Code's history |
| Paste store with hash reference for >1024-char pastes | `history.ts:355-409` | None |
| Removable history entries (ESC-undo integration) | `removeLastFromHistory` + skippedTimestamps | None |
| Per-session transcript file (`{projectDir}/{sessionId}.jsonl`) | `utils/sessionStorage.ts` | None; moai has `.moai/state/` but that's SPEC progress, not transcript |
| 50 MB max transcript read guard | `MAX_TRANSCRIPT_READ_BYTES` | N/A |
| Eager-flush mode for cowork/desktop kill races | `CLAUDE_CODE_EAGER_FLUSH` env | N/A |
| Compact boundary GC of pre-compact messages | `QueryEngine.ts:925-933` | N/A |
| Agent transcript subdir grouping | `agentTranscriptSubdirs: Map` | N/A |

### 9.4 Cost tracker

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Per-model token accumulation (in/out/cache-read/cache-write) | `cost-tracker.ts:250-276 addToTotalModelUsage` | None |
| Per-model USD cost table | `calculateUSDCost(model, usage)` | None |
| Session-scoped cost restore on resume | `restoreCostStateForSession(sessionId)` | None |
| OpenTelemetry counters | `getCostCounter`, `getTokenCounter` per token type | None — CLAUDE.local.md explicitly bans OTEL env vars in tests |
| Advisor recursive cost accounting | `addToTotalSessionCost` + advisor usage loop | None |
| FPS perf metrics in cost state | `lastFpsAverage`, `lastFpsLow1Pct` | None |
| USD budget enforcement as hard ceiling | `maxBudgetUsd` in QueryEngineConfig | None |
| Fast-mode attribution | `usage.speed === 'fast'` → `{ model, speed: 'fast' }` attr | N/A |

### 9.5 Memory system (memdir)

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Typed memory taxonomy (user/feedback/project/reference) as HARD constitution | `memdir/memoryTypes.ts` with full XML prompts | **moai-adk MEMORY.md already uses this schema** (see `.../memory/MEMORY.md` format in-conversation); conventions align but no enforcement |
| MEMORY.md 200-line + 25K-byte truncation with warning append | `memdir/memdir.ts:57-103 truncateEntrypointContent` | No truncation — MoAI CLAUDE.md has its own 40K char cap (for CLAUDE.md, not MEMORY.md) |
| Auto-directory resolution (git-root-canonical, tilde expansion, env override chain) | `memdir/paths.ts:223-235` | moai uses fixed `.claude/agent-memory/` and `~/.claude/projects/{hash}/memory/` paths — no env override |
| Settings.json `autoMemoryDirectory` override (trusted sources only) | `getAutoMemPathSetting` — excludes projectSettings | None |
| Memory scan with mtime sort, cap 200 files, 30-line frontmatter read | `memoryScan.ts` | None |
| LLM-based relevance selection (Sonnet sideQuery, max 5 memories) | `findRelevantMemories.ts` | moai "Context Search Protocol" scans `~/.claude/projects/{hash}/` sessions but not memories; uses Grep not LLM selection |
| `alreadySurfaced` dedup within session | `findRelevantMemories` param | None |
| Memory age string (today/yesterday/N days ago) | `memoryAge.ts:15-20` | None |
| Memory freshness caveat for > 1 day old memories (`<system-reminder>` wrapped) | `memoryAge.ts:33-53` | None |
| Memory prefetch in parallel with model streaming | `startRelevantMemoryPrefetch` / `pendingMemoryPrefetch` in query loop | None — moai context search is blocking |
| Duplicate detection via readFileState | `filterDuplicateMemoryAttachments` | None |
| KAIROS daily-log mode (append-only logs distilled nightly) | `buildAssistantDailyLogPrompt` | None |
| TEAMMEM sub-dir with path-key sanitization | `memdir/teamMemPaths.ts` | None; moai team memory is implicit in `.moai/project/brand/` but not path-validated |
| ensureMemoryDirExists before prompt build | `memdir/memdir.ts:129-147` | moai assumes directories exist |

### 9.6 Migration framework

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| `CURRENT_MIGRATION_VERSION` global config field + ordered migration runner | `main.tsx:325-352` | **None** — CLAUDE.local.md §5 explicitly notes "Migrations: none (moai update for template sync only)" |
| 11 sync migrations + 1 async migration | `migrations/` directory | None |
| Source-isolation (userSettings only) | all migrations | Template-first rule does manual sync |
| Provider/subscriber/build-channel gating | `migrateSonnet45ToSonnet46`, `migrateFennecToOpus` | None |
| Migration telemetry events | `tengu_migrate_{name}` + error variants | None |
| Timestamp-based notification trigger (`sonnet45To46MigrationTimestamp`) | `migrateSonnet45ToSonnet46.ts:55-58` | None |
| Forward-only, non-rollback design | All migrations | N/A |

### 9.7 Autocompaction (unique gaps)

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Proactive token-based compaction threshold | `effectiveWindow - 13K` | None |
| Blocking limit with `MANUAL_COMPACT_BUFFER_TOKENS=3K` reserve | `autoCompact.ts:123-136` | None |
| Circuit breaker after 3 consecutive compaction failures | `autoCompact.ts:70, 257-265` | None |
| Session-memory compaction as first attempt | `trySessionMemoryCompaction` | None |
| Reactive compaction on withheld 413/media errors | `services/compact/reactiveCompact.js` wired into `query.ts:1119-1166` | None |
| Context-collapse (90/95% commit-then-block architecture) | CONTEXT_COLLAPSE feature | None |
| Snip compaction (zombie message removal) | HISTORY_SNIP feature | None |
| Microcompact (cache-level trimming, API-reported deletion counts) | CACHED_MICROCOMPACT feature | None |

### 9.8 Token budget (soft)

| Capability | Claude Code | moai-adk-go |
|------------|-------------|-------------|
| Turn-scoped token budget with 90% completion threshold | `query/tokenBudget.ts:3, 45-92` | moai Skills have token budget in Progressive Disclosure but no turn-level accounting |
| Diminishing-returns detection (3+ continuations, <500 token delta) | `query/tokenBudget.ts:59-62` | None |
| Budget continuation nudge message per turn | `getBudgetContinuationMessage` → meta user message | None |
| `incrementBudgetContinuationCount` global counter | `bootstrap/state.js` | None |

### 9.9 Features present in both (alignment opportunity)

- Project-scoped memory directory keyed on canonical git root: **CC uses `findCanonicalGitRoot`, moai uses sanitized project hash** — conceptually identical but hashing differs.
- Memory index file named `MEMORY.md`: **matches exactly**.
- Memory type taxonomy `user|feedback|project|reference`: **matches exactly** (moai-adk's memory/MEMORY.md references user/feedback/project types).
- 200-line cap intent: **moai has 40K char cap on CLAUDE.md at coding-standards.md, CC has 200 line / 25K byte cap on MEMORY.md** — complementary.
- `<system-reminder>` wrapping: **CC uses for staleness; moai uses for task reminders** (visible in this very conversation).

---

## 10. Source References (file:line)

### 10.1 QueryEngine.ts

- `QueryEngineConfig` type: L130-173
- `QueryEngine` class: L184-1177
- `submitMessage` entry: L209-1156
- canUseTool wrapping: L244-271
- Model + thinking config: L273-282
- `fetchSystemPromptParts` call: L288-300
- Memory mechanics prompt injection: L316-319
- System prompt assembly: L321-325
- processUserInputContext (first pass): L335-395
- Orphaned permission handler: L398-408
- processUserInput call: L410-428
- Transcript pre-record: L450-463
- Permission context update: L477-486
- Skill + plugin parallel load: L534-538
- System init message: L540-551
- Local command fast path: L556-638
- Main query loop: L675-1049
- maxBudgetUsd enforcement: L972-1002
- Structured output retry cap: L1004-1048
- Compact boundary GC splice: L925-933
- Snip replay callback: L905-915
- `ask()` wrapper: L1186

### 10.2 query.ts

- `QueryParams`: L181-199
- `State`: L204-217
- `query()`: L219-239
- `queryLoop()`: L241-1728
- `MAX_OUTPUT_TOKENS_RECOVERY_LIMIT`: L164
- `isWithheldMaxOutputTokens`: L175-179
- Memory prefetch: L301-304
- Skill prefetch: L331-335
- Tool result budget: L379-394
- HISTORY_SNIP: L401-410
- Microcompact: L414-426
- CONTEXT_COLLAPSE apply: L440-447
- Autocompact call: L453-467
- Post-compact handling: L470-543
- StreamingToolExecutor: L561-568
- Runtime model (plan-mode 200k guard): L572-578
- createDumpPromptsFetch: L588-590
- Blocking limit check: L628-648
- Model streaming: L653-954
- Fallback model (FallbackTriggeredError): L894-952
- Withheld error tracking: L788-822
- Abort handling: L1015-1052
- PTL recovery (collapse drain): L1085-1117
- Reactive compact retry: L1119-1166
- Max-output-tokens escalate: L1185-1221
- Max-output-tokens 3-retry nudge: L1223-1256
- API error skip stop hooks: L1262-1265
- Stop hooks: L1267-1306
- Token budget check: L1308-1355
- Tool execution dispatch: L1366-1408
- Tool use summary (Haiku): L1415-1482
- Abort during tools: L1485-1516
- Attachment injection: L1580-1590
- Memory prefetch consume: L1599-1614
- Skill prefetch consume: L1620-1628
- BG_SESSIONS task summary: L1685-1702
- Max turns ceiling: L1705-1712

### 10.3 query/

- `query/config.ts` — `QueryConfig` snapshot, Statsig gates (streamingToolExecution, emitToolUseSummaries, isAnt, fastModeEnabled)
- `query/deps.ts` — `QueryDeps = { callModel, microcompact, autocompact, uuid }` dependency injection for tests
- `query/tokenBudget.ts` — `BudgetTracker`, `checkTokenBudget`, `COMPLETION_THRESHOLD=0.9`, `DIMINISHING_THRESHOLD=500`
- `query/stopHooks.ts` — `handleStopHooks` generator with progress, blocking errors, preventContinuation; TeammateIdle + TaskCompleted sub-hooks

### 10.4 context.ts + context/

- `context.ts` — `getSystemContext`, `getUserContext`, `getGitStatus` (memoized)
- `MAX_STATUS_CHARS = 2000` (L20)
- `systemPromptInjection` for cache breaking (L22-34)
- `context/` subdir — React/ink UI contexts (stats, notifications, overlays, mailbox, voice) — **not query-context; excluded from scope**

### 10.5 history.ts

- `MAX_HISTORY_ITEMS = 100` (L19)
- `MAX_PASTED_CONTENT_LENGTH = 1024` (L20)
- `StoredPastedContent` type: L25-32
- Reference parsing / formatting: L47-100
- `makeLogEntryReader`: L106-143
- `getTimestampedHistory` / `getHistory`: L162-217
- `LogEntry` schema: L219-225
- `addToHistory` entry: L411-434
- `immediateFlushHistory`: L292-327
- `removeLastFromHistory`: L453-464

### 10.6 cost-tracker.ts

- `StoredCostState`: L71-80
- `getStoredSessionCosts`: L87-123
- `restoreCostStateForSession`: L130-137
- `saveCurrentSessionCosts`: L143-175
- `formatCost`: L177-179
- `formatModelUsage`: L181-226
- `formatTotalCost`: L228-244
- `addToTotalModelUsage`: L250-276
- `addToTotalSessionCost`: L278-323

### 10.7 memdir/

- `paths.ts:30-55` — `isAutoMemoryEnabled` priority chain
- `paths.ts:109-150` — `validateMemoryPath` security rules
- `paths.ts:194-196` — `hasAutoMemPathOverride`
- `paths.ts:223-235` — `getAutoMemPath` memoized resolution
- `paths.ts:246-251` — `getAutoMemDailyLogPath`
- `paths.ts:274-278` — `isAutoMemPath`
- `memdir.ts:34-38` — entrypoint constants
- `memdir.ts:57-103` — `truncateEntrypointContent`
- `memdir.ts:129-147` — `ensureMemoryDirExists`
- `memdir.ts:199-266` — `buildMemoryLines`
- `memdir.ts:272-316` — `buildMemoryPrompt`
- `memdir.ts:327-370` — `buildAssistantDailyLogPrompt` (KAIROS)
- `memdir.ts:375-407` — `buildSearchingPastContextSection`
- `memdir.ts:419-507` — `loadMemoryPrompt` dispatcher
- `memoryTypes.ts:14-19` — `MEMORY_TYPES` tuple
- `memoryTypes.ts:37-106` — `TYPES_SECTION_COMBINED`
- `memoryTypes.ts:113-*` — `TYPES_SECTION_INDIVIDUAL`
- `memoryScan.ts:21-22` — `MAX_MEMORY_FILES`, `FRONTMATTER_MAX_LINES`
- `memoryScan.ts:35-77` — `scanMemoryFiles`
- `memoryScan.ts:84-94` — `formatMemoryManifest`
- `memoryAge.ts:6-53` — day counting + freshness text
- `findRelevantMemories.ts:18-24` — selector system prompt
- `findRelevantMemories.ts:39-75` — `findRelevantMemories`
- `findRelevantMemories.ts:77-141` — `selectRelevantMemories` Sonnet sideQuery
- `teamMemPaths.ts:10-64` — path sanitization including NFKC check
- `teamMemPaths.ts:73-94` — team memory path resolution

### 10.8 projectOnboardingState.ts

- `getSteps` (L19-41): `workspace` + `claudemd` steps.
- `isProjectOnboardingComplete` (L43-47).
- `maybeMarkProjectOnboardingComplete` (L49-61).
- `shouldShowProjectOnboarding` (L63-76): caps at 4 seen.
- `incrementProjectOnboardingSeenCount` (L78-83).

### 10.9 migrations/

- Runner: `main.tsx:325-352`
- `CURRENT_MIGRATION_VERSION = 11`
- `migrateAutoUpdatesToSettings.ts` — autoUpdates false → env var `DISABLE_AUTOUPDATER=1`
- `migrateBypassPermissionsAcceptedToSettings.ts` — `bypassPermissionsModeAccepted` → `skipDangerousModePermissionPrompt`
- `migrateEnableAllProjectMcpServersToSettings.ts` — project config MCP enable-all → settings
- `migrateFennecToOpus.ts` (ant-only) — fennec-* aliases → opus/opus[1m]
- `migrateLegacyOpusToCurrent.ts` — old opus → current
- `migrateOpusToOpus1m.ts` — opus → opus[1m] when merge enabled
- `migrateReplBridgeEnabledToRemoteControlAtStartup.ts` — `replBridgeEnabled` key rename
- `migrateSonnet1mToSonnet45.ts` — `sonnet[1m]` → explicit `claude-sonnet-4-5-*`
- `migrateSonnet45ToSonnet46.ts` — explicit 4.5 → `sonnet` alias (Pro/Max/TeamPremium only)
- `resetAutoModeOptInForDefaultOffer.ts` — reset auto-mode opt-in for new offer
- `resetProToOpusDefault.ts` — reset Pro default to Opus

### 10.10 constants/

- `apiLimits.ts` — image/PDF/media server-side limits
- `toolLimits.ts` — tool result size caps (per-result, per-message, token estimation)
- `betas.ts` — dated beta header strings + provider allowlists (Bedrock, Vertex countTokens)
- `common.ts` — `getLocalISODate`, `getSessionStartDate`, `getLocalMonthYear`
- `system.ts` — `DEFAULT_PREFIX`, `AGENT_SDK_*` sysprompt prefixes, `getCLISyspromptPrefix`, `getAttributionHeader` (billing `x-anthropic-billing-header`)
- `files.ts` — binary extension set, `hasBinaryExtension`, `isBinaryContent`
- `prompts.ts` (54K bytes) — the main system prompt template (`getSystemPrompt`); NOT inspected in depth but referenced at `utils/queryContext.ts:64`
- `xml.ts` — `LOCAL_COMMAND_STDOUT_TAG`, `LOCAL_COMMAND_STDERR_TAG` used in `QueryEngine.ts:24-26`
- `querySource.ts` — `QuerySource` type (enum of query origins)
- `systemPromptSections.ts` — section identifier constants (referenced in memory prompt caching)
- `toolLimits.ts` — above
- `messages.ts` — minimal (49 bytes; likely just re-exports)

### 10.11 Cross-cutting

- `utils/queryContext.ts:44-74` — `fetchSystemPromptParts` entry
- `utils/queryContext.ts:88-130` — `buildSideQuestionFallbackParams` (resume cache-prefix rebuild)
- `utils/context.ts:51-98` — `getContextWindowForModel` resolution chain
- `utils/context.ts:149-215` — `getModelMaxOutputTokens`
- `utils/sessionStorage.ts:139` — `isTranscriptMessage` type guard
- `utils/sessionStorage.ts:194-196` — `isEphemeralToolProgress`
- `utils/sessionStorage.ts:202-205` — `getTranscriptPath`
- `utils/sessionStorage.ts:229` — `MAX_TRANSCRIPT_READ_BYTES = 50 MB`
- `services/compact/autoCompact.ts:30-70` — constants
- `services/compact/autoCompact.ts:33-49` — `getEffectiveContextWindowSize`
- `services/compact/autoCompact.ts:72-91` — `getAutoCompactThreshold`
- `services/compact/autoCompact.ts:93-145` — `calculateTokenWarningState`
- `services/compact/autoCompact.ts:147-158` — `isAutoCompactEnabled`
- `services/compact/autoCompact.ts:160-239` — `shouldAutoCompact`
- `services/compact/autoCompact.ts:241-351` — `autoCompactIfNeeded`

---

End of Wave 1.2 research. Total thresholds cataloged: 50+. Total file:line references: 150+. Key insight for moai-adk-go v3: **moai-adk currently has ZERO of the core query-engine capabilities** (compaction, budget enforcement, streaming tool exec, fallback handling, max-output recovery, cost tracking, memory freshness, migration framework) because it sits entirely on top of Claude Code. The v3 design should decide explicitly for each capability: (a) inherit from Claude Code runtime, (b) reimplement in Go for moai CLI parity, or (c) expose a passthrough. Memory taxonomy and MEMORY.md format already align strongly — graduating moai's ad-hoc `.claude/agent-memory/` into Claude Code's memdir spec is the lowest-friction path.

# Research — SPEC-V3R4-HARNESS-002 (Multi-Event Observer Expansion)

This research artifact is the Phase 0.5 deliverable of the plan-phase for SPEC-V3R4-HARNESS-002. It surveys the Claude Code hook surface for the four target events (PostToolUse, Stop, SubagentStop, UserPromptSubmit), audits the V3R4-HARNESS-001 baseline implementation, proposes per-event payload extensions, codifies the PII handling policy for UserPromptSubmit, and outlines the 5-Layer Safety integration boundary.

All file references use project-root-relative paths.

---

## 1. Claude Code Hook Surface (4 Target Events)

The four target hook events are all Claude Code first-class lifecycle hooks. They share a common invocation contract: Claude Code posts JSON to stdin of the registered command and reads JSON from stdout. Exit codes and `decision`/`systemMessage` fields control flow. The runtime contracts are codified in `.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout Reference.

### 1.1 PostToolUse (baseline — already wired)

**Trigger semantics**: Fires after each tool invocation by Claude completes (Write, Edit, Bash, Read, Glob, Grep, Agent, AskUserQuestion, etc.). Settings.json matcher in the existing template (`internal/template/templates/.claude/settings.json.tmpl`) is `Write|Edit` — only edits are observed by the existing PostToolUse → MX/LSP/codemap pipeline, but the harness observer wraps a DIFFERENT hook entry.

**stdin JSON fields supplied by Claude Code** (per hooks-system.md table):
- `toolName` (string): e.g. `"Edit"`, `"Bash"`, `"Agent"`
- `toolInput` (object): tool-specific input shape
- `toolOutput` (object | null): tool result shape
- `duration_ms` (number, v2.1.119+): tool execution duration
- `session.id`, `session.cwd`, `session.projectDir`
- `is_interrupt` (boolean, on failures)

**Existing wiring (V3R4-001 baseline)**: The harness observer is wired via a SEPARATE hook entry, not embedded into the PostToolUse handler chain. The wrapper script `.claude/hooks/moai/handle-harness-observe.sh` (template at `internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl`) forwards stdin JSON to `moai hook harness-observe`, which calls `runHarnessObserve` at `internal/cli/hook.go:474`. This separates the harness observation pipeline from the existing PostToolUse handler (which runs MX/LSP/codemap logic via `moai hook post-tool`).

**Settings.json registration status**: The existing `settings.json.tmpl` registers `PostToolUse` → `handle-post-tool.sh` (matcher `Write|Edit`, timeout 10s, async) but DOES NOT register a separate harness-observe entry. This is a gap (see §2.3 finding).

### 1.2 Stop

**Trigger semantics**: Fires when Claude finishes responding to a user turn (i.e., the assistant's last message is emitted and the turn closes). Per Claude Code v2.1.49+, the stdin includes `last_assistant_message` so handlers see what the assistant just said.

**stdin JSON fields**:
- `last_assistant_message` (string, v2.1.49+): the assistant's final message text
- `session.id`, `session.cwd`, `session.projectDir`
- `agent_id`, `agent_type` (v2.1.69+ if triggered inside an agent context, but Stop is the orchestrator-turn signal so usually absent for the main session)

**stdout contract**: `systemMessage` (optional). Stop can block (`exit code 2`), but blocking is the wrong semantic for an observer — we always exit 0.

**Existing wiring**: `.claude/hooks/moai/handle-stop.sh` forwards stdin to `moai hook stop`, which dispatches into the central `runHookEvent` registry (no harness observation yet). Settings.json registers `Stop` → `handle-stop.sh` with timeout 5s.

**Trigger semantics for harness learning signal**: Stop is the per-turn boundary. Each user turn ends with exactly one Stop. This makes Stop the canonical signal for "session-level completion" patterns — useful for downstream classifiers (e.g., turn frequency per project, average turn duration, common turn-end states).

### 1.3 SubagentStop

**Trigger semantics**: Fires when an `Agent()`-spawned subagent (sub-task or teammate) terminates. Each subagent invocation produces exactly one SubagentStop on completion.

**stdin JSON fields** (per hooks-system.md table, v2.1.42/v2.1.69 evolution):
- `agentType` (string): e.g. `"manager-spec"`, `"expert-backend"`, `"general-purpose"`
- `agentName` (string): the dispatch name
- `last_assistant_message` (string): the subagent's final message
- `agent_id` (string, v2.1.69+): unique agent invocation ID
- `agent_transcript_path` (string, v2.1.69+): on-disk path to the subagent's transcript
- `session.id`, `session.cwd`, `session.projectDir`

**stdout contract**: `systemMessage` (optional). SubagentStop can block (`exit code 2`); observers must always exit 0.

**Existing wiring**: `.claude/hooks/moai/handle-subagent-stop.sh` forwards to `moai hook subagent-stop`. Settings.json registers `SubagentStop` → `handle-subagent-stop.sh` with timeout 5s.

**Trigger semantics for harness learning signal**: SubagentStop reveals which agents are invoked most often, which agents fail vs succeed, and the population mix of role profiles (researcher / analyst / implementer / etc.). This is the canonical signal for "agent invocation outcome" patterns.

### 1.4 UserPromptSubmit

**Trigger semantics**: Fires immediately when the user submits a prompt — BEFORE Claude processes it. This is the earliest hook in the user-turn lifecycle.

**stdin JSON fields**:
- `prompt` (string): the user's raw prompt content
- `session.id`, `session.cwd`, `session.projectDir`

**stdout contract**: `additionalContext` (string), `reason` (string). UserPromptSubmit CAN block (`exit code 2`); observers must always exit 0 to remain non-blocking.

**Existing wiring**: `.claude/hooks/moai/handle-user-prompt-submit.sh` forwards to `moai hook user-prompt-submit`. Settings.json registers `UserPromptSubmit` → `handle-user-prompt-submit.sh` with timeout 5s. The MoAI orchestrator already uses this hook (per hooks-system.md § Smart Hook Behaviors) to set session title with SPEC ID or project/branch info.

**Trigger semantics for harness learning signal**: UserPromptSubmit reveals user intent patterns — what kinds of prompts the user issues most, prompt length distribution, language patterns, and (potentially) which SPEC IDs / commands are most referenced. This is the canonical signal for "user intent" patterns.

**PII concern**: `prompt` contains raw user content. Recording it in plaintext risks leaking conversation history, secrets accidentally typed by users, third-party identifiers, and personal data. Per §3.3 below, the default observation strategy MUST be privacy-safe (SHA-256 hash + metadata).

---

## 2. V3R4-HARNESS-001 Observer Implementation (Baseline)

### 2.1 Code Path

The existing observer pipeline (shipped by SPEC-V3R4-HARNESS-001, merged via PR #910 at commit `bb80ea0f4`) consists of:

| Component | File:Line | Purpose |
|-----------|-----------|---------|
| `runHarnessObserve` | `internal/cli/hook.go:474` | Cobra RunE handler for `moai hook harness-observe`; reads stdin JSON, calls `Observer.RecordEvent`. |
| `isHarnessLearningEnabled` | `internal/cli/hook.go:439` | Gate function (REQ-HRN-FND-009). Reads `.moai/config/sections/harness.yaml`, returns `false` only if `learning.enabled: false` is explicit. Fail-open on missing config or parse error. |
| `hookCmd.AddCommand(&cobra.Command{Use: "harness-observe", ...})` | `internal/cli/hook.go:94-99` | Cobra subcommand registration. Note: this is the ONLY existing harness-* subcommand. |
| `Observer` struct | `internal/harness/observer.go:18` | Owns logPath, retention, nowFn. Created via `NewObserver(logPath)` or `NewObserverWithRetention(logPath, retention)`. |
| `Observer.RecordEvent` | `internal/harness/observer.go:53` | Marshals an `Event` struct to JSONL and appends to `.moai/harness/usage-log.jsonl` (O_APPEND\|O_CREATE\|O_WRONLY). Triggers lazy retention pruning when `retention` is set. |
| `Event` struct | `internal/harness/types.go:36` | The canonical JSONL line schema: `Timestamp`, `EventType`, `Subject`, `ContextHash`, `TierIncrement`, `SchemaVersion`. |

### 2.2 JSONL Schema (Baseline)

The `Event` struct serializes to JSONL with these field names (REQ-HRN-FND-010 contract):

```json
{
  "timestamp": "2026-05-14T12:34:56.789Z",
  "event_type": "agent_invocation",
  "subject": "Edit",
  "context_hash": "",
  "tier_increment": 0,
  "schema_version": "v1"
}
```

`EventType` is an enum (`internal/harness/types.go:15`) with four constants: `moai_subcommand`, `agent_invocation`, `spec_reference`, `feedback`. **Critically, these are SEMANTIC categories, not Claude Code event names.** A PostToolUse hook records `event_type: agent_invocation` (the tool itself is observed as an invocation event), not `event_type: PostToolUse`.

This semantic-vs-runtime distinction is load-bearing for V3R4-002. The new four events (`Stop`, `SubagentStop`, `UserPromptSubmit`, plus the existing `PostToolUse`) are Claude Code RUNTIME events. They must map to harness SEMANTIC event types, which means V3R4-002 must either:
- (a) extend the `EventType` enum with new SEMANTIC values that map per runtime event (e.g., `session_stop`, `subagent_stop`, `user_prompt`), OR
- (b) keep the existing four enum values and reuse them across runtime events with the new `Subject` and extended-payload fields disambiguating the runtime source.

Option (a) is cleaner. Option (b) would require Subject-overloading semantics that obscure the runtime event source. The plan deliverable will recommend option (a).

### 2.3 Test Surface (existing patterns)

`internal/cli/hook_harness_observe_test.go` (already merged, 274 lines) provides:

- `TestIsHarnessLearningEnabled` — table-driven 7 cases covering the gate truth table (missing config / empty / no learning block / explicit true / explicit false / parse error / etc.). Pattern: fixtures via `writeHarnessYAML(t, dir, body)` helper.
- `TestRunHarnessObserve_NoOpWhenLearningDisabled` — verifies usage-log.jsonl is NOT created when `learning.enabled: false`.
- `TestRunHarnessObserve_PreservesExistingLogWhenDisabled` — verifies existing JSONL entries are NOT mutated when gate is closed.
- `TestRunHarnessObserve_RecordsWhenEnabled` — verifies one JSONL entry with REQ-HRN-FND-010 baseline 4 fields when gate is open.

These patterns are directly reusable for the four new event handlers. The new tests follow the same fixture pattern (`writeHarnessYAML` + `withStdin` + `t.Chdir(dir)` + assertion on `.moai/harness/usage-log.jsonl`).

### 2.4 Wrapper Script Pattern

The existing `handle-harness-observe.sh` (also has `.tmpl` template at `internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl`) is a 33-line bash script:

1. Sets `MOAI_HOOK_STDERR_LOG` for stderr redirection
2. Best-effort log rotation at 10MB
3. Searches for the `moai` binary in three locations (PATH → `~/go/bin/moai` → `~/.local/bin/moai`)
4. `exec`s `moai hook harness-observe 2>>"$MOAI_HOOK_STDERR_LOG"`
5. Exits 0 silently if binary not found (Claude Code handles missing hooks gracefully)

The pattern is identical across all existing wrappers (`handle-stop.sh`, `handle-user-prompt-submit.sh`, `handle-subagent-stop.sh`, etc.). The only varying lines are the binary subcommand name (line 19/24/29) and the file-header comment. This makes the wrapper extension trivial.

### 2.5 Settings.json.tmpl Gap Finding (load-bearing)

The existing `internal/template/templates/.claude/settings.json.tmpl` registers:
- `PostToolUse` → `handle-post-tool.sh` (matcher `Write|Edit`, timeout 10, async)
- `Stop` → `handle-stop.sh` (timeout 5)
- `SubagentStop` → `handle-subagent-stop.sh` (timeout 5)
- `UserPromptSubmit` → `handle-user-prompt-submit.sh` (timeout 5)

But it does NOT register a separate entry for `handle-harness-observe.sh`. Per V3R4-001 merged state, the harness observer is reachable as `moai hook harness-observe` CLI subcommand and via the `handle-harness-observe.sh` wrapper, but the wrapper is currently UNINVOKED at Claude Code runtime because the settings.json.tmpl has no hook entry routing to it.

This gap is OUT OF SCOPE for V3R4-002 to retroactively fix for PostToolUse (V3R4-001 shipped the gate + wrapper but not the registration — that is V3R4-001's responsibility for cleanup in a follow-up). V3R4-002 must adopt one of two strategies for the new three events:

- **Strategy WIRE-A (separate hook entries, additive)**: register a SECOND entry under each event slot in settings.json.tmpl that routes to a per-event harness wrapper (e.g., add `handle-harness-observe-stop.sh` under the `Stop` event slot in addition to the existing `handle-stop.sh`). This keeps observer logic fully isolated from the existing hook handlers.
- **Strategy WIRE-B (integrate into existing wrappers)**: modify the existing `handle-stop.sh`, `handle-subagent-stop.sh`, `handle-user-prompt-submit.sh` to ALSO call `moai hook harness-observe-<event>` after the main handler. This means modifying templates that are co-owned by other features.

The plan deliverable will recommend WIRE-A because it preserves separation of concerns, simplifies testing (each observer is independently testable), and keeps the existing `handle-*-sh` templates untouched (zero conflict with other concurrent work). The PostToolUse wiring gap from V3R4-001 is documented here but is NOT in scope for this SPEC's fix.

---

## 3. Per-Event Payload Design Analysis

For each new event, V3R4-002 proposes a payload-extension schema. All extensions follow the additivity rule (§4): they are OPTIONAL fields, `omitempty` JSON-serialized, and old entries lacking them remain valid under the baseline 4-field contract from REQ-HRN-FND-010.

### 3.1 Stop — session-level fields

Proposed extension to the `Event` struct (Stop-specific fields, omitempty):

| Field | JSON name | Source | Notes |
|-------|-----------|--------|-------|
| `SessionID` | `session_id` | `session.id` from stdin | UUID-style identifier supplied by Claude Code; safe to record. |
| `MessageCount` | `message_count` | computed (turn index) | Optional. Requires per-session counter held in `.moai/state/`; if not available, omit. |
| `LastAssistantMessageHash` | `last_assistant_message_hash` | SHA-256 of `last_assistant_message` truncated to 16 hex chars | Privacy-safe digest. Raw message NOT recorded. |
| `LastAssistantMessageLen` | `last_assistant_message_len` | `len(last_assistant_message)` in bytes | Signal for turn-length distribution. |

`EventType` (semantic) for Stop entries: `session_stop` (new enum value).

`Subject` semantics for Stop: empty string OR the SPEC ID extracted from the cwd if a SPEC directory is detected (`.moai/specs/SPEC-XXX/`). The Subject field carries the per-event keying that the downstream classifier uses.

### 3.2 SubagentStop — agent invocation outcome fields

Proposed extension:

| Field | JSON name | Source | Notes |
|-------|-----------|--------|-------|
| `AgentName` | `agent_name` | `agentName` from stdin | The dispatch name (e.g., `manager-spec`, `expert-backend`). |
| `AgentType` | `agent_type` | `agentType` from stdin | The subagent_type, often `general-purpose` for dynamic teammates. |
| `AgentID` | `agent_id` | `agent_id` from stdin (v2.1.69+) | Unique invocation ID. Useful for correlating with SubagentStart hook. |
| `ParentSessionID` | `parent_session_id` | `session.id` from stdin | The parent session (orchestrator) that spawned the subagent. |

`EventType` for SubagentStop entries: `subagent_stop` (new enum value).

`Subject` semantics for SubagentStop: the `AgentName` (e.g., `manager-spec`). This makes the downstream classifier naturally aggregate per-agent observation counts — the original PostToolUse observer used the same pattern (Subject = tool name).

**Edge case**: SubagentStop fires with no parent session_id (orphaned teammate after session-resume). Per AC-HRN-OBS-EDGE-003 (see acceptance.md), the observer still records the event with `parent_session_id` empty string; the downstream classifier handles the empty case.

### 3.3 UserPromptSubmit — PII handling (CRITICAL DECISION)

The `prompt` field is raw user content. Three observation strategies are explicitly considered:

#### Strategy A — Hash + Length + Language (RECOMMENDED DEFAULT)

| Field | JSON name | Value |
|-------|-----------|-------|
| `PromptHash` | `prompt_hash` | First 16 hex chars of SHA-256(prompt) — provides clustering signal without reversibility |
| `PromptLen` | `prompt_len` | Byte length of the prompt |
| `PromptLang` | `prompt_lang` | ISO-639-1 language code if detectable (e.g., `ko`, `en`); empty if undetectable |

Rationale: SHA-256 is one-way. Frequency clustering still works (identical prompts produce identical hashes). Length and language are weak signals that do not leak content. Privacy posture: STRONG — even if the JSONL is accidentally shared, raw user content cannot be recovered. Acceptable for default-on operation.

#### Strategy B — Preview Prefix + Hash + Length (OPT-IN)

| Field | JSON name | Value |
|-------|-----------|-------|
| `PromptPreview` | `prompt_preview` | First N (default N=64) bytes of the prompt |
| `PromptHash` | `prompt_hash` | Full SHA-256 as in Strategy A |
| `PromptLen` | `prompt_len` | Byte length |

Rationale: Provides debugging signal (operator can identify what the user actually typed) while still hashing for clustering. Privacy posture: WEAK — preview leaks the first N characters of every prompt, which often contains the user's intent. Acceptable only with explicit opt-in.

#### Strategy C — Full Content (OPT-IN, EXPLICIT)

| Field | JSON name | Value |
|-------|-----------|-------|
| `PromptContent` | `prompt_content` | The full raw prompt string |
| `PromptLen` | `prompt_len` | Byte length (for redundant consistency) |

Rationale: Maximum signal for downstream NLP classification (e.g., topic clustering, sentiment analysis). Privacy posture: NONE — full content recorded in plaintext. Acceptable only with explicit, separate opt-in via a dedicated config key.

#### Recommendation: Strategy A as default; B/C require explicit opt-in

The plan adopts Strategy A as the **default**. Strategies B and C are gated by a new config key `learning.user_prompt_content` in `.moai/config/sections/harness.yaml`:

| Config value | Strategy active |
|--------------|-----------------|
| absent (default) | Strategy A |
| `hash` (explicit confirm) | Strategy A |
| `preview` | Strategy B |
| `full` | Strategy C |
| `none` | Skip UserPromptSubmit observation entirely (event becomes no-op) |

The default-A choice is privacy-first and aligns with MoAI's local-first posture. Users who want richer telemetry must explicitly opt in.

`EventType` for UserPromptSubmit entries: `user_prompt` (new enum value).

`Subject` semantics for UserPromptSubmit: empty string, or the detected SPEC ID if the prompt mentions one via regex (`SPEC-[A-Z][A-Z0-9]+-[0-9]+`). Subject extraction does NOT record the full prompt — it captures only the SPEC ID match, which is itself a public identifier.

### 3.4 PostToolUse (existing — no schema change required)

V3R4-002 does NOT modify the PostToolUse observer's payload. The existing 4-field baseline (REQ-HRN-FND-010) holds. New fields added for the OTHER three events are `omitempty` and therefore absent from PostToolUse entries.

However, V3R4-002 DOES introduce a new `event_type` enum value to disambiguate: PostToolUse entries continue to use `event_type: agent_invocation` (preserving REQ-HRN-FND-010), while new event entries use `session_stop`, `subagent_stop`, `user_prompt`. The PostToolUse handler `runHarnessObserve` does not change.

---

## 4. Schema Additivity Policy

Per REQ-HRN-FND-010, the baseline 4-field contract (`timestamp`, `event_type`, `subject`, `context_hash`) must hold for every entry in `.moai/harness/usage-log.jsonl`. V3R4-002 adopts the following additivity rules:

1. **Optional fields**: All new fields introduced by V3R4-002 are tagged with Go `json:",omitempty"` so absent fields do not appear in the JSONL line. This means old PostToolUse entries (4-field) and new event entries (4+N fields) can coexist in the same log file.
2. **EventType discriminator**: The `event_type` field is the discriminator that tells downstream classifiers which payload shape to expect. The four new enum values (`session_stop`, `subagent_stop`, `user_prompt`, plus the existing `agent_invocation` for PostToolUse and the legacy `moai_subcommand`/`spec_reference`/`feedback`) form a closed set.
3. **No historical migration**: Per seed §1.3 Non-Goals item 7, V3R4-002 does NOT migrate or rewrite existing `usage-log.jsonl` entries. Pre-V3R4-002 entries remain valid under the baseline contract.
4. **Append-only**: The JSONL file is always opened with `O_APPEND|O_CREATE|O_WRONLY` (per `observer.go:78`). No in-place edits.
5. **SchemaVersion field**: The existing `SchemaVersion` field (`schema_version: "v1"`) is NOT bumped. V3R4-002 stays at v1 because schema EXTENSION (additive optional fields) does not require a version bump. A bump would be required only if the baseline 4 fields changed.

---

## 5. 5-Layer Safety Integration

The 5-Layer Safety architecture (REQ-HRN-FND-005, preserved from design constitution §5) applies UNCHANGED to V3R4-002. Each new event hook feeds the same downstream pipeline:

### Layer 1 — Frozen Guard

V3R4-002 introduces new REQ IDs (REQ-HRN-OBS-NNN) and new cobra subcommands. Neither weakens the FROZEN constitutional contract: the path-prefix protection (REQ-HRN-FND-006) is unchanged, and no new path is added to the FROZEN list. The Frozen Guard's behavior on observation pipeline writes is identical to PostToolUse (observer writes to `.moai/harness/usage-log.jsonl`, which is in the EVOLVABLE zone).

### Layer 2 — Canary Check

When a Tier-promotion is proposed (Tier 1 → 2 → 3 → 4), the Canary layer runs a shadow evaluation against the last 3 projects. V3R4-002 expands the observation surface, which means the Canary layer now sees event_type values it has not seen before (`session_stop`, `subagent_stop`, `user_prompt`). The Canary check itself does not need code changes — it operates on the aggregated Pattern objects, which are indexed by `(event_type, subject, context_hash)`. New event_type values produce new pattern buckets, and the Canary's score-drop > 0.10 rejection rule applies uniformly.

### Layer 3 — Contradiction Detector

The Contradiction Detector flags when a new learning conflicts with an existing rule. With V3R4-002, Stop and SubagentStop signals may sometimes contradict PostToolUse-derived tier promotions (e.g., a tool seen 10 times in PostToolUse but always followed by an immediate SubagentStop with failure code may indicate the pattern should NOT be promoted). The Contradiction Detector layer does not need code changes for V3R4-002 — it operates on rule text, which is unaffected by the underlying event source. The richer signal merely improves the Detector's accuracy in future iterations.

### Layer 4 — Rate Limiter

The existing weekly rate limit (≤3 evolutions/week, ≥24h cooldown, ≤50 active learnings) applies unchanged. The expanded observation surface MAY produce more tier-4 candidates per unit time, but the Rate Limiter cap is independent of how many candidates exist — it just throttles applications.

### Layer 5 — Human Oversight

The AskUserQuestion gate at Tier 4 (REQ-HRN-FND-004) is invoked by the MoAI orchestrator only. V3R4-002 does not introduce any new AskUserQuestion call sites. Subagents (including any new harness-related subagents) remain prohibited from calling AskUserQuestion (REQ-HRN-FND-015).

### Net result

V3R4-002 increases observation coverage WITHOUT modifying any safety layer's behavior, threshold, or contract. The 5-Layer Safety preservation requirement (REQ-HRN-FND-005) is satisfied by construction.

---

## 6. Risks and Open Questions

### 6.1 Risks

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| PII leakage via UserPromptSubmit if Strategy A is bypassed by a misconfiguration | Low | High | Strategy A is the default; opt-in for B/C is gated by an explicit config key. New REQ requires that the loader rejects unknown values with a fail-open to Strategy A (never escalate to weaker privacy). |
| Performance regression — three new JSONL writes per turn (Stop) + one per subagent (SubagentStop) + one per user prompt (UserPromptSubmit) | Medium | Low | Each Observer.RecordEvent is <100ms (REQ-HL-001 baseline). Three additional writes add ~300ms per turn worst case. Lazy retention pruning is bounded. Per-event observer is fail-closed on error (returns nil to keep hook non-blocking). |
| Settings.json.tmpl conflict with concurrent SPECs editing the same template | Medium | Medium | V3R4-002 strategy WIRE-A adds NEW entries (additive) rather than modifying existing entries. Concurrent SPECs that don't touch the Stop / SubagentStop / UserPromptSubmit slots cannot conflict. |
| Cobra subcommand naming conflict — V3R4-002 wants `harness-observe-stop`, `harness-observe-subagent-stop`, `harness-observe-user-prompt-submit` but `harness-observe` already exists | Low | Low | New subcommands use a different prefix sequence. The existing `harness-observe` (no-suffix) remains the PostToolUse handler. New subcommands are explicit. |
| Test isolation — three new observer code paths each need NoOp/Records/PreservesExisting test triple | Low | Low | Pattern reuse from `hook_harness_observe_test.go` (already in tree). Total new test cases ≈ 9 (3 events × 3 cases). Table-driven pattern keeps LOC low. |
| Backward compat with users on `learning.enabled: false` who never opted in to new events | Low | Low | Gate function `isHarnessLearningEnabled` is shared. All four event handlers call the same gate. Disabled = silent no-op for all. |
| Template-first discipline: V3R4-002 must edit `internal/template/templates/.claude/settings.json.tmpl` AND `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl` for any new wrapper, then run `make build` | Low | Medium | Pre-merge CI runs `make build` and verifies `internal/template/embedded.go` is up to date. Documented in CLAUDE.local.md §2 "Template-First Rule". |

### 6.2 Open Questions (forward to plan-auditor / orchestrator)

| OQ | Question | Default answer (proposed) | Decision required from |
|----|----------|--------------------------|------------------------|
| OQ-1 | Should the new EventType enum values (`session_stop`, `subagent_stop`, `user_prompt`) be added to `internal/harness/types.go`, or should the observer reuse `agent_invocation` for all four events with payload disambiguation via the new optional fields? | Add new enum values. Cleaner downstream classifier; aligns Subject semantics per event. | Plan-auditor; orchestrator if disagreement |
| OQ-2 | Should the per-event observer logic live in three separate cobra subcommands (e.g., `harness-observe-stop`), or in a single subcommand with a `--event <name>` flag? | Three separate subcommands. Matches existing per-event handler pattern (`moai hook stop`, `moai hook subagent-stop`). Easier wiring in settings.json.tmpl per event slot. | Plan-auditor; orchestrator |
| OQ-3 | Should V3R4-002 also retroactively wire `handle-harness-observe.sh` into the PostToolUse settings.json.tmpl slot (closing the V3R4-001 gap identified in §2.5), or leave PostToolUse wiring for a separate follow-up? | Leave PostToolUse wiring as a V3R4-001 follow-up. V3R4-002 scope is observer EXPANSION; PostToolUse baseline registration is the foundation SPEC's responsibility. | Orchestrator |
| OQ-4 | What is the default value of N in Strategy B preview prefix? | N = 64 bytes. Empirically captures the first sentence in most cases without leaking full content. | Plan-auditor; user can adjust via config |
| OQ-5 | If a user explicitly opts into Strategy C (full content), should there be a banner warning at session start? | Yes — emit one banner per session at the first UserPromptSubmit observation. This is a user-experience choice; the warning emits through `systemMessage` in the hook stdout. | Plan-auditor; deferred to run-phase implementation detail |
| OQ-6 | Should the `language` field in Strategy A use a heuristic (e.g., first-character Unicode block) or invoke a dedicated language-detection library? | Heuristic only (Unicode block, no dependency). Library imports are out-of-scope for observer hot path. | Plan-auditor |

---

## 7. Cross-References

- SPEC-V3R4-HARNESS-001 spec.md (§5 REQ-HRN-FND-005, REQ-HRN-FND-009, REQ-HRN-FND-010, REQ-HRN-FND-011, REQ-HRN-FND-015): the foundation contracts that V3R4-002 preserves verbatim.
- SPEC-V3R4-HARNESS-001 acceptance.md AC-HRN-FND-007, AC-HRN-FND-008: the no-op / baseline-observation patterns V3R4-002 inherits and extends.
- `internal/cli/hook.go:439` (isHarnessLearningEnabled), `:474` (runHarnessObserve), `:94-99` (harness-observe cobra registration).
- `internal/harness/observer.go:18` (Observer struct), `:53` (RecordEvent).
- `internal/harness/types.go:36` (Event struct), `:15` (EventType enum).
- `internal/cli/hook_harness_observe_test.go` (test patterns to clone).
- `.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout Reference.
- `.claude/rules/moai/design/constitution.md` §5 (5-Layer Safety, FROZEN), §2 (Frozen vs Evolvable).
- `internal/template/templates/.claude/settings.json.tmpl` — hook registration template (Template-First per CLAUDE.local.md §2).
- `internal/template/templates/.claude/hooks/moai/handle-*.sh.tmpl` — wrapper script templates.

End of research.md.

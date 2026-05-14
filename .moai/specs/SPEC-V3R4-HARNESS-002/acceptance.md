# Acceptance Criteria — SPEC-V3R4-HARNESS-002

This document defines the acceptance criteria for the Multi-Event Observer Expansion SPEC. Each AC corresponds to one or more REQ-HRN-OBS-NNN requirements in `spec.md` §6. Every leaf AC uses Given / When / Then format and includes an objectively verifiable Verification block. Where multiple variations of the same scenario differ only in input parameters, the hierarchical AC schema is used (`AC-...-NN.a`, `AC-...-NN.b`) per `.claude/rules/moai/workflow/spec-workflow.md` § Hierarchical Acceptance Criteria Schema.

---

## REQ → AC Coverage Matrix

| REQ ID | AC ID | Description |
|--------|-------|-------------|
| REQ-HRN-OBS-001 | AC-HRN-OBS-001.a | Cobra subcommand `harness-observe-stop` registered |
| REQ-HRN-OBS-002 | AC-HRN-OBS-001.b | Cobra subcommand `harness-observe-subagent-stop` registered |
| REQ-HRN-OBS-003 | AC-HRN-OBS-001.c | Cobra subcommand `harness-observe-user-prompt-submit` registered |
| REQ-HRN-OBS-004 | AC-HRN-OBS-002 | Stop observation entry with extended fields |
| REQ-HRN-OBS-005 | AC-HRN-OBS-003 | SubagentStop observation entry with extended fields |
| REQ-HRN-OBS-006 | AC-HRN-OBS-004 | UserPromptSubmit observation entry (Strategy A default) |
| REQ-HRN-OBS-007 | AC-HRN-OBS-012 | 5-Layer Safety preservation (architectural assertion) |
| REQ-HRN-OBS-008 | AC-HRN-OBS-005 | Unified gate — all 4 handlers no-op when learning disabled |
| REQ-HRN-OBS-009 | AC-HRN-OBS-006 | Schema additivity — old entries remain valid |
| REQ-HRN-OBS-010 | AC-HRN-OBS-012 | 4-tier ladder preservation (architectural assertion) |
| REQ-HRN-OBS-011 | AC-HRN-OBS-012 | Subagent AskUserQuestion prohibition preservation |
| REQ-HRN-OBS-012 | AC-HRN-OBS-007 | PII privacy default (Strategy A) |
| REQ-HRN-OBS-013 | AC-HRN-OBS-008 | UserPromptSubmit opt-in strategies (B, C, none) |
| REQ-HRN-OBS-014 | AC-HRN-OBS-009 | Invalid PII config fails open to Strategy A |
| REQ-HRN-OBS-015 | AC-HRN-OBS-010 | EventType enum extension (3 new values) |
| REQ-HRN-OBS-016 | AC-HRN-OBS-011.a | Wrapper script template files exist with canonical pattern |
| REQ-HRN-OBS-017 | AC-HRN-OBS-011.b | settings.json.tmpl additive registration |
| REQ-HRN-OBS-018 | AC-HRN-OBS-013 | Hook latency budget (100ms) |

Coverage: 18 REQs ↔ 13 ACs (parent counts; leaf nodes per hierarchical schema = 17). Every REQ appears in at least one AC.

---

## Acceptance Criteria

### AC-HRN-OBS-001 — Cobra Subcommand Registration (parent)

**Linked REQs**: REQ-HRN-OBS-001, REQ-HRN-OBS-002, REQ-HRN-OBS-003

**Given (shared parent context)** the moai-adk-go binary built from the V3R4-002 implementation branch with all three new cobra subcommands registered in `hookCmd.Commands()`,

#### AC-HRN-OBS-001.a — Stop subcommand

**When** the user runs `moai hook harness-observe-stop --help` from the terminal,

**Then** the system MUST return cobra's help text describing the subcommand including a `Short` description referencing "Stop" and "harness usage log". The exit code MUST be 0. (maps REQ-HRN-OBS-001)

**Verification**: `moai hook harness-observe-stop --help 2>&1 | head -3` includes the word "Stop"; `echo $?` returns 0.

#### AC-HRN-OBS-001.b — SubagentStop subcommand

**When** the user runs `moai hook harness-observe-subagent-stop --help`,

**Then** the system MUST return help text referencing "SubagentStop" / "subagent". Exit code 0. (maps REQ-HRN-OBS-002)

**Verification**: `moai hook harness-observe-subagent-stop --help 2>&1 | head -3` includes "subagent" or "SubagentStop"; `echo $?` returns 0.

#### AC-HRN-OBS-001.c — UserPromptSubmit subcommand

**When** the user runs `moai hook harness-observe-user-prompt-submit --help`,

**Then** the system MUST return help text referencing "UserPromptSubmit" / "user prompt". Exit code 0. (maps REQ-HRN-OBS-003)

**Verification**: `moai hook harness-observe-user-prompt-submit --help 2>&1 | head -3` includes "prompt" or "UserPromptSubmit"; `echo $?` returns 0.

---

### AC-HRN-OBS-002 — Stop Observation Entry With Extended Fields

**Linked REQs**: REQ-HRN-OBS-004

**Given** a project with `.moai/config/sections/harness.yaml` containing `learning.enabled: true` and an empty (or absent) `.moai/harness/usage-log.jsonl`,

**When** the Claude Code runtime fires the `Stop` hook and the wrapper script invokes `moai hook harness-observe-stop` with a representative stdin JSON containing:

```json
{
  "last_assistant_message": "Done. PR #911 created.",
  "session": {"id": "sess-abc-123", "cwd": "/tmp/proj", "projectDir": "/tmp/proj"}
}
```

**Then** the system MUST append exactly ONE JSONL line to `.moai/harness/usage-log.jsonl`. The line MUST contain:
- the four baseline fields per REQ-HRN-FND-010: `timestamp` (ISO-8601 UTC), `event_type` (value `session_stop`), `subject` (empty string OR a detected SPEC ID), `context_hash` (empty string OK).
- the optional extended fields populated: `session_id: "sess-abc-123"`, `last_assistant_message_hash: <16-hex-char-sha256-prefix>`, `last_assistant_message_len: 22`.
- the raw `last_assistant_message` text MUST NOT appear anywhere in the entry. (maps REQ-HRN-OBS-004)

**Verification**:
- Set up test fixture with `t.TempDir() + writeHarnessYAML(learning.enabled: true)`.
- Call `runHarnessObserveStop(cmd, nil)` with stdin piped JSON.
- Read `.moai/harness/usage-log.jsonl` and JSON-unmarshal the single line.
- Assert: `event_type == "session_stop"`, `session_id == "sess-abc-123"`, `last_assistant_message_len == 22`, `last_assistant_message_hash` is a 16-char hex string, raw text `"Done."` does NOT appear in the file contents.

---

### AC-HRN-OBS-003 — SubagentStop Observation Entry With Extended Fields

**Linked REQs**: REQ-HRN-OBS-005

**Given** a project with `learning.enabled: true`,

**When** the SubagentStop hook fires with stdin JSON:

```json
{
  "agentType": "general-purpose",
  "agentName": "manager-spec",
  "last_assistant_message": "spec.md drafted.",
  "agent_id": "agent-xyz-456",
  "agent_transcript_path": "/tmp/transcript-xyz.txt",
  "session": {"id": "sess-parent-789"}
}
```

**Then** the system MUST append exactly ONE JSONL line containing:
- baseline 4 fields: `timestamp`, `event_type == "subagent_stop"`, `subject == "manager-spec"` (the agentName), `context_hash == ""` (or a derived hash if implementation chooses).
- optional extended fields: `agent_name == "manager-spec"`, `agent_type == "general-purpose"`, `agent_id == "agent-xyz-456"`, `parent_session_id == "sess-parent-789"`. (maps REQ-HRN-OBS-005)

**Verification**: Test fixture calls `runHarnessObserveSubagentStop(cmd, nil)` with piped JSON. Assert all required fields present; `subject` field equals `agentName`; raw `last_assistant_message` ("spec.md drafted.") does NOT appear in JSONL.

---

### AC-HRN-OBS-004 — UserPromptSubmit Observation Entry (Strategy A Default)

**Linked REQs**: REQ-HRN-OBS-006

**Given** a project with `learning.enabled: true` AND `learning.user_prompt_content` absent from harness.yaml (so default Strategy A applies),

**When** the UserPromptSubmit hook fires with stdin JSON:

```json
{
  "prompt": "안녕하세요. SPEC-V3R4-HARNESS-002 plan 부탁드립니다.",
  "session": {"id": "sess-prompt-1"}
}
```

**Then** the system MUST append exactly ONE JSONL line containing:
- baseline 4 fields: `event_type == "user_prompt"`, `subject == "SPEC-V3R4-HARNESS-002"` (extracted via regex from the prompt, OR empty string if no SPEC ID matches), `context_hash == ""`.
- Strategy A fields: `prompt_hash` (16 hex chars from SHA-256), `prompt_len` (byte length of the prompt), `prompt_lang == "ko"` (Korean detected by Unicode block heuristic).
- The raw prompt text "안녕하세요..." MUST NOT appear anywhere in the JSONL entry. Specifically: `prompt_content` MUST be absent, `prompt_preview` MUST be absent. (maps REQ-HRN-OBS-006)

**Verification**: Test fixture calls `runHarnessObserveUserPromptSubmit(cmd, nil)` with piped JSON. Assert: presence of `prompt_hash`, `prompt_len`, `prompt_lang` fields; absence of `prompt_content` and `prompt_preview` fields; `grep -c "안녕하세요" usage-log.jsonl` returns 0.

---

### AC-HRN-OBS-005 — Unified Gate (All 4 Handlers No-Op When Disabled)

**Linked REQs**: REQ-HRN-OBS-008

**Given** a project with `.moai/config/sections/harness.yaml` containing `learning.enabled: false`,

**When** all four observer handlers (`runHarnessObserve`, `runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) are invoked in sequence, each with a representative stdin JSON,

**Then** for EACH handler:
- The function MUST return nil (no error).
- The file `.moai/harness/usage-log.jsonl` MUST NOT be created (`os.Stat` returns `IsNotExist`).
- stderr output MUST be empty.

**And** if `.moai/harness/usage-log.jsonl` was pre-seeded with two baseline-format entries before the test, the file contents AFTER all four handler invocations MUST be byte-identical to the pre-test contents (no append, no truncation, no rewrite). (maps REQ-HRN-OBS-008)

**Verification**: Table-driven Go test `TestHookHarnessGateUniformity` in `internal/cli/hook_harness_gate_uniformity_test.go` iterates over the four handler functions, invokes each with `learning.enabled: false`, asserts file absence; then re-runs with pre-seeded log and asserts byte-identical preservation.

---

### AC-HRN-OBS-006 — Schema Additivity (Old Entries Remain Valid)

**Linked REQs**: REQ-HRN-OBS-009

**Given** a `.moai/harness/usage-log.jsonl` file containing two pre-V3R4-002 entries with only the baseline 4 fields:

```jsonl
{"timestamp":"2026-05-13T10:00:00Z","event_type":"agent_invocation","subject":"Edit","context_hash":""}
{"timestamp":"2026-05-13T10:05:00Z","event_type":"agent_invocation","subject":"Bash","context_hash":""}
```

**When** the V3R4-002 implementation reads the log file (via any tier-classification or aggregation code path),

**Then** the read MUST succeed for all entries without error; each entry MUST parse into the extended `Event` struct with the new optional fields defaulting to their Go zero values (empty strings, zero ints). No JSON unmarshal error MUST occur.

**And When** a new Stop entry is appended to the same file by `runHarnessObserveStop`,

**Then** the resulting file MUST contain THREE lines: the two original baseline entries followed by ONE new entry with extended fields. The two original lines MUST be byte-identical to their pre-append state. (maps REQ-HRN-OBS-009)

**Verification**: Test fixture writes the two baseline entries, runs a new observer with `learning.enabled: true`, then reads the file and asserts: total line count 3, first two lines unchanged, third line is the new event with optional fields populated.

---

### AC-HRN-OBS-007 — PII Privacy Default (Strategy A)

**Linked REQs**: REQ-HRN-OBS-012

**Given** a project with `learning.enabled: true` AND `learning.user_prompt_content` ABSENT from harness.yaml (no opt-in),

**When** the UserPromptSubmit hook fires with a stdin JSON containing any non-empty prompt text (e.g., `"Hello World"`),

**Then** the system MUST apply Strategy A: the JSONL entry MUST contain `prompt_hash` + `prompt_len` + `prompt_lang` ONLY. The fields `prompt_preview` and `prompt_content` MUST be absent (omitted via `omitempty`).

**And** the raw prompt text MUST NOT appear in the JSONL entry under any field name. (maps REQ-HRN-OBS-012)

**Verification**: Test fixture sets `learning.enabled: true` only (no `user_prompt_content` key). Invokes handler. Reads JSONL. Asserts: `prompt_hash` present, `prompt_preview` absent, `prompt_content` absent. `grep "Hello World" usage-log.jsonl` returns 0 matches.

---

### AC-HRN-OBS-008 — UserPromptSubmit Opt-In Strategies (parent)

**Linked REQs**: REQ-HRN-OBS-013

**Given (shared parent context)** a project with `learning.enabled: true` and the UserPromptSubmit handler invoked with a stdin prompt of "Hello world test prompt" (23 bytes),

#### AC-HRN-OBS-008.a — Strategy B (preview)

**When** `.moai/config/sections/harness.yaml` contains `learning.user_prompt_content: preview`,

**Then** the JSONL entry MUST contain `prompt_hash`, `prompt_len`, `prompt_lang`, AND `prompt_preview` (first 64 bytes, or full content if prompt is shorter). `prompt_content` MUST be absent. (maps REQ-HRN-OBS-013)

**Verification**: Test sets config to `preview`. Asserts: `prompt_preview == "Hello world test prompt"` (the full prompt because it is shorter than 64 bytes); `prompt_content` field absent.

#### AC-HRN-OBS-008.b — Strategy C (full)

**When** `learning.user_prompt_content: full`,

**Then** the JSONL entry MUST contain `prompt_hash`, `prompt_len`, `prompt_lang`, AND `prompt_content` (the full raw prompt text). (maps REQ-HRN-OBS-013)

**Verification**: Test sets config to `full`. Asserts: `prompt_content == "Hello world test prompt"`.

#### AC-HRN-OBS-008.c — Strategy "none" (skip)

**When** `learning.user_prompt_content: none`,

**Then** the UserPromptSubmit handler MUST NOT append any JSONL entry. The file `.moai/harness/usage-log.jsonl` MUST remain absent (or, if pre-existing, MUST remain byte-identical). The handler MUST still return nil (no error). (maps REQ-HRN-OBS-013)

**Verification**: Test sets config to `none`, asserts file absence after handler invocation.

---

### AC-HRN-OBS-009 — Invalid PII Config Fails Open to Strategy A

**Linked REQs**: REQ-HRN-OBS-014

**Given** a project with `learning.enabled: true` AND `.moai/config/sections/harness.yaml` containing `learning.user_prompt_content: garbage_value_xyz` (a string not in the set `{hash, preview, full, none}`),

**When** the UserPromptSubmit handler fires with a stdin prompt,

**Then** the system MUST fall back to Strategy A (the most privacy-preserving). The JSONL entry MUST contain `prompt_hash` + `prompt_len` + `prompt_lang` only. The fields `prompt_preview` and `prompt_content` MUST be absent. The raw prompt MUST NOT appear in the entry.

**And** the handler MAY emit a non-blocking warning to stderr indicating the unknown config value was rejected. The handler MUST still return nil and MUST NOT block the Claude Code pipeline. (maps REQ-HRN-OBS-014)

**Verification**: Test sets `learning.user_prompt_content: garbage_value_xyz`. Asserts Strategy A field presence; absence of preview/content; raw prompt not in JSONL. Optional: stderr capture contains a warning string.

---

### AC-HRN-OBS-010 — EventType Enum Extension (3 New Values)

**Linked REQs**: REQ-HRN-OBS-015

**Given** the V3R4-002 implementation merged into `internal/harness/types.go`,

**When** a developer (or automated test) inspects the `EventType` constants block,

**Then** the file MUST contain the original four constants (`EventTypeMoaiSubcommand`, `EventTypeAgentInvocation`, `EventTypeSpecReference`, `EventTypeFeedback`) AND three new constants: `EventTypeSessionStop = "session_stop"`, `EventTypeSubagentStop = "subagent_stop"`, `EventTypeUserPrompt = "user_prompt"`. (maps REQ-HRN-OBS-015)

**Verification**: Static check — `grep -nE 'EventType(SessionStop|SubagentStop|UserPrompt) = ' internal/harness/types.go` returns 3 matches. Go test asserts each new constant evaluates to its expected string literal.

---

### AC-HRN-OBS-011 — Wrapper Scripts and settings.json.tmpl Registration (parent)

**Linked REQs**: REQ-HRN-OBS-016, REQ-HRN-OBS-017

**Given (shared parent context)** the V3R4-002 implementation merged with template-first compliance,

#### AC-HRN-OBS-011.a — Wrapper script template files exist

**When** a developer inspects `internal/template/templates/.claude/hooks/moai/`,

**Then** the directory MUST contain three new files:
- `handle-harness-observe-stop.sh.tmpl`
- `handle-harness-observe-subagent-stop.sh.tmpl`
- `handle-harness-observe-user-prompt-submit.sh.tmpl`

**And** each file MUST follow the existing 33-line canonical pattern from `handle-harness-observe.sh.tmpl`: stderr log path setup, 10MB rotation guard, binary search in PATH → `~/go/bin/moai` → `~/.local/bin/moai`, `exec moai hook <subcommand>`, silent exit 0 if binary not found.

**And** `internal/template/embedded.go` MUST be regenerated such that the new templates are accessible at runtime. (maps REQ-HRN-OBS-016)

**Verification**: `ls internal/template/templates/.claude/hooks/moai/handle-harness-observe-*.sh.tmpl | wc -l` returns 4 (the original + 3 new). `grep -c "moai hook harness-observe-stop" internal/template/templates/.claude/hooks/moai/handle-harness-observe-stop.sh.tmpl` returns at least 1. `make build` succeeds.

#### AC-HRN-OBS-011.b — settings.json.tmpl additive registration

**When** a developer diffs `internal/template/templates/.claude/settings.json.tmpl` against its pre-V3R4-002 state,

**Then** the diff MUST show ADDITIONS only to the `Stop`, `SubagentStop`, and `UserPromptSubmit` hook event slots — one new hook entry per slot pointing to the corresponding `handle-harness-observe-*.sh` wrapper. NO existing lines (the entries pointing to `handle-stop.sh`, `handle-subagent-stop.sh`, `handle-user-prompt-submit.sh`) MUST be modified. (maps REQ-HRN-OBS-017)

**Verification**: `git diff main -- internal/template/templates/.claude/settings.json.tmpl | grep -E '^[-][^-]' | grep -v 'handle-harness-observe'` returns zero output (no lines deleted EXCEPT lines that were re-inserted as part of harness-observe additions). Manual review by plan-auditor.

---

### AC-HRN-OBS-012 — V3R4-001 Contract Preservation (architectural assertion)

**Linked REQs**: REQ-HRN-OBS-007, REQ-HRN-OBS-010, REQ-HRN-OBS-011

**Given** the V3R4-002 PR diff against `main`,

**When** the diff is examined,

**Then** the following files MUST NOT appear in the diff (they are FROZEN per V3R4-001):
- `.claude/rules/moai/design/constitution.md` (§5 5-Layer Safety unchanged)
- `.claude/rules/moai/core/agent-common-protocol.md` (subagent prohibition unchanged)
- `.claude/rules/moai/core/askuser-protocol.md` (orchestrator-only contract unchanged)

**And** the four V3R4-001 functions/constants MUST NOT be modified in V3R4-002:
- `internal/cli/hook.go::isHarnessLearningEnabled` — function body unchanged (only callers added)
- `internal/cli/hook.go::runHarnessObserve` — function body unchanged
- `internal/harness/observer.go::Observer.RecordEvent` — function body unchanged (struct fields extended but write path unchanged)
- `internal/harness/types.go::LogSchemaVersion` — constant value remains `"v1"`

**And** no new AskUserQuestion call site MUST appear in any modified file (`grep -rn 'AskUserQuestion' internal/cli/hook.go internal/harness/` returns zero matches in V3R4-002-added code).

**And** no new file path MUST be added to the FROZEN zone path-prefix list (the list in REQ-HRN-FND-006 is unchanged). (maps REQ-HRN-OBS-007, REQ-HRN-OBS-010, REQ-HRN-OBS-011)

**Verification**: Automated git-diff inspection: `git diff main..HEAD --name-only | grep -E '(design/constitution.md|agent-common-protocol.md|askuser-protocol.md)$'` returns zero matches. Static grep: `grep -A 5 'func isHarnessLearningEnabled' internal/cli/hook.go` shows the function body is identical to V3R4-001's merged version. Manual plan-auditor review confirms.

---

### AC-HRN-OBS-013 — Hook Latency Budget (100ms per handler)

**Linked REQs**: REQ-HRN-OBS-018

**Given** the four observer handlers compiled into the moai binary, run on a typical developer workstation (no extreme load),

**When** each handler is invoked with a representative stdin payload and `learning.enabled: true` (no retention pruning),

**Then** the wall-clock time for the handler's `Observer.RecordEvent` call MUST complete within 100ms at the 95th percentile across 100 invocations. (maps REQ-HRN-OBS-018)

**Verification**: Benchmark Go test `BenchmarkHookHarnessObserveLatency` in `internal/cli/hook_harness_latency_bench_test.go` measures per-handler wall-clock time. Report 50th, 95th, 99th percentile. The 95th percentile MUST be < 100ms. Test reports the values without failing the build if exceeded (informational); the latency budget is enforced as a quality gate in the run-phase Wave C documentation, not as a CI-blocking assertion (since latency can vary by machine).

---

## Edge Cases

### EDGE-001 — Disk-full at hook fire time

If the operating system reports `ENOSPC` when the observer attempts to append to `.moai/harness/usage-log.jsonl`, the handler MUST:
- log a non-blocking warning to stderr (using `cmd.ErrOrStderr()`),
- return nil (NOT propagate the error) so the Claude Code hook pipeline remains non-blocking,
- NOT leave a partial JSONL line in the log file.

### EDGE-002 — settings.json.tmpl missing one of the new hook entries

If a developer accidentally removes one of the three new hook entries from `settings.json.tmpl` (e.g., during rebase), the rendered settings.json will simply not invoke the corresponding wrapper script. Claude Code handles missing hook entries gracefully; the observer for that event becomes silently inactive. The CI guard for this case is OUT OF SCOPE for V3R4-002; the responsibility falls on PR review.

### EDGE-003 — SubagentStop fires with empty parent_session_id (orphaned teammate after session-resume)

If `session.id` is absent or empty in the SubagentStop stdin JSON (e.g., post-resume orphan), the observer MUST still record the entry with `parent_session_id` as empty string (the field is `omitempty` so it will be absent from the JSONL line, which is acceptable — the downstream classifier handles the empty case).

### EDGE-004 — UserPromptSubmit with empty prompt

If the user submits an empty prompt (zero bytes), the observer MUST:
- compute `prompt_hash` as SHA-256("") → first 16 hex chars (deterministic),
- record `prompt_len: 0`,
- record `prompt_lang: ""` (empty since no characters to detect),
- still append exactly ONE JSONL entry (the event is still observed).

### EDGE-005 — Concurrent Stop + SubagentStop (race on JSONL append)

If Stop and SubagentStop fire near-simultaneously (e.g., a subagent exits at the same moment the main turn closes), the JSONL append uses `O_APPEND|O_CREATE|O_WRONLY` semantics. On POSIX-compliant filesystems, `O_APPEND` writes are atomic at the line level for writes ≤ `PIPE_BUF` (4096 bytes on Linux/macOS). Both entries MUST land in the log file as separate, complete JSONL lines. No interleaving or truncation MUST occur.

**Verification**: The existing `Observer.RecordEvent` in `internal/harness/observer.go:78` already uses this pattern. V3R4-002 reuses the same write path, so concurrent safety is inherited.

### EDGE-006 — `learning.user_prompt_content: full` opt-in followed by user opting back out

If the user transitions from `full` to `hash` (or removes the key entirely), the behavior MUST change immediately on the next handler invocation (no caching of the previous strategy). Existing log entries (which may contain `prompt_content`) are NOT retroactively redacted by V3R4-002; users who want to clear historical content must manually delete the file.

### EDGE-007 — harness.yaml YAML parse error (corrupt file)

If `.moai/config/sections/harness.yaml` is malformed YAML, the existing `isHarnessLearningEnabled` returns `true` (fail-open). The new strategy resolver `resolveUserPromptStrategy` MUST follow the same pattern: on YAML parse error, fall back to Strategy A (the strongest privacy default), NOT to "none" or "full".

---

## Test Scenarios (Given-When-Then)

### Scenario 1: Full Turn Lifecycle Observation (Happy Path)

**Given** a project with `learning.enabled: true` and an empty `.moai/harness/usage-log.jsonl`,

**When** a user submits a prompt "/moai plan SPEC-EXAMPLE-001", Claude spawns `manager-spec` as an Agent(), `manager-spec` writes one file via Edit tool, then completes; Claude's main turn then ends,

**Then** the resulting `.moai/harness/usage-log.jsonl` MUST contain (in order) four entries:
1. `event_type: user_prompt`, `subject: "SPEC-EXAMPLE-001"` (extracted from prompt), `prompt_hash` + `prompt_len` + `prompt_lang` populated.
2. `event_type: agent_invocation`, `subject: "Edit"` (the PostToolUse observer from V3R4-001 — unmodified).
3. `event_type: subagent_stop`, `subject: "manager-spec"`, `agent_name: "manager-spec"`.
4. `event_type: session_stop`, `subject: ""` (or detected SPEC ID from cwd), `session_id` populated.

---

### Scenario 2: Privacy-Conscious User (Default Strategy A)

**Given** a privacy-conscious user with `learning.enabled: true` and NO `learning.user_prompt_content` key,

**When** the user types a prompt containing sensitive content like a personal email address,

**Then** the JSONL entry for that prompt MUST NOT contain the email address (or any substring of it). Only `prompt_hash`, `prompt_len`, `prompt_lang` MUST be recorded.

---

### Scenario 3: Power User Opt-In to Strategy C (Full Content)

**Given** a power user who explicitly sets `learning.user_prompt_content: full` to receive the richest downstream classification signal,

**When** the user submits any prompt,

**Then** the JSONL entry MUST contain `prompt_content` with the full raw prompt text in addition to the Strategy A fields. This is the user's informed consent.

---

### Scenario 4: Configuration Drift (Invalid Value)

**Given** a `.moai/config/sections/harness.yaml` file edited by a developer who typo'd `learning.user_prompt_content: fulll` (extra "l"),

**When** the UserPromptSubmit handler fires,

**Then** the system MUST fall back to Strategy A (REQ-HRN-OBS-014). The raw prompt MUST NOT be recorded. The handler MAY emit a one-line stderr warning identifying the invalid value (for developer debugging).

---

### Scenario 5: Pre-V3R4-002 Log Compatibility

**Given** a project that was on V3R4-001 (PostToolUse-only observer) and has accumulated 152 baseline-format entries in `.moai/harness/usage-log.jsonl`,

**When** the user upgrades to V3R4-002 and starts a new Claude Code session,

**Then** the existing 152 entries MUST remain in place, unmodified. New observations from Stop / SubagentStop / UserPromptSubmit hooks MUST be appended after the existing entries. The tier classifier (REQ-HRN-FND-011, unchanged) MUST continue to read all entries (old and new) without parse errors.

---

### Scenario 6: Subagent Blocker Report Instead of AskUserQuestion (Contract Preservation)

**Given** a future scenario where the `manager-spec` subagent is invoked to draft a new SPEC and encounters ambiguity requiring user input,

**When** the subagent runs,

**Then** the subagent MUST return a structured blocker report (per REQ-HRN-OBS-011 → REQ-HRN-FND-015) and MUST NOT invoke `AskUserQuestion`. The orchestrator handles the user interaction. V3R4-002 introduces no exception to this contract.

---

### Scenario 7: 5-Layer Safety Preservation Audit

**Given** the V3R4-002 plan PR ready for merge,

**When** a developer or plan-auditor inspects the PR,

**Then** the diff MUST NOT include any modification of `.claude/rules/moai/design/constitution.md`. The 5-Layer Safety section (§5) MUST be byte-identical between `main` and `HEAD`. The FROZEN zone path-prefix list (REQ-HRN-FND-006) MUST be unchanged.

---

## Definition of Done

This SPEC's plan-phase deliverables are considered complete when ALL of the following are satisfied:

1. **All five SPEC artifacts present**: `research.md`, `spec.md`, `plan.md`, `acceptance.md`, `tasks.md` exist under `.moai/specs/SPEC-V3R4-HARNESS-002/` with the canonical 9-field frontmatter on `spec.md`.
2. **All 18 REQs in EARS format**: every REQ in `spec.md` §6 uses one of the five EARS patterns.
3. **All 13 ACs verifiable**: every AC includes Given-When-Then plus a Verification section with concrete commands or assertion descriptions.
4. **Coverage map complete**: every REQ appears in at least one AC; every AC links back to at least one REQ.
5. **No tech-stack implementation in spec.md**: file paths, function bodies, exact struct definitions are in plan.md only.
6. **No modification of FROZEN files**: `.claude/rules/moai/design/constitution.md`, `.claude/rules/moai/core/agent-common-protocol.md`, `.claude/rules/moai/core/askuser-protocol.md` unchanged in this PR.
7. **V3R4-001 contracts preserved verbatim**: REQ-HRN-FND-005, REQ-HRN-FND-009, REQ-HRN-FND-010, REQ-HRN-FND-011, REQ-HRN-FND-015 all referenced in spec.md §1.2 and not weakened.
8. **No emojis** in artifact content.
9. **No time estimates**: priority labels (P0-P3) and phase ordering only.
10. **Plan-auditor PASS**: independent subagent audit returns PASS verdict (or unresolvable findings escalated as blocker).
11. **PR delegated to `manager-git`**: this manager-spec session does NOT create the PR directly; the orchestrator (parent agent) delegates to `manager-git` via `Agent()`.
12. **Conventional Commits**: commit message follows `plan(SPEC-V3R4-HARNESS-002): multi-event observer expansion plan` format.
13. **PII handling explicit**: Strategy A is documented as default; opt-in mechanism for B, C, none is documented in spec.md §6 and tested in acceptance.md AC-HRN-OBS-007/008/009.

---

End of acceptance.md.

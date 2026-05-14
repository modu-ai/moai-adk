# Tasks — SPEC-V3R4-HARNESS-002

This document is the task-level breakdown of the implementation plan for the Multi-Event Observer Expansion SPEC. Tasks are organized by Wave (A through C) and use IDs `T-A1` through `T-C5`. Each task lists its linked REQ IDs, linked AC IDs, MX tag implications, complexity (XS/S/M/L), file ownership, dependencies, and Wave membership.

All priorities use P0/P1/P2/P3 labels; no time estimates are used per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation.

---

## Task Summary

| Wave | Tasks | Priority Distribution | Total |
|------|-------|-----------------------|-------|
| Wave A — Schema Extension + Cobra Subcommand Scaffolding | T-A1, T-A2, T-A3, T-A4, T-A5 | P0: 5 | 5 |
| Wave B — Wrapper Script Templates + settings.json.tmpl Registration | T-B1, T-B2, T-B3, T-B4 | P0: 4 | 4 |
| Wave C — Test Suite + 5-Layer Safety Integration Verification | T-C1, T-C2, T-C3, T-C4, T-C5 | P1: 5 | 5 |
| **Total** | | **P0: 9, P1: 5** | **14** |

Coverage target (per `quality.yaml` baseline): Wave A ≥ 85%, Wave B (template files only, no Go coverage applies), Wave C drives coverage on new handlers to ≥ 90% (each handler has 3 dedicated test cases).

---

## Wave A — Schema Extension + Cobra Subcommand Scaffolding

### T-A1 (P0) — Extend EventType enum with three new constants

**Description**: Edit `internal/harness/types.go` to add three new `EventType` constants under the existing `const (...)` block:
- `EventTypeSessionStop EventType = "session_stop"`
- `EventTypeSubagentStop EventType = "subagent_stop"`
- `EventTypeUserPrompt EventType = "user_prompt"`

Preserve the existing four constants verbatim. Add a Go comment block citing `REQ-HRN-OBS-015` and `SPEC-V3R4-HARNESS-002`.

**Linked REQs**: REQ-HRN-OBS-015

**Linked ACs**: AC-HRN-OBS-010

**Owner role**: implementer

**Files touched**:
- `internal/harness/types.go` (modify the EventType const block, add 3 new constants + comment)

**Depends on**: none

**MX tag implications**: `@MX:NOTE` on the new constants block citing REQ-HRN-OBS-015 enum extension and downstream consumer (SPEC-V3R4-HARNESS-003).

**Complexity**: XS (3 lines of new code + comment).

**Wave**: Wave A

**Definition of Done**:
- Three new `EventType` constants exist in `internal/harness/types.go`.
- Existing four constants are byte-identical to pre-edit state.
- `go build ./...` succeeds.
- `go vet ./...` shows zero new issues.

---

### T-A2 (P0) — Extend Event struct with optional per-event fields

**Description**: Edit `internal/harness/types.go` `Event` struct to add the following NEW fields, each tagged `json:"...,omitempty"`. Preserve the existing six fields (Timestamp, EventType, Subject, ContextHash, TierIncrement, SchemaVersion) verbatim.

New fields (Stop-specific):
- `SessionID string \`json:"session_id,omitempty"\``
- `LastAssistantMessageHash string \`json:"last_assistant_message_hash,omitempty"\``
- `LastAssistantMessageLen int \`json:"last_assistant_message_len,omitempty"\``

New fields (SubagentStop-specific):
- `AgentName string \`json:"agent_name,omitempty"\``
- `AgentType string \`json:"agent_type,omitempty"\``
- `AgentID string \`json:"agent_id,omitempty"\``
- `ParentSessionID string \`json:"parent_session_id,omitempty"\``

New fields (UserPromptSubmit-specific, Strategy A):
- `PromptHash string \`json:"prompt_hash,omitempty"\``
- `PromptLen int \`json:"prompt_len,omitempty"\``
- `PromptLang string \`json:"prompt_lang,omitempty"\``

New fields (UserPromptSubmit, opt-in B and C):
- `PromptPreview string \`json:"prompt_preview,omitempty"\``
- `PromptContent string \`json:"prompt_content,omitempty"\``

**Linked REQs**: REQ-HRN-OBS-009, REQ-HRN-OBS-004, REQ-HRN-OBS-005, REQ-HRN-OBS-006, REQ-HRN-OBS-013

**Linked ACs**: AC-HRN-OBS-002, AC-HRN-OBS-003, AC-HRN-OBS-004, AC-HRN-OBS-006, AC-HRN-OBS-008

**Owner role**: implementer

**Files touched**:
- `internal/harness/types.go` (extend Event struct)

**Depends on**: T-A1 (enum values are referenced in struct comments).

**MX tag implications**: `@MX:NOTE` on the Event struct citing REQ-HRN-OBS-009 schema additivity and the `omitempty` invariant.

**Complexity**: S (12 new fields with tags + comment).

**Wave**: Wave A

**Definition of Done**:
- All 12 new fields present in `Event` struct with correct `omitempty` JSON tags.
- Existing 6 fields unchanged.
- `go build ./...` succeeds.
- A unit test (added in T-C1+ or separately as a struct-shape test) verifies that an `Event{}` zero value serializes to a JSONL line containing ONLY the existing baseline fields (none of the new optional fields appear).

---

### T-A3 (P0) — Implement runHarnessObserveStop handler

**Description**: Edit `internal/cli/hook.go` to add a new function `runHarnessObserveStop(cmd *cobra.Command, args []string) error` that mirrors the existing `runHarnessObserve` pattern (line 474+). The function MUST:

1. Resolve cwd via `os.Getwd()` (fall back to `"."`).
2. Call `isHarnessLearningEnabled(cwd)`; if false, return nil immediately (no-op).
3. Decode stdin JSON into a struct with fields: `LastAssistantMessage string \`json:"last_assistant_message"\``, `Session struct { ID string \`json:"id"\` } \`json:"session"\``.
4. Compute extended fields: `last_assistant_message_hash` = first 16 hex chars of SHA-256(last_assistant_message); `last_assistant_message_len` = byte length; `session_id` = session.id; `subject` = empty or SPEC ID extracted from cwd if `.moai/specs/SPEC-...` exists.
5. Construct an `Event{}` with `EventType: EventTypeSessionStop`, baseline fields, and the extended fields populated where their source values are non-empty.
6. Use `Observer.RecordEvent` (existing function) with the new EventType; OR add an `Observer.RecordExtendedEvent(e Event)` helper if the existing `RecordEvent(EventType, subject, contextHash)` signature is insufficient. (The plan-auditor will decide; the recommended approach is to add a thin helper, since the existing signature does not accept the extended fields.)
7. Register the cobra subcommand under `hookCmd.AddCommand(&cobra.Command{Use: "harness-observe-stop", Short: "Record Stop event to harness usage log", RunE: runHarnessObserveStop})`.
8. Always return nil on error (non-blocking observer); log to stderr via `cmd.ErrOrStderr()` if `RecordEvent` fails.

**Linked REQs**: REQ-HRN-OBS-001, REQ-HRN-OBS-004, REQ-HRN-OBS-008

**Linked ACs**: AC-HRN-OBS-001.a, AC-HRN-OBS-002, AC-HRN-OBS-005

**Owner role**: implementer

**Files touched**:
- `internal/cli/hook.go` (new function + new cobra subcommand registration)
- `internal/harness/observer.go` (POTENTIALLY: add `RecordExtendedEvent(e Event)` helper if required)

**Depends on**: T-A1, T-A2

**MX tag implications**: `@MX:NOTE` on `runHarnessObserveStop` citing REQ-HRN-OBS-004 + REQ-HRN-OBS-008. If `RecordExtendedEvent` is added, `@MX:ANCHOR` on it (fan_in ≥ 3 from the three new handlers).

**Complexity**: M (~60 LOC including the new helper).

**Wave**: Wave A

**Definition of Done**:
- `runHarnessObserveStop` function exists in `internal/cli/hook.go`.
- Cobra subcommand `harness-observe-stop` appears in `hookCmd.Commands()`.
- A unit test in T-C1 passes.

---

### T-A4 (P0) — Implement runHarnessObserveSubagentStop handler

**Description**: Mirror T-A3 pattern for SubagentStop. Stdin decode struct:
```go
struct {
  AgentType string `json:"agentType"`
  AgentName string `json:"agentName"`
  LastAssistantMessage string `json:"last_assistant_message"`
  AgentID string `json:"agent_id"`
  AgentTranscriptPath string `json:"agent_transcript_path"`
  Session struct { ID string `json:"id"` } `json:"session"`
}
```

Construct `Event{}` with `EventType: EventTypeSubagentStop`, `Subject: agentName`, and extended fields `agent_name`, `agent_type`, `agent_id`, `parent_session_id` (from session.id). Register cobra subcommand `harness-observe-subagent-stop`.

**Linked REQs**: REQ-HRN-OBS-002, REQ-HRN-OBS-005, REQ-HRN-OBS-008

**Linked ACs**: AC-HRN-OBS-001.b, AC-HRN-OBS-003, AC-HRN-OBS-005

**Owner role**: implementer

**Files touched**:
- `internal/cli/hook.go` (new function + new cobra registration)

**Depends on**: T-A1, T-A2, T-A3 (the `RecordExtendedEvent` helper if added in A3 is reused)

**MX tag implications**: `@MX:NOTE` on `runHarnessObserveSubagentStop` citing REQ-HRN-OBS-005 + REQ-HRN-OBS-008.

**Complexity**: S (~40 LOC, simpler than A3 because no SPEC ID extraction).

**Wave**: Wave A

**Definition of Done**:
- `runHarnessObserveSubagentStop` exists.
- Cobra subcommand `harness-observe-subagent-stop` registered.
- A unit test in T-C2 passes.

---

### T-A5 (P0) — Implement runHarnessObserveUserPromptSubmit handler + resolveUserPromptStrategy

**Description**: This is the most complex Wave A task because of PII strategy switching.

Step 1: Implement a NEW helper function `resolveUserPromptStrategy(projectRoot string) UserPromptStrategy` in `internal/cli/hook.go`:
- Read `.moai/config/sections/harness.yaml`.
- Locate the `learning.user_prompt_content` key.
- Map the string value to a `UserPromptStrategy` enum: `StrategyHash` (default), `StrategyPreview`, `StrategyFull`, `StrategyNone`.
- On parse error, missing file, absent key, or UNKNOWN value: return `StrategyHash` (fail-open per REQ-HRN-OBS-014).
- Emit a non-blocking stderr warning ONLY on unknown value detection (to aid debugging).

Step 2: Implement `runHarnessObserveUserPromptSubmit` mirroring T-A3 pattern. Stdin decode struct:
```go
struct {
  Prompt string `json:"prompt"`
  Session struct { ID string `json:"id"` } `json:"session"`
}
```

Apply strategy per `resolveUserPromptStrategy`:
- `StrategyNone`: skip event (return nil without RecordEvent call).
- `StrategyHash` / default: set `prompt_hash`, `prompt_len`, `prompt_lang` only.
- `StrategyPreview`: set the Strategy A fields PLUS `prompt_preview` (first 64 bytes of prompt; full prompt if shorter).
- `StrategyFull`: set the Strategy A fields PLUS `prompt_content` (full prompt).

`prompt_lang` heuristic: examine the first non-whitespace rune:
- Korean (Hangul Unicode block): `"ko"`
- Japanese (Hiragana/Katakana): `"ja"`
- Chinese (CJK Unified Ideographs): `"zh"`
- Latin-1 (ASCII letters): `"en"`
- Otherwise: `""` (empty)

`subject` extraction: apply regex `SPEC-[A-Z][A-Z0-9]+-[0-9]+` to the prompt; if a match exists, use the first match as Subject; else empty string.

Register cobra subcommand `harness-observe-user-prompt-submit`.

**Linked REQs**: REQ-HRN-OBS-003, REQ-HRN-OBS-006, REQ-HRN-OBS-008, REQ-HRN-OBS-012, REQ-HRN-OBS-013, REQ-HRN-OBS-014

**Linked ACs**: AC-HRN-OBS-001.c, AC-HRN-OBS-004, AC-HRN-OBS-005, AC-HRN-OBS-007, AC-HRN-OBS-008.a, AC-HRN-OBS-008.b, AC-HRN-OBS-008.c, AC-HRN-OBS-009

**Owner role**: implementer

**Files touched**:
- `internal/cli/hook.go` (new function + new cobra registration + new helper)

**Depends on**: T-A1, T-A2, T-A3 (the RecordExtendedEvent helper if added)

**MX tag implications**:
- `@MX:WARN` on `runHarnessObserveUserPromptSubmit` with `@MX:REASON` citing PII handling sensitivity and the fail-open invariant (REQ-HRN-OBS-014).
- `@MX:ANCHOR` on `resolveUserPromptStrategy` with `@MX:REASON` citing the fail-open invariant and fan_in (it is called from the handler, the test fixture, and the gate-uniformity test).

**Complexity**: L (~120 LOC including helper, strategy enum, language heuristic, SPEC regex).

**Wave**: Wave A

**Definition of Done**:
- `runHarnessObserveUserPromptSubmit` exists.
- `resolveUserPromptStrategy` helper exists.
- Strategy enum (`UserPromptStrategy`) defined with four values.
- Cobra subcommand `harness-observe-user-prompt-submit` registered.
- Unit tests in T-C3 pass (NoOp/Records/PreservesExisting × strategy switch matrix + fail-open).

---

## Wave B — Wrapper Script Templates + settings.json.tmpl Registration

### T-B1 (P0) — Create handle-harness-observe-stop.sh.tmpl

**Description**: Clone `internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl` to a new file `handle-harness-observe-stop.sh.tmpl` in the same directory. Change exactly two things:
1. The header comment from `# This script forwards stdin JSON to the moai hook harness-observe command.` to `# This script forwards stdin JSON to the moai hook harness-observe-stop command.`
2. The `exec` lines (three occurrences) from `moai hook harness-observe` to `moai hook harness-observe-stop`.

All other lines (stderr log setup, log rotation, binary search) are byte-identical to the source.

**Linked REQs**: REQ-HRN-OBS-016

**Linked ACs**: AC-HRN-OBS-011.a

**Owner role**: template-author

**Files touched**:
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-stop.sh.tmpl` (new file)

**Depends on**: none (T-A3 implements the corresponding subcommand but file existence is independent)

**MX tag implications**: none (shell template, no Go code).

**Complexity**: XS (32 lines, identical to source with 2 string substitutions).

**Wave**: Wave B

**Definition of Done**:
- File exists.
- `diff handle-harness-observe.sh.tmpl handle-harness-observe-stop.sh.tmpl | grep -c '^>'` returns approximately 4 (the header comment line + 3 exec lines).
- File has executable bit (matches source file mode).

---

### T-B2 (P0) — Create handle-harness-observe-subagent-stop.sh.tmpl

**Description**: Same as T-B1 but for SubagentStop. Substitution: `harness-observe` → `harness-observe-subagent-stop`.

**Linked REQs**: REQ-HRN-OBS-016

**Linked ACs**: AC-HRN-OBS-011.a

**Owner role**: template-author

**Files touched**:
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-subagent-stop.sh.tmpl` (new file)

**Depends on**: none

**MX tag implications**: none.

**Complexity**: XS.

**Wave**: Wave B

**Definition of Done**: same as T-B1.

---

### T-B3 (P0) — Create handle-harness-observe-user-prompt-submit.sh.tmpl

**Description**: Same as T-B1 but for UserPromptSubmit. Substitution: `harness-observe` → `harness-observe-user-prompt-submit`.

**Linked REQs**: REQ-HRN-OBS-016

**Linked ACs**: AC-HRN-OBS-011.a

**Owner role**: template-author

**Files touched**:
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-user-prompt-submit.sh.tmpl` (new file)

**Depends on**: none

**MX tag implications**: none.

**Complexity**: XS.

**Wave**: Wave B

**Definition of Done**: same as T-B1.

---

### T-B4 (P0) — Add three additive entries to settings.json.tmpl

**Description**: Edit `internal/template/templates/.claude/settings.json.tmpl` to add ONE new hook entry under each of the three event slots (`Stop`, `SubagentStop`, `UserPromptSubmit`). Use the platform-conditional path quoting pattern matching the existing entries.

For the `Stop` event slot (around line 84-97), after the existing `hooks: [{ ...handle-stop.sh... }]` array, add a SECOND array element pointing to `handle-harness-observe-stop.sh`. The resulting structure:

```json
"Stop": [
  {
    "hooks": [
      {
        "command": "<existing handle-stop.sh path>",
        "timeout": 5,
        "type": "command"
      },
      {
        "command": "<NEW handle-harness-observe-stop.sh path>",
        "timeout": 5,
        "type": "command"
      }
    ]
  }
]
```

(Note: the platform-conditional `{{if eq .Platform "windows"}}` blocks for both Unix and Windows paths must be replicated for the new entry, matching the existing pattern.)

Apply the same additive pattern to the `SubagentStop` and `UserPromptSubmit` slots.

After edit, run `make build` from the project root to regenerate `internal/template/embedded.go`. Verify the embedded.go reflects the new templates.

**Linked REQs**: REQ-HRN-OBS-017

**Linked ACs**: AC-HRN-OBS-011.b

**Owner role**: template-author

**Files touched**:
- `internal/template/templates/.claude/settings.json.tmpl` (modify 3 event slot entries)
- `internal/template/embedded.go` (regenerated by `make build`)

**Depends on**: T-B1, T-B2, T-B3 (template files must exist before settings.json.tmpl references them)

**MX tag implications**: none (template file).

**Complexity**: M (approximately 30 new lines in settings.json.tmpl plus `make build` regeneration).

**Wave**: Wave B

**Definition of Done**:
- Three new hook entries present under their respective event slots.
- Existing entries are byte-identical to pre-edit state (verified via `git diff`).
- `make build` succeeds.
- `internal/template/embedded.go` is regenerated and reflects the new templates.
- A test render of settings.json (from a fresh fixture) shows the expected JSON structure.

---

## Wave C — Test Suite + 5-Layer Safety Integration Verification

### T-C1 (P1) — Author hook_harness_observe_stop_test.go

**Description**: Create `internal/cli/hook_harness_observe_stop_test.go` with three test functions following the V3R4-001 baseline pattern (`hook_harness_observe_test.go`):

- `TestRunHarnessObserveStop_NoOpWhenLearningDisabled`: sets `learning.enabled: false`, invokes handler, asserts `.moai/harness/usage-log.jsonl` is NOT created and stderr is empty.
- `TestRunHarnessObserveStop_PreservesExistingLogWhenDisabled`: pre-seeds the log with two baseline entries, invokes handler with `learning.enabled: false`, asserts the file contents are byte-identical to seed.
- `TestRunHarnessObserveStop_RecordsWhenEnabled`: sets `learning.enabled: true`, invokes handler with representative stdin JSON, asserts exactly one JSONL line is appended with the four baseline fields PLUS the three Stop-specific fields (`session_id`, `last_assistant_message_hash`, `last_assistant_message_len`); asserts `event_type == "session_stop"`; asserts raw `last_assistant_message` text does NOT appear in the file.

Reuse the `writeHarnessYAML` and `withStdin` helpers from `hook_harness_observe_test.go` (extract them to a shared `_test_helpers.go` file if necessary).

**Linked REQs**: REQ-HRN-OBS-004, REQ-HRN-OBS-008, REQ-HRN-OBS-009

**Linked ACs**: AC-HRN-OBS-002, AC-HRN-OBS-005, AC-HRN-OBS-006

**Owner role**: tester

**Files touched**:
- `internal/cli/hook_harness_observe_stop_test.go` (new file)
- `internal/cli/hook_harness_test_helpers.go` (new, if helper extraction is chosen; otherwise duplicate helpers within the new test file)

**Depends on**: T-A3

**MX tag implications**: `@MX:NOTE` on the test file header citing REQ-HRN-OBS-004 and the V3R4-001 test pattern source.

**Complexity**: M (~100 LOC for 3 test functions + helper imports).

**Wave**: Wave C

**Definition of Done**:
- 3 test functions pass via `go test -run TestRunHarnessObserveStop ./internal/cli/...`.
- Coverage on `runHarnessObserveStop` reaches ≥ 90%.

---

### T-C2 (P1) — Author hook_harness_observe_subagent_stop_test.go

**Description**: Same pattern as T-C1, for `runHarnessObserveSubagentStop`. Three test functions: `NoOpWhenLearningDisabled`, `PreservesExistingLogWhenDisabled`, `RecordsWhenEnabled`. The Records test asserts `event_type == "subagent_stop"`, `subject == "<agentName>"`, presence of `agent_name`, `agent_type`, `agent_id`, `parent_session_id`.

**Linked REQs**: REQ-HRN-OBS-005, REQ-HRN-OBS-008, REQ-HRN-OBS-009

**Linked ACs**: AC-HRN-OBS-003, AC-HRN-OBS-005, AC-HRN-OBS-006

**Owner role**: tester

**Files touched**:
- `internal/cli/hook_harness_observe_subagent_stop_test.go` (new file)

**Depends on**: T-A4

**MX tag implications**: `@MX:NOTE` citing REQ-HRN-OBS-005.

**Complexity**: M (~100 LOC).

**Wave**: Wave C

**Definition of Done**:
- 3 test functions pass.
- Coverage on `runHarnessObserveSubagentStop` ≥ 90%.

---

### T-C3 (P1) — Author hook_harness_observe_user_prompt_submit_test.go (most complex)

**Description**: Create `internal/cli/hook_harness_observe_user_prompt_submit_test.go` with the following test functions:

1. `TestRunHarnessObserveUserPromptSubmit_NoOpWhenLearningDisabled` (mirrors T-C1).
2. `TestRunHarnessObserveUserPromptSubmit_PreservesExistingLogWhenDisabled` (mirrors T-C1).
3. `TestRunHarnessObserveUserPromptSubmit_RecordsWhenEnabledStrategyA`: default strategy (no `user_prompt_content` key). Asserts `prompt_hash`, `prompt_len`, `prompt_lang` present; `prompt_preview` and `prompt_content` ABSENT; raw prompt text not in JSONL.
4. `TestRunHarnessObserveUserPromptSubmit_StrategyMatrix` (table-driven): tests all four strategy values (`hash`, `preview`, `full`, `none`) with a fixed prompt input. For each row, asserts the field-presence matrix:
   | Strategy | prompt_hash | prompt_preview | prompt_content | Entry written? |
   |----------|-------------|----------------|----------------|----------------|
   | hash (default) | yes | no | no | yes |
   | preview | yes | yes | no | yes |
   | full | yes | no | yes | yes |
   | none | n/a | n/a | n/a | no (file absent) |
5. `TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy`: sets `learning.user_prompt_content: garbage_value_xyz`, asserts Strategy A behavior (hash only, no preview/content, raw prompt absent), MAY assert a stderr warning.
6. `TestRunHarnessObserveUserPromptSubmit_LanguageHeuristic` (table-driven): tests `prompt_lang` detection for Korean ("안녕"), Japanese ("こんにちは"), Chinese ("你好"), English ("Hello"), Mixed ("Hello 안녕"), Empty ("").
7. `TestRunHarnessObserveUserPromptSubmit_SpecIdExtraction`: tests Subject field extraction for prompts containing "SPEC-V3R4-HARNESS-002" (Subject = the SPEC ID), no SPEC mention (Subject = empty).

**Linked REQs**: REQ-HRN-OBS-006, REQ-HRN-OBS-008, REQ-HRN-OBS-009, REQ-HRN-OBS-012, REQ-HRN-OBS-013, REQ-HRN-OBS-014

**Linked ACs**: AC-HRN-OBS-004, AC-HRN-OBS-005, AC-HRN-OBS-006, AC-HRN-OBS-007, AC-HRN-OBS-008.a, AC-HRN-OBS-008.b, AC-HRN-OBS-008.c, AC-HRN-OBS-009

**Owner role**: tester

**Files touched**:
- `internal/cli/hook_harness_observe_user_prompt_submit_test.go` (new file)

**Depends on**: T-A5

**MX tag implications**: `@MX:NOTE` citing REQ-HRN-OBS-006/012/013/014. `@MX:ANCHOR` on the PII strategy matrix test (it is the canonical regression guard for the fail-open invariant).

**Complexity**: L (~250 LOC including 7 test functions and tables).

**Wave**: Wave C

**Definition of Done**:
- All 7 test functions pass.
- Coverage on `runHarnessObserveUserPromptSubmit` and `resolveUserPromptStrategy` ≥ 90%.
- The PII fail-open path is verified by `TestRunHarnessObserveUserPromptSubmit_FailOpenOnInvalidStrategy`.

---

### T-C4 (P1) — Author hook_harness_gate_uniformity_test.go

**Description**: Create `internal/cli/hook_harness_gate_uniformity_test.go` with a single table-driven test `TestHookHarnessGateUniformity` that iterates over all FOUR observer handlers (`runHarnessObserve`, `runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) and asserts that ALL of them no-op when `learning.enabled: false`. The test:

1. Pre-seeds `.moai/harness/usage-log.jsonl` with two baseline entries.
2. Sets `learning.enabled: false`.
3. For each handler, invokes with a representative stdin JSON.
4. After ALL four handlers run, asserts the file is byte-identical to the seed.
5. Asserts no handler returned an error.

This test prevents accidental future regression where a developer adds a new event handler that forgets the gate check.

**Linked REQs**: REQ-HRN-OBS-008

**Linked ACs**: AC-HRN-OBS-005

**Owner role**: tester

**Files touched**:
- `internal/cli/hook_harness_gate_uniformity_test.go` (new file)

**Depends on**: T-A3, T-A4, T-A5

**MX tag implications**: `@MX:NOTE` citing REQ-HRN-OBS-008 unified gate; flag the test as the regression guard.

**Complexity**: S (~60 LOC).

**Wave**: Wave C

**Definition of Done**:
- Test passes for all 4 handlers.
- Pre-seed verification confirms byte-identical preservation.

---

### T-C5 (P1) — Author safety_preservation_test.go (architectural assertion)

**Description**: Create `internal/harness/safety_preservation_test.go` containing two architectural assertion tests:

1. `TestSafetyArchitecture_LayerCount`: reads `.claude/rules/moai/design/constitution.md`, parses §5 (Safety Architecture), asserts EXACTLY 5 layers exist with their canonical names (Frozen Guard, Canary Check, Contradiction Detector, Rate Limiter, Human Oversight). This is a STRING-MATCHING test against the constitution body.
2. `TestSafetyArchitecture_FrozenZoneUnchanged`: asserts that the FROZEN zone path-prefix list (per REQ-HRN-FND-006) contains exactly four entries: `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, `.moai/project/brand/`. V3R4-002 MUST NOT add new entries.

These tests run during the standard `go test ./internal/harness/...` invocation. They are not coverage-targets; they are regression guards.

**Linked REQs**: REQ-HRN-OBS-007, REQ-HRN-OBS-010, REQ-HRN-OBS-011

**Linked ACs**: AC-HRN-OBS-012

**Owner role**: tester

**Files touched**:
- `internal/harness/safety_preservation_test.go` (new file)

**Depends on**: none (this test reads existing files; not blocked by other tasks)

**MX tag implications**: `@MX:ANCHOR` on the test functions with `@MX:REASON` citing them as the V3R4-001 contract preservation regression guard for V3R4-002 and all downstream V3R4 SPECs.

**Complexity**: S (~80 LOC).

**Wave**: Wave C

**Definition of Done**:
- Both tests pass.
- Test failure message clearly indicates a 5-Layer Safety regression if a layer is removed or renamed.

---

## Dependency Graph (Task-Level)

```
Wave A: T-A1 → T-A2 → T-A3 → T-A4 → T-A5
        (T-A1 enum dependency for all subsequent;
         T-A4 and T-A5 can run in parallel after T-A3 because they reuse the same helper)

Wave B: T-B1, T-B2, T-B3 (parallel — no inter-dependency)
        T-B1 + T-B2 + T-B3 → T-B4 (settings.json.tmpl references the templates)

Wave C: T-C1 (depends on T-A3)
        T-C2 (depends on T-A4)
        T-C3 (depends on T-A5)
        T-C4 (depends on T-A3, T-A4, T-A5)
        T-C5 (no task dependency — reads existing constitution)
        (All Wave C tasks can run in parallel after Wave A is complete)

Inter-Wave:
  Wave A complete → Wave B entry (Wave B's template-first regeneration via make build benefits from the new Go subcommands being present, although strictly the templates can be authored without the binary changes)
  Wave A complete + Wave B complete → Wave C entry
  
  Wave B can technically begin before Wave A completes (template files are pure text), but the `make build` step in T-B4 will not have anything new to embed until Wave A merges. Recommended order: Wave A → Wave B → Wave C.
```

---

## Coverage Targets per Wave

| Wave | Coverage Target | Measurement |
|------|----------------|-------------|
| Wave A | n/a (implementation; coverage measured in Wave C tests) | — |
| Wave B | n/a (template files only, no Go coverage applies) | — |
| Wave C | New code in Wave A reaches ≥ 90% line coverage | `go test -cover ./internal/cli/... ./internal/harness/...` |
| Overall package | `internal/cli`: 85% minimum, currently around 80%. V3R4-002 adds approximately 200 LOC + 9 tests; net coverage should rise to ≥ 85%. | CI quality gate |

---

## Out of Scope (Task-Level)

The following tasks are explicitly NOT in any Wave of this SPEC. They are deferred to downstream SPECs:

| Downstream SPEC | Task Domain |
|-----------------|-------------|
| SPEC-V3R4-HARNESS-003 | Embedding-cluster algorithm; embedding model selection; replacing the frequency-count tier classifier with the embedding-cluster classifier. |
| SPEC-V3R4-HARNESS-004 | Reflexion-style verbal self-critique; episodic memory of reflections; iteration-cap mechanism. |
| SPEC-V3R4-HARNESS-005 | Constitution principle parser; principle-based scoring rubric; pre-screen integration before AskUserQuestion. |
| SPEC-V3R4-HARNESS-006 | Multi-objective scoring tuple (quality + cost + latency + iteration); auto-rollback-on-regression. |
| SPEC-V3R4-HARNESS-007 | Voyager-style embedding-indexed skill library; top-K retrieval; compositional skill reuse. |
| SPEC-V3R4-HARNESS-008 | Anonymization layer; cross-project federation; opt-in approval flow; namespace isolation. |
| Follow-up V3R4-001 cleanup | Retroactively wiring `handle-harness-observe.sh` into the PostToolUse slot of `settings.json.tmpl` (V3R4-001 gap per research.md §2.5 OQ-3). |
| Follow-up migration SPEC | Tooling to migrate pre-V3R4-002 `usage-log.jsonl` entries to a normalized extended schema (NOT in scope; pre-existing entries remain valid under additivity rule). |

---

## Run-Phase Entry Point

After this plan PR merges into main:

1. Execute `/clear` to reset context.
2. Verify the SPEC worktree is still anchored: `cd ~/.moai/worktrees/MoAI-ADK/SPEC-V3R4-HARNESS-002 && git merge origin/main` to incorporate the merged plan PR.
3. Execute `/moai run SPEC-V3R4-HARNESS-002` from inside the worktree.
4. `manager-develop` (with `cycle_type` per `quality.yaml development_mode`) takes over execution of Waves A → B → C sequentially.
5. Each Wave merges as a separate squash PR per Enhanced GitHub Flow doctrine (`CLAUDE.local.md` § 18). Conventional Commits prefix: `feat(SPEC-V3R4-HARNESS-002):` for Wave A and B; `test(SPEC-V3R4-HARNESS-002):` for Wave C.
6. After all three Waves merge, execute `/moai sync SPEC-V3R4-HARNESS-002` (same worktree) to generate the final documentation sync PR.
7. After sync PR merges, `manager-git` may execute the worktree cleanup (`moai worktree done SPEC-V3R4-HARNESS-002`) from the host checkout.

---

End of tasks.md.

> **[REMOVED 2026-07-13]** The `gateguard-fact-force.sh` PreToolUse advisory hook referenced in AC-026 below has been FULLY DECOMMISSIONED from both the local working tree and the distribution template. The hook script, `MOAI_FACT_FORCE` env var, `.moai/state/fact-force/` state dir, and `.moai/logs/fact-force-skip.log` skip log are all removed. References below are preserved as historical decision-record content only.

# acceptance.md — SPEC-HOOK-OFFICIAL-COMPLIANCE-001

> Binary acceptance criteria. Each AC states its verification command, expected output, and REQ traceability. Run-phase reports each per the verification-claim-integrity 5-section format.

---

## §D. AC Matrix

### M1 — Blocking gate JSON contract + PreToolUse exit-code (HIGH)

#### AC-HOC-001 (REQ-HOC-001, GAP-GD-01/STT-02) — team-ac-verify TaskCompleted JSON form

**Given** the `team-ac-verify.sh` gate is invoked with a `--reject` stub payload,
**When** the gate's reject path emits stdout JSON,
**Then** the JSON **shall** contain `"continue":false` and `"stopReason"` keys, and **shall not** contain a top-level `"decision"` key.

```bash
# Verification (windowed-grep — anchored to the reject-path printf window, NOT whole-file).
# Rationale: post-fix, comment lines documenting the OLD form (around lines 10/12/14/16/29)
# still match 'decision.*block', so a whole-file grep would false-pass. Per memory
# feedback_windowed_grep_undercount_authoring.md.
sed -n '40,50p' internal/template/templates/.claude/hooks/moai/team-ac-verify.sh | grep -E 'continue.*false|stopReason'
# Expected: ≥1 match (the corrected reject-path form WITHIN the bounded printf window)
sed -n '40,50p' internal/template/templates/.claude/hooks/moai/team-ac-verify.sh | grep -c 'decision.*block'
# Expected: 0 (no 'decision:block' within the bounded reject-path window; comment lines
#           outside this window do NOT count toward the AC)
```

#### AC-HOC-002 (REQ-HOC-002, GAP-GD-02) — sync-phase-quality-gate Stop hookSpecificOutput nesting

**Given** the `sync-phase-quality-gate.sh` gate blocks a Stop event,
**When** the block path emits stdout JSON,
**Then** the `decision` and `reason` keys **shall** be nested inside a `hookSpecificOutput` object that also contains `"hookEventName":"Stop"`.

```bash
grep -A2 'hookSpecificOutput' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh | grep -E 'hookEventName.*Stop|decision.*block'
# Expected: 1+ match showing hookEventName:"Stop" co-located with decision:block inside hookSpecificOutput
```

#### AC-HOC-003 (REQ-HOC-003, GAP-TOOL-01) — handle-pre-tool exit-code passthrough

**Given** the moai binary returns exit 2 on a PreToolUse reject,
**When** the `handle-pre-tool.sh.tmpl` wrapper completes,
**Then** the wrapper's process exit code **shall** be 2 (NOT hardcoded 0), and stderr **shall** be preserved as the Claude-feedback channel (no `2>>"$MOAI_HOOK_STDERR_LOG"` on the PreToolUse branch, OR the redirect is event-conditional).

```bash
grep -n 'exit 0\|exit \$?' internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl | head -10
# Expected: the exit-0 hardcoding AFTER 'moai hook pre-tool' is REPLACED by exit-code passthrough
# AND Go inspection of internal/hook/pre_tool.go confirms which path the handler uses (recorded in the run-phase E7 report)
```

#### AC-HOC-004 (REQ-HOC-003, GAP-TOOL-01) — handle-pre-tool stderr preservation

**Given** the moai binary writes a reject reason to stderr,
**When** the wrapper invokes `moai hook pre-tool`,
**Then** the stderr **shall** reach Claude Code (NOT be fully redirected to the log file).

```bash
grep -n '2>>.*MOAI_HOOK_STDERR_LOG' internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh.tmpl
# Expected: either 0 matches on the PreToolUse branch, OR the redirect is event-conditional (documented)
```

#### AC-HOC-005 (REQ-HOC-004, GAP-LC2-01) — SessionStart probe hookEventName

**Given** the SessionStart moai-resolvability probe fires (moai unresolvable in all 3 tiers),
**When** the probe emits its `hookSpecificOutput.additionalContext` JSON,
**Then** the JSON **shall** contain `"hookEventName":"SessionStart"` alongside `additionalContext`.

```bash
grep 'hookSpecificOutput' internal/template/templates/.claude/hooks/moai/handle-session-start.sh.tmpl | grep 'hookEventName.*SessionStart'
# Expected: 1 match — the hookSpecificOutput now carries hookEventName:"SessionStart"
```

#### AC-HOC-006 (REQ-HOC-005, GAP-STT-02) — Ledger Closure doctrine mirror

**Given** `agent-common-protocol.md` § Ledger Closure clause (b) describes the team-ac-verify reject form,
**When** the doctrine is read,
**Then** it **shall** encode the corrected `continue:false` + `stopReason` form and **shall not** encode the old `decision:block` form.

```bash
grep -A3 'team-ac-verify.sh' .claude/rules/moai/core/agent-common-protocol.md | grep -E 'continue.*false|stopReason'
# Expected: 1+ match; NO 'decision.*block' in the team-ac-verify description context
```

#### AC-HOC-007 (M1 template-mirror parity)

**Given** M1 edits template hook files,
**When** `go test ./internal/template/...` runs,
**Then** `TestRuleTemplateMirror*` **shall** PASS (template and live mirrors byte-identical for all doctrine files).

```bash
go test ./internal/template/... -run TestRuleTemplateMirror -v 2>&1 | grep -E 'PASS|FAIL'
# Expected: PASS
```

### M2 — Doctrine refresh (MEDIUM)

#### AC-HOC-008 (REQ-HOC-006, GAP-LC-01) — PreCompact Can Block

**Given** the `hooks-system.md` event reference table,
**When** the PreCompact row is read,
**Then** "Can Block" **shall** be `Yes` (was `No`).

```bash
grep -i 'PreCompact' .claude/rules/moai/core/hooks-system.md | grep -i 'yes\|block'
# Expected: the PreCompact row shows Can Block: Yes
```

#### AC-HOC-009 (REQ-HOC-006, GAP-STT-03) — SubagentStop stdout columns

**Given** the SubagentStop row,
**When** the stdout column is read,
**Then** it **shall** list `decision:block+reason | additionalContext` alongside `systemMessage`.

```bash
grep -i 'SubagentStop' .claude/rules/moai/core/hooks-system.md | grep -iE 'decision.*block|additionalContext'
# Expected: 1+ match
```

#### AC-HOC-010 (REQ-HOC-006, GAP-OBS-04) — Notification matcher completeness

**Given** the Notification row,
**When** the matcher column is read,
**Then** it **shall** list all 8 official matcher values (the 4 current + elicitation_complete, elicitation_response, agent_needs_input, agent_completed). SessionEnd row = 6 values; StopFailure row = 10 values.

```bash
grep -iE 'Notification|SessionEnd|StopFailure' .claude/rules/moai/core/hooks-system.md | head -10
# Expected: matcher counts match official (8 / 6 / 10)
```

#### AC-HOC-011 (REQ-HOC-006, GAP-CEW-03) — FileChanged literal-not-regex note

**Given** the FileChanged row (or stdin/stdout table),
**When** the matcher description is read,
**Then** it **shall** state "literal filenames, NOT regex/glob".

```bash
grep -i 'FileChanged' .claude/rules/moai/core/hooks-system.md | grep -iE 'literal|not.*regex|not.*glob'
# Expected: 1+ match
```

#### AC-HOC-012 (REQ-HOC-006, GAP-PE-02) — UserPromptSubmit/PermissionRequest field sets

**Given** the UserPromptSubmit row,
**When** the stdout column is read,
**Then** it **shall** include `decision:{block,reason}`, `sessionTitle`, `suppressOriginalPrompt`. PermissionRequest row **shall** include `decision.behavior`, `updatedInput`, `updatedPermissions`.

```bash
grep -iE 'UserPromptSubmit|PermissionRequest' .claude/rules/moai/core/hooks-system.md | grep -iE 'sessionTitle|suppressOriginal|updatedInput|updatedPermissions|behavior'
# Expected: 1+ match per event row
```

#### AC-HOC-013 (REQ-HOC-006, GAP-PE-03) — event-specific handler-type column

**Given** the event table,
**When** the handler-type column is read,
**Then** Elicitation/ElicitationResult **shall** show "command+http+mcp_tool only (prompt/agent silently discarded)".

```bash
grep -iE 'Elicitation' .claude/rules/moai/core/hooks-system.md | grep -iE 'command.*http.*mcp_tool|prompt.*agent.*discard'
# Expected: 1+ match
```

#### AC-HOC-014 (REQ-HOC-006, GAP-GD-06) — exit-2 per-event enum

**Given** the Rules section,
**When** the exit-2 description is read,
**Then** it **shall not** say "universal"; it **shall** enumerate which events honor exit 2 and which ignore it (PermissionDenied, Notification, SessionStart, SessionEnd, Setup, CwdChanged, FileChanged, PostCompact, SubagentStart, InstructionsLoaded, StopFailure, MessageDisplay ignore exit 2).

```bash
grep -i 'exit.*2.*universal\|exit.*2.*block' .claude/rules/moai/core/hooks-system.md
# Expected: NO "universal" framing; replaced by per-event enum
```

#### AC-HOC-015 (REQ-HOC-006, GAP-TOOL-04) — PostToolUse async constraint

**Given** the PostToolUse row,
**When** the async constraint is read,
**Then** the doctrine **shall** state that async PostToolUse can only deliver `additionalContext` (cannot control `decision` or `updatedToolOutput`).

```bash
grep -i 'PostToolUse' .claude/rules/moai/core/hooks-system.md | grep -iE 'async.*additionalContext|async.*cannot.*block'
# Expected: 1+ match
```

#### AC-HOC-016 (REQ-HOC-007, GAP-GD-03) — hook-independence.md row (g) timeout

**Given** the `hook-independence.md` cross-tab row (g),
**When** the timeout cell is read,
**Then** it **shall** show `60s (Stop)` (was `10s`).

```bash
grep 'sync-phase-quality-gate\|row.*g' .claude/rules/moai/development/hook-independence.md | grep '60'
# Expected: 1+ match showing 60s
```

### M3 — Async observation taps (MEDIUM)

#### AC-HOC-017 (REQ-HOC-008, GAP-OBS-01/02) — 3 observe taps async

**Given** the settings.json.tmpl UserPromptSubmit/Stop/SubagentStop harness-observe entries,
**When** the JSON is parsed,
**Then** each entry **shall** contain `"async": true`.

```bash
python3 -c "
import json
with open('internal/template/templates/.claude/settings.json.tmpl') as f: src=f.read()
# strip Go-template delimiters naively for the check
for evt in ['UserPromptSubmit','Stop','SubagentStop']:
    # find harness-observe co-located with the event
    print(evt, 'harness-observe' in src)
"
grep -B5 -A15 'harness-observe-user-prompt-submit\|harness-observe-stop\|harness-observe-subagent-stop' internal/template/templates/.claude/settings.json.tmpl | grep -c 'async.*true'
# Expected: ≥3 (one per UserPromptSubmit + Stop + SubagentStop observe entry)
```

#### AC-HOC-018 (REQ-HOC-009, GAP-LC2-02) — InstructionsLoaded async

**Given** the InstructionsLoaded entry,
**When** the JSON is parsed,
**Then** it **shall** contain `"async": true`.

```bash
grep -A15 'InstructionsLoaded' internal/template/templates/.claude/settings.json.tmpl | grep 'async.*true'
# Expected: 1+ match
```

### M4 — Timeout headroom (MEDIUM)

#### AC-HOC-019 (REQ-HOC-010, GAP-STT-01) — TeammateIdle/TaskCompleted timeout

**Given** the settings.json.tmpl TeammateIdle and TaskCompleted blocks,
**When** the timeout value is read,
**Then** it **shall** be > 5 (raised toward the sync-gate 60s precedent; exact value per run-phase decision).

```bash
grep -A10 'TeammateIdle\|TaskCompleted' internal/template/templates/.claude/settings.json.tmpl | grep timeout | head -5
# Expected: timeout > 5 for both events (or the fast-pre-check + async-deferred alternative is documented in the run-phase report)
```

#### AC-HOC-020 (REQ-HOC-011, GAP-LC-03) — PreCompact timeout

**Given** the settings.json.tmpl PreCompact block (line 30),
**When** the timeout value is read,
**Then** it **shall** be > 5 (raised toward SessionStart 30s class).

```bash
sed -n '/PreCompact/,/timeout/p' internal/template/templates/.claude/settings.json.tmpl | head -15
# Expected: timeout > 5
```

### M5 — Matcher resolution + Go verification (MEDIUM)

#### AC-HOC-021 (REQ-HOC-012, GAP-GD-04) — FileChanged matcher resolution

**Given** Go inspection of the FileChanged handler / matcher engine,
**When** the matcher semantics are observed,
**Then** the SPEC **shall** document whether the runtime treats FileChanged matchers as literal filenames or regex; **where** literal-only, settings.json.tmpl **shall** register individual entries per filename (no pipe-delimited single matcher).

```bash
# Go inspection evidence recorded in run-phase E7 report:
#   internal/hook/file_changed.go + matcher-engine path observation
# PLUS either:
#   (a) settings.json.tmpl FileChanged block shows individual literal entries, OR
#   (b) hooks-system.md documents the shipping runtime treats it as regex
```

#### AC-HOC-022 (REQ-HOC-013, GAP-CEW-02) — ConfigChange block channel

**Given** Go inspection of `internal/hook/config_change.go`,
**When** the block emission path is observed,
**Then** the run-phase report **shall** state whether the handler emits stdout JSON (`decision:block+reason`, `hookEventName:ConfigChange`) or exit 2 + stderr. **Where** exit 2 + stderr, the `handle-config-change.sh.tmpl` stderr redirect **shall** be made event-conditional.

```bash
# Go inspection evidence recorded in run-phase E7 report:
#   internal/hook/config_change.go observation
```

### M6 — Fail-open semantics correction (MEDIUM)

#### AC-HOC-023 (REQ-HOC-014, GAP-CEW-01) — WorktreeCreate fail-closed guard

**Given** the `handle-worktree-create.sh.tmpl` missing-binary fallback,
**When** the wrapper is read,
**Then** it **shall** either (a) carry a header comment stating it MUST NOT be registered unless moai is guaranteed resolvable + corrected "Claude Code handles missing hooks gracefully" comment, OR (b) emit non-zero exit + stderr diagnostic on the missing-binary branch.

```bash
head -20 internal/template/templates/.claude/hooks/moai/handle-worktree-create.sh.tmpl | grep -iE 'MUST NOT.*register|guaranteed.*resolvable|fail.*closed|abort'
# Expected: 1+ match (header guard OR loud-abort branch)
```

#### AC-HOC-024 (REQ-HOC-015, GAP-PE-01) — PermissionRequest fail-open warning priority

**Given** the standing SessionStart probe (SPEC-HOOK-SESSIONSTART-PROBE-001),
**When** the probe fires for the PermissionRequest wrapper fail-open case,
**Then** the probe output **shall** prioritize / explicitly flag the security-negative silent-allow (exit 0 = ALLOW neutralizes the permission gate) over the observer-event silent-degradation.

```bash
grep -iE 'PermissionRequest|permission.*allow|security.*negative' .claude/hooks/moai/handle-session-start.sh | head -5
# Expected: 1+ match showing the PermissionRequest-priority dimension in the probe warning
```

### M7 — Input hardening (LOW)

#### AC-HOC-025 (REQ-HOC-016, GAP-STT-04/TOOL-02) — MOAI_HOOK_STDERR_LOG allowlist

**Given** the shared wrapper pattern reads `MOAI_HOOK_STDERR_LOG`,
**When** the value is used in `mkdir -p` / `mv -f` / `2>>`,
**Then** the value **shall** be validated against an allowlist prefix (`$HOME/.moai/logs` or `$CLAUDE_PROJECT_DIR/.moai/logs`); mismatches **shall** fall back to a default.

```bash
grep -rn 'MOAI_HOOK_STDERR_LOG' internal/template/templates/.claude/hooks/moai/ | grep -iE 'allowlist|prefix|default.*fallback|case.*in' | head -5
# Expected: 1+ match showing the allowlist-validation applied in the shared pattern
```

#### AC-HOC-026 (REQ-HOC-017, GAP-GD-08) — gateguard state-file escaping

**Given** the `gateguard-fact-force.sh` state-file write (line ~107),
**When** the write interpolates `$session_id` / `$file_path` / `$tool_name`,
**Then** the values **shall** be awk-escaped (as the systemMessage path already does) OR recorded as plain `key=value` (no JSON).

```bash
sed -n '100,115p' internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: state-file write uses awk-escaping OR plain key=value (NOT raw interpolation into printf '{...}')
```

### M8 — Coverage holes + defects (LOW)

#### AC-HOC-027 (REQ-HOC-018, GAP-GD-05) — MultiEdit matcher extension

**Given** the settings.json.tmpl PostToolUse status-transition matcher,
**When** the matcher value is read,
**Then** it **shall** be `Write|Edit|MultiEdit`.

```bash
grep -A30 'PostToolUse' internal/template/templates/.claude/settings.json.tmpl | grep -E 'Write.*Edit.*MultiEdit|Write.*Edit'  | head -3
# Expected: matcher includes MultiEdit
```

#### AC-HOC-028 (REQ-HOC-019, GAP-GD-07) — csharp glob fix

**Given** the `sync-phase-quality-gate.sh` `detect_language` csharp branch,
**When** the csproj check is read,
**Then** it **shall** use `find "$root" -maxdepth 1 -name '*.csproj' -print -quit` (NOT `[ -f "$root/*.csproj" ]`).

```bash
grep -n 'csproj' internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh
# Expected: the find form (NOT the [ -f .../*.csproj ] dead glob)
```

#### AC-HOC-029 (REQ-HOC-020, GAP-OBS-03) — agent-hook exec form

**Given** the 4 agent files (manager-develop, manager-docs, manager-spec, sync-auditor) frontmatter `handle-agent-hook.sh` invocation,
**When** the command string is read,
**Then** the `{action}` token **shall** be quoted (`"{action}"`) or the invocation **shall** use exec form (`args=["{action}"]`).

```bash
grep -rn 'handle-agent-hook' .claude/agents/moai/*.md | grep -cE '\\"[a-z-]+\\"'
# Expected: 7 — all 7 invocations across the 4 agent files carry an explicitly
#           quoted action token (e.g. \"docs-verification\", \"develop-completion\").
#           The tokens are YAML-escaped (backslash-quote) in the frontmatter
#           command strings, hence the \\" anchor in the grep pattern.
```

#### AC-HOC-030 (REQ-HOC-021, GAP-LC-02) — compact naming/comment fix

**Given** the `handle-compact.sh.tmpl` comment naming the deployed hook,
**When** the comment is read,
**Then** it **shall** name the deployed `handle-compact.sh` (NOT the stale `compact.sh`). If renamed to `handle-pre-compact.sh.tmpl`, the settings.json.tmpl + moai hook subcommand mapping **shall** be updated atomically.

```bash
sed -n '12,16p' internal/template/templates/.claude/hooks/moai/handle-compact.sh.tmpl
# Expected: comment names handle-compact.sh (or the file is renamed to handle-pre-compact.sh.tmpl with atomic settings.json.tmpl update)
```

### Cross-cutting

#### AC-HOC-031 — Template-mirror parity (all milestones)

**Given** all milestone edits land,
**When** `go test ./internal/template/...` runs,
**Then** `TestRuleTemplateMirror*` **shall** PASS for all doctrine files touched.

```bash
go test ./internal/template/... -run TestRuleTemplateMirror -v 2>&1 | tail -5
# Expected: PASS
```

#### AC-HOC-032 — Template-neutrality CI guard

**Given** all template edits land,
**When** the template-neutrality CI guard runs,
**Then** it **shall** PASS (no SPEC IDs / REQ tokens / commit SHAs leaked into templates).

```bash
# CI guard: .github/workflows/template-neutrality-check.yaml
go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v 2>&1 | tail -5
# Expected: PASS
```

#### AC-HOC-033 — Subagent boundary (C-HRA-008)

**Given** the run-phase completes,
**When** `grep -rn 'AskUserQuestion' internal/hook/ internal/cli/ | grep -v _test.go | grep -v "// "` runs,
**Then** there **shall** be 0 matches.

```bash
grep -rn 'AskUserQuestion' internal/hook/ internal/cli/ | grep -v _test.go | grep -v "// "
# Expected: 0 matches (no NEW introduction; baseline preserved)
```

#### AC-HOC-034 — spec-lint clean

**Given** the SPEC artifacts land,
**When** `moai spec lint SPEC-HOOK-OFFICIAL-COMPLIANCE-001` runs,
**Then** it **shall** PASS (no findings beyond documented debt).

```bash
moai spec lint SPEC-HOOK-OFFICIAL-COMPLIANCE-001 2>&1 | tail -10
# Expected: PASS (or PASS-WITH-DEBT with documented debt items)
```

#### AC-HOC-035 — Whole-repo green

**Given** the run-phase completes,
**When** `go test ./...` runs,
**Then** it **shall** be green (no NEW regressions; pre-existing baseline preserved).

```bash
go test ./... 2>&1 | tail -10
# Expected: no NEW failures vs baseline
```

#### AC-HOC-036 — Cross-platform build

**Given** the run-phase completes,
**When** `GOOS=windows GOARCH=amd64 go build ./...` runs,
**Then** it **shall** exit 0.

```bash
GOOS=windows GOARCH=amd64 go build ./... 2>&1 | tail -5; echo "exit=$?"
# Expected: exit=0
```

---

## §D.1 Severity classification

| Severity | ACs | Blocking? |
|----------|-----|-----------|
| Must-pass (M1 HIGH) | AC-HOC-001 through AC-HOC-007 | YES — M1 is the HIGH-severity milestone |
| Must-pass (M2 doctrine) | AC-HOC-008 through AC-HOC-016 | YES — doctrine correctness cascades to all future hook authors |
| Should-pass (M3-M6 MEDIUM) | AC-HOC-017 through AC-HOC-024 | YES for sync-auditor PASS; PASS-WITH-DEBT allowed if any single AC is deferred with rationale |
| Nice-to-have (M7-M8 LOW) | AC-HOC-025 through AC-HOC-030 | PASS-WITH-DEBT allowed; deferral requires rationale |
| Cross-cutting | AC-HOC-031 through AC-HOC-036 | YES (CI guards) |

---

## §D.2 Traceability summary

| REQ | Gaps covered | Primary AC | Severity tier |
|-----|-------------|------------|---------------|
| REQ-HOC-001 | GAP-GD-01, GAP-STT-02 | AC-HOC-001 | Must-pass |
| REQ-HOC-002 | GAP-GD-02 | AC-HOC-002 | Must-pass |
| REQ-HOC-003 | GAP-TOOL-01 | AC-HOC-003, AC-HOC-004 | Must-pass |
| REQ-HOC-004 | GAP-LC2-01 | AC-HOC-005 | Must-pass |
| REQ-HOC-005 | GAP-STT-02 (doctrine) | AC-HOC-006 | Must-pass |
| REQ-HOC-006 | GAP-LC-01, GAP-STT-03, GAP-OBS-04, GAP-CEW-03, GAP-PE-02, GAP-PE-03, GAP-GD-06, GAP-TOOL-04 | AC-HOC-008..015 | Must-pass |
| REQ-HOC-007 | GAP-GD-03 | AC-HOC-016 | Must-pass |
| REQ-HOC-008 | GAP-OBS-01, GAP-OBS-02 | AC-HOC-017 | Should-pass |
| REQ-HOC-009 | GAP-LC2-02 | AC-HOC-018 | Should-pass |
| REQ-HOC-010 | GAP-STT-01 | AC-HOC-019 | Should-pass |
| REQ-HOC-011 | GAP-LC-03 | AC-HOC-020 | Should-pass |
| REQ-HOC-012 | GAP-GD-04 | AC-HOC-021 | Should-pass |
| REQ-HOC-013 | GAP-CEW-02 | AC-HOC-022 | Should-pass |
| REQ-HOC-014 | GAP-CEW-01 | AC-HOC-023 | Should-pass |
| REQ-HOC-015 | GAP-PE-01 | AC-HOC-024 | Should-pass |
| REQ-HOC-016 | GAP-STT-04, GAP-TOOL-02 | AC-HOC-025 | Nice-to-have |
| REQ-HOC-017 | GAP-GD-08 | AC-HOC-026 | Nice-to-have |
| REQ-HOC-018 | GAP-GD-05 | AC-HOC-027 | Nice-to-have |
| REQ-HOC-019 | GAP-GD-07 | AC-HOC-028 | Nice-to-have |
| REQ-HOC-020 | GAP-OBS-03 | AC-HOC-029 | Nice-to-have |
| REQ-HOC-021 | GAP-LC-02 | AC-HOC-030 | Nice-to-have |

---

## §D.3 Indirect verification (Go inspection — M1/M5)

The following gaps are hypotheses pending Go inspection per spec.md §E.2. The run-phase records the Go observation in the E7 report; the wrapper adjustment follows the observed Go behavior. The Go inspection itself is an indirect verification requirement (NOT a wrapper-only fix):

| Gap | Go file to inspect | What to observe |
|-----|-------------------|-----------------|
| GAP-TOOL-01 (REQ-HOC-003) | `internal/hook/pre_tool.go` + dispatcher | Does the handler emit exit 2 + stderr OR stdout JSON `permissionDecision:deny` for reject? |
| GAP-CEW-02 (REQ-HOC-013) | `internal/hook/config_change.go` | Does the handler emit stdout JSON `decision:block+reason` OR exit 2 + stderr? |
| GAP-GD-04 (REQ-HOC-012) | `internal/hook/file_changed.go` + matcher engine | Does the runtime treat FileChanged matchers as literal filenames or regex? |
| GAP-LC2-01 (REQ-HOC-004) | (no Go inspection — wrapper-only fix; hookEventName is a doctrine-field addition) | N/A |

---

## §D.4 Closure gates

**Definition of Done (spec-lint + acceptance):**
- AC-HOC-001 through AC-HOC-016 (M1 + M2 must-pass) ALL PASS
- AC-HOC-017 through AC-HOC-030 (M3-M8) ≥ 80% PASS; remainder PASS-WITH-DEBT with documented rationale
- AC-HOC-031 through AC-HOC-036 (cross-cutting) ALL PASS
- spec-lint clean
- Whole-repo green (no NEW regressions)
- Template-mirror parity holds

---

## §D.7 Forward-looking checks (post-close)

- **Runtime-tightening readiness**: after this SPEC closes, a future Claude Code runtime that tightens JSON-schema validation will NOT silently disable the blocking quality gates (M1 fixes).
- **Doctrine cascade**: the hooks-system.md refresh (M2) cascades to all future hook authors — they will read correct field names, matcher semantics, and exit-code enums.
- **SessionStart probe extension point**: REQ-HOC-015 (PermissionRequest warning priority) extends the existing probe; future per-event fail-open semantics can extend the same probe dimensionally.

---

## §E. Known Audit Debt (plan-auditor PASS-WITH-DEBT 0.86 — documented, not fixed)

The following 4 minor defects were identified by the independent plan-auditor and are documented here as known debt rather than fixed inline, to avoid REQ/AC renumbering and keep the audit trail transparent. They do NOT block run-phase entry (verdict was PASS-WITH-DEBT, not FAIL).

- **D3 (MINOR) — AC-HOC-017 dead `python3 -c` block.** The AC-HOC-017 verification contains a `python3 -c` introspection block that is non-functional/decorative; the actual verification load is the second `grep` command. The Python block should be removed in a future cleanup pass. Does not affect AC validity (the grep is the load-bearing check).
- **D5 (MINOR) — REQ heading labels "Event-detected" vs canonical "Event-driven".** REQ-HOC-001, REQ-HOC-012, REQ-HOC-014, REQ-HOC-015 carry the parenthetical label "(Event-detected, ...)" in their heading, whereas the canonical GEARS primary-pattern name is "Event-driven" (the REQ bodies already use the correct `**When**` keyword, so the REQs are GEARS-compliant in substance). Note: "Event-detected" is a recognized GEARS sub-pattern per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format (the form that replaces the deprecated IF/THEN for unwanted-condition-detected responses); the label choice is defensible for REQs whose trigger is an undesired condition, but if the project house style collapses to "Event-driven" the heading labels should be updated in a future cleanup pass. Bodies unchanged.
- **D6 (MINOR) — AC-HOC-019/020 timeout boundary loose.** The two ACs assert "shall be > 5" without anchoring to the report's upper-target precedent (60s for TeammateIdle/TaskCompleted per sync-gate, 30s for PreCompact per SessionStart). The loose lower bound is intentionally permissive (run-phase decides the exact value within the documented range), but a future tightening should reference the precedent ceiling to narrow the acceptance band.
- **D7 (MINOR) — ACs use Given/When/Then rather than canonical GEARS-AC form.** The REQs are GEARS-compliant (Ubiquitous/Event-driven/State-driven/Capability-gate forms with `shall`); the ACs are hybrid Given/When/Then-bearing-`shall`. A future pass may realign the ACs to a canonical GEARS-AC form, but the current hybrid form is readable and the `shall` binding is preserved on both REQ and AC surfaces.

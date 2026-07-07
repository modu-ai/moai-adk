# SPEC-HOOK-FACTFORCE-ADVISORY-001 — Acceptance Criteria

> Companion to `spec.md` (canonical requirements) and `plan.md` (implementation approach). This document owns the Given-When-Then acceptance criteria, edge cases, quality gates, and Definition of Done.

## §A. AC Index

| AC ID | Title | Severity | Requirement |
|-------|-------|----------|-------------|
| AC-FA-001 | First-edit advisory on Edit/Write/MultiEdit (exit 0 + systemMessage) | MUST | REQ-FA-001 |
| AC-FA-002 | Second-edit no-advisory on same file path | MUST | REQ-FA-002 |
| AC-FA-003 | State file keyed by session_id + absolute path (per-session advisory) | MUST | REQ-FA-003 |
| AC-FA-004 | MOAI_FACT_FORCE=off advisory opt-out + skip log | MUST | REQ-FA-004 |
| AC-FA-005 | Zero exit-2 paths + no AskUserQuestion (hook discipline grep) | MUST | REQ-FA-005 |
| AC-FA-006 | Self-termination within 5s timeout | MUST | REQ-FA-006 |
| AC-FA-007 | Existing handle-pre-tool.sh behavior preserved | MUST | REQ-FA-007 |
| AC-FA-008 | No PostToolUse "*" registration | MUST | REQ-FA-008 |
| AC-FA-009 | Subagent context handled correctly | SHOULD | REQ-FA-009 |
| AC-FA-010 | Self-loop prevention on non-Edit/Write/MultiEdit/Read payloads | MUST | REQ-FA-010 |
| AC-FA-011 | Template-First mirror parity (template + local) | MUST | REQ-FA-011 |
| AC-FA-012 | State file hygiene (0o600, JSON-line, no active deletion) | SHOULD | REQ-FA-012 |
| AC-FA-013 | Read-as-investigation (prior Read → next Write skips advisory) | MUST | REQ-FA-001 + REQ-FA-002 |
| AC-FA-014 | systemMessage JSON validity (parses via jq) | MUST | REQ-FA-001 + §C.6 |
| AC-FA-015 | Fail-open on missing session_id (exit 0, no advisory) | MUST | REQ-FA-005 + §C.2 |

**MUST count**: 13. **SHOULD count**: 2. **Total**: 15.

## §B. AC Matrix (Given-When-Then)

### AC-FA-001: First-edit advisory on Edit/Write/MultiEdit

**Given** the GateGuard Fact-Force hook is registered in `.claude/settings.json` as a PreToolUse matcher group scoped to `Edit|Write|MultiEdit`, AND the state directory `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/` is writable, AND no state file exists for the tuple `(session_id, absolute_file_path)`.

**When** the hook receives a PreToolUse payload with `tool_name` ∈ {`Edit`, `Write`, `MultiEdit`} and a `tool_input.file_path` field.

**Then** the hook MUST (a) write a new state file at `${state_dir}/<hash>` containing a single JSON line with `session_id`, `path` (absolute), `first_seen` (ISO-8601 UTC timestamp), and `via` (the tool_name), (b) emit a `systemMessage` advisory JSON on **stdout** whose text contains the substrings `First-edit advisory`, `IMPORTERS`, `DATA SCHEMAS`, and `USER INSTRUCTION`, and (c) exit with code **0** (allow the operation to proceed).

**Verification command**:
```bash
rm -rf /tmp/test-state && mkdir -p /tmp/test-state
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/test-file.go"},"session_id":"test-session-001"}
EOF
echo "exit=$?"
ls /tmp/test-state/.moai/state/fact-force/
```
**Expected**: exit code **0**; one file under `/tmp/test-state/.moai/state/fact-force/`; stdout contains a valid JSON object whose `systemMessage` field contains the 4 substrings.

---

### AC-FA-002: Second-edit no-advisory on same file path

**Given** AC-FA-001 has just executed successfully (the state file for `(test-session-001, /tmp/test-file.go)` now exists).

**When** the hook receives a second PreToolUse payload with the SAME `session_id` (`test-session-001`) and the SAME `tool_input.file_path` (`/tmp/test-file.go`).

**Then** the hook MUST exit with code 0 (allow), emit NO output on stdout (no systemMessage), and NOT modify the state file's `first_seen` timestamp.

**Verification command**:
```bash
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/test-file.go"},"session_id":"test-session-001"}
EOF
echo "exit=$?"
```
**Expected**: exit code 0; **no stdout output**; state file unchanged.

---

### AC-FA-003: State file keyed by session_id + absolute path (per-session advisory)

**Given** the hook has emitted an advisory on `/tmp/test-file.go` in session `test-session-001` (the state file for that tuple exists).

**When** the hook receives a PreToolUse payload with a DIFFERENT session_id (`test-session-002`) but the SAME `file_path` (`/tmp/test-file.go`).

**Then** the hook MUST emit a new advisory (exit 0 + systemMessage on stdout), proving that first-edit detection is preserved despite the switch from block to advisory. The state directory MUST now contain TWO files (one per session_id).

**Verification command**:
```bash
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/test-file.go"},"session_id":"test-session-002"}
EOF
echo "exit=$?"
ls /tmp/test-state/.moai/state/fact-force/ | wc -l
```
**Expected**: exit code **0**; stdout contains the advisory JSON; TWO files under the state directory.

---

### AC-FA-004: MOAI_FACT_FORCE=off advisory opt-out + skip log

**Given** the environment variable `MOAI_FACT_FORCE=off` is set in the hook's environment, AND no state file exists for the target tuple.

**When** the hook receives a PreToolUse payload that would otherwise trigger AC-FA-001 (first edit on a new file path).

**Then** the hook MUST (a) exit with code 0, (b) emit NO advisory systemMessage on stdout, (c) append a one-line bypass record to `.moai/logs/fact-force-skip.log` containing timestamp, session_id, and file_path, AND (d) NOT write a state file.

**Verification command**:
```bash
rm -f /tmp/test-state/.moai/logs/fact-force-skip.log
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=off \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/new-file.go"},"session_id":"test-session-001"}
EOF
echo "exit=$?"
cat /tmp/test-state/.moai/logs/fact-force-skip.log
ls /tmp/test-state/.moai/state/fact-force/ 2>/dev/null | wc -l
```
**Expected**: exit code 0; no stdout advisory; one line in `fact-force-skip.log`; zero state files written for this tuple.

---

### AC-FA-005: Zero exit-2 paths + no AskUserQuestion (hook discipline grep)

**Given** the hook script exists at `.claude/hooks/moai/gateguard-fact-force.sh` and its template mirror at `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh`.

**When** a reviewer greps for exit-2 paths and AskUserQuestion invocations.

**Then** BOTH greps MUST return empty (zero matches for `exit 2`, zero matches for `AskUserQuestion` / `mcp__askuser`), confirming the hook never blocks and never prompts.

**Verification command**:
```bash
# Check 1: zero exit-2 paths
grep -n 'exit 2' \
  .claude/hooks/moai/gateguard-fact-force.sh \
  internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty (zero matches)

# Check 2: no AskUserQuestion
grep -E 'AskUserQuestion|mcp__askuser' \
  .claude/hooks/moai/gateguard-fact-force.sh \
  internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty (zero matches)
```

---

### AC-FA-006: Self-termination within 5s timeout

**Given** the hook is registered with `"timeout": 5` in settings.json.

**When** the hook receives a standard PreToolUse payload (~200 bytes JSON).

**Then** each hook invocation's wall-clock execution time (stdin read → exit) MUST be < 100ms, well within the 5s timeout. The implementation MUST NOT perform network I/O, LSP queries, or subprocess spawns beyond file-existence checks and single-line writes.

**Verification command**:
```bash
for i in 1 2 3 4 5; do
  /usr/bin/time -p bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF' 2>&1 | grep real
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/perf-test.go"},"session_id":"perf-$i"}
EOF
done
```
**Expected**: each `real` line is < 0.1s (100ms).

---

### AC-FA-007: Existing handle-pre-tool.sh behavior preserved

**Given** the settings.json carries the PreToolUse matcher group added by the predecessor SPEC.

**When** a reviewer inspects `.claude/settings.json`.

**Then** the existing `Write|Edit|Bash` matcher group MUST be UNCHANGED (same command, same timeout, same matcher string). This SPEC does NOT modify settings.json.

**Verification command**:
```bash
grep -A 6 '"PreToolUse"' .claude/settings.json
# Expected: the Edit|Write|MultiEdit → gateguard-fact-force.sh group (added by predecessor) is present;
#           the Write|Edit|Bash → handle-pre-tool.sh group is UNCHANGED.
```

---

### AC-FA-008: No PostToolUse "*" registration

**Given** the spec.md REQ-FA-008 prohibition on PostToolUse registration.

**When** a reviewer inspects `.claude/settings.json`.

**Then** there MUST be NO PostToolUse change attributable to this SPEC (the predecessor did not add one either).

**Verification command**:
```bash
git diff HEAD~1 HEAD -- .claude/settings.json | grep -E '^\+.*"PostToolUse"|^\-.*"PostToolUse"'
# Expected: empty (this SPEC makes no settings.json changes).
```

---

### AC-FA-009: Subagent context handled correctly

**Given** a PreToolUse payload that includes the Claude Code v2.1.69+ `agent_id` field AND a `session_id` matching the parent session.

**When** the hook processes the payload.

**Then** the hook MUST apply the same per-file-per-session advisory (NO subagent exemption), AND the state MUST be keyed by the parent `session_id` (NOT the `agent_id`), so that a file already advised in the parent session is NOT re-advised when a subagent edits it later in the same session.

**Verification command**:
```bash
# Step 1: orchestrator triggers advisory on /tmp/shared.go in session parent-001
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/shared.go"},"session_id":"parent-001"}
EOF
echo "step1 exit=$?"

# Step 2: subagent (agent_id=sub-abc) edits the same file in the same parent session
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/shared.go"},"session_id":"parent-001","agent_id":"sub-abc","agent_type":"general-purpose"}
EOF
echo "step2 exit=$?"
```
**Expected**: step1 exit=0 (advisory emitted), step2 exit=0 (no advisory — parent session already advised). One state file.

---

### AC-FA-010: Self-loop prevention on non-Edit/Write/MultiEdit/Read payloads

**Given** the hook is invoked with a payload whose `tool_name` is NOT one of `Edit`, `Write`, `MultiEdit`, or `Read` (e.g., `Bash`).

**When** the hook processes the payload.

**Then** the hook MUST exit 0 immediately (allow), write NO state file, emit NO advisory.

**Verification command**:
```bash
rm -rf /tmp/test-state-sl && mkdir -p /tmp/test-state-sl
CLAUDE_PROJECT_DIR=/tmp/test-state-sl MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls"},"session_id":"sl-001"}
EOF
echo "bash-tool exit=$?"
ls /tmp/test-state-sl/.moai/state/fact-force/ 2>/dev/null | wc -l
```
**Expected**: exit=0; zero state files.

---

### AC-FA-011: Template-First mirror parity (template + local)

**Given** the implementation followed the Template-First ordering (edit template tree → `make build` → mirror to local).

**When** a reviewer diffs the local hook script against its template source.

**Then** the two files MUST be byte-identical.

**Verification command**:
```bash
diff .claude/hooks/moai/gateguard-fact-force.sh \
     internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty diff.

test -x .claude/hooks/moai/gateguard-fact-force.sh && echo "local: executable"
test -x internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh && echo "template: executable"
```

---

### AC-FA-012: State file hygiene (0o600, JSON-line, no active deletion)

**Given** the hook has written at least one state file.

**When** a reviewer inspects the state file's permissions and content.

**Then** the state file MUST (a) be a single JSON line, (b) have filesystem permissions `0o600`, AND (c) NOT be deleted by the hook during the same session.

**Verification command**:
```bash
STATE_FILE=$(ls /tmp/test-state/.moai/state/fact-force/ | head -1)
# macOS
stat -f '%p %Lz' /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Linux
stat -c '%a %s' /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Expected: starts with 100600 (macOS) or 600 (Linux).

wc -l /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Expected: 1 line (single JSON line).

cat /tmp/test-state/.moai/state/fact-force/$STATE_FILE | python3 -m json.tool
# Expected: parses without error.
```

---

### AC-FA-013: Read-as-investigation (prior Read → next Write skips advisory)

**Given** the hook receives a `tool_name=Read` payload for `/tmp/read-first.go` in session `read-sess-001`, AND no state file exists for that tuple.

**When** (step 1) the Read payload is processed, THEN (step 2) a subsequent `tool_name=Write` payload for the SAME path and SAME session is processed.

**Then** step 1 MUST exit 0 with NO advisory on stdout (Read pre-populates state silently), AND a state file MUST be created. Step 2 MUST exit 0 with NO advisory on stdout (the state file from step 1 suppresses it — Read-as-investigation).

**Verification command**:
```bash
rm -rf /tmp/test-state-ri && mkdir -p /tmp/test-state-ri

# Step 1: Read on /tmp/read-first.go (pre-populate state)
CLAUDE_PROJECT_DIR=/tmp/test-state-ri MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"/tmp/read-first.go"},"session_id":"read-sess-001"}
EOF
echo "step1 exit=$?"
ls /tmp/test-state-ri/.moai/state/fact-force/ | wc -l
# Expected: step1 exit=0; no stdout advisory; 1 state file created.

# Step 2: Write on same path (should skip advisory — Read-as-investigation)
CLAUDE_PROJECT_DIR=/tmp/test-state-ri MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"/tmp/read-first.go"},"session_id":"read-sess-001"}
EOF
echo "step2 exit=$?"
# Expected: step2 exit=0; no stdout advisory (state file already exists from step 1).
```

---

### AC-FA-014: systemMessage JSON validity (parses via jq)

**Given** AC-FA-001 has produced an advisory JSON on stdout.

**When** a reviewer pipes that stdout through `jq`.

**Then** the output MUST be valid JSON such that `jq -e '.systemMessage | type'` returns `"string"` (the systemMessage field exists and is a string type). This proves the awk-based escape (§C.5) produced valid JSON without raw newlines.

**Verification command**:
```bash
# Capture stdout from AC-FA-001
stdout=$(CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"/tmp/json-validity.go"},"session_id":"json-test"}
EOF
)

# Verify JSON validity
echo "$stdout" | jq -e '.systemMessage | type'
# Expected: "string"

# Verify no raw newlines inside the JSON string (all escaped as \n)
echo "$stdout" | wc -l
# Expected: 1 line (the entire JSON is on one line — no embedded raw newlines)
```

---

### AC-FA-015: Fail-open on missing session_id (exit 0, no advisory)

**Given** a PreToolUse payload that lacks the `session_id` field (malformed or truncated payload).

**When** the hook processes the payload.

**Then** the hook MUST exit 0 (fail-open — allow the edit), emit NO advisory on stdout, and write NO state file. This is consistent with the §C.2 fail-open default: the hook cannot key a state file without a session_id, and fail-open (allow) is safer than fail-closed (block).

**Verification command**:
```bash
rm -rf /tmp/test-state-fail && mkdir -p /tmp/test-state-fail
CLAUDE_PROJECT_DIR=/tmp/test-state-fail MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/no-session.go"}}
EOF
echo "exit=$?"
ls /tmp/test-state-fail/.moai/state/fact-force/ 2>/dev/null | wc -l
```
**Expected**: exit=0; no stdout output; zero state files.

## §C. Edge Cases

| # | Edge case | Expected behavior | Covered by |
|---|-----------|-------------------|------------|
| E1 | Empty `file_path` in payload | Exit 0, no advisory (cannot key without a path) | AC-FA-015 implies |
| E2 | Unwritable state directory (`mkdir -p` fails) | Exit 0, no advisory (fail-open) | §C.2 |
| E3 | Malformed JSON payload (grep extracts nothing) | Exit 0, no advisory (fail-open) | §C.2 |
| E4 | Absent `session_id` field | Exit 0, no advisory (fail-open — AC-FA-015) | AC-FA-015 |
| E5 | Relative `file_path` (e.g., `./foo.go`) | Resolve to absolute via `${PWD}/<path>` before hashing | REQ-FA-003 |
| E6 | Path containing spaces or shell metacharacters | Hash-based filename avoids the issue; state file content is JSON-quoted | REQ-FA-003 + §C.2 |
| E7 | `MOAI_FACT_FORCE` set to value other than `off` (e.g., `false`, `0`, `no`) | Treated as `on` (only exact `off` opts out) | REQ-FA-004 |
| E8 | Hook invoked twice in rapid succession (race) | State file write is idempotent at FS level; second writer overwrites first; no corruption | §C.2 |
| E9 | Path uses `~` (home shorthand) | Treated literally (not expanded) — may produce a state file keyed on `~`-prefixed path | Open question (deferred) |
| E10 | Symlinked file path | Resolved to the literal path in the payload (no symlink resolution) | §C.2 |

## §D. Quality Gate Criteria

### §D.1 plan-auditor PASS threshold (Tier S)

The plan-auditor's aggregate score for this SPEC MUST be ≥ 0.75 (Tier S threshold per `spec-workflow.md` § SPEC Complexity Tier).

### §D.2 MUST-pass ACs at run-phase exit

All 13 MUST ACs (AC-FA-001 through AC-FA-008, AC-FA-010, AC-FA-011, AC-FA-013, AC-FA-014, AC-FA-015) MUST be PASS at M1 exit. The 2 SHOULD ACs (AC-FA-009, AC-FA-012) MAY be PASS-WITH-DEBT.

### §D.3 LSP / lint baseline preservation

`golangci-lint run` and `go test ./...` MUST show no NEW findings or failures (no Go changes were made). `moai spec lint .moai/specs/SPEC-HOOK-FACTFORCE-ADVISORY-001/` MUST return clean.

## §E. Definition of Done

The SPEC is "run-phase complete" when ALL of the following hold:

- [ ] All 13 MUST ACs PASS (verifiable per §B commands)
- [ ] The 2 SHOULD ACs are PASS or PASS-WITH-DEBT (with debt annotation)
- [ ] No NEW findings in `golangci-lint run` (no Go changes expected)
- [ ] No NEW failures in `go test ./...` (no Go changes expected)
- [ ] `moai spec lint` clean on this SPEC's artifacts
- [ ] The hook-independence.md § 3 Mode G row rationale is updated (`"no gate"` → `"no advisory notice"`)
- [ ] progress.md §E.2 / §E.3 are populated with run-phase evidence

The SPEC is "fully closed" when the sync-phase (manager-docs) additionally:
- [ ] Updates CHANGELOG.md with the advisory-rewrite entry
- [ ] Populates progress.md §E.4 with the sync_commit_sha
- [ ] Transitions frontmatter `status: in-progress → implemented → completed` on the single sync commit

## §F. Indirect Verification (where direct grep is insufficient)

For AC-FA-006 (5s timeout), direct wall-clock measurement is the primary verification. `/usr/bin/time -p` (POSIX form, `real <seconds>` on stderr) is used. Per-invocation check (< 100ms each), not a statistical percentile.

For AC-FA-009 (subagent context), the verification is functional — run the hook with a payload containing `agent_id`, verify the state-keying behavior. Requires the two-step sequence in §B.

For AC-FA-012 (0o600 permissions), the verification depends on the host's `stat` command (BSD vs GNU). Both forms provided in §B.

For AC-FA-014 (JSON validity), the verification USES `jq` to validate the hook's output, but the hook itself does NOT depend on `jq` at runtime (§C.5 jq prohibition). The verification tool (`jq`) and the hook's runtime dependency (none — jq-free) are separate concerns.

## §G. Forward-Looking Checks (out of current scope but documented for traceability)

| Check | Why documented | Owner |
|-------|----------------|-------|
| State file cleanup GC | REQ-FA-012 defers active cleanup; a follow-up SPEC should add SessionEnd sweep | follow-up SPEC |
| Go-side `internal/hook/fact_force.go` | spec.md §E defers; shell-only is sufficient for Tier S | follow-up SPEC IF measurement shows insufficient |
| Subagent `agent_type` exemption list | Not needed for advisory mode (advisory is non-blocking); may be revisited if a future SPEC restores blocking | follow-up SPEC |
| `~` home-shorthand path resolution | Edge case E9; not blocking | follow-up SPEC |
| Non-ASCII Unicode in systemMessage | §C.5 awk escape covers ASCII + newlines only; Unicode may require extended escape | follow-up SPEC IF advisory text is extended |

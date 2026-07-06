# SPEC-HOOK-PREEDIT-INVESTIGATE-001 — Acceptance Criteria

> Companion to `spec.md` (canonical requirements) and `plan.md` (implementation approach). This document owns the Given-When-Then acceptance criteria, edge cases, quality gates, and Definition of Done.

## §A. AC Index

| AC ID | Title | Severity | Requirement |
|-------|-------|----------|-------------|
| AC-FF-001 | First-edit block on Edit/Write/MultiEdit | MUST | REQ-FF-001 |
| AC-FF-002 | Second-edit allow on same file path | MUST | REQ-FF-002 |
| AC-FF-003 | State file keyed by session_id + absolute path | MUST | REQ-FF-003 |
| AC-FF-004 | MOAI_FACT_FORCE=off advisory opt-out + log | MUST | REQ-FF-004 |
| AC-FF-005 | No AskUserQuestion invocation | MUST | REQ-FF-005 |
| AC-FF-006 | Self-termination within 5s timeout | MUST | REQ-FF-006 |
| AC-FF-007 | Existing handle-pre-tool.sh behavior preserved | MUST | REQ-FF-007 |
| AC-FF-008 | No PostToolUse "*" registration | MUST | REQ-FF-008 |
| AC-FF-009 | Subagent context handled correctly | SHOULD | REQ-FF-009 |
| AC-FF-010 | Self-loop prevention on non-Edit/Write/MultiEdit payloads | MUST | REQ-FF-010 |
| AC-FF-011 | Template-First mirror parity (template + local) | MUST | REQ-FF-011 |
| AC-FF-012 | State file hygiene (0o600, JSON-line, no active deletion) | SHOULD | REQ-FF-012 |

**MUST count**: 10. **SHOULD count**: 2. **Total**: 12.

## §B. AC Matrix (Given-When-Then)

### AC-FF-001: First-edit block on Edit/Write/MultiEdit

**Given** the GateGuard Fact-Force hook is registered in `.claude/settings.json` as a PreToolUse matcher group scoped to `Edit|Write|MultiEdit`, AND the state directory `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/` is writable, AND no state file exists for the tuple `(session_id, absolute_file_path)`.

**When** the hook receives a PreToolUse payload with `tool_name` ∈ {`Edit`, `Write`, `MultiEdit`} and a `tool_input.file_path` field.

**Then** the hook MUST (a) write a new state file at `${state_dir}/<hash>` containing a single JSON line with `session_id`, `path` (absolute), and `first_seen` (ISO-8601 UTC timestamp), (b) emit a guidance message on stdout demanding the agent investigate (i) importers via grep, (ii) data schemas / contracts the file touches, (iii) the user instruction that justifies the edit, and (c) exit with code 2 to block the operation.

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
**Expected**: exit code 2; one file under `/tmp/test-state/.moai/state/fact-force/`; stdout contains the substring `FACT-FORCE GATE` and `IMPORTERS` and `DATA SCHEMAS` and `USER INSTRUCTION`.

---

### AC-FF-002: Second-edit allow on same file path

**Given** AC-FF-001 has just executed successfully (the state file for `(test-session-001, /tmp/test-file.go)` now exists).

**When** the hook receives a second PreToolUse payload with the SAME `session_id` (`test-session-001`) and the SAME `tool_input.file_path` (`/tmp/test-file.go`).

**Then** the hook MUST exit with code 0 (allow), emit NO guidance message on stdout, and NOT modify the state file's `first_seen` timestamp.

**Verification command**:
```bash
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/test-file.go"},"session_id":"test-session-001"}
EOF
echo "exit=$?"
```
**Expected**: exit code 0; no stdout output; state file unchanged.

---

### AC-FF-003: State file keyed by session_id + absolute path

**Given** the hook has blocked an edit on `/tmp/test-file.go` in session `test-session-001`.

**When** the hook receives a PreToolUse payload with a DIFFERENT session_id (`test-session-002`) but the SAME `file_path` (`/tmp/test-file.go`).

**Then** the hook MUST treat this as a first-edit (block, exit 2, write a new state file), proving that state is keyed by the COMPOSITE `(session_id, absolute_path)` tuple, not by `absolute_path` alone.

**Verification command**:
```bash
CLAUDE_PROJECT_DIR=/tmp/test-state MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/test-file.go"},"session_id":"test-session-002"}
EOF
echo "exit=$?"
ls /tmp/test-state/.moai/state/fact-force/ | wc -l
```
**Expected**: exit code 2; TWO files under the state directory (one per session_id).

---

### AC-FF-004: MOAI_FACT_FORCE=off advisory opt-out + log

**Given** the environment variable `MOAI_FACT_FORCE=off` is set in the hook's environment, AND no state file exists for the target tuple.

**When** the hook receives a PreToolUse payload that would otherwise trigger AC-FF-001 (first edit on a new file path).

**Then** the hook MUST (a) exit with code 0 (skip the gate), (b) emit NO guidance message, (c) append a one-line bypass record to `.moai/logs/fact-force-skip.log` containing timestamp, session_id, and file_path, AND (d) NOT write a state file (so a subsequent edit with the env var unset WILL be gated as a first edit).

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
**Expected**: exit code 0; no stdout guidance; one line in `fact-force-skip.log`; zero state files written for this tuple.

---

### AC-FF-005: No AskUserQuestion invocation

**Given** the hook script exists at `.claude/hooks/moai/gateguard-fact-force.sh` and its template mirror at `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh`.

**When** a reviewer runs the subagent-boundary grep acceptance criterion.

**Then** the grep MUST return empty (zero matches), confirming the hook script has no AskUserQuestion surface.

**Verification command**:
```bash
grep -E 'AskUserQuestion|mcp__askuser' \
  .claude/hooks/moai/gateguard-fact-force.sh \
  internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty output (zero matches)
```

---

### AC-FF-006: Self-termination within 5s timeout

**Given** the hook is registered with `"timeout": 5` in settings.json (the PreToolUse matcher group entry).

**When** the hook receives a standard PreToolUse payload (~200 bytes JSON).

**Then** each hook invocation's wall-clock execution time (stdin read → exit) MUST be < 100ms, well within the 5s timeout. The 5-iteration sample in the verification below is a practical per-invocation check (each sample must be < 100ms), NOT a statistical p99 measurement. The implementation MUST NOT perform network I/O, LSP queries, or subprocess spawns beyond `[ -f ... ]` and `cat > ...`.

**Verification command**:
```bash
for i in 1 2 3 4 5; do
  /usr/bin/time -p bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF' 2>&1 | grep real
{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/tmp/perf-test.go"},"session_id":"perf-$i"}
EOF
done
```
**Expected**: each `real` line is < 0.1s (100ms). macOS `/usr/bin/time -p` outputs `real <seconds>` on stderr.

---

### AC-FF-007: Existing handle-pre-tool.sh behavior preserved

**Given** the settings.json edit adds a NEW PreToolUse matcher group scoped to `Edit|Write|MultiEdit`.

**When** a reviewer diffs `.claude/settings.json` against its pre-commit baseline.

**Then** the diff MUST show (a) the NEW matcher group added as a new array entry inside the `PreToolUse` array, AND (b) the EXISTING `Write|Edit|Bash` matcher group UNCHANGED (same command, same timeout, same matcher string, same array position relative to the start of the PreToolUse array).

**Verification command**:
```bash
git diff HEAD~1 HEAD -- .claude/settings.json | grep -A 6 '"PreToolUse"'
# Expected: TWO matcher groups visible — original (Write|Edit|Bash) unchanged, new (Edit|Write|MultiEdit) added.

git show HEAD:.claude/settings.json | grep -A 6 '"PreToolUse"' | grep -c 'Write|Edit|Bash'
# Expected: 1 (the original group is preserved in the new HEAD).

git diff HEAD~1 HEAD -- .claude/settings.json | grep -E '^[+-].*"matcher".*"Write\|Edit\|Bash"'
# Expected: empty (no +/- lines touching the original matcher).
```

---

### AC-FF-008: No PostToolUse "*" registration

**Given** the spec.md REQ-FF-008 prohibition on PostToolUse registration of any kind.

**When** a reviewer diffs `.claude/settings.json` against its pre-commit baseline.

**Then** the diff MUST show ZERO changes inside the `PostToolUse` array — no new matcher, no removed matcher, no modified command.

**Verification command**:
```bash
git diff HEAD~1 HEAD -- .claude/settings.json | grep -E '^\+.*"PostToolUse"|^\-.*"PostToolUse"'
# Expected: empty.

git diff HEAD~1 HEAD -- .claude/settings.json | awk '/"PostToolUse"/{p=1} /^[+-]/{if(p) print} /^]/{p=0}'
# Expected: no +/- lines between the PostToolUse key and the closing bracket.
```

---

### AC-FF-009: Subagent context handled correctly

**Given** a PreToolUse payload that includes the Claude Code v2.1.69+ `agent_id` field (indicating subagent context) AND a `session_id` matching the parent session.

**When** the hook processes the payload.

**Then** the hook MUST apply the same per-file-per-session gate (NO subagent exemption), AND the state MUST be keyed by the parent `session_id` (NOT the `agent_id`), so that a file investigated by the orchestrator in the parent session is NOT re-gated when a subagent edits it later in the same session.

**Verification command**:
```bash
# Step 1: orchestrator investigates the file (first edit on /tmp/shared.go in session parent-001)
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
**Expected**: step1 exit=2 (first edit blocked), step2 exit=0 (subagent edit allowed because parent session already investigated). One state file under the state directory.

---

### AC-FF-010: Self-loop prevention on non-Edit/Write/MultiEdit payloads

**Given** the hook is invoked with a payload whose `tool_name` is NOT one of `Edit`, `Write`, `MultiEdit` (e.g., the hook was accidentally invoked on its own Write to the state file, or on a `Read` tool call).

**When** the hook processes the payload.

**Then** the hook MUST exit 0 immediately (allow), write NO state file, emit NO guidance message. This ensures the hook cannot gate its own state-file writes or otherwise recurse.

**Verification command**:
```bash
rm -rf /tmp/test-state-sl && mkdir -p /tmp/test-state-sl
# tool_name=Read → not gated
CLAUDE_PROJECT_DIR=/tmp/test-state-sl MOAI_FACT_FORCE=on \
  bash .claude/hooks/moai/gateguard-fact-force.sh <<'EOF'
{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"/tmp/anything.go"},"session_id":"sl-001"}
EOF
echo "read-tool exit=$?"
ls /tmp/test-state-sl/.moai/state/fact-force/ 2>/dev/null | wc -l
```
**Expected**: exit=0; zero state files.

---

### AC-FF-011: Template-First mirror parity (template + local)

**Given** the implementation followed the Template-First ordering (edit template tree → `make build` → mirror to local).

**When** a reviewer diffs the local hook script against its template source.

**Then** the two files MUST be byte-identical (the local mirror is a verbatim copy of the template, modulo any future template-variable substitution that does not apply to a shell script with no template variables).

**Verification command**:
```bash
diff .claude/hooks/moai/gateguard-fact-force.sh \
     internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty diff.

# Also verify both are executable
test -x .claude/hooks/moai/gateguard-fact-force.sh && echo "local: executable"
test -x internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh && echo "template: executable"
```

---

### AC-FF-012: State file hygiene (0o600, JSON-line, no active deletion)

**Given** the hook has written at least one state file.

**When** a reviewer inspects the state file's permissions and content.

**Then** the state file MUST (a) be a single JSON line (no multi-line JSON, no trailing newline-only lines), (b) have filesystem permissions `0o600` (read/write owner only — protects absolute file paths from other users on shared hosts), AND (c) NOT be deleted by the hook during the same session (cleanup is a separate follow-up SPEC).

**Verification command**:
```bash
STATE_FILE=$(ls /tmp/test-state/.moai/state/fact-force/ | head -1)
# macOS
stat -f '%p %Lz' /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Linux
stat -c '%a %s' /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Expected: starts with 100600 (macOS) or 600 (Linux); file size is reasonable (100-300 bytes).

# Verify single JSON line
wc -l /tmp/test-state/.moai/state/fact-force/$STATE_FILE
# Expected: 1 line (or 2 if a trailing newline counts, but content is single-line JSON).

# Verify JSON validity
cat /tmp/test-state/.moai/state/fact-force/$STATE_FILE | python3 -m json.tool
# Expected: parses without error, shows session_id / path / first_seen fields.
```

## §C. Edge Cases

| # | Edge case | Expected behavior | Covered by |
|---|-----------|-------------------|------------|
| E1 | Empty `file_path` in payload | Exit 0 (cannot gate without a path) | REQ-FF-001 implies |
| E2 | Unwritable state directory (`mkdir -p` fails) | Exit 0 (fail-open, no gate) | §C.2 Robustness |
| E3 | Malformed JSON payload (grep extracts nothing) | Exit 0 (fail-open) | §C.2 Robustness |
| E4 | Absent `session_id` field | Exit 0 (fail-open, allow edit) — cannot key state file without session_id; consistent with §C.2 robustness fail-open default | §C.2 Robustness |
| E5 | Relative `file_path` (e.g., `./foo.go`) | Resolve to absolute via `${PWD}/<path>` before hashing | REQ-FF-003 implies |
| E6 | Path containing spaces or shell metacharacters | Hash-based filename avoids the issue; state file content is JSON-quoted | REQ-FF-003 + §C.2 |
| E7 | `MOAI_FACT_FORCE` set to value other than `off` (e.g., `false`, `0`, `no`) | Treated as `on` (only exact `off` opts out) | REQ-FF-004 |
| E8 | Hook invoked twice in rapid succession (race) | The state file write is idempotent at the filesystem level; second writer overwrites first; no corruption | §C.2 |
| E9 | Path uses `~` (home shorthand) | Treated literally (not expanded) — may produce a state file keyed on `~`-prefixed path; the agent should use absolute paths | Open question for plan-auditor |
| E10 | Symlinked file path | Resolved to the literal path in the payload (no symlink resolution) — symlinks to the same file count as different paths | §C.2 |

## §D. Quality Gate Criteria

### §D.1 plan-auditor PASS threshold (Tier S)

The plan-auditor's aggregate score for this SPEC MUST be ≥ 0.75 (Tier S threshold per `spec-workflow.md` § SPEC Complexity Tier). The plan-auditor evaluates the plan-phase artifacts (spec.md + plan.md + acceptance.md + progress.md) on the standard 4 dimensions (Scope / Coverage / Traceability / Anti-pattern). A score < 0.75 triggers iteration.

### §D.2 MUST-pass ACs at run-phase exit

All 10 MUST ACs (AC-FF-001 through AC-FF-008, AC-FF-010, AC-FF-011) MUST be PASS at M1 exit. The 2 SHOULD ACs (AC-FF-009, AC-FF-012) MAY be PASS-WITH-DEBT with an explicit debt annotation in the M1 commit body — but ONLY if the debt is recorded in `hook-independence.md` § 3 with a follow-up owner.

### §D.3 LSP / lint baseline preservation

`golangci-lint run` and `go test ./...` MUST show no NEW findings or failures (no Go changes were made). `moai spec lint .moai/specs/SPEC-HOOK-PREEDIT-INVESTIGATE-001/` MUST return clean (zero findings on this SPEC's artifacts).

## §E. Definition of Done

The SPEC is "run-phase complete" when ALL of the following hold:

- [ ] All 10 MUST ACs PASS (verifiable per §B commands)
- [ ] The 2 SHOULD ACs are PASS or PASS-WITH-DEBT (with debt annotation)
- [ ] No NEW findings in `golangci-lint run` (no Go changes expected)
- [ ] No NEW failures in `go test ./...` (no Go changes expected)
- [ ] `moai spec lint` clean on this SPEC's artifacts
- [ ] The 4-file artifact set (spec.md + plan.md + acceptance.md + progress.md) is committed with `Authored-By-Agent: manager-develop` trailer on the M1 commit
- [ ] The commit subject follows the canonical pattern: `feat(SPEC-HOOK-PREEDIT-INVESTIGATE-001): M1 ...` with `🗿 MoAI` trailer
- [ ] The hook-independence.md § 3 catalogue is updated with the new shared-failure-mode row
- [ ] progress.md §E.2 / §E.3 are populated with run-phase evidence + audit-ready signal

The SPEC is "fully closed" when the sync-phase (manager-docs) additionally:
- [ ] Updates CHANGELOG.md with the new hook entry
- [ ] Populates progress.md §E.4 with the sync_commit_sha
- [ ] Transitions frontmatter `status: in-progress → implemented → completed` on the single sync commit

## §F. Indirect Verification (where direct grep is insufficient)

For AC-FF-006 (5s timeout), direct wall-clock measurement is the primary verification. The `time` command on macOS vs Linux has different output formats; the verification command in §B uses `/usr/bin/time -p` (POSIX form, `real <seconds>` on stderr). If the implementing agent reports each invocation's wall-clock < 100ms (all 5 samples under the threshold), that satisfies the AC. This is a per-invocation check, not a statistical percentile — the 5-sample size is insufficient for p99 framing.

For AC-FF-009 (subagent context), the verification is functional (run the hook with a payload containing `agent_id`, verify the state-keying behavior). A static grep cannot prove this — it requires the two-step sequence in §B.

For AC-FF-012 (0o600 permissions), the verification depends on the host's `stat` command (BSD vs GNU). The verification command in §B provides both forms.

## §G. Forward-Looking Checks (out of current scope but documented for traceability)

| Check | Why documented | Owner |
|-------|----------------|-------|
| State file cleanup GC | REQ-FF-012 defers active cleanup; a follow-up SPEC should add SessionEnd sweep | follow-up SPEC |
| Go-side `internal/hook/fact_force.go` | spec.md §E defers; shell-only is sufficient for Tier S | follow-up SPEC IF measurement shows insufficient |
| Subagent `agent_type` exemption list | spec.md §F.3 open question; a follow-up SPEC may add an exemption list (e.g., builder-harness) | follow-up SPEC |
| `~` home-shorthand path resolution | Edge case E9; not blocking but should be resolved in a follow-up | follow-up SPEC |
| Confidence scoring / learning | spec.md §E explicitly out of scope; NEVER a follow-up — abandoned candidate A | (never) |

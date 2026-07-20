# SPEC-HOOK-FACTFORCE-ADVISORY-001 — Implementation Plan

> Companion to `spec.md` (canonical requirements) and `acceptance.md` (canonical ACs). This document owns the implementation approach: file-by-file change list, milestone structure, the 3 change-sites in the hook script, and self-verification deliverables.

## §A. Context

### §A.1 Scope summary

Rewrite the existing `gateguard-fact-force.sh` PreToolUse hook (shipped by the predecessor SPEC-HOOK-PREEDIT-INVESTIGATE-001) from exit-2-blocking to exit-0-advisory. The FIRST Edit/Write/MultiEdit per (session, file path) now emits a one-time advisory `systemMessage` JSON on stdout and ALLOWS the operation (exit 0). State-file logic, MOAI_FACT_FORCE=off opt-out, Read-as-investigation, self-loop prevention, and fail-open are all preserved unchanged. The settings.json registration is NOT touched (the predecessor already wired it).

### §A.2 Tier classification

**Tier S** — single-milestone, ≤ 3 files touched (hook script template + hook script local + hook-independence.md Mode G row wording), shell-only (no Go changes), additive behavior change. The 4-artifact set (spec.md + plan.md + acceptance.md + progress.md) is used per explicit user request for AC traceability and V3R6 era-classification readiness.

### §A.3 Files affected (file-by-file change list, Template-First ordering)

| # | Path | Template-First? | Action | LOC delta |
|---|------|------------------|--------|-----------|
| 1 | `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | YES (template) | MODIFY — 3 change-sites (§F below): header comment, fail-open comment, §11 emit-and-block → emit-advisory-exit-0 | ~20 LOC net change |
| 2 | (run `make build` to regenerate embedded assets) | n/a | BUILD | 0 |
| 3 | `.claude/hooks/moai/gateguard-fact-force.sh` | local mirror | MODIFY (mirror of #1, byte-identical) | ~20 LOC net change |
| 4 | `internal/template/templates/.claude/rules/moai/development/hook-independence.md` §3 Mode G row | YES (template) | MODIFY — rationale column: `"no gate"` → `"no advisory notice"` | rationale-column wording change |
| 5 | `.claude/rules/moai/development/hook-independence.md` §3 Mode G row | local mirror | MODIFY (mirror of #4, byte-identical) | rationale-column wording change |

**Total**: 2 template edits + 1 build + 2 local mirrors = 5 file operations. LOC delta ≈ 40 LOC net (mostly the §11 replacement). settings.json is NOT touched (predecessor already wired it).

### §A.4 PRESERVE list (DO NOT TOUCH)

| Path | Why preserved |
|------|---------------|
| `.claude/settings.json` | Predecessor already added the PreToolUse matcher group; this SPEC changes only hook behavior, not registration |
| `internal/template/templates/.claude/settings.json.tmpl` | Same — registration unchanged |
| `.claude/hooks/moai/handle-pre-tool.sh` | Existing PreToolUse Write\|Edit\|Bash group — additive-only |
| `.claude/hooks/moai/handle-harness-observe.sh` | Existing PostToolUse no-matcher group |
| `.claude/hooks/moai/handle-post-tool.sh` | Existing PostToolUse Write\|Edit group |
| `.claude/hooks/moai/status-transition-ownership.sh` | Existing PostToolUse advisory gate |
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | Existing Stop gate |
| `.claude/hooks/moai/team-ac-verify.sh` | Existing TaskCompleted gate (dormant) |
| `internal/hook/pre_tool.go` | Existing Go-side PreToolUse handler — out of scope |
| All other `internal/hook/*.go` files | Out of scope — no Go changes |

### §A.5 Anti-conflict constraints (audit-verified)

| Constraint | How respected |
|------------|---------------|
| settings.json unchanged | Plan does NOT touch settings.json — the predecessor's registration is preserved as-is |
| Template-First Rule | File #1 and #4 are template tree; #2 runs `make build`; #3 and #5 are local mirrors |
| jq prohibition (§C.5 NFR) | The §11 replacement uses `awk`-only JSON escape — no `jq` dependency |
| systemMessage validity (§C.6 NFR) | The awk escape produces `\n`-escaped single-line JSON valid for `jq -e '.systemMessage \| type'` |
| Subagent boundary | Shell script has no AskUserQuestion surface; AC-FA-005 verifies via grep |

## §B. Known Issues (filtered to relevant categories)

### B3. C-HRA-008 / Subagent Boundary Discipline

The hook script is shell-only and has no AskUserQuestion surface. AC-FA-005 verifies via grep that both `exit 2` and `AskUserQuestion` return zero matches.

### B5. CI 3-tier awareness

- spec-lint: this SPEC must pass `moai spec lint` (frontmatter schema, GEARS compliance, Out of Scope section presence).
- golangci-lint: NO Go changes → no NEW lint baseline.
- Test: NO Go test changes. Shell smoke tests (AC-FA-001 through AC-FA-015) are the verification.

### B6. spec-lint Heading Convention

`spec.md` §E uses `### Out of Scope — <topic>` (h3 sub-heading) per the `OutOfScopeRule` lint.

### B8. Working Tree Hygiene

State files under `.moai/state/fact-force/` are already gitignored (predecessor verified). No `.gitignore` changes needed.

### B10. Untouched Paths PRESERVE (Scope Discipline)

The implementing agent MUST NOT touch:
- Any file under `internal/hook/` (no Go changes).
- `.claude/settings.json` or its template (predecessor already wired registration).
- Any SPEC directory other than `SPEC-HOOK-FACTFORCE-ADVISORY-001/` and the predecessor's frontmatter `status:` field.

## §C. Pre-flight Check List

```bash
# 1. Verify branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Confirm the hook script exists (predecessor shipped it)
test -f .claude/hooks/moai/gateguard-fact-force.sh && echo "local hook: present"
test -f internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh && echo "template hook: present"

# 3. Confirm template == local (baseline parity)
diff .claude/hooks/moai/gateguard-fact-force.sh \
     internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty (identical baseline)

# 4. Confirm the current hook has exit 2 (the behavior we're removing)
grep -c 'exit 2' .claude/hooks/moai/gateguard-fact-force.sh
# Expected: 1 (the §11 block path)

# 5. Confirm make build is functional
make build 2>&1 | tail -5
# Expected: exit 0
```

## §D. Constraints (DO NOT VIOLATE)

| Constraint | Enforcement |
|------------|-------------|
| Shell-only — NO Go changes | `git diff --name-only HEAD~1 HEAD \| grep -E '\.go$'` MUST be empty |
| settings.json UNCHANGED | `git diff .claude/settings.json` MUST be empty; `git diff internal/template/templates/.claude/settings.json.tmpl` MUST be empty |
| jq-free implementation (§C.5) | `grep -c 'jq ' .claude/hooks/moai/gateguard-fact-force.sh` MUST be 0 |
| No `exit 2` anywhere (AC-FA-005) | `grep -c 'exit 2' .claude/hooks/moai/gateguard-fact-force.sh` MUST be 0 after the rewrite |
| No AskUserQuestion (AC-FA-005) | `grep -E 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh` MUST be empty |
| Template-First ordering | File #1 (template) MUST be edited BEFORE #3 (local); #4 (template) BEFORE #5 (local); #2 (`make build`) between |
| Template == local byte-identical | `diff` MUST return empty for both the hook script and hook-independence.md |

## §E. Self-Verification Deliverables

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-FA-001 | PASS | `echo '<mock first-edit payload>' \| bash .claude/hooks/moai/gateguard-fact-force.sh; echo $?` | advisory JSON on stdout + exit code **0** |
| AC-FA-002 | PASS | (after AC-FA-001) re-run with same session_id + path | exit 0, no stdout |
| AC-FA-003 | PASS | re-run with different session_id, same path | exit 0 + advisory JSON (proves per-session state) |
| AC-FA-004 | PASS | `MOAI_FACT_FORCE=off bash ...` | exit 0, no stdout; skip log appended |
| AC-FA-005 | PASS | `grep -n 'exit 2' <hook>` + `grep -E 'AskUserQuestion' <hook>` | both empty |
| AC-FA-006 | PASS | `time (echo '<payload>' \| bash ...)` | < 100ms |
| AC-FA-007 | PASS | `grep -A 6 '"PreToolUse"' .claude/settings.json` | existing groups unchanged |
| AC-FA-008 | PASS | `git diff .claude/settings.json` | empty (no settings.json change) |
| AC-FA-009 | SHOULD | `echo '<payload with agent_id>' \| bash ...` | same as AC-FA-001/002 (no subagent exemption) |
| AC-FA-010 | PASS | `echo '<payload with tool_name=Bash>' \| bash ...; echo $?` | exit 0, no state file |
| AC-FA-011 | PASS | `diff .claude/hooks/moai/gateguard-fact-force.sh internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | empty |
| AC-FA-012 | SHOULD | `stat -f '%p' .moai/state/fact-force/<hash>` | 0o600 |
| AC-FA-013 | PASS | Read payload then Write payload on same path | step1: exit 0, state created, no advisory; step2: exit 0, no advisory |
| AC-FA-014 | PASS | `echo "$stdout" \| jq -e '.systemMessage \| type'` | `"string"` |
| AC-FA-015 | PASS | `echo '<payload without session_id>' \| bash ...; echo $?` | exit 0, no advisory, no state file |

### E2-E7. (standard — cross-reference acceptance.md §B verification commands)

## §F. Single Milestone (Tier S scope)

### §F.1 M1 — Hook rewrite + Template-First mirror + Mode G row wording

**Single milestone**. The work is small enough (≤ 40 LOC net across 5 file operations) that splitting into multiple milestones would create artificial phase boundaries. The milestone executes in this strict order:

1. **Edit the template hook script** — `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` (3 change-sites, documented below).
2. **Run `make build`** — regenerate the embedded assets (`//go:embed all:templates`).
3. **Mirror to local** — copy the template hook to `.claude/hooks/moai/gateguard-fact-force.sh`.
4. **Edit the template hook-independence.md** — `internal/template/templates/.claude/rules/moai/development/hook-independence.md` §3 Mode G row: rationale column `"no gate"` → `"no advisory notice"`.
5. **Mirror hook-independence.md to local** — `.claude/rules/moai/development/hook-independence.md`.
6. **Smoke test locally** — run the AC-FA-001 through AC-FA-015 verifications from acceptance.md §B.
7. **Commit** — `feat(SPEC-HOOK-FACTFORCE-ADVISORY-001): M1 gateguard fact-force advisory rewrite (exit 2 → exit 0 + systemMessage)` with `🗿 MoAI` trailer.
8. **Push** — `git push origin main` (Tier S — Hybrid Trunk Route A).

**Exit criteria for M1**: all 15 ACs PASS (MUST) or PASS-WITH-DEBT (SHOULD only).

### §F.2 The 3 change-sites in the hook script

The implementing agent applies exactly 3 changes to `gateguard-fact-force.sh`. Each change-site is identified by its current text (the agent searches for this text and replaces it).

#### Change-site 1: Header comment (lines 1-15)

**Current text** (find):
```
# gateguard-fact-force.sh — PreToolUse hook (first-edit investigation gate)
#
# @MX:ANCHOR High fan-in gate — every edit passes through this hook
#
# Blocks the FIRST Edit/Write/MultiEdit on each file path per session and
# demands investigation (importers, data schemas, user instruction). Subsequent
# edits to the same path in the same session are allowed. A prior Read on the
# same path pre-populates the fact state, so the first post-Read Edit skips the
# gate (Read-as-investigation). Advisory opt-out via MOAI_FACT_FORCE=off.
# Shell-only, O(1), fail-open, self-terminates < 5s.
```

**Replacement text**:
```
# gateguard-fact-force.sh — PreToolUse hook (first-edit investigation advisory)
#
# @MX:ANCHOR High fan-in advisory — every edit passes through this hook
#
# Emits a one-time ADVISORY notice on the FIRST Edit/Write/MultiEdit on each file
# path per session, recommending investigation (importers, data schemas, user
# instruction). The edit is ALLOWED to proceed (exit 0) — this hook never blocks.
# Subsequent edits to the same path in the same session produce no advisory. A
# prior Read on the same path pre-populates the fact state, so the first post-Read
# Edit skips the advisory (Read-as-investigation). Advisory opt-out via
# MOAI_FACT_FORCE=off. Shell-only, O(1), fail-open, self-terminates < 5s, jq-free.
```

**Rationale**: updates the description from "gate" (blocking) to "advisory" (non-blocking). Adds `jq-free` to signal the §C.5 NFR compliance.

#### Change-site 2: Fail-open comment (line 17-18)

**Current text** (find):
```
# Fail-open wrapper: any unexpected error → exit 0 (allow). The ONLY exit 2
# path is the first-edit block after the state-file write succeeds.
```

**Replacement text**:
```
# Fail-open wrapper: any unexpected error → exit 0 (allow). This hook NEVER
# exits 2 — it is advisory-only (exit 0 on every path).
```

**Rationale**: removes the reference to "the ONLY exit 2 path" since there is no longer any exit 2 path.

#### Change-site 3: §11 emit-and-block → emit-advisory-exit-0 (lines 110-127)

**Current text** (find — the entire §11 block from the comment to `exit 2`):
```bash
# --- 11. Emit guidance + block (REQ-FF-001) ---
# GUIDANCE → stderr: Claude Code exit-2 semantics require the block reason on
# stderr (stdout-only exit 2 surfaces as "No stderr output" error). See CC 2.1.202.
cat <<GUIDANCE >&2
FACT-FORCE GATE: first edit on $file_path blocked.

Before proceeding, investigate:
  1. IMPORTERS — who imports / depends on this file?
       grep -rn "<file-basename>" --include='*.go' --include='*.ts' --include='*.py' .  (adapt to language)
  2. DATA SCHEMAS — what data structures / contracts / API types does this file touch?
       Read the struct / interface / type definitions and their consumers.
  3. USER INSTRUCTION — what user instruction justifies this edit?
       Re-read the SPEC acceptance criteria or the explicit user request.

This is a one-time gate per (session, file path). Your NEXT edit to this path
will be allowed. To disable for the session: MOAI_FACT_FORCE=off
GUIDANCE
exit 2
```

**Replacement text** (awk-based JSON escape, jq-free per §C.5 NFR; the advisory text is preserved verbatim — only the delivery channel changes from stderr+exit-2 to stdout+exit-0):
```bash
# --- 11. Emit advisory systemMessage + allow (REQ-FA-001) ---
# stdout JSON systemMessage: Claude Code renders this as informational context
# (NOT a red error box). exit 0 = allow. This hook NEVER blocks.
guidance=$(cat <<GUIDANCE
First-edit advisory on $file_path.

Before proceeding, investigate:
  1. IMPORTERS — who imports / depends on this file?
       grep -rn "<file-basename>" --include='*.go' --include='*.ts' --include='*.py' .  (adapt to language)
  2. DATA SCHEMAS — what data structures / contracts / API types does this file touch?
       Read the struct / interface / type definitions and their consumers.
  3. USER INSTRUCTION — what user instruction justifies this edit?
       Re-read the SPEC acceptance criteria or the explicit user request.

This is a one-time advisory per (session, file path). Your NEXT edit to this path
will not produce this notice. To disable for the session: MOAI_FACT_FORCE=off
GUIDANCE
)

# JSON-escape via awk (jq-free, per §C.5 NFR): backslash to double-backslash,
# double-quote to backslash-quote, inter-line newline to literal backslash-n.
escaped=$(printf '%s' "$guidance" | awk '
BEGIN { sep = "" }
{ gsub(/\\/, "\\\\"); gsub(/"/, "\\\""); printf "%s%s", sep, $0; sep = "\\n" }
')
printf '{"systemMessage":"%s"}\n' "$escaped"
exit 0
```

**Rationale**: this is the core behavior change. The guidance text is preserved verbatim (same 3 investigation points: IMPORTERS, DATA SCHEMAS, USER INSTRUCTION). The delivery channel changes from `cat >&2` + `exit 2` to `printf` on stdout + `exit 0`. The awk escape produces valid single-line JSON with `\n`-escaped newlines (no raw newlines inside the JSON string). The `exit 2` is removed entirely — this is the ONLY code path that previously exited 2.

**Key invariant**: after this change, `grep -c 'exit 2' gateguard-fact-force.sh` MUST return 0 (AC-FA-005).

### §F.3 hook-independence.md Mode G row wording change

**Current text** (find in the rationale column of the Mode G row):
```
the failure mode degrades to "no gate" rather than "session halt"
```

**Replacement text**:
```
the failure mode degrades to "no advisory notice" rather than "session halt"
```

**Rationale**: the hook no longer gates (blocks); it advises. The classification (`acceptable-by-design`) is unchanged. This is a 1-word substitution in the template and its local mirror.

## §G. Anti-Patterns

| Anti-pattern | Why prohibited |
|--------------|----------------|
| Adding `jq` as a runtime dependency | §C.5 NFR prohibits; breaks hook-independence.md Mode G self-contained classification |
| Emitting raw newlines inside the JSON string | §C.6 NFR prohibits; would produce invalid JSON |
| Keeping any `exit 2` path | REQ-FA-001 + REQ-FA-005 prohibit; the hook is advisory-only |
| Modifying settings.json | Predecessor already wired registration; this SPEC is hook-behavior-only |
| Modifying the existing PreToolUse Write\|Edit\|Bash group | REQ-FA-007 requires preservation |
| Changing the state-file keying logic | REQ-FA-003 is unchanged from predecessor |
| Removing the MOAI_FACT_FORCE=off opt-out | REQ-FA-004 is unchanged |
| Removing the Read-as-investigation logic | Preserved per §A.2; the `case` block for Read is NOT in the 3 change-sites |
| Invoking AskUserQuestion from the implementing agent | Subagent boundary; return blocker report instead |
| Changing hook-independence.md beyond the Mode G row wording | §E Out of Scope prohibits |

## §H. Cross-References

- `spec.md` — canonical requirements (REQ-FA-001 through REQ-FA-012) and Exclusions
- `acceptance.md` — canonical AC matrix (AC-FA-001 through AC-FA-015) with Given-When-Then structure
- `progress.md` — §E.1 plan-phase audit-ready signal scaffold
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec
- `.claude/rules/moai/development/hook-independence.md` § 3 — Mode G row (M1 step 4-5 updates the wording)
- Predecessor SPEC: `.moai/specs/SPEC-HOOK-PREEDIT-INVESTIGATE-001/spec.md` — the completed SPEC being superseded
- CLAUDE.local.md §2 Template-First Rule, §7 hook timeout policy

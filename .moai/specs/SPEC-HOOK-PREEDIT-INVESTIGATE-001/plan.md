# SPEC-HOOK-PREEDIT-INVESTIGATE-001 — Implementation Plan

> Companion to `spec.md` (canonical requirements) and `acceptance.md` (canonical ACs). This document owns the implementation approach: file-by-file change list, milestone structure, known-issues injection, and self-verification deliverables.

## §A. Context

### §A.1 Scope summary

A single new PreToolUse shell hook (`gateguard-fact-force.sh`) that blocks the FIRST Edit/Write/MultiEdit on each file path within a session, emits a guidance message demanding investigation (importers / data schemas / user instruction), and allows subsequent edits to the same path. State is session-scoped under `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/`, keyed by a hash of `session_id + absolute_file_path`. Advisory opt-out via `MOAI_FACT_FORCE=off` env var.

### §A.2 Tier classification

**Tier S** — single-milestone, ≤ 5 files, shell-only (no Go changes), additive-only (existing hooks untouched), fail-open default. The 4-artifact set (spec.md + plan.md + acceptance.md + progress.md) is the explicit user request; convention is 2 artifacts for Tier S, but the user's request for explicit acceptance.md + progress.md is honored for AC traceability and V3R6 era-classification readiness.

### §A.3 Files affected (file-by-file change list, Template-First ordering)

| # | Path | Template-First? | Action | LOC delta |
|---|------|------------------|--------|-----------|
| 1 | `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | YES (template) | NEW | ~80 LOC shell script |
| 2 | `internal/template/templates/.claude/settings.json.tmpl` | YES (template) | MODIFY — add a NEW PreToolUse matcher group `Edit\|Write\|MultiEdit` → `gateguard-fact-force.sh` (5s timeout, type: command), placed AFTER the existing `Write\|Edit\|Bash` group and BEFORE the closing `]` of the PreToolUse array | ~7 LOC addition |
| 3 | (run `make build` to regenerate embedded assets) | n/a | BUILD | 0 |
| 4 | `.claude/hooks/moai/gateguard-fact-force.sh` | local mirror | NEW (copy of #1) | ~80 LOC shell script |
| 5 | `.claude/settings.json` | local mirror | MODIFY (mirror of #2) | ~7 LOC addition |
| 6 | `.claude/rules/moai/development/hook-independence.md` § 3 | local doc | MODIFY — append a new row to the shared-failure-mode catalogue recording the new state-file dependency introduced by this hook | ~5 LOC addition |

**Total**: 4 file changes + 1 doc update + 1 build step. LOC delta ≈ 175 LOC (mostly the shell script).

### §A.4 PRESERVE list (DO NOT TOUCH)

| Path | Why preserved |
|------|---------------|
| `.claude/hooks/moai/handle-pre-tool.sh` | Existing PreToolUse Write\|Edit\|Bash group — additive-only, REQ-FF-007 |
| `.claude/hooks/moai/handle-harness-observe.sh` | Existing PostToolUse no-matcher group — REQ-FF-008 |
| `.claude/hooks/moai/handle-post-tool.sh` | Existing PostToolUse Write\|Edit group |
| `.claude/hooks/moai/status-transition-ownership.sh` | Existing PostToolUse advisory gate, separate concern |
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | Existing Stop gate, separate concern |
| `.claude/hooks/moai/team-ac-verify.sh` | Existing TaskCompleted gate (dormant), separate concern |
| `internal/hook/pre_tool.go` | Existing Go-side PreToolUse handler — out of scope (shell-only implementation) |
| All other `internal/hook/*.go` files | Out of scope — no Go changes in this SPEC |

### §A.5 Anti-conflict constraints (audit-verified)

The following constraints are audit-confirmed and MUST be respected by the implementation. Cite these in any plan-auditor review.

| Constraint | Audit source | How respected |
|------------|--------------|---------------|
| PostToolUse `"*"` is occupied | `.claude/settings.json:67-76` | Plan does NOT add any PostToolUse entry. |
| PreToolUse `Write\|Edit\|Bash` group MUST be preserved | `.claude/settings.json:38-49` | Plan adds a SEPARATE matcher group, not modification. |
| `status-transition-ownership.sh` is a separate concern | `.claude/settings.json:59-65` | Plan does NOT touch PostToolUse at all. |
| State MUST be session-scoped | `session-handoff.md` multi-session coordination policy | State file keyed by `session_id` hash + absolute path hash. |
| Subagent boundary (no AskUserQuestion) | C-HRA-008, `agent-common-protocol.md` § User Interaction Boundary | Shell script has no AskUserQuestion surface; grep acceptance criterion: `grep -E 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh` MUST return empty. |
| 5s hook timeout | `CLAUDE.local.md` §7, `hooks-system.md` § Timeout Configuration | Implementation is shell-only `O(1)` operations; no network, no LSP, no subprocess beyond `[ -f ... ]` + `cat > ...`. |
| Template-First Rule | `CLAUDE.local.md` §2 [HARD] | File #1 and #2 in §A.3 are template tree; #3 runs `make build`; #4 and #5 are local mirror. |

## §B. Known Issues (filtered to relevant categories)

This section applies the canonical 12-category known-issues rubric (B1-B12) filtered to the categories relevant to a shell-only hook addition. Categories B1 (cross-platform build tags — no Go), B3 (C-HRA-008 — explicitly relevant, included), B5 (CI 3-tier — included), B6 (spec-lint headings — included), B8 (working-tree hygiene — included), B10 (untouched paths preserve — explicitly relevant, included), B11 (AskUserQuestion prohibition — shell script has no AskUserQuestion surface, included for completeness). B2 / B4 / B7 / B9 / B12 are filtered out (no cross-SPEC conflict, no frontmatter issue, no observer.go CWD issue, no git-commit concern for plan-phase, no manager-docs CHANGELOG concern in plan-phase).

### B3. C-HRA-008 / Subagent Boundary Discipline

The hook script is shell-only and has no AskUserQuestion surface. The acceptance criterion is the grep check:

```bash
grep -E 'AskUserQuestion|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty output
```

This is a binary constraint (zero matches). A CI guard test is NOT strictly necessary for Tier S (the implementation is shell, not Go), but the plan-auditor SHOULD verify the grep returns empty as part of AC-FF-005.

### B5. CI 3-tier awareness

- spec-lint: this SPEC must pass `moai spec lint` (frontmatter schema, EARS/GEARS compliance, Out of Scope section presence, ownership-transition rule).
- golangci-lint: NO Go changes in this SPEC → no NEW lint baseline. The shell script is not covered by golangci-lint.
- Test: NO Go test changes. Shell tests are OPTIONAL for Tier S (shell-only); a manual smoke test (send a mock PreToolUse payload via stdin, verify exit code 2 + guidance message) is the minimum verification.

### B6. spec-lint Heading Convention

`spec.md` §E uses `### Out of Scope — <topic>` (h3 sub-heading) per the `OutOfScopeRule` lint. The plan-auditor SHOULD verify all 7 Out-of-Scope entries use this format.

### B8. Working Tree Hygiene

The hook creates state files under `${CLAUDE_PROJECT_DIR:-$PWD}/.moai/state/fact-force/`. This directory MUST be added to `.gitignore` if not already covered. Verification:

```bash
grep -E '(\.moai/state/|fact-force)' .gitignore
# Expected: at least one match covering .moai/state/
```

If `.moai/state/` is NOT already gitignored, the plan MUST add it. (Most moai-adk projects already gitignore `.moai/state/` — verify in pre-flight.)

### B10. Untouched Paths PRESERVE (Scope Discipline)

The implementing agent MUST NOT touch:
- Any file under `internal/hook/` (no Go changes in this SPEC).
- Any file under `.claude/hooks/moai/` OTHER than the new `gateguard-fact-force.sh`.
- Any file under `internal/template/templates/.claude/hooks/moai/` OTHER than the new `gateguard-fact-force.sh`.
- Any SPEC directory other than `SPEC-HOOK-PREEDIT-INVESTIGATE-001/`.
- Any runtime-managed file (`.moai/harness/*`, `.moai/cache/*`, `.moai/logs/*`).

### B11. AskUserQuestion Prohibition (Subagent Boundary)

The implementing agent (manager-develop or per-spawn `Agent(general-purpose)`) MUST NOT invoke AskUserQuestion during run-phase. If the implementing agent discovers an ambiguity NOT resolved by this plan, it MUST return a structured blocker report to the orchestrator, who runs the AskUserQuestion round and re-delegates.

## §C. Pre-flight Check List

Before any code change, the implementing agent runs:

```bash
# 1. Verify branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Confirm the PreToolUse broad slot is still empty (audit was at plan time; re-verify at run time)
grep -A 8 '"PreToolUse"' .claude/settings.json
# Expected: ONE matcher group (Write|Edit|Bash → handle-pre-tool.sh). If TWO groups already exist, STOP and surface to orchestrator.

# 3. Confirm PostToolUse "*" is still occupied by handle-harness-observe.sh (do not touch)
grep -A 4 '"PostToolUse"' .claude/settings.json | grep -E 'handle-harness-observe|harness-observe'
# Expected: at least one match. If absent, the audit baseline has drifted — STOP.

# 4. Verify .moai/state/ is gitignored (B8)
grep -E '\.moai/state/|fact-force' .gitignore
# If empty: add .moai/state/ to .gitignore as part of this SPEC.

# 5. Verify the template tree mirrors the local hooks directory
ls internal/template/templates/.claude/hooks/moai/ | wc -l
ls .claude/hooks/moai/ | wc -l
# Note: counts may differ if local-only hooks exist; the SPEC adds ONE new file to BOTH directories.

# 6. Verify shell portability utilities are available
command -v shasum || command -v sha1sum
# Expected: at least ONE of these is available (OR semantics — macOS ships shasum only; Linux typically ships sha1sum only). The implementation detects which.

# 7. Verify make build is functional
make build 2>&1 | tail -5
# Expected: exit 0, no errors.
```

## §D. Constraints (DO NOT VIOLATE)

| Constraint | Source | Enforcement |
|------------|--------|-------------|
| Shell-only implementation — NO Go changes | Tier S scope, §A.4 PRESERVE list | Cite: `git diff --name-only HEAD~1 HEAD | grep -E '\.go$'` MUST be empty after the run-phase commit |
| Additive-only on settings.json — separate matcher group | REQ-FF-007 | Cite: the new group is a NEW array entry, not a modification of the existing `Write\|Edit\|Bash` group |
| No PostToolUse registration of ANY kind | REQ-FF-008 | Cite: `git diff .claude/settings.json | grep -E '^\+.*PostToolUse'` MUST be empty |
| No AskUserQuestion in the hook script | REQ-FF-005, C-HRA-008 | Cite: `grep -E 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh` MUST be empty |
| Template-First ordering | CLAUDE.local.md §2 [HARD] | Cite: file #1 (template) and #2 (template) MUST be edited BEFORE #4 (local) and #5 (local) |
| `make build` between template edit and local mirror | CLAUDE.local.md §2 | Cite: step #3 in §A.3 file table |
| 5s hook timeout, fail-open on any error | REQ-FF-006, §C.1, §C.2 | Cite: hook script ends with `exit 0` on every unexpected-error path; the only `exit 2` is the first-edit block |
| State file permissions 0o600 | REQ-FF-012 | Cite: `(umask 077; cat > "$state_file" <<EOF...)` pattern in the script |

## §E. Self-Verification Deliverables

When the implementing agent reports completion, it MUST include the following self-verification matrix (per the canonical E1-E7 deliverables structure, adapted for shell-only scope):

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Expected Output |
|----|--------|---------------------|-----------------|
| AC-FF-001 | PASS | `echo '<mock first-edit payload>' \| bash .claude/hooks/moai/gateguard-fact-force.sh; echo $?` | guidance message on stdout + exit code 2 |
| AC-FF-002 | PASS | (after AC-FF-001) re-run the same command | exit code 0, no guidance message |
| AC-FF-003 | PASS | `ls .moai/state/fact-force/` | one state file per (session_id, file_path) pair |
| AC-FF-004 | PASS | `MOAI_FACT_FORCE=off bash ... <<< '<payload>'` | exit code 0, no guidance; `tail -1 .moai/logs/fact-force-skip.log` shows the bypass record |
| AC-FF-005 | PASS | `grep -E 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh` | empty |
| AC-FF-006 | PASS | `time (echo '<payload>' \| bash .claude/hooks/moai/gateguard-fact-force.sh)` | < 100ms wall-clock |
| AC-FF-007 | PASS | `grep -A 6 '"PreToolUse"' .claude/settings.json` | TWO matcher groups: original (Write\|Edit\|Bash → handle-pre-tool.sh) + new (Edit\|Write\|MultiEdit → gateguard-fact-force.sh); original UNCHANGED |
| AC-FF-008 | PASS | `git diff .claude/settings.json \| grep -E '^\+.*"PostToolUse"'` | empty |
| AC-FF-009 | SHOULD | `echo '<payload with agent_id>' \| bash .claude/hooks/moai/gateguard-fact-force.sh` | same as AC-FF-001 (no subagent exemption) |
| AC-FF-010 | PASS | `echo '<payload with tool_name=Read>' \| bash .claude/hooks/moai/gateguard-fact-force.sh; echo $?` | exit 0, no state file written |
| AC-FF-011 | PASS | `diff .claude/hooks/moai/gateguard-fact-force.sh internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | empty (the two are identical) |
| AC-FF-012 | SHOULD | `stat -f '%p' .moai/state/fact-force/<hash>` (macOS) or `stat -c '%a' ...` (Linux) | 0o600 permissions |

### E2. Cross-Platform Shell Compatibility

```bash
# macOS (bash 3.2)
echo '<payload>' | bash .claude/hooks/moai/gateguard-fact-force.sh

# Linux (bash 4+)
echo '<payload>' | bash .claude/hooks/moai/gateguard-fact-force.sh
```

Both MUST produce the same exit code and stdout. Any divergence is a defect.

### E3. Lint Status

```bash
# spec-lint
moai spec lint .moai/specs/SPEC-HOOK-PREEDIT-INVESTIGATE-001/

# golangci-lint (NO Go changes expected)
golangci-lint run --timeout=2m 2>&1 | tail -5
# Expected: no NEW findings (baseline preserved)
```

### E4. Subagent Boundary Grep (C-HRA-008 family)

```bash
grep -rn 'AskUserQuestion|mcp__askuser' .claude/hooks/moai/gateguard-fact-force.sh internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh
# Expected: empty
```

### E5. Build Verification

```bash
make build 2>&1 | tail -3
# Expected: exit 0
```

### E6. Branch HEAD + Push state

Implementing agent reports:
- List of new commit SHAs (Tier S — Hybrid Trunk main-direct per Route A; commit lands on main).
- Result of `git push origin main`.

### E7. Blocker Report (if any)

If the implementing agent discovers an ambiguity NOT resolved by this plan (e.g., the audit baseline has drifted because a parallel session added a different PreToolUse matcher group), it MUST return a structured blocker report. The orchestrator runs AskUserQuestion and re-delegates.

## §F. Single Milestone (Tier S scope)

### §F.1 M1 — Hook implementation + Template-First mirror + Doc update

**Single milestone**. The work is small enough (≤ 175 LOC across 4 files + 1 doc) that splitting into multiple milestones would create artificial phase boundaries. The milestone executes in this strict order:

1. **Author the template hook script** — `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` (~80 LOC shell, fail-open, 5s self-termination, 0o600 state file, shasum/sha1sum portability detection).
2. **Modify the template settings.json** — `internal/template/templates/.claude/settings.json.tmpl` (add a NEW PreToolUse matcher group scoped to `Edit|Write|MultiEdit`, 5s timeout, type: command, placed AFTER the existing `Write|Edit|Bash` group).
3. **Run `make build`** — regenerate the embedded assets (`//go:embed all:templates`).
4. **Mirror to local** — copy the template hook to `.claude/hooks/moai/gateguard-fact-force.sh`; mirror the settings.json edit to `.claude/settings.json`.
5. **Smoke test locally** — run the AC-FF-001 through AC-FF-012 verifications from §E.1.
6. **Update `hook-independence.md` § 3** — append a new row to the shared-failure-mode catalogue recording the new state-file dependency.
7. **Commit (single commit, Conventional Commits)** — `feat(SPEC-HOOK-PREEDIT-INVESTIGATE-001): M1 gateguard fact-force PreToolUse hook + template mirror + hook-independence catalogue update` with `🗿 MoAI` trailer.
8. **Push** — `git push origin main` (Tier S — Hybrid Trunk Route A).

**Exit criteria for M1**: all 12 ACs in acceptance.md PASS (MUST) or PASS-WITH-DEBT (SHOULD only, with explicit debt annotation in the commit body).

## §G. Anti-Patterns

| Anti-pattern | Why prohibited |
|--------------|----------------|
| Implementing in Go (`internal/hook/fact_force.go`) | Out of scope per §E in spec.md; shell-only is sufficient for the Tier S scope |
| Combining this SPEC with candidate B (Search-First) | User explicitly scoped to F only — scope creep |
| Modifying the existing `Write\|Edit\|Bash` matcher group | REQ-FF-007 requires additive-only |
| Registering a PostToolUse `"*"` matcher | REQ-FF-008 prohibits; slot occupied by `handle-harness-observe.sh` |
| Implementing confidence scoring or learning | Out of scope per §E in spec.md; deterministic gate only |
| Skipping the Template-First ordering | CLAUDE.local.md §2 [HARD] |
| Invoking AskUserQuestion from the implementing agent | Subagent boundary; return blocker report instead |
| Forgetting the `make build` step between template edit and local mirror | Will cause the embedded assets to drift from the template tree |
| Implementing state cleanup / GC | Out of scope; deferred to follow-up SPEC |

## §H. Cross-References

- `spec.md` — canonical requirements (REQ-FF-001 through REQ-FF-012) and Exclusions
- `acceptance.md` — canonical AC matrix (AC-FF-001 through AC-FF-012) with Given-When-Then structure
- `progress.md` — §E.1 plan-phase audit-ready signal scaffold
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S minimal delegation form permitted (~500-800 tokens)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S = 2 artifacts (this SPEC uses 4 per explicit user request)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec (this plan-phase commit); `draft → in-progress` owned by manager-develop (the M1 commit in §F.1)
- `.claude/rules/moai/development/hook-independence.md` § 3 — shared-failure-mode catalogue (M1 step 6 updates this)
- CLAUDE.local.md §2 Template-First Rule, §7 hook timeout policy

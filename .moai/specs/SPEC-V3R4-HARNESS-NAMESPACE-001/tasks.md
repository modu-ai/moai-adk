# SPEC-V3R4-HARNESS-NAMESPACE-001 — Run-Phase Tasks

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.0   | 2026-05-16 | manager-spec | plan-auditor REVISE 0.69 round-1 fix. §5 Step 2 Wave 1 shell snippet — T-Wave1-001 SPEC ID regex broadened (D-002); T-Wave1-004 lint switched to no-args full-repo scan (D-003). §5 Step 3 Wave 2 — T-Wave2-001 PR HEAD check reframed for EXPECTED 2-commit divergence with prefix matching + EC-001 escalation (D-004); T-Wave2-003 verb-surface regex corrected for numbered headers (D-001). T-Wave2-002 attribution comment + close now conditional on EC-001 option selected. |
| 0.1.0   | 2026-05-16 | manager-spec | Initial draft. Run-phase task breakdown: 7 governance verification tasks (Wave 1) + 5 PR closeout tasks (Wave 2). ZERO code tasks. All tasks produce markdown or audit artifacts only. Worktree NOT used (plan-in-main mode → run-phase uses `feat/SPEC-V3R4-HARNESS-NAMESPACE-001` branch). |

---

## 1. Task Overview

Pure governance SPEC — all tasks listed below produce ONLY:

- Markdown updates (`CHANGELOG.md`)
- Audit report artifacts (`.moai/reports/governance/*.md`)
- GitHub PR state changes (close, delete-branch, comment)
- Memory entries (auto-memory taxonomy, type=project)

NO tasks modify source code (`.go`, `.ts`, `.js`, etc.). NO tasks modify skill bodies, agent bodies, or rule files. NO tasks create runtime data files.

### 1.1 Branch Strategy

- Plan-phase branch: `feat/SPEC-V3R4-HARNESS-NAMESPACE-001-plan` (current branch as of plan-phase Step 1).
- Run-phase branch: `feat/SPEC-V3R4-HARNESS-NAMESPACE-001` (created from plan-PR merged main).
- Sync-phase branch: `sync/SPEC-V3R4-HARNESS-NAMESPACE-001` (created from run-PR merged main).
- Worktree: NOT used (plan-in-main mode per CLAUDE.local.md §18.12 BODP).

### 1.2 Total Task Count

| Wave | Tasks | Priority | Estimated artifact output |
|------|-------|----------|---------------------------|
| Wave 1 — Governance Verification | 7 (T-Wave1-001 through T-Wave1-007) | P1 | 1 CHANGELOG entry + 1 audit report |
| Wave 2 — PR #908 Closeout | 5 (T-Wave2-001 through T-Wave2-005) | P0 | 1 PR comment + 1 branch deletion + 1 audit report + 1 memory entry |
| **Total** | **12 tasks** | mixed | ~3-4 markdown files modified, 1 PR closed, 1 branch deleted |

### 1.3 Task Dependency Graph (intra-SPEC)

```
T-Wave1-001 → T-Wave1-002 → T-Wave1-003 → T-Wave1-004 → T-Wave1-005 → T-Wave1-006 → T-Wave1-007
                                                                                          ↓
                                                                              [Wave 1 acceptance gate]
                                                                                          ↓
                                                              T-Wave2-001 → T-Wave2-002 → T-Wave2-003
                                                                                          ↓
                                                                                      T-Wave2-004
                                                                                          ↓
                                                                                      T-Wave2-005
                                                                                          ↓
                                                                              [Wave 2 acceptance gate]
                                                                                          ↓
                                                                                    [PR ready]
```

All tasks within a Wave are sequential. Inter-Wave: Wave 1 MUST complete before Wave 2 begins.

---

## 2. Wave 1 — Governance Verification Tasks

### T-Wave1-001 — SPEC ID format compliance verification

- **Type**: Governance verification (read-only shell)
- **Priority**: P1
- **Inputs**: `.moai/specs/SPEC-V3R4-HARNESS-*/spec.md`, `.moai/specs/SPEC-V3R3-HARNESS-*/spec.md`
- **Action**: Execute SPEC ID regex check per AC-HRN-NS-001 verification block.
- **Outputs**: stdout report (zero VIOLATION lines expected).
- **Acceptance**: AC-HRN-NS-001 PASS.
- **Owning agent**: manager-spec or any read-only agent.

### T-Wave1-002 — Lifecycle independence dependency-graph inspection

- **Type**: Governance verification (read-only shell + manual review)
- **Priority**: P1
- **Inputs**: `.moai/specs/SPEC-V3R4-HARNESS-*/spec.md` frontmatter `dependencies:` blocks.
- **Action**: Extract `dependencies:` from each spec.md, verify acyclic graph + foundation-only edges per AC-HRN-NS-002.
- **Outputs**: Dependency graph rendered to audit report (T-Wave1-007).
- **Acceptance**: AC-HRN-NS-002 PASS.
- **Owning agent**: manager-spec.

### T-Wave1-003 — `.moai/harness/*` hierarchy probe

- **Type**: Governance verification (read-only `find`)
- **Priority**: P1
- **Inputs**: `.moai/harness/` (if present).
- **Action**: Non-destructive `find` probe to enumerate sub-paths; compare against canonical set in design.md §2.1.
- **Outputs**: Hierarchy snapshot in audit report.
- **Acceptance**: AC-HRN-NS-004 PASS (canonical subset OR NOT_PRESENT).
- **Owning agent**: manager-spec.

### T-Wave1-004 — Supersede consistency `moai spec lint --strict`

- **Type**: Governance verification (lint invocation)
- **Priority**: P1
- **Inputs**: Full SPEC corpus (`.moai/specs/`) — lint operates on entire repo via no-args invocation.
- **Action**: Execute `moai spec lint --strict` (no positional arguments) and assert exit code 0 + `✓ No findings — all SPEC documents are valid`.
- **Outputs**: Lint output piped to `/tmp/harness-lint.txt`, copied to audit report.
- **Acceptance**: AC-HRN-NS-005 PASS.
- **Owning agent**: manager-spec.

### T-Wave1-005 — CLI deprecation grace re-assertion

- **Type**: Governance verification (read-only shell `test` + `grep`)
- **Priority**: P1
- **Inputs**: `internal/cli/harness.go`, `internal/cli/root.go`.
- **Action**: Verify deprecation marker file exists AND no harness subcommand registered in cobra tree.
- **Outputs**: PASS/FAIL line per check in audit report.
- **Acceptance**: AC-HRN-NS-006 PASS (MARKER_PRESENT + NO_REGISTRATION).
- **Owning agent**: manager-spec.

### T-Wave1-006 — CHANGELOG v3.0.0-rc1 entry append

- **Type**: Markdown edit
- **Priority**: P1
- **Inputs**: `CHANGELOG.md` (current main HEAD).
- **Action**: Append a `### Governance` block under the v3.0.0-rc1 section per plan.md §2.2 T-Wave1-006 content sketch. Include references to SPEC-V3R4-HARNESS-NAMESPACE-001, PR #908 closeout decision, and REQ-HRN-NS-010 retention-policy reference.
- **Outputs**: Modified `CHANGELOG.md`.
- **Acceptance**: AC-HRN-NS-008 PASS (retention reference present via `grep`).
- **Owning agent**: manager-spec or manager-docs.

### T-Wave1-007 — Wave 1 audit artifact persistence

- **Type**: Markdown create
- **Priority**: P1
- **Inputs**: All outputs from T-Wave1-001 through T-Wave1-006.
- **Action**: Create `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-<YYYY-MM-DD>.md` containing PASS/FAIL summary per task + raw command outputs.
- **Outputs**: New audit report file.
- **Acceptance**: File exists, contains all 6 task results, each with PASS or documented exception.
- **Owning agent**: manager-spec.

### Wave 1 Acceptance Gate

All 7 tasks above MUST complete with PASS before Wave 2 begins. If any task surfaces a BLOCKER (e.g., AC-HRN-NS-005 lint regression caused by external SPEC merge), Wave 2 is deferred and the orchestrator escalates via blocker report.

---

## 3. Wave 2 — PR #908 Closeout Tasks

### T-Wave2-001 — PR #908 HEAD probe (EC-001 escalation EXPECTED)

- **Type**: GitHub API read (gh CLI)
- **Priority**: P0
- **Inputs**: PR #908 metadata from `gh pr view 908 --json headRefOid,headRefName,state`.
- **Action**: Probe HEAD commit SHA with **prefix matching** (`[[ "$head" == 452aa638f* ]]`, NOT string equality). Plan-phase verified HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191` → EXPECTED `EC_001_PATH=true`. On divergence, orchestrator (NOT subagent) invokes `AskUserQuestion` with 4 EC-001 options.
- **Outputs**: HEAD SHA + `EC_001_PATH` flag recorded in audit report; option selection recorded after AskUserQuestion completes.
- **Acceptance**: Probe completes (PASS). Subsequent T-Wave2-002 conditional on user-selected option.
- **Owning agent**: orchestrator (probe) + `manager-git` (option execution after orchestrator's AskUserQuestion round).

### T-Wave2-002 — PR #908 disposition execution (conditional on user-selected EC-001 option)

- **Type**: GitHub API write (gh CLI) — conditional on option chosen
- **Priority**: P0
- **Inputs**: PR #908 (HEAD probe + `EC_001_PATH` from T-Wave2-001) + user-selected option from orchestrator's AskUserQuestion round.
- **Action** (option-dependent):
  - **Option 1 (Recommended) — rollback-tip-then-close**:
    1. `git fetch origin feat/cmd-harness-slash-wrapper`
    2. `git checkout feat/cmd-harness-slash-wrapper`
    3. `git reset --hard 452aa638f`
    4. `git push --force-with-lease origin feat/cmd-harness-slash-wrapper`
    5. `gh pr comment 908 --body "<attribution-text>"`
    6. `gh pr close 908 --delete-branch`
  - **Option 2 — cherry-pick-to-new-PR-then-close**:
    1. `git checkout main && git pull origin main`
    2. `git checkout -b docs/harness-audit-2026-05-13`
    3. `git cherry-pick a41d6d139`
    4. `git push -u origin docs/harness-audit-2026-05-13`
    5. `gh pr create --base main --title "docs(audit): harness 서브시스템 end-to-end 진단 보고서 (PR #908 cherry-pick)"`
    6. `gh pr comment 908 --body "<attribution-text + cherry-pick PR reference>"`
    7. `gh pr close 908 --delete-branch`
  - **Option 3 — abandon-new-commits**:
    1. `gh pr comment 908 --body "<attribution-text + audit doc discarded notice>"`
    2. `gh pr close 908 --delete-branch`
  - **Option 4 — leave-open**:
    1. No git or gh actions.
    2. Record deferral rationale in `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave2-<DATE>.md` (T-Wave2-005).
- **Outputs**: PR state transition (options 1-3) or unchanged (option 4); branch deletion (options 1-3); audit log entry.
- **Acceptance**: AC-HRN-NS-007 PASS:
  - Options 1-3: `gh pr view 908 --json state -q .state` = `CLOSED` AND `git ls-remote --heads origin feat/cmd-harness-slash-wrapper | wc -l` = `0`
  - Option 4: deferral rationale in T-Wave2-005 audit report.
- **Owning agent**: `manager-git` (executes option). Orchestrator owns AskUserQuestion round; user owns option selection.

### T-Wave2-003 — Verb-surface re-assertion verification

- **Type**: Governance verification (read-only `grep` + `diff`)
- **Priority**: P0
- **Inputs**: `.claude/skills/moai/workflows/harness.md` (numbered headers `### 2.1 status` ... `### 2.4 disable`).
- **Action**: Extract verb headers using **corrected regex** `^### [0-9.]+ +(status|apply|rollback|disable)\b` (admits numeric prefixes), `diff` against expected 4-verb set per AC-HRN-NS-003.
- **Outputs**: Diff result (empty = PASS).
- **Acceptance**: AC-HRN-NS-003 PASS (exactly 4 verbs, no drift). Verified at plan-phase against lines 132/166/196/208 — all 4 matched.
- **Owning agent**: manager-spec.

### T-Wave2-004 — Memory entry persistence

- **Type**: Memory file write (auto-memory taxonomy compliance)
- **Priority**: P1
- **Inputs**: All run-phase verification results.
- **Action**: Persist `~/.claude/projects/{hash}/memory/project_v3r4_harness_namespace_001_complete.md` per plan.md §5.5 Memory Entry template. Update `MEMORY.md` index with single-line entry under 150 chars. Mark any superseded prior memory entries with `[SUPERSEDED by ...]` prefix per Lessons Protocol.
- **Outputs**: New memory file + MEMORY.md index update.
- **Acceptance**: Memory file conforms to 4-type taxonomy (type: project), body includes **Why:** and **How to apply:** sub-lines.
- **Owning agent**: orchestrator (MoAI directly — auto-memory not delegable to subagents).

### T-Wave2-005 — Wave 2 audit artifact persistence

- **Type**: Markdown create
- **Priority**: P1
- **Inputs**: All outputs from T-Wave2-001 through T-Wave2-004.
- **Action**: Create `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave2-<YYYY-MM-DD>.md` containing PR #908 final state, verb-surface diff result, branch deletion confirmation, memory entry path.
- **Outputs**: New audit report file.
- **Acceptance**: File exists, contains all 4 prior task results.
- **Owning agent**: manager-spec or manager-git.

### Wave 2 Acceptance Gate

All 5 tasks above MUST complete with PASS. If T-Wave2-001 surfaces EC-001 BLOCKER, T-Wave2-002 through T-Wave2-005 are deferred until orchestrator resolves with user via AskUserQuestion.

---

## 4. PR Composition Strategy

### 4.1 Run PR (`feat/SPEC-V3R4-HARNESS-NAMESPACE-001` → `main`)

**Title**: `feat(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 run-phase — governance verification + PR #908 closeout`

**Files modified**:

- `CHANGELOG.md` (T-Wave1-006 — append v3.0.0-rc1 entry)
- `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-<DATE>.md` (T-Wave1-007 — new)
- `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave2-<DATE>.md` (T-Wave2-005 — new)

**External side effects** (not in PR diff):

- PR #908 state transition to CLOSED (T-Wave2-002)
- `feat/cmd-harness-slash-wrapper` branch deletion from origin (T-Wave2-002)
- Memory entry creation in `~/.claude/projects/{hash}/memory/` (T-Wave2-004 — local, not in repo)
- MEMORY.md index update (local, not in repo)

**Expected PR size**: ~3 files modified, ~50-200 LOC across all 3 files (mostly audit-report markdown).

**Merge strategy**: SQUASH (per CLAUDE.local.md §18.3 for feat/* → main).

### 4.2 Sync PR (`sync/SPEC-V3R4-HARNESS-NAMESPACE-001` → `main`)

**Title**: `sync(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 lifecycle COMPLETE — namespace + lifecycle governance closed`

**Files modified**:

- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/spec.md` (status: implemented → completed, updated date bump, HISTORY append)
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/plan.md` (HISTORY append referencing run-PR + sync-PR)
- `CHANGELOG.md` (closeout confirmation)

**Expected PR size**: ~3 files modified, ~20-50 LOC (metadata-only).

**Merge strategy**: SQUASH.

**[HARD] Commit prefix**: `sync(spec):` per V3R4-SDF-001 lesson #16 (walker filter requires this prefix; `chore(spec):` is SKIPPED by drift walker).

---

## 5. Verification Procedure (End-to-End Walkthrough)

This section is a step-by-step run-phase walkthrough for the agent executing this SPEC. It mirrors plan.md §2 + §3 but in actionable command form.

### Step 1: Setup

```bash
# Verify on correct branch (after plan-PR merged)
git checkout main && git pull origin main
git checkout -b feat/SPEC-V3R4-HARNESS-NAMESPACE-001

# Verify lint baseline (full-repo scan — CLI rejects directory args)
moai spec lint --strict
```

Expected: `✓ No findings — all SPEC documents are valid`.

### Step 2: Execute Wave 1 tasks

```bash
# T-Wave1-001 — SPEC ID format probe (broadened regex)
mkdir -p .moai/reports/governance/
DATE=$(date +%Y-%m-%d)
REPORT=.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-${DATE}.md
echo "# Wave 1 Audit — SPEC-V3R4-HARNESS-NAMESPACE-001 (${DATE})" > "$REPORT"
echo "" >> "$REPORT"

echo "## T-Wave1-001 — SPEC ID format compliance (broadened regex)" >> "$REPORT"
echo '```' >> "$REPORT"
for d in .moai/specs/SPEC-V3R4-HARNESS-*/ .moai/specs/SPEC-V3R3-HARNESS-*/ .moai/specs/SPEC-V3R3-*-HARNESS-*/; do
  [ -d "$d" ] || continue
  id=$(grep -E '^id:' "$d/spec.md" 2>/dev/null | head -1 | awk '{print $2}')
  echo "$id" | grep -qE '^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$' \
    || echo "VIOLATION: $d → $id"
done | tee -a "$REPORT"
echo '```' >> "$REPORT"
echo "" >> "$REPORT"

# T-Wave1-002 — Dependency graph (declared-dependency discipline per REQ-HRN-NS-002)
echo "## T-Wave1-002 — declared-dependency lifecycle discipline" >> "$REPORT"
# Extract dependencies per AC-HRN-NS-002 verification block; record sibling deps as registered blockers

# T-Wave1-003 — Hierarchy probe (canonical reserved names per REQ-HRN-NS-004)
echo "## T-Wave1-003 — .moai/harness/* hierarchy (incl. README.md + main.md)" >> "$REPORT"
# Run find or NOT_PRESENT per AC-HRN-NS-004; expect main.md + README.md + usage-log.jsonl as canonical

# T-Wave1-004 — Lint (no-args full-repo scan — CLI rejects directory args)
echo "## T-Wave1-004 — Supersede consistency lint (full-repo scan)" >> "$REPORT"
moai spec lint --strict 2>&1 | tee -a "$REPORT"
# Expected: ✓ No findings — all SPEC documents are valid

# T-Wave1-005 — CLI deprecation grace
echo "## T-Wave1-005 — CLI deprecation grace" >> "$REPORT"
test -f internal/cli/harness.go && echo "MARKER_PRESENT" >> "$REPORT"
grep -qE '\bharnessCmd\b' internal/cli/root.go \
  && echo "VIOLATION_REGISTERED" >> "$REPORT" \
  || echo "NO_REGISTRATION" >> "$REPORT"

# T-Wave1-006 — CHANGELOG append (manual edit per plan.md §2.2 content sketch)
# (Edit CHANGELOG.md via Edit tool, NOT shell heredoc — preserves diff context)
# MUST include "retention" + "7-day rolling window" + "REQ-HRN-NS-010" keywords for AC-HRN-NS-008.

# T-Wave1-007 — Audit artifact already being built in $REPORT
```

### Step 3: Execute Wave 2 tasks

```bash
# T-Wave2-001 — PR #908 HEAD probe (prefix matching, EC-001 EXPECTED)
PR908_HEAD=$(gh pr view 908 --json headRefOid -q .headRefOid)
if [[ "$PR908_HEAD" == 452aa638f* ]]; then
  echo "SINGLE_ABSORBED_COMMIT: simple abandon path (rare — branch was reset to absorbed tip)"
  EC_001_PATH=false
else
  echo "DIVERGENCE_CONFIRMED: HEAD=$PR908_HEAD — orchestrator MUST invoke AskUserQuestion (4 EC-001 options)"
  EC_001_PATH=true
fi
# At plan-phase 2026-05-16, EC_001_PATH=true was confirmed (HEAD = a41d6d139...)

# T-Wave2-002 — Conditional disposition execution (see tasks.md §3 T-Wave2-002 for 4 option scripts)
# Orchestrator invokes AskUserQuestion; manager-git executes user-selected option.
# Final verification for options 1-3:
#   gh pr view 908 --json state -q .state         # expect: CLOSED
#   git ls-remote --heads origin feat/cmd-harness-slash-wrapper | wc -l  # expect: 0
# For option 4 (leave-open): record deferral rationale in T-Wave2-005 audit report.

# T-Wave2-003 — Verb-surface diff (corrected regex for numbered headers)
grep -E '^### [0-9.]+ +(status|apply|rollback|disable)\b' .claude/skills/moai/workflows/harness.md \
  | sed -E 's/^### [0-9.]+ +//' | awk '{print tolower($1)}' | sort -u > /tmp/verbs.txt
printf '%s\n' apply disable rollback status > /tmp/expected.txt
diff /tmp/verbs.txt /tmp/expected.txt && echo "PASS: 4 verbs only" || echo "VIOLATION: verb drift"

# T-Wave2-004 — Memory entry (orchestrator action, not subagent — see plan.md §5.5)
# T-Wave2-005 — Wave 2 audit artifact (similar pattern to T-Wave1-007; record EC-001 option chosen)
```

### Step 4: Create run-phase PR

```bash
git add CHANGELOG.md .moai/reports/governance/
git commit -m "$(cat <<'EOF'
feat(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 run-phase — governance verification + PR #908 closeout

Wave 1: 7 governance verification tasks (SPEC ID format, dependency graph,
hierarchy probe, lint, CLI deprecation, CHANGELOG, audit artifact). All PASS.

Wave 2: PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13)
closed as fully absorbed by PR #910 commit bb80ea0f4. Branch deleted.
Verb-surface re-asserted (exactly 4 verbs). Memory entry persisted.

All 8 ACs satisfied (binary PASS). Zero LOC code change.

Governance authority for downstream SPEC-V3R4-HARNESS-{004..008} namespace
compliance now in force.

🗿 MoAI <email@mo.ai.kr>
EOF
)"
git push -u origin feat/SPEC-V3R4-HARNESS-NAMESPACE-001
gh pr create --base main --title "feat(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 run-phase — governance verification + PR #908 closeout" --body "..."
```

### Step 5: Sync-phase (separate PR after run PR merges)

```bash
git checkout main && git pull origin main
git checkout -b sync/SPEC-V3R4-HARNESS-NAMESPACE-001

# Edit spec.md: status: implemented → completed, updated: <today>, HISTORY append
# Edit plan.md: HISTORY append referencing run-PR + sync-PR
# Edit CHANGELOG.md: closeout confirmation

git add .moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/ CHANGELOG.md
git commit -m "sync(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 lifecycle COMPLETE — namespace + lifecycle governance closed

🗿 MoAI <email@mo.ai.kr>"
git push -u origin sync/SPEC-V3R4-HARNESS-NAMESPACE-001
gh pr create --base main --title "sync(spec): SPEC-V3R4-HARNESS-NAMESPACE-001 lifecycle COMPLETE — namespace + lifecycle governance closed" --body "..."
```

---

## 6. Out of Scope (Run-Phase Reminders)

Tasks NOT in this SPEC's run-phase (per spec.md §1.3 + §4):

- Modifying any `internal/cli/*.go` file.
- Modifying any `.claude/skills/moai/workflows/harness.md`, `.claude/commands/moai/harness.md`, or any other skill body / command body.
- Modifying any `.claude/agents/` file (FROZEN-zone-adjacent for moai/* agents).
- Modifying `.claude/rules/moai/design/constitution.md` (FROZEN).
- Modifying SPEC-V3R4-HARNESS-001 or any V3R3 SPEC files.
- Adding new lint rules or walker filters.
- Authoring SPEC-V3R4-HARNESS-{004..008}.
- Modifying any file under `.moai/harness/*` runtime state (T-Wave1-003 is read-only probe).
- Modifying `.claude/agents/my-harness/` or `.claude/skills/my-harness-*/`.
- Performance benchmarking, UX validation, integration testing.
- Networking, telemetry, GUI/dashboard creation.

---

## 7. References

- spec.md §5 (REQs), §6 (Acceptance Coverage Map), §10 (Glossary).
- plan.md §2 (Wave 1), §3 (Wave 2), §4 (Cross-Wave Dependencies), §5 (Technical Approach), §6 (Risks).
- acceptance.md §2 (AC Given-When-Then), §3 (Edge Cases), §4 (Definition of Done).
- design.md §2 (Namespace hierarchy), §3 (Verb state machine), §4 (Lifecycle dependency matrix), §5 (PR #908 closeout architecture).
- SPEC-V3R4-SPECLINT-DEBT-002 tasks.md — precedent for governance SPEC task structure.
- SPEC-V3R4-SDF-001 tasks.md — precedent for `sync(spec):` commit prefix + governance SPEC pattern.
- CLAUDE.local.md §18.3 (Merge Strategy), §18.12 (BODP plan-in-main rationale).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
- `.claude/rules/moai/workflow/moai-memory.md` § Agent Memory Taxonomy.
- `.claude/rules/moai/workflow/session-handoff.md` § Auto-Memory Integration (for T-Wave2-004 memory entry format).

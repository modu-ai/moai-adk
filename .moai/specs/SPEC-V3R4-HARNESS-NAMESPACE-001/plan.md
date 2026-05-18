# SPEC-V3R4-HARNESS-NAMESPACE-001 — Implementation Plan

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.3.0   | 2026-05-16 | MoAI orchestrator | **Lifecycle COMPLETE annotation.** Plan PR #944 (`ea1c10647`) → Run PR #945 (`571258a1e`, merged 2026-05-16T05:19:54Z) → Sync PR #N (this commit). Wave 1 (7 governance verification tasks) + Wave 2 (5 PR #908 closeout tasks) all PASS. PR #908 closed via user-selected EC-001 Option 1 (rollback-tip-then-close): branch reset `a41d6d139c8c → 452aa638f60c` + force-with-lease + attribution comment + close-with-delete-branch. AC-HRN-NS-001~008 binary PASS. Plan steps §2 Wave 1 + §3 Wave 2 + §5 Step 2 Wave 1 shell snippet + §5 Step 3 Wave 2 shell snippet all executed verbatim. EC-001 4-option matrix from §6 Risks resolved via Option 1. Zero deferrals; all 12 tasks (T-Wave1-001~007 + T-Wave2-001~005) marked complete in respective audit reports. |
| 0.2.0   | 2026-05-16 | manager-spec | plan-auditor REVISE 0.69 round-1 fix. T-Wave1-001 SPEC ID regex broadened (D-002); T-Wave1-004 lint switched to no-args full-repo scan (D-003); T-Wave1-006 CHANGELOG content sketch enriched with retention keywords (D-007); T-Wave2-001 PR #908 HEAD divergence reframed as EXPECTED with prefix matching (D-004); T-Wave2-003 verb-surface regex corrected for numbered headers (D-001). |
| 0.1.0   | 2026-05-16 | manager-spec | Initial draft. Pure governance plan: zero LOC code change. Run-phase deliverables = CHANGELOG entry + PR #908 closeout (manager-git) + governance verification (lint + cross-reference). plan-in-main mode (BODP signals all-negative → base `origin/main`). 2-Wave structure (Wave 1 governance verification + Wave 2 PR #908 closeout). target_release v2.20.0-rc1. |

---

## 1. Overview

This plan decomposes SPEC-V3R4-HARNESS-NAMESPACE-001's 10 EARS requirements into 2 Waves. Each Wave is independently PR-able (though both ship in a single sync PR per governance SPEC convention from SPECLINT-DEBT-002 / SDF-001). Both Waves produce only markdown + audit artifacts — NO Go code, NO skill body edits, NO agent body edits.

### 1.1 Execution Mode

- **Plan-in-main**: BODP signals A=¬ (no `depends_on` path overlap with current branch), B=¬ (no co-located SPEC in `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/` outside this PR), C=¬ (no open PR on current branch) → base `origin/main` per CLAUDE.local.md §18.12.
- **Worktree**: NOT used (Phase Discipline Step 1 — plan-in-main + single SPEC directory + docs-only).
- **Branch (plan-phase)**: `feat/SPEC-V3R4-HARNESS-NAMESPACE-001-plan`.
- **Branch (run-phase)**: `feat/SPEC-V3R4-HARNESS-NAMESPACE-001` (created from plan-PR merged main).
- **Branch (sync-phase)**: `sync/SPEC-V3R4-HARNESS-NAMESPACE-001` (created from run-PR merged main).
- **Lifecycle**: spec-anchored (governance SPEC + foundation evolve together — but this SPEC ships independent of foundation evolution).

### 1.2 Wave Breakdown

| Wave | Goal | Files Modified | EARS coverage | Priority | Owning Agent |
|------|------|----------------|---------------|----------|--------------|
| Wave 1 | Governance verification (lint + cross-reference) | `CHANGELOG.md` (append v2.20.0-rc1 entry), audit report | REQ-HRN-NS-001, REQ-HRN-NS-002, REQ-HRN-NS-004, REQ-HRN-NS-005, REQ-HRN-NS-006, REQ-HRN-NS-010 | P1 | `manager-spec` (run-phase delegation) |
| Wave 2 | PR #908 closeout + verb-surface re-assertion documentation | PR #908 close comment + branch delete + memory entry | REQ-HRN-NS-003, REQ-HRN-NS-007, REQ-HRN-NS-008, REQ-HRN-NS-009 | P0 | `manager-git` |

**Plan-phase corpus context**: `.moai/specs/` contains 198 SPEC directories as of main HEAD `82880cd49`. `.moai/specs/SPEC-V3R4-HARNESS-*` = 3 entries (-001, -002, -003). `.moai/specs/SPEC-V3R3-HARNESS-*` = 3 entries (HARNESS-001, HARNESS-LEARNING-001, PROJECT-HARNESS-001). `moai spec lint --strict` PASS state as of plan-phase = ✓ No findings (post-SPECLINT-DEBT-002 + SDF-001 cleanup).

---

## 2. Wave 1 — Governance Verification

### 2.1 Objectives

- Verify SPEC ID format compliance (REQ-HRN-NS-001) across the harness family.
- Verify lifecycle independence via dependency-graph inspection (REQ-HRN-NS-002).
- Verify `.moai/harness/*` hierarchy matches REQ-HRN-NS-004 canonical sub-paths.
- Verify supersede consistency under `moai spec lint --strict` (REQ-HRN-NS-005).
- Re-assert CLI deprecation grace from V3R4-HARNESS-001 (REQ-HRN-NS-006).
- Document usage-log retention floor (REQ-HRN-NS-010).
- Append CHANGELOG v2.20.0-rc1 entry documenting the governance closeout.

### 2.2 Tasks

#### T-Wave1-001: SPEC ID format compliance verification

- **What**: For every directory under `.moai/specs/SPEC-V3R4-HARNESS-*/`, `.moai/specs/SPEC-V3R3-HARNESS-*/`, and `.moai/specs/SPEC-V3R3-*-HARNESS-*/` (SCOPE-leading variant), read the `id:` field from spec.md frontmatter and verify it matches the **broadened regex** `^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$` (admits both HARNESS-leading and SCOPE-leading variants).
- **Where**: Shell command in run-phase verification report.
- **Verification**:
  ```bash
  for d in .moai/specs/SPEC-V3R4-HARNESS-*/ .moai/specs/SPEC-V3R3-HARNESS-*/ .moai/specs/SPEC-V3R3-*-HARNESS-*/; do
    [ -d "$d" ] || continue
    id=$(grep -E '^id:' "$d/spec.md" 2>/dev/null | head -1 | awk '{print $2}')
    echo "$id" | grep -qE '^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$' \
      || echo "VIOLATION: $d → $id"
  done
  ```
  Expected: zero VIOLATION lines. Verified at plan-phase against 7 corpus IDs — all PASS.

#### T-Wave1-002: Lifecycle independence — dependency graph inspection

- **What**: For each SPEC in the harness family, extract `dependencies:` and verify no cyclic dependency exists. Verify that V3R4-HARNESS-{002, 003} declare `dependencies: [SPEC-V3R4-HARNESS-001]` (or empty) only.
- **Where**: Shell command + manual review in run-phase verification report.
- **Verification**:
  ```bash
  for d in .moai/specs/SPEC-V3R4-HARNESS-*/; do
    echo "=== $d ==="
    grep -A 5 '^dependencies:' "$d/spec.md" | head -10
  done
  ```
  Expected output schema: each SPEC's `dependencies:` block either empty (`[]`) or contains only `SPEC-V3R4-HARNESS-001` (foundation). No SPEC depends on `-002`, `-003`, etc. (no transitive blocker).

#### T-Wave1-003: `.moai/harness/*` hierarchy probe

- **What**: Document the current state of `.moai/harness/` if it exists. If absent (fresh-start project), document the canonical structure as forward-looking guidance. The probe is non-destructive; no files are created.
- **Where**: Shell command + report in run-phase audit artifact.
- **Verification**:
  ```bash
  if [ -d .moai/harness ]; then
    find .moai/harness -maxdepth 3 -type f -o -type d | sort > /tmp/harness-hierarchy.txt
    cat /tmp/harness-hierarchy.txt
  else
    echo "NOT_PRESENT: .moai/harness/ does not exist (expected for fresh-start)"
  fi
  ```
  Expected: present sub-paths (if any) MUST be a subset of `{usage-log.jsonl, proposals/, learning-history/snapshots/, learning-history/applied/, learning-history/frozen-guard-violations.jsonl, learning-history/tier-promotions.jsonl}`. NOT_PRESENT is acceptable.

#### T-Wave1-004: Supersede consistency lint (no-args full-repo scan)

- **What**: Run `moai spec lint --strict` (no args, full-repo scan) and assert exit code 0 + `✓ No findings — all SPEC documents are valid`. The CLI does NOT accept directory args (`moai spec lint --strict <dir>` returns `ParseFailure: is a directory`); full-repo scan provides stronger governance guarantee anyway.
- **Where**: Shell command in run-phase verification.
- **Verification**:
  ```bash
  moai spec lint --strict 2>&1 | tee /tmp/harness-lint.txt
  grep -q '✓ No findings' /tmp/harness-lint.txt
  ```
  Expected: exit code 0, stdout contains `✓ No findings — all SPEC documents are valid`.

#### T-Wave1-005: CLI deprecation grace re-assertion

- **What**: Verify `internal/cli/harness.go` exists in the tree but is NOT registered in `internal/cli/root.go` cobra command tree.
- **Where**: Shell command in run-phase verification.
- **Verification**:
  ```bash
  test -f internal/cli/harness.go || echo "VIOLATION: deprecation marker file missing"
  grep -E 'harness.*Cmd' internal/cli/root.go && echo "VIOLATION: harness subcommand registered" || echo "PASS: no harness registration"
  ```
  Expected: deprecation marker present, no registration.

#### T-Wave1-006: CHANGELOG v2.20.0-rc1 entry append

- **What**: Append a new entry under the v2.20.0-rc1 section of `CHANGELOG.md` documenting (a) the namespace governance closeout, (b) the PR #908 abandonment decision, (c) reference to SPEC-V3R4-HARNESS-001 + this SPEC.
- **Where**: `CHANGELOG.md` (top-of-file, under "## [Unreleased]" or "## [v2.20.0-rc1]" header per existing convention).
- **Verification**: `grep -A 3 'SPEC-V3R4-HARNESS-NAMESPACE-001' CHANGELOG.md` returns non-empty output.
- **Content sketch**:
  ```markdown
  ### Governance

  - `SPEC-V3R4-HARNESS-NAMESPACE-001`: harness family namespace + lifecycle governance.
    Formalizes the canonical SPEC ID format (admitting both HARNESS-leading and
    SCOPE-leading variants per REQ-HRN-NS-001), the `.moai/harness/*` directory
    hierarchy with canonical reserved names (`README.md`, `main.md`, `usage-log.jsonl`,
    `proposals/`, `learning-history/`), the `/moai harness` 4-verb stable surface
    (`status`, `apply`, `rollback`, `disable`), and declared-dependency lifecycle
    discipline (sibling deps act as registered blockers) across the eight-SPEC family.
    Complements `SPEC-V3R4-HARNESS-001` foundation without amendment. Zero LOC code change.
  - **Retention policy**: usage-log.jsonl 7-day rolling window (REQ-HRN-NS-010);
    older entries archived to learning-history/snapshots/ when downstream rotation SPEC ships.
    learning-history/frozen-guard-violations.jsonl and learning-history/snapshots/
    remain inviolate (no rotation).
  - PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) closeout via
    orchestrator-issued AskUserQuestion (4 options: rollback-tip-then-close (Recommended) /
    cherry-pick-to-new-PR-then-close / abandon-new-commits / leave-open) per EC-001.
    Branch HEAD as of plan-phase: `a41d6d139c8c769bf395a25a055d59c14e180191` (2-commit
    divergence: Commit 1 `452aa638f` absorbed by PR #910 commit `bb80ea0f4`; Commit 2
    `a41d6d139` audit doc unabsorbed).
  ```

#### T-Wave1-007: Wave 1 audit artifact

- **What**: Persist all Wave 1 verification outputs into `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-<YYYY-MM-DD>.md`.
- **Where**: New report file under `.moai/reports/governance/`.
- **Verification**: Report file exists, contains sections "T-Wave1-001 through T-Wave1-006 results", each section shows PASS or documented exception.

### 2.3 Wave 1 Acceptance Gate

Wave 1 passes when:

1. T-Wave1-001 produces zero VIOLATION lines.
2. T-Wave1-002 dependency graph shows no cycle and no cross-SPEC blocker beyond foundation.
3. T-Wave1-003 hierarchy probe shows canonical sub-paths only (or NOT_PRESENT).
4. T-Wave1-004 lint shows ✓ No findings.
5. T-Wave1-005 confirms deprecation marker present, no registration.
6. T-Wave1-006 CHANGELOG entry present.
7. T-Wave1-007 audit artifact persisted.

---

## 3. Wave 2 — PR #908 Closeout

### 3.1 Objectives

- Close PR #908 (`feat/cmd-harness-slash-wrapper`) as fully absorbed.
- Delete the `feat/cmd-harness-slash-wrapper` branch post-closeout.
- Record the closeout in `.moai/reports/governance/` and in an auto-memory project entry.
- Verify verb-surface re-assertion (REQ-HRN-NS-003) survives PR #908 closure with no functional regression.

### 3.2 Tasks

#### T-Wave2-001: PR #908 HEAD probe (divergence is EXPECTED)

- **What**: Probe PR #908 HEAD via `gh pr view 908 --json headRefOid` and detect divergence using **prefix matching** (NOT string equality — `gh` returns full 40-char SHAs). Divergence is the EXPECTED scenario at SPEC merge time (plan-phase verified HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191`). On divergence, set escalation flag for orchestrator to invoke AskUserQuestion (4 options per EC-001).
- **Where**: Shell command in run-phase verification.
- **Verification**:
  ```bash
  gh pr view 908 --json headRefOid,headRefName,state | tee /tmp/pr908.json
  head=$(jq -r .headRefOid /tmp/pr908.json)
  if [[ "$head" == 452aa638f* ]]; then
    echo "SINGLE_ABSORBED_COMMIT: simple abandon path (rare — branch was reset)"
    EC_001_PATH=false
  else
    echo "DIVERGENCE_CONFIRMED: HEAD=$head — orchestrator MUST invoke AskUserQuestion (4 options per EC-001)"
    EC_001_PATH=true
  fi
  echo "EC_001_PATH=$EC_001_PATH" >> /tmp/wave2-state.env
  ```
  Expected at plan-phase 2026-05-16: `EC_001_PATH=true` (HEAD = `a41d6d139...`).

#### T-Wave2-002: PR #908 close with attribution comment

- **What**: `manager-git` posts a closing comment to PR #908 referencing PR #910 commit `bb80ea0f4` as the absorption commit and this SPEC (`SPEC-V3R4-HARNESS-NAMESPACE-001`) as the governance closure authority. Then close the PR.
- **Where**: `gh pr close` invocation.
- **Verification**:
  ```bash
  gh pr comment 908 --body "Closed as fully absorbed by PR #910 commit \`bb80ea0f4\` (V3R4-HARNESS-001 run-phase). Governance closure authority: SPEC-V3R4-HARNESS-NAMESPACE-001 REQ-HRN-NS-007."
  gh pr close 908 --delete-branch
  gh pr view 908 --json state | jq -r .state
  ```
  Expected: final output = `CLOSED`. Branch `feat/cmd-harness-slash-wrapper` deleted from origin.

#### T-Wave2-003: Verb-surface re-assertion verification

- **What**: Confirm the `/moai harness` workflow body at `.claude/skills/moai/workflows/harness.md` still exposes exactly the four canonical verbs (`status`, `apply`, `rollback`, `disable`) with no silent additions. The headers are **numbered** (e.g., `### 2.1 status`); the regex MUST admit numeric prefixes.
- **Where**: Shell command + manual diff against V3R4-HARNESS-001 REQ-HRN-FND-003 reference.
- **Verification**:
  ```bash
  grep -E '^### [0-9.]+ +(status|apply|rollback|disable)\b' .claude/skills/moai/workflows/harness.md \
    | sed -E 's/^### [0-9.]+ +//' | awk '{print tolower($1)}' | sort -u > /tmp/verbs.txt
  printf '%s\n' apply disable rollback status > /tmp/expected.txt
  diff /tmp/verbs.txt /tmp/expected.txt \
    && echo "PASS: 4 verbs only" || echo "VIOLATION: verb-surface drift detected"
  ```
  Expected: PASS (4 lines in `/tmp/verbs.txt` matching `/tmp/expected.txt`). Verified at plan-phase: lines 132/166/196/208 — all 4 verbs matched. If diff shows any other verb header, REQ-HRN-NS-008 is triggered.

#### T-Wave2-004: Memory entry for closeout

- **What**: Persist a project memory entry documenting the PR #908 closeout per session-handoff conventions (auto-memory taxonomy: type=project, body includes Why + How-to-apply).
- **Where**: `~/.claude/projects/{hash}/memory/project_v3r4_harness_namespace_001_complete.md` (created in run-phase by orchestrator, not by this SPEC). Update `MEMORY.md` index with a one-line entry.
- **Verification**: Memory entry conforms to 4-type taxonomy frontmatter (type: project) and includes verbatim resume message per `.claude/rules/moai/workflow/session-handoff.md` §Auto-Memory Integration.

#### T-Wave2-005: Wave 2 audit artifact

- **What**: Persist all Wave 2 verification outputs into `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave2-<YYYY-MM-DD>.md`.
- **Where**: New report file under `.moai/reports/governance/`.
- **Verification**: Report file exists, includes PR #908 final state, verb-surface diff result, branch deletion confirmation.

### 3.3 Wave 2 Acceptance Gate

Wave 2 passes when:

1. T-Wave2-001 PR #908 HEAD probe completes; if `EC_001_PATH=true` (HEAD ≠ `452aa638f*` prefix), orchestrator AskUserQuestion round completes with user-selected option (rollback-tip / cherry-pick / abandon / leave-open). If `EC_001_PATH=false` (single absorbed commit), proceed directly to T-Wave2-002.
2. T-Wave2-002 PR #908 state transitions to `CLOSED` with attribution comment posted (per user-selected EC-001 option); branch deleted unless `leave-open` was chosen.
3. T-Wave2-003 verb-surface diff shows PASS (exactly 4 verbs).
4. T-Wave2-004 memory entry persisted per taxonomy.
5. T-Wave2-005 audit artifact persisted.

---

## 4. Cross-Wave Dependencies

Wave 1 MUST complete before Wave 2 begins:

- Wave 1 establishes the governance verification baseline (lint PASS, hierarchy probe, dependency graph).
- Wave 2 acts on the baseline (PR #908 closeout assumes lint + verb-surface integrity from Wave 1).

Within each Wave, tasks are sequential (T-Wave1-001 → T-Wave1-007 in order; T-Wave2-001 → T-Wave2-005 in order). Parallelization is NOT applicable — this is a docs-only governance SPEC with minimal task interdependency.

---

## 5. Technical Approach

### 5.1 Pure Governance Pattern

This SPEC follows the pure-governance pattern established by:

- `SPEC-V3R4-SPECLINT-DEBT-002` (PR #943, merged) — 12-field canonical frontmatter SSOT, zero code change in spec.md / acceptance.md surface (small change in plan workflow body only).
- `SPEC-V3R4-SDF-001` (PR #940, merged) — 77-SPEC status drift sweep + terminal-state exemption, zero code change (metadata sweep + walker filter only).

The pattern:

1. SPEC artifacts (spec.md, plan.md, acceptance.md) ship in plan-phase.
2. Run-phase produces audit artifacts + CHANGELOG entry + cross-PR coordination (close/comment).
3. Sync-phase merges the run PR, then a separate sync PR (small, metadata-only) closes the lifecycle with status transition to `completed`.
4. Each phase produces an independent PR; lint passes at every gate.

### 5.2 No Code Change Discipline

- No `.go` file modification.
- No `.claude/skills/` body modification (except SSOT documentation if a new file is added — NOT needed for this SPEC).
- No `.claude/agents/` body modification.
- No `.claude/rules/` modification (FROZEN-adjacent — only design constitution is strictly FROZEN, but other rules are inherited unchanged).
- No `.moai/config/sections/` modification.

Only file modifications:
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/spec.md` (plan-phase, this PR)
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/plan.md` (plan-phase, this PR)
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/acceptance.md` (plan-phase, this PR)
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/design.md` (plan-phase, this PR — namespace hierarchy diagram + verb state machine)
- `.moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/tasks.md` (plan-phase, this PR — run-phase task list, governance verification only)
- `CHANGELOG.md` (run-phase, T-Wave1-006)
- `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave{1,2}-*.md` (run-phase audit artifacts, T-Wave1-007 + T-Wave2-005)

### 5.3 Lint Gate Strategy

`moai spec lint --strict` (no-args full-repo scan — the CLI rejects directory args with `ParseFailure: is a directory`) is invoked at three checkpoints:

1. **Plan-phase post-author** (this Step 3): `moai spec lint --strict` — verifies the full SPEC corpus including the new SPEC artifacts pass 12-field canonical schema + EARS format + REQ-AC coverage.
2. **Run-phase T-Wave1-004**: `moai spec lint --strict` — full-repo regression check.
3. **Sync-phase pre-merge**: `moai spec lint --strict` — final verification before status transition to `completed`.

All three checkpoints use the same no-args invocation. Any lint failure at any checkpoint blocks PR merge.

### 5.4 PR #908 Closeout — Safety Considerations

- The closeout protocol uses **prefix matching** (`[[ "$head" == 452aa638f* ]]`) on PR HEAD, not string equality — `gh` returns full 40-char SHAs and prefix matching survives both single-absorbed-commit and multi-commit divergence cases.
- Plan-phase verified PR #908 HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191` (2-commit divergence: Commit 1 `452aa638f` absorbed, Commit 2 `a41d6d139` audit doc unabsorbed). This means EC-001 is the EXPECTED disposition path at run-phase.
- The orchestrator (NOT `manager-git`) invokes `AskUserQuestion` with the 4 EC-001 options. `manager-git` only executes the user-selected option. Force-close is prohibited for all 4 options.
- For options 1-3 (rollback / cherry-pick / abandon), branch deletion is via `gh pr close --delete-branch`, reversible within GitHub's deletion grace period if needed. Option 4 (leave-open) makes no changes to the PR or branch.
- The attribution comment (option 2 chooses to add it; option 3 may omit it) is the audit trail when present; the comment text MUST reference both PR #910 commit hash and this SPEC's ID.

### 5.5 Memory Entry — Taxonomy Compliance

The Wave 2 T-Wave2-004 memory entry MUST conform to `.claude/rules/moai/workflow/moai-memory.md` § Agent Memory Taxonomy:

```markdown
---
name: project-v3r4-harness-namespace-001-complete
description: Harness family namespace governance closeout — PR #908 abandoned, lifecycle independence locked in for V3R4-HARNESS-{001..008}.
type: project
---

PR #908 (`feat/cmd-harness-slash-wrapper`) closed as fully absorbed by PR #910 commit `bb80ea0f4`.
SPEC-V3R4-HARNESS-NAMESPACE-001 documents the 4-verb stable surface, `.moai/harness/*` hierarchy,
and lifecycle independence guarantee for the eight-SPEC harness family.

**Why:** PR #908 was the original `/moai harness` slash thin-wrapper, but its content was already
imported into V3R4-HARNESS-001 run-phase (PR #910). Leaving PR #908 OPEN risked accidental
re-merge of conflicting content. Governance SPEC now binds the namespace + verb-surface stability
for downstream SPECs 002-008.

**How to apply:** Future harness-family SPEC authors MUST follow REQ-HRN-NS-001 through
REQ-HRN-NS-009 — no fifth verb, no path drift, no cross-SPEC lifecycle blocker. Verify via
`moai spec lint --strict` PASS.

## 다음 세션 시작점 (paste-ready resume message)

(populated by orchestrator at run-phase completion)
```

---

## 6. Risks

| Risk | Likelihood | Wave | Mitigation |
|------|------------|------|------------|
| PR #908 HEAD diverged from `452aa638f` (CONFIRMED — HEAD = `a41d6d139...`) | **High (EXPECTED)** | Wave 2 | T-Wave2-001 detects divergence via prefix matching; orchestrator invokes `AskUserQuestion` with 4 EC-001 options. EXPECTED path, not edge case. |
| `moai spec lint --strict` regression on harness-family directories due to V3R4-HARNESS-002 or -003 status drift | Medium | Wave 1 | T-Wave1-004 surfaces any drift; if found, defer to a follow-up status-sync SPEC rather than expanding this SPEC's scope. |
| `.moai/harness/` hierarchy on dev machine differs from canonical layout (e.g., user has experimental sub-dirs) | Low | Wave 1 | T-Wave1-003 reports NOT_PRESENT as acceptable; non-canonical paths are flagged in the audit but NOT auto-removed (no destructive operation). |
| CHANGELOG merge conflict with concurrent v2.20.0-rc1 entries | Low | Wave 1 | T-Wave1-006 appends under existing v2.20.0-rc1 section; conflict resolution defers to `manager-git` rebase. |
| Memory entry exceeds MEMORY.md 200-line limit | Low | Wave 2 | T-Wave2-004 keeps the index entry to a single line under 150 chars per moai-memory.md §MEMORY.md Line Cap. |
| plan-auditor rejects this SPEC due to docs-only run-phase being misread as "no implementation" | Low | n/a (audit) | SPECLINT-DEBT-002 and SDF-001 precedents established that pure-governance SPECs are valid; plan-auditor scoring rubric accepts governance verification + cross-PR coordination as legitimate run-phase deliverables. |

---

## 7. Acceptance Criteria Summary

See `acceptance.md` for full Given-When-Then scenarios. Binary AC summary:

| AC ID | Wave | Pass criterion |
|-------|------|----------------|
| AC-HRN-NS-001 | Wave 1 | T-Wave1-001 + T-Wave1-004 PASS |
| AC-HRN-NS-002 | Wave 1 | T-Wave1-002 dependency graph acyclic, only foundation depends |
| AC-HRN-NS-003 | Wave 2 | T-Wave2-003 verb-surface diff PASS (4 verbs) |
| AC-HRN-NS-004 | Wave 1 | T-Wave1-003 hierarchy probe matches canonical sub-paths or NOT_PRESENT |
| AC-HRN-NS-005 | Wave 1 | T-Wave1-004 `moai spec lint --strict` ✓ No findings |
| AC-HRN-NS-006 | Wave 1 | T-Wave1-005 deprecation marker present, no registration |
| AC-HRN-NS-007 | Wave 2 | T-Wave2-001 + T-Wave2-002 PR #908 state = CLOSED, branch deleted |
| AC-HRN-NS-008 | Wave 1 | T-Wave1-006 CHANGELOG entry present (retention policy reference) |

All 8 ACs are binary: PASS / FAIL with no partial credit.

---

## 8. Out of Scope (Plan-Phase Reminders)

Repeating spec.md §1.3 and §4 for plan-phase emphasis — DO NOT include any of these in run-phase tasks:

1. Modifying `internal/cli/harness.go`, `internal/cli/root.go`, or any `internal/cli/*.go` file.
2. Modifying `.claude/skills/moai/workflows/harness.md` or `.claude/commands/moai/harness.md`.
3. Modifying SPEC-V3R4-HARNESS-001 spec.md / plan.md / acceptance.md / tasks.md / follow-up.md.
4. Modifying the three superseded V3R3 SPECs.
5. Modifying `.claude/rules/moai/design/constitution.md`.
6. Adding a new lint rule or walker filter.
7. Authoring SPEC-V3R4-HARNESS-{004..008}.
8. Modifying `.moai/harness/*` runtime files (read-only probe in T-Wave1-003).
9. Modifying `.claude/agents/my-harness/` or `.claude/skills/my-harness-*/`.
10. Migration tooling.
11. GUI / dashboard / non-text interface.
12. Networking / telemetry.
13. Adding worktree, tmux, or Agent Teams configuration.

---

## 9. References

- spec.md §1, §2, §5 (REQs), §11 (References) — same file.
- acceptance.md — sibling file.
- design.md — sibling file (namespace hierarchy diagram + verb state machine).
- tasks.md — sibling file (run-phase task list).
- SPEC-V3R4-SPECLINT-DEBT-002 plan.md — governance SPEC plan-pattern precedent.
- SPEC-V3R4-SDF-001 plan.md — governance SPEC + sync commit prefix discipline precedent.
- CLAUDE.local.md §18.12 — BODP for plan-in-main mode justification.
- `.claude/rules/moai/workflow/spec-workflow.md` § Plan Phase + § Run Phase + § Sync Phase.
- `.claude/rules/moai/workflow/moai-memory.md` § Agent Memory Taxonomy.
- `.claude/rules/moai/workflow/session-handoff.md` § Auto-Memory Integration.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema SSOT.

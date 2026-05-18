# SPEC-V3R4-HARNESS-NAMESPACE-001 — Acceptance Criteria

## 1. Coverage Map

### 1.1 AC ↔ REQ Mapping

| AC ID | Wave | Covers REQ IDs | Binary outcome | Owning verifier |
|-------|------|----------------|-----------------|------------------|
| AC-HRN-NS-001 (SPEC ID format compliance) | Wave 1 | REQ-HRN-NS-001 | PASS = zero VIOLATION lines | `moai spec lint` + shell regex |
| AC-HRN-NS-002 (declared-dependency lifecycle discipline) | Wave 1 | REQ-HRN-NS-002 | PASS = acyclic + sibling deps are explicit registered blockers | shell dependency-graph inspection |
| AC-HRN-NS-003 (verb-surface stability) | Wave 2 | REQ-HRN-NS-003, REQ-HRN-NS-008 | PASS = exactly 4 verbs, no drift | shell grep + diff |
| AC-HRN-NS-004 (namespace hierarchy) | Wave 1 | REQ-HRN-NS-004, REQ-HRN-NS-009 | PASS = canonical sub-paths or NOT_PRESENT | shell `find` |
| AC-HRN-NS-005 (supersede consistency lint) | Wave 1 | REQ-HRN-NS-005 | PASS = `moai spec lint --strict` ✓ No findings | `moai spec lint --strict` |
| AC-HRN-NS-006 (CLI deprecation grace) | Wave 1 | REQ-HRN-NS-006 | PASS = file present, not registered | shell `test` + `grep` |
| AC-HRN-NS-007 (PR #908 closeout — orchestrator escalation EXPECTED) | Wave 2 | REQ-HRN-NS-007 | PASS = user-selected disposition (options 1-3: CLOSED + branch deleted; option 4: deferral persisted) | orchestrator `AskUserQuestion` + `manager-git` execution |
| AC-HRN-NS-008 (usage-log retention floor) | Wave 1 | REQ-HRN-NS-010 | PASS = CHANGELOG retention reference present | shell `grep` |

Total: 8 ACs covering 10 REQs (REQ-HRN-NS-008 + REQ-HRN-NS-009 share AC-HRN-NS-003 + AC-HRN-NS-004 respectively as enforcement clauses).

---

## 2. Acceptance Criteria — Given-When-Then

### AC-HRN-NS-001 (SPEC ID format compliance)

**Given** the harness-family SPEC directories exist under `.moai/specs/SPEC-V3R4-HARNESS-*/`, `.moai/specs/SPEC-V3R3-HARNESS-*/`, and `.moai/specs/SPEC-V3R3-*-HARNESS-*/` (SCOPE-leading variant such as `SPEC-V3R3-PROJECT-HARNESS-001`),
**When** the orchestrator executes the SPEC ID format probe using the **broadened regex** that admits both HARNESS-leading and SCOPE-leading variants:
```bash
for d in .moai/specs/SPEC-V3R4-HARNESS-*/ .moai/specs/SPEC-V3R3-HARNESS-*/ .moai/specs/SPEC-V3R3-*-HARNESS-*/; do
  [ -d "$d" ] || continue
  id=$(grep -E '^id:' "$d/spec.md" 2>/dev/null | head -1 | awk '{print $2}')
  echo "$id" | grep -qE '^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$' \
    || echo "VIOLATION: $d → $id"
done
```
**Then** the command output **shall** contain zero lines starting with `VIOLATION:`. Verified at plan-phase against 7 corpus IDs (SPEC-V3R3-HARNESS-001, SPEC-V3R3-HARNESS-LEARNING-001, SPEC-V3R3-PROJECT-HARNESS-001, SPEC-V3R4-HARNESS-001, SPEC-V3R4-HARNESS-002, SPEC-V3R4-HARNESS-003, SPEC-V3R4-HARNESS-NAMESPACE-001) — all PASS.

**Binary outcome**: PASS if zero VIOLATION lines, FAIL otherwise.
**Verification context**: Run-phase T-Wave1-001.
**Covers**: REQ-HRN-NS-001.

---

### AC-HRN-NS-002 (declared-dependency lifecycle discipline)

**Given** each harness-family SPEC declares its `dependencies:` field in spec.md frontmatter,
**When** the orchestrator extracts and analyzes the dependency graph:
```bash
for d in .moai/specs/SPEC-V3R4-HARNESS-*/; do
  spec_id=$(grep -E '^id:' "$d/spec.md" | head -1 | awk '{print $2}')
  deps=$(awk '/^dependencies:/{flag=1;next} /^[a-z_]+:/{flag=0} flag' "$d/spec.md" \
    | grep -oE 'SPEC-V3R[0-9]+-[A-Z][A-Z0-9-]*-[0-9]{3}' | sort -u)
  echo "$spec_id depends_on: ${deps:-NONE}"
done
```

**Then** the dependency graph **shall** satisfy ALL of the following:

1. **Acyclic**: No cycle (no `A → B → A` chain). Verifiable by topological sort or by manual inspection of the linear declaration list.
2. **No undeclared blocker**: Every blocker between family members is a declared `dependencies:` entry, never an implicit cross-PR convention.
3. **Sibling deps are explicit registered blockers** (NOT prohibited): If SPEC X declares `dependencies: [SPEC-V3R4-HARNESS-Y]` where Y is a sibling, then X's run PR and sync PR are blocked until Y reaches `status: implemented` or `status: completed`. This is allowed and is the discipline required by REQ-HRN-NS-002.
4. **Foundation dep always permitted**: Every SPEC may declare `SPEC-V3R4-HARNESS-001` as a dependency without further justification.

**Verified at plan-phase against main HEAD `82880cd49`**:
- `SPEC-V3R4-HARNESS-001 depends_on: NONE` (foundation, no deps)
- `SPEC-V3R4-HARNESS-002 depends_on: SPEC-V3R4-HARNESS-001` (foundation only — independent)
- `SPEC-V3R4-HARNESS-003 depends_on: SPEC-V3R4-HARNESS-001, SPEC-V3R4-HARNESS-002` (sibling dep on -002 = registered blocker; -003 cannot merge run/sync until -002 is implemented or completed)
- `SPEC-V3R4-HARNESS-NAMESPACE-001 depends_on: SPEC-V3R4-HARNESS-001` (this SPEC — foundation only)

**Binary outcome**: PASS if all 4 conditions above hold AND graph is acyclic, FAIL otherwise.
**Verification context**: Run-phase T-Wave1-002.
**Covers**: REQ-HRN-NS-002.

---

### AC-HRN-NS-003 (verb-surface stability — exactly 4 verbs)

**Given** the `/moai harness` workflow body exists at `.claude/skills/moai/workflows/harness.md` with **numbered section headers** (verified at plan-phase: `### 2.1 status (file-IO only — no binary)` at line 132, `### 2.2 apply (5-Layer Safety pipeline, file-IO only)` at line 166, `### 2.3 rollback` at line 196, `### 2.4 disable` at line 208),
**When** the orchestrator probes for verb headers using the **corrected regex** that admits numeric prefixes:
```bash
grep -E '^### [0-9.]+ +(status|apply|rollback|disable)\b' .claude/skills/moai/workflows/harness.md \
  | sed -E 's/^### [0-9.]+ +//' | awk '{print tolower($1)}' | sort -u > /tmp/verbs.txt
printf '%s\n' apply disable rollback status > /tmp/expected.txt
diff /tmp/verbs.txt /tmp/expected.txt
```
**Then** the `diff` output **shall** be empty (exit code 0), confirming exactly four verbs (`status`, `apply`, `rollback`, `disable`) are exposed and no fifth verb has been silently introduced. Verified at plan-phase — all 4 verbs matched.

**And** any future SPEC that adds a fifth verb header to `.claude/skills/moai/workflows/harness.md` without explicitly amending REQ-HRN-NS-003 in a new governance SPEC **shall** trigger plan-auditor rejection per REQ-HRN-NS-008.

**Binary outcome**: PASS if diff exit code = 0, FAIL otherwise.
**Verification context**: Run-phase T-Wave2-003.
**Covers**: REQ-HRN-NS-003, REQ-HRN-NS-008.

---

### AC-HRN-NS-004 (namespace hierarchy — canonical reserved names only)

**Given** the project may have a `.moai/harness/` directory (created during V3R4-HARNESS-001 run-phase) or may not (fresh-start),
**When** the orchestrator probes the hierarchy:
```bash
if [ -d .moai/harness ]; then
  find .moai/harness -maxdepth 3 \( -type f -o -type d \) | sed 's|^.moai/harness/||' \
    | grep -v '^$' | sort > /tmp/hierarchy.txt
else
  echo "NOT_PRESENT" > /tmp/hierarchy.txt
fi
```
**Then** every entry in `/tmp/hierarchy.txt` **shall** match one of the canonical reserved names:
- `README.md` (directory orientation, optional)
- `main.md` (harness instance metadata — V3R4-HARNESS-001 run-phase artifact)
- `usage-log.jsonl` (PostToolUse observation log, JSONL append-only)
- `proposals/` (Tier-X candidate proposal artifacts)
- `proposals/*` (any file beneath)
- `learning-history/`
- `learning-history/snapshots/` (pre-application snapshots)
- `learning-history/snapshots/*` (date-named sub-dirs and their contents)
- `learning-history/applied/` (applied evolution records)
- `learning-history/applied/*`
- `learning-history/frozen-guard-violations.jsonl` (L1 Frozen Guard rejection audit log)
- `learning-history/tier-promotions.jsonl` (Tier 1→2→3→4 promotion event log)

**Or** the file `/tmp/hierarchy.txt` **shall** contain the single line `NOT_PRESENT` (acceptable fresh-start state).

**Verified at plan-phase against actual project state**: `.moai/harness/` contains `main.md`, `README.md`, `usage-log.jsonl` — all 3 entries match the canonical reserved names list above. AC PASS.

**And** any future SPEC that introduces a harness-state file outside this canonical set without amending REQ-HRN-NS-004 **shall** trigger plan-auditor rejection per REQ-HRN-NS-009.

**Binary outcome**: PASS if hierarchy is a subset of canonical OR NOT_PRESENT, FAIL if any non-canonical entry is present.
**Verification context**: Run-phase T-Wave1-003.
**Covers**: REQ-HRN-NS-004, REQ-HRN-NS-009.

---

### AC-HRN-NS-005 (supersede consistency — lint PASS)

**Given** the full SPEC corpus exists under `.moai/specs/`,
**When** the orchestrator runs the **no-args full-repo scan** (the CLI does NOT accept directory args — `moai spec lint --strict <dir>` returns `ParseFailure: is a directory`):
```bash
moai spec lint --strict
```
**Then** the command **shall** exit with code 0,
**And** stdout **shall** contain the indicator `✓ No findings — all SPEC documents are valid`,
**And** no warning **shall** reference orphan `supersedes:` chains, missing `superseded_by:` fields on the three V3R3 SPECs, or status-transition drift.

**Rationale for full-repo scan**: governance SPECs (this SPEC, SPECLINT-DEBT-002, SDF-001) verify cross-SPEC consistency where harness-family integrity depends on the broader corpus being lint-clean. A directory-scoped scan would mask drift in unrelated SPECs that could regress harness namespace governance via shared lint rules.

**Note**: The status transition of the three V3R3 SPECs to `superseded` is the responsibility of an inherited follow-up `manager-git` commit per V3R4-HARNESS-001 REQ-HRN-FND-013. If that commit has not yet been executed, the V3R3 SPECs may still show `status: completed` or earlier — but this **shall not** produce a lint finding because the `supersedes:` declaration in V3R4-HARNESS-001 already establishes the relationship.

**Binary outcome**: PASS if exit code = 0 AND zero-findings indicator present, FAIL otherwise.
**Verification context**: Run-phase T-Wave1-004.
**Covers**: REQ-HRN-NS-005.

---

### AC-HRN-NS-006 (CLI deprecation grace — file present, not registered)

**Given** the CLI deprecation contract from V3R4-HARNESS-001 BC-V3R4-HARNESS-001-CLI-RETIREMENT,
**When** the orchestrator probes the source tree:
```bash
test -f internal/cli/harness.go && echo "MARKER_PRESENT" || echo "VIOLATION_MISSING"
grep -E '\bharnessCmd\b|harness\..*Cmd' internal/cli/root.go && echo "VIOLATION_REGISTERED" || echo "NO_REGISTRATION"
```
**Then** the first command **shall** print `MARKER_PRESENT` (deprecation marker file exists),
**And** the second command **shall** print `NO_REGISTRATION` (no harness subcommand wired into the cobra command tree).

**Binary outcome**: PASS if both conditions hold (MARKER_PRESENT + NO_REGISTRATION), FAIL otherwise.
**Verification context**: Run-phase T-Wave1-005.
**Covers**: REQ-HRN-NS-006.

---

### AC-HRN-NS-007 (PR #908 closeout — orchestrator escalation EXPECTED)

**Given** PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) is targeted for closeout, **and** plan-phase 2026-05-16 verified PR #908 HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191` (2-commit divergence from absorbed `452aa638f`),

**When** the orchestrator detects divergence using **prefix matching** (not string equality — `gh` returns full 40-char SHAs):
```bash
head=$(gh pr view 908 --json headRefOid -q .headRefOid)
if [[ "$head" == 452aa638f* ]]; then
  echo "SINGLE_ABSORBED_COMMIT: pre-existing simple abandon path"
  EC_001_PATH=false
else
  echo "DIVERGENCE_CONFIRMED: HEAD=$head (additional unabsorbed commits exist)"
  EC_001_PATH=true
fi
```

**Then** because `EC_001_PATH=true` is the EXPECTED state at SPEC merge time, the orchestrator **shall** invoke `AskUserQuestion` with the 4 EC-001 disposition options (rollback-tip-then-close (Recommended) / cherry-pick-to-new-PR-then-close / abandon-new-commits / leave-open). Subagents (`manager-git`) **shall not** invoke `AskUserQuestion` directly — they execute the user-selected option only.

**And** for options 1 (rollback-tip-then-close), 2 (cherry-pick-to-new-PR-then-close), and 3 (abandon-new-commits), final verification:
```bash
gh pr view 908 --json state -q .state                         # expect: CLOSED
git ls-remote --heads origin feat/cmd-harness-slash-wrapper | wc -l   # expect: 0
```
Both checks MUST pass for options 1-3.

**And** for option 4 (leave-open), the orchestrator **shall** record the deferral rationale in the Wave 2 audit artifact and skip the verification checks above. AC PASS via deferral path.

**Binary outcome**: PASS if (options 1-3) PR state = CLOSED + branch count = 0, OR (option 4) deferral rationale persisted. FAIL if force-close attempted or option selection bypassed.
**Verification context**: Run-phase T-Wave2-001 + T-Wave2-002.
**Covers**: REQ-HRN-NS-007.

---

### AC-HRN-NS-008 (usage-log retention reference in CHANGELOG)

**Given** the run-phase T-Wave1-006 appends a v2.20.0-rc1 CHANGELOG entry for this SPEC,
**When** the orchestrator probes:
```bash
grep -A 5 'SPEC-V3R4-HARNESS-NAMESPACE-001' CHANGELOG.md | grep -iE '(retention|7.day|rolling.window|REQ-HRN-NS-010)' \
  && echo "PASS: retention reference present" || echo "FAIL: retention reference missing"
```
**Then** the output **shall** be `PASS: retention reference present`, confirming that the CHANGELOG entry references either the 7-day rolling-window retention floor OR REQ-HRN-NS-010 explicitly.

**Note**: This AC verifies that REQ-HRN-NS-010 (the optional-feature retention policy contract) is documented as a downstream-SPEC constraint in the release notes, even though no rotation mechanism ships in this SPEC.

**Binary outcome**: PASS if grep output contains `PASS:`, FAIL otherwise.
**Verification context**: Run-phase T-Wave1-006 + post-CHANGELOG probe.
**Covers**: REQ-HRN-NS-010.

---

## 3. Edge Cases

### EC-001: PR #908 HEAD divergence (CURRENT EXPECTED state)

**This is no longer an edge case — this is the EXPECTED scenario at SPEC merge time.** Plan-phase 2026-05-16 verified PR #908 HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191` (2-commit divergence: Commit 1 `452aa638f` absorbed by PR #910 commit `bb80ea0f4`; Commit 2 `a41d6d139` audit doc unabsorbed).

**Detection** (AC-HRN-NS-007 step 1 — prefix matching, not string equality):
```bash
head=$(gh pr view 908 --json headRefOid -q .headRefOid)
if [[ "$head" == 452aa638f* ]]; then
  EC_001_PATH=false   # rare — only if branch was reset to absorbed-tip
else
  EC_001_PATH=true    # expected path
fi
```

**Resolution**: `manager-git` MUST NOT force-close. The orchestrator (NOT the subagent) invokes `AskUserQuestion` with the following **4 disposition options** (ordered by recommendation):

**Option 1 (권장 / Recommended)**: rollback-tip-then-close
- `git reset --hard 452aa638f && git push --force-with-lease origin feat/cmd-harness-slash-wrapper`
- Then `gh pr close 908 --delete-branch`
- Preserves SPEC intent (absorbed commit is the only branch content); discards audit doc commit; clean closeout.

**Option 2**: cherry-pick-to-new-PR-then-close
- Cherry-pick `a41d6d139` (audit doc) onto a fresh branch off main (e.g., `docs/harness-audit-2026-05-13`)
- Open a new documentation PR for the cherry-picked commit
- Then `gh pr close 908 --delete-branch`
- Preserves audit doc value in a properly-scoped PR; cleanest preservation path.

**Option 3**: abandon-new-commits
- `gh pr close 908 --delete-branch` directly (no rollback, no cherry-pick)
- Discards audit doc commit silently; relies on git reflog for recovery if needed.
- Audit doc loss accepted by user.

**Option 4**: leave-open
- Take no action on PR #908; close out this SPEC's run-phase without resolving #908
- Defers #908 disposition to a later SPEC
- Wave 2 audit artifact records the deferral rationale; AC-HRN-NS-007 PASSes via deferral path.

**Acceptance outcome**: AC-HRN-NS-007 PASSes when user selects any of the 4 options and the orchestrator executes/records accordingly. NOT a BLOCKER — orchestrator routes through `AskUserQuestion`, user selection completes the path.

---

### EC-002: `.moai/harness/` contains non-canonical experimental sub-directory

A developer experimenting with the harness may have created `.moai/harness/experiments/` or similar non-canonical paths.

**Detection**: AC-HRN-NS-004 hierarchy probe.

**Resolution**: T-Wave1-003 is a non-destructive probe — flag the non-canonical path in the audit report, do NOT auto-remove. Document in `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-<DATE>.md` as `EXPERIMENTAL_PATH_DETECTED`. AC-HRN-NS-004 still PASSES because the probe surface ALSO documents the canonical set; non-canonical paths trigger a soft warning only.

**Note**: If the project is intended for `v2.20.0-rc1` release, the experimental paths SHOULD be cleaned up before release — but that cleanup is owned by the release process, not this SPEC.

---

### EC-003: `moai spec lint --strict` regression introduced by concurrent SPEC merge

Between plan-phase verification (Step 3 of this SPEC) and run-phase verification (T-Wave1-004), a concurrent SPEC merge (e.g., V3R4-HARNESS-002 status transition) may introduce a transient lint finding.

**Detection**: T-Wave1-004 fails with non-zero exit code.

**Resolution**:
1. Identify the offending SPEC via lint output.
2. Verify the finding is NOT caused by this SPEC's artifacts (re-run `moai spec lint --strict .moai/specs/SPEC-V3R4-HARNESS-NAMESPACE-001/` in isolation).
3. If finding is external, defer to a follow-up status-sync SPEC (do NOT expand this SPEC's scope).
4. Escalate to orchestrator via blocker report if finding cannot be deferred.

**Acceptance outcome**: AC-HRN-NS-005 is BLOCKER until external regression is resolved or deferred.

---

### EC-004: CHANGELOG merge conflict during T-Wave1-006

Concurrent v2.20.0-rc1 entries from other governance SPECs may produce merge conflicts in `CHANGELOG.md`.

**Detection**: `git merge` or rebase fails during run-phase PR creation.

**Resolution**: Standard `manager-git` conflict resolution per CLAUDE.local.md §18 — rebase on main, re-apply CHANGELOG diff, re-verify AC-HRN-NS-008. Worst case: split CHANGELOG entry into a separate sync PR.

**Acceptance outcome**: AC-HRN-NS-008 deferred to sync-phase if conflict cannot be resolved in run-phase.

---

### EC-005: Verb-surface diff shows additional verbs from future downstream SPEC

If SPEC-V3R4-HARNESS-{004..008} merges before this governance SPEC and silently introduces a fifth verb, AC-HRN-NS-003 will FAIL.

**Detection**: T-Wave2-003 diff is non-empty.

**Resolution**: Per REQ-HRN-NS-008, the offending downstream SPEC should have been rejected by plan-auditor. If it wasn't, this SPEC's sync-phase becomes BLOCKED until either (a) the downstream SPEC is reverted, or (b) REQ-HRN-NS-003 is explicitly amended via an amendment SPEC.

**Acceptance outcome**: AC-HRN-NS-003 is BLOCKER until resolution.

---

## 4. Definition of Done (DoD)

This SPEC's lifecycle is COMPLETE when ALL of the following hold:

1. Plan PR merged into main (squash, per spec-workflow.md § Step 1).
2. Run PR merged into main with all 8 ACs satisfied (binary PASS), per spec-workflow.md § Step 2.
3. PR #908 state = `CLOSED`, branch `feat/cmd-harness-slash-wrapper` deleted from origin.
4. CHANGELOG v2.20.0-rc1 entry present, references this SPEC + retention policy.
5. `moai spec lint --strict` on the entire repo passes (final regression check at sync-phase).
6. Sync PR merged into main (squash) with status transition `in-progress` → `implemented` (run-phase) → `completed` (sync-phase).
7. Memory entry `project_v3r4_harness_namespace_001_complete.md` persisted with auto-memory taxonomy compliance.
8. MEMORY.md index updated with single-line entry under 150 chars.

**Status transition timeline**:
- Plan PR open: `draft`
- Plan PR merged: `planned`
- Run PR open: `in-progress`
- Run PR merged: `implemented`
- Sync PR merged: `completed`

Per spec-workflow.md Status Lifecycle (8-value enum).

---

## 5. Out of Scope (Acceptance-Phase Reminders)

The following are explicitly NOT in any AC. Verifying any of these is a scope violation:

- Performance benchmarks of `/moai harness` verbs (no performance SLOs in governance SPEC).
- Functional testing of harness lifecycle (verbs are implementation-owned, not governance-owned).
- User-acceptance testing of the AskUserQuestion gating UX (inherited from V3R4-HARNESS-001).
- Migration of pre-existing `.moai/harness/usage-log.jsonl` schema.
- Cross-machine governance synchronization.

---

## 6. References

- spec.md §5 (REQs), §6 (Acceptance Coverage Map).
- plan.md §2 (Wave 1), §3 (Wave 2), §7 (Acceptance Criteria Summary).
- design.md — namespace hierarchy diagram + verb state machine.
- tasks.md — run-phase task list.
- SPEC-V3R4-HARNESS-001 acceptance.md — pattern precedent for binary AC.
- SPEC-V3R4-SPECLINT-DEBT-002 acceptance.md — pattern precedent for governance SPEC AC + EC sections.
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase Transitions (status lifecycle).
- `.claude/rules/moai/workflow/moai-memory.md` § Agent Memory Taxonomy.

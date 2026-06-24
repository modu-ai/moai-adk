# Acceptance Criteria — SPEC-V3R6-WORKFLOW-AGENT-DOC-ALIGN-001

> Each AC includes a verifiable grep/test command. For documentation-consistency ACs, the command MUST prove BOTH the live `.claude/` tree AND the `internal/template/templates/.claude/` mirror are remediated (where a mirror exists — dev-only `release.md`/`github.md`/`release-update.md` have no mirror; see AC-WADA-016).
>
> `ARCHIVED` below abbreviates the **11 actually-archived names** alternation (NOTE: `researcher` is DELIBERATELY EXCLUDED — it is a LIVE `role_profiles:` profile in `.moai/config/sections/workflow.yaml`, not an archived agent; purging it would break team-mode. See REQ-WADA-001a KEEP carve-out): `manager-strategy\|manager-quality\|manager-brain\|manager-project\|claude-code-guide\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring`.
>
> Directory truth: there is NO `.claude/agents/builder/`. All agent bodies (incl. `builder-harness.md`) live under `.claude/agents/moai/`.

## §D — AC Matrix

### Group 1 — Archived-agent spawn purge

**AC-WADA-001 — Zero archived spawn targets in skill files (live + template)**
Given the remediated skill assets, when grepping for archived-agent names as spawn/delegation references (excluding legitimate pointers to the rejection rule), then zero matches in both trees.
```bash
ARCHIVED11='manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring'
grep -rnE "$ARCHIVED11" .claude/skills/moai/ internal/template/templates/.claude/skills/moai/ \
  | grep -v 'archived-agent-rejection' \
  | grep -v '^[^:]*:[0-9]*:[ \t]*#'   # allow archived names only inside comment pointers to the rule
# Expected: no spawn-instruction matches (any residual line must be a documented pointer to §C)
# NOTE: researcher is DELIBERATELY ABSENT from $ARCHIVED11 — it is a live role_profile (KEEP, AC-WADA-001a).
```
Closure note: a residual line is acceptable ONLY if it is an explicit pointer to `archived-agent-rejection.md` (manual review of any match).

**AC-WADA-001a — `researcher` role_profile references PRESERVED (KEEP carve-out)**
Given the run-phase remediation, when counting `researcher` references, then the count is UNCHANGED from baseline (14) — no `researcher` role_profile reference is purged.
```bash
# researcher is a LIVE role_profile, NOT an archived agent. Baseline count must be preserved.
grep -rcn 'researcher' .claude/skills/moai/ .claude/agents/moai/ | awk -F: '{s+=$2} END {print s}'
# Expected: 14 (unchanged from plan-phase baseline — the 7 live role_profiles incl. researcher are KEPT)
# Negative check: researcher must NOT appear as an Agent() spawn target (it never did — this confirms no accidental rewrite):
grep -rn 'subagent_type.*"researcher"\|Agent(.*"researcher"' .claude/skills/moai/ .claude/agents/moai/
# Expected: 0 (researcher is only ever a role_profile name, never a subagent_type)
```
Closure note: this AC GUARDS AGAINST over-purging. A drop below 14 means a legitimate role_profile reference was wrongly removed.

**AC-WADA-002 — manager-strategy → manager-spec routing**
```bash
# No file routes planning work to manager-strategy as a spawn target (live + template)
grep -rn 'subagent_type.*manager-strategy\|Agent(.*manager-strategy\|delegate to.*manager-strategy' \
  .claude/skills/moai/ .claude/agents/moai/ internal/template/templates/.claude/
# Expected: 0
```

**AC-WADA-003 — manager-quality → sync-auditor / orchestrator batch**
```bash
grep -rn 'subagent_type.*manager-quality\|delegate to.*manager-quality\|Delegate to manager-quality' \
  .claude/skills/moai/ .claude/agents/moai/ internal/template/templates/.claude/
# Expected: 0
```

**AC-WADA-004 — expert-* → per-spawn general-purpose**
```bash
grep -rn 'subagent_type.*expert-\|Agent(.*expert-' \
  .claude/skills/moai/ .claude/agents/moai/ internal/template/templates/.claude/
# Expected: 0
# (No .claude/agents/builder/ — builder-harness.md is under .claude/agents/moai/, already covered.)
```

**AC-WADA-005 — feedback.md GitHub-issue creation routed orchestrator-direct (gh CLI)**
```bash
# feedback.md no longer delegates issue creation to manager-quality (live + template)
grep -n 'manager-quality' .claude/skills/moai/workflows/feedback.md \
  internal/template/templates/.claude/skills/moai/workflows/feedback.md
# Expected: 0
# AND the gh CLI issue-creation path is preserved:
grep -n 'gh issue create' .claude/skills/moai/workflows/feedback.md
# Expected: >= 1
```

**AC-WADA-006 — agent bodies clean of archived spawn refs**
```bash
ARCHIVED11='manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring'
grep -rnE "$ARCHIVED11" .claude/agents/moai/ \
  internal/template/templates/.claude/agents/moai/ \
  | grep -v 'archived-agent-rejection'
# Expected: no spawn-instruction matches (residuals only as §C pointers, manual review)
# (No .claude/agents/builder/ dir exists — the 6 agent bodies incl. builder-harness.md are all under .claude/agents/moai/.)
# NOTE: researcher absent from $ARCHIVED11 — live role_profile KEEP (AC-WADA-001a).
```

**AC-WADA-007 — skill frontmatter `agents:` lists clean (incl. references/ files)**
```bash
# Inspect every skill frontmatter agents: list for archived names (live + template).
# Scope explicitly includes references/{reference,anti-patterns,mx-tag}.md which carry frontmatter agents: lists.
# Match an `agents:` list line that names any of the 11 archived-proper agents:
ARCHIVED11='manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring'
grep -rnE "agents:.*($ARCHIVED11)" .claude/skills/moai/ internal/template/templates/.claude/skills/moai/
# Expected: 0  (specifically feedback.md:24 'agents: ["manager-quality"]' is replaced;
#               references/reference.md, references/anti-patterns.md, references/mx-tag.md also cleaned)
# NOTE: researcher is NOT in $ARCHIVED11 — a 'researcher' entry in an agents: list is a role_profile reference, KEEP it (AC-WADA-001a).
```

### Group 2 — Broken cross-refs & invalid dispatch

**AC-WADA-008 — release dispatch points to existing file (conditional)**
```bash
# If a release-update dispatch exists, it must be release.md; release.md must exist
grep -rn 'release-update' .claude/skills/moai/ internal/template/templates/.claude/skills/moai/
# Expected: 0 (token absent — vacuously satisfied) OR every occurrence repointed to release.md
test -f .claude/skills/moai/workflows/release.md && echo "release.md EXISTS"
```
Closure note: if grep returns 0 (plan-phase state), AC passes vacuously — no broken dispatch exists.

**AC-WADA-009 — team files use general-purpose subagent_type**
```bash
grep -rn 'team-reader\|team-validator' .claude/skills/moai/team/ \
  internal/template/templates/.claude/skills/moai/team/
# Expected: 0
# AND team files use general-purpose:
grep -rln 'subagent_type: "general-purpose"' .claude/skills/moai/team/plan.md \
  .claude/skills/moai/team/review.md
# Expected: both files match
```

**AC-WADA-010 — worktree glossary cross-reference resolves**
```bash
# Either the section exists in worktree-integration.md (live + template) ...
grep -n 'Terminology Glossary' .claude/rules/moai/workflow/worktree-integration.md \
  internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
# ... OR the 2 cross-references no longer point at a non-existent section:
grep -n 'Terminology Glossary' .claude/rules/moai/workflow/spec-workflow.md \
  .claude/rules/moai/workflow/worktree-state-guard.md
# Closure: the section named by the cross-references MUST exist in the referenced file (manual verify of resolution)
```

### Group 3 — Stale ground-truth

**AC-WADA-011 — GLM model names current**
```bash
grep -n 'glm-5.1\|glm-4.7\|glm-4.5-air\|glm-4.7-flashx' .claude/skills/moai/team/glm.md \
  internal/template/templates/.claude/skills/moai/team/glm.md
# Expected: 0 (replaced with glm-5.2[1m])
grep -n 'glm-5.2' .claude/skills/moai/team/glm.md
# Expected: >= 1
```

**AC-WADA-012 — retired terms replaced**
```bash
grep -rn 'Round split\|sprint contract' .claude/skills/moai/team/run.md \
  .claude/skills/moai/workflows/run.md \
  internal/template/templates/.claude/skills/moai/team/run.md \
  internal/template/templates/.claude/skills/moai/workflows/run.md
# Expected: 0 (replaced with Milestone / canonical harness phrasing)
```

**AC-WADA-013 — SKILL.md version footer bumped**
```bash
grep -n 'Version: 2.6.0\|Last Updated: 2026-02-25' .claude/skills/moai/SKILL.md \
  internal/template/templates/.claude/skills/moai/SKILL.md
# Expected: 0 (footer bumped to a post-consolidation version/date)
```

### Group 4 — Orphaned RED test fold-in

**AC-WADA-014 — run.md under 200 LOC (live + template)**
```bash
wc -l .claude/skills/moai/workflows/run.md \
  internal/template/templates/.claude/skills/moai/workflows/run.md
# Expected: both < 200
```

**AC-WADA-015 — TestEntryRouterLOCCeiling GREEN**
```bash
go test ./internal/skills/ -run TestEntryRouterLOCCeiling
# Expected: PASS (exit 0)
```

### Cross-cutting — Template mirror & neutrality

**AC-WADA-016 — changed-line mirror parity (bidirectional, scoped — NOT whole-file byte parity)**
This AC verifies that the LINES THIS SPEC CHANGED are present in BOTH trees. It does NOT require whole-file byte parity, because `manager-spec.md`/`manager-docs.md`/`manager-git.md` are already diverged at baseline (pre-existing lifecycle content owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001). The check is BIDIRECTIONAL — it iterates BOTH the live-side and template-side change sets so a template-only edit cannot escape detection.

```bash
# Step 1 — the set-comparison test (catches a file present in one tree but missing in the other):
go test ./internal/skills/ -run TestTemplateMirrorParity
# Expected: PASS (devOnlyLocalFiles release.md/github.md/release-update.md are excluded by the test itself)
# SCOPE NOTE (D-NEW-4): TestTemplateMirrorParity walks ONLY .claude/skills/moai/workflows/ (workflow_split_test.go:173).
#   It does NOT cover agents/moai/, skills/moai/team/, or skills/moai/references/. Content parity for those
#   directories is covered by Step 2 below (the bidirectional content-grep over skills/moai/ + agents/moai/).

# Step 2 — bidirectional changed-archived-line parity. For every file THIS SPEC touched that HAS a
#          mirror (exclude the 3 dev-only no-mirror files), confirm zero archived-proper refs survive
#          in EITHER tree. Iterating the union of both trees makes the check bidirectional:
ARCHIVED11='manager-strategy|manager-quality|manager-brain|manager-project|claude-code-guide|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring'
EXCLUDE='release.md|github.md|release-update.md'
for TREE in .claude internal/template/templates/.claude; do
  grep -rnE "$ARCHIVED11" "$TREE/skills/moai/" "$TREE/agents/moai/" 2>/dev/null \
    | grep -v 'archived-agent-rejection' \
    | grep -Ev "($EXCLUDE):" \
    | grep -v '^[^:]*:[0-9]*:[ \t]*#'
done
# Expected: no output from EITHER tree (a match in template-only would surface here — the #1 regression).
# NOTE: researcher absent from $ARCHIVED11 — live role_profile KEEP (AC-WADA-001a). Step 2 ALSO covers the
#       3 baseline-diverged agent bodies (manager-spec/docs/git.md) since it greps all of agents/moai/ in
#       both trees — so a per-file Step 3 loop is unnecessary.

# Step 3 — baseline-diverged agent bodies (manual-review closure, non-load-bearing):
#   manager-spec.md / manager-docs.md / manager-git.md carry a pre-existing live↔template §E.5/Mx
#   lifecycle divergence (owned by SPEC-V3R6-LIFECYCLE-REDESIGN-001, OUT OF SCOPE). This SPEC verifies
#   only that its archived-purge edit landed in BOTH trees — and Step 2 above already proves that
#   (zero $ARCHIVED11 in either tree's agents/moai/). No additional command needed; whole-file diff is
#   EXPECTED to differ for these 3 and is NOT a failure.
```
Closure note: whole-file `diff` PASS is NOT required for manager-spec/docs/git.md (baseline-diverged). The dev-only no-mirror trio is excluded. The bidirectional iteration (Step 2 loops BOTH trees) is the defense against a template-only residual.

**AC-WADA-017 — template neutrality preserved**
```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit
# Expected: PASS (no internal SPEC ID / REQ / audit / date / SHA leak in replacement text)
```

## §D.1 — Severity & Closure Gates

| AC | REQ | Severity | Must-pass? |
|----|-----|----------|------------|
| AC-WADA-001..007 | REQ-WADA-001..007 | MUST-FIX | yes |
| AC-WADA-001a | REQ-WADA-001a | MUST-FIX (over-purge guard — researcher KEEP) | yes |
| AC-WADA-008 | REQ-WADA-008 | SHOULD-FIX (conditional/vacuous) | yes if token present |
| AC-WADA-009..010 | REQ-WADA-009..010 | MUST-FIX | yes |
| AC-WADA-011..013 | REQ-WADA-011..013 | SHOULD-FIX | yes |
| AC-WADA-014..015 | REQ-WADA-014..015 | MUST-FIX (restores GREEN main) | yes |
| AC-WADA-016..017 | REQ-WADA-016..017 | MUST-FIX (changed-line mirror + neutrality) | yes |

## §D.2 — Definition of Done

- All MUST-FIX ACs pass with command evidence captured in `progress.md §E.2`/`§E.3`.
- `go test ./...` GREEN (including `TestEntryRouterLOCCeiling` AND `TestTemplateMirrorParity`).
- `golangci-lint run` reports no new issues (no Go production change expected).
- For every in-scope MIRRORED file, this SPEC's changed lines are present in BOTH trees (changed-line mirror per AC-WADA-016 — NOT whole-file byte parity; the 3 baseline-diverged agent bodies and the dev-only no-mirror trio are excluded per their carve-outs).
- `make build` regenerates embedded templates without diff surprises.
- No archived-agent name (of the 11 archived-proper) survives as a spawn target anywhere in template-managed `/moai` assets; the 14 `researcher` role_profile references are PRESERVED (AC-WADA-001a).

## §D.3 — Edge Cases

1. A file references an archived agent ONLY as a pointer to `archived-agent-rejection.md` — KEEP (legitimate). Verified by `grep -v 'archived-agent-rejection'` + manual review.
2. `claude-code-guide` appears referencing the Anthropic built-in (NOT the archived MoAI file) — KEEP (the built-in is valid). Disambiguate by surrounding context per `archived-agent-rejection.md §C.3`. (Live occurrence count is 0, so this is a forward-guard only.)
3. `release-update` token absent at run-phase — AC-WADA-008 passes vacuously (no broken dispatch).
4. `run.md` archived-purge edit (M2) reduces LOC but not below 200 — M6 trim still required; coordinate so trim is the final edit.
5. `researcher` appears in a workflow.yaml-driven `role_profiles` context OR an `agents: [...]` list OR a Team-Mode table — KEEP (live role_profile, NOT an archived agent — AC-WADA-001a). Over-purging it below the baseline count of 14 is a defect.
6. A file is dev-only (`release.md`/`github.md`/`release-update.md`) — its archived-purge applies to the LIVE file only; no template mirror exists, so it is excluded from AC-WADA-016 and from `TestTemplateMirrorParity`.
7. `manager-spec.md`/`manager-docs.md`/`manager-git.md` show a whole-file `diff` against their template mirror — EXPECTED (pre-existing §E.5/Mx lifecycle divergence, out of scope). Only this SPEC's archived-purge edit must be mirrored, not the whole file.

## §E — Plan/Run/Sync Evidence Sections

(Canonical §E lifecycle section skeleton — populated by the respective phase owners. See `progress.md`.)

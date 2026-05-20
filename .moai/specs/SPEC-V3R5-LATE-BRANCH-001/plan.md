# SPEC-V3R5-LATE-BRANCH-001 — Implementation Plan

> Plan-phase artifact. Companion to `spec.md` (WHAT/WHY) and `acceptance.md` (HOW VERIFIED). This document covers HOW BUILT (implementation strategy, milestone breakdown, verification commands, risk mitigation).

## 1. Implementation Strategy

### 1.1 High-level approach

Late-branch is a **configuration + documentation** change, not a code change. There is no new Go package, no new CLI verb, no new test infrastructure required. The implementation is delivered as:

1. One YAML configuration change (4 keys in `team` section of `git-strategy.yaml`)
2. One conditional branch in `spec-assembly.md` Phase 3 skill body + Phase 2.5 GitHub Issue creation MUST default to skip; gated only by explicit `--issue` flag check (added v0.1.1 per REQ-LB-009)
3. `--issue` flag semantics flip in `.claude/skills/moai/SKILL.md` from "opt-out via `--no-issue`" to "opt-in via `--issue`" — default behavior changes from create to skip (added v0.1.1)
4. One row + one new subsection in `manager-git.md` agent body
5. One precondition update + one cleanup procedure update in `spec-workflow.md` rule
6. 5 template mirror updates (`internal/template/templates/`) to keep `moai update` outputs aligned

Total LOC delta estimate: ~90-170 lines across 10 files (5 local + 5 template). No Go code, no new tests required at the binary level — verification is via grep/yq commands on the modified files (see §4 Verification Plan).

**Affected files** (canonical list for run-phase navigation):

| Local source | Template mirror |
|---|---|
| `.moai/config/sections/git-strategy.yaml` | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` |
| `.claude/skills/moai/SKILL.md` (NEW v0.1.1) | `internal/template/templates/.claude/skills/moai/SKILL.md` (NEW v0.1.1) |
| `.claude/agents/moai/manager-git.md` | `internal/template/templates/.claude/agents/moai/manager-git.md` |
| `.claude/rules/moai/workflow/spec-workflow.md` | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` |

### 1.2 Why this is a configuration-only SPEC

The branch creation decision point already exists in `spec-assembly.md` Phase 3 — currently it always creates a branch. The change is to add a conditional branch (`if branch_creation.auto_enabled == false: skip`) and propagate the same default through `git-strategy.yaml`. No new logic is invented; existing config-driven branching gains a new value.

`manager-git.md` and `spec-workflow.md` updates are pure documentation — they teach the orchestrator and downstream agents to recognize and operate within the Late-branch pattern. The 4-phase procedure (A→D) is already documented in `feedback_late_branch_workflow.md` memory; the SPEC formalizes it into the agent body.

### 1.3 Execution surface

| Layer | Modification | LOC est. | Verification |
|-------|--------------|----------|--------------|
| Config (yaml) | 4 keys → false/true flip | ~6 | `yq` query returns expected value |
| Skill (md) | Phase 3 conditional + display | ~25 | `grep` for "auto_enabled" + "Late-branch" |
| Agent (md) | Table row + 4-phase subsection | ~40 | `grep` for "main_late_branch" + "Late-Branch Invocation Pattern" |
| Rule (md) | Step 1 precondition + Step 4 cleanup | ~20 | `grep` for "main checkout" + "git reset --hard origin/main" |
| Template mirrors | byte-equivalent of 4 above | ~80 (sum) | `diff` after macro expansion returns 0 lines |

---

## 2. D-Decisions

This SPEC has one critical design decision documented here. Subsequent D-Decisions may emerge during run-phase and will be appended.

### D1 — Backward Compatibility for Template Default

**Question**: Should the template default for `team.branch_creation.auto_enabled` change from `true` to `false`, and how should existing user projects (`moai update`) inherit (or not) this change?

**Options Evaluated**:

#### Option (a) Breaking default change (REJECTED for this SPEC)
- Template `.tmpl` default: `team.branch_creation.auto_enabled: false`
- New projects created via `moai init` get Late-branch by default
- Existing projects via `moai update` retain their current value (no auto-rewrite)
- **Pros**: New projects align with sole-maintainer-friendly default; no migration burden.
- **Cons**: Surprises new team-mode users who expect `feat/SPEC-*` auto-branch; documentation must explicitly call out the change in v3.5.0 release notes; risk of confused new contributors thinking branch creation is broken.

#### Option (b) Opt-in via `workflow` value (REJECTED for this SPEC)
- Template default unchanged: `team.branch_creation.auto_enabled: true`
- Add a new top-level value `workflow: late-branch` (sibling of current `github-flow`) that activates Late-branch behavior regardless of `auto_enabled`
- **Pros**: Explicit opt-in, no behavior surprise for existing users; clear name (`late-branch`) communicates intent.
- **Cons**: Two settings to coordinate (`workflow` + `auto_enabled`) with potential conflict (what if `workflow: late-branch` but `auto_enabled: true`?); more config surface area; precedence rules need defining.

#### Option (c) Interview-driven on `moai init` (REJECTED for this SPEC)
- Template default unchanged: `auto_enabled: true`
- `moai init` interview asks "Late-branch vs auto-branch" and writes the user choice into `git-strategy.yaml`
- **Pros**: User-aware, no defaults assumption; aligns with existing init-time interview UX.
- **Cons**: Requires `moai init` interview wiring (deferred — EXCL-LB-001 prohibits migration tooling in this SPEC); `moai update` path still ambiguous (existing users unaffected by interview).

**Selected Option: (a) Breaking default change**

**Rationale**: The MoAI-ADK project is a sole-maintainer (GOOS) tool whose primary user base is also sole-maintainer or small-team developers (per `CLAUDE.local.md` §22 dev settings intent and the user directive 2026-05-20). The auto-branch pattern is heavyweight for this audience and provides little marginal value when CI gates run on the eventual PR regardless. Concrete consequences:

1. **New projects** (`moai init` after v3.5.0) start with Late-branch as the working default. This matches the intended user base. v3.5.0 release notes must explicitly call this out as a breaking default change with one-line migration guidance (`auto_enabled: true` in `git-strategy.yaml` restores prior behavior).

2. **Existing projects** (`moai update` after v3.5.0) are unaffected because `moai update` does not silently overwrite user-modified config sections (per `CLAUDE.local.md` §2 Protected Directories rule). The user's existing `git-strategy.yaml` is preserved verbatim; only template files under `internal/template/templates/` change. Users opting into Late-branch on existing projects can manually edit their `git-strategy.yaml`.

3. **Documentation surface**: Option (a) avoids the precedence-rule complexity of (b) and the interview-wiring scope of (c). The change is documented in one place (`git-strategy.yaml.tmpl` comment + `manager-git.md` Personal Mode table + v3.5.0 release notes).

**Risk acknowledged**: New team-mode users with expectations of feat/SPEC-* auto-branch may be initially confused. Mitigation: explicit `prompt_always: true` paired with `auto_enabled: false` ensures the user is informed once at first SPEC creation (R-LB-003).

**Future revisit**: If usage telemetry or feedback indicates the breaking default is a frequent friction point, Option (b)'s `workflow: late-branch` sibling value can be added in a follow-on SPEC without conflict with (a) — Option (b) layered on top of (a) is a strict superset.

### D2 — Frontmatter `issue_number` Field Removal (added v0.1.1)

**Question**: Should new SPECs carry an `issue_number` frontmatter field at all, given the policy decision (2026-05-20) to disable automatic GitHub Issue creation? How should existing SPECs with populated `issue_number` values be handled?

**Decision**: Remove the `issue_number` field entirely from the spec.md frontmatter for new SPECs from v0.1.1 onward. Existing SPECs that already carry `issue_number: <N>` retain the field as immutable history records — NO migration is performed.

**Rationale**:

1. **Field is meaningful only when issue exists**: When `/moai plan` does not auto-create a GitHub Issue (REQ-LB-009), the `issue_number` field is always `null`/`TBD` for new SPECs. Carrying an always-null field clutters the frontmatter and forces lint rules to special-case the null value.

2. **`spec-frontmatter-schema.md` already lists `issue_number` as Optional**: The canonical schema (`.claude/rules/moai/development/spec-frontmatter-schema.md`) classifies `issue_number` as an optional field (alongside `depends_on`, `lint.skip`, `bc_id`). Omitting it for new SPECs is consistent with the schema.

3. **History preservation (EXCL-LB-008)**: Existing SPECs (e.g., SPEC-V3R5-CONSTITUTION-DUAL-001 with merged PRs #1015/#1016/#1017) carry `issue_number` values that serve as historical PR cross-references. Retroactively stripping these values would lose archival context. No migration tool will be built (EXCL-LB-008).

4. **Distinction from D1**:
   - **D1** = behavior change (auto-branch creation default flips from `true` to `false`)
   - **D2** = metadata simplification (no `issue_number` field on new SPECs)
   - Both share motivation: **sole-maintainer workflow optimization** (sole maintainer = SPEC author = implementer; external Issue gateway provides no marginal visibility value).

5. **Opt-in re-introduction**: A user invoking `/moai plan --issue` for an externally-tracked SPEC (e.g., bug report from community) MAY add `issue_number: <N>` manually after `gh issue create`. The field is not forbidden — only the AUTOMATIC creation is suppressed.

**Implementation surface**:
- `manager-spec` agent prompt template: remove `issue_number: TBD` from generated frontmatter
- `spec-assembly.md` Phase 2.5: silent skip default (covered by REQ-LB-009 + AC-LB-007)
- `.claude/skills/moai/SKILL.md`: `--issue` flag flip from `--no-issue` opt-out to `--issue` opt-in
- This SPEC's own spec.md frontmatter: `issue_number: TBD` line removed at v0.1.1

**Risk acknowledged**: PR templates with `closes #N` references may break for SPECs lacking `issue_number`. Mitigation: see R-LB-005 (PR template `closes #N` lines remain optional).

---

## 3. Implementation Sequence (M1-M5)

Priority-based milestone breakdown. No time estimates per coding-standards Rule. Sequential within a single run-phase using `manager-develop` (cycle_type=ddd per `quality.yaml` development_mode), since modifications are to existing files (PRESERVE-IMPROVE pattern).

### M1 — Config switch (D1 source) — Priority P0

**Deliverable**: `.moai/config/sections/git-strategy.yaml` and `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` updated.

**Specific edits** (`team` section, lines 50-74 of git-strategy.yaml):
```yaml
team:
    automation:
        auto_branch: false      # was: true
        auto_commit: true       # unchanged
        auto_pr: false          # was: true
        auto_push: true         # unchanged
    branch_creation:
        auto_enabled: false     # was: true
        prompt_always: true     # was: false
    # remaining fields unchanged
```

Template mirror (`git-strategy.yaml.tmpl`) receives the identical 4-key change. Inline comment added explaining "Late-branch default per SPEC-V3R5-LATE-BRANCH-001" to preserve discoverability of the breaking change.

**Verification**: §4 AC-LB-001.

### M2 — Skill body Phase 3 conditional (D2) — Priority P0

**Deliverable**: `.claude/skills/moai/workflows/plan/spec-assembly.md` and template mirror updated.

**Specific edits** (line 253-340 region, Phase 3: Git Environment Setup):

1. Add a new sub-clause at the top of Phase 3 logic:
   > "If `team.branch_creation.auto_enabled == false`, skip branch creation. Cwd remains on the current branch (typically `main`). Emit mode display 'Late-branch (main commit + late switch)' to communicate the deferred-branch state to the user. SPEC files are committed to the current branch via the standard commit pipeline."

2. Surface "Late-branch (main commit + late switch)" as a recognized mode in Phase 3.5 mode-selection display alongside existing modes.

3. Phase 3.0 BODP Gate (entry point handling) remains active — Late-branch does not bypass BODP. BODP for Late-branch entry point uses `EntryPlanLateBranch` (NEW), parallel to existing `EntryPlanWorktree`/`EntryPlanBranch`.

**Verification**: §4 AC-LB-002.

### M3 — Agent body Personal Mode + Late-Branch Invocation Pattern (D3) — Priority P0

**Deliverable**: `.claude/agents/moai/manager-git.md` and template mirror updated.

**Specific edits**:

1. Add row to Personal Mode SPEC Git Workflow options table:

   | Workflow option | Branch strategy | PR strategy | Cleanup |
   |---|---|---|---|
   | `main_late_branch` | main commit + late `git switch -c feat/SPEC-*` | PR squash + delete-branch | local main `reset --hard origin/main` |

2. Add new subsection "Late-Branch Invocation Pattern" under Branch Management:
   - Phase A — SPEC creation on main (cwd unchanged after `/moai plan`)
   - Phase B — Implementation commits accumulate on main (no push)
   - Phase C — At PR time: `git switch -c feat/SPEC-*` → push → `gh pr create` → squash merge
   - Phase D — Local main reset: `git reset --hard origin/main` → `git pull origin main` (verify)
   - Detection cue: `git rev-list main..HEAD --count > 0 && current branch matches feat/SPEC-*` after manual `git switch -c`
   - Caveat: `git push origin main` is BLOCKED in Phase A/B (REQ-LB-007); even with `auto_push: true`, the orchestrator must hold push until Phase C branch creation.

**Verification**: §4 AC-LB-003.

### M4 — Rule update (D4) — Priority P0

**Deliverable**: `.claude/rules/moai/workflow/spec-workflow.md` and template mirror updated.

**Specific edits**:

1. Step 1 (Plan) entry precondition section: add line:
   > "When `team.branch_creation.auto_enabled == false`, the entry precondition is `git rev-parse --abbrev-ref HEAD == main` (or the user's chosen base branch if `main_branch` differs). No `plan/SPEC-*` branch is created at this step; plan-phase commits land directly on main and are pushed only after Phase C `git switch -c plan/SPEC-*` at PR time."

2. Step 4 (Cleanup) cleanup procedure section: add Late-branch closure subsection:
   > "**Late-branch closure** (when `auto_enabled == false`): After squash merge of run-PR and sync-PR, the user (or `manager-git` automation) MUST execute:
   > ```bash
   > git checkout main
   > git fetch origin
   > git reset --hard origin/main
   > git pull origin main   # verify
   > ```
   > Post-condition: `git status --porcelain` returns empty AND `git rev-parse main` == `git rev-parse origin/main`.
   > Failure mode: skipping this step leaves local main with un-squashed history that conflicts with the next `git pull` (R-LB-002)."

3. Cross-reference added: "For the 4-phase Late-branch invocation pattern (A→D), see `manager-git.md` § Late-Branch Invocation Pattern and `feedback_late_branch_workflow.md`."

**Verification**: §4 AC-LB-004.

### M5 — Template mirror parity (D5) — Priority P0

**Deliverable**: All 4 changes from M1-M4 byte-equivalently mirrored to `internal/template/templates/` counterparts.

**Specific edits**: Identical content as M1-M4 source files, with only template variable substitutions (`{{.GoBinPath}}`, `{{.HomeDir}}`, etc. — currently none of the affected sections use template variables, so equivalence is literal byte-equality after diff).

**Verification**: §4 AC-LB-005. Existing `internal/template/rule_template_mirror_test.go` is expected to PASS on these mirrored files without modification, OR (if mismatch detected) the test extension noted in REQ-LB-008 is folded into this milestone.

---

## 4. Verification Plan

Every AC in `acceptance.md` is executable as a single shell command. The full command set runs as a smoke-suite during run-phase REFACTOR step and again at sync-phase pre-PR check.

### AC-LB-001 — Config switch verification
```bash
# Expect: false, false, false, true
yq '.git_strategy.team.automation.auto_branch' .moai/config/sections/git-strategy.yaml
yq '.git_strategy.team.automation.auto_pr' .moai/config/sections/git-strategy.yaml
yq '.git_strategy.team.branch_creation.auto_enabled' .moai/config/sections/git-strategy.yaml
yq '.git_strategy.team.branch_creation.prompt_always' .moai/config/sections/git-strategy.yaml
```

### AC-LB-002 — Skill body Phase 3 conditional verification
```bash
# Expect: ≥1 (must contain auto_enabled conditional)
grep -c "auto_enabled" .claude/skills/moai/workflows/plan/spec-assembly.md
# Expect: ≥1 (Late-branch mode display)
grep -c "Late-branch" .claude/skills/moai/workflows/plan/spec-assembly.md
```

### AC-LB-003 — Agent body Personal Mode + Late-Branch Invocation Pattern verification
```bash
# Expect: ≥1 (Personal Mode table row)
grep -c "main_late_branch" .claude/agents/moai/manager-git.md
# Expect: ≥1 (subsection header)
grep -c "Late-Branch Invocation Pattern" .claude/agents/moai/manager-git.md
# Expect: ≥4 (Phase A/B/C/D explicit)
grep -cE "^(Phase A|Phase B|Phase C|Phase D)" .claude/agents/moai/manager-git.md
```

### AC-LB-004 — Rule update verification
```bash
# Expect: ≥1 (Step 1 main checkout precondition + Step 4 cleanup procedure)
grep -c "main checkout" .claude/rules/moai/workflow/spec-workflow.md
# Expect: ≥1 (canonical Late-branch closure step)
grep -c "git reset --hard origin/main" .claude/rules/moai/workflow/spec-workflow.md
```

### AC-LB-005 — Template mirror parity verification
```bash
# Expect: 0 (no drift between local and template)
# .yaml.tmpl uses Go template variables — diff with envsubst-equivalent expansion
for f in \
  ".moai/config/sections/git-strategy.yaml:internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl" \
  ".claude/skills/moai/workflows/plan/spec-assembly.md:internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md" \
  ".claude/agents/moai/manager-git.md:internal/template/templates/.claude/agents/moai/manager-git.md" \
  ".claude/rules/moai/workflow/spec-workflow.md:internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md"; do
  src="${f%%:*}"; tgt="${f##*:}"
  # For .md files: literal diff (no template variables in affected sections expected)
  if [[ "$src" == *.md ]]; then
    diff -u "$src" "$tgt" | wc -l   # Expect: 0
  fi
done
# Existing CI: go test ./internal/template/... must PASS
go test ./internal/template/ -run TestRuleTemplateMirror -v
```

### AC-LB-006 — E2E Late-branch scenario verification (manual or scripted)
```bash
# Scripted scenario — runs in /tmp test repo
set -e
TESTDIR=$(mktemp -d)
cd "$TESTDIR"
git init -b main
git commit --allow-empty -m "initial"
git remote add origin .   # self-remote for test
git push origin main 2>/dev/null || git update-ref refs/remotes/origin/main HEAD

# Phase A: SPEC creation simulation
echo "spec content" > spec.md
git add . && git commit -m "spec(SPEC-TEST-001): initial plan"

# Phase B: Implementation commits
echo "red" > impl.go && git add . && git commit -m "RED"
echo "green" > impl.go && git add . && git commit -m "GREEN"

# Phase C: late branch + push + (simulate squash merge with rebase)
git switch -c feat/SPEC-TEST-001
git checkout main
git reset --hard $(git merge-base main feat/SPEC-TEST-001) || true
# Simulate squash merge on origin
git checkout feat/SPEC-TEST-001
git checkout main
git merge --squash feat/SPEC-TEST-001
git commit -m "feat: SPEC-TEST-001 (squash)"
git update-ref refs/remotes/origin/main HEAD

# Phase D: local main reset
git reset --hard refs/remotes/origin/main
test "$(git status --porcelain)" = "" && echo "AC-LB-006 PASS" || echo "AC-LB-006 FAIL"
```

A simplified non-scripted check during run-phase: dogfooding the pattern on this very SPEC's PR — if the maintainer can complete Phase A through D on the plan-PR + run-PR + sync-PR cycle without error, AC-LB-006 is empirically satisfied.

---

## 5. Risk Mitigation

### R-LB-001 — Accidental `git push origin main` during Phase A/B
**Risk**: User runs `git push` or `git push origin` (which defaults to current branch) during Phase B, attempting to push uncommitted-yet-PR-bound work to main. Branch protection blocks the push, but the error message may confuse the user.

**Mitigation**:
1. REQ-LB-007 mandates that automated paths NEVER push during Phase A/B.
2. `manager-git.md` Late-Branch Invocation Pattern subsection explicitly documents this caveat.
3. **Defense-in-depth** (out of scope, separate SPEC): pre-push hook that blocks direct push to `main` regardless of branch protection — recommended follow-on SPEC.

**Detection**: If branch protection rejects the push, the user observes a `gh` or `git` error. The orchestrator (next `/moai` invocation) detects via `git log origin/main..main --count > 0` that local main is ahead and surfaces the Phase C `git switch -c` recommendation.

### R-LB-002 — Skipping Phase D `git reset --hard origin/main`
**Risk**: User forgets to run Phase D after squash merge. Next `git pull` produces merge conflict against squashed remote because local main has the un-squashed history.

**Mitigation**:
1. REQ-LB-006 + Step 4 documentation in `spec-workflow.md`.
2. `manager-git.md` Late-Branch Invocation Pattern includes Phase D verification commands.
3. `feedback_late_branch_workflow.md` memory documents this caveat for orchestrator recall.

**Detection**: `git pull` failure or `git log main..origin/main --count > 0 && git log origin/main..main --count > 0` both non-zero (divergence). The orchestrator catches this on next `/moai` invocation and recommends Phase D.

### R-LB-003 — Backward compatibility for existing users
**Risk**: Existing users with `mode: team` and `auto_branch: true` upgrade to v3.5.0 expecting unchanged behavior. The breaking default in Option (a) might confuse them.

**Mitigation**:
1. `moai update` does NOT auto-rewrite user `git-strategy.yaml` (D-Decision §2 rationale point 2).
2. v3.5.0 release notes call out the breaking default change with one-line migration guidance.
3. `prompt_always: true` (paired with `auto_enabled: false`) ensures first SPEC creation surfaces the choice to the user, providing a teachable moment.

**Detection**: User feedback through `/moai feedback` or GitHub issues post-v3.5.0 release.

### R-LB-004 — Template mirror drift
**Risk**: Future modifications to `git-strategy.yaml.tmpl` (or the other 3 template files) diverge from local copies, breaking `moai update` parity.

**Mitigation**:
1. AC-LB-005 enforced as a one-shot verification at sync-PR.
2. Existing `internal/template/rule_template_mirror_test.go` runs on every `go test` (per `CLAUDE.local.md` Template-First rule).
3. If the existing test does not cover the new fields, REQ-LB-008 extends the test suite during run-phase.

**Detection**: CI failure on `go test ./internal/template/...` post-merge of any future modification.

### R-LB-005 — PR template `closes #N` references break for SPECs lacking `issue_number` (added v0.1.1)
**Risk**: Existing PR templates and CI workflows may reference `closes #N` lines tied to the SPEC's `issue_number` frontmatter. New SPECs (post-v0.1.1) lack this field per D2, so `closes #N` substitution either fails silently or produces malformed PR bodies.

**Mitigation**:
1. PR template `closes #N` lines remain **optional** — SPECs without `issue_number` simply omit the close reference in the PR body.
2. No tooling change required: `manager-git` PR-body generation already treats `issue_number` as nullable. The generation logic falls through gracefully when the field is absent.
3. For SPECs explicitly tracked via `/moai plan --issue` (opt-in path per REQ-LB-009), the user manually adds `issue_number: <N>` to frontmatter after `gh issue create`, restoring the `closes #N` substitution.
4. v3.5.0 release notes document the policy change alongside D1, with one-line guidance: "If your workflow depends on `closes #N` in PR bodies, use `/moai plan --issue` to retain Issue creation."

**Detection**: PR body inspection at sync-phase — if `closes #` placeholder remains unsubstituted, surface warning to user. No CI guard required; the issue is cosmetic (Issue does not exist to be closed).

---

## 6. References

- **Memory files**: `project_v3r5_late_branch_decision.md`, `feedback_late_branch_workflow.md`
- **Source files (with line anchors)**:
  - `.moai/config/sections/git-strategy.yaml` (lines 50-74 `team` section)
  - `.claude/skills/moai/workflows/plan/spec-assembly.md` (lines 250-340 Phase 3 region)
  - `.claude/agents/moai/manager-git.md` (Personal Mode SPEC Git Workflow section)
  - `.claude/rules/moai/workflow/spec-workflow.md` (Step 1 entry, Step 4 cleanup)
- **Template mirrors**: parallel paths under `internal/template/templates/`
- **Related SPECs**: SPEC-V3R5-HARNESS-AUTONOMY-001 (W3 reference structure), SPEC-V3R4-WORKTREE (worktree boundary), SPEC-V3R2-WF-001 (workflow phase ordering FROZEN)
- **Related rules**: `.claude/rules/moai/development/branch-origin-protocol.md` (BODP signals), `.claude/rules/moai/workflow/worktree-integration.md` (terminology glossary)
- **Coding standards**: §22 `CLAUDE.local.md` Dev Settings Intent, §15 Template Language Neutrality

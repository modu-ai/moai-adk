# SPEC-SYNC-STRATEGY-KEY-001 — Acceptance Criteria

All grep paths are repo-root-relative. RED-now baselines were measured on tree `d29b8942e` (worktree `t303`, branch `WT-sync-strategy-key`); each criterion states why it is red now (verification-completeness §2 two-cell discipline). Green-path cells name the milestone in `plan.md` that flips them.

Template tree shorthand: `internal/template/templates` = `T`. Local tree shorthand: repo root = `L`.

---

## §D AC Matrix

### AC-SYK-001 — Canonical key read (REQ-SYK-001, REQ-SYK-011)

- **Given** the shipped sync delivery workflow, **When** Step 3.0 (Detect Git Workflow Strategy) executes, **Then** it reads `git_strategy.{mode}.workflow` from `git-strategy.yaml` (mode resolved from `git_strategy.mode`) and names no other primary source.
- RED-now: `grep -n 'git_strategy' T/.claude/skills/moai/workflows/sync/delivery.md` matches only the `main_branch` resolution (L218-220); the Step 3.0 instruction reads `github.spec_git_workflow` (L17). Red because the canonical read does not exist.
- Green path: M1 rewrites Step 3.0; `grep -c 'git_strategy.{mode}.workflow' T/...delivery.md` ≥ 1 AND `grep -n 'Read `github.spec_git_workflow`' T/...delivery.md` = 0.

### AC-SYK-002 — Legacy key removed from shipped surfaces (REQ-SYK-001, REQ-SYK-009)

- **Given** the template tree after implementation, **When** `grep -rn 'spec_git_workflow' T/` runs, **Then** the count is 0 (the deprecation fallback in AC-SYK-003 mentions the legacy key by its yaml path `github.spec_git_workflow` at most inside the fallback block; if that blocks this grep, the sentinel `spec_git_workflow` must appear only in the fallback and the AC grep is refined to exclude the deprecation block — state the refinement in the evidence, do not weaken the removal claim).
- RED-now: count = **10** (delivery.md ×5, doc-execution.md ×1, tab_schema.json ×3, system.yaml.tmpl ×1).
- Green path: M1-M3 remove all ten; `grep -rn 'spec_git_workflow' T/ | wc -l` → 0 (or refined-exclusion count with the fallback block cited).

### AC-SYK-003 — Legacy fallback with deprecation warning (REQ-SYK-002)

- **Given** a config where `git_strategy.{mode}.workflow` is absent and `github.spec_git_workflow: main_direct` is present, **When** Step 3.0 executes, **Then** the mapped direct-push strategy is applied AND a deprecation warning naming `git_strategy.{mode}.workflow` and the removal version (v3.3.0) is emitted.
- RED-now: `grep -c 'deprecat' T/...delivery.md` = 0 — no fallback or warning exists.
- Green path: M1 adds the D1 fallback table; `grep -c 'v3.3.0' T/...delivery.md` ≥ 1 AND the D1 mapping table tokens (`main_direct`, `feature_branch`) appear only inside the fallback block.

### AC-SYK-004 — Unmatched value stops (REQ-SYK-003, REQ-SYK-004)

- **Given** `git_strategy.manual.workflow: gitflow` (out-of-domain token), **When** Step 3.0 resolves the strategy, **Then** delivery stops with a report naming the offending value and the canonical domain `{github-flow, git-flow}`; no PR is created and nothing is pushed.
- RED-now: no domain statement or value-stop exists in `T/...delivery.md` (`grep -c 'github-flow' T/...delivery.md` = 0).
- Green path: M1 adds the domain sentence + stop; grep for the domain enumeration ≥ 1 in Step 3.0.

### AC-SYK-005 — WT-* route: no PR, integration-worktree merge (REQ-SYK-005) [card regression (a)]

- **Given** the `git-flow` strategy active and the current branch matching `WT-*`, **When** Step 3.2 executes, **Then** no `gh pr create` is invoked for the branch, and the procedure directs: coordinate the merge window with the coordinating session → enter the designated develop integration worktree → `git merge --no-ff <branch>` → push the integration branch → exit; the route text contains no force-push and no PR creation.
- RED-now: `grep -c 'WT-' T/...delivery.md` = **0** — the route does not exist, so a WT-* sync today falls through and improvises.
- Green path: M1 adds the route; `grep -c 'WT-' T/...delivery.md` ≥ 1 AND `sed -n` on the WT-* block shows zero `gh pr` tokens.

### AC-SYK-006 — Unmatched branch stops, no fall-through (REQ-SYK-006) [card regression (b), anti-mutant pair with AC-SYK-005]

- **Given** the `git-flow` strategy active and the current branch matching none of `WT-*`, `feature/*`, `release/*`, `hotfix/*`, `develop`, `main` (e.g. `zzz-typo-branch`), **When** Step 3.2 executes, **Then** delivery stops and reports the branch name with the defined route list; no PR is created and nothing is pushed.
- **Anti-mutant contract**: adding the WT-* route WITHOUT removing fall-through must fail this AC (the `zzz-*` branch still falls through), and removing fall-through WITHOUT the WT-* route must fail AC-SYK-005 (WT-* becomes a stop instead of the defined integration path). The two ACs are independently falsifiable: AC-SYK-005 is red whenever the WT route is absent; AC-SYK-006 is red whenever any silent default remains.
- RED-now: `grep -n 'matches no\|no defined route\|unmatched' T/...delivery.md` = **0 hits** — the stop clause does not exist.
- Green path (mechanized, per-block attribution — audit iter-2 D3): every `##### Strategy:` block in `T/...delivery.md` ends with the stop clause. Mechanical check:
  ```bash
  awk '/^##### Strategy:/{s=$0} /matches no defined route/{c[s]++} END{for(k in c) print k, c[k]}' \
    internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
  ```
  → exactly 2 lines emitted, one per strategy block (`github-flow`, `git-flow`).
- Green path (negative probe, AP-6 mutant must fail — audit iter-2 D3): the rewritten Step 3.0/3.2 contains NO surviving default/catch-all route language:
  ```bash
  grep -n 'Default strategy\|Otherwise,' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
  ```
  → **0 hits** (red now: 1 hit — the `Default strategy (if not configured): github_flow` line at L25). A document carrying BOTH the stop clause AND a reworded default line fails this probe while passing the sentinel check — the two greps are jointly required.

### AC-SYK-007 — Single-source procedure ownership (REQ-SYK-007)

- **Given** the WT-* merge-window procedure exists in two candidate homes (shipped `delivery.md`, dev-only `gitflow-lane-protocol.md`), **When** the change lands, **Then** the canonical procedure (enter → merge --no-ff → push develop → exit) appears in full exactly in `T/...delivery.md` + its local mirror, and `.claude/rules/local/gitflow-lane-protocol.md` references `delivery.md` Step 3.2 instead of restating it.
- RED-now: the procedure lives ONLY in `gitflow-lane-protocol.md` (§2); `grep -c 'WT-' T/...delivery.md` = 0; the dev rule contains the full command sequence.
- Green path: M4 amends the dev rule; `grep -c 'delivery.md' .claude/rules/local/gitflow-lane-protocol.md` ≥ 1 AND a signature phrase from the procedure (e.g. `merge --no-ff`) appears in exactly one of the two rule/skill homes per tree (template carries it; dev rule cites, does not restate the sequence as its own canonical text).

### AC-SYK-008 — Template neutrality preserved (REQ-SYK-008)

- **Given** the template tree after implementation, **When** neutrality probes run, **Then**: `grep -rn 'gitflow-lane-protocol' T/` = 0; `grep -n 'workflow:' T/.moai/config/sections/git-strategy.yaml.tmpl` still yields only `github-flow` values (no `git-flow`/`gitflow` default leaked); AND the **diff-scoped SPEC-token probe** (widened iter-2, audit D2) returns zero:
  ```bash
  git diff d29b8942e -- internal/template/templates/ | grep '^+' | grep -oE 'SPEC-[A-Z0-9-]+-[0-9]{3}'
  ```
  → **0 matches** — no canonical-shape SPEC-ID from ANY SPEC (not only this SPEC's own ID) may appear on added lines of the template diff. This is what catches the mirror-direction mutant of AC-SYK-012 (a foreign ID such as `SPEC-SYNC-PARALLEL-DOCS-001` leaking into the shipped template would pass the old self-ID-only probe and read GREEN while violating REQ-SYK-008).
- RED-now: probes pass on the untouched surfaces (0 / github-flow-only; diff probe 0 on an empty change) — this AC is a **preserve** criterion whose red form is a leakage introduced BY this SPEC's edits (including the mirror-sync step); it must be re-measured post-M4. A tree-wide SPEC-token zero-assertion is NOT valid here — 152 canonical-shape tokens pre-exist in the template tree as documentation examples (measured on `d29b8942e`) — which is precisely why the probe is scoped to this change's added lines.
- Green path: M1-M4 keep all three probes at their current values.

### AC-SYK-009 — Inventory parity (REQ-SYK-009)

- **Given** `github.spec_git_workflow` is removed from `system.yaml.tmpl`, **When** the inventory parity test runs, **Then** `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` exits 0 AND `grep -c 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml` = 0.
- RED-now (for the row): the row exists at `shipped_key_inventory.yaml:2505`, class R. The Go test itself currently passes; it fails only if a template key lacks a row — so the ordering constraint is: remove the tmpl key and the inventory row in the SAME milestone (M3).
- Green path: M3 removes both; test exit 0 observed verbatim.

### AC-SYK-010 — Interview schema rebound (REQ-SYK-010)

- **Given** the project interview schema, **When** the change lands, **Then** no field binds to `github.spec_git_workflow` (`grep -c 'github.spec_git_workflow' T/.claude/skills/moai-workflow-project/schemas/tab_schema.json` = 0) and the SPEC-branching question is rebound to `git_strategy.{mode}.automation.auto_branch` with boolean options (grep for the field ≥ 1).
- RED-now: 3 hits at L1006/1008/1029 binding the dead key; no `automation.auto_branch` field exists.
- Green path: M2 rebinds the question.

### AC-SYK-011 — doc-execution reads canonical keys (REQ-SYK-011)

- **Given** sync document execution startup, **When** Step 1.2 reads project configuration, **Then** it names `git_strategy.mode` and `git_strategy.{mode}.workflow` and does not name the legacy key.
- RED-now: `L/.claude/skills/moai/workflows/sync/doc-execution.md:25` reads "git_strategy.mode, conversation_language, spec_git_workflow".
- Green path: M2 rewrites the line in template; M4 syncs the local mirror.

### AC-SYK-012 — Local surfaces synced at token parity + local value in-domain (REQ-SYK-008) [rescoped iter-2, audit D2]

- **Given** `make build` regenerated the embedded templates and the SPEC-scoped edits landed on both trees, **When** the local surfaces are checked, **Then**:
  1. **Token parity on SPEC-edited regions** (NOT byte-identity, fallback-sentinel excluded): the refined probe `grep -rn 'spec_git_workflow' .claude/skills/ .moai/config/sections/system.yaml | grep -v 'Legacy key fallback' | wc -l` = 0 on the local tree — the dead-key tokens are gone from template AND local alike (the greps of AC-SYK-002 cover the template side, and the exclusion is the same one AC-SYK-002 applies there). The **unrefined** raw grep is expected to return exactly **1**, and that single hit MUST be the D1 fallback sentinel in `L/.claude/skills/moai/workflows/sync/delivery.md` — which AC-SYK-003 REQUIRES to exist, so its presence is the criterion holding, not weakening. A raw count other than 1, or a raw hit outside that fallback block, FAILS this sub-criterion: a NEW non-sentinel `spec_git_workflow` reference anywhere under `.claude/skills/` or in `.moai/config/sections/system.yaml` is a removal-claim regression, not an accepted exception.
  2. **Byte-identical mirror** where the pair is genuine: `diff T/.claude/skills/moai-workflow-project/schemas/tab_schema.json L/.claude/skills/moai-workflow-project/schemas/tab_schema.json` = empty.
  3. **Preserved local-only differences intact** (the destructive-mutant guard): local `doc-execution.md` still carries the `SPEC-SYNC-PARALLEL-DOCS-001 A5` attribution block (`grep -c 'SPEC-SYNC-PARALLEL-DOCS-001' L/.claude/skills/moai/workflows/sync/doc-execution.md` ≥ 1); the delivery.md footer drift is likewise a preserved difference, not a sync target.
  4. **Local value in-domain**: `grep -n 'workflow:' L/.moai/config/sections/git-strategy.yaml` shows `manual.workflow: git-flow` (was `gitflow`).
- RED-now (correct reason, corrected iter-2): the parity criterion is red because the local tree still carries **10 dead-key token hits** (template also 10 — both trees carry the tokens equally; the pre-existing byte-diffs of delivery.md/doc-execution.md are unrelated drift and are NOT the red cause, nor a sync target); local workflow value is `gitflow` (out-of-domain).
- Green path: M4 performs `make build`, region-scoped mirror sync (see plan M4.2 exclusions), and the local value fix. Mirror-sync MUST NOT overwrite local files wholesale: the enforced criterion is (1)-(4) above, not diff-empty — a blanket `cp` template→local would delete the A5 block and FAIL sub-criterion 3 even while making diffs "clean".

---

## §D.1 Severity

- **MUST-PASS**: AC-SYK-001, 002, 004, 005, 006, 008, 009, 010, 012 (key unification, both card regressions, neutrality, parity, mirrors).
- **SHOULD**: AC-SYK-003 (fallback window — degradation to hard removal is a documented downgrade, not a silent one), AC-SYK-007 (single-source — a tolerated temporary duplication must be recorded as debt), AC-SYK-011 (auxiliary read path).

## §D.2 Traceability

REQ-SYK-001 → AC-001, AC-002 · REQ-SYK-002 → AC-003 · REQ-SYK-003 → AC-004 · REQ-SYK-004 → AC-004 · REQ-SYK-005 → AC-005 · REQ-SYK-006 → AC-006 · REQ-SYK-007 → AC-007 · REQ-SYK-008 → AC-008, AC-012 · REQ-SYK-009 → AC-009 · REQ-SYK-010 → AC-010 · REQ-SYK-011 → AC-011.

## §D.3 Edge cases

1. Config has BOTH keys with conflicting values → canonical wins; legacy ignored without warning (the fallback fires only on canonical-absent).
2. `git_strategy.mode` names a profile whose `workflow` subkey is missing (e.g. stripped `manual` block) → treated as unmatched-value stop, not github-flow default; the old silent default `github_flow` (delivery.md **L25**, corrected iter-2) is removed with the dead axis.
3. Branch simultaneously matching patterns (a branch literally named `WT-release/x`) → first-match order: `WT-*` checked before `release/*`; the route list order is normative.
4. WT-* sync while the develop integration worktree does not exist → stop-and-report (the route names the precondition; provisioning the worktree is the coordinating session's act, out of scope).
5. Legacy value `per_spec` / `develop_direct` (the third valuespace) → explicit stop with migration hint (no mapping defined).

## §D.4 Closure gates

- All MUST-PASS ACs green with verbatim command + output evidence (5-section format per verification-claim-integrity §3).
- `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` exit 0 on the final tree.
- Template neutrality CI guard (`.github/workflows/template-neutrality-check.yaml`) green on the PR head.

## §D.5 Forward-looking checks

- After v3.3.0 removes the D1 fallback: re-run AC-SYK-002/003 greps to confirm the fallback block (and its legacy tokens) are gone.
- After v3.3.0 removes the D1 fallback sentinel (card **t315** owns that removal): AC-SYK-012.1's refined form stays correct as written (raw count becomes 0, refined count stays 0), but its stated raw-count expectation MUST be updated from exactly 1 to 0, and AC-SYK-003 itself becomes obsolete at that point.
- The t298 integration-lock defect (holders always appear reclaimable) is the known-broken serialization underneath the WT-* merge window; when its fix lands, re-check the WT-* route text still names the coordinating-session notification as the serialization mechanism.

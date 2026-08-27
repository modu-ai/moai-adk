---
id: SPEC-SYNC-STRATEGY-KEY-001
title: "Unify git delivery strategy on git_strategy.<mode>.workflow — retire github.spec_git_workflow, clean the value axis, route WT-* lanes, stop on unmatched branches"
version: "0.2.0"
status: implemented
created: 2026-08-27
updated: 2026-08-27
author: manager-spec (card t303)
priority: P1
phase: "v3.2.0"
module: ".claude/skills/moai/workflows/sync, internal/template/templates, .moai/config/sections, internal/config/testdata"
lifecycle: spec-anchored
tags: "git-strategy, sync-delivery, config-key-unification, value-domain, wt-branch-routing, template-first, deprecation"
tier: M
related_specs:
  - SPEC-V3R5-GIT-STRATEGY-SCHEMA-001
---

# SPEC-SYNC-STRATEGY-KEY-001 — Unify the git delivery strategy key, value axis, and branch routing

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-27 | manager-spec (card t303) | Initial plan-phase emission. Tier M (lead-decided, not re-classified). 11 GEARS REQs / 12 binary ACs. RED-now baselines measured on tree `d29b8942e` (this worktree, branch `WT-sync-strategy-key`). Lead-approved decisions 2026-08-27 treated as fixed inputs; three delegated design calls resolved here: D1 deprecation-with-window, D2 canonical value domain `{github-flow, git-flow}`, D3 shipped-template-owns-the-WT-procedure. |
| 0.2.0 | 2026-08-27 | manager-spec (card t303) | Plan-audit iteration-2 revision per verdict `.moai/reports/t303/plan-audit-SPEC-SYNC-STRATEGY-KEY-001-v0.1.0.md` (FAIL 0.86, blocking D2). Fixes: audit-D2 — §4 mirror-pair premise corrected (only `tab_schema.json` is a byte-identical pair; `delivery.md`/`doc-execution.md` are neutralized mirrors with preserved local-only content); audit-D1 — tab_schema option line cites L351/357→L354/359. Companion edits in acceptance.md (AC-006/008/012 rescope), plan.md (M1/M4.2/§F), progress.md (plan_status). Audit D6/D7 routed to orchestrator. |

---

## 1. Problem — measured shape

The sync-phase delivery routing reads a **dead config key** and dispatches on a **polluted value axis**, and its gitflow strategy has **no route for the branch shape this repository's lanes actually run on**.

### 1.1 The dead key

`github.spec_git_workflow` (`.moai/config/sections/system.yaml:44`) has **zero Go runtime consumers**. `internal/config/loader_system.go` loads the system section but binds only `hook.*` keys; the `github` block is intentionally unbound. The live chain is `git_strategy` — `internal/config/types.go:20` binding, `types.go:143` struct, `ActiveModeProfile()` accessor, `defaults.go:674`, `Loader.Load()` section registration, plus nested/loader test suites. `delivery.md` L218 (Base Branch Resolution) already reads `git_strategy.mode` from `git-strategy.yaml` — the same skill file mixes both sources today.

### 1.2 The polluted value axis (worse than the card brief stated)

The pre-plan investigation found the strategy axis carries **four incompatible spellings of the same concepts across five surfaces**, plus a third valuespace in the interview schema:

| Surface | Tokens in play | Axis |
|---|---|---|
| `delivery.md` L224-283 strategy dispatch | `github_flow`, `main_direct`, `gitflow` | strategy (underscored / fused) |
| `delivery.md` L27-29 branch handling | `feature_branch`, `main_direct` | branch handling wrongly on strategy key |
| Go defaults (`defaults.go:682,694,707`) + template `git-strategy.yaml.tmpl:13,45,81` | `github-flow` | workflow (hyphenated) |
| `tab_schema.json` interview options for `git_strategy.manual.workflow` (value lines L354/L359) | `github-flow`, `git-flow` | workflow (hyphenated) |
| `tab_schema.json` interview field bound to the dead key (L1006/1008/1029) | `develop_direct`, `feature_branch`, `per_spec` | a THIRD valuespace matching neither of the above |
| This repo's local `git-strategy.yaml:8` | `gitflow` (under `manual.workflow`) | out-of-domain even for the schema's own option set |

So the canonical key's own value domain is already mechanically settled in the wired chain (Go defaults + template + interview schema): **hyphenated `github-flow` / `git-flow`**. The underscored tokens live only in the dead-key consumers.

### 1.3 The missing route and the silent fall-through

Lanes run on `WT-<slug>` branches (observed `WT-t114`, `WT-t60`, `WT-lane-board`). The gitflow strategy in `delivery.md` Step 3.2 defines routes for `feature/*`, `release/*`, `hotfix/*`, `develop`, and `main` (error) — **none match `WT-*`**, and any branch matching no route falls through silently, leaving the agent to improvise PR creation. The merge-window procedure that SHOULD govern `WT-*` integration lives today only in `.claude/rules/local/gitflow-lane-protocol.md` (dev-only, NOT template-mirrored).

### 1.4 RED-now baselines (tree `d29b8942e`)

| Probe | Command (in repo root) | Observed now |
|---|---|---|
| Dead-key references in template tree | `grep -rn 'spec_git_workflow' internal/template/templates/ \| wc -l` | **10** (delivery.md ×5, doc-execution.md ×1, tab_schema.json ×3, system.yaml.tmpl ×1) |
| Dead-key references in local mirrors | `grep -rn 'spec_git_workflow' .claude/skills .moai/config/sections/system.yaml \| wc -l` | **10** (same shape + local system.yaml:44) |
| WT- route in template delivery.md | `grep -c 'WT-' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` | **0** |
| Unmatched-branch stop sentinel | `grep -n 'matches no\|no defined route\|unmatched' …/delivery.md` | **0 hits** |
| Canonical-key read at Step 3.0 | grep for `git_strategy.{mode}.workflow` read instruction | **0** (only `{mode}.main_branch` is read, L220) |
| Inventory row for the dead key | `internal/config/testdata/shipped_key_inventory.yaml:2505` | present, class R |

---

## 2. Requirements (GEARS)

Canonical read path: resolve `mode` from `git_strategy.mode`, then read `git_strategy.{mode}.workflow` from `.moai/config/sections/git-strategy.yaml`.

- **REQ-SYK-001** (Ubiquitous) — The sync delivery workflow shall select the git delivery strategy exclusively by reading `git_strategy.{mode}.workflow` from `git-strategy.yaml`, resolving `{mode}` from `git_strategy.mode`, and shall not read `github.spec_git_workflow` as a primary source.
- **REQ-SYK-002** (When legacy-key-detected) — When `git_strategy.{mode}.workflow` is absent and the legacy `github.spec_git_workflow` key is present in `system.yaml`, the sync delivery workflow shall apply the mapped strategy per the D1 migration table and shall emit a deprecation warning naming the canonical key.
- **REQ-SYK-003** (Ubiquitous) — The `git_strategy.{mode}.workflow` value domain shall be exactly `{github-flow, git-flow}`; tokens from the legacy axes (`feature_branch`, `main_direct`, `github_flow`, `gitflow`, `develop_direct`, `per_spec`) shall not be accepted as strategy selectors on this key.
- **REQ-SYK-004** (When unmatched-value-detected) — When the resolved workflow value matches no canonical token, the delivery step shall stop and report the offending value with the canonical domain, shall not create a pull request, and shall not push.
- **REQ-SYK-005** (While git-flow active, When WT-* branch) — While the resolved strategy is `git-flow`, when the current branch name matches `WT-*`, the delivery step shall create no pull request and shall direct the merge through the designated develop integration worktree per the single-source procedure (enter the integration worktree, merge `--no-ff`, push the integration branch, exit), coordinating the merge window with the coordinating session before entering.
- **REQ-SYK-006** (When unmatched-branch-detected) — When the current branch matches none of the routes defined for the active strategy, the delivery step shall stop and report the branch name together with the defined routes, shall not create a pull request, and shall not push — the delivery step shall not fall through to an improvised default route.
- **REQ-SYK-007** (Ubiquitous) — The WT-* merge-window procedure shall be stated exactly once, owned by the shipped `delivery.md` (template + local mirror), written user-generically; repository-local operating rules shall reference that owner rather than restate the procedure.
- **REQ-SYK-008** (Ubiquitous) — All shipped-surface edits shall land first in `internal/template/templates/**` followed by `make build`, and local mirrors under `.claude/` shall be synced from the regenerated template; the distributed template shall carry no project-private values (no `git-flow`/`gitflow` default on the workflow key, no `.claude/rules/local/*` references, no SPEC-ID tokens).
- **REQ-SYK-009** (Ubiquitous) — The distribution template shall remove the `github.spec_git_workflow` key from `system.yaml.tmpl`, and `internal/config/testdata/shipped_key_inventory.yaml` shall drop the corresponding row in the same change so shipped keys and the triage inventory remain in parity.
- **REQ-SYK-010** (Ubiquitous) — The project interview schema shall bind no field to `github.spec_git_workflow`; the SPEC-branching question shall be rebound to the branch-handling axis live key `git_strategy.{mode}.automation.auto_branch` with boolean options.
- **REQ-SYK-011** (When sync starts) — When the sync workflow begins document execution, the pre-delivery step shall read `git_strategy.mode` and `git_strategy.{mode}.workflow` as the configuration basis and shall not reference the legacy key.

---

## 3. Design decisions (delegated calls, resolved here)

### D1 — Deprecation, not silent removal (existing user configs)

Chosen: **deprecation with a one-minor-version fallback window** (honored through `v3.2.x`, removed in `v3.3.0`).

- Template `system.yaml.tmpl`: the key is REMOVED now (fresh installs and `moai update` redeployments stop carrying it — §2.3 of CLAUDE.local.md wipes `.moai/config` wholesale on update, so the key disappears from updated installs immediately).
- Skill text (`delivery.md` Step 3.0): a legacy fallback path is retained for stale configs (hand-pinned yaml, partial updates). Fallback fires only when the canonical key is absent AND the legacy key is present, applies the mapping below, and emits a deprecation warning naming `git_strategy.{mode}.workflow`.

| Legacy value | Mapped behavior | Rationale |
|---|---|---|
| `main_direct` | direct push to base branch, no PR | preserves the old template default's semantics |
| `github_flow` | `github-flow` route | spelling normalization only |
| `gitflow` | `git-flow` route | spelling normalization only |
| `feature_branch` | `github-flow` route (branch handling already owned by `automation.auto_branch` / `branch_creation.*`) | axis correction: the token never belonged on the strategy axis |
| `develop_direct`, `per_spec`, anything else | explicit stop with migration hint | no safe automatic mapping |

Why not hard removal: `moai update` wipes and re-lays `.moai/config` from the template, but users with pinned or hand-edited configs keep the old key with no signal; a fallback that maps + warns converts a silent wrong-strategy behavior into a visible, self-healing one. Why not permanent support: the key has zero Go consumers; the window only exists to bridge stale files, and a dated removal prevents it becoming a second permanent source.

### D2 — Canonical value domain `{github-flow, git-flow}`

The canonical key's domain follows the **already-wired chain** (Go defaults `defaults.go:682/694/707`, template `git-strategy.yaml.tmpl:13/45/81`, interview schema options) rather than inventing tokens or adopting the underscored dead-key spellings. Consequences:

- `delivery.md` Step 3.2 strategy headings become `github-flow` and `git-flow`; the `main_direct` strategy block is REMOVED as a workflow value (its no-PR direct-push behavior remains reachable through the Tier Route A/B gate at L33-36, which already owns commit/delivery shape per `spec-workflow.md`, and through the D1 legacy mapping for stale configs).
- This repo's local `git-strategy.yaml:8` value `gitflow` becomes `git-flow` (local config fix, not a template value). Known non-durability: `moai update` resets local `.moai/config` to template defaults, so the local value must be re-applied after updates — a pre-existing condition already documented in `gitflow-lane-protocol.md`'s cross-references, not introduced here.
- The old template default transition (dead key `main_direct` → live key `github-flow`) changes the out-of-the-box delivery for fresh installs from "no PR" to "PR". This is accepted: the template's own git-strategy default has been `github-flow` all along; the dead key and the live key never agreed, and unification necessarily picks the live one.

### D3 — Single-source owner for the WT-* merge-window procedure

Chosen: **the shipped template owns the canonical procedure; the dev-only rule points back to it.**

- `delivery.md` (template + local mirror) carries the complete, user-generic WT-* route: no PR; coordinate the merge window with the coordinating session; enter the designated develop integration worktree; `git merge --no-ff <branch>`; push the integration branch (`git push origin develop`); never force; on push rejection, fetch + integrate + retry.
- `.claude/rules/local/gitflow-lane-protocol.md` is amended to reference `delivery.md` Step 3.2 as the canonical statement and keep only the repo-local specifics (integration-lock commands, lead-notification serialization, the t298 lock-defect caveat).

Why this direction and not the reverse: a shipped file referencing a non-shipped file is a neutrality violation — the reference dangles for every downstream user (`.claude/rules/local/*` does not exist outside this repository), and CLAUDE.local.md §2.1 forbids project-private references in the template. A non-shipped file referencing a shipped one dangles nowhere. The `WT-` prefix itself is already a shipped convention (kanban-dispatch worktree-branch naming), so naming it in the template carries no private value.

---

## 4. Constraints

- Plan phase authored documents only; no code changes, no `make build`, no commits in this phase.
- Scope discipline: `delivery.md`'s Tier Route A/B gate (L33-36) stays as-is; no redesign of unrelated delivery.md sections.
- Template content neutrality (CLAUDE.local.md §2.1) binds everything landing under `internal/template/templates/**`: no SPEC-ID tokens, no internal dates, no macOS-bias, no project-private values. This repo's private `git-flow` choice must not leak into the template as a VALUE (structural support for a user-configurable workflow ships; the private choice does not).
- Mirror-scope taxonomy for the four edited surfaces (corrected in v0.2.0; the earlier "four byte-consistent mirror pairs" premise was false):
  - **Byte-identical pair** (diff-empty enforced): `tab_schema.json` only.
  - **Neutralized mirrors** (template carries the neutralized form; local intentionally diverges): `delivery.md` (footer drift, 462 vs 463 lines at audit time) and `doc-execution.md` (local-only 6-line block attributing the FO-SYNC-4 concurrency design to `SPEC-SYNC-PARALLEL-DOCS-001 A5`, including the audit-concurrency scheduling sentence the template deliberately omits). For these, mirror-sync enforces **token-level parity on the SPEC-edited regions only** (the dead-key greps of AC-SYK-002); the preserved local-only differences are NEVER overwritten by a template copy — copying template over local would delete the A5 attribution and the scheduling instruction (a behavioral regression), and copying local into template would leak SPEC-ID tokens into a shipped surface.
  - **Config pair** (never byte-identical by design): `system.yaml` (local, carries repo-private values) vs `system.yaml.tmpl` (template default). Only the SPEC-scoped key removal lands on both.
- Neutrality probe scope: template-wide `SPEC-` token zero-assertions are false by baseline (152 pre-existing canonical-shape tokens in `internal/template/templates/` — doc examples); the enforced probe is **diff-scoped** — zero canonical-shape `SPEC-[A-Z0-9-]+-[0-9]{3}` tokens on ADDED lines of this change's template diff (AC-SYK-008).

---

## 5. Out of Scope

### Out of Scope — delivery.md Tier Route gate

- The Tier S/M/L Route A/B commit-routing gate at delivery.md L33-36 (per `spec-workflow.md` § SPEC Phase Discipline) is not modified beyond removing the `main_direct` strategy block's dependence on the dead axis.

### Out of Scope — Go runtime wiring of the workflow key

- No new Go consumer is added for `git_strategy.{mode}.workflow`; the key remains skill-body + interview-schema consumed (disposition unchanged from SPEC-V3R5-GIT-STRATEGY-SCHEMA-001's wire-through matrix for this field).

### Out of Scope — non-sync consumers of branch handling

- `plan`/`run`-phase branch-creation behavior (`automation.auto_branch`, `branch_creation.*`, Late-Branch closure) is not modified; only the interview question that surfaces the setting is rebound.

### Out of Scope — local-config durability across moai update

- The `moai update` wholesale wipe of local `.moai/config` values (CLAUDE.local.md §2.3) is a known pre-existing defect family with its own cards; this SPEC only re-applies the local `git-flow` value and documents the non-durability.

### Out of Scope — release harness and main-branch policy

- Release branching, `release/vX.Y.Z` handling, and main-branch protection policy (`repo-local-pr-policy.md`) are untouched; the WT-* route integrates to `develop` only.

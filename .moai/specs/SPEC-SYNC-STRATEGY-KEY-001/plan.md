# SPEC-SYNC-STRATEGY-KEY-001 — Implementation Plan

Tier M. Milestones are ordered by decision-reversibility: the design-bearing edits (value domain, routing semantics) come first so human review concentrates where change-likelihood is highest; mechanical mirror-sync work is deferred to the bottom.

## §A Context

- Work location: this lane's worktree (branch `WT-sync-strategy-key`, base develop tip `d29b8942e`).
- SPEC artifacts: `.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/{spec,plan,acceptance,progress}.md`.
- Surfaces (template + local mirror pairs):
  - `.claude/skills/moai/workflows/sync/delivery.md` (463 lines; Step 3.0 L17-29, Step 3.1 Route gate L33-36 — PRESERVE, Step 3.2 L215-283, Completion Criteria L416)
  - `.claude/skills/moai/workflows/sync/doc-execution.md` (L25 config read)
  - `.claude/skills/moai-workflow-project/schemas/tab_schema.json` (L1006/1008/1029 dead-key binding; L345-365 canonical workflow field — reference for the domain)
  - `.moai/config/sections/system.yaml` (+ `.tmpl` L31)
- Non-pair surfaces: `internal/config/testdata/shipped_key_inventory.yaml:2505`, `.claude/rules/local/gitflow-lane-protocol.md`, `.moai/config/sections/git-strategy.yaml:8`.
- Lead-approved decisions (2026-08-27): canonical key, value-axis cleanup, WT-* lane, explicit stop, template-first. Delegated calls resolved in spec.md §3 (D1 deprecation window, D2 domain `{github-flow, git-flow}`, D3 shipped-template owns the WT procedure).

## §B Known Issues

- B1. `moai update` wipes local `.moai/config` wholesale (CLAUDE.local.md §2.3): the local `manual.workflow: git-flow` fix is non-durable and must be re-applied after updates. The skill's canonical-read path must NOT depend on any local value.
- B2. Primary-vs-develop drift: primary's `git-strategy.yaml` says `manual.workflow: github-flow`; develop-based trees say `gitflow`. Sessions in the primary checkout read stale config — the explicit-stop on out-of-domain values (REQ-SYK-004) converts this from silent wrong-strategy to visible stop.
- B3. AC-SYK-002 grep self-collision: the D1 fallback block must mention the legacy key, so the removal grep needs the documented refinement (acceptance.md AC-SYK-002) — do not "fix" this by deleting the fallback.
- B4. Inventory ordering: removing the tmpl key without the inventory row in the same change leaves stale testdata (hygiene); removing the row without the key fails `TestShippedConfigKeysHaveReaders`. Same-milestone only.
- B5. Cross-reference rot: `CLAUDE.local.md` §4.1 and `gitflow-lane-protocol.md` cross-reference `git_strategy.manual.workflow: gitflow`; after the local value becomes `git-flow`, both lines must be updated (CLAUDE.local.md is user-owned — surface to the operator, do not edit silently).
- B6. Doc-only scope: no Go production code changes; the only Go-adjacent touch is testdata (`shipped_key_inventory.yaml`). Lane-local verification is the targeted config test + grep battery — NOT `go test ./...` (load discipline).

## §C Pre-flight (run before M1; record output verbatim)

```bash
git branch --show-current && git rev-parse --short HEAD
grep -rn 'spec_git_workflow' internal/template/templates/ | wc -l   # expect 10
grep -c 'WT-' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md  # expect 0
go test ./internal/config/ -run TestShippedConfigKeysHaveReaders    # expect exit 0 (baseline)
```

## §D Constraints

- PRESERVE: delivery.md Step 3.1 Tier Route A/B gate (L33-36); `git-strategy.yaml.tmpl` workflow defaults (`github-flow`); all template neutrality invariants (no SPEC-IDs, no dates, no `.claude/rules/local/*` refs, no private values).
- FORBIDDEN: direct edits to local `.claude/skills/**` before the template edit (template-first); `git push --force`; editing `.moai/state/`, `.moai/harness/`; touching unrelated delivery.md sections.
- Language: skill bodies in English (distributed artifact); `gitflow-lane-protocol.md` stays Korean (local-only file, match its existing style).

## §E Self-Verification

AC matrix per acceptance.md §D with the (a) command / (b) verbatim output / (c) tree-SHA attribution triple for every row; the full grep battery re-run on the final tree; neutrality probes (AC-SYK-008) explicitly re-measured post-edit since their risk is introduced by this change.

## §F Milestones

> **Milestone-frame delta (operator-visible, iter-2 per audit D4)**: the lead-approved card sketch was 3 milestones (M1 key unification / M2 WT-* route / M3 explicit stop). This plan restructures to M1-M5: the approved M1-M3 content all lands in **this plan's M1** because key read, value domain, WT-* route, and the unmatched-stop are one continuous rewrite of delivery.md Step 3.0/3.2 in a single file — splitting one file edit across milestones creates merge-conflict surface with no isolation benefit. The approved sketch's coverage is complete (key unification → M1.1-1.4 + M2-M3; WT-* route → M1.5; explicit stop → M1.6); M2-M5 are additions (aux consumers, atomic inventory pair, mechanical build/sync, verification pass) ordered by decision-reversibility per the plan-phase ordering rule. If the lead prefers the original 3-milestone shape at kickoff, M1 splits back along items 1-4 / 5 / 6 without content change.

### M1 — Template delivery.md: canonical read, domain, routing semantics (High — design-bearing, most likely to change)

1. Step 3.0: read `git_strategy.{mode}.workflow`; state the domain `{github-flow, git-flow}`; add unmatched-value stop (REQ-SYK-001/003/004).
2. Step 3.0: add the D1 legacy fallback table + deprecation warning (window: removed in v3.3.0) (REQ-SYK-002).
3. Remove the L25 silent `github_flow` default (`Default strategy (if not configured)`, corrected iter-2 per audit D1) and the L27-29 branch-handling-on-strategy-key block (value-axis cleanup).
4. Step 3.2: rename strategies to `github-flow` / `git-flow`; REMOVE the `main_direct` strategy block (semantics live on the Route gate + D1 mapping).
5. Step 3.2 git-flow: add the `WT-*` route (checked FIRST) — no PR; coordinate merge window with the coordinating session; enter the designated develop integration worktree; `git merge --no-ff`; push integration branch; never force (REQ-SYK-005, D3).
6. Step 3.2: end every strategy's route list with the explicit unmatched-branch stop (REQ-SYK-006).
7. Step 3.3.5 and Completion Criteria L416: align vocabulary with the canonical tokens.

### M2 — Template auxiliary consumers (Medium)

1. `doc-execution.md:25`: canonical config read (REQ-SYK-011).
2. `tab_schema.json`: rebind the SPEC-workflow question to `git_strategy.{mode}.automation.auto_branch`, boolean options matching the neighboring `auto_delete_branches` question pattern; drop the dead valuespace (`develop_direct`/`feature_branch`/`per_spec`) (REQ-SYK-010).

### M3 — Template key removal + inventory parity (Medium; atomic)

1. Remove `spec_git_workflow` (+ its comment) from `system.yaml.tmpl` (REQ-SYK-009).
2. Remove the `github.spec_git_workflow` row from `shipped_key_inventory.yaml` — same commit as (1).
3. Verify: `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` exit 0.

### M4 — Build, mirror sync, local surfaces (Low — mechanical)

1. `make build` (regenerate embedded templates).
2. Sync local mirrors — **region-scoped, never a wholesale copy** (rescoped iter-2 per audit D2):
   - `tab_schema.json`: byte-identical pair — full sync, `diff` empty enforced.
   - `delivery.md`, `doc-execution.md`: apply ONLY this SPEC's region edits to the local copies (Step 3.0 / Step 3.2 / L25 default removal / doc-execution L25 config-read line). EXCLUDE from the copy: the doc-execution.md local-only `SPEC-SYNC-PARALLEL-DOCS-001 A5` attribution block (with its audit-concurrency scheduling sentence) and the delivery.md footer drift — both are preserved local-only differences (spec.md §4 taxonomy). A blanket `cp` template→local is FORBIDDEN here (deletes the A5 block = behavioral regression); the mirror-direction copy is likewise forbidden (leaks SPEC-ID tokens — caught by AC-SYK-008's diff-scoped probe). Verification is AC-SYK-012's token-parity sub-criteria, not diff-empty.
3. Local config: `git-strategy.yaml` `manual.workflow: gitflow` → `git-flow`; remove `github.spec_git_workflow` from local `system.yaml`.
4. Amend `gitflow-lane-protocol.md`: reference `delivery.md` Step 3.2 as the canonical WT-* procedure; keep repo-local specifics (lock commands, t298 caveat); update its `gitflow` cross-reference line (REQ-SYK-007).
5. Surface B5 to the operator: `CLAUDE.local.md` §4.1 cross-ref line update (user-owned file).

### M5 — Verification pass (Low)

Full acceptance battery on the final tree; evidence persisted under `.moai/state/verify/`; AC matrix emitted with attribution triples.

## §G Anti-Patterns

- AP-1: editing local mirrors first (breaks template-first; `moai update` would revert the local edits and resurrect the dead key).
- AP-2: keeping the legacy fallback forever "to be safe" — the dated v3.3.0 removal is part of the design; an undated fallback is a second permanent source.
- AP-3: writing the WT-* procedure into the dev-only rule and having the template point at it — dangling reference for every downstream user (neutrality violation; D3 chose the opposite direction).
- AP-4: hardcoding this repo's `git-flow` choice as a template value.
- AP-5: adding the WT-* route while leaving any catch-all/else route — defeats REQ-SYK-006 and fails AC-SYK-006 (the named mutant).
- AP-6: "simplifying" the explicit stop into a default-to-github-flow — reintroduces silent fall-through with different wording.

## §H Cross-References

- `spec.md` §3 D1-D3 (the delegated design calls this plan executes)
- `acceptance.md` §D (AC matrix; anti-mutant contract in AC-SYK-006)
- SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (git_strategy Go-struct disposition; this SPEC changes no disposition)
- CLAUDE.local.md §2 (Template-First Rule), §2.1 (neutrality), §4.1 (develop integration model)
- `.claude/rules/local/gitflow-lane-protocol.md` (repo-local merge-window specifics; becomes a pointer per D3)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (Route A/B gate — preserved)

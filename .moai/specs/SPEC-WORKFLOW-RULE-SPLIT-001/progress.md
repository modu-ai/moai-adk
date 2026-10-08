# Progress — SPEC-WORKFLOW-RULE-SPLIT-001

card t1586, lane-12, run tree `.moai/worktrees/spec-workflow-split`, branch `WT-spec-workflow-split`.
Baseline `81786284e` (original spec-workflow.md unmodified in tree at extraction time — verified via `git status` + `cmp` against its template mirror before assembly).

## §E.2 Run-phase Evidence

### M1 — split (commits 524c9c190, 4ddf0e871)

**Verbatim-move battery (byte-level, stronger than the AC-07 signature greps).** Original extracted from the template mirror (`internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`, `cmp`-identical to the pre-edit local file). Command: python substring assertion of every plan §B M1 kept range (L1-29, L48-61, L68-199, L218-227, L240-332, L334-371, L389-398, L413-446) inside the new core, and every moved range (L30-47, L62-67, L201-217, L229-239, L374-388, L401-412) inside the new companion, plus the single modified skip-contract line (orig L333) and the two promoted headings. Observed: `FAILS: NONE`; companion binding-tag scan found 3 `[HARD]`/`[ZONE]` literals, all inside the NEW header/intro reference sentences (the `factory-dispatch-detail.md` precedent carries the same literal in its blockquote); all 5 original tagged rule lines verified 1-hit each in the new core (`grep -c` = 1 for each of: `[ZONE:Frozen] [HARD] Every MoAI SPEC follows` / `[ZONE:Frozen] [HARD] Step ordering rules` / `[ZONE:Evolvable] [HARD] Every SPEC plan-phase classifies` / `[ZONE:Frozen] [HARD] Execute in main checkout` / `Plan Audit Gate skip policy (single authoritative contract)`). Unaccounted original lines after the battery: 373, 400 only — the blank lines absorbed by the two promoted headings. No content line lost.

**Assembly-defect record:** the first assembly pass misplaced the route-table pointer (after the step-ordering block instead of before it) and omitted the anti-patterns pointer — caught by the core pointer-count check (7 observed vs 8 expected), source restored byte-identical from the template mirror (`cmp` → RESTORED_PAIR_OK), script corrected, re-run; the shipped state is the verified one.

### AC matrix (all commands run in this tree, this run)

| AC | Command (verbatim) | Observed output (verbatim, deciding lines) | Verdict |
|---|---|---|---|
| AC-01 | `python3 -c "print(len(open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"` | `31141` | PASS (≤ 39999) |
| AC-01 | `python3 -c "print(len(open('internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"` | `31141` | PASS |
| AC-01 | `wc -m .claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | `31141` / `31141` / `62282 total` (cross-check agrees) | PASS |
| AC-01 | `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run TestInstructionsLoaded -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/hook	1.135s` (exit 0) | PASS (regression leg; fixture-based, not a live measurement) |
| AC-01 (extra) | companion rune counts, both trees | local `12007`, template `12007` | PASS (companion is paths-scoped; no per-file budget breach) |
| AC-02 | `cmp` pair 1 + pair 2 | `PAIR1_OK` + `PAIR2_OK` (exit 0 both) | PASS |
| AC-03 | `unset … && go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestLateBranchTemplateMirror\|TestDeclaredRuleMirrorForks' -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/template	0.359s` (AC03_EXIT=0) | PASS |
| AC-03 | `grep -c "spec-workflow-detail" internal/template/rule_template_mirror_test.go` | `2` (≥ 1) | PASS |
| AC-04 | `grep -F -c` × 3 clauses + 2 heading greps + skip-contract alternation | `1` `1` `1` `1` `1` `1` (each ≥ 1, headings exactly 1) | PASS |
| AC-04 | `make constitution-check` | `constitution-check: OK` (exit 0; bin/moai registry validation via `MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md ./bin/moai constitution list --format json`) | PASS (clause-text axis, per plan §H) |
| AC-05 | `unset … && go test ./internal/spec/ -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/spec	480.838s` (exit 0; TierArtifact* family included) | PASS |
| AC-05 | `grep -c '"Artifact set"\|Artifact set' .claude/rules/moai/workflow/spec-workflow.md` | `1` | PASS |
| AC-06 (a) | `grep -rn "spec-workflow\.md.*§ Report Persistence" --include="*.md" .claude/ internal/ \| wc -l` | `0` | PASS |
| AC-06 (b) | `grep -rln "spec-workflow-detail\.md § Report Persistence" <4 citation files> \| wc -l` | `4` | PASS |
| AC-06 (c) | 3 anchor-heading greps | `1` `1` `1` | PASS |
| AC-06 (d) | `grep -rln "spec-workflow-detail" <context-loading.md ×2 trees> \| wc -l` | `2` | PASS |
| AC-06 (e) | `grep -c "spec-workflow-detail" .claude/rules/moai/workflow/spec-workflow.md` | `8` (≥ 6) | PASS |
| AC-07 | 7 signature `grep -F -c` + blockquote + `^paths:` greps | `1` ×9 | PASS |
| AC-07 (verbatim) | python substring battery (above) | `FAILS: NONE` | PASS |
| AC-08 | `make embed-manifest` then `make embed-manifest-check` | `ok … embedemit 0.274s` then `ok … embedemit 0.220s`, exit 0 | PASS |
| AC-08 | `grep -c "spec-workflow-detail" internal/template/embed_manifest_gen.go` | `1` (embed line at L217) | PASS |
| AC-09 | `go test ./internal/constitution/ -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/constitution	3.894s` | PASS |
| AC-09 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift\|TestLateBranchTemplateMirror\|TestTemplateNoInternalContentLeak\|TestDeclaredRuleMirrorForks\|TestContractModeAlwaysLoadedBudget' -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/template	4.530s` | PASS |
| AC-09 | `go test ./internal/hook/ -run TestInstructionsLoaded -count=1` | `ok … 1.135s` | PASS |
| AC-09 | `go test ./internal/spec/ -count=1` | `ok … 480.838s` | PASS |
| AC-10 | `make build` | exit 0 (`embed-manifest-check`·`agents-emit-check`·`commands-emit-check`·`tool-policy-drift-check`·`templ-generate` included; `go build … -o bin/moai` completed) | PASS |

All go test invocations used the env-scrubbed compound form (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`) in one invocation each. Full outputs persisted under `.moai/state/verify/t1586-run/` (machine-local scratch; the deciding lines are quoted above — nothing load-bearing lives only in scratch).

### Measured sizes

| File | Runes (both trees) |
|---|---|
| core `spec-workflow.md` | 31,141 (was 40,081 — budget 40,000; ~22% headroom) |
| companion `spec-workflow-detail.md` | 12,007 |

### Deviations from plan (recorded, none silent)

1. **Citation form (AC-06(b) conformance).** Plan §B M3 items 8-9 imply a backtick-preserving citation (`` `spec-workflow.md` § Report Persistence `` → `` `spec-workflow-detail.md` § Report Persistence ``). Measured before editing: the backtick form greps **0** against AC-06(b)'s pattern (`spec-workflow-detail\.md § Report Persistence` requires no backtick between `.md` and ` §`); the unbackticked full path greps **1**. The 6 retargets therefore use the unbackticked full-path form (`.claude/rules/moai/workflow/spec-workflow-detail.md § Report Persistence` / `…detail.md (Run-phase methodology section)`), preserving location AND satisfying the mechanical judge. Backtick styling is lost at exactly 6 sites (4 Report Persistence + 2 methodology pointers).
2. **Generated-file propagation (not in plan §C).** The plan-auditor.md edit made the emitted template twin stale: `make agents-emit` regenerated `internal/template/templates/.codex/agents/moai/plan-auditor.toml` (same one-sentence retarget, verified by diff); `make build`'s templ-generate refreshed `internal/template/catalog.yaml` (hash refresh). Both committed with the M3 evidence — mechanical parity required by `agents-emit-check` and the build gate.
3. **Core size vs projection.** Plan projected ≈28.8KB; measured 31,141 runes. The delta is the fuller header blockquote + 6 pointer lines (new text, not moved content). AC-01's binding bound is ≤ 39,999; no AC binds the projection.
4. **Companion heading promotion.** `### Depends_on Pre-flight Check` and `### Report Persistence` promoted to `##` section headers in the companion (header assembly per plan §B M1 item 2 — bodies verbatim; verified by the byte battery).
5. **Embed-manifest staging order.** `make embed-manifest` reads the *tracked* template set (`git ls-files`): with the new templates companion still untracked, regeneration was a no-op and `embed-manifest-check` failed. Staged the templates files first, then regeneration picked up the embed line. (Matches the t1539 lesson: untracked templates files drop from the embed set.)
6. **Lane-local scope honored.** Full CLI suite NOT run locally (structurally red inside a card worktree; CI owns the full suite). Affected families only, per the dispatch.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-08
evidence: this file §E.2 (all 10 AC observed PASS in this run, tree `c8a6f6fe2`)
commits: 524c9c190 (plan artifacts), 4ddf0e871 (M1 split + draft→in-progress), c8a6f6fe2 (M2 mirror/enrollment/citations), M3 evidence commit (this one)
blockers: none
residual-risk: (1) CI on `origin/develop` is the full-suite verdict surface and has not run yet — local families are the lane-local scope only; (2) AC-07's verbatim guarantee rests on the byte battery run here + sync-phase review against the original L-ranges (the compensating controls acceptance.md names); (3) `go test ./internal/spec/` took 480s wall under concurrent machine load (another session's suite was running) — no flake signal, single run.

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_date: 2026-10-08
card: t1586
pr_surface: github-flow — push + PR to origin/main, lane-executed landing. No CHANGELOG/docs-site surface in this card's scope: the split rules are product-internal deployed assets (`.claude/rules/moai/workflow/` + template mirrors), not user-facing product behavior — release notes carry them.
card_review: advisory fail, 0/3 findings attributable to card scope (all 3 on PR #1772 content / pre-existing content; relays issued t1591-1, t1587-1, t1594) — see `.moai/reports/t1586/card-review.md`
landing_plan: lane pushes the branch and opens the PR to origin/main after this commit
pr: pending
sync_commit_sha: pending-backfill-sync (real SHA backfilled in a follow-up commit — spec-frontmatter-schema.md § SHA placeholder backfill exemption)

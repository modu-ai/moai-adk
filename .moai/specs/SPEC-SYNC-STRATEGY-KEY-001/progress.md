# SPEC-SYNC-STRATEGY-KEY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-27
plan_audit_iter: 2 (iter-1 verdict FAIL 0.86 blocking D2 — `.moai/reports/t303/plan-audit-SPEC-SYNC-STRATEGY-KEY-001-v0.1.0.md`; D2/D3/D1/D4/D5 applied in v0.2.0, D6/D7 routed to orchestrator)
artifacts: spec.md (0.2.0, draft), plan.md, acceptance.md, progress.md — Tier M (4 files incl. progress skeleton)
tree: d29b8942e (worktree t303, branch WT-sync-strategy-key)

## §E.2 Run-phase Evidence

Tree: `63b4628a6` (worktree `t303`, branch `WT-sync-strategy-key`). Full command + verbatim output battery: `.moai/state/verify/t303/ac-battery.txt`; verdict with the 5-section evidence format: `.moai/reports/t303/run-verdict.md`. `T` = `internal/template/templates`, `D` = `T/.claude/skills/moai/workflows/sync/delivery.md`.

| AC | Severity | Command | Actual Output | Status |
|----|----------|---------|---------------|--------|
| AC-SYK-001 | MUST | `grep -c 'git_strategy.{mode}.workflow' D` / `grep -c 'Read \`github.spec_git_workflow\`' D` | `6` / `0` | PASS |
| AC-SYK-002 | MUST | `grep -rn 'spec_git_workflow' T/ \| wc -l`; refined `\| grep -v 'Legacy key fallback' \| wc -l` | `1` (the D1 fallback sentinel only, delivery.md:33); refined `0` — baseline was 10 | PASS (documented refinement) |
| AC-SYK-003 | SHOULD | `grep -c 'v3.3.0' D`; `grep -n 'main_direct\|feature_branch' D` | `1`; hits only at D:39 and D:40, both inside the fallback mapping table | PASS |
| AC-SYK-004 | MUST | `grep -n '{github-flow, git-flow}' D` | D:29 — unmatched-value stop naming the offending value and the canonical domain, "no PR, no push", missing subkey is not a default | PASS |
| AC-SYK-005 | MUST | `grep -c 'WT-' D`; `awk` WT-block `\| grep -c 'gh pr'` | `6`; `0` gh-pr tokens inside the WT-* block | PASS |
| AC-SYK-006 | MUST | `awk '/^##### Strategy:/{s=$0} /matches no defined route/{c[s]++} END{...}' D`; `grep -c 'Default strategy\|Otherwise,' D` | `##### Strategy: github-flow 1` + `##### Strategy: git-flow 1` (exactly 2 lines); negative probe `0` | PASS |
| AC-SYK-007 | SHOULD | `grep -c 'delivery.md'` / `grep -c 'merge --no-ff'` on `.claude/rules/local/gitflow-lane-protocol.md`; same on `D` | `2` / `0`; `D` = `1` — procedure owned by the shipped skill, cited by the dev rule | PASS |
| AC-SYK-008 | MUST | `grep -rn 'gitflow-lane-protocol' T/ \| wc -l`; `grep -n 'workflow:' T/.moai/config/sections/git-strategy.yaml.tmpl`; `git diff 0931789b6 -- T/ \| grep '^+' \| grep -oE 'SPEC-[A-Z0-9-]+-[0-9]{3}' \| wc -l` | `0`; three `workflow: github-flow` lines (13/45/81), no private value leaked; `0` SPEC tokens on added template lines | PASS |
| AC-SYK-009 | MUST | `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1`; `grep -c 'github.spec_git_workflow' internal/config/testdata/shipped_key_inventory.yaml` | `ok github.com/modu-ai/moai-adk/internal/config 0.870s` (exit 0); `0` | PASS |
| AC-SYK-010 | MUST | `grep -c 'github.spec_git_workflow'` / `grep -c 'automation.auto_branch'` on `T/.claude/skills/moai-workflow-project/schemas/tab_schema.json` | `0` / `3`; JSON re-parsed valid after the edit | PASS |
| AC-SYK-011 | SHOULD | `sed -n '25p'` on template and local `doc-execution.md` | both: `- Read project configuration: git_strategy.mode, git_strategy.{mode}.workflow, conversation_language` | PASS |
| AC-SYK-012 | MUST | (1) local dead-key grep refined; (2) `diff` tab_schema template↔local; (3) `grep -c 'SPEC-SYNC-PARALLEL-DOCS-001'` local doc-execution.md; (4) `grep -n 'workflow:' .moai/config/sections/git-strategy.yaml` | (1) `0` **refined** — the raw local grep returns `1`, the same D1 fallback sentinel AC-SYK-002 excludes; AC-SYK-012.1's own text carries no refinement clause, so the exclusion is inherited from AC-SYK-002 and is stated here rather than left implicit (independent sync audit, F3); (2) empty; (3) `3` — A5 block intact; (4) `workflow: git-flow` at L8 | PASS (inherited refinement, disclosed) |

Invariants held (all measured on `63b4628a6`):

| Invariant | Command | Actual Output | Status |
|---|---|---|---|
| Step 3.1 Tier Route A/B gate preserved | `git diff 0931789b6 -- D` inspected for the Route gate hunk | no hunk touches the Route A/B gate block | PASS |
| Template neutrality guard | `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 0.524s` (exit 0) | PASS |
| Adjacent template guards | **Corrected by the independent sync audit (F2).** The run-phase selector `-run 'TestInternalContentLeak\|TestRuleTemplateMirror\|TestCommandsAudit'` named three tests, but only one exists under those names — `grep -rn 'func TestInternalContentLeak\|func TestCommandsAudit' internal/` returns nothing, so that run was a **vacuous green over two of the three**: it established `TestRuleTemplateMirrorDrift` alone. Re-run with the real names on tree `812ee01fc`: `go test ./internal/template/ -run 'TestTemplateNoInternalContentLeak\|TestRuleTemplateMirrorDrift\|TestCommandsThinPattern' -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 6.739s` (exit 0); `-v` shows all three RUN and PASS, so the selector is non-vacuous this time | PASS (re-measured) |
| Embedded templates regenerated | `make build` then `git status --short` | exit 0; empty status — catalog.yaml unchanged because `gen-catalog-hashes.go:22` hashes only each skill's root SKILL.md, not sub-files | PASS |
| Local-only content preserved | `diff` template↔local for delivery.md / doc-execution.md | delivery.md: footer 2 lines only; doc-execution.md: the A5 attribution block only | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: 63b4628a6      # M4; the M5 progress/status commit follows and is the branch tip
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
must_pass_ac_pass_count: 9     # AC-001/002/004/005/006/008/009/010/012
preserve_list_post_run_count: 3   # Step 3.1 Route A/B gate; git-strategy.yaml.tmpl github-flow defaults; local-only mirror content (delivery.md footer, doc-execution.md A5 block)
l44_pre_commit_fetch: not-run     # no push in run phase; git-flow lane integration is a separate gated step
l44_post_push_fetch: not-run      # nothing pushed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-applicable
  reason: documentation/template surface only; no new Go production code. The one Go-adjacent touch is testdata (shipped_key_inventory.yaml), covered by the targeted config test.
total_run_phase_files: 8
m1_to_mN_commit_strategy: one commit per milestone group — M1 c374d9605 (template delivery.md), M2 eb31593d1 (doc-execution + tab_schema), M3 1c9295a07 (tmpl key + inventory row, atomic), M4 63b4628a6 (make build + region-scoped mirror sync + local config + dev rule), M5 (this progress/status commit)
carried_debt:
  - id: B5
    text: "CLAUDE.local.md §2.3 (L180-181, L186) still instructs the operator to re-apply `workflow: gitflow` after `moai update`. That value is now out of the canonical domain and would trip the new unmatched-value stop. CLAUDE.local.md is user-owned — surfaced, not edited."
  - id: OBS-1
    text: "tab_schema.json L517 / L727 bind `git_strategy.personal.auto_branch` / `git_strategy.team.auto_branch`, both missing the `.automation` level the Go struct and git-strategy.yaml.tmpl actually use. Pre-existing, outside this SPEC's scope envelope."
  - id: OBS-2
    text: "The rebound SPEC-branching question now overlaps the per-mode auto_branch questions in the personal/team batches. REQ-SYK-010 prescribed the rebind explicitly; the overlap is reported rather than resolved unilaterally."
  - id: OBS-3
    text: "The template-neutrality CI guard triggers on `pull_request: branches: [main]`. Under the repo's git-flow model card branches carry no PR, so this guard may not fire for this card on CI — it was run locally instead (exit 0)."
```

## §E.4 Sync-phase Audit-Ready Signal

Tree: `812ee01fc` (worktree `t303`, branch `WT-strategy-key-sync`, base develop tip). Coherence-evidence commands re-run on this tree, verbatim outputs below.

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: e9f288473   # this SPEC's sync commit; backfilled in this follow-up commit
sync_status: complete
spec_status_transition: in-progress -> implemented
transition_rationale: >-
  NOT taken to completed. Per the Status Transition Ownership Matrix, manager-docs may carry the
  transition to completed on this same sync commit, but the independent sync audit that decides
  PASS-vs-PASS-WITH-DEBT has not yet run for this SPEC — the dispatching lead stated explicitly
  that the completed call is theirs after that audit, not this agent's. implemented is therefore
  the correct terminal state for this commit.
frontmatter_status_transitions:
  spec.md: "in-progress -> implemented (updated: 2026-08-27)"
  plan.md: "no status: field present in this SPEC's plan.md frontmatter — nothing to transition (see note below)"
  acceptance.md: "no status: field present in this SPEC's acceptance.md frontmatter — nothing to transition (see note below)"
  progress.md: "no status: field present in this SPEC's progress.md frontmatter — nothing to transition (see note below)"
coherence_evidence:
  - claim: "The dead key github.spec_git_workflow survives in exactly one place, the documented D1 fallback sentinel"
    command: "grep -rn \"spec_git_workflow\" internal/template/templates"
    output: "internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:33 (the D1 legacy-fallback sentinel; 1 hit total)"
  - claim: "The canonical key is read 6 times in the shipped sync-delivery skill"
    command: "grep -c \"git_strategy.{mode}.workflow\" internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md"
    output: "6"
  - claim: "REQ-SYK-010's auto_branch rebind survives untouched by t316"
    command: "grep -n \"auto_branch\" internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json"
    output: "3 hits, lines 964/966/982, all git_strategy.{mode}.automation.auto_branch"
  - claim: "t316 (6310dbf28) removed the two malformed personal/team auto_branch bindings this SPEC's OBS-1 named, and did not touch this SPEC's rebind"
    command: "grep -n \"personal.auto_branch\\|team.auto_branch\" internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json"
    output: "exit 1, 0 hits"
  - claim: "Template<->local mirror parity holds for tab_schema.json"
    command: "diff internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json .claude/skills/moai-workflow-project/schemas/tab_schema.json"
    output: "empty diff, exit 0"
  - claim: "The primary checkout's git-strategy.yaml value is inside the canonical domain, resolving carried debt on that config"
    command: "grep -n \"workflow:\\|mode:\" .moai/config/sections/git-strategy.yaml"
    output: "mode: manual (L2); manual.workflow: git-flow (L8) -- inside {github-flow, git-flow}"
  - claim: "Carried-debt item B5 (CLAUDE.local.md instructing the non-canonical 'workflow: gitflow' spelling) is resolved, by card t308"
    command: "grep -n \"workflow: gitflow\" CLAUDE.local.md"
    output: "exit 1, 0 hits"
  - claim: "No docs-site page references the dead key or the mode-scoped canonical path introduced by this SPEC"
    command: "grep -rln \"spec_git_workflow\" docs-site"
    output: "exit 1, 0 hits (existing docs-site references to git_strategy.automation.auto_branch predate this SPEC and are unrelated -- verified by inspection, not edited)"
debt_disposition:
  OBS-1: "resolved-by-t316 (6310dbf28) -- the malformed bindings it named are gone; this SPEC's own rebind is untouched"
  D7b_git_strategy_yaml: "resolved -- primary checkout's git-strategy.yaml carries mode: manual / manual.workflow: git-flow, inside the canonical {github-flow, git-flow} domain"
  B5_claude_local_md: "resolved-by-t308 (da301bbe1) -- the non-canonical 'workflow: gitflow' instruction is gone from CLAUDE.local.md"
  OBS-2: "OPEN -- the rebound SPEC-branching interview question overlaps the per-mode auto_branch questions in the personal/team batches. Reported, not resolved, per REQ-SYK-010's explicit instruction to rebind rather than delete."
  OBS-3: "OPEN -- the template-neutrality CI guard triggers on pull_request: branches: [main]; under this repo's git-flow card branches (no PR) it may not fire on CI for this card. Run locally instead, exit 0 -- not independently re-verified on CI by this sync commit."
ac_count_used_for_b12_self_test: 12   # AC-SYK-001..012, the canonical ### AC-SYK-NNN headings (grep -c '^### AC-SYK-[0-9]\{3\}' acceptance.md -> 12). Corrected by the independent sync audit (F1): the surplus figure is 24 from `grep -oE 'AC-SYK-[0-9]{3}' acceptance.md | wc -l`, and its cause is ACs cross-referencing one another inside their own bodies (AC-005 x4, AC-002 x3, AC-003 x3), NOT the L111 shorthand table -- L111's abbreviated AC-NNN aliases match no AC-SYK- prefix and contribute zero to the 24. The looser regex AC-([A-Z0-9]+-)*[0-9]+ returns 37, and L111's 13 alias tokens are exactly that 37-24 remainder. The original note paired the right conclusion (12 distinct ACs) with the wrong command and the wrong attribution.
b12_self_test_a: "grep -c SPEC-SYNC-STRATEGY-KEY-001 CHANGELOG.md (pre-emission) = 0 -- no duplicate, emission proceeded"
b12_self_test_b: "12 distinct AC-SYK-NNN headings in acceptance.md; CHANGELOG entry states '12 acceptance criteria (AC-SYK-001..012): 12 PASS, 0 FAIL' -- matches"
b12_self_test_c: "all file paths named in the CHANGELOG entry (delivery.md, tab_schema.json, CLAUDE.local.md) verified present via the grep/diff commands above -- no ls-only claim was needed since every path was already read by the coherence-evidence commands"
docs_impact: none   # measured -- see coherence_evidence docs-site row above
changelog_entry_position: "CHANGELOG.md ### Changed section, first bullet (before SPEC-TABSCHEMA-AUTOBRANCH-001)"
```

### Frontmatter transition note (plan.md / acceptance.md / progress.md carry no status: field)

This SPEC's `plan.md`, `acceptance.md`, and `progress.md` were authored at plan-phase with no YAML frontmatter block at all (confirmed: `grep -n "^status:"` across all four artifacts matches only `spec.md`). This differs from at least one other recently-closed SPEC in this repository (SPEC-STATUSLINE-PROFILE-RESPECT-001 / card t293) whose plan.md/acceptance.md/progress.md DID carry a `status:`/`updated:` frontmatter pair — but in that case too, only `spec.md`'s field was actually transitioned at sync-close (its own progress.md §E.4 names the field `spec_status_transition`, singular). The canonical Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields) scopes the 12-field frontmatter schema to `spec.md` specifically; it does not mandate frontmatter on the other three artifacts. Given both the schema scope and the repo precedent, this sync commit transitions `spec.md`'s `status:` field only, and does not add new frontmatter structure to files that never carried it — inventing a frontmatter block on files where none exists would be a body-shape change outside a mechanical status-field transition. Surfaced here rather than silently resolved, per the agent's obligation to report rather than paper over a premise gap between the dispatch instruction and the artifacts actually on disk.

### Independent sync-audit findings — disposition and cross-references

Verdict: PASS-WITH-DEBT, weighted harmonic mean 90.21 against the Tier M threshold 80 (Functionality 96 / Security 93 / Craft 78 / Consistency 90), must-pass firewall cleared on both must-pass dimensions, zero blocking findings. Cold-start auditor, different session from the run and from this sync. Report: `.moai/reports/t303/sync-audit.md`.

**F2 — vacuous green from a selector that matched nothing (the named form).** This is the failure shape card **t241** names: creating a check is not completing it, and "the selector chose nothing, so it printed ok" is one of its canonical forms. The §E.2 "Adjacent template guards" row cited `-run 'TestInternalContentLeak|TestRuleTemplateMirror|TestCommandsAudit'` and reported `ok ... 0.249s`. Two of those three test names do not exist: `grep -rn 'func TestInternalContentLeak\|func TestCommandsAudit' internal/` returns no output. Only `TestRuleTemplateMirrorDrift` matched, so the green established one guard while the row claimed three.

The correction is not "re-run until green" — the original run WAS green. `rc 0` means *everything selected passed*, never *everything that should pass, passed*. The completion act is counting what the selector actually swept: re-run with the real names (`TestTemplateNoInternalContentLeak|TestRuleTemplateMirrorDrift|TestCommandsThinPattern`) gave `ok github.com/modu-ai/moai-adk/internal/template 6.739s`, and `-v` showed all three RUN and PASS — three counted, not inferred. Recorded here at the lead's request as a real-world instance for the t241 implementing session to cite.

**F4 — real defect, deliberately NOT fixed here (outside this SPEC's scope envelope); evidence for card issuance.** File `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md`, line **278**, current text verbatim:

> `6. Push the integration branch: `git push origin develop` — never force-push. On a rejected push, fetch, integrate the fetched integration branch, and push again`

The label says *the integration branch*; the literal hardcodes `develop`. Under a `release/vX.Y.Z` batch — the model this repo's own lane doctrine uses — the literal targets the wrong branch while the surrounding prose reads as if it were parameterized. The mismatch is between the step's own label and its own command, which is why it survives reading. Scope note: the same file's line 299 (`1. Push to develop: `git push origin develop``) is NOT this defect — it sits under a `**develop branch** → Push directly:` heading where `develop` is the correct literal.

**OBS-3 cross-reference — card t333.** The open debt recorded above (the template-neutrality CI guard triggers on `pull_request: branches: [main]`, so under this repo's git-flow card branches it may not fire on CI) is the same trigger-axis family card **t333** covers — a guard that ran under one branch model and silently stops running under another. OBS-3 stays open here and is not resolved by this SPEC; t333 is the place it gets addressed.

**F5 — NOT adopted; the audit's own claim failed measurement.** The audit reported that the new CHANGELOG bullet omits a trailing `MoAI` marker "both adjacent entries carry". Measured on this tree: neither adjacent entry carries one (`sed -n '12p'` and `sed -n '237p'` on CHANGELOG.md, tails read directly). An auditor verdict is itself subject to verification; this one was a false positive and is recorded as such rather than acted on.

### Terminal transition — implemented -> completed (card t303, lane-2, 2026-08-28)

The sync-phase §E.4 above deferred the `completed` call to the dispatching lead, pending an independent re-read of the landed change. That re-read has now run, on a tree at `d566ecc75` (= `origin/develop`, divergence `0 0`), and is recorded verbatim at `.moai/reports/t303/verdict.md`.

```yaml
spec_status_transition: implemented -> completed
transitioned_by: card t303 lane, on the dispatching lead's instruction after the re-read
measured_tree: d566ecc75
landing_ancestry: "git merge-base --is-ancestor 0c7457f8d origin/develop -> exit 0"
ac_recheck: "AC-SYK-001..012, all 12 PASS on this tree; commands and observed outputs in the verdict"
ac_008_baseline_note: >-
  The acceptance recipe pins baseline d29b8942e, which now predates unrelated template edits by
  other cards. Re-running it verbatim would attribute those edits to this SPEC, so the probe was
  scoped to this card's own commit range (0931789b6..63b4628a6): 239 diff lines, zero canonical-shape
  SPEC tokens on added lines.
open_debt_unchanged:
  OBS-2: "OPEN -- interview-schema question overlap; not touched by this transition"
  OBS-3: "OPEN -- neutrality guard trigger axis; card t333"
carried_forward_elsewhere: "D6 (v3.3.0 fallback removal) and D7a (v3.2.0 release-note duty) are card t315's"
```

**F1 (acceptance-text defect, no implementation consequence).** AC-SYK-012 sub-criterion 1 requires `grep -rn 'spec_git_workflow' .claude/skills/ .moai/config/sections/system.yaml` to count **0** on the local tree. The measured count is **1**, and that one hit is the deprecation fallback sentinel at `delivery.md:33` — the very sentinel AC-SYK-003 requires to exist. Applied literally, AC-012.1 and AC-003 cannot both hold. AC-SYK-002 anticipated exactly this on the template side and carried an explicit refinement clause; AC-012.1 did not carry that clause across to the local side. The defect is in the acceptance wording, not in the implementation, and is recorded rather than silently satisfied. A corrective, if the lead wants one, is a one-line refinement of AC-012.1 to "zero excluding the documented fallback sentinel".

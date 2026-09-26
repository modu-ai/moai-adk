# SPEC-MCP-WORKTREE-UNTRACKED-001 — Progress

Card: t1202 | Branch: WT-worktree-moai-root | Base: origin/develop `df526c9a9` | Tier: M

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-26
tier: M
spec_version: "0.7.0"
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
spec_id_check: "Bash regex PASS on SPEC-MCP-WORKTREE-UNTRACKED-001; ID absent from .moai/specs (count 0)"
baseline_tree: df526c9a9
premise_evidence: .moai/reports/t1202/verdict.md
plan_audit_iter1: ".moai/reports/t1202/plan-audit.md — FAIL 0.62"
plan_audit_iter2: ".moai/reports/t1202/plan-audit-iter2.md — FAIL 0.78; lead decision: option B (scope reduction)"
plan_audit_delta: ".moai/reports/t1202/plan-audit-delta.md — FAIL 0.86 (blocked by D27)"
plan_audit_delta2: ".moai/reports/t1202/plan-audit-delta2.md — FAIL 0.87 (blocked by D34)"
plan_audit_delta3: ".moai/reports/t1202/plan-audit-delta3.md — FAIL 0.90 (blocked by D44/D45)"
plan_audit_delta4: ".moai/reports/t1202/plan-audit-delta4.md — FAIL 0.92 (blocked by D50/D51)"
deferred_to: t1213
recommendation: "design (a) alone; operator decides at Kickoff (plan.md §C, 2 open decisions)"
```

### Delta-4 audit map (spec v0.6.0 → v0.7.0)

Source: `.moai/reports/t1202/plan-audit-delta4.md` (FAIL 0.92; D44/D45 measured closed; D50/D51 blocking on the safety axis).

| Item | Change |
|---|---|
| D50 | spec §4.4 condition 4 (admin-dir `gitdir` back-reference) removed; predicate is conditions 1–3 plus the relative-path base rule; two file reads (spec §4.4, §5, plan M4b). Forged-`.git` reasoning restated without condition 4: the root becomes config-orphaned, so REQ-MWU-012 reads the primary git identifies or fails closed — no weakening (spec §4.4, acceptance §D.1). AC-MWU-014 widened: hand-moved `W′` → `fail` + `gate_unmet`. The superseded delta-3 D45 row below remains as history. |
| D51 | AC-MWU-014 widened: relative-path worktree `W_R` evaluated from an unrelated working directory → `fail` + `gate_unmet`. |
| AC budget | Unchanged at 16; both variants added inside AC-MWU-014. |

### Delta-3 audit map (spec v0.5.0 → v0.6.0)

Source: `.moai/reports/t1202/plan-audit-delta3.md` (FAIL 0.90, blocked by D44/D45).

| Item | Change |
|---|---|
| D45 | spec §4.4 predicate now has four conditions: `.git` file with `gitdir:`; target directory exists with parent `worktrees`; target contains `commondir`; target's `gitdir` file canonically equals `<root>/.git`. Three file reads, no subprocess. Excludes submodules at `…/worktrees/<x>` and forged `.git` files. |
| D46 | Relative `gitdir:` resolved against the directory holding the file, never the process cwd (spec §4.4). acceptance.md header: every fixture isolates global/system git config. |
| D44 | AC-MWU-015 widened (count unchanged): (i) adds a non-zero-exit orphaned root (admin-dir `HEAD` removed); (ii) adds the separate-git-dir primary itself, an ordinary submodule root, and a submodule at a `worktrees`-component path. |
| D49 | AC-MWU-015 pins fixture P's workflow.yaml to declare no codex gate, so `fail` can only come from fail-closed. |
| D48 | REQ-MWU-012: fail-closed `gate_unmet` states the gate was assumed `required` because the primary could not be identified; AC-MWU-015(i) asserts it. |
| D47 | REQ-MWU-010 + AC-MWU-013: the `codex_audit` description must not promise refusal of an uncorroborated PASS on such a worktree. |
| §5 | Constraint text updated to three file reads. |

### Delta-2 audit map (spec v0.4.0 → v0.5.0)

Source: `.moai/reports/t1202/plan-audit-delta2.md` (FAIL 0.87, blocked by D34; D27/D29 confirmed closed).

| Item | Change |
|---|---|
| D34 | "Config-orphaned" requires positive, git-free evidence: `<root>/.git` is a file whose `gitdir:` names `<common-dir>/worktrees/<name>`. REQ-MWU-012 fails closed only for such roots; every other root (non-repository, primary checkout, subdirectory, submodule) keeps today's gate path and runs no git for it. The undefined "primary checkout itself" exception is removed. AC-MWU-015 widened: (i) fail-closed on worktree evidence incl. git absent from `PATH`; (ii) `inconclusive` for non-git dir, repository primary, and primary with git absent. |
| D35 | No "not a repository" classification remains. Primary identification runs with scrubbed `GIT_*` plus `LC_ALL=C` and decides by exit status and output shape (REQ-MWU-006, REQ-MWU-012). |
| D36 | Warning moved to `_root.worktree_warning`; the fallback `_root.warning` keeps key and text. AC-MWU-016 asserts both present under `CLAUDE_PROJECT_DIR = W`. |
| D37 | AC-MWU-016 now calls `verify_snapshot`. |
| D38 | AC-MWU-014 requires a non-empty `audit_receipt`. |
| D39 | REQ-MWU-010 and AC-MWU-013 cover the `codex_audit` tool description (and any other description naming the gate source); plan M5. |
| D40 | spec §5 wording: the determination reads one file; git inspections only for config-orphaned roots; the warning needs no git. |
| D41 | `spec_audit` attaches `_root` on every call; existing result fields unchanged. |
| D42 | AC-MWU-014: `audit_multi` must return `overall_verdict: fail` with `gate_unmet` naming codex. |
| D43 | Receipt-id routing pinned to the MCP call site (`mcp_audit_receipt.go`); `auditreceipt.CodexGateRequired` unchanged, hook guard unaffected (REQ-MWU-011, plan §B.3). |
| D32 | No action (auditor: optional, no hand edit). AC count unchanged at 16, so the baseline was not regenerated. |
| Residual (b) of delta-2 | A subdirectory of a worktree has no `.git` file, so it is never config-orphaned — the server-cwd-in-subdirectory case no longer reads the primary's gate. |

### Delta-audit map (spec v0.3.0 → v0.4.0)

| Item | Change |
|---|---|
| D27 | New spec §3.1 (three gate readers) and REQ-MWU-011/012: on a config-orphaned root the `workflow.audit.gates` read comes from the primary checkout; an incomplete determination or ambiguous layout treats the codex gate as `required` (fail closed). Only that key; write destinations stay with t1213. AC-MWU-014 (codex_audit `fail` + `gate_unmet`; audit_multi not `pass`) and AC-MWU-015 (fail-closed; non-worktree unchanged). |
| D29 | "Config-orphaned root" = no own `workflow.yaml` AND linked worktree — durable across receipt writes. REQ-MWU-008 no longer conditioned on the acceptance branch. AC-MWU-014 and AC-MWU-016 repeat after `W/.moai/state/` exists. |
| D30 (lead decision B) | REQ-MWU-013 + AC-MWU-016: `_root` warning on `spec_progress` / `spec_audit` / `spec_drift` / `verify_*` for a config-orphaned root; `spec_audit` gains `_root`. plan §C records the lead decision; the "no warning" default is withdrawn. |
| D28 | AC-MWU-012: bare-`.moai` dirs accepted under a PATH without the VCS binary and a bogus `GIT_DIR`. |
| D31 | REQ-MWU-010 text extended; AC-MWU-013 checks all three statements in both rule copies and `projectRootDescCommon`. |
| D32 | Not hand-edited: the sanctioned regeneration command writes the HEAD at regeneration time into the header (tool behavior; auditor marked no-action). Regenerated again in this commit. |
| D33 | spec §6 new H3: tree-operation `GIT_*` scrubbing is out of scope; REQ-MWU-006/012 scrub only the validator and the gate inspection. |
| AC budget | Old AC-004 and AC-005 merged; ACs renumbered 001–016 (Tier M ceiling). REQ sections reordered so REQ-MWU-001..013 appear in order. |

### Revision map for plan-audit iter-1 (spec v0.1.0 → v0.2.0)

| Defect | Change |
|---|---|
| D1 | plan.md §C keeps the three markers as operator decisions 1–3 plus decisions 4–7, each with a recommended default and the REQs/ACs it changes; REQ text is written under the defaults and carries no marker. |
| D2 | Graph tools are class G (spec §3); REQ-MWU-009 keeps them on the tree root with the existing "graph layer absent" error; AC-MWU-012. |
| D3 | Source-root predicate = MoAI configuration (`.moai/config/sections`), not `.moai` existence; REQ-MWU-011; AC-MWU-014. |
| D4 | spec §3 per-tool read/write inventory with code sites (verified against code); REQ-007 split into REQ-MWU-007..010 by class; ACs per class: T AC-002, C AC-011, G AC-012, S AC-013. |
| D5 | AC-MWU-006 (admin entry removed); AC-MWU-004 now also covers the primary-itself case. |
| D6 | AC-MWU-009 now requires a valid `W` to be accepted under hostile `GIT_DIR`/`GIT_WORK_TREE`; the unrelated-dir rejection is the secondary check. |
| D7 | spec §2.2 states the issue's `moai worktree new` reproduction and its acceptance of (b); §2.3 keeps only measured reasons. |
| D8 | AC-MWU-001 carries three RED-first predicates (is-ancestor, `_test.go`-only RED commit, RED-tree test failure). |
| D9 | plan §B.2 forbids unscrubbed helpers and the shape fallback; REQ-MWU-005/006 require scrubbed git and no alternative retry. |
| D10 | REQ-MWU-004 primary-checkout predicate (common-dir parent + porcelain first entry); ambiguous layouts rejected; AC-MWU-007; spec §6. |
| D11 | spec §2.2 separates 2 execution sites from 3 entry points; credential claim marked unmeasured and dropped as a reason. |
| D12 | AC-MWU-015 checks `projectRootDescCommon`. |
| D13 | AC-MWU-002 names `collectReviewDiff`. |
| D14 | Pattern labels changed to Event-driven. |
| D15 | plan §B.3 file estimate and recount-after-M1 rule. |
| D16 | spec §2.1 describes the check as registration + common-dir, not containment; §2.2 names the `WorktreeCreate` hook as a MoAI entry point. |

### Scope-reduction map for plan-audit iter-2 (spec v0.2.0 → v0.3.0)

| Item | Change |
|---|---|
| Scope (lead option B) | Kept validator acceptance + tree operations. Removed REQ-008 (config/catalogue), REQ-010 (state writes), REQ-011 (no-flip), REQ-012 (dual-root provenance) and AC-011/013/014 of v0.2.0; deferral recorded in spec §6 with card t1213. REQs renumbered contiguously 001–010. |
| D17 | The existing `.moai`-directory branch is first and untouched (REQ-MWU-001, no subprocess); the linked-worktree branch runs only when it fails. New AC-MWU-013: non-git dirs with bare `.moai` or only `.moai/specs/<id>` stay accepted; AC-MWU-014 names the existing fixtures. |
| D18, D20 (C/S rows), D25 | Out of scope → t1213. Inventory trimmed to tree operations; the audit_multi receipt write and gate read are listed in the t1213 exclusion. D20 build-identity row added as a tree operation. |
| D19 | Entry points corrected to 6 with call sites; plan §B.1 records the materializer-caller command. |
| D21 | AC-MWU-001 asserts through the existing `validateProjectRoot(string) (string, error)` only; plan M2 states the RED test compiles pre-fix and fails with the rejection, not a build error. |
| D22 | Moot — the decision it referenced moved to t1213. |
| D23 | REQ-MWU-007 + AC-MWU-011: a dangling or prunable sibling entry is excluded, never a rejection reason. |
| D24 | One-line boundary in acceptance §D.1. |
| D26 | No change; baseline regenerated again with the sanctioned command. |
| §C decisions | Registration and graph decisions resolved in scope; remaining operator list: interim behavior until t1213, design (b), worktree base branch — each with a recommended default; no clarification markers. |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

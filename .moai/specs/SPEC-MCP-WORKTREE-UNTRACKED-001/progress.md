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

Run base: `e29000671` (origin/develop merged in). Commits: RED `f1d578b72`, fix
`2d3bd5061`, rule doc `01d16be19`. All tests below ran against HEAD `01d16be19`
unless marked otherwise; fixtures use `t.TempDir()` with `GIT_CONFIG_GLOBAL` /
`GIT_CONFIG_SYSTEM` pointed at an empty file plus `GIT_CONFIG_NOSYSTEM=1`.

M1 re-verification after the develop merge: the validator's rejection site is
still the `.moai` stat at the end of `validateProjectRoot`
(`internal/cli/mcp_project_root.go`); the three gate readers are unchanged —
`applyGateUnmet` (`mcp_codex.go`), `workflowAuditGates` → `enforceRequiredGateUnmet`
(`mcp_convergence.go`), `recordAuditReceipt` → `auditreceipt.CodexGateRequired`
(`mcp_audit_receipt.go`). `spec_audit` still returned no `_root`.

### RED-first (AC-MWU-001)

| Predicate | Command | Observed |
|---|---|---|
| (i) ancestry | `git merge-base --is-ancestor f1d578b72 2d3bd5061; echo $?` | `is-ancestor exit=0` |
| (ii) `_test.go` only | `git show --name-only --format= f1d578b72` | `internal/cli/mcp_project_root_worktree_test.go` |
| (iii) RED tree fails with the rejection | `go test ./internal/cli/ -run 'TestValidateProjectRoot_AcceptsLinkedWorktreeOfUntrackedMoai$' -count=1` on tree `e29000671` + the RED test file (= tree of `f1d578b72`, run before that commit) | `exit=1`; `mcp_project_root_worktree_test.go:120: validateProjectRoot(W) rejected a linked worktree: project_root ".../W" has no .moai directory, so it is not a MoAI project root`; `FAIL github.com/modu-ai/moai-adk/internal/cli 1.117s` (a test failure, not a build failure) |

### AC matrix

Command for every row unless noted: `go test ./internal/cli/ -run 'TestValidateProjectRoot_AcceptsLinkedWorktreeOfUntrackedMoai|TestLinkedWorktree_|TestValidateProjectRoot_BareMoaiNeedsNoGit|TestConfigOrphanedWorktree_|TestSpecAudit_RootBlockKeepsExistingFields' -count=1 -v` → `exit=0`, `ok github.com/modu-ai/moai-adk/internal/cli 5.449s`.

| AC | Test | Deciding output | Status |
|---|---|---|---|
| 001 | `TestValidateProjectRoot_AcceptsLinkedWorktreeOfUntrackedMoai` | `--- PASS (0.19s)` + RED table above | PASS |
| 002 | `TestLinkedWorktree_TreeOperationsTargetTheWorktree` (diff has W's change only; codex session sent `"cwd":"<W>"`) | `--- PASS (0.44s)` | PASS |
| 003 | `TestLinkedWorktree_RejectsUnrelatedDirectory` | `--- PASS (0.04s)` | PASS |
| 004 | `TestLinkedWorktree_RejectsNonMoaiPrimarySelfAndSubdirectory` | `--- PASS (0.37s)`; logged: `…: its primary checkout …/Q has no .moai directory`; `…: it is the repository's primary checkout, which has no .moai directory`; `…: it is not the top level of a worktree (a subdirectory of …/W)` | PASS |
| 005 | `TestLinkedWorktree_RejectsUnregisteredWorktree` | `--- PASS (0.09s)`; logged: `git could not inspect it as a worktree (not a git repository, an unregistered or unreadable worktree, or git unavailable)` | PASS |
| 006 | `TestLinkedWorktree_RejectsSeparateGitDirLayout` | `--- PASS (0.13s)`; logged: `ambiguous repository layout (separate git dir, submodule, or bare repository)` | PASS |
| 007 | `TestLinkedWorktree_SymlinkCanonicalization` | `--- PASS (0.19s)` | PASS |
| 008 | `TestLinkedWorktree_IgnoresInheritedGitEnvironment` (non-parallel, `GIT_DIR`/`GIT_WORK_TREE` = P) | `--- PASS (0.19s)` | PASS |
| 009 | `TestLinkedWorktree_FailsClosedWithoutGit` (PATH emptied, premise asserted) | `--- PASS (0.07s)` | PASS |
| 010 | `TestLinkedWorktree_DanglingSiblingDoesNotRejectW` | `--- PASS (0.29s)` | PASS |
| 011 | `TestLinkedWorktree_GraphNeverFromPrimary` (positive control: P's query returns `svc.go`) | `--- PASS (0.22s)` | PASS |
| 012 | `TestValidateProjectRoot_BareMoaiNeedsNoGit` (normal run, then PATH emptied + `GIT_DIR` nonexistent) | `--- PASS (0.00s)` | PASS |
| 013 | existing tests: package run (below); docs: `diff .claude/rules/moai/core/moai-mcp-tools.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools.md; echo $?` → `diff exit=0`; neutrality: `grep -nE "SPEC-[A-Z]\|\bt[0-9]{3,4}\b\|20[0-9]{2}-[0-9]{2}-[0-9]{2}" <template copy>` → no match (`exit=1`); `go test ./internal/template/ -run 'TestTemplateNeutralityAudit\|TestTemplateNoInternalContentLeak\|TestMCPTemplateNeutral' -count=1 -v` → `--- PASS: TestTemplateNoInternalContentLeak`, `--- PASS: TestTemplateNeutralityAudit`, `ok … 0.899s`; descriptions: `TestLinkedWorktree_DescriptionsStateTheGateSource` → `--- PASS (0.00s)` | PASS |
| 014 | `TestConfigOrphanedWorktree_PrimaryGateEnforced` — codex_audit `fail` + `gate_unmet` + `audit_receipt`, audit_multi `overall fail` + `gate_unmet` naming codex, both before and after `W/.moai/state` exists; `hand-moved_worktree`; `relative-path_worktree_from_an_unrelated_cwd` | `--- PASS (1.12s)`, `--- PASS: …/hand-moved_worktree (0.17s)`, `--- PASS: …/relative-path_worktree_from_an_unrelated_cwd (0.17s)` | PASS |
| 015 | (i) `TestConfigOrphanedWorktree_FailsClosedWhenPrimaryUnidentified` (P declares no codex gate); (ii) `TestConfigOrphanedWorktree_OtherRootsKeepFailOpen` | (i) `--- PASS (0.40s)` with three subtests PASS; logged `gate_unmet="workflow.audit.gates.codex assumed \`required\` because the primary checkout of this worktree could not be identified, and this audit returned no verdict (fail-open inconclusive)"`; (ii) `--- PASS (0.44s)` | PASS |
| 016 | `TestConfigOrphanedWorktree_CatalogueWarning`; `TestSpecAudit_RootBlockKeepsExistingFields` | `--- PASS (0.63s)`; `--- PASS (0.00s)` | PASS |

### Package runs (quality gate §D.2)

| Check | Command | Observed |
|---|---|---|
| internal/cli package (one run, under slot lease `internal-cli-suite`, HEAD `01d16be19`) | `go test ./internal/cli/ -count=1 -timeout 25m` | `ok github.com/modu-ai/moai-adk/internal/cli 1173.600s`, `exit=0` |
| internal/template | `go test ./internal/template/... -count=1` (after `make build`) | `ok …/internal/template 60.879s`, `ok …/agentemit`, `ok …/commandemit`, `exit=0` |
| vet | `go vet ./internal/cli/` | `vet=0`, no output |
| lint | `golangci-lint run ./internal/cli/...` | `0 issues.`, `exit=0` |
| build | `make build` | `exit=0` |
| SPEC lint | `./bin/moai spec lint SPEC-MCP-WORKTREE-UNTRACKED-001` (binary built by `make build` at `f1d578b72`-dirty, i.e. this card's working tree; spec-lint code not touched by this card) | `✓ No findings — all SPEC documents are valid` |

Package-wide runs were deliberately **not repeated**: after the single
internal/cli run above, the lead reported that full-package runs of
`./internal/cli` and `./internal/hook` write test rows into the real
`~/.moai` profile-leases.db (known leak, fixed on another card not yet on
develop). Further verification was narrowed with `-run`; rows already written
by that one run were left in place as instructed. An earlier unleased
internal/cli run started by this lane was stopped by its own recorded PIDs
before completion when the slot convention was noticed; it produced no verdict.

Coverage of the new file (narrowed run, `-coverprofile`): `scrubbedGitEnv`
100.0%, `runScrubbedGit` 100.0%, `singleGitPath` 75.0%,
`parseWorktreePorcelain` 88.0%, `identifyPrimaryCheckout` 76.2%,
`validateLinkedWorktreeRoot` 83.3%, `isConfigOrphanedRoot` 85.2%,
`resolveAuditGates` 100.0%, `receiptCodexGateRequired` 100.0%,
`withRootBlock` 80.0%.

### Run-phase mandatory checks (verdict.md)

**RUN-MUST-1 — cwd-resolution mutant.** `isConfigOrphanedRoot` temporarily
changed to `gitdir, _ = filepath.Abs(gitdir)` for a relative `gitdir:`; a
temporary probe test evaluated the same relative-path worktree from a
**sibling** cwd. Command: `go test ./internal/cli/ -run 'TestConfigOrphanedWorktree_PrimaryGateEnforced|TestMutantProbe_RelativeWorktreeFromSiblingCwd' -count=1 -v` → `exit=1`:

```
mcp_project_root_worktree_gate_test.go:102: relative-path W_R: want verdict fail with gate_unmet, got verdict="inconclusive" gate_unmet="" summary="codex unavailable: codex binary not found in PATH"
--- FAIL: TestConfigOrphanedWorktree_PrimaryGateEnforced (0.99s)
    --- PASS: TestConfigOrphanedWorktree_PrimaryGateEnforced/hand-moved_worktree (0.17s)
    --- FAIL: TestConfigOrphanedWorktree_PrimaryGateEnforced/relative-path_worktree_from_an_unrelated_cwd (0.03s)
zz_mutant_sibling_test.go:24: sibling cwd=…/002/X rel=../P/.git/worktrees/WR resolves-from-cwd err=<nil>
--- PASS: TestMutantProbe_RelativeWorktreeFromSiblingCwd (0.25s)
```

AC-MWU-014(b) kills the mutant; the sibling-cwd fixture does not (the relative
path resolves from there), which is why the criterion asserts non-resolution
before the call. Mutant reverted, probe file deleted (never committed).

**RUN-MUST-2 — back-reference mutant.** `isConfigOrphanedRoot` temporarily
required the admin directory's `gitdir` file to canonically equal
`<root>/.git`. Command: `go test ./internal/cli/ -run 'TestConfigOrphanedWorktree_PrimaryGateEnforced' -count=1 -v` → `exit=1`:

```
mcp_project_root_worktree_gate_test.go:81: hand-moved W′: want verdict fail with gate_unmet, got verdict="inconclusive" gate_unmet="" summary="codex unavailable: codex binary not found in PATH"
mcp_project_root_worktree_gate_test.go:102: relative-path W_R: want verdict fail with gate_unmet, got verdict="inconclusive" gate_unmet="" summary="codex unavailable: codex binary not found in PATH"
    --- FAIL: TestConfigOrphanedWorktree_PrimaryGateEnforced/hand-moved_worktree (0.05s)
```

AC-MWU-014(a) kills the mutant. Reverted; the post-revert run of
`TestConfigOrphanedWorktree_|TestMutantProbe_` returned `exit=0` with every
subtest PASS.

**Unmeasured premises, now observed** (logged by the committed tests):
- Submodule git dir has no `commondir`: `premise observed: …/super/.git/modules/sub has no commondir` and `…/super/.git/modules/vendor/worktrees/lib has no commondir`.
- Deleting the admin-dir `HEAD` makes git exit non-zero: `premise observed: git rev-parse after admin HEAD deletion: exit status 128`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 01d16be19
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
red_commit: f1d578b72
fix_commit: 2d3bd5061
docs_commit: 01d16be19
new_warnings_or_lints_introduced: 0
cross_platform_build: "darwin only (local); windows/linux left to CI"
total_run_phase_files: 11
m1_to_mN_commit_strategy: "RED test-only commit, then fix+tests, then rule docs, then SPEC status/evidence"
package_wide_runs: "internal/cli once (pre-constraint); not repeated per lead instruction (profile-leases.db leak)"
deferred_to: t1213
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-26
sync_commit_sha: pending-backfill-sync
sync_status: complete
sync_base_head: 8322fb928
b12_self_test_a: "grep -c 'SPEC-MCP-WORKTREE-UNTRACKED-001' CHANGELOG.md -> 0 before emission"
b12_self_test_b: "acceptance.md distinct AC ids = 16 (AC-MWU-001..016), no [RETIRED]/[REF] markers; CHANGELOG entry cites 16"
b12_self_test_c: "ls of every CHANGELOG-claimed path (internal/cli/mcp_worktree_root.go, both moai-mcp-tools.md copies) -> all exist"
changelog_entry_position: "[Unreleased] ### Fixed, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (updated already 2026-09-26)"
  plan_md: "n/a (stateless on the status axis)"
  acceptance_md: "n/a (stateless on the status axis)"
  progress_md: "n/a (no frontmatter)"
docs_synced:
  - docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md (linked-worktree row + paragraph, same commit)
docs_not_changed:
  - "docs-site */advanced/multi-model-audit.md: mentions project_root only as a pointer to the MCP server page; no claim contradicted"
mx_tag_validation: "no new @MX tags added in sync; code unchanged in sync (markdown-only commit)"
deferred_to: t1213
deferred_scope: "state writes, SPEC catalogue and non-gate config routing to the primary checkout; hook-side audit-receipt guard"
tests_run_in_sync: none
```

Sync-phase note: no package-wide `go test ./internal/cli` or `./internal/hook`
run was made (known profile-leases.db leak into the real `~/.moai`); this commit
changes markdown only.

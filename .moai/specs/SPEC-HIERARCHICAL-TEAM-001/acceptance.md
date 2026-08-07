# acceptance.md — SPEC-HIERARCHICAL-TEAM-001

> Canonical AC enumeration. Tier M = 3-artifact set; this file owns the §D AC Matrix + severity / traceability / indirect-verification / closure gates / forward-looking checks. Each AC is binary-testable, written as `AC-XXX Given … When … Then …` per the manager-spec verification-layer format. GEARS requirements live in spec.md §D (REQ-LEAD-001..REQ-CLOSE-001); this file does NOT restate them.

## §A. Verification philosophy

- **Per-AC binary PASS/PARTIAL/FAIL**: every AC verifies via a deterministic command (grep, file existence, lint exit code, agent-body inspection). No "subjective quality" ACs.
- **Peer cross-validation applies to run-phase AC evidence**, not to plan-phase AC verification: plan-auditor runs once at plan-phase gate; REQ-PEER-001's second-worker re-run fires only at run-phase milestone boundaries. Plan-phase AC verification uses the standard manager-spec self-verification batch.
- **Evidence persistence**: every run-phase AC verification command's verbatim output redirects to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}` per REQ-FOLD-002. Cited paths MUST resolve at audit time.
- **Template-mirror parity**: every distributed-surface AC (CLAUDE.md, `.claude/rules/moai/**`, `.claude/agents/moai/manager-lead.md`) verifies BOTH the live edit AND its template mirror under `internal/template/templates/`. Per §25 template-neutrality, template copies MUST NOT carry SPEC IDs / REQ tokens / internal dates.

## §B. Severity model

| Severity | Meaning | Action |
|---|---|---|
| MUST | Blocks run-phase completion; blocks sync-phase entry | Resolve before M6 close |
| SHOULD | Blocks sync-phase PASS (sync-auditor flags); does NOT block run-phase M{n} advance | Resolve before sync close |
| NICE | Documentation-only; tracked as debt | Optional |

All ACs in §D are MUST unless marked otherwise.

## §C. Traceability matrix

| AC ID | REQ ID | Milestone | Severity |
|---|---|---|---|
| AC-LEAD-001 | REQ-LEAD-001 | M1 | MUST |
| AC-LEAD-002 | REQ-LEAD-002 | M1 | MUST |
| AC-LEAD-003 | REQ-LEAD-003 | M1 | MUST |
| AC-DEPTH-001 | REQ-DEPTH-001 | M1 | MUST |
| AC-WORKTREE-001 | REQ-WORKTREE-001 | M2 | MUST |
| AC-WORKTREE-002 | REQ-WORKTREE-002 | M2 | MUST |
| AC-FOLD-001 | REQ-FOLD-001 | M3 | MUST |
| AC-FOLD-002 | REQ-FOLD-002 | M3 | MUST |
| AC-FOLD-003 | REQ-FOLD-003 | M3 | MUST |
| AC-PEER-001 | REQ-PEER-001 | M4 | MUST |
| AC-PEER-002 | REQ-PEER-002 | M4 | MUST |
| AC-FANOUT-001 | REQ-FANOUT-001 | M5 | MUST |
| AC-FANOUT-002 | REQ-FANOUT-002 | M5 | MUST |
| AC-CLOSE-001 | REQ-CLOSE-001 | M6 | MUST |
| AC-REGRESS-001 | (cross-cutting) | M6 | MUST |

## §D. AC Matrix (Given-When-Then)

### AC-LEAD-001 — `manager-lead.md` coordination-only + Agent-carrier + depth-2 seal

**Given** the run-phase M1 has landed (commit on `feat/spec-hierarchical-team-001`) and `.claude/agents/moai/manager-lead.md` exists.
**When** the inspector greps the agent file's YAML frontmatter and body.
**Then** ALL of:
- `tools:` field includes the literal token `Agent`;
- `tools:` field includes `Read`, `Grep`, `Glob`, `Bash`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `Skill`;
- the body prose declares a coordination-only role (does NOT write code, does NOT author SPEC body content);
- the body prose declares it returns blocker reports (does NOT invoke `AskUserQuestion`);
- `grep -c 'AskUserQuestion' .claude/agents/moai/manager-lead.md` returns `0`.

**Verify**: `grep -nE '^(tools:|name:|description:)' .claude/agents/moai/manager-lead.md`; `grep -c 'AskUserQuestion' .claude/agents/moai/manager-lead.md`.

### AC-LEAD-002 — CLAUDE.md §4 Selection Decision Tree 13th row + non-regression

**Given** the M1 distributed-surface edit to `CLAUDE.md` has landed and its template mirror at `internal/template/templates/CLAUDE.md` is byte-identical (modulo whitespace) to the live edit.
**When** the inspector greps CLAUDE.md §4 Selection Decision Tree.
**Then** ALL of:
- the tree carries a row matching `Use the .manager-lead. subagent` (regex `manager-lead`);
- the row's predicate names multi-milestone Tier L coordination (`≥3 milestones AND ≥10 files AND cross-domain fan-out` or semantic equivalent — conjunctive AND, NOT disjunctive OR);
- the existing 12 rows (1-12) remain present verbatim (no row deleted, no row renumbered);
- `grep -c 'Use the' CLAUDE.md` returns ≥ the pre-M1 count + 1 (the new 13th row).

**Verify**: `grep -nE 'manager-lead' CLAUDE.md`; `grep -n '≥3 milestones AND ≥10 files AND cross-domain fan-out' CLAUDE.md` (expect ≥1 match); diff `CLAUDE.md` against `internal/template/templates/CLAUDE.md` (whitespace-normalized).

### AC-LEAD-003 — CLAUDE.md §4 Retained Agents table 12 entries + Watch note amendment

**Given** the M1 distributed-surface edit has landed + mirrored.
**When** the inspector greps CLAUDE.md §4.
**Then** ALL of:
- the Retained Agents table carries a row for `manager-lead` with class `core/manager` and a phase-scope reference;
- the table header row count is now 12 (was 11);
- the Watch note's flat-hierarchy claim is amended to name `manager-lead` as the sole Agent-carrier (regex `sole exception among retained agents` OR `sole Agent-carrier`);
- the Supersession note for `SPEC-SUBAGENT-NESTING-DOCTRINE-001` references this SPEC's depth-2 seal (regex `SPEC-HIERARCHICAL-TEAM-001|depth-2 seal`).

**Verify**: `grep -c '| .manager-lead. |' CLAUDE.md`; `grep -E 'sole (exception|Agent-carrier)' CLAUDE.md`; `grep -E 'SUBAGENT-NESTING-DOCTRINE-001' CLAUDE.md` (the supersession note).

### AC-DEPTH-001 — depth-2 seal CI guard (defense-in-depth)

**Given** the M1 depth-2 seal is in effect and the OQ-4 RESOLVED decision (guard obligatory) has landed.
**When** the inspector runs the CI test under `internal/template/` (or, if the test has not yet been authored, greps for its presence + the grep pattern).
**Then** ALL of:
- a CI test exists under `internal/template/` mirroring the `subagent_boundary_test.go` pattern (grep the test for `subagent_boundary` or `manager-lead` references);
- the test greps every `manager-lead`-spawned leaf-worker agent file for the literal token `Agent` in its `tools:` list;
- the test FAILS the build when a leaf-worker agent file adds `Agent` to `tools:` (regression guard);
- where a leaf-worker agent file omits `Agent` from `tools:` (the REQ-LEAD-001 invariant), the test PASSES.

**Verify**: run the CI test (`go test ./internal/template/... -run <DepthSeal|SubagentBoundary>`); or, if not yet authored, `ls internal/template/*depth* internal/template/*leaf* 2>/dev/null` + `grep -rln 'manager-lead.*Agent\|leaf-worker.*Agent' internal/template/` (expect the test file path to resolve once authored at M1).

### AC-WORKTREE-001 — worktree-integration.md decision tree + HARD rule re-key

**Given** the M2 distributed-surface edit has landed + mirrored.
**When** the inspector greps `.claude/rules/moai/workflow/worktree-integration.md`.
**Then** ALL of:
- the § Worktree Selection Rules decision tree carries the predicate `parallel write workers within a hierarchical team` (or semantic equivalent referencing `manager-lead` fan-out);
- the HARD rule line (~194) carries the same re-keyed predicate;
- the residual phrase `team mode implementation with parallel agents` is ABSENT from the decision tree (regex `team mode implementation` returns 0 matches in § Worktree Selection Rules);
- the L1/L2 layer distinction, the read-only-agent exemption, and the `isolation: "worktree"` mechanism are unchanged (regex `isolation: .worktree.` still matches).

**Verify**: `grep -nE 'parallel write workers within a hierarchical team|manager-lead fan-out' .claude/rules/moai/workflow/worktree-integration.md`; `grep -nE 'team mode implementation' .claude/rules/moai/workflow/worktree-integration.md` (expect 0 in the re-keyed sections; the § Agent Teams Variant RETIRED tombstone in spec-workflow.md is OUT of scope and may still carry the phrase).

### AC-WORKTREE-002 — agent-common-protocol.md § Background Agent re-key + concurrency safeguard retention

**Given** the M2 distributed-surface edit has landed + mirrored.
**When** the inspector greps `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution.
**Then** ALL of:
- the stale team-mode framing is re-keyed to the same predicate as AC-WORKTREE-001;
- the literal sentence `MoAI does not run two write-capable agents concurrently` is present verbatim (regex anchored);
- the `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` reference (default 20) is unchanged;
- the MoAI Mode 4 ceiling (`3-5 concurrent Agent()`) is unchanged.

**Verify**: `grep -nE 'parallel write workers within a hierarchical team' .claude/rules/moai/core/agent-common-protocol.md`; `grep -c 'does not run two write-capable agents concurrently' .claude/rules/moai/core/agent-common-protocol.md` (expect ≥ 1).

### AC-FOLD-001 — Context-Folding 3-step procedure at milestone Mn

**Given** the M3 manager-lead body carries the fold procedure and a Tier L milestone Mn has completed (all M{n} AC rows PASS).
**When** manager-lead detects Mn completion.
**Then** manager-lead executes ALL of:
1. persists each M{n} AC's verification command output to `.moai/state/verify/<session>/M<n>.<AC-id>.{log,out}`;
2. appends a `progress.md` §E.2 fold row of form `M<n>: <AC-id-1>=PASS, ... | evidence: .moai/state/verify/<session>/M<n>.* | fold-at: <ISO-8601>`;
3. invokes `/compact` with explicit retain-current-milestone + retain-fold-rows instructions (and retain-armed-goal per AP-3).

**Verify** (run-phase, post-M3): inspect `progress.md` §E.2 for fold rows; inspect `.moai/state/verify/<session>/M<n>.*` paths resolve; inspect manager-lead agent log for the `/compact` invocation.

### AC-FOLD-002 — Evidence-persistence + GAP marking

**Given** the M3 fold procedure is in effect.
**When** the inspector probes a §E.2 fold row's cited evidence path.
**Then** ALL of:
- the path resolves on disk (NO dangling references);
- the path is under `.moai/state/verify/<session>/` (NOT `/tmp`);
- any AC whose evidence could not be populated is marked `GAP` (NOT `PASS`) in the fold row;
- manager-lead did NOT advance to M{n+1} while a GAP marker was unresolved (return-blocker evidence).

**Verify**: `find .moai/state/verify/<session>/M<n>.* -type f` resolves; `grep -E 'GAP' .moai/specs/SPEC-HIERARCHICAL-TEAM-001/progress.md` (expect 0 in a clean run; any GAP is blocker-evidence).

### AC-FOLD-003 — Bounded-context invariant

**Given** the M3 fold procedure has fired at milestone Mn.
**When** the inspector reads post-fold token usage from the statusline context-usage readout.
**Then** ALL of:
- post-fold token usage < pre-fold token usage (the fold reduced the live context — strict binary);
- post-fold token usage < the model-specific handoff threshold (context-window-management.md § Context Window Targets — 50% on 1M / GLM-5.2, 90% on 200K/256K);
- the armed goal condition (if any) is preserved across the `/compact` (AP-3).

**Verify**: statusline context-usage readout pre-fold vs post-fold (numeric comparison: post < pre); confirm post-fold reading is below the model-specific handoff threshold; `moai goal status` (if armed) returns the same condition post-fold.

### AC-PEER-001 — Peer cross-validation spawn at Tier M/L

**Given** the M4 manager-lead body carries the peer-spawn step and a Tier M/L milestone Mn AC has been marked PASS by `manager-develop`.
**When** manager-lead processes the PASS report.
**Then** manager-lead spawns a second `Agent(general-purpose)` (NOT the author) with `tools:` omitting Write/Edit/NotebookEdit, that re-runs the acceptance.md §D GWT commands for that AC, and returns PASS / PARTIAL / FAIL. Tier S ACs are skipped.

**Verify** (run-phase): inspect manager-lead agent log for the peer-spawn `Agent(general-purpose)` invocation; verify the peer worker's `tools:` list omits Write/Edit.

### AC-PEER-002 — Peer FAIL/PARTIAL blocker behavior

**Given** the M4 peer-spawn step is in effect.
**When** the peer worker returns FAIL or PARTIAL for an AC the author marked PASS.
**Then** manager-lead returns a structured blocker report to the orchestrator (NOT an `AskUserQuestion` call); the orchestrator (NOT manager-lead) runs the AskUser round per REQ-PEER-002; manager-lead did NOT advance to M{n+1}.

**Verify** (run-phase): `grep -c 'AskUserQuestion' .claude/agents/moai/manager-lead.md` returns `0`; blocker-report path is logged.

### AC-FANOUT-001 — Schema-driven fan-out reduce contract

**Given** the M5 manager-lead body carries the reduce step.
**When** manager-lead fans out ≥ 3 explorer agents.
**Then** ALL of:
- each explorer's return matches the `plan-research-fanout` skill's fixed-heading markdown schema;
- manager-lead's reduce step is mechanical merge (no per-spawn re-derivation);
- cross-explorer contradictions are annotated as a named section in the merged result (NOT silently discarded).

**Verify**: diff a sample explorer return against `plan-research-fanout/SKILL.md`'s schema declaration; inspect the merged result for a `Contradictions` / `Cross-lens conflicts` section when contradictions exist.

### AC-FANOUT-002 — Fan-out concurrency ceiling

**Given** the M5 manager-lead body carries the concurrency cap.
**When** manager-lead fans out leaf workers.
**Then** the concurrent leaf-worker spawn count NEVER exceeds 5; where > 5 workers are warranted, manager-lead sequences them in batches. The runtime cap `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` (default 20) is unchanged.

**Verify**: `grep -nE '3-5|≤ 5|concurrent' .claude/agents/moai/manager-lead.md`; inspect manager-lead agent log for batch sequencing on > 5 spawns.

### AC-CLOSE-001 — Phase 4 mode taxonomy non-regression

**Given** the M6 §G.2 note has landed in `orchestration-mode-selection.md`.
**When** the inspector greps `.claude/rules/moai/workflow/orchestration-mode-selection.md`.
**Then** ALL of:
- §A Mode catalog carries Modes 1-6 verbatim (no Mode 7 added);
- the `--mode` dispatch axis values (`autopilot`, `loop`, `team`, `pipeline`) are unchanged;
- the `MODE_TEAM_UNAVAILABLE` sentinel is unchanged;
- §G.2 names `manager-lead` as a Mode-5-shaped delegation target (regex `Mode-5-shaped` OR `Mode 5 shaped`).

**Verify**: `grep -nE 'Mode 7|Mode-7' .claude/rules/moai/workflow/orchestration-mode-selection.md` (expect 0); `grep -nE 'manager-lead' .claude/rules/moai/workflow/orchestration-mode-selection.md` (expect the §G.2 reference).

### AC-REGRESS-001 — spec-lint strict zero on this SPEC directory

**Given** all run-phase milestones M1-M6 have landed.
**When** the inspector runs `moai spec lint --strict .moai/specs/SPEC-HIERARCHICAL-TEAM-001/`.
**Then** exit code 0; output reports 0 errors and 0 strict errors (no `LegacyEARSKeyword`, no `FrontmatterInvalid`, no `MissingExclusions`, no `OwnershipTransitionInvalid`, no `FrontmatterPhaseInvalid`).

**Verify**: `moai spec lint --strict .moai/specs/SPEC-HIERARCHICAL-TEAM-001/ ; echo "exit=$?"`.

## §D.1 Severity promotion rules

- All ACs in §D are MUST. SHOULD/NICE severities apply ONLY to documentation debt explicitly deferred to a follow-up SPEC (none in this SPEC's scope).
- A MUST AC that PARTIAL-passes at M6 close blocks sync-phase entry. The orchestrator runs `AskUserQuestion` to either (a) accept as documented debt with a follow-up SPEC, (b) re-spawn manager-develop for the failing AC, or (c) abort.

## §D.2 Indirect verification (when the direct command is unavailable)

- **`/compact` availability in subagent context (assumption 3)**: if Claude Code v2.1.219+ does NOT honor `/compact` invoked by a subagent (manager-lead), AC-FOLD-001 step 3 cannot fire. **Indirect verification**: manager-lead returns a blocker report; M3 re-plans to either (a) escalate the compact to the orchestrator (parent), or (b) fall back to `/clear` + paste-ready resume (the session-handoff ladder rung 2). The fallback is documented in the manager-lead body; the indirect-verification exit is the blocker-report path.
- **Peer worker's read-only enforcement**: the spawn-time `mode` parameter is deprecated since v2.1.219; read-only enforcement rests on the `tools:` list omission (CLAUDE.md §4 Watch note). **Indirect verification**: AC-PEER-001 verifies the `tools:` list omission directly; runtime behavior is verified by attempting a Write call from the peer worker and confirming it's rejected.

## §D.3 Closure gates

- **Run-phase close**: ALL MUST ACs PASS (AC-LEAD-001..AC-REGRESS-001); §E Self-Verification matrix green; `progress.md` §E.2 fold rows present for M1-M6.
- **Sync-phase close**: `moai spec audit` reports 0 drift on this SPEC; ` OwnershipTransitionRule` reports 0 findings (the `draft → in-progress` transition on M1 commit is owned by manager-develop per the Status Transition Ownership Matrix; the `implemented → completed` transition rides the sync commit owned by manager-docs).
- **Definition of Done**: this SPEC's `status: completed` rides the sync commit; `sync_commit_sha` backfilled in `progress.md` §E.4 per the SHA-placeholder backfill exemption (D3).

## §D.4 Forward-looking checks (post-close invariants)

- **Depth-2 seal regression guard**: if a future SPEC amends `manager-lead.md` to add `Agent` to a leaf-worker agent's `tools:` list, the REQ-DEPTH-001 CI guard (obligatory per OQ-4 RESOLVED) catches it at lint time. The regression never reaches runtime (where it would otherwise be rejected by `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH`).
- **Worktree re-key stability**: if a future SPEC re-introduces team-mode language into `worktree-integration.md`, AC-WORKTREE-001's residual-phrase grep (`team mode implementation` returns 0) catches the regression at sync-phase lint.
- **Context-Folding evidence audit**: a future `moai spec audit` SHOULD resolve all `.moai/state/verify/<session>/M<n>.*` paths cited in §E.2 fold rows. Dangling paths indicate either session-directory cleanup over-reach or a fold-procedure gap — either is a follow-up SPEC, not a silent acceptance.

## §D.5 Plan-audit forward-looking checks (binding at plan-audit gate)

- **`phase:` value audit**: `phase: "v3.x target"` is a release-target label (NOT a lifecycle-stage token). `moai spec lint --strict` MUST NOT emit `FrontmatterPhaseInvalid`. (Self-check passes by construction at plan-phase emission; included here for plan-audit visibility.)
- **OQ resolution gate**: OQ-1 / OQ-2 / OQ-3 / OQ-4 are DEFERRED to Implementation Kickoff per the AUTONOMY-TIERS precedent. plan-auditor verifies the OQs are documented in spec.md §F and surface at the Implementation Kickoff Approval gate; the OQ decisions ride `progress.md` §E.1 as plan-phase-terminated inputs.
- **REQ/AC budget**: Tier M ceiling is 16 REQs / 16 ACs. This SPEC carries 14 REQs (REQ-LEAD-001..REQ-CLOSE-001, including REQ-DEPTH-001) + 15 ACs (AC-LEAD-001..AC-REGRESS-001, including AC-DEPTH-001). Within budget; plan-auditor verifies no ceiling breach.

## §D.6 Anti-patterns binding at acceptance time

- **AP-ACCEPT-001 — Peer worker conflated with sync-auditor**: AC-PEER-001/002 verify a RUN-PHASE per-AC binary check; sync-auditor is a SYNC-PHASE 4-dimension harmonic-mean score. A plan-audit finding that conflates the two (e.g., demanding peer cross-validation produce a 4-dimension score) is a category error — reject.
- **AP-ACCEPT-002 — Fold-row format over-constraint**: AC-FOLD-001 specifies the fold-row prefix `M<n>:` and the evidence-path citation; it does NOT constrain the row's free-form prose beyond the verification-claim-integrity §2 attribution requirement. A plan-audit finding demanding a stricter schema is over-reach — the format intentionally coexists with era.go's matchers without requiring matcher changes.
- **AP-ACCEPT-003 — Template-mirror parity scoped to distributed surfaces only**: AC-LEAD-002/003 and AC-WORKTREE-001/002 verify BOTH live + template-mirror for distributed surfaces (CLAUDE.md, `.claude/rules/moai/**`, `.claude/agents/moai/manager-lead.md`). The plan.md / acceptance.md / spec.md / progress.md files are NOT distributed (they live under `.moai/specs/`, which is explicitly exempt per CLAUDE.local.md §2 "Local-Only Files") — they do NOT require template mirrors.

## §D.7 Summary

15 ACs total, all MUST severity, all binary-testable, all traceable to a REQ in spec.md §D (except AC-REGRESS-001 which is cross-cutting). The matrix is the run-phase verification SSOT; the spec.md §H summary is the cross-reference index. Indirect verification (§D.2) covers the two load-bearing assumptions (`/compact` in subagent context, peer-worker read-only enforcement) whose direct verification is changelog-sourced and run-phase-confirmed.

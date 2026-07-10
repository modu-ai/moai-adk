# Design — SPEC-AGENT-TEAM-RETIRE-001

> Tier L design record. Decisions D1-D8 govern Scope A boundaries and ordering;
> §C/§D specify the two Scope B workflow scripts at design (not implementation)
> level.

## §A. Design Decisions

### D1 — Boundary: remove the STATIC layer, keep the NATIVE runtime

The retirement severs along the ownership line: everything MoAI *built on top
of* Claude Code teammates (role_profiles CONFIG schema, team_spawn coordination
primitives, team rules/skills, web-console team fieldset) is removed;
everything the Claude Code runtime itself provides or that MoAI wired directly
to the runtime (`Agent(name=...)`, `~/.claude/teams/`, `teammateMode`
launchers, `worktree --team` P1-P4, session `MergeTeamCheckpoints`) is
preserved. Rationale: Mode 3 is already default-disabled
(orchestration-mode-selection.md §C.1) and the practical surface is covered by
Modes 4/5/6; but the native runtime remains a supported Claude Code feature the
user can opt into (`moai cg` GLM teammates), so its plumbing must survive.
Alternative rejected: full removal including tmux launchers — breaks `moai cg`
cost-optimization mode, which is actively used.

### D2 — Phase 0 de-risk ordering (migrate before delete)

`ClaimTask` and the lock pair are the only team_spawn symbols with live value:
they carry the SPEC-CLIFIX-CRITICAL-001 P0 concurrency fixes (map RMW,
O_APPEND) and their regression tests. Deleting first and re-adding later would
put the P0 guard RED mid-flight and hide regressions in the gap. Hence M0
migrates (`internal/lockfile`, `internal/cli/taskledger`) with thin aliases
left in team_spawn.go, lands whole-repo green, and only then does M1 delete.
REQ-ATR-003 encodes this as a state-driven gate. Alternative rejected: delete
ClaimTask + repro test together ("the defect was in the deleted code") — the
claim/ledger primitive is generically useful for native-teammate state files
and the repro test is the only executable record of the P0 defect class.

### D3 — Repurpose agentlint instead of deleting the subcommand

`moai workflow lint` currently validates only `role_profiles` isolation.
Removing the checks without replacement leaves a shipped subcommand that always
exits 0 — a no-op CLI surface (worse than absent: it implies validation that
does not happen). The cheapest live check with an existing validator is
`model_routing_profiles` closed-set validation (`ValidateModelRoutingProfiles`
already exists in `internal/config`; the lint command merely gains a caller).
Alternative rejected: delete the subcommand — breaks the documented `moai
workflow lint` surface and its CI/docs references for no gain; the Simplicity
Ladder rung 2 (reuse existing validator) favors repurpose.

### D4 — Verdict arithmetic in script JS, never in an LLM

The harmonic mean `n/Σ(1/sᵢ)` is computed in the workflow script body.
Rationale: (a) LLM arithmetic is non-deterministic and unverifiable —
`agent-common-protocol.md` § Skeptical Evaluation Stance requires the harmonic
mean over dimensions, and a mean computed by a model cannot be audited;
(b) a "meta-judge" aggregation agent adds a 5th LLM call whose only job is math
plus opinion-averaging — it smooths dissent, which is exactly what the
harmonic mean exists to prevent (one low dimension must drag the verdict).
Zero-score guard: `sᵢ = 0` short-circuits to FAIL naming the dimension (no
division). Null-judge guard: any null → `INCOMPLETE` naming the dimension;
never harmonic-mean over 3 (a missing dimension is evidence-absence, and
evidence absent ≠ evidence of success per verification-claim-integrity §1).

### D5 — Markdown outputs for explorers, schema for judges

`plan-research-fanout.js` explorers return fixed-heading markdown, NOT a
forced schema — the codemaps pilot lesson: schema-forced output is brittle
under rate limiting (a truncated JSON is unusable; truncated markdown
degrades gracefully), and research narrative fits headings better than JSON.
The mandatory `### confidence_and_gaps` heading keeps honesty structural
("NONE found" is a valid, valuable answer). `sync-audit-4dim.js` judges DO use
schema-forced output because the verdict computation consumes typed fields
(`score`, `findings[].severity`) — arithmetic needs structure; narrative does
not. This asymmetry is deliberate.

### D6 — INCOMPLETE / insufficient_coverage semantics (fail-honest)

Both scripts prefer honest partial-failure over silent degradation:
sync-audit returns `INCOMPLETE` on ANY null judge (4 dimensions are the
contract; 3/4 is not a weaker verdict, it is no verdict); plan-research
tolerates 1 null lens (coverage gap named in synthesis) but aborts at ≥2
(a synthesis over half the lenses would smooth over unknown unknowns —
`insufficient_coverage` forces a re-run or a conscious lens reduction by the
orchestrator). Thresholds live in args so the orchestrator can tighten, never
loosen silently.

### D7 — Template mirror strategy

Rules/skills/config removals apply to BOTH trees (local `.claude/` +
`internal/template/templates/.claude/`) per the Template-First Rule
(CLAUDE.local.md §2); the Scope B workflow scripts apply to the LOCAL tree
ONLY (`.claude/workflows/` is user-owned, not template-managed per
dynamic-workflows.md — mirroring them would violate §24 namespace policy and
ship maintainer tooling to all users). Per-tree independent `[ -e ]` absence
loops are mandated (SUBCOMMAND-RETIRE-001 D7 lesson: piped `grep -q "No such"`
false-passes when residue survives in exactly one tree).

### D8 — Auto-select thresholds become prose-only SSOT

`TeamAutoSelectionConfig` (domains≥3 / files≥10 / score≥7) is removed with the
team block. The thresholds still drive Phase 0.95 Mode 3/4 auto-selection
heuristics in doctrine. Recommendation: `orchestration-mode-selection.md` §B.1
becomes the sole (prose) SSOT and drops its "machine-readable source is
workflow.yaml auto_selection" pointer — no Go code reads the struct outside
the deleted tests, so the machine-readable claim is already vestigial.
**ADOPTED** (user decision 2026-07-11, orchestrator-relayed): prose-only SSOT
confirmed; encoded in REQ-ATR-010 (extended) + AC-ATR-028.

### D9 — team/glm.md: migrate essentials, then delete (user decision)

The user resolved the glm.md question as **migrate-essentials-then-delete**:
the file stays in the Phase 6 deletion set, but M3 step 0 first relocates the
essential CG Mode (GLM teammate) guidance — LLM mode detection, prerequisites,
tmux environment variables, error recovery — into a new
`## CG Mode (Claude + GLM teammates)` section of
`.claude/rules/moai/core/glm-web-tooling.md` (both trees; the rule has a
template mirror). The Agent Teams orchestration prose in glm.md (team spawn
patterns, role assignments) is NOT migrated — it is the retired static layer.
Rationale: `moai cg` (GLM cost-optimization mode) remains an active, preserved
CLI surface (REQ-ATR-006e); deleting its only teammate-guidance doc would
orphan the feature. Encoded as REQ-ATR-022 (While migrate-then-delete gate) +
AC-ATR-027 (grep token `CG Mode (Claude + GLM`, 0-count baseline verified).

## §B. Scope A Structural Notes

- **internal/lockfile API sketch**: `func Lock(f *os.File) error` /
  `func Unlock(f *os.File) error`, two files (`lock_unix.go` /
  `lock_windows.go`) + migrated test. `@MX:ANCHOR` on the exported pair
  (`@MX:REASON`: advisory-lock invariant contract; fan-in from taskledger and
  future state writers).
- **internal/cli/taskledger contents**: `TeamTaskEntry` (rename decision at
  run-phase: keep name for churn minimization vs rename to `TaskEntry` —
  prefer KEEP to minimize diff; record either way), `AppendTask`, `ClaimTask`,
  `TaskClaimer`, `NewTaskClaimer`. `@MX:ANCHOR` +
  `@MX:SPEC: SPEC-CLIFIX-CRITICAL-001` on `ClaimTask`.
- **Deletion blast-radius (corrected per auditor D3)**: the 0-non-test-caller
  claim holds for the team_spawn.go SYMBOL family only. The CONFIG TYPE
  family has non-test consumers: `workflow_accessors.go`
  `WorkflowTeamAutoSelection()` (deleted in M1), `workflow_lint.go`
  `validateRoleProfiles` (removed in M1, replaced in M2), and the
  settings/web surfaces (M2). "Confined to tests" is false for the type
  family — the corrected enumeration below governs.
- **Deletion blast-radius (team_spawn symbols)**: 0 non-test callers of every deleted
  symbol; compile-coupled edits confined to `internal/config` tests +
  `internal/cli/agentlint/workflow_lint.go` + `internal/settings` +
  `internal/web` (Phase 3 surfaces).

## §C. sync-audit-4dim.js Design

```
meta.phases: [ Context, Judge, Verdict ]
args: { spec_id (required), threshold = 0.85, tier = 'M' }   // Tier S: caller does not launch (gate note in header)

Phase Context  — 1x agent(effort 'medium', agentType 'Explore',
                 schema { spec_id, acceptance_criteria[], changed_files[], test_command })
                 label 'context:<spec_id>'
Phase Judge    — parallel(4) — one per DIMENSIONS = ['Functionality','Security','Craft','Consistency']
                 agent(effort 'xhigh', read-only, schema { dimension, score(0..1),
                   findings[{severity,summary,file,evidence}], evidence_gaps[] })
                 prompt: skeptical-auditor stance; every claim needs command + verbatim output;
                 evidence absent ≠ pass; label 'judge:<dimension>'
Phase Verdict  — SCRIPT JS ONLY (no agent call):
                 nulls = judges where result is null → if any: return { verdict:'INCOMPLETE', missing:[...] }
                 zeros = scores where s === 0 → if any: return { verdict:'FAIL', zero_scored:[...] }
                 h = DIMENSIONS.length / scores.reduce((a,s)=> a + 1/s, 0)
                 return { verdict: h >= args.threshold ? 'PASS' : 'FAIL', harmonic_mean: h, findings, evidence_gaps }
```

Codified anti-patterns (header doctrine comment): no meta-judge agent; no LLM
arithmetic; no Write/Edit for judges; gate on Tier M/L only.

## §D. plan-research-fanout.js Design

```
meta.phases: [ Explore, Synthesize ]
args: { topic (required), lenses = ['codebase-precedent','external-docs','constraints-risks','prior-SPEC-memory'] }
lenses = args.lenses.slice(0, 4)   // hard cap ≤4

Phase Explore   — parallel(lenses) — agent(agentType 'Explore', effort 'medium',
                  label 'explore:<lens>'); 4-element prompt per lens:
                  objective / output-format (fixed headings: ## <lens>,
                  ### findings, ### evidence, ### confidence_and_gaps) /
                  tool guidance (Read+Grep+Glob for codebase lenses; WebSearch+WebFetch
                  for external-docs; memory files for prior-SPEC-memory) /
                  boundaries ("do NOT cover other lenses"; "NONE found" is valid)
Phase Synthesize — nullLenses ≥ 2 → return { verdict:'insufficient_coverage', failed_lenses:[...] } (no agent call)
                  else 1x agent(effort 'high', label 'synthesize:research'):
                  merge per-lens reports; mark every cross-lens contradiction
                  explicitly (### contradictions section); NEVER smooth or
                  average conflicting claims
return { lenses, per_lens_reports, research_md }
```

`research_md` is a STRING returned to the orchestrator; the workflow performs
no file writes — manager-spec/the orchestrator writes
`.moai/specs/<ID>/research.md` outside the workflow (SPEC-artifact ownership +
read-only discipline). Codified anti-patterns: >4 lenses; xhigh on explorers;
in-workflow file writes.

## §E. Rejected Alternatives Summary

| Alternative | Rejected because |
|-------------|------------------|
| Full team removal incl. tmux launchers | Breaks `moai cg` (actively used GLM cost mode) |
| Delete ClaimTask + repro together | Loses P0 regression guard; primitive generically useful |
| Delete `moai workflow lint` subcommand | No-op vs absent tradeoff resolved by cheap repurpose (existing validator) |
| Meta-judge aggregation agent | Smooths dissent; unauditable arithmetic; +1 xhigh call |
| Schema-forced explorer output | Rate-limit brittleness (truncated JSON unusable) |
| Harmonic mean over 3/4 dimensions | Evidence absence ≠ evidence; INCOMPLETE is the honest verdict |
| Template-shipping the workflow pair | `.claude/workflows/` is user-owned; not template-managed |

---
id: SPEC-AGENT-CORE4-001
title: "Plan: agent core four"
version: "0.1.0"
created: 2026-10-18
---

# plan.md — SPEC-AGENT-CORE4-001

Tier L. Development mode: TDD (`.moai/config/sections/quality.yaml` `development_mode: tdd`), with a
characterization-first M1 because the neutral layer replaces behavior that must be preserved byte
for byte. No time estimates; priorities and ordering only. Run phase uses the current agent names
(`manager-develop`, `manager-docs`); the rename takes effect at M8, inside this SPEC.

## §A Decisions most likely to change (reversibility order, reviewed first)

| # | Decision | Why it can change | Reversal cost | Where recorded |
|---|---|---|---|---|
| 1 | Absorbed names retired vs kept as aliases | user-facing; touches every workflow and 4-locale docs | M8 revert + table edit | decision-index Q3 |
| 2 | Resulting roster of nine; Codex count 9 not 12 | the card's "12/12" does not reconcile (acceptance §3) | none (documentation) until M3 | Q2 |
| 3 | Hide mode ladder H1 → H2 → H3 | rests on an unmeasured runtime property | regenerate settings (M6 only) | Q8, Q9 |
| 4 | Neutral frontmatter and capability vocabulary | new data model; every agent depends on it | high after M3 (all renders) | design §1 |
| 5 | Auditor guard on by default | changes a template default | one settings matcher edit | Q6 |
| 6 | Budget split between agents and skills | design values, not measured per agent | rows only; pair ceilings fixed | design §3 |
| 7 | Allowlist policy for historical hits | decides what "zero" means | edit one file | Q10 |

Mechanical work (mirrors, sweeps) is last by construction (M7).

## §B Milestones (execution order = dependency order)

Each milestone is its own commit named `feat(SPEC-AGENT-CORE4-001): M<n> <subject> (card t1649)` with
the line `🗿 MoAI`, staged by explicit pathspec (primary-checkout rules; this work is in a card
worktree). Evidence files under `.moai/reports/t1649/` are gitignored by the blanket rule and are
added with `git add -f <path>` (the established verdict-file convention).

### M0 — Baseline, alone in its commit (priority High)
Re-measure on the then-HEAD (rebase re-pin, `verification-completeness.md` §4): the research §1/§2/§3
numbers, `moai agent size` absence, `prompt_bytes` absence, the 977-file census with per-class
counts, rosterguard result. Output: `.moai/reports/t1649/m0-baseline.md` (force-added). No other
file changes. Reason: baseline-first ordering is witnessed only by the commit graph
(`verification-claim-integrity.md` §2.3).

### M1 — Neutral layer, renderers, generation wiring (priority High)
RED first: `TestNeutralRendersReproduceC2` (red until each of the twelve current definitions is
converted), `TestEmitCheckFailsOnHandEdit`, `TestSizeLibraryCountsComponents`.
Work: `internal/template/agentspec/` schema, loader, shared preamble, `claude-only`/`codex-only`
blocks; Claude renderer; Codex renderer fed by the neutral input (the existing manifest
semantics unchanged); `make agents-emit` / `agents-emit-check` extended with the summary line;
size-accounting library (no CLI verb yet). Convert the twelve current definitions with **zero body
change**; the eight retiring ones carry `retiring: true` and are excluded from the C1/C3p
comparison until M8.
Gate: AC-AC4-005/006 green with `copies=2` (C2, C3); C2 and C3 byte-identical to HEAD
(`git diff --stat ae34f26831 -- internal/template/templates/.claude/agents internal/template/templates/.codex/agents` empty).

### M2 — Carried agents path-free; C1 and C3p generated (priority High)
RED first: AC-AC4-003 and 004 restricted to the four carried agents (`diff -q` per file).
Work: builder-harness, e2e-tester, manager-design, super-advisor bodies lose install-root paths
(skill names instead); regenerate C1/C2/C3/C3p for them. `agents-emit-check` summary becomes
`copies=4`. Gate: the four carried agents' C1 == C2; rosterguard green.

### M3 — Core four and worker (priority High; the size criterion)
RED first: `moai agent size` tests (verb absent), `TestAuditorRubricHeadingsPreserved`,
`TestAuditorCodexContract`, `TestSweepThresholdScalesWithUniverse`, Codex notation/inline-skill
greps.
Work: specs for planner, implementer, documenter, auditor, worker (the `auditor` spec with
`hold_emission: true`, design §2: it renders in memory only until M8 because `auditor` is a
substring of the two live auditor names); the four `moai-agent-*` skill
bodies (≤ 4,000 bytes) and the auditor references (rubric, checklist, shell recipes moved
verbatim); the `moai agent size` verb; Codex permission-contract rows for the new role names;
roster literals and the 65-site registry gain the four emitted names (planner, implementer,
documenter, worker) and the recalibrated sweep threshold lands **in this milestone** so the guard
stays green with a universe of 17 (REQ-AC4-014). Old
agents stay spawnable; workflows still point at them (they flip in M7).
Gate: AC-AC4-001, 002, 009, 011, 012, 015, 019 green; rosterguard green; `make agents-emit-check` green.

### M4 — Guards and records (priority High)
RED first: `TestAuditorWriteGuard`, `TestAgentMemoryCap`, `TestAgentSpawnAuditRecord` (and the
failure demonstrations: a planted out-of-tree auditor write, a planted 8,001-byte index write).
Work: auditor write-path check in the PreToolUse handler (settings matcher widened to
`Write|Edit|MultiEdit|NotebookEdit`; verdict-only name set generated from the specs); memory-cap
check plus the doctor per-agent size/`over_cap` fields; spawn record gains
`phase`/`effort`/`effort_source`/`prompt_bytes`, unattributed stop rows no longer written.
Settings template and its mirror change in the same commit. Gate: AC-AC4-010, 013, 014 green.

### M5 — Replacement paths verified before any removal (priority High; ordering gate)
RED first: `TestLeaderRoleInjectionWithoutManagerLeadFile` (role cores delivered with the agent file
absent), a mapping test that every duty listed in `manager-git.md` § Route B has a command
named in the replacement text.
Work: (a) leader path — role injection + `moai-factory-foreman` marker proven with the
`manager-lead` file absent from the fixture tree; (b) `moai todo` verbs and `/moai:todo` smoke
(`moai todo --help` exit 0) mapped against `manager-todo.md`'s duty list; (c) `moai factory
complete` landing edge mapped against `manager-git.md`'s duties, Route B steps rewritten as
orchestrator-direct steps; (d) `## Tier L Fan-Out` clause added to the run workflow (+ mirror);
(e) `.moai/reports/t1649/m5-verification.md` (force-added) with each path's command and verbatim
output. This commit is the ancestor M8 must have (REQ-AC4-011). Gate: AC-AC4-024 green.

### M6 — Hiding settings and the preload probe (priority Medium)
RED first: the doctor check test (exposed skill → Fail), the generator test (explicit names, no
wildcard).
Work: run the live probe (design §5) for H1; on pass generate H1; on fail try H2; on fail or no
live runtime, ship H3 (inline; no `moai-agent-*` skill published; the generator emits an empty
fragment and the doctor reports `0 exposed`). Evidence: `.moai/reports/t1649/preload-probe.md`
(force-added). Codex: assert nothing published and no ghost entries. Gate: AC-AC4-021, 022 green.

### M7 — Reference flip, in three commits (priority Medium)
Common: `retired_names.yaml` + allowlist; RED first `TestRetiredNamesOnlyInAllowlist` (fails with
the full hit list), `TestOwnershipTrailerAcceptsCoreAndLegacy`, `TestRetiredAgentSpawnRefused`
(refusal code lands inactive-by-roster: it runs whenever the name is in the table, and M8 is when
the definitions disappear). Class order: **M7a** Go production code and registries (profile matrix,
agent-name sets, delegation map, web agent grid and its locale keys, trailer enum, lane
spawn-authority text, hook script contract test); **M7b** rules, skills, workflows, hooks,
configuration (`delegation.yaml`, `llm.yaml`, `workflow.yaml`), `.moai/policies` projection and
every template mirror, in the same commit as their local copy, with the byte delta stated in the
commit body for each always-loaded file changed (`rule-authoring.md`); rule table rows 13-20;
**M7c** tests and fixtures (rename inputs, allowlist the pinned fixtures, no assertion
relaxed, no skip added). Phased names (`plan-auditor`, `sync-auditor`) and command-class names are
rewritten by hand with the table's replacement text; renamed classes may be scripted per file
class. Gate: AC-AC4-017, 018, 020 green; rosterguard green after each of the three commits.

### M8 — Cutover (priority High)
In one commit: emit `auditor` (drop `hold_emission`), delete the eight retired definitions from
C1, C2, C3 and C3p; drop `retiring:` markers; roster
literals, registry sites and counts to nine; refusal live for the eight names (a custom agent file
of the same name outside `.claude/agents/moai/` is honored); `AGENTS.local.md` §2.0 follow-up
recorded as an open item (user-owned file, not edited here). Gate: every AC green except
022-conditional; quality gates (acceptance §6) green; `git merge-base --is-ancestor <M5> <M8>`
exit 0. Landing reaches other lanes only when they absorb the main that carries this card; their
stale spawns degrade to one refusal and a retry (design §6).

## §C Pre-flight (run-phase Section C)

```text
git rev-parse --short HEAD && git branch --show-current
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
go test ./internal/harness/rosterguard/... -count=1
make agents-emit-check embed-manifest-check commands-emit-check
grep -rlE --exclude-dir=.git --exclude-dir=reports --exclude-dir=specs --exclude-dir=node_modules --exclude-dir=archive '\b(manager-lead|manager-todo|manager-git|manager-spec|manager-develop|manager-docs|plan-auditor|sync-auditor)\b' .
```

## §D Constraints

- Template-First: edit `internal/template/templates/` (or the neutral source), run `make agents-emit`,
  then `make build`, then sync local copies; never hand-edit C1/C2/C3/C3p once generated.
- PRESERVE: the semantics of `agents-codex.yaml` (only add role names and rows), the existing hook
  guards, `docs-site/**`, `README*.md`, other SPEC directories, `AGENTS.local.md`.
- Subagent boundary: no `AskUserQuestion` in `internal/hook` or `internal/template` code (grep gate).
- Template neutrality: no SPEC ids, dates, SHAs, or REQ tokens in any template file; new templates
  that cite `factory-dispatch` or `cross-session-messaging` carry `moai:role-rules-required`.
- One writer per tree; stage by pathspec; re-read HEAD before each commit.
- Rule-load budget: report always-loaded count and bytes before and after M7b
  (`rule-loading-budget.md`).

## §E Self-verification deliverables (per milestone, 5-section format)

Claim / Evidence (command + verbatim output) / Baseline-attribution (tree SHA, judging build commit
per `verification-claim-integrity.md` §2.2: build `moai` from the tree and invoke it by path) /
Gaps / Residual-risk. The `moai agent size` and `moai doctor` measurements cite the tree-built
binary, not the installed one.

## §F Risks

| Risk | Mitigation |
|---|---|
| No live runtime for the preload probe | H3 ships (Q9); AC-022 degrades to a guard |
| M7 size | three commits, class order, allowlist with reasons, rosterguard green after each |
| Auditor shrink loses protocol | AC-012 heading containment; references moved verbatim, not rewritten |
| Plan-audit ceiling on an oversized SPEC | split proposal in the plan report: core-4 render+guards (M0-M4), retirement (M5, M7, M8), hiding (M6); milestones are already separable at those seams |
| Roster-guard substring invariant (`auditor` ⊂ `plan-auditor`) | deferred emission of `auditor` to the M8 swap commit; the test is not relaxed |
| Generic names (`planner`, `worker`, `auditor`) may collide with user-defined agents | recorded as decision Q11; `moai agent lint` collision check is a follow-up candidate, not in scope |
| Always-loaded rule growth | M7b net byte delta stated per file; net negative expected (eight names replaced by shorter ones) |
| Stale user-scope copies keep the old agents spawnable | out of scope; the doctor reports them (Gap G-4) |

## §G Anti-patterns

Rewriting rubric text while moving it; hand-editing a generated copy; adding an alias file for a
retired name; widening `agent_type` matching beyond the generated set; treating a green
`[no tests to run]` as a pass; committing the baseline together with the change it measures;
`git add -A`.

## §H Cross-references

`spec.md`, `acceptance.md`, `design.md`, `research.md`, `decision-index.md`;
`.claude/rules/moai/workflow/archived-agent-rejection.md`;
`.claude/rules/moai/development/{verification-completeness,rule-authoring}.md`;
`.claude/rules/moai/core/verification-claim-integrity.md`;
`internal/template/agentemit/agents-codex.yaml`; `internal/harness/rosterguard/`.

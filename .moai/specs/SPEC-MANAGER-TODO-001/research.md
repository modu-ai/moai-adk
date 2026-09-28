# SPEC-MANAGER-TODO-001 — research.md

Plan-phase research. All measurements taken against this worktree (`WT-manager-todo-agent`, develop base `b7ff456b7`), 2026-09-29. Regeneration commands are given per section so any figure can be re-measured.

## A. Measured current state

| Fact | Evidence (command + verbatim output) |
|------|--------------------------------------|
| Agent definition exists, read-only | `Read .claude/agents/moai/mission-governor.md` — frontmatter `tools: Read, Grep, Glob, Skill`, `permissionMode: plan`, `memory: project`; body: "Evaluate a sealed auto-mission snapshot and return one bounded, structured decision. This agent never applies the decision." |
| Three-copy rule | `ls .claude/agents/moai/` shows `mission-governor.md` (1,837 B); `ls internal/template/templates/.codex/agents/moai/` shows `mission-governor.toml` (2,574 B); C2 mirror `internal/template/templates/.claude/agents/moai/mission-governor.md` present |
| Todo state vocabulary | `grep "picked\|queued\|done\|hold" internal/cli/todo.go` — current verbs: `add`(→queued), `list`, `next`, `pick`/`add --pick`, `unpick` (picked→queued recovery), `done`, `undone`, `drop`. `hold` does not exist today |
| Queue store is home SQLite db | `internal/cli/todo.go:197` help text verbatim: "Operate the kanban backlog queue at ~/.moai/db/<project-key>/todo/backlog.db." Line 202 confirms the old `backlog.json` wording is legacy prose only |
| `/moai:todo` is a compatibility alias | `.claude/commands/moai/todo.md` body verbatim: `Use Skill("moai") with arguments: gtd $ARGUMENTS` — the todo slash command routes into the gtd/todo CLI surface, so `--auto` must work through that routing |
| Goal auto-mission wiring | `internal/cli/goal.go:111-273`: `NewAutoMissionCommand` builds `goal --auto` (draft), `approve` (seal), `run` (`--action/--target`, `--governor-receipt` flag described verbatim as "0600 mission-governor decision receipt", `--recommend` described as "compatibility flag only; never grants authority"), `resume`, `revoke` |
| Read-only role roster (Codex) | `internal/cli/codex_audit_mcp.go:199` tool description names the four read-only roles: `plan-auditor, sync-auditor, mission-governor, super-advisor`; `internal/template/agentemit/agents-codex.yaml` lines 83/489 set `role_sandbox: mission-governor: read-only`; rationale block lines 280-285 explains the Codex subagent sandbox inheritance measurement |
| Role fingerprint prose | `internal/cli/codex_role_fingerprint.go:5,16` — manager-lead and mission-governor return their role contracts instead of a nonce |
| delegationmap | `internal/harness/delegationmap/types.go:122` — `"mission-governor": {}` in `retainedCatalog`; comment lines 82-114 record the map went stale once when mission-governor joined the catalog (precedent for this rename's sweep duty) |
| rosterguard | `internal/harness/rosterguard/registry.go` — 10 hits, incl. line 55 ("card t917 landed mission-governor here"), fixture-style reasons enumerating MissingNames, and the whole-file-assertion caveat at 526 |
| internal/web | `git grep -n -i "mission.governor" -- internal/web/` returns **zero hits** — no internal/web surface enumerates the agent name; no change needed there (recorded as the ④-measured-zero result) |
| Jev surface | `scripts/jev/{triage.sh,route.sh,ask.sh}` local-only (no template mirror — confirmed by AGENTS.local.md §29); `internal/cli/todo_jev_finding.go` already carries a todo↔Jev finding surface in the CLI |
| Catalog count surfaces | `CLAUDE.md:63`, `agent-authoring.md:130,147`, `agent-patterns.md:233,268`, `internal/template/retained_agents.go:16`, `catalog.yaml:157,159`, rosterguard, delegationmap — all enumerate 13/12 including mission-governor |
| Parallel factory card | `WT-factory-self-dispatch` mid-run, tip `061614bd5` (dispatch-measured); overlap files include `internal/cli/todo.go` (M1-M3 landed, M4-M7 pending) — seam, not dependency (§D) |
| Liveness measurement precedent | `internal/cli/update_worktree_processes.go` already shells out to `lsof` for process-cwd measurement — the pickup-ownership predicate reuses this pattern; the project lesson "judge a card owner with lsof cwd, not the session registry alone" (feedback_find_card_owner_with_lsof_not_session_registry) makes the two-channel measurement mandatory |

## B. Reference-sweep baseline (M2 checklist)

Baseline command (regenerate verbatim):

```bash
git grep -n -i "mission.governor" -- . ':(exclude).moai/reports' ':(exclude).moai/specs'
```

**Measured 2026-09-29: 262 hits across 95 tracked files.** (Full-catalogue sweep including historical: `git grep -i mission.governor -- .` = 373 hit lines; the `.moai/reports/**` and `.moai/specs/**` remainder — 111 hits — are historical event records and are out of scope per spec.md Out of Scope.) This enumeration is the M2/REQ-MT-005 checklist; each file below carries its disposition column value: **RENAME** (update in place), **FROZEN** (historical fixture, do not edit), or **HIST** (historical prose recording a past event; update only where it describes current behavior).

| File | Lines | Disposition |
|------|-------|-------------|
| `.claude/agents/moai/mission-governor.md` | 2, 15 | RENAME → deleted (REQ-MT-003) |
| `.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | 213 | RENAME (read-only role roster row — see §C conflict) |
| `.claude/rules/moai/development/agent-authoring.md` | 130, 147 | RENAME |
| `.claude/rules/moai/development/agent-patterns.md` | 233, 268 | RENAME |
| `.claude/skills/moai/workflows/goal.md` | 156 | RENAME (goal workflow judgment-role clause) |
| `.moai/project/codemaps/data-flow.md` | 452 | RENAME-at-next-regen (codemaps are generated; note in M2, regenerate via /moai codemaps) |
| `.moai/project/codemaps/docs-truth.md` | 18, 43, 46, 54, 57, 70 | RENAME-at-next-regen (same) |
| `CHANGELOG.md` | 53, 90, 112 | HIST (past release rows) |
| `CLAUDE.md` | 63 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/advanced/agent-guide.md` | en 81,83,86,243 · ja 82,84,87,244 · ko 93,95,98 · zh 81,83,86,243 | RENAME (4 locales) |
| `docs-site/content/{en,ja,ko,zh}/advanced/claude-md-guide.md` | 111 ×4 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/advanced/no-haiku-3tier.md` | en 84,88,118,120,122 · ja/ko/zh 83,87,117,119,121 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/advanced/profile-matrix.md` | en/ja/zh 13,15,41,58 · ko 32,49,51,69 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/claude-code/agentic/sub-agents.md` | en/ko 194 · ja 162 · zh 160 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/core-concepts/what-is-moai-adk.md` | en 260,290,528 · ja 261,291,531 · ko 261,291,530 · zh 261,291,529 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/getting-started/cli.md` | en/ja/zh 454 · ko 456 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/getting-started/faq.md` | 90,120 (en/ja/zh) · 94,124 (ko) | RENAME |
| `docs-site/content/{en,ja,ko,zh}/getting-started/introduction.md` | en 138,149 · ja/ko/zh 137,148 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/multi-llm/model-policy.md` | en 142,163 · ja/zh 97,111 · ko 132,150 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-goal.md` | en/zh 67,69,71 · ja 67,69,71 · ko 71,73,75 | RENAME |
| `docs-site/content/{en,ja,ko,zh}/utility-commands/moai-gtd.md` | 55 ×4 | RENAME |
| `internal/cli/codex_audit_launch_test.go` | 1052, 1054 | RENAME (test expects the role name) |
| `internal/cli/codex_audit_live_test.go` | 393 | RENAME |
| `internal/cli/codex_audit_mcp.go` | 199 | RENAME (tool description) |
| `internal/cli/codex_role_contract_test.go` | 4, 158, 179, 181, 184, 216, 222 | RENAME |
| `internal/cli/codex_role_fingerprint.go` | 5, 16 | RENAME |
| `internal/cli/factory_dispatch_test.go` | 127 | RENAME |
| `internal/cli/goal_mission_test.go` | 56, 464, 567 | RENAME |
| `internal/cli/goal.go` | 273 | RENAME (flag help text "mission-governor decision receipt") |
| `internal/cli/init_codex_only_test.go` | 60 | RENAME |
| `internal/cli/testdata/codex-rollouts-t1171/**` (roles/, roles-other-version/, real/*.jsonl) | 2,4,14-15 / 2,4,15 / recorded rollout lines | FROZEN (t1171 recorded launch fixtures — renaming would falsify the historical record) |
| `internal/graph/gtd_private_test.go` | 325 | RENAME |
| `internal/harness/delegationmap/analyze_test.go` | 114,125,131,132,137,138,141,142 | RENAME |
| `internal/harness/delegationmap/fixturegen_test.go` | 135, 140, 142 | RENAME |
| `internal/harness/delegationmap/testdata/mission_governor_undesignated.jsonl` | 1-9 | RENAME-or-FROZEN (ledger-shaped fixture; run-phase decides rename + filename vs freeze — either must be dispositioned here before close) |
| `internal/harness/delegationmap/types.go` | 82, 96, 107, 112, 122 | RENAME (retainedCatalog + staleness comment) |
| `internal/harness/rosterguard/axis.go` | 7, 19, 21, 79 | RENAME |
| `internal/harness/rosterguard/registry.go` | 55, 177, 179, 193, 206, 208, 223, 241, 526, 529 | RENAME (fixture-reason strings enumerate the role name) |
| `internal/harness/rosterguard/rosterguard_test.go` | 216,217,237,243,256,296,316,322 | RENAME |
| `internal/mission/governance_receipt.go` | 269 | RENAME (receipt contract prose) |
| `internal/mission/governance_receipt_jev_test.go` | 2, 72, 117, 159 | RENAME |
| `internal/mission/governance_receipt_test.go` | 22, 66, 125 | RENAME |
| `internal/mission/governor_test.go` | 21, 27, 31 | RENAME |
| `internal/template/agentemit/agents-codex.yaml` | 83, 282, 489, 518 | RENAME (role_sandbox + addendum anchor — see §C conflict) |
| `internal/template/agentemit/golden_test.go` | 209 | RENAME |
| `internal/template/agentemit/permission_contract_test.go` | 142, 143, 274 | RENAME |
| `internal/template/catalog.yaml` | 157, 159 | RENAME |
| `internal/template/catalog_loader_test.go` | 73 | RENAME |
| `internal/template/catalog_tier_audit_test.go` | 259 | RENAME |
| `internal/template/embed_catalog_test.go` | 64 | RENAME |
| `internal/template/goal_auto_workflow_test.go` | 39 | RENAME |
| `internal/template/retained_agents.go` | 16 | RENAME |
| `internal/template/retained_agents_test.go` | 19 | RENAME |
| `internal/template/templates/.claude/agents/moai/mission-governor.md` | 2, 15 | RENAME → deleted (REQ-MT-003) |
| `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | 213 | RENAME (mirror) |
| `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | 130, 147 | RENAME (mirror) |
| `internal/template/templates/.claude/rules/moai/development/agent-patterns.md` | 233, 268 | RENAME (mirror) |
| `internal/template/templates/.claude/skills/moai/workflows/goal.md` | 156 | RENAME (mirror) |
| `internal/template/templates/.codex/agents/moai/mission-governor.toml` | 2, 4, 14 | RENAME → regenerated (REQ-MT-002; never hand-edited) |
| `internal/template/templates/AGENTS.md.tmpl` | 38 | RENAME |
| `internal/template/templates/CLAUDE.md` | 63 | RENAME (mirror) |
| `README.{md,ko.md,ja.md,zh}.md` | en 374,516 · ko 373,515 · ja 374,516 · zh 373,515 | RENAME (4-locale sync, REQ-MT-018) |
| `reports/moai-dual-harness-{handoff,implementation-status}-20260923.md` | 21 / 15 | HIST (repo-root historical reports) |
| `internal/web/**` | — | measured ZERO hits; no change |

Run-phase MUST re-run the baseline command at M2 start (the sweep decays — line numbers move as parallel cards land on develop) and reconcile against this table; new hits discovered at run time join the checklist with the same disposition rules.

## C. Contract-conflict report (card item ③ — reported, not silently resolved)

**Measured conflict: the read-only role roster.** Today `mission-governor` is one of the four Codex read-only contract roles (`plan-auditor`, `sync-auditor`, `mission-governor`, `super-advisor`; `codex_audit_mcp.go:199`, `agents-codex.yaml` `role_sandbox: mission-governor: read-only`, mirrors of the same four in `moai-mcp-tools-catalogue.md:213`). A pure rename would put **manager-todo** — an agent whose repurposed mission is todo-queue management and dispatch ownership — into a roster whose contract is "read-only sandbox, every MCP server disabled". That is a genuine role-contract conflict, measured on four surfaces, not an inference.

Proposed disposition (design decision D-3): **remove manager-todo from the Codex read-only role roster.** Consequence, reported per the card's "report conflicts back" instruction: the Codex path loses its mission-judgment role — a Codex session can no longer start the sealed-snapshot judgment as a top-level read-only process. The judgment capability survives on the Claude path (goal workflow dispatches manager-todo; `--governor-receipt` schema unchanged, REQ-MT-016), but the *Codex-side* start of that role is a capability removal, and it cannot be preserved without making the dispatch-owning agent codex-sandbox read-only (incorrect) or splitting a second agent file (net catalog addition — violates C-1).

**This removal needs lead confirmation before run-phase.** Recorded as `[NEEDS CLARIFICATION: codex read-only role roster disposition]` in plan.md; if the lead instead elects to retain a read-only roster entry for the judgment sub-role, the roster contract must be rewritten to carry a sub-role sandbox, which is a larger design change than this card.

Secondary conflict check (no conflict measured): the receipt contract (`internal/mission/governance_receipt.go`, `goal.go --governor-receipt`) binds to a schema and path convention, not to the agent name in code; only flag help text and prose mention the old name. Renaming keeps the receipt path `.moai/state/mission/governance/` and schema intact.

## D. t1240 seam (parallel factory card)

`WT-factory-self-dispatch` (SPEC-FACTORY-SELF-DISPATCH-001) is mid-run; its M1-M3 are landed (tip `061614bd5`) and M4-M7 pending, with `internal/cli/todo.go` among its overlap files. This SPEC takes **no dependency** on that unlanded code: `--auto` is designed against the current todo CLI verb surface only (`add/list/next/pick/unpick/done/undone/drop`), creates no factory lease, and claims no factory slot (REQ-MT-013). Both cards' `todo.go` deltas are absorbed by the serial merge window (`moai integration acquire` → merge → re-measure) at run time; a semantic clash there is a run-phase blocker report to the lead, never a forced merge.

## E. Jev boundary sources

The boundary codified in REQ-MT-014/015 already exists as operator doctrine (local instructions §29): three decision grades (grade 1 lead-answers-now; grade 2 model-drafts-lead-confirms; grade 3 never-delegated — completion verdicts, merge approvals, queue mutations, operator gates). Jev remains local-only (`scripts/jev/` has no template mirror; measured: no `scripts/jev` path under `internal/template/templates/`). The CLI already carries a todo↔Jev finding surface (`internal/cli/todo_jev_finding.go`) — the new agent-body boundary and the existing finding surface are complementary, and run-phase must not wire `scripts/jev` into any shipped template.

## F. Template neutrality

Every M1/M2 edit under `internal/template/templates/**` (agent file, rules mirrors, CLAUDE.md/AGENTS.md.tmpl mirrors, catalog.yaml) must carry no forbidden content class: no other card's id, no internal dates, no commit SHAs, no macOS-bias paths, no local-instruction-file references. The rename prose in template mirrors describes the role generically ("todo-queue management and dispatch guidance") — the provenance-bearing narrative stays in `.moai/specs/` and `.claude/` (C1 side is dogfood, mirrors are the neutral source layer). CI guards: `template-neutrality-check.yaml`, `agents-emit-check`, content-leak tests.

## G. Open items carried to plan.md

- `[NEEDS CLARIFICATION: codex read-only role roster disposition]` — §C above.
- Codemaps disposition: refresh in-milestone vs defer to next `/moai codemaps` run (plan decision D-6; either is consistent with REQ-MT-005 provided the disposition is recorded).

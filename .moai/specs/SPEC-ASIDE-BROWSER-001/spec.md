---
id: SPEC-ASIDE-BROWSER-001
title: "Optional Aside browser-CLI integration — documented optional MCP, thin safety-policy skill, explicit-only orchestrator-run e2e toolchain"
version: "0.2.1"
status: completed
created: 2026-10-02
updated: 2026-10-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/template"
lifecycle: spec-anchored
tags: "aside, browser, mcp, e2e, skill, safety-policy, optional-integration"
tier: M
---

# SPEC: Optional Aside browser-CLI integration

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase authoring for card t1439 (Mods design 6/6, Class C). Measurement baseline: tree `4bf547bcad7c155b1e91485921569db709ec3ac2` (worktree t1439, base develop). Aside CLI `1.26.916.1741` observed locally. |
| 0.2.0 | 2026-10-02 | manager-spec | Plan-audit iteration 1 (FAIL 0.75, defects D1-D13) revised. Operator verdicts recorded the same day: Q1 orchestrator-only execution (no subagent ever invokes Aside), Q2 core skill tier, Q3 completely silent fallback. REQ-006 split, REQ-010/012 inverted, fallback-note requirement removed, e2e-tester Aside recipe replaced by one prohibition sentence, catalog expectation corrected, t1434 dependency carried into plan.md, checker anchors made literal. Same measurement tree. |
| 0.2.1 | 2026-10-02 | manager-spec | Delta revision after plan-audit iteration 2 (PASS-WITH-DEBT 0.81), findings N1, N2, N6 only: the Q1 carve-out now reaches every `e2e.md` site by role (REQ-ASB-010, REQ-ASB-012, § 1), development-time `aside` probing is assigned to the orchestrator (§ 3 R-1, A-6), selector count aligned. No requirement added; counts unchanged. |

## 1. Problem — measured shape

Card t1439 asks for an **optional** integration of the Aside browser CLI with three surfaces and a conservative safety posture. Aside drives the operator's own browser, so it acts with the operator's credentials, cookies, and history. That is why the integration is opt-in, read-only by default, and never reachable without an explicit operator act.

Facts measured on baseline tree `4bf547bcad7c155b1e91485921569db709ec3ac2` (full ledger: `acceptance.md` § Evidence ledger; background: `research.md`):

- `moai mcp add|remove|list` already exists (`internal/cli/mcp.go`) with `--command`, repeatable `--args`, `--scope project|user`. The documented line `moai mcp add aside --command aside --args mcp` was executed twice in a scratch project against a binary built from this tree: both exit 0, one `aside` entry `{"args":["mcp"],"command":"aside"}`. **No Go CLI change is needed.**
- The template `.mcp.json` ships exactly `moai` and `context7`; `grep -c -i aside` on it returns `0`. The existing guard `TestMCPNeutralityTemplateShape` already fails on any key outside the allowed set (`mcpAllowedActiveKeys`); no test names `aside` explicitly.
- `aside mcp` exposes two tools: `exec` (browser agent on logged-in sites) and `repl` (Playwright-style JavaScript against open pages, 120 s timeout). `aside exec` takes `--permission ask|guard|full-access` (omitted = Guard; `ask` and `guard` are the same). **`aside repl` has no permission flag at all** — it runs arbitrary page JavaScript in the operator's session — so "read-only" for repl is a discipline the policy must state, not a mode the tool enforces.
- `aside skills install` installs the aside-browser skill into the operator's coding agents. It modifies the operator's agent configuration, which is why MoAI may only *advise* it.
- `/moai e2e --tool` is prose-only (`workflows/e2e.md:40`). The workflow names the e2e-tester as execution owner (`e2e.md:36`), bounds output through that agent (`e2e.md:54`), and carries many sites that either delegate script creation, execution, or recording to the e2e-tester or instruct the missing-toolchain Surface and Install sequence (for example the Phase 2 and Phase 3 delegation lines, the Agent Chain Summary, the Execution Summary, the sequence header at `:108`, and the `--tool` bypass sentence at `:119`). Aside needs an explicit carve-out at every one of them; `plan.md` M3.2 enumerates them by command (15 lines at baseline) so the list cannot go stale.

## 2. Requirements (GEARS)

REQ prefix: `REQ-ASB` (Aside Browser). Five modules (R1 documented optional MCP, R2 policy skill, R3 e2e toolchain, R4 parity, R5 neutrality).

| ID | Pattern | Requirement |
|----|---------|-------------|
| REQ-ASB-001 | Ubiquitous | The policy skill and the docs-site MCP guide (ko, en, ja, zh) SHALL document `moai mcp add aside --command aside --args mcp --scope user` as the optional way to register Aside's MCP server. |
| REQ-ASB-002 | Unwanted | The distributed default MCP configuration (`internal/template/templates/.mcp.json`) and the distributed settings templates SHALL NOT contain an `aside` entry or token. |
| REQ-ASB-003 | Event-driven | When the documented registration command is run any number of times in project or user scope, `moai mcp add` SHALL leave exactly one `aside` entry with `command` `aside` and `args` `["mcp"]` and every unrelated entry unchanged. |
| REQ-ASB-004 | Ubiquitous | The skill `moai-ref-aside-browser` SHALL be a non-user-invocable reference skill that states when Aside is appropriate (an explicit operator request, or a task that needs the operator's logged-in browser session or direct page inspection), when it is not (public pages a plain fetch can read, any CI run, any task another toolchain already covers), and that Aside is operated by the orchestrator alone. |
| REQ-ASB-005 | Unwanted | The skill SHALL prohibit `--permission full-access` in every form (CLI flag, or elevation requested through a prompt), SHALL direct that `--permission` be omitted so the session stays in Guard, and SHALL limit `aside repl` code to navigation, reading, and screenshot capture unless the operator has confirmed a state-changing action. |
| REQ-ASB-006 | Event-driven | When an Aside action would change state in the operator's browser session (submit, send, purchase, delete, post, sign in or out, or any write to a logged-in site), the skill SHALL require explicit operator confirmation, obtained by the orchestrator through its question channel, before the action runs. |
| REQ-ASB-007 | Unwanted | A subagent SHALL NOT invoke Aside (the `aside` CLI in any form, or an Aside MCP tool) under any circumstance; the skill and the e2e-tester definition SHALL each state this prohibition, and a subagent whose task needs Aside SHALL return a blocker report to the orchestrator. |
| REQ-ASB-008 | Unwanted | The skill SHALL prohibit installing Aside or its skills automatically — the only permitted guidance is to advise the operator to run `aside skills install` themselves — and SHALL prohibit placing credentials, cookies, tokens, or session data from the browser in any output, report, commit, or memory entry. |
| REQ-ASB-009 | Capability gate | Where the operator passes `--tool aside` explicitly, the e2e workflow SHALL use Aside as the web toolchain through the `aside repl` CLI; Aside SHALL NOT be auto-detected, SHALL NOT appear among the Phase 0.5 selection options, and SHALL NOT be recommended. |
| REQ-ASB-010 | State-driven | While Aside is the active e2e toolchain, the e2e workflow SHALL assign every Aside step (probe, repl calls, screenshot capture) to the orchestrator, carved out of the e2e-tester's execution ownership at every site in `e2e.md` that delegates script creation, execution, or recording to the e2e-tester, SHALL instruct the orchestrator to load the policy skill before the first Aside step, and SHALL keep Aside output bounded in the orchestrator's context (verbose output redirected to a file, exit code and bounded tail surfaced, path cited). |
| REQ-ASB-011 | State-driven | While `CI=true` is set, the e2e workflow SHALL treat Aside as unavailable even when requested and continue on the platform default toolchain. |
| REQ-ASB-012 | Event-driven | When the Aside probe (`aside --version`) fails, or Aside is excluded (`CI=true`, or a non-web platform class), the e2e workflow SHALL continue on the platform default toolchain with no Aside-specific message, prompt, install attempt, or failure, and no site in `e2e.md` that runs the missing-toolchain Surface and Install sequence, or that delegates test script creation, execution, or recording to the e2e-tester, SHALL apply to Aside. |
| REQ-ASB-013 | State-driven | While Aside is the active e2e toolchain, the journey evidence SHALL be a screenshot captured through `aside repl`, saved under `e2e/` by path and cited by path. |
| REQ-ASB-014 | Ubiquitous | The skill SHALL have a core-tier catalog entry whose hash matches its content; the skill and `workflows/e2e.md` SHALL be byte-identical between template and root copies; the e2e-tester definition SHALL be identical between template and root except for its pre-existing divergence; the committed Codex e2e-tester definition SHALL match its emission from the e2e-tester source; and the repository build SHALL succeed. |
| REQ-ASB-015 | Unwanted | Template content authored by this SPEC SHALL NOT carry card ids, SPEC ids, commit SHAs, or dates, and SHALL NOT position any one programming language as primary. |

### 2.1 User story

The operator (GOOS) sometimes needs an agent to inspect or drive a page that only his logged-in browser can reach. He wants that capability to exist without being on by default, without ever escalating to full access, without any subagent reaching for it, and without `/moai e2e` changing behavior or emitting anything about Aside for anyone who did not ask for it.

## 3. Constraints and assumptions

- **Documented flags are pinned to what was observed** (Aside `1.26.916.1741`): `--permission ask|guard|full-access`, `aside repl`, `aside mcp`, `aside skills install`, `aside --version`. Aside's behavior may vary by version; the skill states no version and, per REQ-ASB-015, no date.
- **Conservative by construction**: every prohibition in REQ-ASB-005 to REQ-ASB-008 has a runnable check with literal anchors (`acceptance.md` § Checker anchors). A prohibition without a check is not adopted.
- **CLI-first, no MCP hard dependency** (`e2e.md:53-56`): the workflow drives Aside through the `aside repl` CLI; the MCP registration is optional documentation and never a prerequisite.
- **Template-First** (`internal/template/CLAUDE.md`) and development mode `tdd` (`quality.yaml`); the process steps live in `plan.md`.
- **Resolved by the operator on 2026-10-02** (decision-index Q1, Q2, Q3):
  - R-1 (was A-2, Q1): Aside is operated by the **orchestrator only**. No subagent, the e2e-tester included, ever invokes `aside` (exec or repl) or an Aside MCP tool. The mechanical anchor is one negative sentence in the e2e-tester definition and the same boundary in the skill. The rule binds the implementation process too: the `aside --version` pre-flight and the screenshot-persistence measurement are run by the orchestrator (the main session), which hands the observed output to manager-develop as run-prompt input; manager-develop and every other subagent never invoke `aside` (`plan.md` § C, M3.0).
  - R-2 (was A-1, Q3): the fallback is **completely silent**. An absent Aside (or `CI=true`) continues on the default toolchain with no Aside-specific message, note, prompt, install attempt, or failure. The run report names the toolchain actually used in the normal way; that is not an Aside note.
  - R-3 (was A-3, Q2): the skill is a **core** skill. The optional-pack placement was rejected because every non-core catalog entry is hidden from slim installs (measured: `TestSlimFS_HidesNonCoreEntries`), while core `moai` and `e2e-tester` would then name a skill default installs do not ship.
- **Still-open assumptions** (each a decision-index row, none settled by the operator):
  - A-4 (Q6): the skill is not added to `delegation.yaml` `domain_skills`, so no mission auto-injects it.
  - A-5 (Q7): docs-site scope is `guides/mcp-server.md` in four locales.
  - A-6 (Q5): `aside repl` can persist a screenshot to a path; unmeasured at plan time, measured first in run phase by the orchestrator (not by a subagent).
  - A-7: these artifacts are written in English per the card instruction.
  - A-8 (Q8): under `--tool aside` the e2e-tester still performs detection and journey mapping (Phases 0 and 1); the orchestrator runs the Aside steps of Phases 2 and 3 itself.

## 4. Tier classification

**Tier M** — 15 touched files counting mirrors and the generated TOML (Tier M range 5-15, at its ceiling), roughly 400-600 new lines. The Q1 verdict did not lower the count: the e2e-tester definition still changes by one sentence, so its root mirror and the regenerated Codex TOML remain, while the recipe block and its subtests were removed. REQ 15 of 16, AC 14 of 16. Surfaces that would exceed the ceiling stay out of scope (§ 7).

## 5. Files affected (all coordinates against `4bf547bcad7c155b1e91485921569db709ec3ac2`)

| # | File | Change |
|---|------|--------|
| 1 | `internal/template/templates/.claude/skills/moai-ref-aside-browser/SKILL.md` | NEW — thin policy skill |
| 2 | `.claude/skills/moai-ref-aside-browser/SKILL.md` | NEW — root mirror, byte-identical to 1 |
| 3 | `internal/template/catalog.yaml` | NEW core-tier skill entry; hashes via `make build` (expected changed set: new entry, `moai`, `e2e-tester`) |
| 4 | `internal/template/templates/.claude/skills/moai/workflows/e2e.md` | `--tool` flag line, execution-owner carve-out, bounded-output carve-out, probe row, Phase 0.5 and CI clauses, both missing-toolchain sites, Tool Matrix row |
| 5 | `.claude/skills/moai/workflows/e2e.md` | root mirror, byte-identical to 4 (baseline `cmp` exit 0) |
| 6 | `internal/template/templates/.claude/agents/moai/e2e-tester.md` | one negative sentence (subagent never invokes Aside) |
| 7 | `.claude/agents/moai/e2e-tester.md` | same sentence; the pair keeps its pre-existing 3-line baseline divergence only |
| 8 | `internal/template/templates/.codex/agents/moai/e2e-tester.toml` | generated from file 6 |
| 9 | `internal/template/mcp_template_neutrality_test.go` | NEW `TestMCPDefaultExcludesAside` |
| 10 | `internal/cli/mcp_test.go` | NEW `TestMCP_Add_AsideDocumentedCommandLine` |
| 11 | `internal/template/aside_skill_policy_test.go` | NEW `TestAsideSkillPolicyAnchors` (checker, anchors, negative controls) |
| 12-15 | `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md` | optional-MCP row, short note, and the "Four" numeral becoming "Five" in heading and sentence |

## 6. Dependencies

No blocking dependency. Card t1434 (Mods 1/6, plugin load-scope probe) measured skills and MCP servers as PLUGIN-OK, capability only; it decided nothing about moving them. The consequences, the verdict's path and stamp, and the reason this SPEC does not depend on it are in `plan.md` § A.3. The existing `moai mcp` CLI is a consumed dependency, not a modified one.

## 7. Out of Scope

### Out of Scope — Aside behavior beyond what was observed

- The permission semantics of the MCP `exec` tool: its argument schema was not measured, so the skill does not claim `--permission` can be passed through MCP.
- `aside exec` agent mode as an e2e driver — only `aside repl` is wired.
- Any Aside account, host, or model option (`--account`, `--host`, `--model`, `--effort`).

### Out of Scope — Other MoAI surfaces

- Any Go change to the `moai mcp` command or the MCP server.
- The default `.mcp.json`, the settings templates, and `delegation.yaml` `domain_skills` (REQ-ASB-002 pins the first two unchanged).
- `.claude/rules/moai/core/settings-management.md`, whose root and template copies already differ (`cmp` reports a difference at byte 18093) — touching it would absorb unrelated drift.
- The docs-site skill-guide table and the README ref-skill lists and counts (`README.md:452`, `:774`): the precedent ref-skill commit touched them, but adding them would exceed the Tier M file ceiling, and no guard forces them (catalog totals are derived from `catalog.yaml`, not hard-coded).
- `docs-site` `utility-commands/moai-e2e.md` in four locales and the `/moai` router entries; the workflow file is the e2e documentation of record.
- Any Aside-specific message, note, or telemetry in the e2e run report (the verdict is silence).

### Out of Scope — Operating Aside

- Running Aside in CI, installing Aside or its skills, or any automatic skill install.
- Moving the skill into a plugin (follow-up if a later card moves skills and MCP servers together).

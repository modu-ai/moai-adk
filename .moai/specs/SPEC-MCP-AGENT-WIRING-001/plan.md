# plan.md — SPEC-MCP-AGENT-WIRING-001

Milestones are ordered by decision reversibility: the grant matrix (a judgment call with a blast-radius consequence) is settled first, the mechanical mirror sweep last.

## §A. Context

### A.1 Base

All citations read at base commit `ed70e4354`.

### A.2 The 17 tools and their write-capability

Registered by `registerMoaiMCPTools` (`internal/cli/mcp_server.go:105`) — count verified at base: `grep -c "mcp.NewTool(" internal/cli/mcp_server.go` = **17**.

| Group | Tools | Write-capable? |
|---|---|---|
| M1 core | `session_list`, `goal_status`, `spec_progress`, `verify_trend`, `spec_audit`, `spec_drift`, `audit_cache` | read-only (carry `WithReadOnlyHintAnnotation(true)`) |
| M1 core | `goal_arm`, `verify_snapshot` | **write-capable** |
| M2 codex | `codex_audit`, `codex_setup` | read-only |
| CODEX-PHASE2 | `codex_job_status`, `codex_job_result` | read-only |
| CODEX-PHASE2 | `codex_task`, `codex_job_cancel` | **write-capable** |
| M3 GLM | `glm_audit`, `audit_multi` | read-only |

### A.3 Existing precedent

`plan-auditor.md:7` and `sync-auditor.md:9` already carry `mcp__moai__audit_multi` in their `tools:` CSV. This SPEC extends an established pattern rather than introducing one.

### A.4 Agent `tools:` lines at base

| Agent | File:line | Already has MCP? |
|---|---|---|
| manager-spec | `manager-spec.md:9` | no |
| manager-develop | `manager-develop.md:9` | no |
| manager-docs | `manager-docs.md:9` | no |
| manager-git | `manager-git.md:8` | no |
| manager-design | `manager-design.md:10` | no |
| manager-lead | `manager-lead.md:8` | no (carries `Agent`) |
| plan-auditor | `plan-auditor.md:7` | `audit_multi` |
| sync-auditor | `sync-auditor.md:9` | `audit_multi` |
| builder-harness | `builder-harness.md:7` | no |
| super-advisor | `super-advisor.md:13` | no |
| e2e-tester | `e2e-tester.md:15` | no |
| Explore | (built-in, no file) | n/a |

## §B. The grant matrix (M1)

Every row states what is added and why; every omission is stated as a decision.

### B.1 Agents receiving grants

| Agent | Grant | Justification |
|---|---|---|
| `manager-spec` | `spec_progress`, `spec_audit`, `spec_drift` | Owns plan-phase artifact authoring. Needs to see the existing SPEC set (dedup, ID uniqueness — Step 2 of its own workflow), and to read lifecycle/drift state when authoring `depends_on` or an amendment. All three read-only. |
| `manager-develop` | `verify_snapshot`, `verify_trend`, `goal_status` | `verify_snapshot` is the **one write-capable grant this SPEC makes**, and it is the agent's existing duty: §E attribution (`manager-develop-prompt-template.md` § Section E) requires recording command + observed output against a HEAD-keyed snapshot — that IS `verify.RecordCheck`, which `verify_snapshot` wraps. `verify_trend` reads the same store. `goal_status` lets it see whether a run-phase goal is armed before reporting completion. |
| `manager-docs` | `spec_progress`, `spec_audit` | Sync-phase close needs the SPEC set and the lifecycle/drift view to confirm a clean close. Read-only. It owns `§E.4` and the `sync_commit_sha`, but those are file writes, not MCP writes. |
| `plan-auditor` | `spec_audit`, `spec_drift`, `codex_audit` (keeps `audit_multi`) | Its verdict rests on artifact state and cross-model second opinions. `codex_audit` is read-only (a review call, not a task spawn). `codex_task` is deliberately **omitted** — an auditor that can spawn a codex task is no longer only auditing. |
| `sync-auditor` | `verify_trend`, `audit_cache`, `glm_audit` (keeps `audit_multi`) | 4-dimension scoring wants the verification history (`verify_trend`), the plan-audit PASS cache for consistency scoring (`audit_cache`, read-only lookup/compute-hash), and the GLM backend for the cross-model read. `verify_snapshot` is **omitted**: an auditor that can write the verification baseline it scores against is grading its own paper. |
| `manager-lead` | `session_list`, `goal_status` | Tier L coordination across worktree-isolated leaf workers needs to see concurrent sessions (`session_list` — directly relevant to the pre-spawn race check) and whether a goal is armed. Both read-only. |
| `super-advisor` | `spec_audit`, `verify_trend` | Consults across all phases and has no Write/Edit at base (`super-advisor.md:13`), so read-only state access fits its shape exactly: a diagnosis grounded in measured state beats one grounded in the prompt's summary. This is the matrix's most discretionary row — the agent functions without it. |

### B.2 The `manager-lead` `goal_arm` omission — a deliberate departure from the proposed matrix

The Epic brief proposed `manager-lead` → `session_list, goal_status, goal_arm`. **`goal_arm` is omitted here**, and the reason is not caution for its own sake.

`goal_arm` changes the session's **termination condition**: an armed goal causes the `stop-goal` Stop-hook evaluator to block turn-end until the condition holds. Per `.claude/rules/moai/workflow/goal-directive.md` § Goal-Presentation Timing, the goal is presented at the **Implementation Kickoff Approval** gate and armed by the orchestrator only after that gate passes. Granting `goal_arm` to a subagent creates a path where a coordinator re-arms or changes the session's termination condition without the gate — which is the one decision the gate exists to hold. `goal_status` (read) gives `manager-lead` everything it needs to coordinate around an armed goal; arming stays with the orchestrator.

If the owner wants `goal_arm` granted anyway, the change is one CSV entry and one AC row — but it should be a stated decision rather than an inherited default.

### B.3 Agents receiving no grant — each a documented decision

| Agent | Decision | Reason |
|---|---|---|
| `manager-git` | **none** | Its description scopes it to commits, branches, PRs, merges, releases and explicitly excludes implementation, testing, architecture, documentation, and audits (`manager-git.md:3-8`). Every moai MCP tool reads or writes SPEC/verification/audit state — none of it is git state. It receives its SPEC context from the orchestrator's spawn prompt. |
| `manager-design` | **none** | D1-D5 design-phase pipeline with a dedicated `DesignSync` tool (`manager-design.md:10`). No moai state tool serves the design surface. |
| `e2e-tester` | **none** | Owns journey scripting and artifacts under project-local `e2e/`. `verify_snapshot` would be the tempting grant — e2e runs are verification — but it is write-capable, and routing e2e results into the verification baseline means a flaky run mutates the snapshot that later attributes other agents' claims. E2E results are surfaced to the orchestrator, which records them. |
| `builder-harness` | **none** | Builds artifact metadata (frontmatter, manifests, dispatch tables) including MCP-server scaffolding (`builder-harness.md:3-6`). It authors MCP configuration; it does not consume moai runtime state. |
| `Explore` | **n/a** | Anthropic built-in with no MoAI file; its tool set is not editable here. |

Five of twelve agents receiving nothing is the expected shape, not an oversight: the moai tool surface is SPEC-lifecycle and verification state, and five of the retained agents work in domains (git, design, e2e, artifact-meta, generic exploration) that do not touch it.

### B.4 Write-capable grant summary

Exactly **one** write-capable tool is granted by this SPEC: `verify_snapshot` → `manager-develop`. `goal_arm`, `codex_task`, and `codex_job_cancel` are granted to **nobody**.

## §C. REQ-B-6 (model/effort SSOT) — already complete at base; no Go code change is part of this SPEC

The Epic brief asked to "extend the `ResolveAgentModelEffort` SSOT beyond `glm_audit` so model version and effort selection is uniform across the tool surface." Measured against the tree (re-verified at base in iter-2), that extension is **already done, and is already structurally guarded**. Stating this plainly rather than inventing work to match the brief:

Only four of the 17 tools invoke a model at all — `codex_audit`, `codex_task`, `glm_audit`, `audit_multi`. The other 13 are state readers with no model in the path. All four already resolve through the SSOT, and **all four are already covered by an existing structural guard**:

| Tool | Resolution site | Existing guard |
|---|---|---|
| `glm_audit` | `internal/cli/mcp_glm.go:134` | `mcp_glm_test.go:227` asserts SSOT-only |
| `codex_audit` | `internal/cli/mcp_codex.go:199` | `TestMCPAudit_NoDirectFrontmatterRead` (`mcp_audit_test.go:148`) scans `mcp_codex.go` and forbids direct frontmatter/override reads |
| `audit_multi` (convergence) | via the same resolvers | `mcp_convergence_test.go:518-526` forbids direct reads |
| `codex_task` | `openCodexSession` (`mcp_codex.go:527`) → `openCodexSessionOn` (`mcp_codex.go:536`) → `threadParams["model"] = me.Model` (`mcp_codex.go:566`); model rides `thread/start` | **Same guard** — `TestMCPAudit_NoDirectFrontmatterRead` (`mcp_audit_test.go:148`) scans `files := []string{"mcp_glm.go", "mcp_audit.go", "mcp_codex.go"}`; the entire codex session path lives in `mcp_codex.go`, so the path is structurally covered. `internal/cli/codex_task.go:222` only delegates to `openCodexSessionOn` and performs no independent model resolution (grep `AgentOverrides|agentfm|ReadFrontmatter|ParseFrontmatter` = empty). |

The iter-1 plan had identified the codex session path as "the one real gap" needing a new structural guard. That premise was **wrong**: the codex session path lives entirely in `mcp_codex.go`, which is already in the `TestMCPAudit_NoDirectFrontmatterRead` scan list, and `codex_task.go` adds no independent resolution of its own. The Epic brief's "extend the SSOT" is therefore satisfied at base; no Go code change is part of this SPEC, and no new guard test is added. REQ-B-6 records the finding; it does not deliver a migration.

The web console side is already guarded too — `internal/web/mcp_audit_surface_test.go:76` asserts `internal/web` imports rather than redefines the resolver.

## §D. Milestones

### M1 — Grant matrix applied to the 7 source agent files

Edit the `tools:` CSV of `manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`, `sync-auditor`, `manager-lead`, `super-advisor` under `.claude/agents/moai/`. Append only; preserve existing entries and the CSV form.

### M2 — Template mirror + `make build`

Mirror the 7 edited files into `internal/template/templates/.claude/agents/moai/`, run `make build`, commit the regenerated catalog alongside. (The iter-1 plan's separate M2 Go-test milestone and the M3 mirror step were collapsed: the Go-test milestone was redundant because the codex session path is already structurally guarded — see §C — so the mirror step renumbers from M3 to M2.)

## §E. Anti-patterns

- **AP-B-1 — `tools:` as a YAML array.** It is a CSV string. Only `skills:` takes an array.
- **AP-B-2 — Granting a tool "because the agent might want it".** Every entry in §B.1 names the specific duty it serves. An unjustifiable grant is an omission.
- **AP-B-3 — Editing the local `.claude/` copy without the template mirror.** CI fails on the parity guard; the local edit is only half the change.
- **AP-B-4 — Inventing SSOT migration work (or a redundant guard) to match the brief's framing.** §C measured the actual state: all four model-invoking tools resolve through the SSOT at base AND the existing `TestMCPAudit_NoDirectFrontmatterRead` already scans `mcp_codex.go`, where the entire codex session path lives. The deliverable here is **nothing on the Go side** — no migration, no new guard test, no code change. A SPEC that adds a redundant codex guard test would be re-asserting what `mcp_audit_test.go:148` already asserts.
- **AP-B-5 — Adding `Agent` to a leaf worker.** The depth-2 seal (`manager_lead_depth_test.go`) is untouched by this SPEC; MCP tools are not spawn capability.

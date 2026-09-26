---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Acceptance — subagent model/effort inheritance"
created: 2026-09-26
---

# Acceptance — SPEC-AGENT-MODEL-INHERIT-001

## §D AC Matrix

- **AC-AMI-001** (maps REQ-AMI-001): Given the run phase starts, When progress.md §E.2 is read, Then it names the develop commit carrying the t1175 merge and a table of re-measured inventory counts taken after absorbing develop, dated before the first edit commit; and if the t1175 merge was absent, no edit commit exists on this branch after the measurement.
- **AC-AMI-002** (maps REQ-AMI-002): Given the run phase starts, When progress.md §E.2 is read, Then it contains the `git diff --name-only <merge-base>...WT-role-naming-docs` command, its output, and the intersection with this SPEC's touch set; and if that intersection held a file already changed on that branch, no edit commit exists on this branch after the measurement.
- **AC-AMI-003** (maps REQ-AMI-004): Given the change is complete, When `grep -rnE '^(model|effort):' .claude/agents internal/template/templates/.claude/agents` runs and `/moai:harness` generation is exercised by its generator test in a temp project, Then the grep prints nothing and the generated specialist file and manifest carry no `model`/`effort` value.
- **AC-AMI-004** (maps REQ-AMI-005): Given the change is complete, When `grep -nE '^(model|model_reasoning_effort)' internal/template/templates/.codex/agents/moai/*.toml` runs and `make agents-emit-check` runs, Then the grep prints nothing and `make agents-emit-check` exits 0.
- **AC-AMI-005** (maps REQ-AMI-006): Given a MoAI agent file without `effort:`, When the agent linter runs over `.claude/agents/moai/`, Then it reports no finding with rule id LR-03 or LR-12.
- **AC-AMI-006** (maps REQ-AMI-007): Given a test project with a user agent declaring `model: opus` and `effort: high`, When `moai update` runs and the agent linter runs, Then the file is byte-identical afterwards and the linter reports no error for it; and an existing harness v4 manifest carrying specialist `model`/`effort` passes the v4manifest validator unchanged, as does one carrying neither.
- **AC-AMI-007** (maps REQ-AMI-008): Given the change is complete, When `grep -rnE 'Agent\([^)]*(model|effort) *:' .claude internal/template/templates/.claude` and `grep -nE '\b(model|effort) *:' .claude/workflows/*.js internal/template/templates/.claude/workflows/*.js` run, Then the first prints nothing outside `.claude/skills/moai-foundation-cc/reference/**` (Claude Code feature documentation) and the second prints no `agent()` option line.
- **AC-AMI-008** (maps REQ-AMI-009): Given a built binary, When `moai model profile --json` runs, Then it exits non-zero with an unknown-command error; and `grep -rnE 'ResolveAgentModelEffort|DefaultProfileMatrix|ResolveHarnessAgentModelEffort' --include=*.go internal cmd pkg` prints nothing.
- **AC-AMI-009** (maps REQ-AMI-010): Given a PreToolUse payload for an `Agent` spawn with a `model` argument, When the hook handles it in a temp project, Then the decision carries no model-based advisory or deny and `.moai/logs/agent-model-audit.jsonl` does not exist afterwards.
- **AC-AMI-010** (maps REQ-AMI-011): Given a `workflow.yaml` containing `agent_model_guard: {enabled: true}`, When configuration loads, Then loading succeeds and a spawn payload is handled identically to one loaded without the key.
- **AC-AMI-011** (maps REQ-AMI-012): Given `moai web` serving a test project, When the settings page renders and a POST carrying former agentfm field names is sent, Then the settings shell renders no agent-settings tab, no agent model/effort or profile-selector control appears in the HTML, the former agentfm handler path is unrouted, and no agent file or `llm.yaml` byte changes.
- **AC-AMI-012** (maps REQ-AMI-013): Given `moai web` serving a test project, When a preference profile is created, renamed, and deleted, and the main-session effort control is saved, Then each operation returns the same status code and writes the same files as the characterisation test recorded in M1.
- **AC-AMI-013** (maps REQ-AMI-014): Given the change is complete, When `grep -nE '^\s*(profile|profiles|harness_agents|agent_overrides|performance_tier):' internal/template/templates/.moai/config/sections/llm.yaml` and `grep -nE '^\s*(workflow_agents|model_routing|model_routing_profiles):' internal/template/templates/.moai/config/sections/workflow.yaml` run, Then both print nothing.
- **AC-AMI-014** (maps REQ-AMI-015): Given test projects for each row of design.md §C, When `moai update` runs, Then the resulting `llm.yaml` and `workflow.yaml` carry none of the REQ-AMI-014 keys nor `agent_model_guard`, the backup carries the original files, and the update output names each removed key.
- **AC-AMI-015** (maps REQ-AMI-016): Given `moai init` in a temp dir with a main-session model policy chosen, When init completes, Then `llm.yaml` carries no `profile:` key and the main-session setting is persisted where it was before this SPEC.
- **AC-AMI-016** (maps REQ-AMI-017): Given `moai init --profile high` and `moai update --profile low`, When each runs in a temp project, Then each exits 0, each prints a deprecation warning naming subagent inheritance, and neither writes a `profile:` key.
- **AC-AMI-017** (maps REQ-AMI-018): Given the M1 characterisation tests for main-session model/effort resolution, When they run after M7, Then they pass unmodified.
- **AC-AMI-018** (maps REQ-AMI-019): Given the M1 characterisation tests for GLM alias mapping and session reasoning, When they run after M7, Then they pass unmodified.
- **AC-AMI-019** (maps REQ-AMI-020): Given a project with `workflow.audit.codex` and `workflow.audit.glm` pins, and another without them, When the codex and GLM tools resolve their model, Then the first returns the pins and the second returns the backend defaults.
- **AC-AMI-020** (maps REQ-AMI-021): Given the change is complete, When the roster SSOT is grepped and rosterguard tests run, Then exactly one Go literal lists the retained agents, it carries no model or effort field, and `go test ./internal/harness/rosterguard/...` passes.
- **AC-AMI-021** (maps REQ-AMI-022): Given design.md §D, When `grep -c '\[HARD\]'` runs on each touched rule file before and after, Then each file's delta equals the number of rows the table marks "Removed" with a `[HARD]` marker for that file.
- **AC-AMI-022** (maps REQ-AMI-023): Given the change is complete, When each mirrored file pair is compared, Then the template side changed in the same or an earlier commit than the local side, and every local-only file edited has no template counterpart created.
- **AC-AMI-023** (maps REQ-AMI-024): Given the change is complete, When the template-neutrality tests run (`go test ./internal/template/ -run 'Neutrality|InternalContentLeak'`), Then they pass.
- **AC-AMI-024** (maps REQ-AMI-025): Given the change is complete, When `go test -run TestAlwaysLoadedTokenBudget -v ./internal/config/` runs, Then it passes and progress.md records the before headroom (61 at `d6992e3a0`, re-baselined after t1175) and the after headroom.
- **AC-AMI-025** (maps REQ-AMI-026): Given the change is complete, When `grep -rlE 'model profile|profile matrix|agent_overrides|harness_agents|agent_model_guard|[Pp]er-[Ss]pawn|agent-model-audit' docs-site/content` runs and the docs-site verify recipe (hugo build, four-locale section parity, redirect check for removed pages) runs, Then the grep prints nothing and the verify recipe exits 0.

## §D.1 Edge cases

- A user `llm.yaml` whose `agent_overrides` holds a codex model id for the audit key: after update the audit falls back to pin or backend default; the update report names the dropped override.
- A user who exports `CLAUDE_CODE_SUBAGENT_MODEL`: subagents stay pinned; documented residual (spec.md §F), not a failure.
- An agent file with `effort:` but no `model:` (user-authored): untouched (AC-AMI-006).

## §D.2 Quality gates

- Affected-package `go test`, `go vet`, `golangci-lint run` clean on changed packages.
- `make build`, `make agents-emit-check`, `make commands-emit-check` exit 0.
- CI green on the develop push that carries this card.

## §D.3 Definition of Done

All AC-AMI-001..025 pass with evidence in progress.md §E.2; Q1–Q6 answers recorded in
progress.md §E.1 (resolved 2026-09-26); design.md §D matches the landed diff.

# Acceptance Criteria — SPEC-MODEL-PROFILE-MATRIX-001

Format: Given-When-Then. Each AC maps to one or more REQ-MPM-xxx. "Green CI guards" means existing `go test ./internal/...` config + template + spec-lint guards stay passing.

## §D. Acceptance Criteria Matrix

### Config schema + migration

**AC-MPM-001** (REQ-MPM-001/002/008)
- Given a `llm.yaml` with `profile: low`
- When the config loads
- Then `LLMConfig.EffectiveProfile()` returns `low`; and given `profile:` absent but `performance_tier: high`, it returns `max`; and given both absent, it returns `medium`; and given `profile: bogus`, validation returns an error naming the value and the set {max, medium, low}.

**AC-MPM-002** (REQ-MPM-003/004)
- Given a legacy `llm.yaml` carrying `plan_type: subscription`, `performance_tier: max`, and a `claude_models:` block
- When the config loads
- Then the load succeeds with no error, `plan_type` and `claude_models` are ignored, and the effective profile resolves to `max`.

**AC-MPM-003** (REQ-MPM-005)
- Given a project whose `llm.yaml` still contains `plan_type` + `claude_models`
- When any persistence write updates the `llm` section (init/update/web save)
- Then the rewritten `llm.yaml` contains no `plan_type` key and no `claude_models` block, and contains `profile:`.

**AC-MPM-004** (REQ-MPM-006/007)
- Given `llm.agent_overrides: { manager-develop: { model: opus, effort: xhigh } }`
- When resolution runs
- Then `manager-develop` resolves to `{opus, xhigh}`; and given an override for a non-catalog agent name or an out-of-enum model/effort, validation returns an error naming the offending agent and field.

### Profile matrix (Matrix A)

**AC-MPM-005** (REQ-MPM-009/010/011/012)
- Given `profile: max` and no overrides
- When each agent is resolved
- Then results equal Matrix A max column exactly: manager-spec/plan-auditor/sync-auditor → {fable, medium}; manager-develop → {fable, low}; super-advisor → {fable, medium}; manager-design/builder-harness/e2e-tester → {opus, high}; manager-docs → {sonnet, medium}; manager-git → {sonnet, low}.

**AC-MPM-006** (REQ-MPM-012)
- Given `profile: medium` and `agent_overrides: { manager-spec: { model: opus, effort: xhigh } }`
- When manager-spec and plan-auditor are resolved
- Then manager-spec → {opus, xhigh} (override wins) and plan-auditor → {opus, high} (medium group cell, unaffected by the manager-spec override).

**AC-MPM-007** (REQ-MPM-013)
- Given any profile
- When `Explore` (or any agent with no group) is resolved
- Then the resolver returns the `inherit` sentinel and no concrete model alias.

**AC-MPM-008** (REQ-MPM-010/033)
- Given the shipped template `llm.yaml`
- When it is read
- Then it carries an `llm.profiles` block with all three profiles × six agent-group keys whose values equal Matrix A, and the local `.moai/config/sections/llm.yaml` carries the same schema shape.

### Selection — init

**AC-MPM-009** (REQ-MPM-014)
- Given `moai init` run interactively
- When the model-routing question(s) are presented
- Then exactly one appears — the profile selection {max, medium, low} — and no api/subscription (plan_type) question is presented.

**AC-MPM-010** (REQ-MPM-015/016)
- Given `moai init --profile medium`
- When init completes
- Then the flag value overrides any wizard answer, is closed-set validated (invalid value → error naming {max, medium, low}), and `llm.profile: medium` is persisted.

**AC-MPM-011** (REQ-MPM-017)
- Given `moai init --plan-type api` (retired flag)
- When init runs
- Then the CLI does not write a `plan_type` key (the flag is rejected as unknown or ignored), and the produced `llm.yaml` has no `plan_type`.

### Selection — web console

**AC-MPM-012** (REQ-MPM-018/019)
- Given the `moai web` Model Policy page
- When it renders
- Then a profile selector {max, medium, low} is present, its selection persists to `llm.profile`, and no plan_type selector or read-only plan_type value appears anywhere on the page.

**AC-MPM-013** (REQ-MPM-020)
- Given `profile: low` active
- When the Model Policy page renders the per-agent grid
- Then each retained agent shows its resolved `{model, effort}` from Matrix A low column (e.g. manager-spec → opus/low, super-advisor → opus/high, manager-git → sonnet/low), derived from the single profile-matrix source (no duplicated literal).

**AC-MPM-014** (REQ-MPM-021/022)
- Given the per-agent override editor
- When a valid override (e.g. manager-develop → opus/xhigh) is submitted
- Then it persists to `llm.agent_overrides`; and when an override with an out-of-enum model or effort is submitted, the handler returns a client error and persists nothing.

### Consumption — runtime injection + reconciliation

**AC-MPM-015** (REQ-MPM-023/024)
- Given a fresh `moai init` (any profile) or `moai update`
- When the shipped agent files under `.claude/agents/moai/` are inspected afterward
- Then their frontmatter `model:` is `inherit` (unmutated) and their `effort:` equals the `agent-authoring.md § Effort-Level Calibration Matrix` value — init/update performed no frontmatter model/effort mutation.

**AC-MPM-016** (REQ-MPM-025/026)
- Given `profile: max` active
- When `moai model profile --json` is run
- Then it emits the resolved per-agent `{model, effort}` for the active profile (Matrix A max); the emitted **model** is the value the orchestrator injects as a per-spawn runtime arg (per DECISION-001, `model`-only), and the emitted **effort** is carried for display + GLM-overlay input + documented intent (NOT a per-spawn override for a named subagent); and the orchestrator spawn-guidance (`model-policy.md`) references this resolver as the runtime-arg **model** injection source (no reliance on a mutated frontmatter model pin).

**AC-MPM-017** (REQ-MPM-027)
- Given `profile: max` (manager-spec group cell = fable/medium) and manager-spec frontmatter effort = xhigh
- When the resolver reports manager-spec's resolved profile cell alongside the frontmatter effort
- Then it reports profile effort `medium` as **documented intent** while the frontmatter effort stays `xhigh`; for a named-subagent spawn the **model** `fable` is the injected runtime arg and the effective spawn **effort** remains the frontmatter `xhigh` (no per-spawn effort override for named subagents, DECISION-001); and this documented divergence produces no validation or lint error.

**AC-MPM-018** (REQ-MPM-028)
- Given the post-implementation tree
- When `grep -rn "tierProfiles\|ApplyTierProfile\|EffectivePlanType\|plan_type" internal/ --include=*.go | grep -v _test.go` is run
- Then there are no production references resolving model/effort through plan_type (only migration-tolerance strip logic and its tests may remain).

### GLM interaction

**AC-MPM-019** (REQ-MPM-029/030/031)
- Given `profile: max` and a GLM backend (`team_mode: glm`)
- When `moai model profile --json` resolves manager-develop
- Then the model alias `fable` maps through `llm.glm.models` (→ glm-5.2) and effort `low` collapses to `thinking-off` — except manager-develop's coding-max override lifts it to `reasoning-max`; and no GLM-specific profile column exists and the GLM env activation flow is unchanged.

### Template / build / tests / boundary

**AC-MPM-020** (REQ-MPM-032/034/035)
- Given the run-phase implementation
- When `make build` and `go test ./internal/config/... ./internal/template/... ./internal/cli/... ./internal/web/...` run
- Then the build succeeds and tests pass, including: new-schema load, legacy-schema migration load, profile→agent resolution with override precedence, retired-key round-trip strip, init flag+wizard precedence, and web selector+override persist/validate.

**AC-MPM-021** (REQ-MPM-036)
- Given run-phase completion
- When the documentation-impact list is produced
- Then it enumerates README profile-count references, docs-site model-routing pages, and `model-policy.md` sections describing the retired plan_type axis, explicitly flagged as follow-up sync-phase scope (no doc prose edited in this SPEC's run phase).

**AC-MPM-022** (REQ-MPM-037/038)
- Given the post-implementation diff
- When it is reviewed
- Then `.moai/config/sections/delegation.yaml`, the GLM env-writing code paths, and `workflow.yaml` `model_routing_profiles` are unmodified, and no agent-frontmatter effort *value* was re-authored (only `ApplyTierProfile` mutation removed).

**AC-MPM-023** (REQ-MPM-039)
- Given any documentation-impact note or code comment about profile-injected GLM effort
- When it is reviewed
- Then it states "implemented + wired, live wire-effectiveness pending" and never claims verified GLM wire-effectiveness.

**AC-MPM-024** (HaikuResidualRule + CI guards)
- Given the new profile matrix + template `llm.yaml`
- When `go test ./internal/spec/...` (HaikuResidualRule) and the config CI guards (`YAML_SECTION_NO_LOADER`, `CONFIG_STRUCT_YAML_MISMATCH`) run
- Then all pass — the matrix contains zero haiku references and the struct↔YAML symmetry holds for the new fields.

### Web save-path frontmatter mutation retirement

**AC-MPM-025** (REQ-MPM-040)
- Given a project whose agent `.md` frontmatter is at `model: inherit` and a `moai web` settings save that changes the model-routing selection (formerly `performance_tier`, now `profile`)
- When the save is submitted and `applyPerfTierEdits` (`internal/web/agentfm.go`) runs
- Then no agent `.md` frontmatter `model:`/`effort:` value is mutated (the `template.ApplyTierProfile` call at `agentfm.go:108` is retired), the selection persists to `llm.profile` (+ any `llm.agent_overrides`) in `llm.yaml`, and `grep -rn "ApplyTierProfile" internal/web/ --include='*.go' | grep -v _test.go` returns no production call.

## §D.1 Must-Pass (blocking) ACs

AC-MPM-001, AC-MPM-002 (migration — no data-loss / no load error), AC-MPM-005 (Matrix A fidelity), AC-MPM-015 + AC-MPM-025 (no frontmatter mutation at any of the 4 `ApplyTierProfile` call sites — the `[1m]`-safety core, init/update + web save-path), AC-MPM-020 (build+tests green), AC-MPM-024 (existing guards green).

## §D.2 Definition of Done

- All 25 ACs verified with observed command output (not asserted).
- `moai spec lint` + `--strict` clean for this SPEC directory.
- Both former `[NEEDS CLARIFICATION]` markers are resolved as DECISION-001 (Agent-tool per-spawn capability = model-only) and DECISION-002 (default profile = medium) in plan.md §E — no open clarification markers remain at run-phase entry.
- Documentation-impact list produced (REQ-MPM-036) and handed to sync phase.

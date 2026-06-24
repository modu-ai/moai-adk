# Implementation Plan — SPEC-V3R6-RULES-CATALOG-SCRUB-001

## §A Context

This is a **content scrub** SPEC: mechanical removal of 12 archived-agent names where they survive as live spawn targets / live examples / live hook actions across `.claude/rules/`, with identical edits applied to the `internal/template/templates/.claude/rules/` mirror. Tier M, ~14 defect groups across ~18 files. No Go code. No new lint rule. The work is high-volume but low-conceptual-complexity EXCEPT for D14 (design-pipeline `expert-frontend`), which carries genuine design tension and is scoped as SHOULD.

The dominant risk is **over-scrubbing**: archived names appear legitimately in 4 PRESERVE files (`archived-agent-rejection.md`, `NOTICE.md`, `agent-patterns.md`, `spec-frontmatter-schema.md`) and as `role_profile` tokens. Every milestone must distinguish a live-reference defect from a documentary/role-profile mention.

## §B Known Issues / Constraints

- **Template-First (CLAUDE.local.md §2)**: all 18 target files verified to have a template mirror. Each edit must be applied to BOTH the deployed `.claude/rules/moai/...` and the template source `internal/template/templates/.claude/rules/moai/...`. The template tree is embedded at compile time via `//go:embed all:templates` (`internal/template/embed.go:28`) — there is NO `embedded.go` file (CLAUDE.local.md's mention of it is stale). `make build` = `templ-generate` + `gen-catalog-hashes.go --all` + `go build` (Makefile:23-25); the embed is picked up directly by `go build`.
- **Mirror parity is Go-test-defined, NOT naive byte-identity**: the parity check is `go test ./internal/template/... -run TestRuleTemplateMirrorDrift` (`internal/template/rule_template_mirror_test.go:108`). Some target files are INTENTIONALLY divergent deployed↔template (the template copy is §25-sanitized): of this SPEC's targets, `spec-workflow.md` is in the byte-parity allowlist (`workflowOptMirroredPaths`) so its template mirror MUST get the identical edit, while `agent-common-protocol.md`, `ci-watch-protocol.md`, and `agent-teams-pattern.md` are §25-sanitized and EXCLUDED from byte-parity (their cleanliness is covered by `TestTemplateNoInternalContentLeak`, NOT byte-identity). Run-phase MUST NOT use raw `diff` to assert parity — it uses the Go-test mechanism.
- **Template neutrality (CLAUDE.local.md §25 / §15)**: replacement text must stay generic — no SPEC IDs/dates/SHAs leaking into template content. `Agent(general-purpose)` and role_profile language is neutral and safe.
- **D14 design tension**: `design/constitution.md` names `expert-frontend` in 10 places including FROZEN HARD clauses + a phase diagram + Sprint Contract negotiation. A mechanical rename risks breaking the design-pipeline narrative. Scoped as SHOULD with a carve-out-note default.
- **claude-code-guide ambiguity (D9)**: the file's escalation target must be the MoAI-custom archived agent (replace with `Explore`), NOT the Anthropic built-in helper (valid). Read the line in context before editing.

## §C Pre-flight Verification (completed during plan-phase)

| Check | Result |
|-------|--------|
| Grep sweep of archived names across `.claude/rules/` | 25 files matched; 18 carry live-reference defects, 4 are PRESERVE, 3 are role_profile-only |
| Template mirror existence (14 sampled files) | All 14 have mirrors → uniform Template-First applies |
| D13 boilerplate count | 8/16 language files have the `manager-quality` line; 8 omit it (removal restores consistency) |
| D10 self-contradiction | `agent-authoring.md` ships 17-agent catalog + its own 8-agent ceiling section → confirmed defect |
| D14 spread | 10 `expert-frontend` occurrences in design/constitution.md, mirrored at zone-registry CONST-V3R2-064 |

## §D Constraints

- [HARD] Do NOT touch `archived-agent-rejection.md`, `NOTICE.md`, `agent-patterns.md:269-273`, `spec-frontmatter-schema.md:64`.
- [HARD] Do NOT scrub any `role_profile` token.
- [HARD] Mirror every `.claude/rules` edit to the template; run `make build`.
- [HARD] Keep replacement language template-neutral.
- [HARD] Each fix must leave a grep-verifiable absence at the live-reference site (per acceptance.md).

## §E Self-Verification

See `progress.md` §E for the audit-ready signal skeleton. Plan-phase self-verification: spec-lint clean (12 frontmatter fields + Exclusions present), SPEC ID regex PASS, defect inventory each carries file:line + canonical replacement.

## §F Milestones

Milestones are grouped by **risk tier and domain**, P0 first. Each milestone applies edits to BOTH `.claude/rules/moai/...` and the template mirror, then a make-build verification batches at M-final.

### M1 — P0 CI-protocol live spawns (D4, D5)

Highest priority: `manager-quality` invoked as a live diagnostic subagent in `ci-autofix-protocol.md` (L83,101,107,122) and `ci-watch-protocol.md` (L79,113). These would trigger `ARCHIVED_AGENT_REJECTED` at runtime.

- Replace the `manager-quality` diagnostic-subagent references with the Stop hook `sync-phase-quality-gate.sh` and/or a per-spawn `Agent(general-purpose)` diagnostic scope (per archived-agent-rejection §C).
- Preserve the surrounding AskUserQuestion-boundary and auxiliary-fail-count prose; only the agent name/spawn mechanism changes.
- Mirror both files.
- Binds REQ-RCS-007, REQ-RCS-004, REQ-RCS-008.

### M2 — P0 agent-authoring catalog rewrite (D10)

`development/agent-authoring.md` §Agent Categories ships the OLD 17-agent catalog (Manager Agents (7) + Expert Agents (6) + Builder (1)), self-contradicting the file's own 8-agent ceiling section.

- Replace §Agent Categories with the 7 retained MoAI-custom agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`) + Anthropic built-in `Explore`.
- Verify internal consistency with the later 8-agent ceiling section.
- Mirror.
- Binds REQ-RCS-005, REQ-RCS-004, REQ-RCS-008.

### M3 — core/ hook-action + normative-example scrub (D1, D2)

- `core/agent-hooks.md` (~L52-55): remove the 4 archived rows (`expert-backend`/`expert-frontend`/`expert-devops`/`manager-quality`) from the `{agent}-{phase}` hook-action table, OR mark the table "archived — non-functional, historical mapping only". Keep the `manager-develop` / `manager-spec` / `manager-docs` retained rows.
- `core/agent-common-protocol.md`: L345 `expert-*` → `general-purpose with domain whitelist`; L401 `manager-quality in diagnostic mode` → `Explore`. Leave L37 prose (`phantom manager-quality / expert-security`) intact — it documents the consolidation.
- Mirror both.
- Binds REQ-RCS-001, REQ-RCS-004, REQ-RCS-008.

### M4 — workflow/ team-mode + worktree + dead-path scrub (D6, D7, D8, D9)

- `agent-teams-pattern.md:3`: remove the dead `paths:` token `.claude/agents/moai/manager-strategy.md` only. (Full file consolidation = SSOT-DEDUP-001.)
- `spec-workflow.md` L220 (`team-validator`/`team-tester`), L368 (`backend-dev`/`frontend-dev`): rewrite to role_profiles (implementer/tester/reviewer) spawned as `general-purpose`.
- `worktree-integration.md` L167 (archived HARD list) + L253 (stale "5 agents declare isolation" count): recount against the 8-agent catalog; replace archived names with retained equivalents / role_profiles. Keep the role_profile lines (L83,166,210,236) untouched.
- `worktree-state-guard.md:62`: `claude-code-guide` (MoAI-custom archived) → `Agent(Explore)` read-only. Confirm in-context it is the MoAI-custom one.
- Mirror all four.
- Binds REQ-RCS-001, REQ-RCS-003, REQ-RCS-004, REQ-RCS-006, REQ-RCS-008.

### M5 — development/ spawn-example + model-policy + skill-authoring scrub (D11, D12, D15)

- `orchestrator-templates.md` L102-104, L175: rewrite invalid `subagent_type: "analyzer"/"designer"/"implementer"/"reviewer"` to `general-purpose` + role profile.
- `model-policy.md` ~L84-90: update the Model Policy Tiers table (Opus/Sonnet/Haiku columns reference `strategy`/`researcher`/`quality`) to the 8-agent catalog OR note tiers are role-profile-based.
- `skill-authoring.md` L102 (`agents: ["expert-backend"]`), L173 (`agents: ["manager-spec", "expert-backend"]`): replace `expert-backend` with a retained agent (e.g. `manager-develop`) in the frontmatter examples.
- Mirror all three.
- Binds REQ-RCS-001, REQ-RCS-004, REQ-RCS-008.

### M6 — languages/ boilerplate consistency (D13)

- Remove the `delegate to \`manager-quality\` agent for AI-powered debugging` line from the 8 language files that carry it (`cpp`, `csharp`, `elixir`, `flutter`, `javascript`, `r`, `ruby`, `typescript`). The other 8 already omit it.
- Mirror all 8.
- Binds REQ-RCS-001, REQ-RCS-008. Note: language files are high-neutrality-sensitivity (16-language equal treatment) — removal restores parity, does not break it.

### M7 — design-pipeline carve-out (D14, D3) — SHOULD, higher-risk

Decision-gated (see §F.D14 below). Default approach: **carve-out note**, not mechanical rename.

- `design/constitution.md`: at the top of the design-pipeline section (near L24) add a carve-out note: "Pipeline `expert-frontend` resolves to `Agent(general-purpose)` with a frontend domain whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C — the archived name below denotes the pipeline ROLE, not a live spawn target." Then the 10 `expert-frontend` mentions are no longer silently load-bearing.
- `zone-registry.md:561` (CONST-V3R2-064): add the same cross-reference to the migration table in the clause text OR re-derive the clause to name the role generically.
- Mirror both.
- Binds REQ-RCS-010 (SHOULD), REQ-RCS-004, REQ-RCS-008.

### M-final — Template build + catalog-wide verification

- Run `make build` (`templ-generate` + `gen-catalog-hashes.go --all` + `go build`; NO `embedded.go` regen step — the template tree is embedded directly via `//go:embed all:templates`). Confirm `go build ./...` exits 0 after the deployed + template-source edits.
- Run the batched grep-verification suite from `acceptance.md` (per-defect absence + 4-PRESERVE-intact + role_profile-unchanged).
- Run `go test ./internal/template/... -run TestRuleTemplateMirrorDrift` to confirm mirror parity for byte-identical-allowlist files (`spec-workflow.md`) and `go test ./internal/template/... -run TestTemplateNoInternalContentLeak` to confirm the §25-sanitized mirrors (`agent-common-protocol.md`, `ci-watch-protocol.md`, `agent-teams-pattern.md`) stay leak-free. Do NOT use raw `diff` for parity.
- Binds AC-RCS-018/019/021.

## §F Anti-Patterns & Risk Decisions

### §F.D14 — design/constitution.md decision (the load-bearing risk)

`expert-frontend` is named in 10 places in `design/constitution.md`, including FROZEN HARD clauses (L72, L339), a phase diagram (L103), an I/O table (L122), and Sprint Contract negotiation (L315-328). A mechanical rename to `Agent(general-purpose)` would (a) read awkwardly in a sequence-diagram phase line, and (b) risk desyncing the FROZEN-zone clause from its mirror in `zone-registry.md` CONST-V3R2-064.

**Recommendation (default, encoded as M7 SHOULD)**: carve-out note, NOT rename. Add one note declaring the pipeline `expert-frontend` resolves to `Agent(general-purpose)`+frontend-whitelist per archived-agent-rejection §C, then leave the 10 role-name mentions as-is (they now denote a documented ROLE, not a live spawn). This:
- keeps the FROZEN clause text stable (minimal blast radius in a `[ZONE:Frozen]` zone),
- makes the archived name non-load-bearing (the AC is satisfied: the name is cross-referenced to the migration table, not silently live),
- defers a deeper design-pipeline-agent rename to a dedicated design SPEC if the team wants the retained-agent name surfaced everywhere.

**Escalation trigger**: if run-phase finds the carve-out note insufficient (e.g. the orchestrator still attempts an `expert-frontend` spawn from the pipeline), return a blocker to re-scope D14 as a HARD retained-rename via a follow-up design SPEC. Do not silently mechanical-rename inside a FROZEN zone.

### §F.D8 — agent recount decision

The stale "5 agents declare isolation: manager-develop, expert-frontend, expert-backend, expert-refactoring, researcher" (worktree-integration.md:253) lists 4 archived + 1 retained. Under the 8-agent catalog, isolation is a property of the WORK (cross-file writes), not a fixed agent list. **Recommendation**: replace the fixed count+list with role-based language: "write-heavy `Agent(general-purpose)` invocations and team-mode role_profiles (implementer/tester/designer) declare `isolation: worktree`; read-only `Explore` and read-only role_profiles do not." This removes the brittle count entirely rather than re-deriving a new number that will drift again.

### Anti-patterns to avoid

- **AP-RCS-001**: scrubbing an archived name inside one of the 4 PRESERVE files. Always grep-confirm the file is not a PRESERVE file before editing.
- **AP-RCS-002**: scrubbing a `role_profile` token (`researcher`/`reviewer`/etc.) mistaking it for an archived agent.
- **AP-RCS-003**: editing `.claude/rules` without mirroring to the template (or vice-versa) → drift.
- **AP-RCS-004**: mechanical-renaming inside a `[ZONE:Frozen]` clause (D14) without the carve-out decision → FROZEN-zone violation.
- **AP-RCS-005**: introducing a SPEC ID / date / SHA into template content during replacement → neutrality CI failure.

## §G Cross-References

- `spec.md` §B defect inventory (D1-D15) + §C GEARS requirements
- `acceptance.md` AC-RCS-001..021
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C migration table
- `CLAUDE.local.md` §2 / §15 / §25, `.moai/docs/template-internal-isolation-doctrine.md`

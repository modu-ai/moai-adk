# Agent Teams Pattern (Workflow Optimization Layer B)

Standardized pattern for spawning 5+ implementer teammates plus tester and
reviewer when a SPEC declares sufficient parallel implementation surface area.

> This rule was added by SPEC-V3R5-WORKFLOW-OPT-001 Layer B to formalize the
> Agent Teams composition that maximizes parallel run-phase throughput. It
> complements the existing `team-pattern-cookbook.md` (5 generic patterns) with
> a workflow-optimization-specific 7-teammate composition derived from the W3
> meta-analysis.

> Cross-reference: `.moai/config/sections/workflow.yaml` `team.role_profiles`
> (canonical role definitions); `.claude/rules/moai/workflow/team-protocol.md`
> (canonical team coordination protocol).

## When to Spawn a 5+1+1 Team

The orchestrator SHOULD spawn a 5-implementer + 1-tester + 1-reviewer team
(total 7 teammates) when ALL of the following hold:

1. `.moai/config/sections/workflow.yaml` `team.enabled: true` AND
   `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in `.claude/settings.json` env.
2. The SPEC declares at least 5 implementation packages with independent file
   ownership (no overlapping write surfaces).
3. The SPEC's harness level is `standard` or `thorough` (not `minimal` —
   minimal harness defaults to single-agent autopilot).
4. Wall-time budget allows ≥ 30 min of parallel execution (Agent Teams have
   higher token cost and coordination overhead than solo mode).

When any condition fails, fall back to solo mode (single `manager-develop`
invocation per package, sequential or parallel via the Bash tool's
multi-invocation).

## Composition Reference

| Role | Count | Mode | Model | Isolation | Purpose |
|------|------:|------|-------|-----------|---------|
| implementer | 5 | acceptEdits | sonnet | worktree | One per independent package |
| tester | 1 | acceptEdits | sonnet | worktree | Cross-package test coverage |
| reviewer | 1 | plan | sonnet | none | Read-only quality validation |

Role profile definitions live in `workflow.yaml` `team.role_profiles`. The
7-teammate composition is a usage pattern; the role definitions themselves
are shared with other team patterns (research, design, debug — see
`team-pattern-cookbook.md`).

## File Ownership Map (Example)

For a SPEC with 5 implementation packages:

```
implementer-1 → internal/pkg_a/**        (own *.go files, no tests)
implementer-2 → internal/pkg_b/**        (own *.go files, no tests)
implementer-3 → internal/pkg_c/**        (own *.go files, no tests)
implementer-4 → internal/pkg_d/**        (own *.go files, no tests)
implementer-5 → internal/pkg_e/**        (own *.go files, no tests)
tester        → internal/{pkg_a,..,pkg_e}/**_test.go  (own all test files)
reviewer      → (read-only)                              (cross-cuts all)
```

The `tester` role owns all `*_test.go` files to prevent write conflicts
between implementer and tester on the same test file. Implementer-N produces
the API; tester writes the test against the API.

## Spawn Order

```
Phase 1: Spawn reviewer (read-only, no isolation overhead)
Phase 2: Spawn 5 implementers in parallel (5 Agent() calls in one orchestrator turn)
Phase 3: Spawn tester after first implementer commits its package (tester needs API surface)
```

Spawning all 7 simultaneously is acceptable when the API surface is defined
upfront in the SPEC; in that case, tester can write tests against the spec
before implementer code lands.

## Communication Protocol

Follow the canonical team-protocol.md mailbox v2 envelope (`message`,
`shutdown_request`, `shutdown_response`, `blocker_report`, `task_handoff`).

Specific to the 5+1+1 pattern:

- Implementers communicate directly with each other when one package's API
  affects another (`task_handoff` or `message` with target peer name).
- Tester listens for `task_handoff` from any implementer signaling
  "API ready for test against my package."
- Reviewer receives a `message` from the team lead (orchestrator) when
  the implementer/tester pair completes; reviewer then runs read-only
  validation and sends results back as `message`.

## Fallback to Solo Mode

If team spawn fails (e.g., `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` not set,
`TeamCreate` returns error, or any teammate spawn fails), the orchestrator
MUST fall back to solo mode without retry. Solo mode invokes
`manager-develop` once per package sequentially. The fallback path is
already documented in `.claude/rules/moai/workflow/spec-workflow.md`
§ Agent Teams Variant § Fallback.

The orchestrator MUST record the fallback decision in `progress.md` for
the SPEC under a `team_fallback_reason:` field, so that subsequent SPEC
cycles can analyze whether team mode is consistently failing on this
environment.

## CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS Requirement

Agent Teams is an experimental Claude Code feature. Activation requires:

1. `.claude/settings.json`:
   ```json
   {
     "env": {
       "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
     }
   }
   ```
2. `.moai/config/sections/workflow.yaml`:
   ```yaml
   workflow:
     team:
       enabled: true
   ```

Both conditions MUST hold; otherwise solo mode is the only available path.
The orchestrator MUST verify both at the start of every run-phase that
intends to use team mode, and fail loudly with a structured error if either
is missing.

## Cross-references

- `.moai/config/sections/workflow.yaml` `team.role_profiles` — canonical role
  definitions (7 role keys).
- `.claude/rules/moai/workflow/team-protocol.md` — canonical team coordination
  protocol (mailbox, ledger, spawn validation).
- `.claude/rules/moai/workflow/team-pattern-cookbook.md` — 5 generic team
  patterns (research, implementation, review, design, debug).
- `.claude/rules/moai/workflow/worktree-integration.md` — `isolation: worktree`
  rules for write-heavy roles.
- SPEC-V3R5-WORKFLOW-OPT-001 acceptance.md AC-WO-009 (role_profiles 7-key check).

---

Version: 1.0.0
Classification: Evolvable workflow pattern, applies when SPEC declares 5+ independent implementation packages
Origin: SPEC-V3R5-WORKFLOW-OPT-001 Layer B (2026-05-20)

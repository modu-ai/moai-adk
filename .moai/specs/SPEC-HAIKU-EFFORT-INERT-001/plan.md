---
id: SPEC-HAIKU-EFFORT-INERT-001
title: "Implementation plan — remove inert effort field from haiku-tier agents"
version: "0.1.1"
status: completed
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tier: S
tags: "agent, effort, haiku, frontmatter, template"
---

# Plan — SPEC-HAIKU-EFFORT-INERT-001

## §A Context

Tier **S** (minimal). A configuration-honesty fix: remove a stray hand-authored `effort: medium` frontmatter field from the two `model: haiku` agents whose model does not support effort, and add a guard test so the invariant cannot silently regress. No generator logic change; no model tier change.

## §B Root-Cause Finding (from investigation)

| Question | Answer |
|----------|--------|
| Does `agentEffortMap` assign an effort to the haiku agents? | **No.** The map has exactly 5 entries, all `model: inherit`. `manager-docs`/`manager-git` are absent → `GetAgentEffort` returns `""` → `ApplyEffortPolicy` no-op (verified by existing test `no_op_for_agent_not_in_effort_map`). |
| Where did `effort: medium` come from? | **Hand-authored in the template frontmatter** of both haiku agent files. The generator did not emit it. |
| Is the dev-tree local file patched output or template source? | **Unpatched working copy — NOT a byte-identical mirror.** The local file carries internal SPEC-IDs/dates that are §25-stripped from the template (enrolled as a non-parity file at `rule_template_mirror_test.go:102-103`; verified by live `diff`). `ApplyModelPolicy`/`ApplyEffortPolicy` run only on user projects at init/update, never on the dev tree. Only the `^model:`/`^effort:` lines are identical across the two trees today. |
| Does the fix need a generator change? | **No.** The map already excludes haiku agents (REQ-HEI-004 is an invariant, not an edit). The fix is frontmatter removal + guard test. |

## §C Resolution Options (decision deferred to Implementation Kickoff Approval gate)

The human kickoff gate chooses. `manager-spec` recommendation is stated but non-binding.

### Option 1 — Remove the `effort` field from both haiku agents  ★ RECOMMENDED

- Delete the `effort: medium` line from `manager-docs.md` and `manager-git.md` (both template source + local mirror), keeping `model: haiku`.
- `agentEffortMap` needs **no change** — it already omits haiku agents (so the task's "adjust the map" sub-clause is a no-op here; the map is already correct).
- Add a guard test asserting the `model: haiku ⇒ no effort` invariant.
- **Effect:** frontmatter becomes honest (no field that the model silently ignores); zero behavior change (effort was already inert).
- **Rationale for recommendation:** matches the observed root cause exactly, is the minimal correct change, preserves the intended lightweight haiku model for sync/git work, and the guard prevents recurrence. Aligns with the "config must not lie" principle.

### Option 2 — Upgrade manager-docs/manager-git to an effort-supporting model tier

- Change `model: haiku` → an effort-supporting model (inherit/sonnet/opus) so `effort: medium` becomes meaningful.
- **Assessment: likely WRONG.** Sync (manager-docs) and git (manager-git) work is deliberately lightweight; `haiku` is the intended correct model per `model-policy.md` § Inherit-by-Default (the two named haiku exceptions). Upgrading the model purely to justify an inert field inverts the fix — it adds cost/latency to satisfy a stray line. Listed for completeness only.

### Option 3 — Keep as-is + add a doc/code comment explaining the inert-by-design state

- Leave `effort: medium` and annotate that it is intentionally inert on Haiku.
- **Assessment: weakest.** Leaves a misleading, non-functional config in place; a comment cannot travel with the frontmatter into every reader's mental model. Documenting a lie is worse than removing it. Listed for completeness only.

## §D Constraints

- **Template-First (CLAUDE.local.md §2):** edit `internal/template/templates/.claude/agents/moai/*.md` FIRST, then `make build`, then apply the identical `effort:`-line removal to local `.claude/agents/moai/*.md`. Parity requirement is scoped to the `^model:`/`^effort:` lines ONLY — these are §25-sanitized non-parity files (`rule_template_mirror_test.go:102-103`), so whole-file byte-identity is neither required nor expected (the pre-existing `description:`/body divergence is by design).
- **Scope discipline:** touch ONLY the two haiku agent files + one test file. Do not touch the 5 effort-supporting agents, `model_policy.go` logic, or any other frontmatter field.
- **No time estimates** (priority labels only).
- **16-language neutrality / internal-content isolation** unaffected — agent frontmatter carries no forbidden content classes.

## §E Self-Verification (plan-phase audit-ready signal)

- [x] Root cause traced to source (stray frontmatter line), not the generator.
- [x] Both trees (template + local) identified as in-scope.
- [x] Guard-test invariant defined (`model: haiku ⇒ no effort`).
- [x] Out-of-scope boundary drawn (no model change, no map logic change, no migration tooling).
- [x] SPEC ID regex self-check printed and PASS.

## §F Milestones (priority-ordered, no time estimates)

### M1 — Remove the inert `effort` field + rebuild + mirror
- Edit `internal/template/templates/.claude/agents/moai/manager-docs.md`: delete the `effort: medium` line (keep `model: haiku`).
- Edit `internal/template/templates/.claude/agents/moai/manager-git.md`: delete the `effort: medium` line (keep `model: haiku`).
- Edit local mirror `.claude/agents/moai/manager-docs.md` and `.claude/agents/moai/manager-git.md` identically.
- Run `make build` to recompile the embedded template FS.
- Verify: `grep -c '^effort:' ` on all four files returns 0; `grep -c '^model: haiku'` returns 1 each.

### M2 — Add regression guard test for the `model: haiku ⇒ no effort` invariant
- Add a test under `internal/template/` (new focused `*_test.go` file OR a function appended to `model_policy_test.go`) that scans agent frontmatter in BOTH trees — `internal/template/templates/.claude/agents/moai/*.md` AND the local `.claude/agents/moai/*.md` (resolve the repo root by walking up from the package dir) — and fails if any file declaring `model: haiku` also declares an `effort:` field (key on the model field, not agent names).
- Confirm the existing `TestApplyEffortPolicy/no_op_for_agent_not_in_effort_map` still passes (evidence the generator will not re-add effort for haiku agents).
- Run `go test ./internal/template/...` — all pass.

> Tier S note: M1 and M2 MAY be executed in a single pass. They are listed separately for clarity, not to mandate two commits.

## §G Anti-Patterns to Avoid

- Adding `manager-docs`/`manager-git` to `agentEffortMap` "for completeness" — the map is correct as-is; adding them would re-emit the inert field.
- Removing `effort` from the 5 effort-supporting agents (collateral over-reach).
- Editing only the local mirror or only the template source (Template-First violation → mirror drift).
- Changing the haiku model tier to justify the field (Option 2 inversion).

## §H Cross-References

- `spec.md` §B (REQ-HEI-001..007), §C (Out of Scope).
- `acceptance.md` (AC matrix, DoD).
- `internal/template/model_policy.go`, `internal/template/model_policy_test.go`.

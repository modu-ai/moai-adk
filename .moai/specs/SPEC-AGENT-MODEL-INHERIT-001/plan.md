---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Plan — subagent model/effort inheritance"
created: 2026-09-26
---

# Plan — SPEC-AGENT-MODEL-INHERIT-001

## §A Context

Tier **L**: ≥3 milestones and well over 10 files (46 agent files, ~38 non-test Go files, ≥29
test files, ~20 rule/skill mirrors, 48 docs-site pages — research.md). Milestones are ordered by
decision-reversibility first (data model, retention seams, user-facing removals), mechanical
edits last. The mechanical doctrine milestones also come last because they share 21 files with
card t1175, which lands before this run starts.

## §B Known issues

- `make agents-emit` fails closed on an empty `effort` until the Codex manifest changes (research.md §B) — M6 changes both in one commit.
- Agent-lint LR-03 turns every MoAI agent into an error once `effort:` is stripped — M6 retires it in the same commit.
- `internal/cli/testdata/codex-rollouts-t1171/**` embeds `DefaultProfileMatrix` text in captured rollouts; these are fixtures and are only regenerated if a test that reads them fails.
- The local `.moai/config/sections/llm.yaml` is untracked in the primary checkout; this SPEC never commits it.

## §C Pre-flight (run entry)

1. t1175 merged into local develop; absorb develop (`git merge develop` inside this worktree) — REQ-AMI-001.
2. Re-measure overlap with `WT-role-naming-docs` and halt on a touched shared file — REQ-AMI-002.
3. Re-run every research.md inventory command on the absorbed tree and record the counts in progress.md §E.2.
4. Baseline: `TestAlwaysLoadedTokenBudget` headroom, `make agents-emit-check`, affected-package `go test`.

## §D Constraints

- Template-First; `make build` after template edits; `make agents-emit` after C2 agent edits; never hand-edit `.codex/agents/moai/*.toml`.
- Tests scoped to affected packages (`internal/config`, `internal/template/...`, `internal/cli/...`, `internal/hook`, `internal/web`, `internal/settings/...`, `internal/harness/...`, `internal/spec`); full suite is CI's job.
- One writer per tree; no push (lead pushes develop).

## §E Self-verification (per milestone)

Each milestone closes with: affected-package `go test` output, `go vet` on those packages,
`golangci-lint run` on changed packages, and for text milestones a grep proving the removed
patterns are gone from both trees.

## §F Milestones

### M0 — Run entry and baseline (Priority High)
REQ-AMI-001, REQ-AMI-002. No edits. Output: progress.md §E.2 table of re-measured counts, overlap list, budget headroom.

### M1 — Retention seams first (Priority High)
REQ-AMI-018..021. Characterisation tests pinning today's main-session model/effort resolution,
GLM alias mapping, session GLM reasoning, and audit-pin precedence. Introduce the model-free
roster SSOT (D4) and re-point rosterguard and `config.retainedAgentNames`. Re-point codex/GLM
task and audit resolution to pin > backend default (D5). These decisions shape every later
milestone, so they come first.

### M2 — Configuration schema and migration (Priority High)
REQ-AMI-014..017. Remove the `LLMConfig` fields and profile validation; template `llm.yaml`
drops the keys and comments; relocate the retired-key strip and extend it to the four keys;
update report lines (D8); `performance_tier` and the `workflow.yaml` routing keys and their
validators (D12); `--profile` becomes a deprecation-warning no-op (D10); `update_wizard` stops writing
`llm.profile`. Tests for each row of design.md §C.

### M3 — Hook guard removal (Priority High)
REQ-AMI-010/011. Delete `agent_model_guard.go` and its test, `pre_tool.go` wiring, config type
and default, `prune_logs` entry, gitignore-artifact test row; reword sibling-guard comments. Test
that a config carrying `agent_model_guard` still loads.

### M4 — CLI resolver removal (Priority Medium)
REQ-AMI-009. Delete `moai model` (`model.go`, root registration), `profile_matrix.go` resolvers
and matrix, orphaned per-agent GLM helpers (grep-proven), `cellguard` (D6), config `Profiles`
mirrors in `cli/glm.go`. Update or delete the tests found in M0.

### M5 — Web console removal (Priority Medium)
REQ-AMI-012/013. Remove the agentfm tab, handlers, app seams, templ blocks (regenerate
`*_templ.go`), `internal/settings/agentfm`, orphaned `v4manifest` display helpers. Test that the
preference-profile routes and main-session controls still respond as before. The whole tab
goes, UI and API (D7).

### M6 — Agent frontmatter and Codex emission (Priority Medium)
REQ-AMI-004..007. In one commit: strip `model:`/`effort:` from 12 template agents, set the
Codex manifest `model_reasoning_effort.emit: false` (D2), retire LR-03/LR-12 (D3), adapt
`haiku_effort_guard_test` and agentemit golden tests, run `make agents-emit` and
`make agents-emit-check`; then strip the 12 local moai agents and the 10 local harness agents.
Harness v4 manifest fields become optional and generation stops emitting them (D11); the
dynamic-workflow `agent()` model/effort literals are removed from template and local scripts (D12).

### M7 — Doctrine text (Priority Low — mechanical, overlaps t1175)
REQ-AMI-008, 022..025. Apply design.md §D H1–H16 template-first, then local; `make build`;
`TestAlwaysLoadedTokenBudget` before/after; template-neutrality guard; zone-registry check.

### M8 — docs-site and records (Priority Low)
REQ-AMI-026. Rewrite or remove the 48 docs-site pages in four locales in one change set, add a
redirect for every removed page, and run the oss-docs verify recipe. CHANGELOG and superseded-SPEC status belong to sync.

## §G Anti-patterns

- Stripping C2 `effort:` before the manifest change — `make agents-emit` fails closed.
- Hand-editing Codex toml files.
- Removing the roster together with the matrix — rosterguard loses its canonical list.
- Treating a leftover `llm.yaml` key as an error — breaks `moai update` for edited configs.
- Running `go test ./...` locally.

## §H Cross-references

research.md (inventory and overlap), design.md (decisions, migration, HARD-clause table),
acceptance.md (criteria), progress.md §E.1 (operator answers Q1–Q6, resolved 2026-09-26).

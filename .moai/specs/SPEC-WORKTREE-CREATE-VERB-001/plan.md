# SPEC-WORKTREE-CREATE-VERB-001 — Implementation Plan

> Card t1070 (class C). Plan phase artifacts authored 2026-09-22 in worktree `.claude/worktrees/t1070` (branch `WT-worktree-verb`, HEAD `0314801c2`).

## §A Context

moai's harness-neutral creation capability already exists (`materializeSessionWorktree`, `internal/cli/session_worktree.go:203`) but is consumed only as a side effect of `moai init` / `moai web` / `moai profile`, gated default-OFF. Creation is exposed as a verb nowhere: `internal/cli/worktree/` has no creation verb (the old `moai worktree new` is retired with bodp), `moai cc -w` outsources creation to the `claude` binary, and `moai codex -w` deliberately resolves-only (`codex_launcher.go:257-267` records the gap). This plan wires the existing capability into ONE new verb surface.

## §B Known Issues

- The original plan deliberately left the direction undecided; M1 has now resolved it to option 가 after the operator-relayed live Codex observation.
- The t1050 investigation initially recorded **0 live Codex session runtime observations**. That historical gap is closed by the M1 evidence in `progress.md` §E.2.

## §C Pre-flight (decision gate — RESOLVED before M2)

**Decision: option 가 — revive `moai worktree new` as the harness-neutral creation verb.**

- **(가)** Revive `moai worktree new` as a first-class verb in `internal/cli/worktree/` (new verb file + `root.go` wiring).
- **(나)** Narrow the surface to `moai codex -w --create` (extend `resolveCodexWorktreeDir`'s resolve path with creation, retiring the resolve-only error for the create flag form).

**Measured basis:** the operator relayed a live Codex-lane observation that `moai codex -w t1070` resolves and enters the existing tree while the missing-tree form remains resolve-only. The remaining capability is therefore provisioning before entry. A Codex-only `--create` flag would not serve scripts, factory leads, or the t1082 lane-worktree handoff, while `worktree new` gives all harnesses one MoAI-owned surface.

**Implementation constraint:** the new verb MUST reuse the existing creation plumbing (`materializeSessionWorktree` + `LoadWorktreeBaseBranch`) through an injected adapter. It must validate a single L1 leaf name before that adapter is called. It must not author a second `git worktree add` invocation path. The retired `moai worktree new` implementation and its `--base`/`--from-current` flags are NOT revived.

**Phase 4 Mode Selection note:** serial mode expected — single-package scope (`internal/cli`), one verb surface, no independent lanes.

## §D Constraints

- Reuse-first (REQ-WCV-001): no parallel creation algorithm.
- L1 placement (REQ-WCV-007): `.claude/worktrees/<name>`; L1/L2 boundary per `worktree-integration.md` § Terminology Glossary.
- Non-interactive CLI (REQ-WCV-006): C-HRA-008 / REQ-PGN-012; static guard test required.
- Session-worktree gate default-OFF baseline (REQ-WCV-005 / REQ-SW-001): byte-identical — the verb's introduction must not flip the gate or change existing consumers.
- TDD mode: RED-GREEN-REFACTOR per milestone.
- Errors/comments English.
- No full-suite local runs; affected-package scope only (`go test ./internal/cli/...` scoped to touched packages), CI owns the full verdict.

## §E Self-Verification

Plan-phase self-checks (executed and cited in `progress.md` §E.1):

1. SPEC-ID regex check → PASS (verbatim output cited in §E.1).
2. Frontmatter canonical 12 fields validated against the schema SSOT.
3. ID uniqueness confirmed against the 908-SPEC catalog (no collision).
4. Requirements in GEARS notation (spec.md §B).
5. Out of Scope section satisfies the `OutOfScopeRule` lint convention (spec.md §D: three `### Out of Scope —` H3 sub-headings with `-` bullets).
6. Artifact set matches the Tier (M: spec.md + plan.md + acceptance.md + progress.md, plus orchestrator-mandated research.md), directory layout (no flat files).
7. spec.md carries no implementation detail (function/file anchors live here and in research.md, not as HOW mandates in REQ text — REQ text names the plumbing as a reuse contract, not a design).

## §F Milestones (priority-ordered by decision reversibility)

- **M1 (High — complete):** Live Codex existing-tree entry observation relayed by the operator; option 가 selected and the cross-harness handoff recorded in `progress.md` §E.2.
- **M2 (High — verb surface shape):** Wire the chosen surface to the creation plumbing. (가): new verb file in `internal/cli/worktree/` + `root.go` registration. (나): `--create` flag on `moai codex -w` extending the resolve path. Argument validation (REQ-WCV-003) and collision refusal (REQ-WCV-004) land here. RED first: failing tests for create-at-conventional-path and refusal-without-value.
- **M3 (Medium — diagnostics + boundary):** Error message family consistent with the existing resolve-error shape; L1 placement validation; English strings. GREEN.
- **M4 (Medium — guards):** Static guard test (`TestNew_NoAskUserQuestion` pattern) for the new surface; gate-baseline test confirming `SessionWorktreeEnabled` default unchanged and existing consumers byte-identical (REQ-WCV-005).
- **M5 (Low — mechanical verification):** `go vet` + `golangci-lint` on touched packages; scoped `go test ./internal/cli/...` for affected packages; evidence persisted under `.moai/state/verify/`.

## §G Anti-Patterns

- Do NOT implement both (가) and (나); the decision gate picks one surface.
- Do NOT author a second `git worktree add` invocation path alongside `materializeSessionWorktree` plumbing.
- Do NOT flip the session-worktree feature gate default as a side effect of the verb.
- Do NOT reintroduce retired `--base`/`--from-current` flag semantics under the revived name.
- Do NOT run the full test suite locally (affected packages only; CI owns the full verdict).

## §H Cross-References

- research.md — measured facts 1-5 with file:line anchors; retired-verb history; Codex-observation-count-0 gap.
- SPEC-SESSION-WORKTREE-001 — the default-OFF gate and `materializeSessionWorktree` origin (REQ-SW-001 baseline).
- SPEC-WORKTREE-BASEREF-001 — base-branch selection (REQ-WBR-010/011).
- SPEC-CLI-WORKTREE-FLAG-RACE-001 — `moai codex -w` resolve path lineage.
- Sibling cards: t1071 (AGENTS.md binding row + `moai codex -w` registration), t1072 (worktree-integration.md:227), t1073 (`moai worktree done` tier semantics) — all out of scope here.

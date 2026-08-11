---
description: "Read-only verification batching rationale + class taxonomy (loads at run/sync-phase completion)"
paths: "**/.moai/specs/**,**/.claude/skills/moai/workflows/run.md,**/.claude/skills/moai/workflows/sync/**,**/verification-batch-pattern.md,**/agent-common-protocol.md"
---

# Verification Batch Pattern

Canonical pattern for orchestrator-side read-only verification batching during run-phase completion. Motivation: reduces serial-verification round-trip latency at run-phase completion.

Cross-reference: `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution defines the HARD batching obligation; this file owns the grouping rationale and class taxonomy.

## Why Batch

When `manager-develop` reports completion, the orchestrator independently verifies seven dimensions: test suite, coverage, subagent-boundary, sentinel-key, CLI smoke, benchmark, lint. Each is read-only and independent. Serial issuance multiplies round-trip latency; multi-Bash batching collapses it to the slowest single command.

## When to Batch (Verification Class Taxonomy)

| Class | Read-only? | Mutates state? | Batch-safe? |
|-------|------------|----------------|-------------|
| Test execution | yes (output only) | no | YES |
| Coverage measurement | yes | writes `cover.out` (no side effect) | YES |
| Grep / find / sentinel scan | yes | no | YES |
| CLI smoke (--version, --help) | yes | no | YES |
| Benchmark | yes | no | YES |
| Lint (golangci-lint, ruff, etc.) | yes | no | YES |
| Build (`go build`, `npm run build`) | depends | writes artifacts | NO if downstream depends |
| Test fixture setup | yes | writes test files | NO if shared state |

All seven canonical batch items in agent-common-protocol §Parallel Execution are read-only batch-safe.

> **Re-sync sentinel**: the verbatim 7-command batch AND the file-redirect contract (redirect + bounded-tail output representation) live in `agent-common-protocol-reference.md` § Canonical 7-item example / § File-redirect contract (the detail sidecar of `agent-common-protocol.md`, which retains the binding summary). If either the 7-item list OR the file-redirect contract representation changes, re-sync this file's grouping rationale and the class taxonomy below to match. This file owns only the *why* (grouping rationale + class taxonomy + anti-patterns), not the *what* (the verbatim command list or its output representation).

## When NOT to Batch

- Explicit dependency (`make build` before tests that invoke its binary).
- Same-file writes (two `coverprofile=cover.out` runs race).
- Shared-state mutation (`git checkout` + `git status` in one tree).

Serialize dependent ops; batch independent read-only verifications by default.

## Grouping Heuristic

| Group | Members | Typical Total Time |
|-------|---------|-------------------:|
| A. Functional | `go test ./...`, coverage | 30-120 s |
| B. Boundary | subagent-boundary grep, sentinel scan, frontmatter check | 1-5 s |
| C. Quality | golangci-lint, spec-lint | 10-60 s |
| D. Smoke | CLI --version, --help | 1-3 s |
| E. Benchmark (optional) | go test -bench | 30-300 s |

Groups A-D issue as one parallel batch. Group E joins when benchmark is in AC.

## Anti-Pattern Catalogue

- **AP-VBP-001 — Serial across turns**: N turns where one suffices. Adds N × round-trip latency plus context-switch overhead.
- **AP-VBP-002 — Pseudo-batch via `&&`**: Chains sequentially in one shell, not parallel. First failure short-circuits.
- **AP-VBP-003 — Pseudo-batch via `&`**: Interleaved output is hard to parse. Orchestrator-level multi-Bash is cleaner — each call produces a separate, structured output block.

## Correct Pattern (Reference)

The orchestrator's response contains multiple Bash tool calls within a single assistant turn. The canonical 7-item example lives in `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (satisfies the canonical verification-batch acceptance criterion).

## Attributable diff-check pattern (SPEC-SYNC-PARALLEL-DOCS-001 A9)

> Thin pointer — the SSOT for the attributable diff-check doctrinal switch + its fallback contract lives in `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch. That section owns: the three-way attribution predicate (snapshot key / command / output), the CONSUME-vs-re-execute default-inversion, the mismatch-reason enum (`snapshot_key_drift` / `command_drift` / `missing_section_e` / `output_drift`), and the fallback-to-re-execution safety boundary. This file owns only the *why* (grouping rationale + class taxonomy above).

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (HARD batching obligation) + `agent-common-protocol-reference.md` (7-item canonical example).
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch (SPEC-SYNC-PARALLEL-DOCS-001 A9 — the composition-time switch that selects consume-vs-re-execute).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E → Attribution discipline (SPEC-SYNC-PARALLEL-DOCS-001 A9 — the §E attribution triple the diff-check consults).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2 (the invariant + attribution contract A9 preserves on every path).
- reduces serial CI wait.

---

Version: 1.1.0 (SPEC-SYNC-PARALLEL-DOCS-001 A9 — attributable diff-check pattern + A9 fallback contract added)
Classification: Evolvable operational rule, applies to all run-phase completion verifications

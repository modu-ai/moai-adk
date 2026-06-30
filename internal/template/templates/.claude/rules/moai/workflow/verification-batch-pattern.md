# Verification Batch Pattern

Canonical pattern for orchestrator-side read-only verification batching during run-phase completion. Added in response to a meta-analysis (≈10 min, ~11% of run-phase wall-time lost to serial verification).

Cross-reference: `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution defines the HARD batching obligation; this file owns the grouping rationale and class taxonomy.

## Why Batch

When `manager-develop` reports completion, the orchestrator independently verifies seven dimensions: test suite, coverage, subagent-boundary (C-HRA-008), sentinel-key, CLI smoke, benchmark, lint. Each is read-only and independent. Serial issuance multiplies round-trip latency; multi-Bash batching collapses it to the slowest single command.

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

> **Re-sync sentinel**: the verbatim 7-command batch lives in `agent-common-protocol.md` § Parallel Execution (the SSOT). If that 7-item list changes, re-sync this file's grouping rationale and the class taxonomy below to match. This file owns only the *why* (grouping rationale + class taxonomy + anti-patterns), not the *what* (the verbatim command list).

## When NOT to Batch

- Explicit dependency (`make build` before tests that invoke its binary).
- Same-file writes (two `coverprofile=cover.out` runs race).
- Shared-state mutation (`git checkout` + `git status` in one tree).

Serialize dependent ops; batch independent read-only verifications by default.

## Grouping Heuristic

| Group | Members | Typical Total Time |
|-------|---------|-------------------:|
| A. Functional | `go test ./...`, coverage | 30-120 s |
| B. Boundary | C-HRA-008 grep, sentinel scan, frontmatter check | 1-5 s |
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

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (HARD batching obligation + 7-item canonical example).
- The canonical workflow-optimization acceptance criterion for the verification batch pattern.
- W3 HARNESS-AUTONOMY-001 meta-analysis (`feedback_w3_metaanalysis_lessons.md`).

---

Version: 1.0.0
Classification: Evolvable operational rule, applies to all run-phase completion verifications
Origin: workflow-optimization Layer D (2026-05-20)

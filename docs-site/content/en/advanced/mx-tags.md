---
title: "@MX TAG System"
weight: 61
draft: false
---

@MX TAG is a code-level annotation that serves as the standard means for AI agents to convey **context · invariants · danger zones** across development sessions. Prompts can be ignored, but comments inscribed in code survive alongside the code, so the next agent can immediately grasp the intent and constraints the moment it first reads the code.

> The operations (scanning · adding · querying) of @MX TAG are performed via the `/moai mx` command. This page covers the protocol and lifecycle of the tag system itself.

## Tag Syntax

```go
// @MX:TAG_TYPE: [description]
// @MX:SUB_KEY: [sub value]
```

A tag is an inline source comment, not a separate JSON ledger. It is collected via `grep` or `moai mx query`.

## Tag Types

| Tag | Purpose | Required Subline |
|------|------|----------------|
| `@MX:NOTE` | Conveys context and intent | — |
| `@MX:WARN` | Marks a danger zone | `@MX:REASON` |
| `@MX:ANCHOR` | Invariant contract (high fan_in) | `@MX:REASON` |
| `@MX:TODO` | Incomplete work | — |
| `@MX:DEBT` | Intentional simplification (working code) | `@MX:CEILING` + `@MX:UPGRADE` |

## Sublines

`@MX:SPEC` · `@MX:LEGACY` · `@MX:REASON` · `@MX:TEST` · `@MX:PRIORITY` · `@MX:CEILING` · `@MX:UPGRADE`

- `@MX:REASON` is **required** for WARN and ANCHOR.
- The `[AUTO]` prefix is **required** for agent-generated tags.

## When to Add

**@MX:NOTE** — Magic constants, exported functions over 100 lines without godoc, business rules without explanation.

**@MX:WARN** — Goroutines/channels without `context.Context`, cyclomatic complexity 15 or above, global state mutation, 8 or more if-branches.

**@MX:ANCHOR** — fan_in 3 or above, public API boundaries, external system integration points.

**@MX:TODO** — Public functions without test files, unimplemented SPEC requirements, errors returned without handling.

**@MX:DEBT** — When an intentional simplification is adopted, correct within the stated limit (`@MX:CEILING`), with a revisit trigger (`@MX:UPGRADE`).

## DEBT — Explicit Limits of a Working Simplification

`@MX:DEBT` is not a marker for incomplete work. The code is **already complete and working correctly**, but it records that it is an intentional simplification within stated limits. Two sublines follow.

```go
// @MX:DEBT: in-memory map cache, no eviction
// @MX:CEILING: < 10k entries
// @MX:UPGRADE: switch to LRU when entry count exceeds 10k
```

A DEBT without `@MX:UPGRADE` has no termination condition and rots silently. `moai mx query --kind DEBT --json` reports this as `"rotRisk": "no-trigger"`. The rot signal is the absence of `@MX:UPGRADE`; the absence of `@MX:CEILING` is merely a quality memo and not a criterion for rot.

> `@MX:TODO` marks incomplete work resolved in the GREEN step (code not yet complete), while `@MX:DEBT` marks a simplification that is complete and works correctly but has an explicit limit (code is complete). DEBT may legitimately persist across multiple GREEN steps, and the "promote to WARN after 3 unresolved occurrences" rule for TODO does not apply.

## When to Update / Remove

- **ANCHOR** — Update on fan_in change or SPEC update. Never auto-delete; demote to NOTE via report.
- **NOTE** — Re-review on function signature change.
- **WARN** — Remove when the dangerous structure is improved.
- **TODO** — Remove on resolution (test passes or implementation complete). Promote to WARN after 3 repeated unresolved occurrences.
- **DEBT** — Update on limit or trigger change. Remove when the `@MX:UPGRADE` trigger fires and the simplification is replaced, independent of other work completion. No auto-promotion.

## Lifecycle Summary

```text
TODO     Created in RED/ANALYZE → Resolved (removed) in GREEN/IMPROVE → Promoted to WARN after 3 unresolved occurrences
ANCHOR   Created when fan_in ≥ 3 → Updated on call-count/SPEC change → Demoted to NOTE (report) when fan_in < 3 → No auto-deletion
WARN     Created on danger detection → Persists if structural → Removed on resolution
NOTE     Created when context is needed → Updated after signature change → Discarded on code deletion
DEBT     Created on intentional simplification → Resolved (simplification replaced) on UPGRADE trigger firing → No auto-promotion
```

## Per-Language Comment Syntax

| Language | Prefix | Example |
|------|--------|------|
| Go · Java · TS · Rust · C/C++ · Swift · Kotlin · Dart · Zig · Scala | `//` | `// @MX:NOTE:` |
| Python · Ruby · Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

## Configuration (`.moai/config/sections/mx.yaml`)

- **thresholds** — `fan_in_anchor`, `complexity_warn`, `branch_warn`
- **limits** — `anchor_per_file` (default 3), `warn_per_file` (default 5). On excess, ANCHOR demotes from lowest fan_in; WARN keeps only P1–P5 priority.
- **exclude** — Tagging-excluded patterns such as `**/*_generated.go`, `**/vendor/**`, `**/mock_*.go`
- **require_reason_for** — Tag types for which REASON is required

## Tag Language

Tag descriptions and `@MX:REASON` follow the `code_comments` setting in `.moai/config/sections/language.yaml` (default `en`). For a Korean project, set `code_comments: ko` and tags will be written in Korean.

## Next Steps

- [Hooks Guide](/en/advanced/hooks-guide) — The foundation for handling code context alongside hooks
- [SPEC-Based Development](/en/core-concepts/spec-based-dev) — SPEC lifecycle and @MX TAG integration
- [TRUST 5 Quality Framework](/en/core-concepts/trust-5) — The Readable principle and @MX:NOTE

# Patterns

> **Module**: moai-foundation-core/patterns
> **Purpose**: Cross-cutting patterns that recur across MoAI-ADK's foundational principles.

---

## Overview

The six foundational principles (TRUST 5, SPEC-First DDD, Delegation, Token Optimization, Progressive Disclosure, Modular System) share a small set of recurring patterns. This module names them so they can be recognized and applied consistently regardless of which principle is currently active.

## Key Patterns

**Delegate, don't execute.** MoAI orchestrates; it does not implement complex tasks directly. Every non-trivial unit of work is handed to a retained agent via natural-language delegation ("Use the {agent} subagent to {task}"), never a `subagent_type` code literal. See [delegation-patterns.md](delegation-patterns.md) and [agents-reference.md](agents-reference.md).

**Phase separation with context reset.** Work is split into phases (plan / run / sync) and `/clear` is executed between them so each phase starts from a clean, minimal context. This is what lets a bounded context budget span an entire SPEC lifecycle. See [token-optimization.md](token-optimization.md) and [spec-first-ddd.md](spec-first-ddd.md).

**Progressive disclosure.** Content is layered (Quick / Implementation / Advanced) and loaded on demand — SKILL.md stays under 500 lines, deep dives live in modules/, and reference material loads only when needed. Readers pay only for the depth they use. See [progressive-disclosure.md](progressive-disclosure.md) and [modular-system.md](modular-system.md).

**Quality as a gate, not a guideline.** TRUST 5 (Tested / Readable / Unified / Secured / Trackable) is enforced before completion; a partial pass blocks. Tooling is language-neutral and auto-detected per project rather than hard-coded to one stack. See [trust-5-framework.md](trust-5-framework.md).

**Specification before implementation.** Requirements are authored (GEARS format) and acceptance criteria defined before code is written, so implementation is verifiable rather than aspirational. See [spec-first-ddd.md](spec-first-ddd.md).

## Composition

These patterns reinforce one another: delegation enables phase separation (each agent owns a phase's context); phase separation enables progressive disclosure (each phase loads only what it needs); progressive disclosure keeps the token budget viable; and TRUST 5 gates each phase's output. Applying any one in isolation under-delivers — applying them together is the foundation.

## Best Practices

- When a task feels too big for one agent, split by phase or by an independent file set — do not invent a new agent name (legacy names are archived and rejected at spawn).
- When context feels heavy, `/clear` between phases rather than loading more into a single session.
- When a reference file grows past ~500 lines, split it into a module and cross-link from the entry point.
- When a quality check is skipped "for speed", treat it as a red flag — the gate exists precisely for that moment.

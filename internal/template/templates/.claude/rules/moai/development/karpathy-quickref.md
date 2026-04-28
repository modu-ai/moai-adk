---
description: "Quick reference mapping 4 Karpathy coding principles to 6 MoAI Agent Core Behaviors with checkpoint questions"
paths: "**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs,**/*.c,**/*.cpp,**/*.rb,**/*.php,**/*.kt,**/*.swift,**/*.dart,**/*.ex,**/*.scala,**/*.hs,**/*.zig"
---

# Karpathy Coding Principles — Quick Reference

Mapping Andrej Karpathy's 4 coding principles to MoAI's 6 Agent Core Behaviors with checkpoint questions.

## Principle-to-Behavior Mapping

| Karpathy Principle | MoAI Behavior(s) | Coverage |
|---|---|---|
| Think Before Coding | Surface Assumptions + Manage Confusion | 95% |
| Simplicity First | Enforce Simplicity + Scope Discipline | 90% |
| Surgical Changes | Scope Discipline + Scope Discipline | 85% |
| Goal-Driven Execution | Verify Don't Assume + Push Back | 80% |

## Checkpoint Questions

### Think Before Coding

- Have I listed my assumptions explicitly?
- Is there conflicting information I'm ignoring?
- Am I about to implement without understanding the "why"?

### Simplicity First

- Can this be done in fewer lines? (3x LOC trigger)
- Are these abstractions earning their complexity?
- Would a staff engineer say "why didn't you just..."?

### Surgical Changes

- Am I touching only what was asked?
- Does my change match the existing code style?
- Am I refactoring adjacent code "while I'm here"?

### Goal-Driven Execution

- What is the testable completion assertion?
- Have I run the tests to verify?
- Am I claiming success without evidence?

## Cross-Reference

For concrete wrong/right code examples, see skill: `moai-reference-anti-patterns`

---

**Version**: 1.0.0  
**Source**: SPEC-KARPATHY-001 M2  
**Last Updated**: 2026-04-28

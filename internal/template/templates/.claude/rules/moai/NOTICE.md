# MoAI-ADK Third-Party Notices

This product includes software developed by revfactory/harness and redistributed under the Apache License 2.0.

## Apache License 2.0

The following source material is licensed under Apache License 2.0:

**Source Repository**: https://github.com/revfactory/harness  
**License**: Apache License 2.0 (https://www.apache.org/licenses/LICENSE-2.0)

### Imported Components

The following reference documents from `revfactory/harness` (imported 2026-04-26) are incorporated into MoAI-ADK as pattern cookbook rules:

1. `agent-design-patterns.md` → `.claude/rules/moai/development/agent-patterns.md`
2. `qa-agent-guide.md` → `.claude/rules/moai/quality/boundary-verification.md`
3. `skill-testing-guide.md` → `.claude/rules/moai/development/skill-ab-testing.md`
4. `team-examples.md` → `.claude/rules/moai/workflow/team-pattern-cookbook.md`
5. `orchestrator-template.md` → `.claude/rules/moai/development/orchestrator-templates.md`
6. `skill-writing-guide.md` → `.claude/rules/moai/development/skill-writing-craft.md`

### Attribution

This product includes software developed by revfactory/harness contributors. The original works and any modifications are provided under the terms of the Apache License 2.0.

The imported documents have been adapted for MoAI-ADK terminology and 16-language neutrality while preserving the original technical content and design patterns. Original source authorship is retained.

### Full Apache License 2.0 Text

For the complete Apache License 2.0 text, visit: https://www.apache.org/licenses/LICENSE-2.0

---

## Karpathy Coding Principles

The following reference material is derived from Andrej Karpathy's coding philosophy:

**Source Repository**: https://github.com/forrestchang/andrej-karpathy-skills

### Imported Concepts

The following concepts from Karpathy's 4 coding principles and anti-pattern catalog (imported 2026-04-28) are incorporated into MoAI-ADK:

1. **4 Coding Principles** → `.claude/rules/moai/development/karpathy-quickref.md`
   - Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven Execution
   - Mapped to MoAI's 6 Agent Core Behaviors with checkpoint questions

2. **Anti-Pattern Catalog (8 categories)** → `.claude/skills/moai/references/anti-patterns.md`
   - Premature Abstraction, Over-Engineering, Drive-By Refactoring, Style Drift
   - Silent Assumption, Guessing Over Clarifying, Sycophantic Agreement, Claiming Without Evidence
   - Adapted with Go/Python/TypeScript code examples for MoAI agent context

3. **Constitution Amendments (3 additions)** → `.claude/rules/moai/core/moai-constitution.md`
   - Behavior 4: Quantitative LOC trigger (Simplicity First)
   - Behavior 5: Style-matching directive (Surgical Changes)
   - Behavior 6: Goal-to-test pattern (Goal-Driven Execution)

### Attribution

Andrej Karpathy's coding principles are shared publicly as educational material. The `forrestchang/andrej-karpathy-skills` repository packages these principles into a structured reference. MoAI-ADK has adapted the concepts, mapped them to existing Agent Core Behaviors, and created concrete code examples specific to MoAI's orchestration context.

---

**Import Date (harness)**: 2026-04-26
**Import Date (Karpathy)**: 2026-04-28
**MoAI-ADK License**: MIT
**Combined Compatibility**: Apache 2.0 imports distributed under MIT with Apache attribution preserved.

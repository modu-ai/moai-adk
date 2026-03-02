# Research: Context Memory System Analysis

## 1. System Inventory

### 1.1 Git-Based Context Memory (DELETE TARGET)

| File | Location | Role |
|------|----------|------|
| `workflows/context.md` | `.claude/skills/moai/workflows/` | Core workflow - git log parsing, context injection |
| `sync.md` Step 3.1.1 | `.claude/skills/moai/workflows/` | Context Memory Generation in commits |
| `manager-git.md` | `.claude/agents/moai/` | `## Context` section in commit message format |
| `CLAUDE.md` Section 16 | Project root | Context Search Protocol |
| `moai-memory.md` | `.claude/rules/moai/workflow/` | Memory hierarchy + session continuity rules |
| `context.yaml` | `.moai/config/sections/` | Context search configuration |
| `SKILL.md` (moai/SKILL.md) | `.claude/skills/moai/` | context/ctx/memory alias routing |

### 1.2 moai-foundation-context (KEEP - Token Management)

This skill is primarily about **token budget management**, NOT git-based context memory:
- Pattern 1: Token budget allocation (200K window)
- Pattern 2: Aggressive /clear strategy
- Pattern 3: Session state persistence
- Pattern 4: Multi-agent handoff protocols
- Pattern 5: Progressive disclosure

Decision: KEEP this skill. It serves a different purpose (token management).
Action: Remove "memory" from trigger keywords to avoid confusion.

### 1.3 SPEC System (ENHANCE TARGET)

Current SPEC structure:
```
.moai/specs/SPEC-XXX/
  spec.md         # Requirements (EARS format)
  plan.md         # Task decomposition
  acceptance.md   # Acceptance criteria
  research.md     # Deep codebase analysis
```

Missing: No section for implementation discoveries (Gotchas, runtime decisions).

### 1.4 Agent Skill References

8 agents reference `moai-foundation-context` in their skills list:
- manager-spec, manager-ddd, expert-backend, expert-frontend
- manager-docs, manager-project, manager-quality, manager-strategy

Action: Keep these references (foundation-context is about token management, not context memory).

## 2. Overlap Analysis

| Data Type | context.md | SPEC spec.md | Verdict |
|-----------|-----------|-------------|---------|
| Technical Decisions | `- Decision:` in git commits | Assumptions, Requirements | 100% overlap |
| Constraints | `- Constraint:` in git commits | Environment, Assumptions | 100% overlap |
| Technical Approach | `- Pattern:` in git commits | Technical Approach + research.md | 100% overlap |
| Risks | `- Risk:` in git commits | Manageable via SPEC | 90% overlap |
| Gotchas | `- Gotcha:` in git commits | NOT in SPEC | Unique - absorb into SPEC |
| User Preferences | `- UserPref:` in git commits | CLAUDE.local.md, Auto Memory | Covered elsewhere |
| MX Tag History | `## MX Tags Changed` | Code itself is truth | Unnecessary |

## 3. Usage Evidence

- context.md writing (manager-git `## Context` sections): Active in commits
- context.md reading (`/moai memory` invocation): Zero evidence of actual usage
- Session boundary git tags: Defined but no tags found in git

## 4. Impact Assessment

### Files to Modify (Both local + template source)

**DELETE entirely:**
- `.claude/skills/moai/workflows/context.md` (+ template)
- `.moai/config/sections/context.yaml` (+ template if exists)

**MODIFY (remove context memory sections):**
- `.claude/skills/moai/SKILL.md` - Remove context/ctx/memory alias
- `CLAUDE.md` - Remove Section 16
- `.claude/skills/moai/workflows/sync.md` - Remove Step 3.1.1
- `.claude/agents/moai/manager-git.md` - Remove `## Context` commit format
- `.claude/rules/moai/workflow/moai-memory.md` - Remove context memory references
- `.claude/skills/moai/workflows/plan.md` - Remove Context Loading section
- `.claude/skills/moai/workflows/run.md` - Remove Context Loading + Context Propagation
- `.claude/skills/moai-foundation-context/SKILL.md` - Remove "memory" trigger keyword
- `.claude/skills/moai-foundation-core/SKILL.md` - Remove context redirect reference
- `.claude/skills/moai-foundation-core/modules/agents-reference.md` - Update skill list
- `.claude/skills/moai-foundation-core/modules/commands-reference.md` - Remove context command
- `.claude/skills/moai-foundation-core/modules/execution-rules.md` - Update reference
- README.md / README.ko.md - Remove Context Memory section

**ENHANCE:**
- SPEC workflow: Add Implementation Log section to spec.md template
- manager-spec agent: Include Implementation Log guidance
- run.md workflow: Record discoveries in SPEC Implementation Log
- moai-memory.md rule: Simplify to SPEC-as-context only

### Estimated File Count
- Delete: 2 files (x2 for template = 4)
- Modify: ~15 files (x2 for template = ~30)
- Enhance: ~4 files (x2 for template = ~8)
- Total operations: ~42 file operations
